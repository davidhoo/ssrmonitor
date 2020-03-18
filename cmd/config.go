/*
Copyright © 2020 David Hu<coolbor@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func checkConfig() {
	if err := viper.ReadInConfig(); err != nil {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {

			if err := viper.WriteConfigAs(home + "/.ssrmonitor.yaml"); err != nil {
				fmt.Print(err)
				os.Exit(1)
			}
		}
	}
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the ssr feed",
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()
		printConfig()
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new ssr feed into config file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		checkConfig()
		urls := viper.GetStringSlice("urls")
		urls = append(urls[:], strings.TrimSpace(args[0]))
		viper.Set("urls", urls)
		_ = viper.WriteConfig()
		printConfig()
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the ssr feed via index",
	Args: func(cmd *cobra.Command, args []string) error {
		c := len(args)
		if c != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d\n", c)
		}
		if _, err := strconv.ParseInt(args[0], 10, 32); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		idx, _ := strconv.ParseInt(args[0], 10, 32)
		urls := viper.GetStringSlice("urls")
		if int(idx) >= len(urls) || int(idx) < 0 {
			cmd.PrintErrf("Error: %v is not a valid index\n", idx)
			_ = cmd.Usage()
			os.Exit(1)
		}
		urls = append(urls[0:idx], urls[idx+1:]...)
		viper.Set("urls", urls)
		_ = viper.WriteConfig()
		printConfig()
	},
}

func init() {
	configCmd.AddCommand(deleteCmd)
	configCmd.AddCommand(addCmd)
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func printConfig() {
	urls := viper.GetStringSlice("urls")
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("#", "url")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithPadding(1)
	for i, url := range urls {
		tbl.AddRow(i, strings.TrimSpace(url))
	}
	tbl.Print()
}
