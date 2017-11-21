# go-lame

Yet another simple wrapper of libmp3lame for golang. 
It focuses on converting __raw PCM__, and __WAV__ files into mp3. 

# Examples

## PCM to MP3

Assuming the sample rate of the PCM file is 44100, and the output's is the same.

```go

func PcmToMp3(pcmFileName, mp3FileName string) {
	pcmFile, _ := os.OpenFile(pcmFileName, os.O_RDONLY, 0555)
	mp3File, _ := os.OpenFile(mp3FileName, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0755)
	defer mp3File.Close()
	wr, err := lame.NewWriter(mp3File)
	if err != nil {
		panic("cannot create lame writer, err: " + err.Error())
	}
	io.Copy(wr, pcmFile)
	wr.Close()
}

```

### PCM to MP3 (Customize params)

Assuming here goes some properties of the PCM file:

    - SampleRate: 16000,
    - Number of Channel: 1 (single),

and the output (.mp3) is expected to meet demands below:

    - SampleRate: 8000
    - Quality: 9 (lowest)
    - Mode: Stereo (2 channels)
   
```go
func PcmToMp3Customize(pcmFileName, mp3FileName string) {
	pcmFile, _ := os.OpenFile(pcmFileName, os.O_RDONLY, 0555)
	mp3File, _ := os.OpenFile(mp3FileName, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0755)
	defer mp3File.Close()
	wr, err := lame.NewWriter(mp3File)
	if err != nil {
		panic("cannot create lame writer, err: " + err.Error())
	}
	wr.InSampleRate = 44100  // input sample rate
	wr.InNumChannels = 1     // number of channels: 1
	wr.OutMode = lame.MODE_STEREO // common, 2 channels
	wr.OutQuality = 9        // 0: highest; 9: lowest 
	wr.OutSampleRate = 8000  // output sample rate
	
	io.Copy(wr, pcmFile)
	wr.Close()
}
```

## WAV to MP3

```go
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
```

# Roadmap

- [x] Wrapping functions from libmp3lame
- [x] WavFile parsing support
- [x] Shortcut to using wrapped functions
- [x] Supporting parsing both little-endian and big-endian PCM files
- [ ] Thorough tests 
- [ ] Supporting bit depth other than 16