package lame

import (
	"testing"
	"os"
	"./compare"
)

func Test_ReadWavHeader(t *testing.T) {
	tests := []struct {
		fn string
		hdr WavHeader
	}{
		{
			fn: "res/1chan_s16ple.wav",
			hdr: WavHeader{
				ChunkId: chunkIdLe,
				WavHeaderRemaining: WavHeaderRemaining{
					ChunkSize:     0x04e224,
					Format:        format,
					SubChunk1Id:   subChunk1Id,
					SubChunk1Size: 16,
					AudioFormat:   1,
					NumChannels:   1,
					SampleRate:    16000,
					ByteRate:      32000,
					BlockAlign:    2,
					BitsPerSample: 16,
					SubChunk2Id:   subChunk2Id,
					SubChunk2Size: 320000,
				},
			},
		},
	}
	for idx, test := range tests {
		t.Run(test.fn, func(t *testing.T) {
			f, err := os.OpenFile(test.fn, os.O_RDONLY, 0700)
			if err != nil {
				t.Errorf("Case#%d, cannot open file %s, err=%s", idx, test.fn, err.Error())
				return
			}
			defer f.Close()
			hdr, err := ReadWavHeader(f)
			if err != nil {
				t.Errorf("Case#%d, %s", idx, err.Error())
				return
			}
			diffs, err := compare.Compare(&test.hdr, &hdr)
			if err != nil {
				t.Errorf("%s", err.Error())
			} else if len(diffs) > 0 {
				t.Errorf("%#v", diffs)
			}
		})
	}
}

func Test_LameStruct(t *testing.T) {
	l, err := NewLame()
	t.Logf("%#v", l.lgs)
	t.Logf("%#v", err)
}

