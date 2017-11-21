package lame

import (
	"io"
	"encoding/binary"
	"errors"
)

// WAV file spec reference:
// http://www.topherlee.com/software/pcm-tut-wavformat.html
// properties mostly name after:
// http://soundfile.sapp.org/doc/WaveFormat/

type (
	// The header of a common wav file, length=44B
	WavHeader struct {
		// RIFF Header
		ChunkId   [4]byte // fixed "RIFF" or "RIFX" if the file is big-endian
		WavHeaderRemaining
	}

	WavHeaderRemaining struct {
		ChunkSize int32   //  Size of the overall file - 8 bytes, in bytes (32-bit integer). Typically, you'd fill this in after creation.
		Format    [4]byte // fixed "WAVE"

		// Format Header
		SubChunk1Id   [4]byte // fixed "fmt\0"
		SubChunk1Size int32   // Length of format data as listed above
		AudioFormat   int16   // Type of format (1 is PCM) - 2 byte integer
		NumChannels   int16   // Number of Channels - 2 byte integer
		SampleRate    int32   // Sample Rate - 32 byte integer. Common values are 44100 (CD), 48000 (DAT). Sample Rate = Number of Samples per second, or Hertz.
		ByteRate      int32   // *NOT BIT RATE*, but byte rate (Sample Rate * BitsPerSample * Channels) / 8.
		BlockAlign    int16   //
		BitsPerSample int16   // Bits per sample

		// Data
		SubChunk2Id   [4]byte // Contains "data"
		SubChunk2Size int32   // Number of bytes in data. Number of samples * num_channels * sample byte size
	}
)

var (
	// ChunkId is invalid
	ErrInvalidWavChunkId = errors.New("invalid wav chunk id, expected RIFF or RIFX")
	// Cannot read ChunkId at all
	ErrCannotReadChunkId = errors.New("cannot read chunkId")
	// Cannot read header at all
	ErrCannotReadHeader = errors.New("cannot read headers")
)

var (
	chunkIdLe = [4]byte{'R', 'I', 'F', 'F'} // chunkId little-endian
	chunkIdBe = [4]byte{'R', 'I', 'F', 'X'} // chunkId big-endian

	format = [4]byte{'W', 'A', 'V', 'E'}
	subChunk1Id = [4]byte{'f', 'm', 't', ' '}
	subChunk2Id = [4]byte{'d', 'a', 't', 'a'}
)

// Try to read the wav header from the given reader
// returns non-nil err if error occurs
// NOTE: the reader's position would be permanently changed, even if the given data is corrupted
func ReadWavHeader(reader io.Reader) (hdr *WavHeader, err error) {
	hdr = new(WavHeader)
	err = binary.Read(reader, binary.LittleEndian, &hdr.ChunkId)
	if err != nil {
		err = ErrCannotReadChunkId
		return
	} else if hdr.ChunkId != chunkIdLe && hdr.ChunkId != chunkIdBe {
		err = ErrInvalidWavChunkId
		return
	}

	if hdr.ChunkId == chunkIdLe {
		err = binary.Read(reader, binary.LittleEndian, &hdr.WavHeaderRemaining)
	} else {
		err = binary.Read(reader, binary.BigEndian, &hdr.WavHeaderRemaining)
	}

	if err != nil {
		err = ErrCannotReadHeader
	}

	return
}

func (hdr *WavHeader) IsBigEndian() bool {
	return hdr.ChunkId == chunkIdBe
}

// build an encodeOptions object by wavHeader
func (hdr *WavHeader) ToEncodeOptions() EncodeOptions {
	return EncodeOptions{
		InBigEndian:     hdr.IsBigEndian(),
		InSampleRate:    int(hdr.SampleRate),
		InBitsPerSample: int(hdr.BitsPerSample),
		InNumChannels:   int(hdr.NumChannels),
		OutSampleRate:   int(hdr.SampleRate), // default: remains unchanged
		OutMode:         MODE_STEREO,
		OutQuality:      0,
	}
}