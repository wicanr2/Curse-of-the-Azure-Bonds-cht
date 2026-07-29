package pc98music

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/wicanr2/golden-box-remake-engine/audio/s98"
)

// S98ParameterStartup reports how one descriptor/first-stream parameter call
// materialized in the YM2203 trace before the channel's first audible note.
type S98ParameterStartup struct {
	Channel                   int      `json:"channel"`
	Phase                     string   `json:"phase"`
	ParameterIndex            int      `json:"parameter_index"`
	CompleteToneLoad          bool     `json:"complete_tone_load"`
	ToneLoadCount             int      `json:"tone_load_count"`
	SignatureSHA256           []string `json:"signature_sha256,omitempty"`
	MatchesEmbeddedBank       bool     `json:"matches_embedded_bank"`
	OverwrittenBeforeFirstKey bool     `json:"overwritten_before_first_key_on"`
	ActiveAtFirstKey          bool     `json:"active_at_first_key_on"`
}

// S98TrackAudit contains non-copyrighted metadata and hashes only. The source
// register trace remains in the user's local evidence workspace.
type S98TrackAudit struct {
	Selector         int                   `json:"selector"`
	S98SHA256        string                `json:"s98_sha256"`
	Bytes            int                   `json:"bytes"`
	TimerNumerator   uint32                `json:"timer_numerator"`
	TimerDenominator uint32                `json:"timer_denominator"`
	YM2203Clock      uint32                `json:"ym2203_clock"`
	DurationTicks    uint64                `json:"duration_ticks"`
	RegisterWrites   int                   `json:"register_writes"`
	ToneLoads        int                   `json:"tone_loads"`
	FirstKeyOns      int                   `json:"first_key_ons"`
	Startup          []S98ParameterStartup `json:"startup"`
}

// AuditS98Track cross-checks one local Hoot S98 v3 trace against the exact
// MSCDRV descriptor and stream parameter events for a public selector.
func AuditS98Track(driver, raw []byte, selector int) (S98TrackAudit, error) {
	if hash := fileSHA256(driver); hash != DriverSHA256 {
		return S98TrackAudit{}, fmt.Errorf("MSCDRV.EXE SHA-256 %s, want %s", hash, DriverSHA256)
	}
	file, err := s98.Parse(raw)
	if err != nil {
		return S98TrackAudit{}, err
	}
	if len(file.Devices) != 1 || file.Devices[0].Type != 2 {
		return S98TrackAudit{}, fmt.Errorf(
			"S98 declares %d devices; want one YM2203 type-2 device",
			len(file.Devices),
		)
	}
	loads, err := s98.YM2203ToneLoads(file.Events, 0)
	if err != nil {
		return S98TrackAudit{}, err
	}
	keyOns, err := s98.YM2203KeyOns(file.Events, 0)
	if err != nil {
		return S98TrackAudit{}, err
	}
	blocks, err := EmbeddedFMParameterBlocks(driver)
	if err != nil {
		return S98TrackAudit{}, err
	}
	known := make([][21]byte, len(blocks))
	for index := range blocks {
		known[index] = [21]byte(blocks[index].YM2203Signature())
	}
	expected, err := startupParameterEvents(driver, selector)
	if err != nil {
		return S98TrackAudit{}, err
	}
	trackStart, err := firstTrackToneTick(loads, expected, known)
	if err != nil {
		return S98TrackAudit{}, err
	}
	transitions := collapseToneLoads(loads, trackStart)
	firstKeys := firstKeyOnByChannel(keyOns, trackStart)

	report := S98TrackAudit{
		Selector: selector, S98SHA256: fileSHA256(raw), Bytes: len(raw),
		TimerNumerator: file.TimerNumerator, TimerDenominator: file.TimerDenominator,
		YM2203Clock: file.Devices[0].Clock, ToneLoads: len(loads),
	}
	for _, event := range file.Events {
		if event.Kind == s98.EventWrite {
			report.RegisterWrites++
		} else {
			report.DurationTicks += event.Wait
		}
	}
	for channel := 0; channel < 3; channel++ {
		observed := transitions[channel]
		if len(expected[channel]) != 2 {
			return S98TrackAudit{}, fmt.Errorf(
				"selector %d channel %d has %d startup parameters, want 2",
				selector, channel, len(expected[channel]),
			)
		}
		first, second := expected[channel][0], expected[channel][1]
		cursor := 0
		firstReport := S98ParameterStartup{
			Channel: channel, Phase: "descriptor", ParameterIndex: first,
		}
		if first < len(known) {
			if cursor >= len(observed) || observed[cursor].Signature != known[first] {
				return S98TrackAudit{}, fmt.Errorf(
					"selector %d channel %d descriptor parameter %d does not match S98: cursor=%d observed=%x expected=%x",
					selector, channel, first, cursor,
					observedSignature(observed, cursor), known[first],
				)
			}
			firstReport = parameterTrace(firstReport, observed[cursor:cursor+1], known, firstKeys[channel])
			cursor++
		} else {
			secondPosition := findToneSignature(observed, cursor, known[second])
			if secondPosition < 0 {
				return S98TrackAudit{}, fmt.Errorf(
					"selector %d channel %d cannot find stream parameter %d after descriptor parameter %d",
					selector, channel, second, first,
				)
			}
			firstReport = parameterTrace(
				firstReport, observed[cursor:secondPosition], known, firstKeys[channel],
			)
			cursor = secondPosition
		}
		_, hasFirstKey := firstKeys[channel]
		firstReport.OverwrittenBeforeFirstKey = hasFirstKey

		if second >= len(known) || cursor >= len(observed) ||
			observed[cursor].Signature != known[second] {
			return S98TrackAudit{}, fmt.Errorf(
				"selector %d channel %d first stream parameter %d does not match S98: cursor=%d observed=%x expected=%x",
				selector, channel, second, cursor,
				observedSignature(observed, cursor), known[second],
			)
		}
		secondReport := parameterTrace(S98ParameterStartup{
			Channel: channel, Phase: "first_stream", ParameterIndex: second,
		}, observed[cursor:cursor+1], known, firstKeys[channel])
		if hasFirstKey && firstReport.CompleteToneLoad && firstReport.ActiveAtFirstKey {
			return S98TrackAudit{}, fmt.Errorf(
				"selector %d channel %d descriptor parameter remained active at first key-on",
				selector, channel,
			)
		}
		if hasFirstKey && !secondReport.ActiveAtFirstKey {
			return S98TrackAudit{}, fmt.Errorf(
				"selector %d channel %d stream parameter %d was not active at first key-on",
				selector, channel, second,
			)
		}
		report.Startup = append(report.Startup, firstReport, secondReport)
		if hasFirstKey {
			report.FirstKeyOns++
		}
	}
	return report, nil
}

func observedSignature(loads []s98.YM2203ToneLoad, index int) [21]byte {
	if index < 0 || index >= len(loads) {
		return [21]byte{}
	}
	return loads[index].Signature
}

func startupParameterEvents(driver []byte, selector int) ([3][]int, error) {
	var result [3][]int
	playback, events, err := NewTrackPlayback(driver, selector)
	if err != nil {
		return result, err
	}
	record := func(events []MusicEvent) {
		for _, event := range events {
			if event.Kind == EventSetParameterBlock && event.Channel >= 0 && event.Channel < 3 &&
				len(result[event.Channel]) < 2 {
				result[event.Channel] = append(result[event.Channel], int(event.Parameter))
			}
		}
	}
	record(events)
	for tick := 0; tick < 4096; tick++ {
		complete := true
		for channel := 0; channel < 3; channel++ {
			complete = complete && len(result[channel]) >= 2
		}
		if complete {
			return result, nil
		}
		events, err = playback.Tick(4096)
		if err != nil {
			return result, err
		}
		record(events)
	}
	return result, fmt.Errorf("selector %d did not initialize all FM parameters", selector)
}

func collapseToneLoads(loads []s98.YM2203ToneLoad, minimumTick uint64) [3][]s98.YM2203ToneLoad {
	var result [3][]s98.YM2203ToneLoad
	var last [3][21]byte
	var have [3]bool
	for _, load := range loads {
		if load.Tick < minimumTick || load.Signature == [21]byte{} ||
			load.Channel < 0 || load.Channel >= 3 {
			continue
		}
		if have[load.Channel] && last[load.Channel] == load.Signature {
			continue
		}
		have[load.Channel] = true
		last[load.Channel] = load.Signature
		result[load.Channel] = append(result[load.Channel], load)
	}
	return result
}

func firstTrackToneTick(
	loads []s98.YM2203ToneLoad,
	expected [3][]int,
	known [][21]byte,
) (uint64, error) {
	var minimum uint64
	found := false
	for _, load := range loads {
		for channel := 0; channel < 3; channel++ {
			for _, index := range expected[channel] {
				if index < len(known) && load.Signature == known[index] &&
					(!found || load.Tick < minimum) {
					minimum = load.Tick
					found = true
				}
			}
		}
	}
	if !found {
		return 0, fmt.Errorf("S98 has no tone matching the selector startup parameters")
	}
	return minimum, nil
}

func firstKeyOnByChannel(keyOns []s98.YM2203KeyOn, minimumTick uint64) map[int]s98.YM2203KeyOn {
	result := make(map[int]s98.YM2203KeyOn, 3)
	for _, keyOn := range keyOns {
		if keyOn.Tick < minimumTick {
			continue
		}
		if _, exists := result[keyOn.Channel]; !exists {
			result[keyOn.Channel] = keyOn
		}
	}
	return result
}

func parameterTrace(
	result S98ParameterStartup,
	loads []s98.YM2203ToneLoad,
	known [][21]byte,
	firstKey s98.YM2203KeyOn,
) S98ParameterStartup {
	result.CompleteToneLoad = len(loads) != 0
	result.ToneLoadCount = len(loads)
	for _, load := range loads {
		sum := sha256.Sum256(load.Signature[:])
		result.SignatureSHA256 = append(result.SignatureSHA256, hex.EncodeToString(sum[:]))
		for _, signature := range known {
			if load.Signature == signature {
				result.MatchesEmbeddedBank = true
				break
			}
		}
		result.ActiveAtFirstKey = result.ActiveAtFirstKey || load.Signature == firstKey.Signature
	}
	return result
}

func findToneSignature(loads []s98.YM2203ToneLoad, start int, signature [21]byte) int {
	for index := start; index < len(loads); index++ {
		if loads[index].Signature == signature {
			return index
		}
	}
	return -1
}
