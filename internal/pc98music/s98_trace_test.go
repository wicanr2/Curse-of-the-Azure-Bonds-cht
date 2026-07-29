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
