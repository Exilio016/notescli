/*
Copyright (C) 2024 Bruno Flávio Ferreira

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
package snippet

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/atotto/clipboard"
	"github.com/koki-develop/go-fzf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Snippet struct {
	name string
	content string
}

var snippets = []Snippet{}
var mutex sync.Mutex

var SnippetCmd = &cobra.Command{
	Use: "snippet",
	Short: "Find and copy a snippet of code",
	Long: "An fzf-like menu to find a snippet of code and copy it to clipboard",
	Run: func(cmd *cobra.Command, args []string) {
		dir := getSnippetDir()
		processFilesInDir(dir)
		needToPrint, err := cmd.Flags().GetBool("print")
		cobra.CheckErr(err)
		isTmux, err := cmd.Flags().GetBool("tmux")
		cobra.CheckErr(err)
		
		for _,i := range searchKeys() {
			clipboard.WriteAll(snippets[i].content)
			if needToPrint {
				fmt.Print(snippets[i].content)
			}
			if isTmux {
				cmd := exec.Command("tmux", "load-buffer", "-w", "-")
				in, err := cmd.StdinPipe()
				cobra.CheckErr(err)
				go func() {
					defer in.Close()
					io.WriteString(in, snippets[i].content)
				}()
				cobra.CheckErr(cmd.Run())
			}
		}
	},
}

func init() {
	SnippetCmd.Flags().BoolP("print", "p", false, "Print selected snippet insted of just copying it")
	SnippetCmd.Flags().BoolP("tmux", "t", false, "Store selected snippet to tmux buffer -")
}

func getSnippetDir() *os.File {
	path := viper.GetString("snippetsdir")
	dir, err := os.Open(path)
	if err != nil {
		fmt.Fscanln(os.Stderr, "Could not find snippet folder")
		os.Exit(1)
	}
	return dir
}

func searchKeys() []int {
	f, err := fzf.New(fzf.WithHotReload(&mutex))
	cobra.CheckErr(err)
	res, err := f.Find(&snippets, func(i int) string { return snippets[i].name }, 
		fzf.WithPreviewWindow(func(i, width, height int) string {
			return snippets[i].content
	}))
	cobra.CheckErr(err)
	return res
}

func processFilesInDir(dir *os.File) {
	children, err := dir.Readdir(0)
	cobra.CheckErr(err)
	for _,stat := range children {
		if !stat.IsDir() {
			if f, err := os.Open(dir.Name() + "/" + stat.Name()); err == nil {
				if content, err := io.ReadAll(f); err == nil {
					go parseContent(string(content), stat.Name())
				}
			}
		}
	}
}

func parseContent(content string, filename string) {
	keyRegex := regexp.MustCompile("(?m)\\s*#{3} (.*)$")
	valueRegex := regexp.MustCompile("(?s)\\x60{3}\\w*(.*?)(?:\\x60\\x60\\x60)")

	keys := keyRegex.FindAllStringSubmatch(content, -1)
	values := valueRegex.FindAllStringSubmatch(content, -1)

	if len(keys) != len(values) {
		fmt.Printf("WARNING: snippet file '%s' malformated! Number of headers doesn't match number of snippets\n", filename)
		fmt.Printf("Headers: %d, snippets: %d\n", len(keys), len(values))
	}
	len := min(len(keys), len(values))
	for i := 0; i < len; i++ {
		mutex.Lock()
		key := keys[i][1]
		value := strings.Trim(values[i][1], "\r\n\t ")
		snippets = append(snippets, Snippet{key, value})
		mutex.Unlock()
	}
}

