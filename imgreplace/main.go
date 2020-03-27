package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	_ "github.com/mattn/getwild"
)

func download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	bin, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return bin, err
}

var rxAnchor = regexp.MustCompile(`\!\[.*?\]\(https://.*?\)`)

func mains(args []string) error {
	for _, fname := range args {
		bin, err := ioutil.ReadFile(fname)
		if err != nil {
			return err
		}
		count := 0

		bin = rxAnchor.ReplaceAllFunc(bin, func(m []byte) []byte {
			kakko := bytes.IndexByte(m, '(')
			url := m[kakko+1 : len(m)-1]

			pic, err := download(string(url))
			if err != nil {
				return m
			}
			count++

			contentType := http.DetectContentType(pic)

			newFname := uuid.New().String()
			if p := strings.SplitN(contentType, "/", 2); len(p) >= 2 {
				newFname = fmt.Sprintf("%s.%s", newFname, p[1])
			}
			ioutil.WriteFile(newFname, pic, 0644)

			result := make([]byte, 0, len(m))
			result = append(result, m[:kakko+1]...)
			result = append(result, newFname...)
			result = append(result, ')')

			return result
		})

		if count > 0 {
			backup := fname + ".bak"
			os.Remove(backup)
			if os.Rename(fname, backup) == nil {
				ioutil.WriteFile(fname, bin, 0644)
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
