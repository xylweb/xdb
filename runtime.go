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

type Ibase []string
type Dbase map[string]interface{}
type Xdbase struct {
	Path    string
	IsIndex bool
	IData   Ibase
	Data    Dbase
	Chan    chan time.Time
	Lock    sync.RWMutex
}

func NewXdb(path, dname string, index bool) *Xdbase {
	os.MkdirAll(path, 0777)
	return &Xdbase{
		Path:    filepath.Join(path, dname),
		IsIndex: index,
		Data:    make(map[string]interface{}),
		Chan:    make(chan time.Time, 10),
		Lock:    sync.RWMutex{},
	}
}
func (this *Xdbase) Open() {
	this.fromIFile()
	this.fromDFile()
	this.run()
}

//index path
func (this *Xdbase) getIdataPath() string {
	return this.Path + subffixindex
}

//data path
func (this *Xdbase) getPath() string {
	return this.Path
}
func (this *Xdbase) getDataPath() string {
	return this.Path + subffix
}
func (this *Xdbase) getITmpPath() string {
	return this.Path + subffixindex + subffixtmp
}
func (this *Xdbase) getDTmpPath() string {
	return this.Path + subffix + subffixtmp
}
func (this *Xdbase) sort(key string) {
	if !this.IsIndex {
		return
	}
	if len(this.IData) == 0 {
		this.IData = append(this.IData, key)
		return
	}
	min := this.IData[0]
	max := this.IData[len(this.IData)-1]
	if len(key) < len(max) {
		var tmp = []string{key}
		tmp = append(tmp, this.IData...)
		this.IData = tmp
	}
	if len(key) == len(min) && key < min {
		var tmp = []string{key}
		tmp = append(tmp, this.IData...)
		this.IData = tmp
	}
	if len(key) == len(max) && key > max {
		this.IData = append(this.IData, key)
	}
	if len(key) > len(max) {
		this.IData = append(this.IData, key)
	}
}
func (this *Xdbase) Add(key string, val interface{}) bool {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	if _, ok := this.Data[key]; !ok {
		this.sort(key)
	}
	this.Data[key] = val
	this.setChan()
	return true
}
func (this *Xdbase) Get(key string) (interface{}, bool) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	val, ok := this.Data[key]
	return val, ok
}
func (this *Xdbase) Del(key string) bool {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	delete(this.Data, key)
	this.setChan()
	return true
}
func (this *Xdbase) Count() int {
	return len(this.Data)
}
func (this *Xdbase) OrderKey(order string, limit int) []string {
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
func (this *Xdbase) setChan() {
	if len(this.Chan) > 0 {
		return
	}
	select {
	case this.Chan <- time.Now():
	default:
	}
}
func (this *Xdbase) run() {
	go func(this *Xdbase) {
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
func (this *Xdbase) ClearCh() {
	for i := 0; i <= len(this.Chan); i++ {
		<-this.Chan
	}
}
func (this *Xdbase) toFile(path, tmppath string, d interface{}) bool {
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
func (this *Xdbase) fromIFile() bool {
	data, err := ioutil.ReadFile(this.getIdataPath())
	if err != nil {
		return false
	}
	ibase := make(Ibase, 0)
	err = msgpack.Unmarshal(data, &ibase)
	if err != nil {
		return false
	}
	this.IData = ibase
	return true
}

func (this *Xdbase) fromDFile() bool {
	data, err := ioutil.ReadFile(this.getDataPath())
	if err != nil {
		return false
	}
	dbase := make(Dbase, 0)
	err = msgpack.Unmarshal(data, &dbase)
	if err != nil {
		return false
	}
	this.Data = dbase
	return true
}

func (this *Xdbase) Save() bool {
	return this.toFile(this.getIdataPath(), this.getITmpPath(), this.IData) &&
		this.toFile(this.getDataPath(), this.getDTmpPath(), this.Data)
}
