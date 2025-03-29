package add

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var name string

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new fleeting note",
	Long:  "Create a new fleeting note on inbox directory",
	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now()
		if name == "" {
			rl, err := readline.New("Name of new note file: ")
			cobra.CheckErr(err)
			defer rl.Close()

			line, err := rl.Readline()
			cobra.CheckErr(err)
			name = strings.Trim(line, " \t\r\n")
		}
		path := viper.GetString("inboxdir")
		notePath := path + "/" + now.Format("2006-01-02 15:04") + " - " + name + ".md"
		file, err := os.Create(notePath)
		cobra.CheckErr(err)

		values := make(map[string]string)
		values["name"] = name
		values["date"] = now.Format("2006-01-02")

		templateStr := viper.GetString("template")
		templ := template.Must(template.New("note").Parse(templateStr))

		templ.Execute(file, values)
		file.Close()

		open, _ := cmd.Flags().GetBool("open")
		if open {
			editor := viper.GetString("editor")
			cmd := exec.Command(editor, notePath)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		} else {
			fmt.Println("Note created at", notePath)
		}
	},
}

func init() {
	AddCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the note file")
	AddCmd.Flags().BoolP("open", "o", false, "Open the note file after creation")
}
