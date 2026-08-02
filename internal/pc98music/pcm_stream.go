package pc98music

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/wicanr2/golden-box-remake-engine/audio/pcm"
	"github.com/wicanr2/golden-box-remake-engine/audio/ym2203"
	"github.com/wicanr2/golden-box-remake-engine/audio/ym2203/ymfm"
)

const ymfmNativeClockDivisor = 24

// GameMusicStartDelay is the delay used by GAME.EXE MSCPLAY after MSCSTOP and
// before the new selector is sent through interrupt vector 7Eh.
const GameMusicStartDelay = 800 * time.Millisecond

// TrackPCMStream renders one verified PC-98 track as signed 16-bit little
// endian stereo PCM. It implements io.Reader for renderer-side audio players.
type TrackPCMStream struct {
	mu       sync.Mutex
	playback *TrackPlayback
	renderer *YM2203EventRenderer
	synth    *ymfm.Synth
	resample *pcm.LinearResampler

	timerValue  byte
	silentBytes int64
	pending     []byte
	failed      error
	closed      bool
	selector    int
	outputRate  uint32
	emitted     uint64
	history     []byte
}

const trackPCMStreamSnapshotVersion = 1

type TrackPCMStreamSnapshot struct {
	Version          int                         `json:"version"`
	Selector         int                         `json:"selector"`
	OutputSampleRate uint32                      `json:"output_sample_rate"`
	Playback         TrackPlaybackSnapshot       `json:"playback"`
	Renderer         YM2203RendererSnapshot      `json:"renderer"`
	SynthState       []byte                      `json:"synth_state"`
	Resampler        pcm.LinearResamplerSnapshot `json:"resampler"`
	TimerValue       byte                        `json:"timer_value"`
	SilentBytes      int64                       `json:"silent_bytes"`
	Pending          []byte                      `json:"pending,omitempty"`
}

// ValidatePersistent applies allocation and identity bounds before a snapshot
// from JSON is retained by game state. Driver-specific sequence identities are
// validated again by Restore against the locally supplied MSCDRV.EXE.
func (snapshot TrackPCMStreamSnapshot) ValidatePersistent() error {
	if snapshot.Version != trackPCMStreamSnapshotVersion {
		return fmt.Errorf("unsupported PC-98 PCM snapshot version %d", snapshot.Version)
	}
	if snapshot.Selector < 1 || snapshot.Selector > driverPublicTracks {
		return fmt.Errorf("PC-98 PCM selector %d outside 1..%d", snapshot.Selector, driverPublicTracks)
	}
	if snapshot.OutputSampleRate < 8_000 || snapshot.OutputSampleRate > 192_000 {
		return fmt.Errorf("PC-98 PCM output rate %d outside 8000..192000", snapshot.OutputSampleRate)
	}
	if snapshot.Playback.Version != playbackSnapshotVersion {
		return fmt.Errorf("unsupported track playback snapshot version %d", snapshot.Playback.Version)
	}
	if len(snapshot.SynthState) == 0 || len(snapshot.SynthState) > 1<<20 {
		return fmt.Errorf("PC-98 YM2203 state size %d outside 1..1048576", len(snapshot.SynthState))
	}
	maxBuffered := int(snapshot.OutputSampleRate) * 4 * 8
	if snapshot.SilentBytes < 0 || snapshot.SilentBytes > int64(snapshot.OutputSampleRate)*4 || len(snapshot.Pending) > maxBuffered {
		return fmt.Errorf("PC-98 PCM snapshot buffers are outside bounded limits")
	}
	for channel, state := range snapshot.Playback.Channels {
		if len(state.Machine.CallStack) > 16 || len(state.Machine.LoopStack) > 16 {
			return fmt.Errorf("PC-98 channel %d sequence stack exceeds 16 entries", channel)
		}
	}
	return nil
}

// NewGameTrackPCMStream opens a selector with the GAME.EXE MSCPLAY transition
// delay. The silence does not advance the driver's Timer B playback state.
func NewGameTrackPCMStream(
	driver []byte, selector int, outputSampleRate uint32,
) (*TrackPCMStream, error) {
	stream, err := NewTrackPCMStream(driver, selector, outputSampleRate)
	if err != nil {
		return nil, err
	}
	stream.prependSilence(GameMusicStartDelay, outputSampleRate)
	return stream, nil
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
	stream, err := newTrackPCMStream(playback, renderer, initial, outputSampleRate)
	if err != nil {
		return nil, err
	}
	stream.selector = selector
	return stream, nil
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
		playback:   playback,
		renderer:   renderer,
		synth:      synth,
		resample:   resampler,
		outputRate: outputSampleRate,
	}
	if err := stream.apply(initial); err != nil {
		stream.Close()
		return nil, err
	}
	return stream, nil
}

func (stream *TrackPCMStream) prependSilence(duration time.Duration, sampleRate uint32) {
	frames := int64(duration) * int64(sampleRate) / int64(time.Second)
	stream.silentBytes += frames * 4
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
	if stream == nil {
		return 0, io.EOF
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if len(destination) == 0 {
		return 0, nil
	}
	if stream == nil || stream.closed {
		return 0, io.EOF
	}
	if stream.silentBytes != 0 {
		count := int64(len(destination))
		if count > stream.silentBytes {
			count = stream.silentBytes
		}
		clear(destination[:int(count)])
		stream.silentBytes -= count
		stream.recordEmitted(destination[:int(count)])
		return int(count), nil
	}
	for len(stream.pending) < len(destination) && stream.failed == nil {
		stream.failed = stream.renderPeriod()
	}
	count := copy(destination, stream.pending)
	stream.pending = stream.pending[count:]
	if count != 0 {
		stream.recordEmitted(destination[:count])
		return count, nil
	}
	if stream.failed != nil {
		return 0, stream.failed
	}
	return 0, io.EOF
}

// Close releases the native synthesizer.
func (stream *TrackPCMStream) Close() error {
	if stream == nil {
		return nil
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if stream == nil || stream.closed {
		return nil
	}
	stream.closed = true
	stream.silentBytes = 0
	stream.pending = nil
	stream.history = nil
	if stream.synth != nil {
		stream.synth.Close()
	}
	return nil
}

func (stream *TrackPCMStream) recordEmitted(data []byte) {
	stream.emitted += uint64(len(data))
	limit := int(stream.outputRate) * 4 * 4
	if limit <= 0 {
		return
	}
	if len(data) >= limit {
		stream.history = append(stream.history[:0], data[len(data)-limit:]...)
		return
	}
	if excess := len(stream.history) + len(data) - limit; excess > 0 {
		copy(stream.history, stream.history[excess:])
		stream.history = stream.history[:len(stream.history)-excess]
	}
	stream.history = append(stream.history, data...)
}

// SnapshotAtFrame captures the decoder at its current read-ahead point while
// prepending already emitted-but-not-yet-audible PCM from the bounded history.
// frame is the renderer player's actual stereo sample-frame position.
func (stream *TrackPCMStream) SnapshotAtFrame(frame uint64) (TrackPCMStreamSnapshot, error) {
	if stream == nil {
		return TrackPCMStreamSnapshot{}, fmt.Errorf("PC-98 PCM stream is nil")
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if frame > ^uint64(0)/4 {
		return TrackPCMStreamSnapshot{}, fmt.Errorf("PCM frame position %d overflows byte offset", frame)
	}
	audible := frame * 4
	if audible > stream.emitted {
		return TrackPCMStreamSnapshot{}, fmt.Errorf("audible PCM byte %d exceeds emitted byte %d", audible, stream.emitted)
	}
	historyStart := stream.emitted - uint64(len(stream.history))
	if audible < historyStart {
		return TrackPCMStreamSnapshot{}, fmt.Errorf("audible PCM byte %d precedes retained history %d", audible, historyStart)
	}
	backlog := append([]byte(nil), stream.history[audible-historyStart:]...)
	return stream.snapshotLocked(backlog)
}

// Snapshot captures the stream at the exact source position without player
// read-ahead. Tests and non-buffering adapters use this form.
func (stream *TrackPCMStream) Snapshot() (TrackPCMStreamSnapshot, error) {
	if stream == nil {
		return TrackPCMStreamSnapshot{}, fmt.Errorf("PC-98 PCM stream is nil")
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	return stream.snapshotLocked(nil)
}

func (stream *TrackPCMStream) snapshotLocked(prefix []byte) (TrackPCMStreamSnapshot, error) {
	if stream == nil || stream.closed || stream.failed != nil {
		return TrackPCMStreamSnapshot{}, fmt.Errorf("PC-98 PCM stream is not snapshotable")
	}
	playback, err := stream.playback.Snapshot()
	if err != nil {
		return TrackPCMStreamSnapshot{}, err
	}
	renderer, err := stream.renderer.Snapshot()
	if err != nil {
		return TrackPCMStreamSnapshot{}, err
	}
	synthState, err := stream.synth.Snapshot()
	if err != nil {
		return TrackPCMStreamSnapshot{}, err
	}
	resampler, err := stream.resample.Snapshot()
	if err != nil {
		return TrackPCMStreamSnapshot{}, err
	}
	pending := append(prefix, stream.pending...)
	return TrackPCMStreamSnapshot{
		Version: trackPCMStreamSnapshotVersion, Selector: stream.selector,
		OutputSampleRate: stream.outputRate, Playback: playback, Renderer: renderer,
		SynthState: synthState, Resampler: resampler, TimerValue: stream.timerValue,
		SilentBytes: stream.silentBytes, Pending: pending,
	}, nil
}

// restore replaces a freshly constructed stream with a validated snapshot.
// It remains private so a failed restore cannot expose a partially mutated
// live decoder; RestoreGameTrackPCMStream discards the fresh stream on error.
func (stream *TrackPCMStream) restore(snapshot TrackPCMStreamSnapshot) error {
	if stream == nil {
		return fmt.Errorf("PC-98 PCM stream is nil")
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if stream == nil || stream.closed {
		return fmt.Errorf("PC-98 PCM stream is closed")
	}
	if snapshot.Version != trackPCMStreamSnapshotVersion {
		return fmt.Errorf("unsupported PC-98 PCM snapshot version %d", snapshot.Version)
	}
	if snapshot.Selector != stream.selector || snapshot.OutputSampleRate != stream.outputRate {
		return fmt.Errorf("PC-98 PCM snapshot stream identity does not match selector/rate")
	}
	maxBuffered := int(stream.outputRate) * 4 * 8
	if snapshot.SilentBytes < 0 || snapshot.SilentBytes > int64(stream.outputRate)*4 || len(snapshot.Pending) > maxBuffered {
		return fmt.Errorf("PC-98 PCM snapshot buffers are outside bounded limits")
	}
	if err := stream.playback.Restore(snapshot.Playback); err != nil {
		return err
	}
	if err := stream.renderer.Restore(snapshot.Renderer); err != nil {
		return err
	}
	if err := stream.synth.Restore(snapshot.SynthState); err != nil {
		return err
	}
	if err := stream.resample.Restore(snapshot.Resampler); err != nil {
		return err
	}
	stream.timerValue = snapshot.TimerValue
	stream.silentBytes = snapshot.SilentBytes
	stream.pending = append(stream.pending[:0], snapshot.Pending...)
	stream.failed = nil
	stream.emitted = 0
	stream.history = nil
	return nil
}

func RestoreGameTrackPCMStream(driver []byte, snapshot TrackPCMStreamSnapshot) (*TrackPCMStream, error) {
	if err := snapshot.ValidatePersistent(); err != nil {
		return nil, err
	}
	stream, err := NewGameTrackPCMStream(driver, snapshot.Selector, snapshot.OutputSampleRate)
	if err != nil {
		return nil, err
	}
	if err := stream.restore(snapshot); err != nil {
		_ = stream.Close()
		return nil, err
	}
	return stream, nil
}
