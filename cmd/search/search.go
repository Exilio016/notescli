/*
Copyright (C) 2025 Bruno Flávio Ferreira

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package search

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/koki-develop/go-fzf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"notescli/cmd/utils"
)

type Tag struct {
	name     string
	filename string
	path     string
}

var mutex sync.Mutex
var tags []Tag

var SearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search a note by its tags and open it with the editor",
	Long:  "A fzf-like menu to search notes using tags and open it with the default editor",
	Run: func(cmd *cobra.Command, args []string) {
		dir := getVaultDir()
		f, err := fzf.New(fzf.WithHotReload(&mutex))
		cobra.CheckErr(err)
		go parseFileTags(dir)
		findResult, err := f.Find(&tags, func(i int) string { return tags[i].name }, fzf.WithPreviewWindow(fzfPreviewWindow))
		cobra.CheckErr(err)
		for _, i := range findResult {
			note := tags[i].path
			editor := viper.GetString("editor")
			editorCmd := exec.Command(editor, note)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr
			editorCmd.Run()
		}
	},
}

func fzfPreviewWindow(i, width, height int) string {
	f, err := os.Open(tags[i].path)
	cobra.CheckErr(err)
	content, _ := io.ReadAll(f)
	return tags[i].filename + "\n\n" + utils.TruncateContent(width, height-4, string(content))
}

func getVaultDir() *os.File {
	path := viper.GetString("vault")
	dir, err := os.Open(path)
	if err != nil {
		fmt.Fscanln(os.Stderr, "Could not find snippet folder")
		os.Exit(1)
	}
	return dir
}

func parseFileTags(dir *os.File) {
	children, err := dir.ReadDir(0)
	cobra.CheckErr(err)
	for _, stat := range children {
		if !stat.IsDir() {
			f, err := os.Open(dir.Name() + "/" + stat.Name())
			cobra.CheckErr(err)
			fileTags := processFile(f)
			for _, tag := range fileTags {
				mutex.Lock()
				tags = append(tags, Tag{name: tag, filename: stat.Name(), path: dir.Name() + "/" + stat.Name()})
				mutex.Unlock()
			}
		}
	}
}

func processFile(f *os.File) []string {
	content, err := io.ReadAll(f)
	cobra.CheckErr(err)
	line := []byte{}
	foundTags := false
	list := []string{}
	if "---\n" == string(content[0:4]) {
		for _, c := range content[4:] {
			if c != '\n' {
				line = append(line, c)
			} else {
				s := strings.Trim(string(line), " \n\t\r")
				if foundTags && strings.HasPrefix(s, "- ") {
					list = append(list, s[2:])
				} else if foundTags {
					break
				}

				if s == "tags:" {
					foundTags = true
				} else if s == "---" {
					break
				}
				line = []byte{}
			}
		}
	}
	return list
}
