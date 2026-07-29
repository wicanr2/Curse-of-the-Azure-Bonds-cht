package pc98music

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"sort"
)

type MusicEventKind string

const (
	EventRegisterWrite     MusicEventKind = "register_write"
	EventSetVolume         MusicEventKind = "set_volume"
	EventSetParameterBlock MusicEventKind = "set_parameter_block"
)

// MusicEvent is one side effect proven in MSCDRV.EXE sub_10410 or sub_10253.
// Tick zero contains track initialization; later ticks correspond to one
// invocation of the timer-driven stream interpreter.
type MusicEvent struct {
	Tick         uint64         `json:"tick"`
	Channel      int            `json:"channel"`
	Kind         MusicEventKind `json:"kind"`
	Register     byte           `json:"register,omitempty"`
	Value        byte           `json:"value,omitempty"`
	Parameter    byte           `json:"parameter,omitempty"`
	StreamOffset int            `json:"stream_offset,omitempty"`
	Opcode       byte           `json:"opcode,omitempty"`
}

type playbackChannel struct {
	machine         *SequenceMachine
	duration        uint16
	baseVolume      uint16
	currentVolume   uint16
	envelopePointer int
	mode            byte
}

// TrackPlayback deterministically executes the normal (non-fade, non-SFX)
// path of one imported seven-channel track.
type TrackPlayback struct {
	driverData []byte
	channels   [driverTrackChannels]playbackChannel
	registers  [256]byte
	tick       uint64
	counter    uint16
	tempo      uint16
}

type PlaybackAudit struct {
	Selector                   int    `json:"selector"`
	Ticks                      int    `json:"ticks"`
	EventCount                 int    `json:"event_count"`
	EventSHA256                string `json:"event_sha256"`
	ParameterIndices           []int  `json:"parameter_indices"`
	EmbeddedParametersComplete bool   `json:"embedded_parameters_complete"`
}

func NewTrackPlayback(driver []byte, selector int) (*TrackPlayback, []MusicEvent, error) {
	sum := sha256.Sum256(driver)
	if hex.EncodeToString(sum[:]) != DriverSHA256 {
		return nil, nil, fmt.Errorf("MSCDRV.EXE SHA-256 does not match known input")
	}
	if selector < 1 || selector > driverPublicTracks {
		return nil, nil, fmt.Errorf("selector %d is outside 1..%d", selector, driverPublicTracks)
	}
	tracks, err := auditTrackDescriptors(driver)
	if err != nil {
		return nil, nil, err
	}
	return newTrackPlayback(driver[driverDataFileBase:], tracks[selector-1])
}

func newTrackPlayback(driverData []byte, track TrackDescriptor) (*TrackPlayback, []MusicEvent, error) {
	playback := &TrackPlayback{
		driverData: append([]byte(nil), driverData...),
		tempo:      uint16(track.HeaderWords[1]),
	}
	initial := make([]MusicEvent, 0, 16)
	playback.writeRegister(&initial, -1, 0x26, byte(playback.tempo), 0, 0)
	for channel, descriptor := range track.Channels {
		end := descriptor.SequenceOffset + descriptor.SequenceLength
		if channel == 6 {
			end = len(playback.driverData)
		}
		machine, err := NewSequenceMachine(channel, descriptor.SequenceOffset, end)
		if err != nil {
			return nil, nil, fmt.Errorf("channel %d: %w", channel, err)
		}
		state := &playback.channels[channel]
		state.machine = machine
		switch {
		case channel < 3:
			initial = append(initial,
				MusicEvent{
					Channel: channel, Kind: EventSetParameterBlock,
					Parameter: byte(descriptor.RawParameter1),
				},
				MusicEvent{
					Channel: channel, Kind: EventSetVolume,
					Value: byte(descriptor.RawParameter2),
				},
			)
			playback.writeRegister(&initial, channel, 0x28, byte(channel), 0, 0)
		case channel < 6:
			state.envelopePointer = descriptor.RawParameter1<<5 + 0x10
			state.baseVolume = uint16(descriptor.RawParameter2)
			state.currentVolume = state.baseVolume
			playback.writeRegister(
				&initial, channel, byte(channel+5), byte(state.currentVolume), 0, 0,
			)
		default:
			state.baseVolume = uint16(descriptor.RawParameter2)
			state.currentVolume = state.baseVolume
		}
	}
	playback.writeRegister(&initial, -1, 0x07, 0xB8, 0, 0)
	return playback, initial, nil
}

func (playback *TrackPlayback) writeRegister(
	events *[]MusicEvent,
	channel int,
	register, value byte,
	offset int,
	opcode byte,
) {
	playback.registers[register] = value
	*events = append(*events, MusicEvent{
		Tick: playback.tick, Channel: channel, Kind: EventRegisterWrite,
		Register: register, Value: value, StreamOffset: offset, Opcode: opcode,
	})
}

func (playback *TrackPlayback) word(offset int) (uint16, error) {
	if offset < 0 || offset+2 > len(playback.driverData) {
		return 0, fmt.Errorf("driver data word 0x%X is outside input", offset)
	}
	return uint16(playback.driverData[offset]) |
		uint16(playback.driverData[offset+1])<<8, nil
}

// Tick advances all seven channels once, matching the normal timer path.
func (playback *TrackPlayback) Tick(maxCommands int) ([]MusicEvent, error) {
	playback.tick++
	playback.counter++
	events := make([]MusicEvent, 0, 16)
	for channel := range playback.channels {
		state := &playback.channels[channel]
		if state.duration != 0 {
			state.duration--
			if channel >= 3 && channel < 6 {
				if err := playback.sustainPSG(channel, state, &events); err != nil {
					return nil, err
				}
			}
			continue
		}
		commands, err := state.machine.NextTimed(playback.driverData, maxCommands)
		if err != nil {
			return nil, fmt.Errorf("channel %d tick %d: %w", channel, playback.tick, err)
		}
		for _, command := range commands {
			if err := playback.applyCommand(channel, state, command, &events); err != nil {
				return nil, fmt.Errorf(
					"channel %d stream offset 0x%X: %w",
					channel, command.Offset, err,
				)
			}
		}
	}
	return events, nil
}

func writeEventDigest(digest hash.Hash, event MusicEvent) {
	_, _ = fmt.Fprintf(
		digest, "%d|%d|%s|%02x|%02x|%02x|%d|%02x\n",
		event.Tick, event.Channel, event.Kind, event.Register, event.Value,
		event.Parameter, event.StreamOffset, event.Opcode,
	)
}

func auditTrackPlayback(
	driverData []byte,
	track TrackDescriptor,
	ticks int,
) (PlaybackAudit, error) {
	playback, initial, err := newTrackPlayback(driverData, track)
	if err != nil {
		return PlaybackAudit{}, err
	}
	digest := sha256.New()
	eventCount := 0
	parameterSet := make(map[int]bool)
	for _, event := range initial {
		writeEventDigest(digest, event)
		eventCount++
		if event.Kind == EventSetParameterBlock {
			parameterSet[int(event.Parameter)] = true
		}
	}
	for tick := 0; tick < ticks; tick++ {
		events, err := playback.Tick(4096)
		if err != nil {
			return PlaybackAudit{}, err
		}
		for _, event := range events {
			writeEventDigest(digest, event)
			eventCount++
			if event.Kind == EventSetParameterBlock {
				parameterSet[int(event.Parameter)] = true
			}
		}
	}
	parameterIndices := make([]int, 0, len(parameterSet))
	for index := range parameterSet {
		parameterIndices = append(parameterIndices, index)
	}
	sort.Ints(parameterIndices)
	return PlaybackAudit{
		Selector: track.Selector, Ticks: ticks, EventCount: eventCount,
		EventSHA256:                hex.EncodeToString(digest.Sum(nil)),
		ParameterIndices:           parameterIndices,
		EmbeddedParametersComplete: parametersWithinEmbeddedBank(parameterIndices),
	}, nil
}

func (playback *TrackPlayback) applyCommand(
	channel int,
	state *playbackChannel,
	command StreamCommand,
	events *[]MusicEvent,
) error {
	switch command.Name {
	case "note", "rest":
		state.duration = uint16(command.Operands[0]) - 1
		if channel < 3 {
			return playback.startFM(channel, command, events)
		}
		if channel < 6 {
			return playback.startPSG(channel, state, command, events)
		}
	case "parameter_85":
		if channel < 3 {
			*events = append(*events, MusicEvent{
				Tick: playback.tick, Channel: channel,
				Kind: EventSetParameterBlock, Parameter: command.Operands[0],
				StreamOffset: command.Offset, Opcode: command.Opcode,
			})
		} else if channel < 6 {
			state.envelopePointer = int(command.Operands[0])<<5 + 0x10
		}
	case "parameter_8a":
		if channel < 3 {
			*events = append(*events, MusicEvent{
				Tick: playback.tick, Channel: channel, Kind: EventSetVolume,
				Value: command.Operands[0], StreamOffset: command.Offset,
				Opcode: command.Opcode,
			})
		} else if channel < 6 {
			state.baseVolume = uint16(command.Operands[0])
			state.currentVolume = state.baseVolume
			playback.writeRegister(
				events, channel, byte(channel+5), command.Operands[0],
				command.Offset, command.Opcode,
			)
		}
	case "tempo_step":
		if playback.tempo < 0xC9 {
			playback.tempo += 4
			playback.writeRegister(
				events, channel, 0x26, byte(playback.tempo),
				command.Offset, command.Opcode,
			)
		}
	case "mode_91":
		state.mode = 1
	case "mode_92":
		state.mode = 2
	case "extension_b0":
		if channel >= 3 && channel < 6 {
			playback.writeRegister(
				events, channel, command.Operands[0], command.Operands[1],
				command.Offset, command.Opcode,
			)
		}
	}
	return nil
}

func (playback *TrackPlayback) startFM(
	channel int,
	command StreamCommand,
	events *[]MusicEvent,
) error {
	playback.writeRegister(
		events, channel, 0x28, byte(channel), command.Offset, command.Opcode,
	)
	if command.Name == "rest" {
		return nil
	}
	note := int(command.Opcode) - 3
	if note < 0 {
		return fmt.Errorf("FM note 0x%02X produces negative table index", command.Opcode)
	}
	octave, index := note/12, note%12
	fnumber, err := playback.word(0x210 + index*2)
	if err != nil {
		return err
	}
	playback.writeRegister(
		events, channel, byte(0xA4+channel),
		byte(octave<<3)|byte(fnumber>>8), command.Offset, command.Opcode,
	)
	playback.writeRegister(
		events, channel, byte(0xA0+channel),
		byte(fnumber), command.Offset, command.Opcode,
	)
	playback.writeRegister(
		events, channel, 0x28, byte(channel+0xF0), command.Offset, command.Opcode,
	)
	return nil
}

func (playback *TrackPlayback) startPSG(
	channel int,
	state *playbackChannel,
	command StreamCommand,
	events *[]MusicEvent,
) error {
	amplitudeRegister := byte(channel + 5)
	playback.writeRegister(
		events, channel, amplitudeRegister, 0, command.Offset, command.Opcode,
	)
	if command.Name == "rest" {
		return nil
	}
	note := int(command.Opcode)
	if note > 0x18 {
		note -= 0x18
	}
	index := note - 2
	if index < 0 || index >= 71 {
		return fmt.Errorf("PSG note 0x%02X maps outside 71-entry period table", command.Opcode)
	}
	period, err := playback.word(0x228 + index*2)
	if err != nil {
		return err
	}
	toneRegister := byte((channel - 3) * 2)
	playback.writeRegister(
		events, channel, toneRegister, byte(period), command.Offset, command.Opcode,
	)
	playback.writeRegister(
		events, channel, toneRegister+1, byte(period>>8),
		command.Offset, command.Opcode,
	)
	state.currentVolume = state.baseVolume
	value, err := playback.word(state.envelopePointer)
	if err != nil {
		return err
	}
	state.currentVolume += value
	playback.writeRegister(
		events, channel, amplitudeRegister, byte(state.currentVolume),
		command.Offset, command.Opcode,
	)
	return nil
}

func (playback *TrackPlayback) sustainPSG(
	channel int,
	state *playbackChannel,
	events *[]MusicEvent,
) error {
	toneRegister := byte((channel - 3) * 2)
	switch state.mode {
	case 1:
		value := playback.registers[toneRegister]
		if playback.counter&0x04 != 0 {
			value++
		} else {
			value--
		}
		playback.writeRegister(events, channel, toneRegister, value, 0, 0)
	case 2:
		if playback.counter&0x01 != 0 {
			value := playback.registers[toneRegister]
			if playback.counter&0x80 != 0 {
				value++
			} else {
				value--
			}
			playback.writeRegister(events, channel, toneRegister, value, 0, 0)
		}
	}

	state.envelopePointer += 2
	value, err := playback.word(state.envelopePointer)
	if err != nil {
		return err
	}
	if value == 0x80 {
		state.envelopePointer -= 2
		return nil
	}
	state.currentVolume += value
	if int16(state.currentVolume) > 0x0F {
		state.currentVolume = 0
	}
	playback.writeRegister(
		events, channel, byte(channel+5), byte(state.currentVolume), 0, 0,
	)
	return nil
}
