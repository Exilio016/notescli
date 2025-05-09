/*
Copyright (C) 2024-2025 Bruno Flávio Ferreira

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
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/Masterminds/sprig/v3"
	"github.com/atotto/clipboard"
	"github.com/chzyer/readline"
	"github.com/koki-develop/go-fzf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"notescli/cmd/utils"
)

type Snippet struct {
	name    string
	content string
	inputs  []Input
}

type Input struct {
	name         string
	defaultValue string
}

var snippets = []Snippet{}
var mutex sync.Mutex

var SnippetCmd = &cobra.Command{
	Use:   "snippet",
	Short: "Find and copy a snippet of code",
	Long:  "An fzf-like menu to find a snippet of code and copy it to clipboard",
	Run: func(cmd *cobra.Command, args []string) {
		dir := getSnippetDir()
		processFilesInDir(dir)
		needToPrint, err := cmd.Flags().GetBool("print")
		cobra.CheckErr(err)
		isTmux, err := cmd.Flags().GetBool("tmux")
		cobra.CheckErr(err)
		for _, i := range searchKeys() {
			result := handleInputs(snippets[i])
			clipboard.WriteAll(result)
			if needToPrint {
				fmt.Print(result)
			}
			if isTmux {
				cmd := exec.Command("tmux", "load-buffer", "-w", "-")
				in, err := cmd.StdinPipe()
				cobra.CheckErr(err)
				go func() {
					defer in.Close()
					io.WriteString(in, result)
				}()
				cobra.CheckErr(cmd.Run())
			}
		}
	},
}

func handleInputs(snippet Snippet) string {
	result := snippet.content
	writter := bytes.NewBufferString("")
	if len(snippet.inputs) > 0 {
		templ := template.Must(template.New("result").Funcs(sprig.FuncMap()).Parse(result))
		values := map[string]string{}
		for _, in := range snippet.inputs {
			prompt := fmt.Sprintf("Please provide value for \"%s\": ", in.name)
			if in.defaultValue != "" {
				prompt = fmt.Sprintf("Please provide value for \"%s\" or press enter for default \"%s\": ", in.name, in.defaultValue)
			}
			rl, err := readline.New(prompt)
			cobra.CheckErr(err)
			defer rl.Close()
			value, err := rl.Readline()
			cobra.CheckErr(err)
			if value == "" {
				values[in.name] = in.defaultValue
			} else {
				values[in.name] = value
			}
		}
		templ.Execute(writter, values)
		result = writter.String()
	}
	return result
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
			return utils.TruncateContent(width, height, snippets[i].content)
		}))
	cobra.CheckErr(err)
	return res
}

func processFilesInDir(dir *os.File) {
	children, err := dir.Readdir(0)
	cobra.CheckErr(err)
	for _, stat := range children {
		if !stat.IsDir() && stat.Name() != "README.md" {
			if f, err := os.Open(dir.Name() + "/" + stat.Name()); err == nil {
				if content, err := io.ReadAll(f); err == nil {
					go parseContent(string(content), stat.Name())
				}
			}
		}
	}
}

func parseContent(content string, filename string) {
	keyRegex := regexp.MustCompile("(?m)^\\s*#{3} (.*)$")
	valueRegex := regexp.MustCompile("(?s)(INPUTS:\\n(?:- (?:\\w+(?::[^\\n]+)*)\\n)+)*\\x60{3}\\w*(.*?)(?:\\x60\\x60\\x60)")
	inputRegex := regexp.MustCompile("(?m)-\\s+(\\w+)\\s*(?::\\s*(.+))*$")

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
		input := values[i][1]
		list := []Input{}
		inputs := inputRegex.FindAllStringSubmatch(input, -1)
		for _, in := range inputs {
			name := strings.Trim(in[1], "\r\n\t ")
			defaultValue := in[2]
			list = append(list, Input{name, defaultValue})
		}
		value := strings.Trim(values[i][2], "\r\n\t ")
		snippets = append(snippets, Snippet{key, value, list})
		mutex.Unlock()
	}
}
