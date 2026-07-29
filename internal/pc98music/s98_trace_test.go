package pc98music

import (
	"testing"

	"github.com/wicanr2/golden-box-remake-engine/audio/s98"
)

func TestCollapseToneLoadsDropsResetAndConsecutiveRewrites(t *testing.T) {
	a := [21]byte{1}
	b := [21]byte{2}
	got := collapseToneLoads([]s98.YM2203ToneLoad{
		{Channel: 0},
		{Tick: 2, Channel: 0, Signature: a},
		{Tick: 2, Channel: 0, Signature: a},
		{Tick: 2, Channel: 1, Signature: b},
		{Tick: 3, Channel: 0, Signature: b},
	}, 2)
	if len(got[0]) != 2 || got[0][0].Signature != a || got[0][1].Signature != b ||
		len(got[1]) != 1 || got[1][0].Signature != b {
		t.Fatalf("collapsed loads = %+v", got)
	}
}

func TestParameterTraceHashesAndMatchesKnownTone(t *testing.T) {
	signature := [21]byte{1, 2, 3}
	key := s98.YM2203KeyOn{Channel: 0, Signature: signature}
	got := parameterTrace(
		S98ParameterStartup{ParameterIndex: 4},
		[]s98.YM2203ToneLoad{{Channel: 0, Signature: signature}},
		[][21]byte{signature},
		key,
	)
	if !got.CompleteToneLoad || !got.MatchesEmbeddedBank || !got.ActiveAtFirstKey ||
		got.ToneLoadCount != 1 || len(got.SignatureSHA256) != 1 ||
		len(got.SignatureSHA256[0]) != 64 {
		t.Fatalf("parameter trace = %+v", got)
	}
}

func TestAuditStartupOutputLevelsMatchesBaseAndCarrierRewrites(t *testing.T) {
	block := FMParameterBlock{
		FeedbackAlgorithm: 4,
		OutputLevel:       [4]byte{102, 107, 102, 107},
	}
	signature := [21]byte(block.YM2203Signature())
	sequence, err := block.YM2203LevelSequence(105)
	if err != nil {
		t.Fatal(err)
	}
	loads := make([]s98.YM2203ToneLoad, len(sequence))
	for index, levels := range sequence {
		loads[index] = s98.YM2203ToneLoad{
			Tick: 7, Channel: 0, Signature: signature, Levels: levels,
		}
	}
	controls := [3][]startupParameterVolume{
		{{parameter: 0, volume: 105, hasVolume: true}},
	}
	reports, err := auditStartupOutputLevels(loads, 7, []FMParameterBlock{block}, controls)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 || !reports[0].BaseLevelMatches ||
		!reports[0].CarrierOrderMatches || !reports[0].CompleteSequenceSeen ||
		reports[0].ExpectedToneLoads != 3 {
		t.Fatalf("output-level report=%+v", reports)
	}

	loads[2].Levels[2]++
	if _, err := auditStartupOutputLevels(loads, 7, []FMParameterBlock{block}, controls); err == nil {
		t.Fatal("bad second-carrier rewrite unexpectedly accepted")
	}
}
