# READY 340：摩安德之坑最後阻擊與離隊

## 成果

- ECL3 block `0x11` SearchLocation selector `0x0F` 已定位至 `+0x1249`。
- 上層 `(0,12)` 觸發最後阻擊：MON3 `0x11×10`、`0x1C×5`、`0x19×5`。
- 戰勝後恢復同一份 dungeon ECL runtime，不誤回世界選單。
- `(0,11,W)` 執行出口 handler，驗證 `4C5B=FF`、`7F12=1` 與
  `NEWECL 0x51`。
- 愛麗雅絲、龍餌告別已繁中化，並同步移出 roster 與 combat projection；
  Alias 死亡分支保留龍餌抱著遺體前往希爾斯法的作品語意。
- packed-text CLI 現顯示 block-relative offset，便於後續 Gold Box
  disassembly／文字／selector 交叉比對。

## 驗收

Real-image integration 從摩貢、摩安德殘軀、護手與祭壇藏寶一路跑到：

1. 上樓並踏入 `(0,12)`；
2. 顯示繁中最後阻擊；
3. 驗證 20 名敵人的原始 MON3 組成；
4. 戰勝並回到 dungeon；
5. 從 `(0,11,W)` 離坑；
6. 驗證兩個劇情旗標、block `0x51`、NPC 離隊與繁中告別；
7. Continue 後回到 `ENTER CITY／JOURNEY ON／CAMP` 世界選單。

`go test ./internal/ecl ./internal/game` 必須通過；提交前另以 Docker／Xvfb
執行 `go test ./...`。
