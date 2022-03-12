package main

import (
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var (
	version = "unknown"
)

func startListingOfDirs(path string) map[string]NodeModulesDirectory {
	result := map[string]NodeModulesDirectory{}
	w := Walker{
		root: path,
	}
	w.Walk("", func(directory NodeModulesDirectory, err error) error {
		result[directory.path] = directory

		return nil
	})

	return result
}

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "npkill in GO"
	app.Version = version
	app.Action = func(c *cli.Context) error {
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
			return err
		}

		list := startListingOfDirs(path)

		if len(list) == 0 {
			color.Yellow("No node_modules found. Exiting.")
			return nil
		}
		nodeModulesTable(list, false)
		answers := Checkboxes(
			"Which directories do You want to remove?",
			list,
		)
		if len(answers) > 0 {
			for key, dir := range answers {
				err := os.RemoveAll(key)
				if err != nil {
					dir.deleted = err.Error()
				} else {
					dir.deleted = "Success"
				}
				answers[key] = dir
			}
			nodeModulesTable(answers, true)
		} else {
			color.Yellow("Nothing selected. Exiting.")
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
