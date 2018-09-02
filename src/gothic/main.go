package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/gocolly/colly"
)

var RegExpHTTPURL, err1 = regexp.Compile("^https?://")
var RegExpMailto, err2 = regexp.Compile("^[a-z]+:[^/]")
var agents = []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36 OPR/55.0.2994.44", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.1.2 Safari/605.1.15"}

type byLength []string

func (s byLength) Len() int {
	return len(s)
}
func (s byLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byLength) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}

type Spider struct {
	origin *url.URL
	links  []string
	parsed []string
}

func (spider Spider) Parse(url string, ch chan string) {
	c := colly.NewCollector(colly.UserAgent(agents[rand.Int()%len(agents)]))
	origin := spider.origin.String()
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := strings.Split(e.Attr("href"), "#")[0]
		if len(strings.Trim(link, " ")) > 0 {
			if RegExpHTTPURL.MatchString(link) {
				if 0 == strings.Index(link, origin) {
					ch <- link
				}
			} else if len(link) > 0 {
				if '/' == link[0] {
					ch <- origin + link
				} else if !RegExpMailto.MatchString(link) {
					ch <- origin + "/" + link
				}
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

func (spider Spider) Next(url string) {
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
				// if 0 == len(spider.parsed)%10 {
				// 	spider.Save()
				// }
				sort.Sort(byLength(spider.links))
				next := spider.links[0]
				spider.links = spider.links[1:]
				fmt.Println(next)
				spider.Next(next)
			}
			break
		}
	}
}

func (spider Spider) Run() {
	spider.links = append(spider.links, spider.origin.String())
	for _, link := range spider.links {
		spider.Next(link)
	}
}

func (spider Spider) GetLinksFilename() string {
	return spider.origin.Hostname() + "-links.txt"
}

func (spider Spider) GetParsedFilename() string {
	return spider.origin.Hostname() + "-parsed.txt"
}

func (spider Spider) Load() {
	links, err1 := ioutil.ReadFile(spider.GetLinksFilename())
	if nil == err1 {
		spider.links = strings.Split(strings.Trim(string(links), " "), "\n")
	}
	parsed, err2 := ioutil.ReadFile(spider.GetParsedFilename())
	if nil == err2 {
		spider.parsed = strings.Split(strings.Trim(string(parsed), " "), "\n")
	}
}

func (spider Spider) Save() error {
	err := ioutil.WriteFile(spider.GetLinksFilename(), []byte(strings.Join(spider.links, "\n")), 0644)
	if nil != err {
		return err
	}
	err = ioutil.WriteFile(spider.GetParsedFilename(), []byte(strings.Join(spider.parsed, "\n")), 0644)
	if nil != err {
		return err
	}
	return nil
}

func main() {
	rand.Seed(time.Now().Unix())
	origin, err0 := url.Parse(os.Args[1])
	if nil != err0 {
		panic(err0)
	}

	s := Spider{origin: origin}
	signals := make(chan os.Signal, 1)
	go s.Run()
	signal.Notify(signals, syscall.SIGINT)
	<-signals
	s.Save()
}
