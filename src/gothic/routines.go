package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var s chan bool

type T struct {
	i     int
	array []string
}

func (t T) dump(name string) {
	fmt.Println(strconv.Itoa(t.i) + " " + name + " " + strings.Join(t.array, " "))
}

func (t T) first() {
	time.Sleep(8 * time.Second)
	t.dump("first ")
	s <- true
}

func (t T) second() {
	time.Sleep(2 * time.Second)
	t.array[0] = "c"
	t.dump("second")
	// s <- true
}

func (t T) third() {
	time.Sleep(4 * time.Second)
	t.array[0] = "d"
	t.dump("third ")
}

func (t T) fourth() {
	ticker := time.NewTicker(time.Second)
	for {
		t.i++
		t.array = append(t.array, strconv.Itoa(t.i))
		t.dump("fourth")
		t.array[0] = "b"
		<-ticker.C
	}
}

func main() {
	s = make(chan bool, 1)
	t := T{}
	t.i = 0
	t.array = []string{"a"}
	go t.first()
	go t.second()
	go t.third()
	go t.fourth()
	<-s
}
