# 第二百八十五輪：剛德神殿治療服務

狀態：READY

## Reference evidence

ECL2 block `0x01` 的 SearchLocation dispatch 在 GEO2 block 1 terrain `0x92`
先發出 PICTURE `6`，當下 scene selectors 是 HEAD2 block `9`／BODY2 block `6`；
接著寫入 Area2 `EnterTemple=1` 並執行 `COMBAT (0x24)`。GEO2 parser 證明此 terrain
位於 `(0,7)` 的祭壇格。`ovr003.CMD_Combat` 消耗 EnterTemple 後呼叫
`ovr005.temple_shop()`，離開服務後從 COMBAT 下一條 instruction 恢復 ECL。

`ovr005` 列出十種治療與固定價格：

| 服務 | GP | reference effect |
|---|---:|---|
| Cure Blindness | 1000 | 移除 blinded |
| Cure Disease | 1000 | 移除六種 disease affects |
| Cure Light Wounds | 100 | `1d8` |
| Cure Serious Wounds | 350 | `2d8+1` |
| Cure Critical Wounds | 600 | `3d8+3` |
| Heal | 5000 | 補至最大 HP 再減 `1d4`，並清 blindness／disease／feeblemind |
| Neutralize Poison | 1000 | 移除 poison damage／slow poison／poisoned |
| Raise Dead | 5500 | dead／animated 回到 1 HP，清 animate-dead／poison |
| Remove Curse | 3500 | 移除 bestow-curse，解除 cursed items |
| Stone to Flesh | 2000 | stoned 回到 1 HP |

付款沿用 CityShop：先扣目前角色五種 typed coins 的 gold worth，再使用 pooled money。
進神殿會清空既有 pool。治療前會顯示價格並要求確認。

## Remake transaction

- Temple signal 與同一 result 的 PICTURE 分成兩個可恢復 boundary：先顯示原圖，
  Enter 後才進神殿，不可讓 service signal 吃掉 PICTURE。
- 神殿提供治療、查看、集中／分配金幣、估價與離開；治療清單保存十個固定價格。
- 治療使用 deterministic RNG，更新 Character HP／health／effects／equipment，
  並同步既有 combat fighter projection。
- 新增 internal `stoned` health state；Raise Dead 的原版 Constitution／最大 HP
  penalty 尚未具備完整多職業 stat transaction，因此本輪只實作已確定的復活、
  1 HP 與 affect 清理，不宣稱該 penalty 完成。
- Scene compositor 改為建立可擴張 canvas：BODY 位於 `y+5`，HEAD masked pixels
  最後覆蓋。這修正 HEAD9／BODY6 不同 selector 時，舊同號預產圖缺失與頭部裁切。

## Regression

- synthetic EnterTemple SAVE → COMBAT 產生 temple signal、清旗標並可 resume PRINT；
- cure-light-wounds 驗證 100 GP typed-coin 扣款與 `1d8` healing；
- remove-curse 驗證 effect 與 cursed equipment；
- formal new game → Tilverton `(0,7)` → PICTURE 6／HEAD9＋BODY6 → Temple →
  Cure Light Wounds／確認／付款 → EXIT → 返回原格；
- `ComposeHeadBody` 驗證 canvas 高度與 BODY y offset。
