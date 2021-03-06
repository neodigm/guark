// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/guark/guark/app/utils"
	"github.com/urfave/cli/v2"
)

func path(elem ...string) string {
	return filepath.Join(append([]string{wdir}, elem...)...)
}

func CheckWorkingDir(c *cli.Context) (err error) {

	if utils.IsFile("guark.yaml") == false {
		err = fmt.Errorf("could not find: guark.yaml, cd to a guark project!")
	}

	return
}

func getHost() string {

	cmd := exec.Command("go", "env", "GOHOSTOS")
	out, err := cmd.Output()

	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(out))
}

func GitFile(repo string, file string, auth string) (content []byte, e error) {

	url, e := url.Parse(repo)

	if e != nil {
		return
	}

	switch url.Host {
	case "github.com":
		content, e = getGithubFile(url, file, auth)
		return
	case "bitbucket.org":
		content, e = getBitbucketFile(url, file, auth)
		return
	}

	e = fmt.Errorf("Unknown host: %s", url.Host)

	return
}

func getGithubFile(url *url.URL, file string, auth string) (content []byte, e error) {

	content, e = GetContentFromUrl(fmt.Sprintf("https://api.github.com/repos%s/contents/%s", url.Path, file), auth)

	if e != nil {
		return
	}

	var dl struct {
		URL string `json:"download_url"`
	}

	e = json.Unmarshal(content, &dl)

	if e != nil {
		return
	}

	content, e = GetContentFromUrl(dl.URL, auth)
	return
}

func getBitbucketFile(url *url.URL, file string, auth string) ([]byte, error) {

	return GetContentFromUrl(fmt.Sprintf("https://api.bitbucket.org/2.0/repositories%s/src/master/%s", url.Path, file), auth)
}

func GetContentFromUrl(url string, auth string) (content []byte, e error) {

	res, e := http.Get(UrlAuth(url, auth))

	if e != nil {
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {

		e = fmt.Errorf("Request error: %v for %s", res.StatusCode, strings.Replace(url, auth+"@", "", 1))
		return
	}

	content, e = ioutil.ReadAll(res.Body)

	return
}

func IsUrl(u string) bool {

	if u == "" {
		return false
	}

	_, err := url.ParseRequestURI(u)

	return err == nil
}

func UrlAuth(url string, auth string) string {

	if auth != "" {
		return strings.Replace(url, "https://", fmt.Sprintf("https://%s@", auth), 1)
	}

	return url
}
