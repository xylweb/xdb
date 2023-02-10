package xdb

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"
)

var (
	cnum    = 100
	subffix = ".dat"
)

type Dbase map[string]interface{}
type Xdbase struct {
	Path string
	Data Dbase
	Chan chan time.Time
	Lock sync.RWMutex
}

func NewXdb(path, dname string) *Xdbase {
	return &Xdbase{Path: filepath.Join(path, dname+subffix), Data: make(map[string]interface{}), Chan: make(chan time.Time, cnum), Lock: sync.RWMutex{}}
}
func (this *Xdbase) Init() {
	this.fromFile()
	this.run()
}
func (this *Xdbase) Add(key string, val interface{}) bool {
	this.Lock.Lock()
	defer this.Lock.Unlock()
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
func (this *Xdbase) setChan() {
	select {
	case this.Chan <- time.Now():
	default:
	}
}
func (this *Xdbase) run() {
	go func(this *Xdbase) {
		now := time.Now()
		savetime := time.Now()
		savebool := false
		for {
			select {
			case tm := <-this.Chan:
				savetime = time.Now()
				savebool = true
				if tm.Unix() > now.Unix() {
					this.toFile()
					now = tm
					savebool = false
				}
			default:
				if time.Now().Unix() > savetime.Unix() && savebool {
					this.toFile()
					savebool = false
				}
			}
		}
	}(this)
}
func (this *Xdbase) toFile() bool {
	this.Lock.Lock()
	data, err := json.Marshal(this.Data)
	this.Lock.Unlock()
	if err != nil {
		return false
	}
	err = ioutil.WriteFile(this.Path, data, 0777)
	if err != nil {
		return false
	}
	return true
}
func (this *Xdbase) fromFile() bool {
	data, err := ioutil.ReadFile(this.Path)
	if err != nil {
		return false
	}
	dbase := make(Dbase)
	err = json.Unmarshal(data, &dbase)
	if err != nil {
		return false
	}
	this.Data = dbase
	return true
}

func (this *Xdbase) Save() bool {
	return this.toFile()
}
