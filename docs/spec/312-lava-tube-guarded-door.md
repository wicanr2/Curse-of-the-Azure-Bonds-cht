# 第三百一十二輪：熔岩洞火蜥蜴守門巡邏

狀態：`READY`

## 反組譯證據

- ECL5 block `0x32` SearchLocation 在 `+0x05B5` 將 `C04F & 0x7F` 送入
  `ON GOTO`。selector 1 對應第一個 table target；因此 terrain `0x8A`
  對應第十個 target `+0x10C6`，不是零起算誤判的 terrain `0x89`。
- GEO5 block `0x32` 的 `(9,10)` 是可驗證的 terrain `0x8A` 格。
- handler 僅在 `4C60=0` 且 `4C48 & 0x08=0` 時觸發，顯示由火蜥蜴率領的
  守門巡邏，建立 MON5 `0x39×3 + 0x31×3 + 0x33×1`。
- 勝利後令 `4C48 |= 0x08`，並顯示腦中夢境般聲音所發出的「前方危機重重，
  務必做好萬全準備」警告。

## 實作與中文契約

- `SALAMANDER` 沿用「火蜥蜴」，黑暗精靈戰士／牧師沿用既有 MON5 本地化。
- 戰勝後 ECL 可直接停在 press-button menu，不一定先暴露泛用
  「戰鬥勝利」畫面；State 必須以同一 resumable PC 接續夢境警告。
- terrain dispatch 文件與測試應保存 selector base，不能只記 handler offset。
- 事件沿用 640×480、24px 中文敘事及原始 icon nearest-neighbour 整數放大契約。

## 驗收

- 真實長流程由提爾佛頓抵達熔岩洞，打贏入口伏擊後，以原始 GEO `(9,10)`
  觸發守門戰。
- 驗證敵軍數量 `3/3/1`、夢境警告、`4C48 & 0x08` 及最後返回 block `0x32`
  dungeon。
