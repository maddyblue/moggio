package ogg

import (
	"io"
)

type page struct {
	header        pageHeader
	needsContinue bool
	packetCount   int
	packets       []packetInfo
	content       []byte
	lastPageRest  []byte
}

type packetInfo struct {
	offset, size int
}

func (p *page) ReadFrom(r io.Reader) error {
	err := p.header.ReadFrom(r)
	if err != nil {
		return err
	}
	p.packetCount = 0
	for _, entry := range p.header.SegmentTable {
		if entry < 0xFF {
			p.packetCount++
		}
	}
	p.needsContinue = p.header.SegmentTable[len(p.header.SegmentTable)-1] == 0xFF
	p.packets = make([]packetInfo, p.packetCount+1)
	totalSize := 0
	i := 0
	for _, entry := range p.header.SegmentTable {
		totalSize += int(entry)
		p.packets[i].size += int(entry)
		if entry < 0xFF {
			i++
			p.packets[i].offset = totalSize
		}
	}
	p.content = make([]byte, totalSize)
	_, err = io.ReadFull(r, p.content)
	if err != nil {
		return err
	}
	return nil
}

// Prepend adds the last, unfinished packet from the last page to the start of this page.
// Returns an error if the argument is not an empty slice and the header indicates that the last page had no unfinished packet.
func (p *page) Prepend(last []byte) error {
	if len(last) > 0 && !p.header.IsContinue() {
		return ErrCorruptStream
	}
	p.lastPageRest = last
	return nil
}

func (p *page) PacketCount() int {
	return p.packetCount
}

// Packet returns the content of the packet at the given index.
// Only finished packets will be returned.
func (p *page) Packet(i int) []byte {
	if i == 0 && p.header.IsContinue() {
		return append(p.lastPageRest, p.content[:p.packets[0].size]...)
	}
	if i >= p.packetCount {
		panic("index out of bounds")
	}
	info := p.packets[i]
	return p.content[info.offset : info.offset+info.size]
}

// Rest returns the last packet on the page if it is unfinished, or an empty slice.
// Intended to be used as the argument to the next page's Prepend method.
func (p *page) Rest() []byte {
	if p.packetCount == 0 {
		// a packet spans multiple pages
		return append(p.lastPageRest, p.content...)
	}
	last := p.packets[p.packetCount-1]
	return p.content[last.offset+last.size:]
}
