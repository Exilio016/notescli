package add

import (
	"fmt"
	"os"
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
			fmt.Print("Name of new note file: ")
			fmt.Scanf("%s\n", &name)
		}
		path := viper.GetString("inboxdir")
		file, err := os.Create(path + "/" + now.Format("2006-01-02 15:04") + " - " + name + ".md")
		cobra.CheckErr(err)
		defer file.Close()

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
