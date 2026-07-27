# 第三百零九輪：伊弗利特與哈普解放

狀態：`READY`

## terrain 0x88 barn

ECL5 block `0x31` SearchLocation 的 terrain `0x88` 指向 `+0x076D`。若
`4C5E != 1`，事件顯示空穀倉裡的伊弗利特與黑暗精靈黨羽，接著執行
`SETUP MONSTER 0x34,2,0x34`、`APPROACH / DELAY / APPROACH`，再依已擊敗
巡邏數 `4C47` 計算黨羽數，結果上限為 `6`。

本輪 regression 在 `4C47=1` 時依原始 arithmetic 建立：

- MON5 `0x34` `EFREETI` ×1，icon `0x34`，HP 55；
- MON5 `0x32` `DARK ELF MAGE` ×6，icon `0x32`；
- MON5 `0x33` `DARK ELF CLERIC` ×6，icon `0x33`。

## APPROACH signal

Opcode `0x0D APPROACH` 不是 unsupported command。reference 行為是將
encounter sprite 向隊伍推進；bounded VM 現以 `ApproachCount` 保存
presentation intent，BlockSession 跨 block 聚合。幾何距離與逐幀動畫仍由
frontend adapter 負責。

State 過去忽略 interactive runner error，會把 unsupported opcode 偽裝成空白
事件。本輪改為回傳 runtime error；只有已明確支援的 opcode 才能繼續主線。

## victory continuation

勝利後 ECL：

- `4C01=5`；
- `4C5E=1`，表示伊弗利特已被擊敗／哈普解放；
- 在屍體上找到標示村莊與洞穴的地圖；
- 顯示 PICTURE `0x32`（decimal 50）與村民歡呼；
- 長老感謝隊伍，並指出黑暗精靈受附近法師塔控制。

所有 Continue 與 PICTURE 關閉都恢復同一 ECL runtime，最後停在法師塔主線
提示；不能在戰後先行重建 Hap place menu。

## UI contract

原始 PICTURE 50 與 MON5 icon 採 nearest-neighbour 整數倍放大；640×480
畫布上的伊弗利特威脅、地圖線索、群眾與長老對話均以 24px 繁中重繪，
緊湊戰鬥資訊可使用約 16×15。
