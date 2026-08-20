package eclcells

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

// 分派器的偵測是逐格盤點的地圖：少偵測到一個，那個 block 的內容就整批看起來
// 「沒有每格事件」——那是假零。這一條把形狀釘住。
func TestAnalyzeFindsBothMaskShapes(t *testing.T) {
	const imagePath = "../../curseoftheazurebonds.zip"
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	found, masks, tableForm := 0, map[int]int{}, 0
	blocks := 0
	for member := 1; member <= 6; member++ {
		payload := memberPayload(t, archive, fmt.Sprintf("ECL%d.DAX", member))
		parsed, err := dax.Parse(payload)
		if err != nil {
			t.Fatal(err)
		}
		for _, block := range parsed {
			blocks++
			dispatch := Analyze(block.Data)
			switch {
			case dispatch.Found:
				found++
				masks[dispatch.Mask]++
				assertAscending(t, member, block.Entry.ID, dispatch.Indexes)
			case dispatch.TableForm:
				tableForm++
			}
		}
	}
	if blocks != 25 {
		t.Fatalf("掃到 %d 個 block，應該是 25 個", blocks)
	}
	if found != 14 {
		t.Errorf("有地形分派的 block 是 %d 個，宣告的是 14 個", found)
	}
	// ⚠ 遮罩不是固定的：`0x7F` 與 `0x3F` 都量到過。寫死一種會讓另一種整批落空。
	if masks[0x7F] == 0 || masks[0x3F] == 0 {
		t.Errorf("遮罩分布是 %v，兩種都應該還在", masks)
	}
	if tableForm != 5 {
		t.Errorf("用 GETTABLE 查表分派的 block 是 %d 個，宣告的是 5 個", tableForm)
	}
}

// 索引一律升冪：`Targets` 是 map，直接走訪的順序每次都不一樣，產出的對照表就
// 每跑一次 diff 一次。
func assertAscending(t *testing.T, member int, block uint8, indexes []int) {
	t.Helper()
	for i := 1; i < len(indexes); i++ {
		if indexes[i-1] >= indexes[i] {
			t.Fatalf("ECL%d/0x%02X 的索引沒有升冪排序：%v", member, block, indexes)
			return
		}
	}
}

// 守衛欄要停在第一句話之前，否則它印的是內容不是守衛。
func TestGuardStopsBeforeTheFirstLine(t *testing.T) {
	const imagePath = "../../curseoftheazurebonds.zip"
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("找不到遊戲 image，跳過")
	}
	archive, err := zip.OpenReader(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	parsed, err := dax.Parse(memberPayload(t, archive, "ECL4.DAX"))
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range parsed {
		if block.Entry.ID != 0x20 {
			continue
		}
		dispatch := Analyze(block.Data)
		// 散提爾堡城區索引 19 是「闖進民宅」，由 `4C01` 守著——那一格正是
		// spec 1142（`4C00` 區在原作是地圖本地的 bank）擋掉的內容之一。
		guard := dispatch.Guards[19]
		if !strings.Contains(guard, "COMPARE 4C01") {
			t.Fatalf("索引 19 的守衛是 %q，看不到 4C01", guard)
		}
		if strings.Contains(guard, "\"") {
			t.Fatalf("索引 19 的守衛裡混進了台詞：%q", guard)
		}
		return
	}
	t.Fatal("ECL4.DAX 裡沒有 block 0x20")
}

func memberPayload(t *testing.T, archive *zip.ReadCloser, name string) []byte {
	t.Helper()
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, name) {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		payload, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		return payload
	}
	t.Fatalf("image 裡沒有 %s", name)
	return nil
}
