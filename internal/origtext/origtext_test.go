package origtext

import (
	"testing"

	"golang.org/x/text/encoding/traditionalchinese"
)

// 英文原版的名字必須逐位元組不變，否則這個改動會動到既有的英文資料。
func TestASCIIIsUnchanged(t *testing.T) {
	for _, name := range []string{"DRAGONBAIT", "Long Sword +1", "", "a b  c", "~COMBAT"} {
		if got := Decode([]byte(name)); got != name {
			t.Fatalf("Decode(%q) = %q, want unchanged", name, got)
		}
	}
}

// 中文版的名字是 Big5：兩個位元組要組成一個字，不是兩個壞位元組。
func TestBig5NameBecomesOneRunePerCharacter(t *testing.T) {
	const want = "青色枷"
	encoded, err := traditionalchinese.Big5.NewEncoder().Bytes([]byte(want))
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) != 6 {
		t.Fatalf("Big5 編碼長度 %d，預期 6 個位元組", len(encoded))
	}
	got := Decode(encoded)
	if got != want {
		t.Fatalf("Decode(Big5(%q)) = %q", want, got)
	}
	if len([]rune(got)) != 3 {
		t.Fatalf("解出 %d 個字元，預期 3 個", len([]rune(got)))
	}
	// 舊行為的對照：直接當 UTF-8 會得到一串無效位元組。
	if string(encoded) == want {
		t.Fatal("測試前提不成立：Big5 位元組不該直接等於 UTF-8 字串")
	}
}

// 定寬欄位的補位位元組不能被當成字的一部分。
func TestDecodeFieldTrimsPadding(t *testing.T) {
	encoded, err := traditionalchinese.Big5.NewEncoder().Bytes([]byte("長劍"))
	if err != nil {
		t.Fatal(err)
	}
	field := make([]byte, 0, 0x2A)
	field = append(field, encoded...)
	for len(field) < 0x2A {
		field = append(field, 0x00)
	}
	if got := DecodeField(field); got != "長劍" {
		t.Fatalf("DecodeField = %q, want 長劍", got)
	}
	if got := DecodeField([]byte("Dagger    ")); got != "Dagger" {
		t.Fatalf("DecodeField = %q, want Dagger", got)
	}
}

// 解不開的位元組要退回舊行為，不能讓整筆記錄讀取失敗。
func TestUndecodableBytesFallBackInsteadOfFailing(t *testing.T) {
	raw := []byte{0xFF, 0xFF}
	if got := Decode(raw); got == "" {
		t.Fatal("無法解碼時不應回空字串")
	}
}
