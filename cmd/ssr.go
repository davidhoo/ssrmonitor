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
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/sparrc/go-ping"
)

// SSR struct
type SSR struct {
	Method        string
	Password      string
	Server        string
	Port          string
	Protocol      string
	Obfs          string
	Scheme        string
	ObfsHost      string
	Plugins       string
	RowPlugins    string
	Group         string
	ObfsParam     string
	ProtocolParam string
	Remarks       string
}

// Ping server
func (s *SSR) Ping() (*ping.Statistics, error) {
	pinger, err := ping.NewPinger(s.Server)
	if err != nil {
		return new(ping.Statistics), err
	}
	pinger.Timeout = 10000000000
	pinger.Count = 5
	pinger.Run()
	st := pinger.Statistics()
	return st, nil
}

// EmojiFlag is return emoji flag
func (s *SSR) EmojiFlag() string {
	for c, f := range EmojiFlags {
		if strings.Contains(s.Remarks, c) {
			return f + " "
		}
	}
	return UnknownEmojiFlag
}

// Parse ssr url
func Parse(rawurl string) *SSR {
	ssr := new(SSR)
	u, _ := url.Parse(rawurl)
	switch u.Scheme {
	case "ss":
		ssr = parseSS(u)
	case "ssr":
		ssr = parseSSR(u)
	}

	return ssr
}

func parseSS(u *url.URL) *SSR {
	ssr := new(SSR)
	ssr.Scheme = u.Scheme
	ssr.Server = u.Hostname()
	ssr.Port = u.Port()
	method, password := parseAuthority(u.User.Username())
	ssr.Method = method
	ssr.Password = password
	ssr.RowPlugins = u.Query().Get("plugin")
	ssr.Plugins, ssr.Obfs, ssr.ObfsHost = parsePlugin(ssr.RowPlugins)
	return ssr
}

func parseSSR(u *url.URL) *SSR {
	ssr := new(SSR)
	ssr.Scheme = u.Scheme
	urlsegments := strings.Split(Decode(u.Host), "?")
	segs := strings.Split(urlsegments[0], ":")
	ssr.Server = segs[0]
	ssr.Port = segs[1]
	ssr.Protocol = segs[2]
	ssr.Method = segs[3]
	ssr.Obfs = segs[4]
	ssr.Password = Decode(strings.TrimRight(segs[5], "/"))
	query, _ := url.ParseQuery(urlsegments[1])
	ssr.Group = Decode(query.Get("group"))
	ssr.ObfsParam = Decode(query.Get("obfsparam"))
	ssr.ProtocolParam = Decode(query.Get("protoparam"))
	ssr.Remarks = Decode(query.Get("remarks"))
	return ssr
}

func parseAuthority(authority string) (method string, password string) {
	authority = Decode(authority)
	methodAndPassword := strings.Split(authority, ":")
	return methodAndPassword[0], methodAndPassword[1]
}

func Decode(rawstr string) string {
	s, _ := base64.RawURLEncoding.DecodeString(rawstr)
	return string(s)
}

func parsePlugin(rowplugin string) (plugin string, obfs string, obfshost string) {
	// var plugin, obfs, obfshost string
	params := strings.Split(rowplugin, ";")

	if len(params) > 0 {
		plugin = params[0]
		for _, kvstring := range params[1:] {
			kv := strings.Split(kvstring, "=")
			switch kv[0] {
			case "obfs":
				obfs = kv[1]
			case "obfs-host":
				obfshost = kv[1]
			}
		}
	}

	return plugin, obfs, obfshost
}
