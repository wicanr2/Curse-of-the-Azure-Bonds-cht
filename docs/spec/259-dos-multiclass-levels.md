# 第 259 輪：DOS multi-class level preservation

狀態：`READY`（限 raw player class／level preservation 與 primary-class adapter）

## 已確認

- reference `ClassId` 使用 `0..7` 表示單一職業，`8..16` 表示 cleric/fighter/ranger/magic-user/thief 的多職業組合。
- Player `ClassLevel[8]` 位於 `.SAV/.GUY` record `0x109..0x110`；`multiclassLevel` 位於 `0xE6`。
- 目前 remake `Character.Class` 仍是單一 primary-class 欄位，因此 raw multi-class ID 先投影為主要規則分支（`mc_c_*`→cleric、`mc_f_*`→fighter、`mc_mu_t`→magic-user）。

## 實作

- `DOSPlayerRecord` 與 `party.Character` 保存完整 `[8]uint8 ClassLevels` 與 `MulticlassLevel`。
- parser 接受 raw class IDs `8..16`，以 `0xE6`／各 class level 決定目前 level。
- `PatchDOSPlayerRecord` 只在 Character 帶有 class-level data 時回寫 `0x109..0x110`，不覆蓋未知 bytes。

## 保留邊界

多職業的獨立命中骰、法術／能力限制、training／duel-class UI 與 remake character-creation 選單仍待接入；本輪不把 primary-class projection 誤報成完整 multi-class rules。

## 回歸

synthetic human `mc_f_mu` record 驗證 raw class `13`、fighter／magic-user levels、primary projection 與 DOS writeback round-trip。
