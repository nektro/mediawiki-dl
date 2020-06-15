package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/nektro/go-util/types"
	"github.com/nektro/go-util/util"
	"github.com/nektro/go-util/vflag"
	"github.com/schollz/progressbar/v3"
	"go.etcd.io/bbolt"
)

var (
	sites []string
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

		db, err := bbolt.Open(dir2+".bolt", 0600, nil)
		util.DieOnError(err)
		defer db.Close()

		for {
			i := 0
			u1 := item + "?title=Special:AllPages&from=" + last
			fetchDoc(http.MethodGet, u1, "#bodyContent a").Each(func(_ int, el *goquery.Selection) {
				h, _ := el.Attr("href")
				if h[0] == '#' {
					return
				}
				if strings.Contains(h, "Special:AllPages") {
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
				page, _ := url.QueryUnescape(h)
				u2 := item + "?title=Special:Export&action=submit&history=1&pages=" + h
				go saveFile(u2, page, bar, db)
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
	}
	wg.Wait()
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

func saveFile(urlS string, page string, bar *progressbar.ProgressBar, db *bbolt.DB) {
	wg.Add(1)
	sem.Add()

	defer sem.Done()
	defer bar.Add(1)
	defer wg.Done()

	time.Sleep(time.Second / 100)
	brk := false
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pages"))
		if b == nil {
			return nil
		}
		if b.Get([]byte(page)) != nil {
			brk = true
		}
		return nil
	})
	if brk {
		return
	}

	req, _ := http.NewRequest(http.MethodGet, urlS, nil)
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64; rv:68.0) Gecko/20100101 Firefox/68.0")
	req.Header.Add("connection", "close")

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	bys, _ := ioutil.ReadAll(res.Body)

	db.Update(func(tx *bbolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte("pages"))
		b.Put([]byte(page), bys)
		return nil
	})
}
