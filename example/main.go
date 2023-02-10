package main

import (
	"fmt"
	"time"

	"github.com/xylweb/xdb"
)

func main() {
	xdb.Init(xdb.Config{DbPath: "./", DbName: "base"})
	st := time.Now()
	for i := 0; i <= 100000; i++ {
		xdb.Insert(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
	d, e := xdb.Find("key100000")
	xdb.Del("key01")
	fmt.Println(d, e, time.Now().Sub(st))
	time.Sleep(10 * time.Second)
}
