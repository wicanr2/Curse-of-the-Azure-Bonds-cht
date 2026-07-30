# 第三百八十三輪：立石群揭露與 Myth Drannor 解鎖

狀態：`READY`

## 範圍

本輪只證明「三個後續主人已解除控制後，立石群揭露提朗瑟克斯並指向
Myth Drannor」的原始 ECL 條件，以及修正重製 VM 第一次明確 `RunFrom`
遺失既有記憶體的問題。這不代表 Myth Drannor 地圖、最終神殿或結局已完成。

## 原始 ECL 證據

DOS `ECL1.DAX` block `0x50` 的立石群事件由 `+0x01C4` 開始。它先把工作
位址 `0x7F79` 清為 0，再依序檢查：

- `0x4C59 == 1`
- `0x4C5B != 0`
- `0x4C5A == 1`

每個成立條件都把 `0x7F79` 加一。三者都成立時，原文在 `+0x03CF` 顯示
灰袍人起身、長袍滑落，揭露他就是 Tyranthraxus，並要求隊伍到
`MYTH DRANNOR` 與他會面。

三個旗標不是測試臆造值。六章原始 ECL 的 SAVE 寫入掃描得到：

- `ECL5` block `0x33` `+0x0FB6`：`0x4C59 ← 1`
- `ECL4` block `0x22` `+0x0C8E`：`0x4C5A ← 1`
- `ECL3` block `0x11` `+0x04E7`：`0x4C5B ← 0xFF`

`ECL3` block `0x12` `+0x18AE` 另會把 `0x4C5B` 暫寫為 1；立石群因此採
「非零」判定，而不是只接受 `0xFF`。

## VM 缺陷與修正

`BlockSession.SetMemoryValue` 可在事件啟動前載入 AREA、玩家存檔及劇情
工作記憶體，但舊版第一次呼叫明確位址的 `RunFrom` 時沒有把 runtime
標成已啟動。底層 runner 因而略過共享記憶體匯入，執行結束後還會用空白
memory 覆蓋 session；三個已完成旗標全部變成 0。

現在第一次 `RunFrom` 會先設定起始 PC 與 `Started`，再匯入共享 memory。
合成回歸測試證明預載的 `0x9000=7` 首輪可讀且執行後仍保存；真實
`ECL1.DAX` 回歸則證明三旗標保持 `1／1／0xFF`、工作計數為 3，並取得
上述 Tyranthraxus／Myth Drannor 原文。

## 尚未完成

- `JOURNEY ON → MYTH DRANNOR` 後的完整 ECL6／GEO6 正常玩家路徑。
- Burial Glen、森林／道路分支、最終神殿、最終戰與結局。
- 本段完整繁中 game-pack 文本、事件畫面與原版／remake 截圖對照。

