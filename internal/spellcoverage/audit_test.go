package spellcoverage

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
)

func TestFormalPlayerSpellCoverageReportsEveryPackEntry(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	report, err := Build(pack, filepath.Join("..", "game", "combat_state.go"))
	if err != nil {
		t.Fatal(err)
	}
	if report.SpellCount != len(pack.CombatPlayerSpells) || report.SpellCount == 0 {
		t.Fatalf("report spell count=%d pack=%d", report.SpellCount, len(pack.CombatPlayerSpells))
	}
	for _, spell := range report.Spells {
		if spell.ID == "" || spell.Behavior == "" || spell.TargetMode == "" {
			t.Fatalf("incomplete report row=%+v", spell)
		}
		if spell.Handler.Status == "" || spell.Visual.Status == "" || spell.Sound.Status == "" {
			t.Fatalf("missing machine status for %s: %+v", spell.ID, spell)
		}
	}
	if report.MissingCount == 0 {
		t.Fatal("audit unexpectedly claims every spell has complete runtime, visual, and sound evidence")
	}
}

func TestRuntimeScannerDoesNotTreatPackDeclarationAsCoverage(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	report, err := Build(pack, filepath.Join("..", "game", "combat_state.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, spell := range report.Spells {
		if spell.Visual.Status == "missing" || spell.Sound.Status == "missing" {
			if len(spell.Limitations) == 0 {
				t.Fatalf("missing evidence has no limitation for %s", spell.ID)
			}
		}
	}
}
