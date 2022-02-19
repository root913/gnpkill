package main

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
)

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func diskUsage(currentPath string, info os.FileInfo) int64 {
	size := info.Size()

	if !info.IsDir() {
		return size
	}

	dir, err := os.Open(currentPath)

	if err != nil {
		fmt.Println(err)
		return size
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, file := range files {
		if file.Name() == "." || file.Name() == ".." {
			continue
		}
		size += diskUsage(currentPath+"/"+file.Name(), file)
	}

	return size
}

func nodeModulesTable(list map[string]NodeModulesDirectory, withDelete bool) {
	tw := table.NewWriter()
	if withDelete {
		tw.SetRowPainter(table.RowPainter(func(row table.Row) text.Colors {
			if row[3] == "Success" {
				return text.Colors{text.BgGreen, text.FgBlack}
			} else {
				return text.Colors{text.BgRed, text.FgBlack}
			}
		}))
		total := 0
		tw.AppendHeader(table.Row{"Name", "Path", "Size", "Deleted?"})
		for _, dir := range list {
			tw.AppendRow(table.Row{dir.name, dir.path, dir.sizeFormatted, dir.deleted})
			total = total + int(dir.size)
		}
		tw.AppendFooter(table.Row{"", "Total", ByteCountSI(int64(total))})
	} else {
		total := 0
		tw.AppendHeader(table.Row{"Name", "Path", "Size"})
		for _, dir := range list {
			tw.AppendRow(table.Row{dir.name, dir.path, dir.sizeFormatted})
			total = total + int(dir.size)
		}
		tw.AppendFooter(table.Row{"", "Total", ByteCountSI(int64(total))})
	}

	fmt.Println(tw.Render())
}

func getKeys(m map[string]NodeModulesDirectory) []string {
	j := 0
	keys := make([]string, len(m))
	for k := range m {
		keys[j] = k
		j++
	}
	return keys
}

func Checkboxes(label string, opts map[string]NodeModulesDirectory) map[string]NodeModulesDirectory {
	res := []string{}
	prompt := &survey.MultiSelect{
		Message: label,
		Options: getKeys(opts),
	}
	survey.AskOne(prompt, &res)

	result := map[string]NodeModulesDirectory{}
	for _, f := range res {
		result[f] = opts[f]
	}

	return result
}
