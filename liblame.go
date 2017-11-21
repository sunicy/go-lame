package lame

/*
#cgo LDFLAGS: -lmp3lame

#include "lame/lame.h"
*/
import "C"
import (
	"runtime"
	"errors"
	"fmt"
	"unsafe"
)

// All functions are ported from libmp3lame www.mp3dev.org

type (
	// Bridge towards lame_global_struct
	Lame struct {
		// ptr to lame_global_struct
		lgs *C.struct_lame_global_struct
		// lame_init_params not called after some settings?
		paramUpdated bool
	}

	// MPEG_mode_e
	Mode int

	// the VBR mode vbr_mode_e
	VBRMode int

	// asm_optimizations_e
	AsmOptimizations int
)

// let us define modes here
const (
	MODE_STEREO        Mode = iota
	MODE_JOINT_STEREO
	MODE_DUAL_CHANNEL   /* LAME doesn't supports this! */
	MODE_MONO
	MODE_NOT_SET        /* Don't use this! It's used for sanity checks. */
	MODE_MAX_INDICATOR
)

// let us define vbr mode here
const (
	VBR_OFF           VBRMode = iota
	VBR_MT
	VBR_RH
	VBR_ABR
	VBR_MTRH
	VBR_MAX_INDICATOR
	VBR_DEFAULT       = VBR_MTRH
)

// let us define asm optimizations
const (
	AO_INVALID   AsmOptimizations = iota
	AO_MMX
	AO_AMD_3DNOW
	AO_SSE
)

const (
	_SAFE_MP3_BUF_SIZE = 7200 // the buffer size which is safe to hold possible data at a time
)

var (
	ErrInsufficientMemory = errors.New("cannot allocate memory for Lame")
	ErrCannotInitParams   = errors.New("cannot init params, most likely invalid params, especially sample rates")
	ErrParamsNotInit      = errors.New("params not initialized")
	ErrTooSmallBuffer     = errors.New("too small buffer")
	ErrUnknown            = errors.New("unknown")
	ErrEmptyArguments     = errors.New("some arguments are empty")
	ErrInvalidSampleRate  = errors.New("invalid sample rate, supports only 8, 12, 16, 22, 32, 44.1, 48k")
)

// create and init a lame struct
func NewLame() (l *Lame, err error) {
	l = new(Lame)
	l.paramUpdated = false
	l.lgs = C.lame_init()
	if l.lgs == nil {
		return nil, ErrInsufficientMemory
	}
	runtime.SetFinalizer(l, finalizer)
	return l, nil

}

/*
 * NOTE: MUST BE CALLED BEFORE CONVERSION
 * sets more internal configuration based on data provided above.
 */
func (l *Lame) InitParams() error {
	l.checkLgs()
	if retCode := int(C.lame_init_params(l.lgs)); retCode != 0 {
		return ErrCannotInitParams
	}
	l.paramUpdated = true // Yep!
	return nil
}

/*
 * NOTE: MUST BE CALLED AFTER CONVERSION
 * returns the length of trailing bytes
 */
func (l *Lame) EncodeFlush() (residual []byte, err error) {
	l.checkLgs()
	buf := make([]byte, _SAFE_MP3_BUF_SIZE)
	cMp3Buf := (*C.uchar)(unsafe.Pointer(&buf[0]))
	residualSize := int(C.lame_encode_flush(l.lgs, cMp3Buf, C.int(len(buf))))
	if _, err = l.encodeError(residualSize); err != nil {
		return
	}
	return buf[:residualSize], nil
}

// bind to release the memory
func finalizer(l *Lame) {
	C.lame_close(l.lgs)
}

// PANIC if Lame is not initialized
func (l *Lame) checkLgs() {
	if l.lgs == nil {
		panic("uninitialized Lame struct")
	}
}

/*
 * BELOW ARE <CONVERT> FUNCTIONS
 */
// encode pcm to mp3, given buffer
// according to LAME, there is a loose bound for the buf, len=1.25*numSamped + 7200
func (l *Lame) EncodeInt16(dataLeft, dataRight []int16, mp3Buf []byte) (int, error) {
	if len(mp3Buf) == 0 || len(dataLeft) == 0 || len(dataRight) == 0 {
		return 0, ErrEmptyArguments
	}
	cMp3Buf := (*C.uchar)(unsafe.Pointer(&mp3Buf[0]))
	cDataLeft := (*C.short)(unsafe.Pointer(&dataLeft[0]))
	cDataRight := (*C.short)(unsafe.Pointer(&dataRight[0]))
	ret := int(C.lame_encode_buffer(l.lgs, cDataLeft, cDataRight, C.int(len(dataLeft)), cMp3Buf, C.int(len(mp3Buf))))
	return l.encodeError(ret)
}

// encode pcm to mp3, given buffer. same with encodeInt64, except data for left and right channels being interleaved
// according to LAME, there is a loose bound for the buf, len=1.25*numSamped + 7200
func (l *Lame) EncodeInt16Interleaved(data []int16, mp3Buf []byte) (int, error) {
	if len(mp3Buf) == 0 || len(data) == 0 {
		return 0, ErrEmptyArguments
	}
	cMp3Buf := (*C.uchar)(unsafe.Pointer(&mp3Buf[0]))
	cData := (*C.short)(unsafe.Pointer(&data[0]))
	ret := int(C.lame_encode_buffer_interleaved(l.lgs, cData, C.int(len(data)), cMp3Buf, C.int(len(mp3Buf))))
	return l.encodeError(ret)
}

func (l *Lame) EncodeInt32(dataLeft, dataRight []int32, mp3Buf []byte) (int, error) {
	if len(mp3Buf) == 0 || len(dataLeft) == 0 || len(dataRight) == 0 {
		return 0, ErrEmptyArguments
	}
	cMp3Buf := (*C.uchar)(unsafe.Pointer(&mp3Buf[0]))
	cDataLeft := (*C.int)(unsafe.Pointer(&dataLeft[0]))
	cDataRight := (*C.int)(unsafe.Pointer(&dataRight[0]))
	ret := int(C.lame_encode_buffer_int(l.lgs, cDataLeft, cDataRight, C.int(len(dataLeft)), cMp3Buf, C.int(len(mp3Buf))))
	return l.encodeError(ret)
}

func (l *Lame) EncodeInt64(dataLeft, dataRight []int32, mp3Buf []byte) (int, error) {
	if len(mp3Buf) == 0 || len(dataLeft) == 0 || len(dataRight) == 0 {
		return 0, ErrEmptyArguments
	}
	cMp3Buf := (*C.uchar)(unsafe.Pointer(&mp3Buf[0]))
	cDataLeft := (*C.long)(unsafe.Pointer(&dataLeft[0]))
	cDataRight := (*C.long)(unsafe.Pointer(&dataRight[0]))
	ret := int(C.lame_encode_buffer_long2(l.lgs, cDataLeft, cDataRight, C.int(len(dataLeft)), cMp3Buf, C.int(len(mp3Buf))))
	return l.encodeError(ret)
}

func (l *Lame) encodeError(ret int) (size int, err error) {
	if ret >= 0 {
		return ret, nil
	}
	switch ret {
	case -1:
		return 0, ErrTooSmallBuffer
	case -2:
		return 0, ErrInsufficientMemory
	case -3:
		return 0, ErrParamsNotInit
	default:
		return 0, ErrUnknown
	}
}

/*
 * BELOW ARE <PARAM> SETTERS & GETTERS
 */

// btw, we will set paramUpdated = false if retCode=0(succeeded)
func (l *Lame) setterError(name string, retCode int) error {
	if retCode != 0 {
		return fmt.Errorf("cannot %s, code=%d", name, retCode)
	}
	l.paramUpdated = false
	return nil
}

func (l *Lame) checkSampleRate(rate int) error {
	if rate == 8000 || rate == 11025 || rate == 12000 || rate == 16000 || rate == 22050 || rate == 24000 || rate == 32000 || rate == 44100 || rate == 48000 {
		return nil
	} else {
		return ErrInvalidSampleRate
	}
}

// set input sample rate
func (l *Lame) SetInSampleRate(sampleRate int) error {
	l.checkLgs() // try to panic
	if err := l.checkSampleRate(sampleRate); err != nil {
		return err
	}
	return l.setterError("lame_set_in_samplerate", int(C.lame_set_in_samplerate(l.lgs, C.int(sampleRate))))
}

// get input sample rate
func (l *Lame) GetInSampleRate() int {
	l.checkLgs()
	return int(C.lame_get_in_samplerate(l.lgs))
}

/* number of channels in input stream. default=2  */
// set number of channels
func (l *Lame) SetNumChannels(numChannels int) error {
	l.checkLgs()
	return l.setterError("lame_set_num_channels", int(C.lame_set_num_channels(l.lgs, C.int(numChannels))))
}

// get the number of channels
func (l *Lame) GetNumChannels() int {
	l.checkLgs()
	return int(C.lame_get_in_samplerate(l.lgs))
}

/*
  scale the input by this amount before encoding.  default=1
  (not used by decoding routines)
*/
func (l *Lame) SetScale(scale float32) error {
	l.checkLgs()
	return l.setterError("lame_set_scale", int( C.lame_set_scale(l.lgs, C.float(scale))))
}
func (l *Lame) GetScale() float32 {
	l.checkLgs()
	return float32(C.lame_get_scale(l.lgs))
}

/*
  scale the channel 0 (left) input by this amount before encoding.  default=1
  (not used by decoding routines)
  ref: https://github.com/gypified/libmp3lame/blob/master/include/lame.h#L206
*/
func (l *Lame) SetScaleLeft(scale float32) error {
	l.checkLgs()
	return l.setterError("lame_set_scale_left", int(C.lame_set_scale_left(l.lgs, C.float(scale))))
}

/*
  scale the channel 1 (right) input by this amount before encoding.  default=1
  (not used by decoding routines)
*/
func (l *Lame) SetScaleRight(scaleRight float32) error {
	l.checkLgs()
	return l.setterError("lame_set_scale_right", int(C.lame_set_scale_right(l.lgs, C.float(scaleRight))))
}

func (l *Lame) GetScaleRight() float32 {
	l.checkLgs()
	return float32(C.lame_get_scale_right(l.lgs))
}

/*
  output sample rate in Hz.  default = 0, which means LAME picks best value
  based on the amount of compression.  MPEG only allows:
  MPEG1    32, 44.1,   48khz
  MPEG2    16, 22.05,  24
  MPEG2.5   8, 11.025, 12
  (not used by decoding routines)
*/
func (l *Lame) SetOutSampleRate(outSampleRate int) error {
	l.checkLgs()
	if err := l.checkSampleRate(outSampleRate); err != nil {
		return err
	}
	return l.setterError("lame_set_out_samplerate", int(C.lame_set_out_samplerate(l.lgs, C.int(outSampleRate))))
}

func (l *Lame) GetOutSampleRate() int {
	l.checkLgs()
	return int(C.lame_get_out_samplerate(l.lgs))
}

/*  below are general control parameters
	default: 0
	set to 1 if you need LAME to ollect data for an MP3 frame analyzer
*/
func (l *Lame) SetAnalysis(analysis int) error {
	l.checkLgs()
	return l.setterError("lame_set_analysis", int(C.lame_set_analysis(l.lgs, C.int(analysis))))
}

func (l *Lame) GetAnalysis() int {
	l.checkLgs()
	return int(C.lame_get_analysis(l.lgs))
}

/*
  1 = write a Xing VBR header frame.
  default = 1
  this variable must have been added by a Hungarian notation Windows programmer :-)
*/
func (l *Lame) SetBWriteVbrTag(bWriteVbrTag int) error {
	l.checkLgs()
	return l.setterError("lame_set_bWriteVbrTag", int(C.lame_set_bWriteVbrTag(l.lgs, C.int(bWriteVbrTag))))
}

func (l *Lame) GetBWriteVbrTag() int {
	l.checkLgs()
	return int(C.lame_get_bWriteVbrTag(l.lgs))
}

/* 1=decode only.  use lame/mpglib to convert mp3/ogg to wav.  default=0 */
func (l *Lame) SetDecodeOnly(decodeOnly int) error {
	l.checkLgs()
	return l.setterError("lame_set_decode_only", int(C.lame_set_decode_only(l.lgs, C.int(decodeOnly))))
}

func (l *Lame) GetDecodeOnly() int {
	l.checkLgs()
	return int(C.lame_get_decode_only(l.lgs))
}

/*
  internal algorithm selection.  True quality is determined by the bitrate
  but this variable will effect quality by selecting expensive or cheap algorithms.
  quality=0..9.  0=best (very slow).  9=worst.
  recommended:  2     near-best quality, not too slow
                5     good quality, fast
                7     ok quality, really fast
*/
func (l *Lame) SetQuality(quality int) error {
	l.checkLgs()
	return l.setterError("lame_set_quality", int(C.lame_set_quality(l.lgs, C.int(quality))))
}

func (l *Lame) GetQuality() int {
	l.checkLgs()
	return int(C.lame_get_quality(l.lgs))
}

/*
  mode = 0,1,2,3 = stereo, jstereo, dual channel (not supported), mono
  default: lame picks based on compression ration and input channels
*/
func (l *Lame) SetMode(mode Mode) error {
	l.checkLgs()
	return l.setterError("lame_set_mode", int(C.lame_set_mode(l.lgs, C.MPEG_mode(mode))))
}
func (l *Lame) GetMode() Mode {
	l.checkLgs()
	return Mode(C.lame_get_mode(l.lgs))
}

/*
  force_ms.  Force M/S for all frames.  For testing only.
  default = 0 (disabled)
*/
func (l *Lame) SetForceMs(forceMs int) error {
	l.checkLgs()
	return l.setterError("lame_set_force_ms", int(C.lame_set_force_ms(l.lgs, C.int(forceMs))))
}

func (l *Lame) GetForceMs() int {
	l.checkLgs()
	return int(C.lame_get_force_ms(l.lgs))
}

/* use free_format?  default = 0 (disabled) */
func (l *Lame) SetFreeFormat(freeFormat int) error {
	l.checkLgs()
	return l.setterError("lame_set_free_format", int(C.lame_set_free_format(l.lgs, C.int(freeFormat))))
}

func (l *Lame) GetFreeFormat() int {
	l.checkLgs()
	return int(C.lame_get_free_format(l.lgs))
}

/* perform ReplayGain analysis?  default = 0 (disabled) */
func (l *Lame) SetFindReplayGain(findReplayGain int) error {
	l.checkLgs()
	return l.setterError("lame_set_findReplayGain", int(C.lame_set_findReplayGain(l.lgs, C.int(findReplayGain))))
}

func (l *Lame) GetFindReplayGain() int {
	l.checkLgs()
	return int(C.lame_get_findReplayGain(l.lgs))
}

/* decode on the fly. Search for the peak sample. If the ReplayGain
 * analysis is enabled then perform the analysis on the decoded data
 * stream. default = 0 (disabled)
 * NOTE: if this option is set the build-in decoder should not be used */
func (l *Lame) SetDecodeOnTheFly(decode_on_the_fly int) error {
	l.checkLgs()
	return l.setterError("lame_set_decode_on_the_fly", int(C.lame_set_decode_on_the_fly(l.lgs, C.int(decode_on_the_fly))))
}

func (l *Lame) GetDecodeOnTheFly() int {
	l.checkLgs()
	return int(C.lame_get_decode_on_the_fly(l.lgs))
}

/* counters for gapless encoding */
func (l *Lame) SetNogapTotal(nogapTotal int) error {
	l.checkLgs()
	return l.setterError("lame_set_nogap_total", int(C.lame_set_nogap_total(l.lgs, C.int(nogapTotal))))
}

func (l *Lame) GetNogapTotal() int {
	l.checkLgs()
	return int(C.lame_get_nogap_total(l.lgs))
}

func (l *Lame) SetNogapCurrentindex(nogapCurrentindex int) error {
	l.checkLgs()
	return l.setterError("lame_set_nogap_currentindex", int(C.lame_set_nogap_currentindex(l.lgs, C.int(nogapCurrentindex))))
}

func (l *Lame) GetNogapCurrentindex() int {
	l.checkLgs()
	return int(C.lame_get_nogap_currentindex(l.lgs))
}

/* set one of brate compression ratio.  default is compression ratio of 11.  */
func (l *Lame) SetBrate(brate int) error {
	l.checkLgs()
	return l.setterError("lame_set_brate", int(C.lame_set_brate(l.lgs, C.int(brate))))
}

func (l *Lame) GetBrate() int {
	l.checkLgs()
	return int(C.lame_get_brate(l.lgs))
}

func (l *Lame) SetCompressionRatio(compressionRatio float32) error {
	l.checkLgs()
	return l.setterError("lame_set_compression_ratio", int(C.lame_set_compression_ratio(l.lgs, C.float(compressionRatio))))
}

func (l *Lame) GetCompressionRatio() float32 {
	l.checkLgs()
	return float32(C.lame_get_compression_ratio(l.lgs))
}

func (l *Lame) SetPreset(preset int) error {
	l.checkLgs()
	return l.setterError("lame_set_preset", int(C.lame_set_preset(l.lgs, C.int(preset))))
}

func (l *Lame) SetAsmOptimizations(optim AsmOptimizations, mode int) error {
	l.checkLgs()
	return l.setterError("lame_set_asm_optimizations", int(C.lame_set_asm_optimizations(l.lgs, C.int(optim), C.int(mode))))
}

/********************************************************************
 *  frame params
 ***********************************************************************/
/* mark as copyright.  default=0 */
func (l *Lame) SetCopyright(copyright int) error {
	l.checkLgs()
	return l.setterError("lame_set_copyright", int(C.lame_set_copyright(l.lgs, C.int(copyright))))
}

func (l *Lame) GetCopyright() int {
	l.checkLgs()
	return int(C.lame_get_copyright(l.lgs))
}

/* mark as original.  default=1 */
func (l *Lame) SetOriginal(original int) error {
	l.checkLgs()
	return l.setterError("lame_set_original", int(C.lame_set_original(l.lgs, C.int(original))))
}

func (l *Lame) GetOriginal() int {
	l.checkLgs()
	return int(C.lame_get_original(l.lgs))
}

/* error_protection.  Use 2 bytes from each frame for CRC checksum. default=0 */
func (l *Lame) SetErrorProtection(errorProtection int) error {
	l.checkLgs()
	return l.setterError("lame_set_error_protection", int(C.lame_set_error_protection(l.lgs, C.int(errorProtection))))
}

func (l *Lame) GetErrorProtection() int {
	l.checkLgs()
	return int(C.lame_get_error_protection(l.lgs))
}

/* MP3 'private extension' bit  Meaningless.  default=0 */
func (l *Lame) SetExtension(extension int) error {
	l.checkLgs()
	return l.setterError("lame_set_extension", int(C.lame_set_extension(l.lgs, C.int(extension))))
}

func (l *Lame) GetExtension() int {
	l.checkLgs()
	return int(C.lame_get_extension(l.lgs))
}

/* enforce strict ISO compliance.  default=0 */
func (l *Lame) SetStrictISO(strictISO int) error {
	l.checkLgs()
	return l.setterError("lame_set_strict_ISO", int(C.lame_set_strict_ISO(l.lgs, C.int(strictISO))))
}

func (l *Lame) GetStrictISO() int {
	l.checkLgs()
	return int(C.lame_get_strict_ISO(l.lgs))
}

/********************************************************************
 * quantization/noise shaping
 ***********************************************************************/

/* disable the bit reservoir. For testing only. default=0 */
func (l *Lame) SetDisableReservoir(disable_reservoir int) error {
	l.checkLgs()
	return l.setterError("lame_set_disable_reservoir", int(C.lame_set_disable_reservoir(l.lgs, C.int(disable_reservoir))))
}

func (l *Lame) GetDisableReservoir() int {
	l.checkLgs()
	return int(C.lame_get_disable_reservoir(l.lgs))
}

/* select a different "best quantization" function. default=0  */
func (l *Lame) SetQuantComp(quantComp int) error {
	l.checkLgs()
	return l.setterError("lame_set_quant_comp", int(C.lame_set_quant_comp(l.lgs, C.int(quantComp))))
}

func (l *Lame) GetQuantComp() int {
	l.checkLgs()
	return int(C.lame_get_quant_comp(l.lgs))
}

func (l *Lame) SetQuantCompShort(quantCompShort int) error {
	l.checkLgs()
	return l.setterError("lame_set_quant_comp_short", int(C.lame_set_quant_comp_short(l.lgs, C.int(quantCompShort))))
}

func (l *Lame) GetQuantCompShort() int {
	l.checkLgs()
	return int(C.lame_get_quant_comp_short(l.lgs))
}

func (l *Lame) SetExperimentalX(experimentalX int) error {
	l.checkLgs()
	return l.setterError("lame_set_experimentalX", int(C.lame_set_experimentalX(l.lgs, C.int(experimentalX))))
}

/* compatibility*/
func (l *Lame) GetExperimentalX() int {
	l.checkLgs()
	return int(C.lame_get_experimentalX(l.lgs))
}

/* another experimental option.  for testing only */
func (l *Lame) SetExperimentalY(experimentalY int) error {
	l.checkLgs()
	return l.setterError("lame_set_experimentalY", int(C.lame_set_experimentalY(l.lgs, C.int(experimentalY))))
}

func (l *Lame) GetExperimentalY() int {
	l.checkLgs()
	return int(C.lame_get_experimentalY(l.lgs))
}

/* another experimental option.  for testing only */
func (l *Lame) SetExperimentalZ(experimentalZ int) error {
	l.checkLgs()
	return l.setterError("lame_set_experimentalZ", int(C.lame_set_experimentalZ(l.lgs, C.int(experimentalZ))))
}

func (l *Lame) GetExperimentalZ() int {
	l.checkLgs()
	return int(C.lame_get_experimentalZ(l.lgs))
}

/* Naoki's psycho acoustic model.  default=0 */
func (l *Lame) SetExpNspsytune(expNspsytune int) error {
	l.checkLgs()
	return l.setterError("lame_set_exp_nspsytune", int(C.lame_set_exp_nspsytune(l.lgs, C.int(expNspsytune))))
}

func (l *Lame) GetExpNspsytune() int {
	l.checkLgs()
	return int(C.lame_get_exp_nspsytune(l.lgs))
}

// void lame_set_msfix(lame_global_flags *, double);
func (l *Lame) SetMsfix(msfix float32) {
	l.checkLgs()
	C.lame_set_msfix(l.lgs, C.double(msfix))
}

func (l *Lame) GetMsfix() float32 {
	l.checkLgs()
	return float32(C.lame_get_msfix(l.lgs))
}

/* VBR stuff */
/* Types of VBR.  default = VBR_OFF = CBR */
func (l *Lame) SetVBR(vbr VBRMode) error {
	l.checkLgs()
	return l.setterError("lame_set_VBR", int(C.lame_set_VBR(l.lgs, C.vbr_mode(vbr))))
}

func (l *Lame) GetVBR() VBRMode {
	l.checkLgs()
	return VBRMode(C.lame_get_VBR(l.lgs))
}

/* VBR quality level.  0=highest  9=lowest  */
func (l *Lame) SetVBRQ(VBR_q int) error {
	l.checkLgs()
	return l.setterError("lame_set_VBR_q", int(C.lame_set_VBR_q(l.lgs, C.int(VBR_q))))
}

func (l *Lame) GetVBRQ() int {
	l.checkLgs()
	return int(C.lame_get_VBR_q(l.lgs))
}

/* VBR quality level.  0=highest  9=lowest, Range [0,...,10[  */
func (l *Lame) SetVBRQuality(VBRQuality float32) error {
	l.checkLgs()
	return l.setterError("lame_set_VBR_quality", int(C.lame_set_VBR_quality(l.lgs, C.float(VBRQuality))))
}

func (l *Lame) GetVBRQuality() float32 {
	l.checkLgs()
	return float32(C.lame_get_VBR_quality(l.lgs))
}

/* Ignored except for VBR=vbr_abr (ABR mode) */
func (l *Lame) SetVBRMeanBitrateKbps(VBRMeanBitrateKbps int) error {
	l.checkLgs()
	return l.setterError("lame_set_VBR_mean_bitrate_kbps", int(C.lame_set_VBR_mean_bitrate_kbps(l.lgs, C.int(VBRMeanBitrateKbps))))
}

func (l *Lame) GetVBRMeanBitrateKbps() int {
	l.checkLgs()
	return int(C.lame_get_VBR_mean_bitrate_kbps(l.lgs))
}

func (l *Lame) SetVBRMinBitrateKbps(VBRMinBitrateKbps int) error {
	l.checkLgs()
	return l.setterError("lame_set_VBR_min_bitrate_kbps", int(C.lame_set_VBR_min_bitrate_kbps(l.lgs, C.int(VBRMinBitrateKbps))))
}

func (l *Lame) GetVBRMinBitrateKbps() int {
	l.checkLgs()
	return int(C.lame_get_VBR_min_bitrate_kbps(l.lgs))
}

func (l *Lame) SetVBRMaxBitrateKbps(VBRMaxBitrateKbps int) error {
	l.checkLgs()
	return l.setterError("lame_set_VBR_max_bitrate_kbps", int(C.lame_set_VBR_max_bitrate_kbps(l.lgs, C.int(VBRMaxBitrateKbps))))
}

func (l *Lame) GetVBRMaxBitrateKbps() int {
	l.checkLgs()
	return int(C.lame_get_VBR_max_bitrate_kbps(l.lgs))
}

/*
  1=strictly enforce VBR_min_bitrate.  Normally it will be violated for
  analog silence
*/
func (l *Lame) SetVBRHardMin(VBRHardMin int) error {
	l.checkLgs()
	return l.setterError("lame_set_VBR_hard_min", int(C.lame_set_VBR_hard_min(l.lgs, C.int(VBRHardMin))))
}

func (l *Lame) GetVBRHardMin() int {
	l.checkLgs()
	return int(C.lame_get_VBR_hard_min(l.lgs))
}

/* filtering... */
/* freq in Hz to apply lowpass. Default = 0 = lame chooses.  -1 = disabled */
func (l *Lame) SetLowpassfreq(lowpassfreq int) error {
	l.checkLgs()
	return l.setterError("lame_set_lowpassfreq", int(C.lame_set_lowpassfreq(l.lgs, C.int(lowpassfreq))))
}

func (l *Lame) GetLowpassfreq() int {
	l.checkLgs()
	return int(C.lame_get_lowpassfreq(l.lgs))
}

/* width of transition band, in Hz.  Default = one polyphase filter band */
func (l *Lame) SetLowpasswidth(lowpasswidth int) error {
	l.checkLgs()
	return l.setterError("lame_set_lowpasswidth", int(C.lame_set_lowpasswidth(l.lgs, C.int(lowpasswidth))))
}

func (l *Lame) GetLowpasswidth() int {
	l.checkLgs()
	return int(C.lame_get_lowpasswidth(l.lgs))
}

/* freq in Hz to apply highpass. Default = 0 = lame chooses.  -1 = disabled */
func (l *Lame) SetHighpassfreq(highpassfreq int) error {
	l.checkLgs()
	return l.setterError("lame_set_highpassfreq", int(C.lame_set_highpassfreq(l.lgs, C.int(highpassfreq))))
}

func (l *Lame) GetHighpassfreq() int {
	l.checkLgs()
	return int(C.lame_get_highpassfreq(l.lgs))
}

/* width of transition band, in Hz.  Default = one polyphase filter band */
func (l *Lame) SetHighpasswidth(highpasswidth int) error {
	l.checkLgs()
	return l.setterError("lame_set_highpasswidth", int(C.lame_set_highpasswidth(l.lgs, C.int(highpasswidth))))
}

func (l *Lame) GetHighpasswidth() int {
	l.checkLgs()
	return int(C.lame_get_highpasswidth(l.lgs))
}

/* only use ATH for masking */
func (l *Lame) SetATHonly(ATHonly int) error {
	l.checkLgs()
	return l.setterError("lame_set_ATHonly", int(C.lame_set_ATHonly(l.lgs, C.int(ATHonly))))
}

func (l *Lame) GetATHonly() int {
	l.checkLgs()
	return int(C.lame_get_ATHonly(l.lgs))
}

/* only use ATH for short blocks */
func (l *Lame) SetATHshort(ATHshort int) error {
	l.checkLgs()
	return l.setterError("lame_set_ATHshort", int(C.lame_set_ATHshort(l.lgs, C.int(ATHshort))))
}

func (l *Lame) GetATHshort() int {
	l.checkLgs()
	return int(C.lame_get_ATHshort(l.lgs))
}

/* disable ATH */
func (l *Lame) SetNoATH(noATH int) error {
	l.checkLgs()
	return l.setterError("lame_set_noATH", int(C.lame_set_noATH(l.lgs, C.int(noATH))))
}

func (l *Lame) GetNoATH() int {
	l.checkLgs()
	return int(C.lame_get_noATH(l.lgs))
}

/* select ATH formula */
func (l *Lame) SetATHtype(ATHtype int) error {
	l.checkLgs()
	return l.setterError("lame_set_ATHtype", int(C.lame_set_ATHtype(l.lgs, C.int(ATHtype))))
}

func (l *Lame) GetATHtype() int {
	l.checkLgs()
	return int(C.lame_get_ATHtype(l.lgs))
}

/* lower ATH by this many db */
func (l *Lame) SetATHlower(ATHlower float32) error {
	l.checkLgs()
	return l.setterError("lame_set_ATHlower", int(C.lame_set_ATHlower(l.lgs, C.float(ATHlower))))
}

func (l *Lame) GetATHlower() float32 {
	l.checkLgs()
	return float32(C.lame_get_ATHlower(l.lgs))
}

/* select ATH adaptive adjustment type */
func (l *Lame) SetAthaaType(athaaType int) error {
	l.checkLgs()
	return l.setterError("lame_set_athaa_type", int(C.lame_set_athaa_type(l.lgs, C.int(athaaType))))
}

func (l *Lame) GetAthaaType() int {
	l.checkLgs()
	return int(C.lame_get_athaa_type(l.lgs))
}

/* adjust (in dB) the point below which adaptive ATH level adjustment occurs */
func (l *Lame) SetAthaaSensitivity(athaaSensitivity float32) error {
	l.checkLgs()
	return l.setterError("lame_set_athaa_sensitivity", int(C.lame_set_athaa_sensitivity(l.lgs, C.float(athaaSensitivity))))
}

func (l *Lame) GetAthaaSensitivity() float32 {
	l.checkLgs()
	return float32(C.lame_get_athaa_type(l.lgs))
}

/*
  allow blocktypes to differ between channels?
  default: 0 for jstereo, 1 for stereo
*/
func (l *Lame) SetAllowDiffShort(allowDiffShort int) error {
	l.checkLgs()
	return l.setterError("lame_set_allow_diff_short", int(C.lame_set_allow_diff_short(l.lgs, C.int(allowDiffShort))))
}

func (l *Lame) GetAllowDiffShort() int {
	l.checkLgs()
	return int(C.lame_get_allow_diff_short(l.lgs))
}

/* use temporal masking effect (default = 1) */
func (l *Lame) SetUseTemporal(useTemporal int) error {
	l.checkLgs()
	return l.setterError("lame_set_useTemporal", int(C.lame_set_useTemporal(l.lgs, C.int(useTemporal))))
}

func (l *Lame) GetUseTemporal() int {
	l.checkLgs()
	return int(C.lame_get_useTemporal(l.lgs))
}

/* use temporal masking effect (default = 1) */
func (l *Lame) SetInterChRatio(interChRatio float32) error {
	l.checkLgs()
	return l.setterError("lame_set_interChRatio", int(C.lame_set_interChRatio(l.lgs, C.float(interChRatio))))
}

func (l *Lame) GetInterChRatio() float32 {
	l.checkLgs()
	return float32(C.lame_get_interChRatio(l.lgs))
}

/* disable short blocks */
func (l *Lame) SetNoShortBlocks(noShortBlocks int) error {
	l.checkLgs()
	return l.setterError("lame_set_no_short_blocks", int(C.lame_set_no_short_blocks(l.lgs, C.int(noShortBlocks))))
}

func (l *Lame) GetNoShortBlocks() int {
	l.checkLgs()
	return int(C.lame_get_no_short_blocks(l.lgs))
}

/* force short blocks */
func (l *Lame) SetForceShortBlocks(forceShortBlocks int) error {
	l.checkLgs()
	return l.setterError("lame_set_force_short_blocks", int(C.lame_set_force_short_blocks(l.lgs, C.int(forceShortBlocks))))
}

func (l *Lame) GetForceShortBlocks() int {
	l.checkLgs()
	return int(C.lame_get_force_short_blocks(l.lgs))
}

/* Input PCM is emphased PCM (for instance from one of the rarely
   emphased CDs), it is STRONGLY not recommended to use this, because
   psycho does not take it into account, and last but not least many decoders
   ignore these bits */
func (l *Lame) SetEmphasis(emphasis int) error {
	l.checkLgs()
	return l.setterError("lame_set_emphasis", int(C.lame_set_emphasis(l.lgs, C.int(emphasis))))
}

func (l *Lame) GetEmphasis() int {
	l.checkLgs()
	return int(C.lame_get_emphasis(l.lgs))
}

/* mp3 version  0=MPEG-2  1=MPEG-1  (2=MPEG-2.5)     */
func (l *Lame) GetVersion() int {
	l.checkLgs()
	return int(C.lame_get_version(l.lgs))
}

/* encoder delay   */
func (l *Lame) GetEncoderDelay() int {
	l.checkLgs()
	return int(C.lame_get_encoder_delay(l.lgs))
}

/*
  padding appended to the input to make sure decoder can fully decode
  all input.  Note that this value can only be calculated during the
  call to lame_encoder_flush().  Before lame_encoder_flush() has
  been called, the value of encoder_padding = 0.
*/
func (l *Lame) GetEncoderPadding() int {
	l.checkLgs()
	return int(C.lame_get_encoder_padding(l.lgs))
}

/* size of MPEG frame */
func (l *Lame) GetFramesize() int {
	l.checkLgs()
	return int(C.lame_get_framesize(l.lgs))
}

/* number of PCM samples buffered, but not yet encoded to mp3 data. */
func (l *Lame) GetMfSamplesToEncode() int {
	l.checkLgs()
	return int(C.lame_get_mf_samples_to_encode(l.lgs))
}

/*
  size (bytes) of mp3 data buffered, but not yet encoded.
  ref: https://github.com/gypified/libmp3lame/blob/master/include/lame.h#L594
*/
func (l *Lame) GetSizeMp3buffer() int {
	l.checkLgs()
	return int(C.lame_get_size_mp3buffer(l.lgs))
}

/* number of frames encoded so far */
func (l *Lame) GetFrameNum() int {
	l.checkLgs()
	return int(C.lame_get_frameNum(l.lgs))
}

/*
  lame's estimate of the total number of frames to be encoded
   only valid if calling program set num_samples
*/
func (l *Lame) GetTotalframes() int {
	l.checkLgs()
	return int(C.lame_get_totalframes(l.lgs))
}

/* RadioGain value. Multiplied by 10 and rounded to the nearest. */
func (l *Lame) GetRadioGain() int {
	l.checkLgs()
	return int(C.lame_get_RadioGain(l.lgs))
}

/* AudiophileGain value. Multipled by 10 and rounded to the nearest. */
func (l *Lame) GetAudiophileGain() int {
	l.checkLgs()
	return int(C.lame_get_AudiophileGain(l.lgs))
}

/* the peak sample */
func (l *Lame) GetPeakSample() float32 {
	l.checkLgs()
	return float32(C.lame_get_PeakSample(l.lgs))
}

/* Gain change required for preventing clipping.
   ref: https://github.com/gypified/libmp3lame/blob/master/include/lame.h#L623
*/
func (l *Lame) GetNoclipGainChange() int {
	l.checkLgs()
	return int(C.lame_get_noclipGainChange(l.lgs))
}

/*
 * ref: https://github.com/gypified/libmp3lame/blob/master/include/lame.h#L623
 */
func (l *Lame) GetNoclipScale() float32 {
	l.checkLgs()
	return float32(C.lame_get_noclipScale(l.lgs))
}
