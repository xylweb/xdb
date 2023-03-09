package test

import (
	"testing"

	"github.com/xylweb/xdb"
)

func TestAdd(t *testing.T) {
	db := xdb.NewXdb[int64]()
	db.SetParams(xdb.Config{DbPath: "./data/", DbName: "base", IsIndex: true}).Open()
	for i := 0; i < 10000; i++ {
		db.Add(int64(i), i)
	}
	t.Log("ok")
}
func BenchmarkAdd(t *testing.B) {
	db := xdb.NewXdb[int64]()
	db.SetParams(xdb.Config{DbPath: "./data/", DbName: "base", IsIndex: true}).Open()
	for i := 0; i < t.N; i++ {
		db.Get(int64(i))

	}
	db.Close()
}
