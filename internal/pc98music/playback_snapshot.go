package pc98music

import "fmt"

const playbackSnapshotVersion = 1

type SequenceMachineSnapshot struct {
	Family    ChannelFamily  `json:"family"`
	PC        int            `json:"pc"`
	Start     int            `json:"start"`
	End       int            `json:"end"`
	CallStack []int          `json:"call_stack,omitempty"`
	LoopStack []LoopSnapshot `json:"loop_stack,omitempty"`
}

type LoopSnapshot struct {
	Target    int  `json:"target"`
	Remaining byte `json:"remaining"`
}

type PlaybackChannelSnapshot struct {
	Machine         SequenceMachineSnapshot `json:"machine"`
	Duration        uint16                  `json:"duration"`
	BaseVolume      uint16                  `json:"base_volume"`
	CurrentVolume   uint16                  `json:"current_volume"`
	EnvelopePointer int                     `json:"envelope_pointer"`
	Mode            byte                    `json:"mode"`
	OperatorMask    byte                    `json:"operator_mask"`
}

type TrackPlaybackSnapshot struct {
	Version   int                                          `json:"version"`
	Channels  [driverTrackChannels]PlaybackChannelSnapshot `json:"channels"`
	Registers [256]byte                                    `json:"registers"`
	Tick      uint64                                       `json:"tick"`
	Counter   uint16                                       `json:"counter"`
	Tempo     uint16                                       `json:"tempo"`
}

func (playback *TrackPlayback) Snapshot() (TrackPlaybackSnapshot, error) {
	if playback == nil || len(playback.driverData) == 0 {
		return TrackPlaybackSnapshot{}, fmt.Errorf("track playback is not initialized")
	}
	snapshot := TrackPlaybackSnapshot{
		Version: playbackSnapshotVersion, Registers: playback.registers,
		Tick: playback.tick, Counter: playback.counter, Tempo: playback.tempo,
	}
	for index := range playback.channels {
		channel := &playback.channels[index]
		if channel.machine == nil {
			return TrackPlaybackSnapshot{}, fmt.Errorf("channel %d sequence machine is nil", index)
		}
		machine := channel.machine
		machineSnapshot := SequenceMachineSnapshot{
			Family: machine.Family, PC: machine.PC, Start: machine.Start, End: machine.End,
			CallStack: append([]int(nil), machine.callStack...),
			LoopStack: make([]LoopSnapshot, len(machine.loopStack)),
		}
		for loopIndex, loop := range machine.loopStack {
			machineSnapshot.LoopStack[loopIndex] = LoopSnapshot{Target: loop.Target, Remaining: loop.Remaining}
		}
		snapshot.Channels[index] = PlaybackChannelSnapshot{
			Machine: machineSnapshot, Duration: channel.duration,
			BaseVolume: channel.baseVolume, CurrentVolume: channel.currentVolume,
			EnvelopePointer: channel.envelopePointer, Mode: channel.mode,
			OperatorMask: channel.operatorMask,
		}
	}
	return snapshot, nil
}

func (playback *TrackPlayback) Restore(snapshot TrackPlaybackSnapshot) error {
	if playback == nil || len(playback.driverData) == 0 {
		return fmt.Errorf("track playback is not initialized")
	}
	if snapshot.Version != playbackSnapshotVersion {
		return fmt.Errorf("unsupported track playback snapshot version %d", snapshot.Version)
	}
	for index := range playback.channels {
		current := playback.channels[index].machine
		machine := snapshot.Channels[index].Machine
		if current == nil || machine.Family != current.Family || machine.Start != current.Start || machine.End != current.End {
			return fmt.Errorf("channel %d sequence identity does not match track", index)
		}
		if machine.PC < machine.Start || machine.PC > machine.End {
			return fmt.Errorf("channel %d PC 0x%X outside 0x%X..0x%X", index, machine.PC, machine.Start, machine.End)
		}
		if len(machine.CallStack) > 16 || len(machine.LoopStack) > 16 {
			return fmt.Errorf("channel %d sequence stack exceeds 16 entries", index)
		}
		for _, address := range machine.CallStack {
			if address < machine.Start || address > machine.End {
				return fmt.Errorf("channel %d call return 0x%X outside sequence", index, address)
			}
		}
		for _, loop := range machine.LoopStack {
			if loop.Target < machine.Start || loop.Target >= machine.End || loop.Remaining == 0 {
				return fmt.Errorf("channel %d invalid loop target/count 0x%X/%d", index, loop.Target, loop.Remaining)
			}
		}
	}
	playback.registers = snapshot.Registers
	playback.tick, playback.counter, playback.tempo = snapshot.Tick, snapshot.Counter, snapshot.Tempo
	for index := range playback.channels {
		target := &playback.channels[index]
		source := snapshot.Channels[index]
		target.duration, target.baseVolume, target.currentVolume = source.Duration, source.BaseVolume, source.CurrentVolume
		target.envelopePointer, target.mode, target.operatorMask = source.EnvelopePointer, source.Mode, source.OperatorMask
		target.machine.PC = source.Machine.PC
		target.machine.callStack = append(target.machine.callStack[:0], source.Machine.CallStack...)
		target.machine.loopStack = target.machine.loopStack[:0]
		for _, loop := range source.Machine.LoopStack {
			target.machine.loopStack = append(target.machine.loopStack, sequenceLoop{Target: loop.Target, Remaining: loop.Remaining})
		}
	}
	return nil
}

type YM2203RendererSnapshot struct {
	Parameters [3]int  `json:"parameters"`
	Have       [3]bool `json:"have"`
}

func (renderer *YM2203EventRenderer) Snapshot() (YM2203RendererSnapshot, error) {
	if renderer == nil {
		return YM2203RendererSnapshot{}, fmt.Errorf("YM2203 event renderer is nil")
	}
	return YM2203RendererSnapshot{Parameters: renderer.parameters, Have: renderer.have}, nil
}

func (renderer *YM2203EventRenderer) Restore(snapshot YM2203RendererSnapshot) error {
	if renderer == nil {
		return fmt.Errorf("YM2203 event renderer is nil")
	}
	for channel, parameter := range snapshot.Parameters {
		have := parameter >= 0 && parameter < len(renderer.blocks)
		if snapshot.Have[channel] != have {
			return fmt.Errorf("renderer channel %d parameter %d/have %v is inconsistent", channel, parameter, snapshot.Have[channel])
		}
	}
	renderer.parameters, renderer.have = snapshot.Parameters, snapshot.Have
	return nil
}
