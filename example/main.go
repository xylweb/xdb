package main

import (
	"fmt"
	"time"

	"github.com/xylweb/xdb"
)

func main() {
	db := xdb.NewXdb[string]()
	db.SetParams(xdb.Config{DbPath: "./data/", DbName: "base", IsIndex: true}).Open()
	st := time.Now()
	for i := 0; i <= 1000000; i++ {
		db.Add(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
	d, e := db.Get("key1000000")
	db.Del("01")
	keys := db.OrderKey("desc", 5)
	fmt.Println(keys)
	fmt.Println(d, e, time.Now().Sub(st))
	time.Sleep(10 * time.Second)
}
