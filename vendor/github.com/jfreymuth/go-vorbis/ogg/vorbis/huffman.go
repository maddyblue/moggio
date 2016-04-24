package vorbis

import (
	"github.com/jfreymuth/go-vorbis/ogg"
)

type huffmanTable []uint32

func (t huffmanTable) Lookup(r *ogg.BitReader) uint32 {
	i := uint32(0)
	for i&0x80000000 == 0 {
		i = t[i*2+r.Read1()]
	}
	return i & 0x7FFFFFFF
}

func (t huffmanTable) Put(entry uint32, length uint8) {
	t.put(0, entry, length-1)
}

func (t huffmanTable) put(index, entry uint32, length uint8) bool {
	if length == 0 {
		if t[index*2] == 0 {
			t[index*2] = entry | 0x80000000
			return true
		} else if t[index*2+1] == 0 {
			t[index*2+1] = entry | 0x80000000
			return true
		}
		return false
	} else {
		if t[index*2]&0x80000000 == 0 {
			if t[index*2] == 0 {
				t[index*2] = t.findEmpty(index + 1)
			}
			if t.put(t[index*2], entry, length-1) {
				return true
			}
		}
		if t[index*2+1]&0x80000000 == 0 {
			if t[index*2+1] == 0 {
				t[index*2+1] = t.findEmpty(index + 1)
			}
			if t.put(t[index*2+1], entry, length-1) {
				return true
			}
		}
		return false
	}
}

func (t huffmanTable) findEmpty(index uint32) uint32 {
	for t[index*2] != 0 {
		index++
	}
	return index
}
