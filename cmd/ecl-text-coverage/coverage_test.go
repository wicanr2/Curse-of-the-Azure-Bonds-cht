package main

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

// 分母（walkPages）與比對（walkRuns）走的是同一條控制流，但狀態不同：
// 前者只帶一個位元、一定走得完；後者帶整份已累積文字，可能碰到上限。
//
// 因此**唯一成立的方向是 walkRuns ⊆ walkPages**。這條測試釘住那個包含關係，
// 並把兩者的差額印出來——差額不是 0 不算失敗（那正是截斷的代價），
// 但它必須是**看得到的數字**，不能只在報告裡寫一句「可能有頁沒進分母」。
func TestPageWalkCoversEveryPageTheRunWalkFinds(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer archive.Close()

	totalPages, missedByRunWalk := 0, 0
	for _, member := range members {
		data, err := readMember(&archive.Reader, member)
		if err != nil {
			t.Fatal(err)
		}
		blocks, err := dax.Parse(data)
		if err != nil {
			t.Fatalf("%s: %v", member, err)
		}
		for _, block := range blocks {
			label := fmt.Sprintf("%s/0x%02X", member, block.Entry.ID)
			pages, err := walkPages(block.Data)
			if err != nil {
				t.Fatalf("%s page walk: %v", label, err)
			}
			paths, runPages, err := walkRuns(label, block.Data)
			if err != nil {
				t.Fatalf("%s run walk: %v", label, err)
			}
			_ = paths
			totalPages += len(pages)
			for offset := range runPages {
				if _, ok := pages[offset]; !ok {
					t.Fatalf("%s %s 只被帶文字的走訪找到，分母走訪沒找到——"+
						"分母是過近似，這個方向不該發生", label, offset)
				}
			}
			for offset := range pages {
				if _, ok := runPages[offset]; !ok {
					missedByRunWalk++
				}
			}
		}
	}
	if totalPages == 0 {
		t.Fatal("分母走訪一頁都沒找到")
	}
	t.Logf("分母 %d 頁，其中帶文字的走訪沒走到 %d 頁（截斷的代價，這些頁一律算 unmatched）",
		totalPages, missedByRunWalk)
}
