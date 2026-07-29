// Package pc98music verifies title-specific PC-98 music bridge evidence.
package pc98music

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	GameSHA256   = "8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0"
	DriverSHA256 = "bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5"

	DriverMissingStart = 0x4000
	DriverMissingEnd   = 0x4400

	driverDataFileBase   = 0x14D0
	driverTrackTableFile = 0x1800
	driverPublicTracks   = 12
	driverTrackChannels  = 7
)

// Anchor is a short instruction sequence independently located by IDA Pro and
// then checked against the original executable bytes.
type Anchor struct {
	Binary     string `json:"binary"`
	Label      string `json:"label"`
	FileOffset int    `json:"file_offset"`
	Bytes      []byte `json:"-"`
}

type AnchorResult struct {
	Binary     string `json:"binary"`
	Label      string `json:"label"`
	FileOffset int    `json:"file_offset"`
	Size       int    `json:"size"`
}

// SoundBIOSService describes one INT D2h client that is present in this
// MSCDRV.EXE. Names follow NEC's official PC-9800 Sound BIOS command table.
type SoundBIOSService struct {
	Command    byte   `json:"command"`
	Name       string `json:"name"`
	FileOffset int    `json:"file_offset"`
	Arguments  string `json:"arguments"`
}

// TrackChannel reports one raw eight-byte channel record. SequenceOffset is
// relative to the driver's data segment; the remaining two words stay raw
// until both producer and playback consumers are fully traced.
type TrackChannel struct {
	Channel              int            `json:"channel"`
	SequenceOffset       int            `json:"sequence_offset"`
	SequenceLength       int            `json:"sequence_length"`
	RawParameter1        int            `json:"raw_parameter_1"`
	RawParameter2        int            `json:"raw_parameter_2"`
	FileStart            int            `json:"file_start"`
	FileEnd              int            `json:"file_end"`
	Complete             bool           `json:"complete"`
	SHA256               string         `json:"sha256"`
	CommandCount         int            `json:"command_count"`
	OpcodeCounts         map[string]int `json:"opcode_counts"`
	ValidatedTimedEvents int            `json:"validated_timed_events"`
	ValidationMode       string         `json:"validation_mode"`
}

type TrackDescriptor struct {
	Selector         int            `json:"selector"`
	DriverIndex      int            `json:"driver_index"`
	DescriptorOffset int            `json:"descriptor_offset"`
	DescriptorFile   int            `json:"descriptor_file"`
	HeaderWords      [4]int         `json:"header_words"`
	Channels         []TrackChannel `json:"channels"`
	Complete         bool           `json:"complete"`
}

type BridgeReport struct {
	GameSHA256        string             `json:"game_sha256"`
	DriverSHA256      string             `json:"driver_sha256"`
	DriverMissingFrom int                `json:"driver_missing_from"`
	DriverMissingTo   int                `json:"driver_missing_to"`
	Anchors           []AnchorResult     `json:"anchors"`
	SoundBIOSServices []SoundBIOSService `json:"sound_bios_services"`
	Tracks            []TrackDescriptor  `json:"tracks"`
	PlaybackAudits    []PlaybackAudit    `json:"playback_audits"`
}

var gameAnchors = []Anchor{
	{
		Binary:     "GAME.EXE",
		Label:      "MSCPLAY_MSCSTOP_VECTOR_7E_CALLS",
		FileOffset: 0x9410,
		Bytes: mustHex(
			"c606397f008a4606a2387fb07e50bf387f1e579a0b00bd08" +
				"89ec5dca02005589e5c606397f01b07e50bf387f1e579a0b00bd08",
		),
	},
	{
		Binary:     "GAME.EXE",
		Label:      "REGISTER_IMAGE_IVT_TRAMPOLINE",
		FileOffset: 0x957B,
		Bytes: mustHex(
			"551e8bec9cbb45000e5333db8edb8a5e0cd1e3d1e3c51f1e53" +
				"c57608fcad50ad8bd8ad8bc8ad8bd0ad8be8ad50ad8bf8ad50ad" +
				"8ec01f5e58facb9c0657558becc47e10fcab8bc3ab8bc1ab8bc2" +
				"ab58ab8bc6ab58ab8cd8ab58ab58ab1f5dca",
		),
	},
}

var driverAnchors = []Anchor{
	{
		Binary:     "MSCDRV.EXE",
		Label:      "VECTOR_7E_HANDLER",
		FileOffset: 0x0280,
		Bytes: mustHex(
			"505351525657551e06bb2d018edb833e060000752d84e4752650" +
				"e82d10e84a10e8300ce8910ee8370e58bbff003c0e7503bb0100" +
				"5350e8650183c404eb0490e8bc02071f5d5f5e5a595b58cf",
		),
	},
	{
		Binary:     "MSCDRV.EXE",
		Label:      "INSTALL_VECTOR_7E",
		FileOffset: 0x02E3,
		Bytes:      mustHex("2ec41ef20026c7078000268c4f02c3f8010000"),
	},
	{
		Binary:     "MSCDRV.EXE",
		Label:      "INSTALL_INT_D2",
		FileOffset: 0x12CA,
		Bytes: mustHex(
			"b8d235cd21891e21038cc0a323031eb8e0ce8ed8bb06008b17" +
				"b8d225cd211fc3",
		),
	},
	{
		Binary:     "MSCDRV.EXE",
		Label:      "INT_D2_INITIALIZE_AND_SHUTDOWN_CLIENTS",
		FileOffset: 0x135D,
		Bytes:      mustHex("b400cdd2b82d018ed8c3558bec8a4604e802005dc3b402cdd2c3"),
	},
	{
		Binary:     "MSCDRV.EXE",
		Label:      "DIRECT_YM2203_WRITE_PORTS_188_18A",
		FileOffset: 0x1275,
		Bytes: mustHex(
			"515250ba8801eca88075f85850ee8ac34242b92000e2feee585a59c3",
		),
	},
	{
		Binary:     "MSCDRV.EXE",
		Label:      "DIRECT_YM2203_READ_PORTS_188_18A",
		FileOffset: 0x12AB,
		Bytes: mustHex(
			"515250ba8801eca88075f858eeb90400e2fe504242ba8a01ec8ad8585a59c3",
		),
	},
}

var soundBIOSServices = []SoundBIOSService{
	{0x00, "INITIALIZE", 0x135D, "ES=控制／工作區段"},
	{0x02, "CLEAR", 0x1372, "無"},
	{0x10, "READREG", 0x1384, "AL=OPN 暫存器；回傳 BX"},
	{0x11, "WRITEREG", 0x13A1, "AL=OPN 暫存器，BL=資料"},
	{0x12, "SETTOUCH", 0x13B4, "AL=聲道，BL=觸鍵／閘門比"},
	{0x13, "NOTE", 0x13CA, "AL=聲道，BH=音高，BL=音長"},
	{0x14, "SETLENGTH", 0x13DD, "AL=聲道，BL=預設音長"},
	{0x16, "SETPARABLOCK", 0x140D, "AL=聲道，ES:BX=參數塊，DL=種類"},
	{0x17, "READPARA", 0x1420, "AL=聲道，BL=參數編號；回傳 BX"},
	{0x18, "WRITEPARA", 0x1438, "AL=聲道，BL=參數編號，DX=值"},
	{0x19, "ALLSTOP", 0x143D, "無"},
	{0x1A, "CONTPLAY", 0x1442, "無"},
	{0x1B, "MODUON", 0x1462, "AL=聲道"},
	{0x1C, "MODUOFF", 0x1472, "AL=聲道"},
	{0x1D, "SETINTCOND", 0x148A, "AL=聲道，ES:BX=回呼，CX=條件"},
	{0x1E, "HOLDSTATE", 0x1452, "AL=聲道，BL=維持音長"},
	{0x1F, "SETVOLUME", 0x149D, "AL=聲道，BL=音量"},
}

func mustHex(value string) []byte {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		panic(err)
	}
	return decoded
}

func fileSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func verifyAnchors(data []byte, anchors []Anchor) ([]AnchorResult, error) {
	results := make([]AnchorResult, 0, len(anchors))
	for _, anchor := range anchors {
		end := anchor.FileOffset + len(anchor.Bytes)
		if anchor.FileOffset < 0 || end > len(data) {
			return nil, fmt.Errorf(
				"%s %s range 0x%X..0x%X is outside %d-byte input",
				anchor.Binary, anchor.Label, anchor.FileOffset, end, len(data),
			)
		}
		if !bytes.Equal(data[anchor.FileOffset:end], anchor.Bytes) {
			return nil, fmt.Errorf(
				"%s %s byte mismatch at file offset 0x%X",
				anchor.Binary, anchor.Label, anchor.FileOffset,
			)
		}
		results = append(results, AnchorResult{
			Binary:     anchor.Binary,
			Label:      anchor.Label,
			FileOffset: anchor.FileOffset,
			Size:       len(anchor.Bytes),
		})
	}
	return results, nil
}

func verifySoundBIOSServices(data []byte) error {
	for _, service := range soundBIOSServices {
		want := []byte{0xB4, service.Command, 0xCD, 0xD2}
		end := service.FileOffset + len(want)
		if service.FileOffset < 0 || end > len(data) {
			return fmt.Errorf(
				"MSCDRV.EXE Sound BIOS %s range 0x%X..0x%X is outside input",
				service.Name, service.FileOffset, end,
			)
		}
		if !bytes.Equal(data[service.FileOffset:end], want) {
			return fmt.Errorf(
				"MSCDRV.EXE Sound BIOS %s byte mismatch at file offset 0x%X",
				service.Name, service.FileOffset,
			)
		}
	}
	return nil
}

func wordAt(data []byte, offset int) (int, error) {
	if offset < 0 || offset+2 > len(data) {
		return 0, fmt.Errorf("word at file offset 0x%X is outside %d-byte input", offset, len(data))
	}
	return int(data[offset]) | int(data[offset+1])<<8, nil
}

func rangeComplete(start, end int) bool {
	return !(start < DriverMissingEnd && end > DriverMissingStart)
}

func auditTrackDescriptors(data []byte) ([]TrackDescriptor, error) {
	tracks := make([]TrackDescriptor, 0, driverPublicTracks)
	for index := 0; index < driverPublicTracks; index++ {
		descriptorOffset, err := wordAt(data, driverTrackTableFile+index*2)
		if err != nil {
			return nil, fmt.Errorf("track index %d pointer: %w", index, err)
		}
		descriptorFile := driverDataFileBase + descriptorOffset
		if descriptorFile < 0 || descriptorFile+64 > len(data) {
			return nil, fmt.Errorf(
				"track index %d descriptor 0x%X maps outside input at 0x%X",
				index, descriptorOffset, descriptorFile,
			)
		}
		track := TrackDescriptor{
			Selector:         index + 1,
			DriverIndex:      index,
			DescriptorOffset: descriptorOffset,
			DescriptorFile:   descriptorFile,
			Complete:         rangeComplete(descriptorFile, descriptorFile+64),
		}
		for headerIndex := range track.HeaderWords {
			value, err := wordAt(data, descriptorFile+headerIndex*2)
			if err != nil {
				return nil, err
			}
			track.HeaderWords[headerIndex] = value
		}
		track.Channels = make([]TrackChannel, 0, driverTrackChannels)
		for channel := 0; channel < driverTrackChannels; channel++ {
			record := descriptorFile + 8 + channel*8
			sequenceOffset, err := wordAt(data, record)
			if err != nil {
				return nil, err
			}
			sequenceLength, err := wordAt(data, record+2)
			if err != nil {
				return nil, err
			}
			raw1, err := wordAt(data, record+4)
			if err != nil {
				return nil, err
			}
			raw2, err := wordAt(data, record+6)
			if err != nil {
				return nil, err
			}
			fileStart := driverDataFileBase + sequenceOffset
			fileEnd := fileStart + sequenceLength
			if fileStart < 0 || fileEnd < fileStart || fileEnd > len(data) {
				return nil, fmt.Errorf(
					"track %d channel %d sequence 0x%X+0x%X maps outside input",
					index+1, channel, sequenceOffset, sequenceLength,
				)
			}
			complete := rangeComplete(fileStart, fileEnd)
			track.Complete = track.Complete && complete
			stream := data[fileStart:fileEnd]
			opcodeCounts := make(map[string]int)
			commandCount := 0
			validationMode := "declared-range-static-and-control"
			machineEnd := sequenceOffset + sequenceLength
			if channel == 6 {
				// sub_10410's timing branch does not honor the descriptor end
				// or execute A0-A4. It reads onward until a timed byte occurs.
				validationMode = "bounded-runtime-read-through"
				machineEnd = len(data) - driverDataFileBase
			} else {
				commands, err := DecodeStreamStructure(channel, stream)
				if err != nil {
					return nil, fmt.Errorf(
						"track %d channel %d decode structure: %w",
						index+1, channel, err,
					)
				}
				commandCount = len(commands)
				for _, command := range commands {
					opcodeCounts[command.Name]++
				}
			}
			machine, err := NewSequenceMachine(
				channel, sequenceOffset, machineEnd,
			)
			if err != nil {
				return nil, err
			}
			validatedTimedEvents := 0
			driverData := data[driverDataFileBase:]
			for validatedTimedEvents < 256 &&
				(channel == 6 || machine.PC != machine.End) {
				commands, err := machine.NextTimed(driverData, 4096)
				if IsSequenceEnd(err) {
					break
				} else if err != nil {
					return nil, fmt.Errorf(
						"track %d channel %d control flow after %d timed events: %w",
						index+1, channel, validatedTimedEvents, err,
					)
				}
				if channel == 6 {
					commandCount += len(commands)
					for _, command := range commands {
						opcodeCounts[command.Name]++
					}
				}
				validatedTimedEvents++
			}
			track.Channels = append(track.Channels, TrackChannel{
				Channel:              channel,
				SequenceOffset:       sequenceOffset,
				SequenceLength:       sequenceLength,
				RawParameter1:        raw1,
				RawParameter2:        raw2,
				FileStart:            fileStart,
				FileEnd:              fileEnd,
				Complete:             complete,
				SHA256:               fileSHA256(stream),
				CommandCount:         commandCount,
				OpcodeCounts:         opcodeCounts,
				ValidatedTimedEvents: validatedTimedEvents,
				ValidationMode:       validationMode,
			})
		}
		tracks = append(tracks, track)
	}
	return tracks, nil
}

// ExtractTrackSequences returns copies of the seven channel byte streams for
// one 1-based public selector. It accepts only the identified user-media
// driver and refuses any stream touching the absent-sector range.
func ExtractTrackSequences(driver []byte, selector int) ([][]byte, error) {
	if selector < 1 || selector > driverPublicTracks {
		return nil, fmt.Errorf("track selector %d is outside 1..%d", selector, driverPublicTracks)
	}
	if hash := fileSHA256(driver); hash != DriverSHA256 {
		return nil, fmt.Errorf("MSCDRV.EXE SHA-256 %s, want %s", hash, DriverSHA256)
	}
	tracks, err := auditTrackDescriptors(driver)
	if err != nil {
		return nil, err
	}
	track := tracks[selector-1]
	if !track.Complete {
		return nil, fmt.Errorf("track selector %d overlaps absent-sector range", selector)
	}
	sequences := make([][]byte, 0, len(track.Channels))
	for _, channel := range track.Channels {
		sequences = append(
			sequences,
			append([]byte(nil), driver[channel.FileStart:channel.FileEnd]...),
		)
	}
	return sequences, nil
}

// AuditBridge verifies the exact supplied GAME.EXE and incomplete MSCDRV.EXE.
// It deliberately rejects other dumps so evidence from the known missing
// sector cannot be silently generalized.
func AuditBridge(game, driver []byte) (BridgeReport, error) {
	gameHash := fileSHA256(game)
	if gameHash != GameSHA256 {
		return BridgeReport{}, fmt.Errorf("GAME.EXE SHA-256 %s, want %s", gameHash, GameSHA256)
	}
	driverHash := fileSHA256(driver)
	if driverHash != DriverSHA256 {
		return BridgeReport{}, fmt.Errorf("MSCDRV.EXE SHA-256 %s, want %s", driverHash, DriverSHA256)
	}
	gameResults, err := verifyAnchors(game, gameAnchors)
	if err != nil {
		return BridgeReport{}, err
	}
	driverResults, err := verifyAnchors(driver, driverAnchors)
	if err != nil {
		return BridgeReport{}, err
	}
	if err := verifySoundBIOSServices(driver); err != nil {
		return BridgeReport{}, err
	}
	tracks, err := auditTrackDescriptors(driver)
	if err != nil {
		return BridgeReport{}, err
	}
	playbackAudits := make([]PlaybackAudit, 0, len(tracks))
	for _, track := range tracks {
		audit, err := auditTrackPlayback(driver[driverDataFileBase:], track, 4096)
		if err != nil {
			return BridgeReport{}, fmt.Errorf(
				"track %d playback audit: %w", track.Selector, err,
			)
		}
		playbackAudits = append(playbackAudits, audit)
	}
	return BridgeReport{
		GameSHA256:        gameHash,
		DriverSHA256:      driverHash,
		DriverMissingFrom: DriverMissingStart,
		DriverMissingTo:   DriverMissingEnd,
		Anchors:           append(gameResults, driverResults...),
		SoundBIOSServices: append([]SoundBIOSService(nil), soundBIOSServices...),
		Tracks:            tracks,
		PlaybackAudits:    playbackAudits,
	}, nil
}
