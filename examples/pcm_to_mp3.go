package examples

import (
	"os"
	"io"
	"github.com/sunicy/go-lame"
)

func PcmToMp3(pcmFileName, mp3FileName string) {
	pcmFile, _ := os.OpenFile(pcmFileName, os.O_RDONLY, 0555)
	mp3File, _ := os.OpenFile(mp3FileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	defer mp3File.Close()
	wr, err := lame.NewWriter(mp3File)
	if err != nil {
		panic("cannot create lame writer, err: " + err.Error())
	}
	io.Copy(wr, pcmFile)
	wr.Close()
}

func PcmToMp3Customize(pcmFileName, mp3FileName string) {
	pcmFile, _ := os.OpenFile(pcmFileName, os.O_RDONLY, 0555)
	mp3File, _ := os.OpenFile(mp3FileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	defer mp3File.Close()
	wr, err := lame.NewWriter(mp3File)
	if err != nil {
		panic("cannot create lame writer, err: " + err.Error())
	}
	wr.InSampleRate = 44100  // input sample rate
	wr.InNumChannels = 1     // number of channels: 1
	wr.OutMode = lame.MODE_STEREO // common, 2 channels
	wr.OutQuality = 9        // 0: highest; 9: lowest
	wr.OutSampleRate = 8000
	io.Copy(wr, pcmFile)
	wr.Close()
}
