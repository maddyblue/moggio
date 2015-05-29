package frame_test

import (
	"bytes"
	"crypto/md5"
	"io"
	"testing"

	"github.com/mjibson/mog/_third_party/gopkg.in/mewkiz/flac.v1"
)

var golden = []struct {
	name string
}{
	// i=0
	{
		name: "../testdata/59996.flac",
	},

	// i=1
	{
		name: "../testdata/172960.flac",
	},

	// i=2
	{
		name: "../testdata/189983.flac",
	},
}

func TestFrameHash(t *testing.T) {
	for i, g := range golden {
		stream, err := flac.ParseFile(g.name)
		if err != nil {
			t.Fatal(err)
		}
		defer stream.Close()

		md5sum := md5.New()
		for frameNum := 0; ; frameNum++ {
			frame, err := stream.ParseNext()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Errorf("i=%d, frameNum=%d: error while parsing frame; %v", i, frameNum, err)
				continue
			}
			frame.Hash(md5sum)
		}
		want := stream.Info.MD5sum[:]
		got := md5sum.Sum(nil)
		// Verify the decoded audio samples by comparing the MD5 checksum that is
		// stored in StreamInfo with the computed one.
		if !bytes.Equal(got, want) {
			t.Errorf("i=%d: MD5 checksum mismatch for decoded audio samples; expected %32x, got %32x", i, want, got)
		}
	}
}
