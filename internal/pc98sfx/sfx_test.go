package pc98sfx

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestImportRejectsUnknownExecutable(t *testing.T) {
	t.Parallel()

	_, err := Import([]byte("not GAME.EXE"))
	if err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("Import error=%v", err)
	}
}

func TestDecodeImmediateFormulaAndTablePaths(t *testing.T) {
	t.Parallel()

	game := make([]byte, tableOffset+selectorCount*wordsPerEffect*2+2)
	put := func(selector, index int, value uint16) {
		offset := tableOffset + (selector*wordsPerEffect+index)*2
		game[offset] = byte(value)
		game[offset+1] = byte(value >> 8)
	}
	put(3, 1, 1000)
	put(3, 2, 3000)
	put(3, 3, 500)
	put(3, 4, 0)

	noOp := decodeEffect(game, 13)
	if !noOp.NoOp || len(noOp.Steps) != 0 {
		t.Fatalf("selector 13=%+v", noOp)
	}
	formula := decodeEffect(game, 2)
	if formula.Source != "formula" || len(formula.Steps) != 1 ||
		formula.Steps[0].FrequencyOrPeriod != 20 ||
		formula.Steps[0].PulseCount != 125 {
		t.Fatalf("selector 2=%+v", formula)
	}
	table := decodeEffect(game, 3)
	if table.Source != "table" || len(table.Steps) != 3 ||
		table.Steps[0].PulseCount != 2 ||
		table.Steps[1].Kind != StepDelay ||
		table.Steps[2].PulseCount != 4 {
		t.Fatalf("selector 3=%+v", table)
	}
}

func TestKnownDigestConstantIsCanonicalHex(t *testing.T) {
	t.Parallel()

	decoded, err := hex.DecodeString(GameSHA256)
	if err != nil || len(decoded) != sha256.Size {
		t.Fatalf("GameSHA256=%q err=%v", GameSHA256, err)
	}
}
