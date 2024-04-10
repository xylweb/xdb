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
	db.Range(func(k int, v any) bool {
		fmt.Println(k, v)
		return false
	})
	db.RangeIndex("desc", 10, func(k int, v any) bool {
		fmt.Println(k, v)
		return false
	})
	fmt.Println(d, e, time.Now().Sub(st))
	for i := 0; i < 20; i++ {
		fmt.Println(db.Refresh())
	}
	time.Sleep(10 * time.Second)
}
