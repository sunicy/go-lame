package examples

import (
	"os"
	"io"
	"github.com/sunicy/go-lame"
)

func WavToMp3(wavFileName, mp3FileName string) {
	// open files
	wavFile, _ := os.OpenFile(wavFileName, os.O_RDONLY, 0555)
	mp3File, _ := os.OpenFile(mp3FileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	defer mp3File.Close()

	// parsing wav info
	// NOTE: reader position moves even if it is not a wav file
	wavHdr, err := lame.ReadWavHeader(wavFile)
	if err != nil {
		panic("not a wav file, err=" + err.Error())
	}

	wr, _ := lame.NewWriter(mp3File)
	wr.EncodeOptions = wavHdr.ToEncodeOptions()
	io.Copy(wr, wavFile) // wavFile's pos has been changed!
	wr.Close()
}
