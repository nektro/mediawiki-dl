package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/nektro/go-util/types"
	"github.com/nektro/go-util/util"
	"github.com/nektro/go-util/vflag"
)

var (
	sites []string
	wg    = new(sync.WaitGroup)
	sem   = types.NewSemaphore(5)
)

func main() {
	vflag.StringArrayVar(&sites, "site", []string{}, "")
	vflag.Parse()

	dir1 := "./data"
	os.Mkdir(dir1, os.ModePerm)
	for _, item := range sites {
		urlO, err := url.Parse(item)
		util.DieOnError(err)
		dir2 := dir1 + "/" + urlO.Host
		os.Mkdir(dir2, os.ModePerm)

		last := ""
		l := 0
		j := 0
		for {
			i := 0
			u1 := item + "?title=Special:AllPages&from=" + last
			fmt.Println(u1)
			fetchDoc(http.MethodGet, u1, "#bodyContent a").Each(func(_ int, el *goquery.Selection) {
				h, _ := el.Attr("href")
				if h[0] == '#' {
					return
				}
				if strings.HasPrefix(h, "/wiki") {
					h = h[5:]
				}
				if strings.Contains(h, "Special:AllPages") {
					return
				}
				if l > 0 && i == 0 {
					i++
					return
				}
				h = strings.TrimPrefix(h, "/index.php")
				h = strings.TrimPrefix(h, "/")
				last = h
				i++
				j++
				//
				page, _ := url.QueryUnescape(h)
				page = strings.ReplaceAll(page, "/", "∕")
				u2 := item + "?title=Special:Export&action=submit&history=1&pages=" + h
				p := dir2 + "/" + page + ".xml"
				go saveFile(u2, p)
			})
			wg.Wait()
			fmt.Println()
			if l == 0 {
				l = i
			}
			if i < l {
				break
			}
		}
	}
	wg.Wait()
	fmt.Println("done")
}

func fetchDoc(method, urlS, selctor string) *goquery.Selection {
	req, _ := http.NewRequest(method, urlS, nil)
	res, _ := http.DefaultClient.Do(req)
	doc, _ := goquery.NewDocumentFromResponse(res)
	return doc.Find(selctor)
}

func fetchBytes(method, urlS string) io.ReadCloser {
	req, _ := http.NewRequest(method, urlS, nil)
	res, _ := http.DefaultClient.Do(req)
	return res.Body
}

func saveFile(urlS string, pathS string) {
	wg.Add(1)
	sem.Add()

	defer sem.Done()
	defer wg.Done()

	f, _ := os.Create(pathS)
	defer f.Close()

	req, _ := http.NewRequest(http.MethodGet, urlS, nil)
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64; rv:68.0) Gecko/20100101 Firefox/68.0")
	req.Header.Add("connection", "close")

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	io.Copy(f, res.Body)
	fmt.Print("|")
}
