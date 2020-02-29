/*
Copyright © 2020 David Hu <coolbor@gmail.com>

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
	"io/ioutil"
	"strings"
	"sync"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

// fengCmd represents the feng command
var fengCmd = &cobra.Command{
	Use:   "feng",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("feng called")
		feeds := getFengHostFeed("")
		var wg sync.WaitGroup
		for _, strFeed := range feeds {
			s := Parse(strFeed)
			wg.Add(1)
			go func() {
				st, err := s.Ping()
				if err != nil {

					fmt.Println(s.EmojiFlag(), aurora.Red(s.Remarks), aurora.Red(s.Server), aurora.Red(err))
				} else {
					if st.PacketLoss == 100 {
						fmt.Printf("%s "+aurora.Red("%s %s los:%d\n").String(), s.EmojiFlag(), s.Remarks, s.Server, int(st.PacketLoss))
					} else {
						fmt.Println(s.EmojiFlag(), s.Remarks, s.Server, "AvgRtt:", aurora.Green(st.AvgRtt), "StdDevRtt:", st.StdDevRtt, "los:", st.PacketLoss)
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
		fmt.Println("========== done =============")
	},
}

func getFengHostFeed(url string) []string {
	b, err := ioutil.ReadFile("./feng.feed")
	if err != nil {
		return nil
	}
	strurls := Decode(string(b))
	return strings.Split(strurls, "\n")
}

func init() {
	rootCmd.AddCommand(fengCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fengCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fengCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
