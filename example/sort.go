package main

import (
	"fmt"
	"math/rand"
	"time"
)

type DType interface {
	int | int32 | int64 | uint | uint32 | uint64
}
type Sort[T DType] struct {
	LNum  int
	LData []T
	RNum  int
	RData []T
}

func (this *Sort[T]) Insert(d T) {
	this.FindAsc(0, len(this.LData), d)
	this.FindDesc(0, len(this.RData), d)
}
func (this *Sort[T]) FindAsc(left, right int, d T) {
	lens := len(this.LData)
	if len(this.LData) == 0 {
		this.LData = append(this.LData, d)
		return
	}
	if d <= this.LData[0] {
		tmp := []T{d}
		tmp = append(tmp, this.LData...)
		this.LData = tmp
		return
	}
	if d >= this.LData[lens-1] {
		this.LData = append(this.LData, d)
		return
	}
	if left > right {
		jk := left
		tmp := make([]T, lens+1)
		copy(tmp[:jk], this.LData[:jk])
		copy(tmp[jk:jk+1], []T{d})
		copy(tmp[jk+1:], this.LData[jk:])
		if len(tmp) > this.LNum {
			this.LData = tmp[:this.LNum]
		} else {
			this.LData = tmp
		}
		return
	}
	mid := (left + right) / 2
	if this.LData[mid] > d {
		this.FindAsc(left, mid-1, d)
	} else if d > this.LData[mid] {
		this.FindAsc(mid+1, right, d)
	} else {
		jk := mid
		tmp := make([]T, lens+1)
		copy(tmp[:jk], this.LData[:jk])
		copy(tmp[jk:jk+1], []T{d})
		copy(tmp[jk+1:], this.LData[jk:])
		if len(tmp) > this.LNum {
			this.LData = tmp[:this.LNum]
		} else {
			this.LData = tmp
		}
		return
	}
}
func (this *Sort[T]) FindDesc(left, right int, d T) {
	lens := len(this.RData)
	if len(this.RData) == 0 {
		this.RData = append(this.RData, d)
		return
	}
	if d <= this.RData[0] {
		tmp := []T{d}
		tmp = append(tmp, this.RData...)
		this.RData = tmp
		return
	}
	if d >= this.RData[lens-1] {
		this.RData = append(this.RData, d)
		return
	}
	if left > right {
		jk := left
		tmp := make([]T, lens+1)
		copy(tmp[:jk], this.RData[:jk])
		copy(tmp[jk:jk+1], []T{d})
		copy(tmp[jk+1:], this.RData[jk:])
		if len(tmp) > this.RNum {
			this.RData = tmp[lens-this.RNum:]
		} else {
			this.RData = tmp
		}
		return
	}
	mid := (left + right) / 2
	if this.RData[mid] > d {
		this.FindDesc(left, mid-1, d)
	} else if d > this.RData[mid] {
		this.FindDesc(mid+1, right, d)
	} else {
		jk := mid
		tmp := make([]T, lens+1)
		copy(tmp[:jk], this.RData[:jk])
		copy(tmp[jk:jk+1], []T{d})
		copy(tmp[jk+1:], this.RData[jk:])
		if len(tmp) > this.RNum {
			this.RData = tmp[lens-this.RNum:]
		} else {
			this.RData = tmp
		}
		return
	}
}
func main() {
	s := time.Now()
	sk := &Sort[int]{LNum: 500, RNum: 500}
	for i := 0; i < 1000000; i++ {
		d := rand.Int()
		//fmt.Println("============", i, d)
		sk.Insert(d)
	}
	fmt.Println("result", len(sk.LData), len(sk.RData), time.Now().Sub(s))
	fmt.Println("left", sk.LData)
	fmt.Println("right", sk.RData)

}
