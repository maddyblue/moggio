package meta_test

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/mjibson/mog/_third_party/gopkg.in/mewkiz/flac.v1"
	"github.com/mjibson/mog/_third_party/gopkg.in/mewkiz/flac.v1/meta"
)

var golden = []struct {
	name   string
	info   *meta.StreamInfo
	blocks []*meta.Block
}{
	// i=0
	{
		name: "../testdata/59996.flac",
		info: &meta.StreamInfo{BlockSizeMin: 0x1000, BlockSizeMax: 0x1000, FrameSizeMin: 0x44c5, FrameSizeMax: 0x4588, SampleRate: 0xac44, NChannels: 0x2, BitsPerSample: 0x18, NSamples: 0x2000, MD5sum: [16]uint8{0x95, 0xba, 0xe5, 0xe2, 0xc7, 0x45, 0xbb, 0x3c, 0xa9, 0x5c, 0xa3, 0xb1, 0x35, 0xc9, 0x43, 0xf4}},
		blocks: []*meta.Block{
			{
				Header: meta.Header{Type: 0x4, Length: 202, IsLast: true},
				Body:   &meta.VorbisComment{Vendor: "reference libFLAC 1.2.1 20070917", Tags: [][2]string{{"Description", "Waving a bamboo staff"}, {"YEAR", "2008"}, {"ARTIST", "qubodup aka Iwan Gabovitch | qubodup@gmail.com"}, {"COMMENTS", "I release this file into the public domain"}}},
			},
		},
	},

	// i=1
	{
		name: "../testdata/172960.flac",
		info: &meta.StreamInfo{BlockSizeMin: 0x1000, BlockSizeMax: 0x1000, FrameSizeMin: 0xb7c, FrameSizeMax: 0x256b, SampleRate: 0x17700, NChannels: 0x2, BitsPerSample: 0x10, NSamples: 0xaaa3, MD5sum: [16]uint8{0x76, 0x3d, 0xa8, 0xa5, 0xb7, 0x58, 0xe6, 0x2, 0x61, 0xb4, 0xd4, 0xc2, 0x88, 0x4d, 0x8e, 0xe}},
		blocks: []*meta.Block{
			{
				Header: meta.Header{Type: 0x4, Length: 180, IsLast: true},
				Body:   &meta.VorbisComment{Vendor: "reference libFLAC 1.2.1 20070917", Tags: [][2]string{{"GENRE", "Sound Clip"}, {"ARTIST", "Iwan 'qubodup' Gabovitch"}, {"Artist Homepage", "http://qubodup.net"}, {"Artist Email", "qubodup@gmail.com"}, {"DATE", "2012"}}},
			},
		},
	},

	// i=2
	{
		name: "../testdata/189983.flac",
		info: &meta.StreamInfo{BlockSizeMin: 0x1200, BlockSizeMax: 0x1200, FrameSizeMin: 0x94d, FrameSizeMax: 0x264a, SampleRate: 0xac44, NChannels: 0x2, BitsPerSample: 0x10, NSamples: 0x50f4, MD5sum: [16]uint8{0x63, 0x28, 0xed, 0x6d, 0xd3, 0xe, 0x55, 0xfb, 0xa5, 0x73, 0x69, 0x2b, 0xb7, 0x35, 0x73, 0xb7}},
		blocks: []*meta.Block{
			{
				Header: meta.Header{Type: 0x4, Length: 40, IsLast: true},
				Body:   &meta.VorbisComment{Vendor: "reference libFLAC 1.2.1 20070917", Tags: nil},
			},
		},
	},

	// i=3
	{
		name: "testdata/input-SCPAP.flac",
		info: &meta.StreamInfo{BlockSizeMin: 0x1200, BlockSizeMax: 0x1200, FrameSizeMin: 0xe, FrameSizeMax: 0x10, SampleRate: 0xac44, NChannels: 0x2, BitsPerSample: 0x10, NSamples: 0x16f8, MD5sum: [16]uint8{0x74, 0xff, 0xd4, 0x73, 0x7e, 0xb5, 0x48, 0x8d, 0x51, 0x2b, 0xe4, 0xaf, 0x58, 0x94, 0x33, 0x62}},
		blocks: []*meta.Block{
			{
				Header: meta.Header{Type: 0x3, Length: 180, IsLast: false},
				Body:   &meta.SeekTable{Points: []meta.SeekPoint{meta.SeekPoint{SampleNum: 0x0, Offset: 0x0, NSamples: 0x1200}, meta.SeekPoint{SampleNum: 0x1200, Offset: 0xe, NSamples: 0x4f8}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}}},
			},
			{
				Header: meta.Header{Type: 0x5, Length: 540, IsLast: false},
				Body:   &meta.CueSheet{MCN: "1234567890123", NLeadInSamples: 0x15888, IsCompactDisc: true, Tracks: []meta.CueSheetTrack{meta.CueSheetTrack{Offset: 0x0, Num: 0x1, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex{meta.CueSheetTrackIndex{Offset: 0x0, Num: 0x1}, meta.CueSheetTrackIndex{Offset: 0x24c, Num: 0x2}}}, meta.CueSheetTrack{Offset: 0xb7c, Num: 0x2, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex{meta.CueSheetTrackIndex{Offset: 0x0, Num: 0x1}}}, meta.CueSheetTrack{Offset: 0x16f8, Num: 0xaa, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex(nil)}}},
			},
			{
				Header: meta.Header{Type: 0x1, Length: 4, IsLast: false},
				Body:   nil,
			},
			{
				Header: meta.Header{Type: 0x2, Length: 4, IsLast: false},
				Body:   &meta.Application{ID: 0x66616b65, Data: nil},
			},
			{
				Header: meta.Header{Type: 0x1, Length: 3201, IsLast: true},
				Body:   nil,
			},
		},
	},

	// i=4
	{
		name: "testdata/input-SCVA.flac",
		info: &meta.StreamInfo{BlockSizeMin: 0x1200, BlockSizeMax: 0x1200, FrameSizeMin: 0xe, FrameSizeMax: 0x10, SampleRate: 0xac44, NChannels: 0x2, BitsPerSample: 0x10, NSamples: 0x16f8, MD5sum: [16]uint8{0x74, 0xff, 0xd4, 0x73, 0x7e, 0xb5, 0x48, 0x8d, 0x51, 0x2b, 0xe4, 0xaf, 0x58, 0x94, 0x33, 0x62}},
		blocks: []*meta.Block{
			{
				Header: meta.Header{Type: 0x3, Length: 180, IsLast: false},
				Body:   &meta.SeekTable{Points: []meta.SeekPoint{meta.SeekPoint{SampleNum: 0x0, Offset: 0x0, NSamples: 0x1200}, meta.SeekPoint{SampleNum: 0x1200, Offset: 0xe, NSamples: 0x4f8}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}}},
			},
			{
				Header: meta.Header{Type: 0x5, Length: 540, IsLast: false},
				Body:   &meta.CueSheet{MCN: "1234567890123", NLeadInSamples: 0x15888, IsCompactDisc: true, Tracks: []meta.CueSheetTrack{meta.CueSheetTrack{Offset: 0x0, Num: 0x1, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex{meta.CueSheetTrackIndex{Offset: 0x0, Num: 0x1}, meta.CueSheetTrackIndex{Offset: 0x24c, Num: 0x2}}}, meta.CueSheetTrack{Offset: 0xb7c, Num: 0x2, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex{meta.CueSheetTrackIndex{Offset: 0x0, Num: 0x1}}}, meta.CueSheetTrack{Offset: 0x16f8, Num: 0xaa, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex(nil)}}},
			},
			{
				Header: meta.Header{Type: 0x4, Length: 203, IsLast: false},
				Body:   &meta.VorbisComment{Vendor: "reference libFLAC 1.1.3 20060805", Tags: [][2]string{{"REPLAYGAIN_TRACK_PEAK", "0.99996948"}, {"REPLAYGAIN_TRACK_GAIN", "-7.89 dB"}, {"REPLAYGAIN_ALBUM_PEAK", "0.99996948"}, {"REPLAYGAIN_ALBUM_GAIN", "-7.89 dB"}, {"artist", "1"}, {"title", "2"}}},
			},
			{
				Header: meta.Header{Type: 0x2, Length: 4, IsLast: true},
				Body:   &meta.Application{ID: 0x66616b65, Data: nil},
			},
		},
	},

	// i=5
	{
		name: "testdata/input-SCVAUP.flac",
		info: &meta.StreamInfo{BlockSizeMin: 0x1200, BlockSizeMax: 0x1200, FrameSizeMin: 0xe, FrameSizeMax: 0x10, SampleRate: 0xac44, NChannels: 0x2, BitsPerSample: 0x10, NSamples: 0x16f8, MD5sum: [16]uint8{0x74, 0xff, 0xd4, 0x73, 0x7e, 0xb5, 0x48, 0x8d, 0x51, 0x2b, 0xe4, 0xaf, 0x58, 0x94, 0x33, 0x62}},
		blocks: []*meta.Block{
			{
				Header: meta.Header{Type: 0x3, Length: 180, IsLast: false},
				Body:   &meta.SeekTable{Points: []meta.SeekPoint{meta.SeekPoint{SampleNum: 0x0, Offset: 0x0, NSamples: 0x1200}, meta.SeekPoint{SampleNum: 0x1200, Offset: 0xe, NSamples: 0x4f8}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}}},
			},
			{
				Header: meta.Header{Type: 0x5, Length: 540, IsLast: false},
				Body:   &meta.CueSheet{MCN: "1234567890123", NLeadInSamples: 0x15888, IsCompactDisc: true, Tracks: []meta.CueSheetTrack{meta.CueSheetTrack{Offset: 0x0, Num: 0x1, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex{meta.CueSheetTrackIndex{Offset: 0x0, Num: 0x1}, meta.CueSheetTrackIndex{Offset: 0x24c, Num: 0x2}}}, meta.CueSheetTrack{Offset: 0xb7c, Num: 0x2, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex{meta.CueSheetTrackIndex{Offset: 0x0, Num: 0x1}}}, meta.CueSheetTrack{Offset: 0x16f8, Num: 0xaa, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex(nil)}}},
			},
			{
				Header: meta.Header{Type: 0x4, Length: 203, IsLast: false},
				Body:   &meta.VorbisComment{Vendor: "reference libFLAC 1.1.3 20060805", Tags: [][2]string{{"REPLAYGAIN_TRACK_PEAK", "0.99996948"}, {"REPLAYGAIN_TRACK_GAIN", "-7.89 dB"}, {"REPLAYGAIN_ALBUM_PEAK", "0.99996948"}, {"REPLAYGAIN_ALBUM_GAIN", "-7.89 dB"}, {"artist", "1"}, {"title", "2"}}},
			},
			{
				Header: meta.Header{Type: 0x2, Length: 4, IsLast: false},
				Body:   &meta.Application{ID: 0x66616b65, Data: nil},
			},
			{
				Header: meta.Header{Type: 0x7e, Length: 0, IsLast: false},
				Body:   nil,
			},
			{
				Header: meta.Header{Type: 0x1, Length: 3201, IsLast: true},
				Body:   nil,
			},
		},
	},
	// i=6
	{
		name: "testdata/input-SCVPAP.flac",
		info: &meta.StreamInfo{BlockSizeMin: 0x1200, BlockSizeMax: 0x1200, FrameSizeMin: 0xe, FrameSizeMax: 0x10, SampleRate: 0xac44, NChannels: 0x2, BitsPerSample: 0x10, NSamples: 0x16f8, MD5sum: [16]uint8{0x74, 0xff, 0xd4, 0x73, 0x7e, 0xb5, 0x48, 0x8d, 0x51, 0x2b, 0xe4, 0xaf, 0x58, 0x94, 0x33, 0x62}},
		blocks: []*meta.Block{
			{
				Header: meta.Header{Type: 0x3, Length: 180, IsLast: false},
				Body:   &meta.SeekTable{Points: []meta.SeekPoint{meta.SeekPoint{SampleNum: 0x0, Offset: 0x0, NSamples: 0x1200}, meta.SeekPoint{SampleNum: 0x1200, Offset: 0xe, NSamples: 0x4f8}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}}},
			},
			{
				Header: meta.Header{Type: 0x5, Length: 540, IsLast: false},
				Body:   &meta.CueSheet{MCN: "1234567890123", NLeadInSamples: 0x15888, IsCompactDisc: true, Tracks: []meta.CueSheetTrack{meta.CueSheetTrack{Offset: 0x0, Num: 0x1, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex{meta.CueSheetTrackIndex{Offset: 0x0, Num: 0x1}, meta.CueSheetTrackIndex{Offset: 0x24c, Num: 0x2}}}, meta.CueSheetTrack{Offset: 0xb7c, Num: 0x2, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex{meta.CueSheetTrackIndex{Offset: 0x0, Num: 0x1}}}, meta.CueSheetTrack{Offset: 0x16f8, Num: 0xaa, ISRC: "", IsAudio: true, HasPreEmphasis: false, Indicies: []meta.CueSheetTrackIndex(nil)}}},
			},
			{
				Header: meta.Header{Type: 0x4, Length: 203, IsLast: false},
				Body:   &meta.VorbisComment{Vendor: "reference libFLAC 1.1.3 20060805", Tags: [][2]string{{"REPLAYGAIN_TRACK_PEAK", "0.99996948"}, {"REPLAYGAIN_TRACK_GAIN", "-7.89 dB"}, {"REPLAYGAIN_ALBUM_PEAK", "0.99996948"}, {"REPLAYGAIN_ALBUM_GAIN", "-7.89 dB"}, {"artist", "1"}, {"title", "2"}}},
			},
			{
				Header: meta.Header{Type: 0x1, Length: 4, IsLast: false},
				Body:   nil,
			},
			{
				Header: meta.Header{Type: 0x2, Length: 4, IsLast: false},
				Body:   &meta.Application{ID: 0x66616b65, Data: nil},
			},
			{
				Header: meta.Header{Type: 0x1, Length: 3201, IsLast: true},
				Body:   nil,
			},
		},
	},

	// i=7
	{
		name: "testdata/input-SVAUP.flac",
		info: &meta.StreamInfo{BlockSizeMin: 0x1200, BlockSizeMax: 0x1200, FrameSizeMin: 0xe, FrameSizeMax: 0x10, SampleRate: 0xac44, NChannels: 0x2, BitsPerSample: 0x10, NSamples: 0x16f8, MD5sum: [16]uint8{0x74, 0xff, 0xd4, 0x73, 0x7e, 0xb5, 0x48, 0x8d, 0x51, 0x2b, 0xe4, 0xaf, 0x58, 0x94, 0x33, 0x62}},
		blocks: []*meta.Block{
			{
				Header: meta.Header{Type: 0x3, Length: 180, IsLast: false},
				Body:   &meta.SeekTable{Points: []meta.SeekPoint{meta.SeekPoint{SampleNum: 0x0, Offset: 0x0, NSamples: 0x1200}, meta.SeekPoint{SampleNum: 0x1200, Offset: 0xe, NSamples: 0x4f8}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}, meta.SeekPoint{SampleNum: 0xffffffffffffffff, Offset: 0x0, NSamples: 0x0}}},
			},
			{
				Header: meta.Header{Type: 0x4, Length: 203, IsLast: false},
				Body:   &meta.VorbisComment{Vendor: "reference libFLAC 1.1.3 20060805", Tags: [][2]string{{"REPLAYGAIN_TRACK_PEAK", "0.99996948"}, {"REPLAYGAIN_TRACK_GAIN", "-7.89 dB"}, {"REPLAYGAIN_ALBUM_PEAK", "0.99996948"}, {"REPLAYGAIN_ALBUM_GAIN", "-7.89 dB"}, {"artist", "1"}, {"title", "2"}}},
			},
			{
				Header: meta.Header{Type: 0x2, Length: 4, IsLast: false},
				Body:   &meta.Application{ID: 0x66616b65, Data: nil},
			},
			{
				Header: meta.Header{Type: 0x7e, Length: 0, IsLast: false},
				Body:   nil,
			},
			{
				Header: meta.Header{Type: 0x1, Length: 3201, IsLast: true},
				Body:   nil,
			},
		},
	},

	// i=8
	{
		name: "testdata/input-VA.flac",
		info: &meta.StreamInfo{BlockSizeMin: 0x1200, BlockSizeMax: 0x1200, FrameSizeMin: 0xe, FrameSizeMax: 0x10, SampleRate: 0xac44, NChannels: 0x2, BitsPerSample: 0x10, NSamples: 0x16f8, MD5sum: [16]uint8{0x74, 0xff, 0xd4, 0x73, 0x7e, 0xb5, 0x48, 0x8d, 0x51, 0x2b, 0xe4, 0xaf, 0x58, 0x94, 0x33, 0x62}},
		blocks: []*meta.Block{
			{
				Header: meta.Header{Type: 0x4, Length: 203, IsLast: false},
				Body:   &meta.VorbisComment{Vendor: "reference libFLAC 1.1.3 20060805", Tags: [][2]string{{"REPLAYGAIN_TRACK_PEAK", "0.99996948"}, {"REPLAYGAIN_TRACK_GAIN", "-7.89 dB"}, {"REPLAYGAIN_ALBUM_PEAK", "0.99996948"}, {"REPLAYGAIN_ALBUM_GAIN", "-7.89 dB"}, {"artist", "1"}, {"title", "2"}}},
			},
			{
				Header: meta.Header{Type: 0x2, Length: 4, IsLast: true},
				Body:   &meta.Application{ID: 0x66616b65, Data: nil},
			},
		},
	},
}

func TestParseBlocks(t *testing.T) {
	for i, g := range golden {
		stream, err := flac.ParseFile(g.name)
		if err != nil {
			t.Fatal(err)
		}
		defer stream.Close()
		blocks := stream.Blocks

		if len(blocks) != len(g.blocks) {
			t.Errorf("i=%d: invalid number of metadata blocks; expected %d, got %d", i, len(g.blocks), len(blocks))
			continue
		}

		got := stream.Info
		want := g.info
		if !reflect.DeepEqual(got, want) {
			t.Errorf("i=%d: metadata StreamInfo block bodies differ; expected %#v, got %#v", i, want, got)
		}

		for j, got := range blocks {
			want := g.blocks[j]
			if !reflect.DeepEqual(got.Header, want.Header) {
				t.Errorf("i=%d, j=%d: metadata block headers differ; expected %#v, got %#v", i, j, want.Header, got.Header)
			}
			if !reflect.DeepEqual(got.Body, want.Body) {
				t.Errorf("i=%d, j=%d: metadata block bodies differ; expected %#v, got %#v", i, j, want.Body, got.Body)
			}
		}
	}
}

func TestParsePicture(t *testing.T) {
	stream, err := flac.ParseFile("testdata/silence.flac")
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	want, err := ioutil.ReadFile("testdata/silence.jpg")
	if err != nil {
		t.Fatal(err)
	}

	for _, block := range stream.Blocks {
		if block.Type == meta.TypePicture {
			pic := block.Body.(*meta.Picture)
			got := pic.Data
			if !bytes.Equal(got, want) {
				t.Errorf("picture data differ; expected %v, got %v", want, got)
			}
			break
		}
	}
}
