package matching

import (
	"errors" 
	"fmt"
)

var (
	tA = [8]byte{1, 2, 4, 8, 16, 32, 64, 128} 
	tB = [8]byte{254, 253, 251, 247, 239, 223, 191, 127}
)

func dataOrCopy(d []byte, c bool) []byte{
	if !c{
		return d 
	}
	data := make([]byte, len(d)) 
	copy(data, d) 
	return data 
}

func NewSlice(l int64) []byte{
	remainder := l % 8 
	if remainder != 0{
		remainder = 1 
	}
	return make([]byte, l/8+remainder)
}

func Get(m []byte, i int64) bool{
	return m[i/8]&tA[i%8] != 0 
}

func Set(m []byte, i int64, v bool){
	index := i/8 
	bit := i % 8 
	if v{
		m[index] = m[index] | tA[bit] 
	} else{
		m[index] = m[index] & tB[bit]
	}
}

func GetBit(b byte, i int64) bool{
	return b&tA[i] != 0 
}

func SetBit(b byte, i int64, v bool) byte{
	if v{
		return b | tA[i] 
	}
	return b & tB[i]
}

func Len(m []byte) int{
	return len(m) * 8 
}

type Bitmap []byte 

func New(l int64) Bitmap{
	return NewSlice(l)
}

func (b Bitmap) Len() int{
	return Len(b)
}

func (b Bitmap) Get(i int64) bool{
	return Get(b, i)
}

func (b Bitmap) Set(i int64, v bool){
	Set(b, i, v)
}

func (b Bitmap) Data(copy bool) []byte{
	return dataOrCopy(b, copy)
}

type Window struct{
	Min int64 
	Max int64 
	Cap int64 
	Bitmap Bitmap 
}

func newWindow(min, max int64) Window{
	return Window{
		Min: min, 
		Max: max, 
		Cap: max - min, 
		Bitmap: New(max - min),
	}
}

func (w Window) put(val int64) error{
	if val <= w.Min{
		return errors.New(fmt.Sprintf("expired val %v, current Window [%v-%v]", val, w.Min, w.Max))
	} else if val > w.Max{
		delta := val - w.Max 
		w.Min += delta 
		w.Max += delta 
		w.Bitmap.Set(val%w.Cap, true)
	} else if w.Bitmap.Get(val%w.Cap){
		return errors.New(fmt.Sprintf("existed val %v", val))
	} else{
		w.Bitmap.Set(val%w.Cap, true)
	}
	return nil 
}

func (w Window) contains(val int64) bool{
	return false 
}