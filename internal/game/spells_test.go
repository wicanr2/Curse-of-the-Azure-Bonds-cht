package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestCampSpellLabelUsesGlobalSpellIDs(t *testing.T) {
	catalog := testCatalog()
	catalog.Strings["spell_cleric_7"] = "防護善良"
	catalog.Strings["spell_magic_user_7"] = "魔法飛彈"

	if got := campSpellLabel(catalog, party.ClassCleric, ProtectionFromGoodSpellID); got != "防護善良" {
		t.Fatalf("cleric global spell 0x%02X label=%q", ProtectionFromGoodSpellID, got)
	}
	if got := campSpellLabel(catalog, party.ClassMagicUser, MagicMissileSpellID); got != "魔法飛彈" {
		t.Fatalf("magic-user global spell 0x%02X label=%q", MagicMissileSpellID, got)
	}
	if got := campSpellLabel(catalog, party.ClassMagicUser, ProtectionFromGoodSpellID); !strings.Contains(got, "0x07") {
		t.Fatalf("magic-user must not reinterpret cleric ID 0x07: %q", got)
	}
}
