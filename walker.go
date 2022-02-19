package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/theckman/yacspin"
)

var NumWorkers = runtime.GOMAXPROCS(0)

var BufferSize = NumWorkers

var excludedDirs = map[string]bool{
	"vendor":  true,
	"cache":   true,
	"plugins": true,
	"uploads": true,
	"public":  true,
	"dist":    true,
	"bundle":  true,
}

type NodeModulesDirectory struct {
	name          string
	path          string
	size          int64
	sizeFormatted string
	info          os.FileInfo
	deleted       string
}

type NodeModulesDirectoryFunc func(directory NodeModulesDirectory, err error) error

var ErrNotDir = errors.New("not a directory")

type WalkerError struct {
	error error
	path  string
}

type WalkerErrorList struct {
	ErrorList []WalkerError
}

func (we WalkerError) Error() string {
	return we.error.Error()
}

func (wel WalkerErrorList) Error() string {
	if len(wel.ErrorList) > 0 {
		out := make([]string, len(wel.ErrorList))
		for i, err := range wel.ErrorList {
			out[i] = err.Error()
		}
		return strings.Join(out, "\n")
	}
	return ""
}

type Walker struct {
	wg                       sync.WaitGroup
	ewg                      sync.WaitGroup
	jobs                     chan string
	root                     string
	errors                   chan WalkerError
	errorList                WalkerErrorList
	nodeModulesDirectoryFunc NodeModulesDirectoryFunc
	spinner                  *yacspin.Spinner
}

func (w *Walker) addJob(path string) {
	w.wg.Add(1)
	select {
	case w.jobs <- path:
	default:
		w.processPath(path)
	}
}

func (w *Walker) collectErrors() {
	defer w.ewg.Done()
	for err := range w.errors {
		w.errorList.ErrorList = append(w.errorList.ErrorList, err)
	}
}

func (w *Walker) processPath(relpath string) error {
	defer w.wg.Done()
	path := filepath.Join(w.root, relpath)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal("processPath: " + err.Error())
		return err
	}

	for _, f := range files {
		if f.IsDir() && !strings.HasPrefix(f.Name(), ".") && !excludedDirs[f.Name()] {
			dirname := f.Name()
			subrelpath := filepath.Join(relpath, dirname)
			subpath := filepath.Join(w.root, subrelpath)

			w.spinner.Message(" Searching in " + subpath)

			if dirname == "node_modules" {
				err = fs.SkipDir

				size := diskUsage(subpath, f)
				nodeModulesDirectory := NodeModulesDirectory{
					name:          dirname,
					path:          subpath,
					info:          f,
					size:          size,
					sizeFormatted: ByteCountSI(size),
				}

				w.nodeModulesDirectoryFunc(nodeModulesDirectory, nil)
			}

			if err == filepath.SkipDir {
				return nil
			}

			if err != nil {
				fmt.Println("loop: " + err.Error())
				w.errors <- WalkerError{
					error: err,
					path:  subpath,
				}
				continue
			}

			w.addJob(subrelpath)
		}
	}
	return nil
}

func (w *Walker) worker() {
	for path := range w.jobs {
		err := w.processPath(path)
		if err != nil {
			fmt.Println("worker: " + err.Error())
			w.errors <- WalkerError{
				error: err,
				path:  path,
			}
		}
	}

}

func (w *Walker) Walk(relpath string, nodeModulesDirectoryFunc NodeModulesDirectoryFunc) error {
	w.errors = make(chan WalkerError, BufferSize)
	w.jobs = make(chan string, BufferSize)
	w.nodeModulesDirectoryFunc = nodeModulesDirectoryFunc

	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[11],
		SuffixAutoColon: true,
		Message:         " Searching",
		StopColors:      []string{"fgGreen"},
	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		log.Println(err)
		return err
	}
	w.spinner = spinner

	w.spinner.Start()

	w.ewg.Add(1)
	go w.collectErrors()

	for n := 1; n <= 4; n++ {
		go w.worker()
	}

	w.addJob(relpath)
	w.wg.Wait()
	close(w.jobs)
	close(w.errors)
	w.ewg.Wait()

	w.spinner.Stop()

	if len(w.errorList.ErrorList) > 0 {
		return w.errorList
	}

	return nil
}
