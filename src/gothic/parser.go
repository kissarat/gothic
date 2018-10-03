package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE link (
	id text primary key not null,
	created timestamp not null default current_timestamp,
	archived timestamp
)`

var RegexHref = regexp.MustCompile("href=\"([^\"]+)\"")

type Spider struct {
	db     *sql.DB
	client http.Client
}

func (spider Spider) Init() error {
	var count int
	spider.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name = 'link'`, &count)
	if 0 == count {
		spider.db.Exec(schema)
	}
	spider.client = http.Client{Timeout: time.Duration(15 * time.Second)}
	return nil
}

func (spider Spider) Fetch(url string) error {
	r, err1 := spider.client.Get(url)
	if nil != err1 {
		return err1
	}
	data, err2 := ioutil.ReadAll(r.Body)
	if nil != err2 {
		return err2
	}
	for _, m := range RegexHref.FindAllStringSubmatch(string(data), -1) {
		s := m[1]
		if strings.HasPrefix(s, "https://te.20minut.ua") {
			spider.db.Exec("INSERT INTO link (id) VALUES ($1)", s)
		} else if len(s) > 0 && '/' == s[0] {
			if len(s) > 1 && '/' == s[1] {
				continue
			}
			s = "https://te.20minut.ua" + s
			// fmt.Println("ABC", s)
			spider.db.Exec("INSERT INTO link (id) VALUES ($1)", s)
			// if nil != err3 {
			// return err3
			// }
		}
	}
	return nil
}

func (spider Spider) Archive(url string) {
	spider.Fetch(url)
	r, err := spider.client.Get("http://web.archive.org/save/" + url)
	spider.db.Exec("UPDATE link SET archived = current_timestamp WHERE id = $1", url)
	if nil == err {
		fmt.Println(r.StatusCode, url)
	} else {
		fmt.Println("ARCHIVE ERROR: "+url, err)
		time.Sleep(8 * time.Second)
	}
}

func (spider Spider) Load() {
	for {
		row := spider.db.QueryRow("SELECT id FROM link WHERE archived IS NULL LIMIT 1")
		var url string
		err := row.Scan(&url)
		if nil == err {
			spider.Archive(url)
		} else if sql.ErrNoRows == err {
			fmt.Errorf("No rows\n")
			return
		}
	}
}

func main() {
	db, err := sql.Open("sqlite3", "parser.db")
	if nil != err {
		panic(err)
	}
	spider := Spider{db: db}
	err = spider.Init()
	if nil != err {
		panic(err)
	}
	err = spider.Fetch("https://te.20minut.ua/")
	if nil != err {
		panic(err)
	}
	spider.Load()
}
