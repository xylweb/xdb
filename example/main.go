package main

import (
	"fmt"
	"time"

	"github.com/xylweb/xdb"
)

func main() {
	db := xdb.NewXdb[int]()
	db.SetParams(xdb.Config{DbPath: "./data/", DbName: "base", IsIndex: true, Pass: "123456"}).Open()
	st := time.Now()
	for i := 0; i <= 1000000; i++ {
		db.Add(i, fmt.Sprintf("value%d", i))
	}
	d, e := db.Get(1000000)
	db.Del(999998)
	keys := db.OrderKey("desc", 5)
	fmt.Println(keys)
	for _, v := range keys {
		val, _ := db.Get(v)
		fmt.Println(v, val)
	}
	fmt.Println(d, e, time.Now().Sub(st))
	time.Sleep(10 * time.Second)
}
