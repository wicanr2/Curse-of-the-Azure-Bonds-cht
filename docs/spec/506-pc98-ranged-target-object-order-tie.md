# 第五百零六輪：PC-98 遠程目標同距物件順序邊界

狀態：`READY`（限已投影候選的同距排序與 bounded ranged-target adapter；不宣稱
完整 `PICKTARGET` producer、pointer chain、原版亂數或逐像素戰鬥演出）

## 本輪結論

第 416、503、505 輪的證據已足以把 `SelectRangedCombatTarget` 的同距排序從
「stable fighter ID 暫代」收斂成較接近原作的 bounded 規則：兩個候選都帶有
非零且不同的 `LegacyObjectID` 時，先依一基底原始 combat-object 身分排序；若
任一方沒有完整 legacy projection，才回到 stable fighter ID。距離排序、一次
抽樣、不可見候選移除與最多 20 次嘗試不變。

這是 `strong inference`，不是 `exact` 的完整 comparator。`PICKTARGET` 的
候選 producer、同距離前的所有欄位、原版亂數 helper 與 pointer-chain 仍未由
目前 bytes 閉合；本輪只消除 remake 自己的字典序 tie-break，不把它寫成原版
所有情況的證明。

## 證據與位址基準

本輪沿用既有唯讀輸入，不重新命名或覆寫 IDA 原始識別：

| 證據 | 定位／摘要 | 等級 |
|---|---|---|
| PC-98 `GAME.EXE` | SHA-256 `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | `exact input` |
| PC-98 `GAME.OVR` | SHA-256 `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | `exact input` |
| overlay 13 `PICKTARGET` | local `3D7Fh..3F5Ch`；依候選數抽樣、`CHECKTARGET`、失敗移除，最多 `14h` 重試 | `exact control flow` |
| overlay 24 候選 builder | `PICKTARGET` 的 raw far call `9A C0 00 4A 01`，解碼為 `014A:00C0h`；完整 producer 欄位仍未命名 | `exact call／unknown semantics` |
| `LegacyObjectID` 投影 | 由保留的 `CHARACTERLIST／OBJECTLIST` 一基底身分連接至 remake fighter | `exact adapter identity／strong inference for tie use` |

`3D7Fh` 等數值是 overlay-local；`014A:00C0h` 是 raw far-pointer 位址空間，
不可與 overlay-local `00C0h` 或其他 segment offset 直接合併。IDA Pro 9.4 只
操作 Docker 內 code-only 副本；原始 `.EXE`、`.OVR`、symbol table 與 `.i64`
仍唯讀保存。

## Remake 契約

`internal/combat.SelectRangedCombatTarget` 現依下列順序產生候選：

1. 先以雙方 footprint、加權距離、地形射線與射程建立合法候選。
2. 距離較近者優先。
3. 距離相同且雙方均有不同非零 `LegacyObjectID` 時，較小的原始物件身分
   優先。
4. legacy projection 不完整時，使用 stable fighter ID，確保 map 迭代不造成
   不可重現結果。
5. 使用同一 Battle PRNG 抽樣；visibility 失敗的候選移除後重抽，最多 20 次。

這個 comparator 只服務已有 ranged-target consumer；它不會建立候選、不解讀
法術、不把 stable ID 寫回原始存檔，也不取代 `combat/quicktarget.SelectOne`
的 JSON／legacy candidate adapter。

## 回歸驗證

`TestSelectRangedCombatTargetUsesLegacyObjectOrderForEqualDistance` 使用兩個距離
完全相同、stable ID 刻意與 `LegacyObjectID` 相反的候選，並以同一 seed 計算預期
的一次抽樣結果。測試因此驗證排序投影與 PRNG 消耗使用同一序列，而不是複製任何
可編輯中文或 game-pack 顯示文字。

## 尚未關閉

- 原版候選 builder 的完整欄位、排序 comparator 與同距離 tie 仍需 runtime trace
  或完整資料流，才能從 `strong inference` 升為 `exact`。
- `3E01:142Dh` 的亂數演算法、Quick 法術專用 candidate policy、完整敵方 AI、
  弓箭／法術逐幀效果、音效與戰後續跑仍未完成。
- 因此本輪不能宣稱完整戰鬥、完整可通關或整作 remake 完成。
