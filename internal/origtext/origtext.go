// Package origtext turns the original game's 8-bit strings into Go strings.
//
// 原版的角色名、怪物名與物品名都是 8-bit 位元組，沒有編碼標記。英文版是
// ASCII，中文版（珍020）是 Big5。先前的做法是 `string(bytes)`，那等於宣告
// 「這些位元組是 UTF-8」——對 ASCII 恰好成立，對 Big5 一定不成立：`0xA4 0x40`
// 這種合法的中文字會變成兩個無效位元組，之後不論比對、換行或查字模都會壞。
//
// 這裡用 Big5 解，因為 Big5 對 0x00..0x7F 與 ASCII 逐位元組相同，**英文原版的
// 解碼結果一個位元組都不會變**；中文版才會被正確組成一個字。
//
// ⚠ 本套件只負責「讀進來」。remake 有自己的存檔格式，不需要把字串寫回原版
// 版面（使用者 2026-08-16 決定），所以這裡刻意不保留原始位元組做 round-trip。
//
// ⚠ **界線是「位元組從哪來」，不是「版面長什麼樣」。** 有些版面（例如物品記錄）
// 原版與 remake sidecar 共用，而 remake 的寫入端寫的是 UTF-8。在共用版面的
// 解析函式裡一律當 Big5 解，會把 remake 自己寫的資料讀壞——要在載入原版檔案
// 的那一層解碼。
package origtext

import (
	"golang.org/x/text/encoding/traditionalchinese"
)

// Decode 解一段原版位元組。無法以 Big5 解讀時退回逐位元組的舊行為，
// 讓損壞的欄位仍然能顯示出「有東西但看不懂」，而不是整筆記錄讀取失敗。
func Decode(raw []byte) string {
	if isASCII(raw) {
		return string(raw)
	}
	decoded, err := traditionalchinese.Big5.NewDecoder().Bytes(raw)
	if err != nil {
		return string(raw)
	}
	return string(decoded)
}

// DecodeField 解一段固定寬度欄位，並去掉原版用來補滿欄位的 NUL 與空白。
//
// 先切再解是安全的：Big5 的第二個位元組只落在 0x40..0x7E 與 0xA1..0xFE，
// 不可能是 0x00 或 0x20，所以砍掉尾端的補位位元組不會砍掉半個中文字。
func DecodeField(raw []byte) string {
	end := len(raw)
	for end > 0 && (raw[end-1] == 0x00 || raw[end-1] == ' ') {
		end--
	}
	return Decode(raw[:end])
}

func isASCII(raw []byte) bool {
	for _, b := range raw {
		if b >= 0x80 {
			return false
		}
	}
	return true
}
