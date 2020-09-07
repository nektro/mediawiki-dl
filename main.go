package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/nektro/go-util/types"
	"github.com/nektro/go-util/util"
	"github.com/nektro/go-util/vflag"
	"github.com/schollz/progressbar/v3"
)

var (
	sites []string
	expor []string
	wg    = new(sync.WaitGroup)
	sem   = types.NewSemaphore(6)
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

		last := ""
		l := 0
		j := 0
		bar := progressbar.Default(int64(1), urlO.Host)

		var doc *goquery.Document

		for {
			i := 0
			u1 := item + "?title=Special:AllPages&from=" + last
			fetchDoc(http.MethodGet, u1, "#bodyContent a", "#WikiaPage a").Each(func(_ int, el *goquery.Selection) {
				h, _ := el.Attr("href")
				if h[0] == '#' {
					return
				}
				if strings.Contains(h, "Special:AllPages") {
					return
				}
				if strings.HasPrefix(h, "https://") || strings.HasPrefix(h, "http://") {
					return
				}
				if l > 0 && i == 0 {
					i++
					return
				}
				h = strings.TrimPrefix(h, "/wiki")
				h = strings.TrimPrefix(h, "/index.php")
				h = strings.TrimPrefix(h, "/")
				last = h
				i++
				j++
				bar.ChangeMax(j)
				//
				u2 := item + "?title=Special:Export&action=submit&history=1&pages=" + h
				go func() {
					wg.Add(1)
					sem.Add()
					defer sem.Done()
					defer bar.Add(1)
					defer wg.Done()

					req, _ := http.NewRequest(http.MethodGet, u2, nil)
					req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64; rv:68.0) Gecko/20100101 Firefox/68.0")
					req.Header.Add("connection", "close")
					res, _ := http.DefaultClient.Do(req)
					bys, _ := ioutil.ReadAll(res.Body)

					d, _ := goquery.NewDocumentFromReader(bytes.NewReader(bys))
					if doc == nil {
						doc = d
						return
					}
					ns := d.Find("page").Nodes
					if len(ns) > 0 {
						doc.Find("siteinfo").AppendNodes(ns[0])
					}
				}()
			})
			if l == 0 {
				l = i
			}
			if i < l {
				break
			}
		}
		wg.Wait()
		bar.Add(1)

		f, _ := os.Create(dir2 + ".xml.gz")
		w := gzip.NewWriter(f)
		s, _ := doc.Find("mediawiki").Parent().Html()
		fmt.Fprintln(w, s)
	}
	wg.Wait()
}

func fetchDoc(method, urlS string, selctor ...string) *goquery.Selection {
	req, _ := http.NewRequest(method, urlS, nil)
	res, _ := http.DefaultClient.Do(req)
	doc, _ := goquery.NewDocumentFromResponse(res)
	for _, item := range selctor {
		s := doc.Find(item)
		if s.Size() > 0 {
			return s
		}
	}
	return nil
}
