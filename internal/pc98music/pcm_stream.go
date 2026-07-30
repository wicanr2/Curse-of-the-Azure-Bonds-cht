package pc98music

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/wicanr2/golden-box-remake-engine/audio/pcm"
	"github.com/wicanr2/golden-box-remake-engine/audio/ym2203"
	"github.com/wicanr2/golden-box-remake-engine/audio/ym2203/ymfm"
)

const ymfmNativeClockDivisor = 24

// TrackPCMStream renders one verified PC-98 track as signed 16-bit little
// endian stereo PCM. It implements io.Reader for renderer-side audio players.
type TrackPCMStream struct {
	playback *TrackPlayback
	renderer *YM2203EventRenderer
	synth    *ymfm.Synth
	resample *pcm.LinearResampler

	timerValue byte
	pending    []byte
	failed     error
	closed     bool
}

// NewTrackPCMStream opens a selector from the user's exact local MSCDRV.EXE.
func NewTrackPCMStream(
	driver []byte, selector int, outputSampleRate uint32,
) (*TrackPCMStream, error) {
	playback, initial, err := NewTrackPlayback(driver, selector)
	if err != nil {
		return nil, err
	}
	renderer, err := NewYM2203EventRenderer(driver)
	if err != nil {
		return nil, err
	}
	return newTrackPCMStream(playback, renderer, initial, outputSampleRate)
}

func newTrackPCMStream(
	playback *TrackPlayback,
	renderer *YM2203EventRenderer,
	initial []MusicEvent,
	outputSampleRate uint32,
) (*TrackPCMStream, error) {
	if playback == nil {
		return nil, fmt.Errorf("PC-98 track playback is nil")
	}
	if renderer == nil {
		return nil, fmt.Errorf("PC-98 register renderer is nil")
	}
	synth, err := ymfm.New(uint32(PC98YM2203ClockHz))
	if err != nil {
		return nil, err
	}
	resampler, err := pcm.NewLinearResampler(
		uint64(synth.NativeSampleRate()),
		uint64(outputSampleRate),
	)
	if err != nil {
		synth.Close()
		return nil, err
	}
	stream := &TrackPCMStream{
		playback: playback,
		renderer: renderer,
		synth:    synth,
		resample: resampler,
	}
	if err := stream.apply(initial); err != nil {
		stream.Close()
		return nil, err
	}
	return stream, nil
}

func (stream *TrackPCMStream) apply(events []MusicEvent) error {
	writes, err := stream.renderer.Render(events)
	if err != nil {
		return err
	}
	for _, write := range writes {
		if write.Register == 0x26 {
			stream.timerValue = write.Value
		}
		if err := stream.synth.Write(write.Register, write.Value); err != nil {
			return err
		}
	}
	return nil
}

func (stream *TrackPCMStream) renderPeriod() error {
	cycles, err := ym2203.TimerBClockCycles(
		stream.timerValue, PC98YM2203DefaultPrescale,
	)
	if err != nil {
		return err
	}
	if cycles%ymfmNativeClockDivisor != 0 {
		return fmt.Errorf(
			"YM2203 Timer B period %d is not divisible by native clock divisor %d",
			cycles, ymfmNativeClockDivisor,
		)
	}
	native, err := stream.synth.Generate(int(cycles / ymfmNativeClockDivisor))
	if err != nil {
		return err
	}
	output, err := stream.resample.Append(nil, native)
	if err != nil {
		return err
	}
	bytes := make([]byte, len(output)*4)
	for index, sample := range output {
		// ymfm exposes FM plus three SSG outputs. Averaging the four paths
		// prevents arithmetic clipping; analog PC-98 mixer gain remains a
		// separate title-oracle calibration.
		value := int64(sample) / 4
		if value > 32767 {
			value = 32767
		} else if value < -32768 {
			value = -32768
		}
		pcmValue := uint16(int16(value))
		binary.LittleEndian.PutUint16(bytes[index*4:], pcmValue)
		binary.LittleEndian.PutUint16(bytes[index*4+2:], pcmValue)
	}
	stream.pending = append(stream.pending, bytes...)

	events, err := stream.playback.Tick(4096)
	if err != nil {
		return err
	}
	return stream.apply(events)
}

// Read emits an endless looping track until the source interpreter reports an
// error or Close is called.
func (stream *TrackPCMStream) Read(destination []byte) (int, error) {
	if len(destination) == 0 {
		return 0, nil
	}
	if stream == nil || stream.closed {
		return 0, io.EOF
	}
	for len(stream.pending) < len(destination) && stream.failed == nil {
		stream.failed = stream.renderPeriod()
	}
	count := copy(destination, stream.pending)
	stream.pending = stream.pending[count:]
	if count != 0 {
		return count, nil
	}
	if stream.failed != nil {
		return 0, stream.failed
	}
	return 0, io.EOF
}

// Close releases the native synthesizer.
func (stream *TrackPCMStream) Close() error {
	if stream == nil || stream.closed {
		return nil
	}
	stream.closed = true
	stream.pending = nil
	if stream.synth != nil {
		stream.synth.Close()
	}
	return nil
}
