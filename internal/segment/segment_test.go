package segment

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// 註冊表的 id 必須與 `docs/plan/segment-labels.json` 的鍵完全一致：
// 標籤那一份是 SEG-03 逐條找原作敘述定下來的，兩邊分岔就代表有一段沒人認領。
func TestRegistryMatchesSegmentLabels(t *testing.T) {
	payload, err := os.ReadFile("../../docs/plan/segment-labels.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Labels map[string]string `json:"labels"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.Labels) != len(registry) {
		t.Fatalf("標籤 %d 段、註冊表 %d 段", len(document.Labels), len(registry))
	}
	for _, candidate := range registry {
		if _, ok := document.Labels[candidate.ID]; !ok {
			t.Errorf("註冊表有 %s，標籤沒有", candidate.ID)
		}
		if want := FormatID(candidate.Member, candidate.Block); want != candidate.ID {
			t.Errorf("%s 的成員／block 對不上 id，應為 %s", candidate.ID, want)
		}
	}
	for id := range document.Labels {
		if _, ok := Lookup(id); !ok {
			t.Errorf("標籤有 %s，註冊表沒有", id)
		}
	}
}

var graphRowPattern = regexp.MustCompile(
	"^\\| `(ECL[1-6]/0x[0-9A-F]{2})` \\| [^|]* \\| ([^|]*) \\| ([^|]*) \\|")

// EnterFrom 不能憑印象填：它必須是轉移圖「進入自」欄位裡真的存在的一條邊。
// 唯二的例外是開場與提爾佛頓第一段——引擎在 LastECL ＝ 0 時才走到它們
// （spec 1141），所以那兩段的 EnterFrom 是 0x00 而不是某個 block。
func TestEnterFromIsARealIncomingEdge(t *testing.T) {
	payload, err := os.ReadFile("../../docs/audit/ecl-block-graph.md")
	if err != nil {
		t.Fatal(err)
	}
	incoming := map[string]map[uint8]bool{}
	for _, line := range strings.Split(string(payload), "\n") {
		match := graphRowPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		blocks := map[uint8]bool{}
		for _, token := range strings.Split(match[2], "、") {
			token = strings.Trim(strings.TrimSpace(token), "`")
			if !strings.HasPrefix(token, "0x") {
				continue
			}
			value, parseErr := strconv.ParseUint(token[2:], 16, 8)
			if parseErr != nil {
				t.Fatalf("轉移圖的 %q 不是 block 編號", token)
			}
			blocks[uint8(value)] = true
		}
		incoming[match[1]] = blocks
	}
	if len(incoming) != len(registry) {
		t.Fatalf("轉移圖段落清單 %d 段、註冊表 %d 段", len(incoming), len(registry))
	}
	for _, candidate := range registry {
		edges, ok := incoming[candidate.ID]
		if !ok {
			t.Errorf("轉移圖沒有 %s 這一段", candidate.ID)
			continue
		}
		if candidate.EnterFrom == 0x00 {
			if len(edges) != 0 {
				t.Errorf("%s 宣告成全新開局，但轉移圖有 %d 條進入邊", candidate.ID, len(edges))
			}
			continue
		}
		if !edges[candidate.EnterFrom] {
			t.Errorf("%s 的 EnterFrom 0x%02X 不在轉移圖的進入自清單裡", candidate.ID, candidate.EnterFrom)
		}
	}
}

// GameArea 一律等於 ECL 成員編號。世界地圖上的段也一樣：它不載 GEO 幾何，
// 但圖片素材仍然按章節分檔，落成別的數字就會去對不存在的檔案。
func TestGameAreaFollowsTheECLMember(t *testing.T) {
	for _, candidate := range registry {
		if candidate.GameArea != candidate.Member {
			t.Errorf("%s 的章節 %d 不等於成員編號 %d",
				candidate.ID, candidate.GameArea, candidate.Member)
		}
	}
}

func TestLookupAcceptsIDBlockAndLegacyFlag(t *testing.T) {
	for _, name := range []string{"ECL5/0x32", "ecl5/0x32", "0x32", "32", "lava-tube"} {
		found, ok := Lookup(name)
		if !ok || found.ID != "ECL5/0x32" {
			t.Errorf("Lookup(%q) = %v, %v", name, found.ID, ok)
		}
	}
	if _, ok := Lookup("no-such-segment"); ok {
		t.Error("Lookup 收下了不存在的段名")
	}
	if _, ok := Lookup(""); ok {
		t.Error("Lookup 收下了空字串")
	}
	seen := map[string]bool{}
	for _, name := range Names() {
		if seen[name] {
			t.Errorf("%q 在別名清單裡重複", name)
		}
		seen[name] = true
		if _, ok := Lookup(name); !ok {
			t.Errorf("Names 列了 %q，但 Lookup 查不到", name)
		}
	}
}
