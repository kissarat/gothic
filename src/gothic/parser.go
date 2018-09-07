package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

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
	db *sql.DB
}

func (spider Spider) fetch(url string) {
	r, err2 := http.Get(url)
	if nil != err2 {
		fmt.Println(err2)
		return
	}
	data, err3 := ioutil.ReadAll(r.Body)
	if nil != err3 {
		fmt.Println(err3)
		return
	}
	for _, m := range RegexHref.FindAllStringSubmatch(string(data), -1) {
		// var count int
		// row := spider.db.QueryRow("SELECT count(*) FROM link WHERE id in $1", m[1])
		// row.Scan(&count)
		// if 0 == count {
		s := m[1]
		if strings.HasPrefix(s, "http://rada.te.ua") {
			spider.db.Exec("INSERT INTO link (id) VALUES ($1)", s)
		}
		// }
	}
}

func (spider Spider) archive(ch chan string) {
	for {
		url, more := <-ch
		if more {
			// time.Sleep(100 * time.Millisecond)
			spider.fetch(url)
			r, err := http.Get("http://web.archive.org/save/" + url)
			if nil != err {
				fmt.Println(err)
			}
			spider.db.Exec("UPDATE link SET archived = current_timestamp WHERE id = $1", url)
			fmt.Println(r.StatusCode, url)
			spider.load(ch)
		} else {
			break
		}
	}
}

func (spider Spider) load(ch chan string) {
	rows, err2 := spider.db.Query("SELECT id FROM link WHERE archived IS NULL ORDER BY created")
	if nil != err2 {
		panic(err2)
	}
	// for i := 0; i < 10; i++ {
	go spider.archive(ch)
	// }
	for {
		if rows.Next() {
			var s string
			rows.Scan(&s)
			ch <- s
		} else {
			// close(ch)
			fmt.Println("Channel closed")
			break
		}
	}
}

func main() {
	db, err1 := sql.Open("sqlite3", "parser.db")
	if nil != err1 {
		panic(err1)
	}
	defer db.Close()
	var count int
	db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name = 'link'`, &count)
	if 0 == count {
		db.Exec(schema)
	}

	spider := Spider{db: db}
	spider.fetch("http://rada.te.ua/")
	ch := make(chan string, 1)
	spider.load(ch)
}
