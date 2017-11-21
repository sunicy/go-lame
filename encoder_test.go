package lame

import (
	"testing"
	"os"
	"io"
)

func Test_Encoder_Full(t *testing.T) {
	fin, _ := os.OpenFile("res/1chan_s16ple.raw", os.O_RDONLY, 0700)
	fout, _ := os.OpenFile("out2.mp3", os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0755)
	wr, err := NewWriter(fout)
	if err != nil {
		t.Errorf("cannot create lame writer, %s", err.Error())
	}
	wr.InNumChannels = 1
	wr.InSampleRate = 16000
	wr.OutSampleRate = 16000
	wr.OutMode = MODE_MONO

	_, err = io.Copy(wr, fin)
	if err != nil {
		t.Errorf("cannot write into file, %s", err.Error())
	}
	wr.Close()
	fout.Close()
}
