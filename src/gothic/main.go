package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

var RegExpHTTPURL, err1 = regexp.Compile("^https?://")

type Spider struct {
	origin string
	links  []string
	parsed []string
}

func (spider Spider) Parse(url string, ch chan string) {
	c := colly.NewCollector(colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36"))
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := strings.Split(e.Attr("href"), "#")[0]
		if RegExpHTTPURL.MatchString(link) {
			if 0 == strings.Index(link, spider.origin) {
				ch <- link
			}
		} else if len(link) > 0 {
			if '/' == link[0] {
				ch <- spider.origin + link
			} else {
				ch <- spider.origin + "/" + link
			}
		}
	})
	c.OnScraped(func(r *colly.Response) {
		close(ch)
	})
	c.OnError(func(r *colly.Response, e error) {
		close(ch)
	})
	c.Visit(url)
}

func (spider Spider) Load(url string) {
	ch := make(chan string)
	go spider.Parse(url, ch)
	spider.parsed = append(spider.parsed, url)
one:
	for {
		link, more := <-ch
		if more {
			for _, x := range spider.parsed {
				if x == link {
					continue one
				}
			}
			for _, x := range spider.links {
				if x == link {
					continue one
				}
			}
			spider.links = append(spider.links, link)
		} else {
			_, err := http.Get("https://web.archive.org/save/" + url)
			if nil != err {
				fmt.Println(err)
			}
			if len(spider.links) > 0 {
				next := spider.links[0]
				spider.links = spider.links[1:]
				fmt.Println(next)
				spider.Load(next)
			}
			break
		}
	}
}

func (spider Spider) Run() {
	spider.links = append(spider.links, spider.origin)
	for _, link := range spider.links {
		spider.Load(link)
	}
}

func main() {
	s := Spider{origin: os.Args[1]}
	s.Run()
}
