package lame

import (
	"io"
	"errors"
)

// A helper for liblame, which is able to,
// 1. set quality
// 2. support mono (single channel) and stereo (2 channels ) mode
// 3. change sample rate
// 4. Big-endian and little-endian (input file)

type (
	// options for encoder
	EncodeOptions struct {
		InBigEndian bool // true if it is in big-endian
		InSampleRate   int  // Hz, e.g., 8000, 16000, 12800, 44100, etc.
		InBitsPerSample int // typically 16 should be working fine. the bit count of each sample, e.g., 2Bytes/sample->16bits
		InNumChannels  int  // count of channels, for mono ones, please remain 1, and 2 if stereo


		OutSampleRate int  // Hz
		OutMode       Mode // MODE_MONO, MODE_STEREO, etc.
		OutQuality    int  // quality: 0-highest, 9-lowest
	}

	Writer struct {
		output io.Writer
		lame *Lame
		EncodeOptions
	}
)

var ErrUnsupportedChannelNum = errors.New("only 1 and 2 channels are supported")

// create a new writer, without initializing the Lame
func NewWriter(output io.Writer) (*Writer, error) {
	lame, err := NewLame()
	if err != nil {
		return nil, err
	}
	return &Writer{
		output: output,
		lame: lame,
		EncodeOptions: EncodeOptions{
			InBigEndian:  false,
			InSampleRate:    44100,
			InBitsPerSample: 16,
			InNumChannels:   2,
			OutSampleRate:   44100,
			OutMode:         MODE_MONO,
			OutQuality:      1,
		},
	}, nil
}

// forced to init the params inside
// NOT NECESSARY
func (w *Writer) ForceUpdateParams() (err error) {
	if err = w.lame.SetInSampleRate(w.InSampleRate); err != nil {
		return
	}
	if err = w.lame.SetOutSampleRate(w.OutSampleRate); err != nil {
		return
	}
	if err = w.lame.SetNumChannels(w.InNumChannels); err != nil {
		return
	}
	if err = w.lame.SetMode(w.OutMode); err != nil {
		return
	}
	if err = w.lame.SetQuality(w.OutQuality); err != nil {
		return
	}
	if err = w.lame.InitParams(); err != nil {
		return
	}
	return nil
}

// NOT thread-safe!
// will check if we have lame object inside first!
// TODO: support 8bit/32bit!
// currently, we support 16bit depth only
func (w *Writer) Write(p []byte) (n int, err error) {
	if !w.lame.paramUpdated {
		if err = w.ForceUpdateParams(); err != nil {
			return 0, err
		}
	}

	var samples = make([]int16, len(p) / 2)
	var lo, hi int16 = 0x0001, 0x0100
	if w.InBigEndian {
		lo, hi = 0x0100, 0x0001
	}
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(p[i * 2]) * lo + int16(p[i * 2 + 1]) * hi
	}
	var outNumChannels = 2
	if w.OutMode == MODE_MONO {
		outNumChannels = 1
	}
	// inSample * (inRate / outRate) / (inNumChan / outNumChan)
	var outSampleCount = int(int64(len(samples)) * int64(w.InSampleRate) / int64(w.OutSampleRate) * int64(outNumChannels) / int64(w.InNumChannels))
	var mp3BufSize = int(1.25 * float32(outSampleCount) + 7200) // follow the instruction from LAME
	var mp3Buf = make([]byte, mp3BufSize)

	if w.InNumChannels != 1 && w.InNumChannels != 2 {
		return 0, ErrUnsupportedChannelNum
	}
	if w.InNumChannels == 1 {
		n, err = w.lame.EncodeInt16(samples, samples, mp3Buf)
	} else if w.InNumChannels == 2 {
		n, err = w.lame.EncodeInt16Interleaved(samples, mp3Buf)
	}
	if err != nil {
		return 0, err
	} else {
		_, err = w.output.Write(mp3Buf[:n])

		return 2 * len(samples), err
	}

}

func (w *Writer) Close() error {
	// try to get some residual data
	if residual, err := w.lame.EncodeFlush(); err != nil {
		return err
	} else {
		if len(residual) > 0 {
			_, err = w.output.Write(residual)
		}
		return err
	}
}
