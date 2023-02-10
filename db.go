package xdb

var (
	Xdb *Xdbase
)

type Config struct {
	DbPath string
	DbName string
}

func Init(c Config) {
	Xdb = NewXdb(c.DbPath, c.DbName)
	Xdb.Init()
}
func Insert(key string, val interface{}) bool {
	Xdb.Add(key, val)
	return true
}
func Find(key string) (interface{}, bool) {
	val, ok := Xdb.Get(key)
	return val, ok
}
func Del(key string) bool {
	return Xdb.Del(key)
}
func Save() bool {
	return Xdb.Save()
}
