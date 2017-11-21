package lame

import (
	"testing"
	"os"
	"io/ioutil"
	"fmt"
)

func Test_LibLame_Full(t *testing.T) {
	fn := "res/1chan_s16ple.raw"
	f, _ := os.OpenFile(fn, os.O_RDONLY, 0700)
	lame, err := NewLame()
	if err != nil {
		t.Errorf("cannot create lame: %s", err.Error())
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("cannot read file")
	}
	var samples []int16
	for i := 0; i < len(data); i += 2 {
		samples = append(samples, int16(data[i]) + int16(data[i + 1]) * 256)
	}
	fmt.Printf("sample: %d\n", len(samples))
	lame.SetNumChannels(1)
	lame.SetInSampleRate(16000)
	lame.SetMode(MODE_MONO)
	lame.SetOutSampleRate(16000)
	lame.InitParams()
	buf := make([]byte, int(1.25 * float64(len(samples)) + 7200))
	size, err := lame.EncodeInt16(samples, samples, buf)
	if err != nil {
		t.Errorf("cannot encode! %s", err.Error())
	}
	residual, _ := lame.EncodeFlush()
	r := append(buf[:size], residual...)
	t.Logf("len=%d", len(r))
	ioutil.WriteFile("out.mp3", r, 0700)
}


