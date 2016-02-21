package rardecode

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"io"
)

const (
	fileSize = 0x1000000

	vmGlobalAddr      = 0x3C000
	vmGlobalSize      = 0x02000
	vmFixedGlobalSize = 0x40
)

// v3Filter is the interface type for RAR V3 filters.
// v3Filter performs the same function as the filter type, except that it also takes
// the initial register values r, and global data as input for the RAR V3 VM.
type v3Filter func(r map[int]uint32, global, buf []byte, offset int64) ([]byte, error)

var (
	// standardV3Filters is a list of known filters. We can replace the use of a vm
	// filter with a custom filter function.
	standardV3Filters = []struct {
		crc uint32   // crc of code byte slice for filter
		len int      // length of code byte slice for filter
		f   v3Filter // replacement filter function
	}{
		{0xad576887, 53, e8FilterV3},
		{0x3cd7e57e, 57, e8e9FilterV3},
		{0x0e06077d, 29, deltaFilterV3},
	}
)

func filterE8(c byte, v5 bool, buf []byte, offset int64) ([]byte, error) {
	off := int32(offset)
	for b := buf; len(b) >= 5; {
		ch := b[0]
		b = b[1:]
		off++
		if ch != 0xe8 && ch != c {
			continue
		}
		if v5 {
			off %= fileSize
		}
		addr := int32(binary.LittleEndian.Uint32(b))
		if addr < 0 {
			if addr+off >= 0 {
				binary.LittleEndian.PutUint32(b, uint32(addr+fileSize))
			}
		} else if addr < fileSize {
			binary.LittleEndian.PutUint32(b, uint32(addr-off))
		}
		off += 4
		b = b[4:]
	}
	return buf, nil
}

func e8FilterV3(r map[int]uint32, global, buf []byte, offset int64) ([]byte, error) {
	return filterE8(0xe8, false, buf, offset)
}

func e8e9FilterV3(r map[int]uint32, global, buf []byte, offset int64) ([]byte, error) {
	return filterE8(0xe9, false, buf, offset)
}

func filterDelta(n int, buf []byte) ([]byte, error) {
	var res []byte
	l := len(buf)
	if cap(buf) >= 2*l {
		res = buf[l : 2*l] // use unused capacity
	} else {
		res = make([]byte, l, 2*l)
	}

	i := 0
	for j := 0; j < n; j++ {
		var c byte
		for k := j; k < len(res); k += n {
			c -= buf[i]
			i++
			res[k] = c
		}
	}
	return res, nil
}

func deltaFilterV3(r map[int]uint32, global, buf []byte, offset int64) ([]byte, error) {
	return filterDelta(int(r[0]), buf)
}

func filterArm(buf []byte, offset int64) ([]byte, error) {
	for i := 0; len(buf)-i > 3; i += 4 {
		if buf[i+3] == 0xeb {
			n := uint(buf[i])
			n += uint(buf[i+1]) * 0x100
			n += uint(buf[i+2]) * 0x10000
			n -= (uint(offset) + uint(i)) / 4
			buf[i] = byte(n)
			buf[i+1] = byte(n >> 8)
			buf[i+2] = byte(n >> 16)
		}
	}
	return buf, nil
}

type vmFilter struct {
	execCount uint32
	global    []byte
	static    []byte
	code      []command
}

// execute implements v3filter type for VM based RAR 3 filters.
func (f *vmFilter) execute(r map[int]uint32, global, buf []byte, offset int64) ([]byte, error) {
	if len(buf) > vmGlobalAddr {
		return buf, errInvalidFilter
	}
	v := newVM(buf)

	// register setup
	v.r[3] = vmGlobalAddr
	v.r[4] = uint32(len(buf))
	v.r[5] = f.execCount
	for i, n := range r {
		v.r[i] = n
	}

	// vm global data memory block
	vg := v.m[vmGlobalAddr : vmGlobalAddr+vmGlobalSize]

	// initialize fixed global memory
	for i, n := range v.r[:vmRegs-1] {
		binary.LittleEndian.PutUint32(vg[i*4:], n)
	}
	binary.LittleEndian.PutUint32(vg[0x1c:], uint32(len(buf)))
	binary.LittleEndian.PutUint64(vg[0x24:], uint64(offset))
	binary.LittleEndian.PutUint32(vg[0x2c:], f.execCount)

	// registers
	v.r[6] = uint32(offset)

	// copy program global memory
	var n int
	if len(f.global) > 0 {
		n = copy(vg[vmFixedGlobalSize:], f.global) // use saved global instead
	} else {
		n = copy(vg[vmFixedGlobalSize:], global)
	}
	copy(vg[vmFixedGlobalSize+n:], f.static)

	v.execute(f.code)

	f.execCount++

	// keep largest global buffer
	if cap(global) > cap(f.global) {
		f.global = global[:0]
	} else if len(f.global) > 0 {
		f.global = f.global[:0]
	}

	// check for global data to be saved for next program execution
	globalSize := binary.LittleEndian.Uint32(vg[0x30:])
	if globalSize > 0 {
		if globalSize > vmGlobalSize-vmFixedGlobalSize {
			globalSize = vmGlobalSize - vmFixedGlobalSize
		}
		if cap(f.global) < int(globalSize) {
			f.global = make([]byte, globalSize)
		} else {
			f.global = f.global[:globalSize]
		}
		copy(f.global, vg[vmFixedGlobalSize:])
	}

	// find program output
	length := binary.LittleEndian.Uint32(vg[0x1c:]) & vmMask
	start := binary.LittleEndian.Uint32(vg[0x20:]) & vmMask
	if start+length > vmSize {
		// TODO: error
		start = 0
		length = 0
	}
	if start != 0 && cap(v.m) > cap(buf) {
		// Initial buffer was to small for vm.
		// Copy output to beginning of vm memory so that decodeReader
		// will re-use the newly allocated vm memory and we will not
		// have to reallocate again next time.
		copy(v.m, v.m[start:start+length])
		start = 0
	}
	return v.m[start : start+length], nil
}

// getV3Filter returns a V3 filter function from a code byte slice.
func getV3Filter(code []byte) (v3Filter, error) {
	// check if filter is a known standard filter
	c := crc32.ChecksumIEEE(code)
	for _, f := range standardV3Filters {
		if f.crc == c && f.len == len(code) {
			return f.f, nil
		}
	}

	// create new vm filter
	f := new(vmFilter)
	r := newRarBitReader(bytes.NewReader(code[1:])) // skip first xor byte check

	// read static data
	n, err := r.readBits(1)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		m, err := r.readUint32()
		if err != nil {
			return nil, err
		}
		f.static = make([]byte, m+1)
		err = r.readFull(f.static)
		if err != nil {
			return nil, err
		}
	}

	f.code, err = readCommands(r)
	if err == io.EOF {
		err = nil
	}

	return f.execute, err
}
