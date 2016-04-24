package ogg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

var pageHeaderPattern = [4]byte{'O', 'g', 'g', 'S'}

var ErrCorruptStream = errors.New("ogg: corrupt stream")
var ErrChecksum = errors.New("ogg: wrong checksum")

const (
	headerFlagContinuedPacket   = 1
	headerFlagBeginningOfStream = 2
	headerFlagEndOfStream       = 4
)

type pageHeaderStart struct {
	CapturePattern          [4]byte
	StreamStructureVersion  uint8
	HeaderTypeFlag          byte
	AbsoluteGranulePosition uint64
	StreamSerialNumber      uint32
	PageSequenceNumber      uint32
	PageChecksum            uint32
	PageSegments            uint8
}

type pageHeader struct {
	pageHeaderStart
	SegmentTable   []uint8
	headerChecksum uint32
}

func (h *pageHeader) ReadFrom(r io.Reader) error {
	data := make([]byte, 27)
	_, err := io.ReadFull(r, data)
	if err != nil {
		return err
	}
	binary.Read(bytes.NewReader(data), binary.LittleEndian, &h.pageHeaderStart)
	if h.CapturePattern != pageHeaderPattern {
		return ErrCorruptStream
	}
	h.SegmentTable = make([]byte, h.PageSegments)
	_, err = io.ReadFull(r, h.SegmentTable)
	if err != nil {
		return err
	}
	data[22], data[23], data[24], data[25] = 0, 0, 0, 0
	h.headerChecksum = crcUpdate(0, data)
	h.headerChecksum = crcUpdate(h.headerChecksum, h.SegmentTable)
	return nil
}

func (h *pageHeader) IsFirstPage() bool { return h.HeaderTypeFlag&headerFlagBeginningOfStream != 0 }
func (h *pageHeader) IsLastPage() bool  { return h.HeaderTypeFlag&headerFlagEndOfStream != 0 }
func (h *pageHeader) IsContinue() bool  { return h.HeaderTypeFlag&headerFlagContinuedPacket != 0 }
