package main

import (
	"fmt"
	"time"

	"github.com/xylweb/xdb"
)

func main() {
	xdb.Init(xdb.Config{DbPath: "./", DbName: "base", Index: true})
	st := time.Now()
	for i := 0; i <= 10; i++ {
		xdb.Insert(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
	d, e := xdb.Find("key10")
	xdb.Del("01")
	keys := xdb.OrderKey("desc", 5)
	fmt.Println(keys)
	fmt.Println(d, e, time.Now().Sub(st))
	time.Sleep(10 * time.Second)
}
