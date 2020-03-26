package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type QiitaJson []struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func readJson(bin []byte) (QiitaJson, error) {
	var q QiitaJson
	err := json.Unmarshal(bin, &q)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	bin, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return bin, err
}

func downloadIndex(user string, page, perPage int) ([]byte, error) {
	url := fmt.Sprintf("https://qiita.com/api/v2/users/%s/items?page=%d&per_page=%d",
		user, page, perPage)
	return download(url)
}

var han2zen = strings.NewReplacer(
	`"`, "\uFF02",
	`*`, "\uFF0A",
	`/`, "\uFF0F",
	`:`, "\uFF1A",
	`<`, "\uFF1C",
	`?`, "\uFF1F",
	`>`, "\uFF1E",
	`\`, "\uFF3C",
	`|`, "\uFF5C")

func safeFileName(name string) string {
	return han2zen.Replace(name)
}

func mains(args []string) error {
	const PER_PAGE = 60
	if len(args) <= 0 {
		exename, err := os.Executable()
		if err != nil {
			return err
		}
		return fmt.Errorf("Usage: %s USERNAMEs...", exename)
	}
	for _, user := range args {
		for page := 1; true; page++ {
			bin, err := downloadIndex(user, page, PER_PAGE)
			if err != nil {
				return err
			}
			if bin[0] != '[' {
				break
			}
			articles, err := readJson(bin)
			if err != nil {
				return err
			}
			if len(articles) <= 0 {
				break
			}
			for _, article1 := range articles {
				bin, err = download(article1.Url + ".md")
				if err != nil {
					return err
				}
				fname := safeFileName(article1.Title) + ".md"
				err = ioutil.WriteFile(fname, bin, 0644)
				if err != nil {
					return err
				}
				fmt.Printf("%s\n-> %s\n\n", article1.Url, fname)
				time.Sleep(time.Second)
			}
			if len(articles) < PER_PAGE {
				break
			}
		}
	}
	return nil
}

func main() {
	if err := mains(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
