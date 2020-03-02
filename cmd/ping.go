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
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/logrusorgru/aurora"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// fengCmd represents the feng command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping all SSR servers and sorting them",
	Long:  `Ping all SSR servers and sorting them.`,
	Run: func(cmd *cobra.Command, args []string) {

		feedURLs := viper.GetStringSlice("urls")
		if len(feedURLs) < 1 {
			cmd.Args = cobra.MinimumNArgs(1)
			if len(args) > 0 {
				feedURLs = append(feedURLs, args[0])
			} else {
				cmd.PrintErrln("Error: requires at least 1 arg(s), only received 0")
				cmd.Usage()
			}
		}

		for _, feedURL := range feedURLs {

			feeds, err := getFengHostFeed(feedURL)
			if err != nil {
				cmd.PrintErrf("Error: %s\n", err)
				cmd.Usage()
				return
			}
			var wg sync.WaitGroup
			var ss SSRs
			for _, strFeed := range feeds {
				s, err := Parse(strFeed)
				if err != nil {
					continue
				}
				wg.Add(1)
				go func() {
					st, err := s.Ping()
					if err != nil {

						fmt.Println(s.EmojiFlag()+" ", aurora.Red(s.Remarks), aurora.Red(s.Server), aurora.Red(err))
					} else {
						if int(st.PacketLoss) == 100 {
							fmt.Printf("%s "+aurora.Red("%s %s los:%d\n").String(), s.EmojiFlag()+" ", s.Remarks, s.Server, int(st.PacketLoss))
						} else {
							fmt.Println(s.EmojiFlag()+" ", s.Remarks, s.Server, "AvgRtt:", aurora.Green(st.AvgRtt), "StdDevRtt:", st.StdDevRtt, "los:", int(st.PacketLoss))
							s.AvgRtt, s.StdDevRtt, s.PacketLoss = st.AvgRtt, st.StdDevRtt, st.PacketLoss
							ss = append(ss, *s)
						}
					}
					wg.Done()
				}()
			}
			wg.Wait()

			sort.Sort(ss)

			headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
			columnFmt := color.New(color.FgYellow).SprintfFunc()
			tbl := table.New("#", "Flag", "Server", "AvgRtt", "StdDevRtt", "PacketLoss", "Remarks")
			tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithPadding(1)
			for i, s := range ss {
				tbl.AddRow(i, s.EmojiFlag(), s.Server, fmt.Sprintf("%5dms   ", s.AvgRtt.Milliseconds()), s.StdDevRtt, fmt.Sprintf("%4d%%", int(s.PacketLoss)), s.Remarks)
			}
			tbl.Print()
		}
	},
}

func getFengHostFeed(url string) ([]string, error) {

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return nil, fmt.Errorf("%s has not content returned.(%s)", url)
	}
	defer res.Body.Close()

	strurls := Decode(string(b))

	return strings.Split(strurls, "\n"), nil
}

func init() {

	// pingCmd.PersistentFlags().StringVar(&Url, "url", "", "ssr subscribe url")
	// pingCmd.ArgsLenAtDash()

	rootCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fengCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fengCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
