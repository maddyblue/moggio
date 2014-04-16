package mp3

import (
	"io"
	"math"
)

type MP3 struct {
	b *bitReader

	syncword           uint16
	ID                 byte
	layer              Layer
	protection_bit     byte
	bitrate_index      byte
	sampling_frequency byte
	padding_bit        byte
	private_bit        byte
	mode               Mode
	mode_extension     byte
	copyright          byte
	original_home      byte
	emphasis           Emphasis
}

func New(r io.Reader) *MP3 {
	b := newBitReader(r)
	return &MP3{
		b: b,
	}
}

func (m *MP3) Sequence() {
	for {
		m.frame()
		break
	}
	return
}

func (m *MP3) frame() {
	m.header()
	m.error_check()
	m.audio_data()
}

func (m *MP3) header() {
	syncword := uint16(m.b.ReadBits64(12))
	for i := 0; syncword != 0xfff; i++ {
		syncword <<= 1
		syncword |= uint16(m.b.ReadBits64(1))
		println("mis sync", i)
	}
	m.syncword = syncword
	m.ID = byte(m.b.ReadBits64(1))
	m.layer = Layer(m.b.ReadBits64(2))
	m.protection_bit = byte(m.b.ReadBits64(1))
	m.bitrate_index = byte(m.b.ReadBits64(4))
	m.sampling_frequency = byte(m.b.ReadBits64(2))
	m.padding_bit = byte(m.b.ReadBits64(1))
	m.private_bit = byte(m.b.ReadBits64(1))
	m.mode = Mode(m.b.ReadBits64(2))
	m.mode_extension = byte(m.b.ReadBits64(2))
	m.copyright = byte(m.b.ReadBits64(1))
	m.original_home = byte(m.b.ReadBits64(1))
	m.emphasis = Emphasis(m.b.ReadBits64(2))
}

func (m *MP3) error_check() {
	if m.protection_bit == 0 {
		m.b.ReadBits64(16)
	}
}

func (m *MP3) audio_data() {
	if m.mode == ModeSingle {
		main_data_end := uint16(m.b.ReadBits64(9))
		m.b.ReadBits64(5) // private_bits
		scfsi := make([]byte, cblimit)
		var part2_3_length [2]uint16
		var big_values [2]uint16
		var global_gain [2]uint16
		var scalefac_compress [2]byte
		var blocksplit_flag [2]byte
		var block_type [2]byte
		var switch_point [2]byte
		var table_select [3][2]byte
		var subblock_gain [3][2]uint8
		var region_address1, region_address2 [2]byte
		var preflag, scalefac_scale, count1table_select [2]byte
		var scalefac [][2]uint8
		var scalefacw [][3][2]uint8
		for scfsi_band := 0; scfsi_band < 4; scfsi_band++ {
			scfsi[scfsi_band] = byte(m.b.ReadBits64(1))
		}
		for gr := 0; gr < 2; gr++ {
			part2_3_length[gr] = uint16(m.b.ReadBits64(12))
			big_values[gr] = uint16(m.b.ReadBits64(9))
			global_gain[gr] = uint16(m.b.ReadBits64(8))
			scalefac_compress[gr] = byte(m.b.ReadBits64(4))
			blocksplit_flag[gr] = byte(m.b.ReadBits64(1))
			if blocksplit_flag[gr] != 0 {
				block_type[gr] = byte(m.b.ReadBits64(2))
				switch_point[gr] = byte(m.b.ReadBits64(1))
				for region := 0; region < 2; region++ {
					table_select[region][gr] = byte(m.b.ReadBits64(5))
				}
				for window := 0; window < 3; window++ {
					subblock_gain[window][gr] = uint8(m.b.ReadBits64(3))
				}
			} else {
				for region := 0; region < 3; region++ {
					table_select[region][gr] = byte(m.b.ReadBits64(5))
				}
				region_address1[gr] = byte(m.b.ReadBits64(4))
				region_address2[gr] = byte(m.b.ReadBits64(3))
			}
			preflag[gr] = byte(m.b.ReadBits64(1))
			scalefac_scale[gr] = byte(m.b.ReadBits64(1))
			count1table_select[gr] = byte(m.b.ReadBits64(1))
		}
		// The main_data follows. It does not follow the above side information in the bitstream. The main_data ends at a location in the main_data bitstream preceding the frame header of the following frame at an offset given by the value of main_data_end (see definition of main_data_end and 3-Annex Fig.3-A.7.1)
		var xr [576]float64
		for gr := 0; gr < 2; gr++ {
			if blocksplit_flag[gr] == 1 && block_type[gr] == 2 {
				scalefac = make([][2]uint8, switch_point_l(switch_point[gr]))
				scalefacw = make([][3][2]uint8, cblimit_short-switch_point_s(switch_point[gr]))
				for cb := 0; cb < switch_point_l(switch_point[gr]); cb++ {
					if (scfsi[cb] == 0) || (gr == 0) {
						slen := scalefactors_len(scalefac_compress[gr], block_type[gr], switch_point[gr], cb)
						scalefac[cb][gr] = uint8(m.b.ReadBits64(slen))
					}
				}
				for cb := switch_point_s(switch_point[gr]); cb < cblimit_short; cb++ {
					slen := scalefactors_len(scalefac_compress[gr], block_type[gr], switch_point[gr], cb)
					for window := 0; window < 3; window++ {
						if (scfsi[cb] == 0) || (gr == 0) {
							scalefacw[cb][window][gr] = uint8(m.b.ReadBits64(slen))
						}
					}
				}
			} else {
				scalefac = make([][2]uint8, cblimit)
				for cb := 0; cb < cblimit; cb++ {
					if (scfsi[cb] == 0) || (gr == 0) {
						slen := scalefactors_len(scalefac_compress[gr], block_type[gr], switch_point[gr], cb)
						scalefac[cb][gr] = uint8(m.b.ReadBits64(slen))
					}
				}
			}
			bits := uint(part2_3_length[gr]) - part2_length(switch_point[gr], scalefac_compress[gr], block_type[gr])
			region := 0
			entry := huffmanTables[table_select[region][gr]]
			isx := 0
			cb := 0
			rcount := region_address1[gr] + 1
			sfbwidthptr := 0
			sfbwidth := sfbwidthTable[m.Sampling()].long
			if block_type[gr] == 2 {
				sfbwidth = sfbwidthTable[m.Sampling()].short
			}
			sfbound := sfbwidth[sfbwidthptr]
			sfbwidthptr++
			var factor float64
			var sfm float64 = 2
			if scalefac_scale[gr] == 1 {
				sfm = 4
			}
			exp := func(i int) {
				cb += i
				if block_type[gr] == 2 {
					// todo: factor = ...
					panic("block type 2 - scale factors")
				} else {
					factor = (float64(global_gain[gr]) - 210) / 4
					factor -= sfm * (float64(scalefac[cb][gr]) + float64(preflag[gr]*pretab[cb]))
					factor = math.Pow(2, factor)
				}
			}
			exp(0)
			read := func(b byte) {
				d := int(b)
				if d == max_table_entry {
					// The spec says that the linbits values should be added to max_table_entry
					// - 1. libmad does not use a -1. I'm not sure if the spec is wrong, libmad
					// is wrong, or I'm misinterpreting the spec.
					d += int(m.b.ReadBits64(entry.linbits))
				}
				if d != 0 {
					xr[isx] = math.Pow(float64(d), 4.0/3.0) * factor
					if m.b.ReadBits64(1) == 1 {
						xr[isx] = -xr[isx]
					}
				}
				isx++
			}
			until := m.b.read + bits
			for big := big_values[gr]; big > 0 && m.b.read < until; big-- {
				if isx == sfbound {
					sfbound += sfbwidth[sfbwidthptr]
					sfbwidthptr++
					rcount--
					if rcount == 0 {
						if region == 0 {
							rcount = region_address1[gr] + 1
						} else {
							rcount = 0
						}
						region++
						entry = huffmanTables[table_select[region][gr]]
					}
					exp(1)
				}
				pair := entry.tree.Decode(m.b)
				read(pair[0])
				read(pair[1])
			}
			if m.b.read >= until {
				panic("huffman overrun")
			}
			table := huffmanQuadTables[count1table_select[gr]]
			setQuad := func(b, offset byte) {
				var v byte
				if b&offset != 0 {
					v = 1
				}
				read(v)
			}
			for m.b.read < until {
				if isx == sfbound {
					sfbound += sfbwidth[sfbwidthptr]
					sfbwidthptr++
					exp(1)
				}
				quad := table.Decode(m.b)[0]
				setQuad(quad, 1<<0) // v
				setQuad(quad, 1<<1) // w
				setQuad(quad, 1<<2) // x
				setQuad(quad, 1<<3) // y
			}
			/*
				for position != main_data_end {
					m.b.ReadBits64(1) // ancillary_bit
				}
			//*/
		}
		_ = main_data_end
		// todo: determine channel blocktype, support blocktype == 2
		aliasReduce(xr[:])
	}
	/* else if (mode == ModeStereo) || (mode == ModeDual) || (mode == ModeJoint) {
		main_data_end := uint16(m.b.ReadBits64(9))
		private_bits := byte(m.b.ReadBits64(3))
		for ch := 0; ch < 2; ch++ {
			for scfsi_band = 0; scfsi_band < 4; scfsi_band++ {
				scfsi[scfsi_band][ch] = byte(m.b.ReadBits64(1))
			}
		}
		for gr := 0; gr < 2; gr++ {
			for ch := 0; ch < 2; ch++ {
				part2_3_length[gr][ch] = uint16(m.b.ReadBits64(12))
				big_values[gr][ch] = uint16(m.b.ReadBits64(9))
				global_gain[gr][ch] = uint16(m.b.ReadBits64(8))
				scalefac_compress[gr][ch] = byte(m.b.ReadBits64(4))
				blocksplit_flag[gr][ch] = byte(m.b.ReadBits64(1))
				if blocksplit_flag[gr][ch] {
					block_type[gr][ch] = byte(m.b.ReadBits64(2))
					switch_point[gr][ch] = uint16(m.b.ReadBits64(1))
					for region := 0; region < 2; region++ {
						table_select[region][gr][ch] = byte(m.b.ReadBits64(5))
					}
					for window := 0; window < 3; window++ {
						subblock_gain[window][gr][ch] = uint8(m.b.ReadBits64(3))
					}
				} else {
					for region := 0; region < 3; region++ {
						table_select[region][gr][ch] = byte(m.b.ReadBits64(5))
					}
					region_address1[gr][ch] = byte(m.b.ReadBits64(4))
					region_address2[gr][ch] = byte(m.b.ReadBits64(3))
				}
				preflag[gr][ch] = byte(m.b.ReadBits64(1))
				scalefac_scale[gr][ch] = byte(m.b.ReadBits64(1))
				count1table_select[gr][ch] = byte(m.b.ReadBits64(1))
				// The main_data follows. It does not follow the above side information in the bitstream. The main_data endsat a location in the main_data bitstream preceding the frame header of the following frame at an offset given by thevalue of main_data_end.
			}
		}
		for gr := 0; gr < 2; gr++ {
			for ch := 0; ch < 2; ch++ {
				if blocksplit_flag[gr][ch] == 1 && block_type[gr][ch] == 2 {
					for cb := 0; cb < switch_point_l[gr][ch]; cb++ {
						if (scfsi[cb] == 0) || (gr == 0) {
							// scalefac[cb][gr][ch]0..4 bits uimsbf
						}
					}
					for cb := switch_point_s[gr][ch]; cb < cblimit_short; cb++ {
						for window := 0; window < 3; window++ {
							if (scfsi[cb] == 0) || (gr == 0) {
								// scalefac[cb][window][gr][ch] 0..4 bits uimsbf
							}
						}
					}
				} else {
					for cb := 0; cb < cblimit; cb++ {
						if (scfsi[cb] == 0) || (gr == 0) {
							// scalefac[cb][gr][ch]0..4 bits uimsbf
						}
					}
				}
				// Huffmancodebits (part2_3_length-part2_length) bits bslbf
				for position != main_data_end {
					ancillary_bit := byte(m.b.ReadBits64(1))
				}
			}
		}
	}
	//*/
}

var (
	CI = [8]float64{
		-0.6,
		-0.535,
		-0.33,
		-0.185,
		-0.095,
		-0.041,
		-0.0142,
		-0.0037,
	}
	CS, CA [8]float64
)

func init() {
	for i, v := range CI {
		den := math.Sqrt(1 + math.Pow(v, 2))
		CS[i] = 1 / den
		CA[i] = v / den
	}
}

func aliasReduce(s []float64) {
	for x := 18; x < len(s); x += 18 {
		for i := 0; i < 8; i++ {
			a := s[x-i-1]
			b := s[x+i]
			s[x-i-1] = a*CS[i] - b*CA[i]
			s[x+i] = b*CS[i] + a*CA[i]
		}
	}
}

// Length returns the frame length in bytes.
func (m *MP3) Length() int {
	padding := 0
	if m.padding_bit != 0 {
		padding = 1
	}
	switch m.layer {
	case LayerI:
		return (12*m.Bitrate()*1000/m.Sampling() + padding) * 4
	case LayerII, LayerIII:
		return 144*m.Bitrate()*1000/m.Sampling() + padding
	default:
		return 0
	}
}

func (m *MP3) Bitrate() int {
	switch {
	case m.layer == LayerIII:
		switch m.bitrate_index {
		case 1:
			return 32
		case 2:
			return 40
		case 3:
			return 48
		case 4:
			return 56
		case 5:
			return 64
		case 6:
			return 80
		case 7:
			return 96
		case 8:
			return 112
		case 9:
			return 128
		case 10:
			return 160
		case 11:
			return 192
		case 12:
			return 224
		case 13:
			return 256
		case 14:
			return 320
		}
	}
	return 0
}

func (m *MP3) Sampling() int {
	switch m.sampling_frequency {
	case 0:
		return 44100
	case 1:
		return 48000
	case 2:
		return 32000
	}
	return 0
}

type Layer byte

const (
	LayerI   Layer = 3
	LayerII        = 2
	LayerIII       = 1
)

func (l Layer) String() string {
	switch l {
	case LayerI:
		return "layer I"
	case LayerII:
		return "layer II"
	case LayerIII:
		return "layer III"
	default:
		return "unknown"
	}
}

type Mode byte

const (
	ModeStereo Mode = 0
	ModeJoint       = 1
	ModeDual        = 2
	ModeSingle      = 3
)

type Emphasis byte

const (
	EmphasisNone  Emphasis = 0
	Emphasis50_15          = 1
	EmphasisCCIT           = 3
)

const (
	cblimit         = 21
	cblimit_short   = 12
	max_table_entry = 15
)

func switch_point_l(b byte) int {
	if b == 0 {
		return 0
	}
	return 8
}

func switch_point_s(b byte) int {
	if b == 0 {
		return 0
	}
	return 3
}

func part2_length(switch_point, scalefac_compress, block_type byte) uint {
	slen1, slen2 := slen12(scalefac_compress)
	switch switch_point {
	case 0:
		switch block_type {
		case 0, 1, 3:
			return 11*slen1 + 10*slen2
		case 2:
			return 18*slen1 + 18*slen2
		}
	case 1:
		switch block_type {
		case 0, 1, 3:
			return 11*slen1 + 10*slen2
		case 2:
			return 17*slen1 + 18*slen2
		}
	}
	panic("unreachable")
}

func slen12(scalefac_compress byte) (slen1, slen2 uint) {
	switch scalefac_compress {
	case 0:
		return 0, 0
	case 1:
		return 0, 1
	case 2:
		return 0, 2
	case 3:
		return 0, 3
	case 4:
		return 3, 0
	case 5:
		return 1, 1
	case 6:
		return 1, 2
	case 7:
		return 1, 3
	case 8:
		return 2, 1
	case 9:
		return 2, 2
	case 10:
		return 2, 3
	case 11:
		return 3, 1
	case 12:
		return 3, 2
	case 13:
		return 3, 3
	case 14:
		return 4, 2
	case 15:
		return 4, 3
	}
	panic("unreachable")
}

func scalefactors_len(scalefac_compress, block_type, switch_point byte, cb int) uint {
	slen1, slen2 := slen12(scalefac_compress)
	switch block_type {
	case 0, 1, 3:
		if cb <= 10 {
			return slen1
		}
		return slen2
	case 2:
		switch {
		case switch_point == 0 && cb <= 5:
			return slen1
		case switch_point == 0 && cb > 5:
			return slen2
		case switch_point == 1 && cb <= 5:
			// FIX: see spec note about long windows
			return slen1
		case switch_point == 1 && cb > 5:
			return slen2
		}
	}
	panic("unreachable")
}
