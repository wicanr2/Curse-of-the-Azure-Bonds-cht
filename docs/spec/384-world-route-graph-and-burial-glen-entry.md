# 第 384 輪：世界路線圖與 Burial Glen 入口

狀態：`READY`

## 結論

- `ECL1.DAX` block `0x50` 尾端保存 14×4 的有序、有向目的地表；`FF` 是空格。
- UI 顯示完整目的地列時，作品 adapter 把該列的原始 location byte 依選項
  順序投影到 `4C02..4C05`。這些 byte 不可一律減一。
- 條件式劇情可能隱藏目的地；此時保留原 ECL 已準備的縮短表，不用完整 JSON
  adjacency row 覆寫。
- AREA 抵達時以 `4C9C` 提供 location，呼叫 ECL1 lifecycle entry 1。
  Myth Drannor 的值是 `13`。
- 選擇 `ENTER CITY` 後會切換至 ECL6 block `0x40`，要求
  `LOAD FILES 40,02,FF`、`LOAD PIECES 17,18,16`，並顯示龍盔指出
  Tyranthraxus 位於北方。

## 資料與程式界線

- 14 個地點與 adjacency graph 位於 CoAB JSON；共用 engine 只提供 schema
  與 graph reference validation。
- `4C02..4C05`、`4C9B`、`4C9C` 與 ECL1 lifecycle 都是 CoAB legacy adapter，
  不得移入共用 engine。
- `myth-drannor.tyranthraxus-reveal`、`myth-drannor.edge` 與
  `myth-drannor.helm-north` 都以穩定 `message_id` 綁定。產品測試從受測
  game pack 查文字，不複製繁中句子。

## 驗證

- 原始 ECL 測試驗證 Standing Stone 揭露、四個目的地、Myth Drannor
  route、`4C9C=13`、ECL6 `0x40` 與 load signals。
- 正常 State 玩家路徑只使用畫面上的按鍵、Continue、Journey On、
  Wilderness、Enter City，最後進入 GEO6 block `0x40`。
- 既有 Tilverton→Ashabenford→Standing Stone→Essembra→Hap 長路徑仍須
  同時通過，防止完整 adjacency row 覆寫條件式目的地。

## 尚未完成

- 現有 wilderness map 只完成有界 travel handoff，尚未還原全部 terrain、
  隨機遭遇、路程時間與每條 route 的 AREA movement。
- Burial Glen 只完成入口與第一人稱 geometry handoff；區域事件、戰鬥、
  最終神殿及結局仍未完成。
