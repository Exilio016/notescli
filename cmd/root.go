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
package cmd

import (
	"log"
	"notescli/cmd/add"
	"notescli/cmd/snippet"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "notescli",
	Short: "A CLI application to manage notes",
	Long: `notescli is an application created to help you manage and use your notes more efficiently.
	This tool provides helpful commands to manage your notes and snippets.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}
var cfgFile = ""

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.config/notescli/config.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	viper.SetDefault("snippetsdir", home+"/Notes/snippets")
	viper.SetDefault("inboxdir", home+"/Notes/inbox")
	viper.SetDefault("editor", "vim")
	viper.SetDefault("template", `---
date: {{.date}}
tags:
	-
hubs:
	- "[[]]"
references:
	-
---
# {{.name}}
	`)

	rootCmd.AddCommand(snippet.SnippetCmd)
	rootCmd.AddCommand(add.AddCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home + "/.config/notescli")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Println(err)
	}
}
