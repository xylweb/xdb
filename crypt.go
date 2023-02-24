package xdb

import (
	"crypto/sha256"
	"math"
	"sync"

	"golang.org/x/crypto/hkdf"
)

const (
	page     = 10
	bytesize = 256
)

type Crypt struct {
	Pass     string
	passbyte [bytesize]byte
	passmap  map[uint8]int
	PerPage  int
	page     int
	lock     *sync.RWMutex
	wg       *sync.WaitGroup
}

func (this *Crypt) SetParams(pass string) {
	this.lock = &sync.RWMutex{}
	this.wg = &sync.WaitGroup{}
	this.Pass = pass
	this.page = page
}
func (this *Crypt) CPass(lens int) bool {
	if lens < bytesize {
		lens = bytesize
	}
	hash := sha256.New
	hkdf := hkdf.New(hash, []byte(this.Pass), nil, nil)
	p := make([]byte, lens)
	hkdf.Read(p)
	ps := [bytesize]byte{}
	mp := make(map[uint8]int, 0)
	i := 0
	for _, v := range p {
		if _, ok := mp[v]; !ok {
			mp[v] = i
			ps[i] = uint8(v)
			i++
		}
		if len(mp) > bytesize {
			break
		}
	}
	if len(mp) < bytesize {
		return this.CPass(lens + bytesize)
	}
	this.passbyte = ps
	this.passmap = mp
	return true
}
func (this *Crypt) Encode(d []byte) []byte {
	lens := len(d)
	this.PerPage = int(math.Ceil(float64(lens) / float64(this.page)))
	rs := make([]byte, lens)
	for i := 0; i < this.page; i++ {
		jkstart := i * this.PerPage
		if jkstart > lens {
			continue
		}
		jkend := (i + 1) * this.PerPage
		if jkend > lens {
			jkend = lens
		}
		this.wg.Add(1)
		go func(jks, jke int, p [bytesize]byte, d []byte, lock *sync.RWMutex, wg *sync.WaitGroup) {
			lock.Lock()
			defer lock.Unlock()
			rstmp := make([]byte, 0)
			for _, val := range d[jks:jke] {
				rstmp = append(rstmp, p[val])
			}
			copy(rs[jks:jke], rstmp)
			this.wg.Done()
		}(jkstart, jkend, this.passbyte, d, this.lock, this.wg)
	}
	this.wg.Wait()
	return rs
}
func (this *Crypt) Decode(d []byte) []byte {
	lens := len(d)
	this.PerPage = int(math.Ceil(float64(lens) / float64(this.page)))
	rs := make([]byte, lens)
	for i := 0; i < this.page; i++ {
		jkstart := i * this.PerPage
		if jkstart > lens {
			continue
		}
		jkend := (i + 1) * this.PerPage
		if jkend > lens {
			jkend = lens
		}
		this.wg.Add(1)
		go func(jks, jke int, p map[uint8]int, d []byte, lock *sync.RWMutex, wg *sync.WaitGroup) {
			lock.Lock()
			defer lock.Unlock()
			rstmp := make([]byte, 0)
			for _, val := range d[jks:jke] {
				k := p[val]
				rstmp = append(rstmp, uint8(k))
			}
			copy(rs[jks:jke], rstmp)
			wg.Done()
		}(jkstart, jkend, this.passmap, d, this.lock, this.wg)
	}
	this.wg.Wait()
	return rs
}
