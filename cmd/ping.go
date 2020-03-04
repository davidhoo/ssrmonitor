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
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/logrusorgru/aurora"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

// pingCmd represents the feng command
var pingCmd = &cobra.Command{
	Use:   "ping [ssr subscribe url]",
	Short: "Ping all SSR servers and sorting them",
	Long:  `Ping all SSR servers and sorting them.`,
	Args: func(cmd *cobra.Command, args []string) error {
		c := len(args)
		if c == 1 {
			return nil
		} else if c > 1 {
			return errors.New(fmt.Sprintf("accepts 1 arg(s), received %d\n", c))
		}
		err := viper.ReadInConfig()
		if err != nil {
			return err
		}
		feedUrls := viper.GetStringSlice("urls")
		if len(feedUrls) < 1 {
			return errors.New("missing configuration for 'configPath'")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		var feedURLs []string
		if len(args) == 1 {
			feedURLs = args
		} else {
			feedURLs = viper.GetStringSlice("urls")
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
			log.Println("pinging ...")
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
			log.Println("sorting ...")

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
	fmt.Printf("Downloading...\n%s\n", url)

	p := mpb.New(mpb.WithWidth(64))
	var total int64
	bar := p.AddBar(
		total,
		mpb.PrependDecorators(decor.Counters(decor.UnitKiB, "% .1f / % .1f")),
		mpb.AppendDecorators(decor.Percentage()),
	)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	var f bytes.Buffer
	for {
		n, err := res.Body.Read(buf)
		if err != nil || n == 0 {
			break
		}
		f.Write(buf[:n])

		total += int64(n)
		bar.SetTotal(total+1024, false)
		bar.IncrBy(n)
	}
	bar.SetTotal(total, true)
	p.Wait()
	defer res.Body.Close()
	strurls := Decode(string(f.Bytes()))
	return strings.Split(strurls, "\n"), nil
}

func init() {
	rootCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fengCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fengCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
