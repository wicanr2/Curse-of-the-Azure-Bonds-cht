# 第二百五十一輪：DOS save format／age evidence

## 狀態

READY

## Contract

`CHRDAT{slot}{1..6}.SAV`／`.GUY` 的角色 record 以 `0x74` race、`0x75` class、
`0x76..0x77` signed little-endian age、`0x78` max HP 為目前已驗證欄位。只修改年齡時，
應 patch 兩個 age bytes，保留其餘 raw record 與 `.FX/.SWG` sidecars。

`SAVGAM?.DAT` 是另一個 slot-level container；其 shared ECL flag window 的公式為
`((ecl_address - 0x4C00) * 2) + 0x201`，不能用來推算 player age。

## 來源與邊界

- 本地 DOS Manual／RefCard：確認遊戲的 save／load 操作，但沒有 raw byte offset table。
- CoAB technical guide：確認角色 record offsets、race/class enum 與 save naming。
- `/tmp/coab-reference`：以 `Gbl.RaceClasses`、`race_ages`、`NormalizeClock` 與 player
  parser 交叉核對 class index、半獸人 fighter `13+1d4` 及 age writeback。

這份規格只定義可安全 patch 的 evidence boundary；尚未完成完整 character serializer、
multi-class record、所有 `.FX/.SWG` 欄位與原版跨檔案刪除副作用。
