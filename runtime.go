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
}
type DType interface {
	~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64 | ~float32 | ~float64 | ~string
}

type Ibase[T DType] []T
type Dbase[T DType] map[T]any
type Xdbase[T DType] struct {
	DPath   string
	Path    string
	IsIndex bool
	IData   Ibase[T]
	Data    Dbase[T]
	Chan    chan time.Time
	Lock    sync.RWMutex
}

func NewXdb[T DType]() *Xdbase[T] {
	return new(Xdbase[T])
}
func (this *Xdbase[T]) SetParams(conf Config) *Xdbase[T] {
	this.DPath = conf.DbPath
	this.Path = filepath.Join(conf.DbPath, conf.DbName)
	this.IsIndex = conf.IsIndex
	this.Data = make(Dbase[T])
	this.Chan = make(chan time.Time, 10)
	this.Lock = sync.RWMutex{}
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
	if len(this.IData) == 0 {
		this.IData = append(this.IData, key)
		return
	}
	min := this.IData[0]
	max := this.IData[len(this.IData)-1]
	if key <= min {
		var tmp = []T{key}
		tmp = append(tmp, this.IData...)
		this.IData = tmp
	}
	if key >= max {
		this.IData = append(this.IData, key)
	}
}
func (this *Xdbase[T]) Add(key T, val any) bool {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	if _, ok := this.Data[key]; !ok {
		this.sort(key)
	}
	this.Data[key] = val
	this.setChan()
	return true
}
func (this *Xdbase[T]) Get(key T) (any, bool) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	val, ok := this.Data[key]
	return val, ok
}
func (this *Xdbase[T]) Del(key T) bool {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	delete(this.Data, key)
	this.setChan()
	return true
}
func (this *Xdbase[T]) Count() int {
	return len(this.Data)
}
func (this *Xdbase[T]) OrderKey(order string, limit int) []T {
	if len(this.IData) <= limit {
		return this.IData
	}
	switch order {
	case "asc":
		return this.IData[:limit]
	case "desc":
		return this.IData[len(this.IData)-limit:]
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
					if len(this.IData) > 0 {
						this.toFile(this.getIdataPath(), this.getITmpPath(), this.IData)
					}
					if len(this.Data) > 0 {
						this.toFile(this.getDataPath(), this.getDTmpPath(), this.Data)
					}
					this.ClearCh()
				} else {
					this.Chan <- times
				}
			default:
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
	this.Lock.Lock()
	data, err := msgpack.Marshal(d)
	this.Lock.Unlock()
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
	ibase := make(Ibase[T], 0)
	err = msgpack.Unmarshal(data, &ibase)
	if err != nil {
		return false
	}
	this.IData = ibase
	return true
}

func (this *Xdbase[T]) fromDFile() bool {
	data, err := ioutil.ReadFile(this.getDataPath())
	if err != nil {
		return false
	}
	dbase := make(Dbase[T], 0)
	err = msgpack.Unmarshal(data, &dbase)
	if err != nil {
		return false
	}
	this.Data = dbase
	return true
}

func (this *Xdbase[T]) Save() bool {
	return this.toFile(this.getIdataPath(), this.getITmpPath(), this.IData) &&
		this.toFile(this.getDataPath(), this.getDTmpPath(), this.Data)
}
