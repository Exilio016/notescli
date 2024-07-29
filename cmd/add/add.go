package add

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

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
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Name of new note file: ")

			line, err := reader.ReadString('\n')
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
	},
}

func init() {
	AddCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the note file")
}
