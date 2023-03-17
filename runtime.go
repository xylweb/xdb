package xdb

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	subffix      string  = ".xdb"
	subffixindex string  = ".idb"
	subffixtmp   string  = ".tmp"
	interval     float64 = 2
)

type Config struct {
	DbPath  string
	DbName  string
	IsIndex bool
	Pass    string
}
type DType interface {
	~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64 | ~float32 | ~float64 | ~string
}

type Ibase[T DType] []T
type Dbase[T DType] map[T]any
type Xdbase[T DType] struct {
	Pass      string
	DPath     string
	Path      string
	IsIndex   bool
	iData     Ibase[T]
	data      Dbase[T]
	Chan      chan time.Time
	lock      sync.RWMutex
	CloseChan chan bool
}

func NewXdb[T DType]() *Xdbase[T] {
	return new(Xdbase[T])
}
func (this *Xdbase[T]) SetParams(conf Config) *Xdbase[T] {
	this.Pass = conf.Pass
	this.DPath = conf.DbPath
	this.Path = filepath.Join(conf.DbPath, conf.DbName)
	this.IsIndex = conf.IsIndex
	this.data = make(Dbase[T])
	this.Chan = make(chan time.Time, 10)
	this.CloseChan = make(chan bool)
	this.lock = sync.RWMutex{}
	return this
}
func (this *Xdbase[T]) Open() *Xdbase[T] {
	os.MkdirAll(this.DPath, 0777)
	this.fromIFile()
	this.fromDFile()
	this.run()
	return this
}

//index path
func (this *Xdbase[T]) getIdataPath() string {
	return this.Path + subffixindex
}

//data path
func (this *Xdbase[T]) getPath() string {
	return this.Path
}
func (this *Xdbase[T]) getDataPath() string {
	return this.Path + subffix
}
func (this *Xdbase[T]) getITmpPath() string {
	return this.Path + subffixindex + subffixtmp
}
func (this *Xdbase[T]) getDTmpPath() string {
	return this.Path + subffix + subffixtmp
}
func (this *Xdbase[T]) sort(key T) {
	if !this.IsIndex {
		return
	}
	if len(this.iData) == 0 {
		this.iData = append(this.iData, key)
		return
	}
	min := this.iData[0]
	max := this.iData[len(this.iData)-1]
	if key <= min {
		var tmp = []T{key}
		tmp = append(tmp, this.iData...)
		this.iData = tmp
	}
	if key >= max {
		this.iData = append(this.iData, key)
	}
}
func (this *Xdbase[T]) Add(key T, val any) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.data[key]; !ok {
		this.sort(key)
	}
	this.data[key] = val
	this.setChan()
	return true
}
func (this *Xdbase[T]) Get(key T) (any, bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	val, ok := this.data[key]
	return val, ok
}
func (this *Xdbase[T]) Del(key T) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.data, key)
	if len(this.iData) > 0 {
		this.delIndex(0, len(this.iData), key)
	}
	this.setChan()
	return true
}
func (this *Xdbase[T]) Count() int {
	return len(this.data)
}
func (this *Xdbase[T]) Range(f func(k T, v any) bool) {
	for k, v := range this.data {
		rs := f(k, v)
		if !rs {
			return
		}
	}
}
func (this *Xdbase[T]) RangeIndex(order string, limit int, f func(k T, v any) bool) {
	var d Ibase[T]
	if len(this.iData) <= limit {
		d = this.iData
	} else if order == "asc" {
		d = this.iData[:limit]
	} else { //desc
		d = this.iData[len(this.iData)-limit:]
	}
	for _, vk := range d {
		v, ok := this.data[vk]
		if ok {
			rs := f(vk, v)
			if !rs {
				return
			}
		}
	}
}
func (this *Xdbase[T]) delIndex(left, right int, d T) {
	if left > right {
		return
	}
	mid := (left + right) / 2
	if this.iData[mid] > d {
		this.delIndex(left, mid-1, d)
	} else if d > this.iData[mid] {
		this.delIndex(mid+1, right, d)
	} else {
		tmp := make([]T, len(this.iData)-1)
		copy(tmp[:mid], this.iData[:mid])
		copy(tmp[mid:], this.iData[mid+1:])
		this.iData = tmp
		return
	}
}
func (this *Xdbase[T]) OrderKey(order string, limit int) []T {
	if len(this.iData) <= limit {
		return this.iData
	}
	switch order {
	case "asc":
		return this.iData[:limit]
	case "desc":
		return this.iData[len(this.iData)-limit:]
	}
	return nil
}
func (this *Xdbase[T]) setChan() {
	if len(this.Chan) > 0 {
		return
	}
	select {
	case this.Chan <- time.Now():
	default:
	}
}
func (this *Xdbase[T]) run() {
	go func(this *Xdbase[T]) {
		times := time.Now()
		for {
			select {
			case times = <-this.Chan:
				if time.Now().Sub(times).Seconds() > interval {
					if len(this.iData) > 0 {
						this.toFile(this.getIdataPath(), this.getITmpPath(), this.iData)
					}
					if len(this.data) > 0 {
						this.toFile(this.getDataPath(), this.getDTmpPath(), this.data)
					}
					this.ClearCh()
				} else {
					time.AfterFunc(time.Second, func() {
						this.Chan <- times
					})
				}
			case cls := <-this.CloseChan:
				if cls {
					return
				}
			}
		}
	}(this)
}
func (this *Xdbase[T]) ClearCh() {
	for i := 0; i <= len(this.Chan); i++ {
		<-this.Chan
	}
}
func (this *Xdbase[T]) toFile(path, tmppath string, d any) bool {
	this.lock.Lock()
	data, err := msgpack.Marshal(d)
	if this.Pass != "" {
		c := Crypt{}
		c.SetParams(this.Pass)
		c.CPass(256)
		data = c.Encode(data)
	}
	this.lock.Unlock()
	if err != nil {
		return false
	}
	err = ioutil.WriteFile(path, data, 0777)
	if err != nil {
		return false
	}
	err = os.Rename(tmppath, path)
	if err != nil {
		return false
	}
	return true
}
func (this *Xdbase[T]) fromIFile() bool {
	data, err := ioutil.ReadFile(this.getIdataPath())
	if err != nil {
		return false
	}
	if this.Pass != "" {
		c := Crypt{}
		c.SetParams(this.Pass)
		c.CPass(bytesize)
		data = c.Decode(data)
	}
	ibase := make(Ibase[T], 0)
	err = msgpack.Unmarshal(data, &ibase)
	if err != nil {
		return false
	}
	this.iData = ibase
	return true
}

func (this *Xdbase[T]) fromDFile() bool {
	data, err := ioutil.ReadFile(this.getDataPath())
	if err != nil {
		return false
	}
	if this.Pass != "" {
		c := Crypt{}
		c.SetParams(this.Pass)
		c.CPass(bytesize)
		data = c.Decode(data)
	}
	dbase := make(Dbase[T], 0)
	err = msgpack.Unmarshal(data, &dbase)
	if err != nil {
		return false
	}
	this.data = dbase
	return true
}

func (this *Xdbase[T]) Save() bool {
	return this.toFile(this.getIdataPath(), this.getITmpPath(), this.iData) &&
		this.toFile(this.getDataPath(), this.getDTmpPath(), this.data)
}

func (this *Xdbase[T]) Close() bool {
	this.CloseChan <- true
	return true
}
