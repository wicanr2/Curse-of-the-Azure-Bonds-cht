package main

import (
	"archive/zip"
	"path/filepath"
	"strings"
	"testing"
)

// ★ **DOS 版沒有 BGM**，而這個結論是**列舉**不是搜尋。
//
// PC-98 那一側的音樂來自一支獨立的驅動程式（`MSCDRV.EXE`，`internal/pc98music`
// 靠它的 SHA-256 認人）。DOS image 裡沒有那支驅動，也沒有任何音樂資料檔——
// 94 個成員逐一看過，不是 grep 沒中。
//
// ⚠ 為什麼要釘住：這件事本來只寫在手打的說明文字裡，而同一份文件的另一處還寫著
// 「DOS 版有沒有那 12 首曲子還沒查」——**兩句互相矛盾，而且都沒有東西會發現**。
// 有了這條，往後 image 裡真的多了音樂檔就會紅，那時候才該重新問這個問題。
//
// ⚠ 正對照：同一份列舉找得到**音效**實際住的地方（`START.EXE`／`GAME.OVR`）。
// 所以「找不到音樂」不是因為看不見檔案。
func TestDOSImageHasNoMusicData(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("讀不到遊戲 image：%v", err)
	}
	defer archive.Close()

	executables := []string{}
	suspects := []string{}
	haveResident, haveOverlay := false, false
	for _, file := range archive.File {
		name := strings.ToUpper(filepath.Base(file.Name))
		switch name {
		case "START.EXE":
			haveResident = true
		case "GAME.OVR":
			haveOverlay = true
		}
		if strings.HasSuffix(name, ".EXE") {
			executables = append(executables, name)
		}
		// 音樂資料檔會長成什麼樣不知道，所以判準放寬：名字裡帶這些字的一律
		// 拉出來人工看。⚠ 寬的判準配上**零命中**才有意義；窄的判準零命中
		// 什麼都證明不了。
		for _, hint := range []string{"MSC", "MUS", "SONG", "BGM", "FM", "OPN", "MID", "SND"} {
			if strings.Contains(name, hint) {
				suspects = append(suspects, name)
				break
			}
		}
	}

	// 正對照：音效住的兩個檔都在，所以列舉本身是看得見東西的。
	if !haveResident || !haveOverlay {
		t.Fatalf("正對照失敗：START.EXE=%v GAME.OVR=%v——列舉本身有問題，"+
			"這時候的「沒有音樂」不能當結論", haveResident, haveOverlay)
	}
	if len(suspects) > 0 {
		t.Fatalf("image 裡出現了可能是音樂的檔案 %v："+
			"重新問「DOS 版有沒有 BGM」這個問題（spec 1192）", suspects)
	}
	// DOS image 只有遊戲本體與一支複製工具，沒有音樂驅動程式。
	want := map[string]bool{"START.EXE": true, "COPYCURS.EXE": true}
	for _, name := range executables {
		if !want[name] {
			t.Errorf("image 裡多了一個執行檔 %s：如果那是音樂驅動程式，"+
				"「DOS 版沒有 BGM」就不成立了", name)
		}
	}
}
