# 1149 — `24h COMBAT` 的三選一順序：場上有怪就直接打

- 證據等級：`exact`（DOS `overlay-02:179Ah` 起的分派逐條讀出）
- 三選一的整體模型見 spec 1095；`8B69h` 由 `1Ch` 清掉見 spec 1145
- 逐處分類見 `docs/audit/ecl-combat-sites.md`

## 分派

```asm
17A0  inc  word ptr ds:4FB4h            ; PC 推進
17A4  cmp  byte ptr ds:8B69h, 0
17A9  jz   short loc_17AE
17AB  jmp  sub_1956                     ; ★ 有怪 → 打
17AE  cmp  byte ptr ds:8B56h, 0
17B3  jz   short loc_17B8
17B5  jmp  sub_1956                     ; ★ 有怪 → 打
17B8  les  di, ds:4F9Dh
17BC  cmp  word ptr es:[di+6D8h], 1     ; 這之後才輪到商店
17C2  jz   short loc_17C7
17C4  jmp  near ptr sub_184B            ; 都不是 → 另一支
17C7  bank1^[6D8h] := 0                 ; 讀到就清零
```

⇒ **順序是「有怪 → 打」在最前面**，商店旗標排在它後面。場上有怪而商店旗標又是 1
的時候，原作打，不開店。

`bank1^[6D8h]` ＝ ECL 格 `7F6Ch`（換算 `0x7C00 ＋ X ÷ 2`，見 spec 1146）——這是第
四個獨立對上的點。`8B69h` 就是 `1Ch CLEARMONSTERS` 清的「有怪要打」旗標
（spec 1145），在 remake 對應的是怪物鏈非空。

## remake

分派順序改成照抄：先看怪物鏈，再看商店旗標，再看營地旗標。

⚠ 兩個請求旗標在 remake **都沒有 producer**（`TestCampRequestFlagHasNoProducer`
釘住 `0x7EE2`），所以今天觀察不到差別；照著寫是為了補上寫入端時不會先走錯支。
`0x7EE2` 的語意是**神殿**，remake 原本的 `TempleRequested` 是對的（spec 1182）。

⚠ 仍是 `partial`：那 46 處沒擺過怪的 `24h` 在 remake 走的是別的機制
（地城裡「零隻怪就續跑 ECL」、世界地圖上是地點選單），不是 `24h` 的請求旗標。

## `sub_184B` 已經讀完

那一支自己再分兩路（spec 1182）：`bank1^[5C4h]`（ECL 格 `7EE2h`）＝ 1 走
**神殿**（`overlay-04` ＝ TEMPLE），否則走**戰後處理**（`overlay-05` ＝ POSTCOM）。
所以 `24h` 是**四選一**，而第四支正是那 46 處「沒擺過怪的 `24h`」的去向。

⚠ 本規格先前依 spec 1030 把 `overlay-04` 當成營地——那是神殿，營地是
`overlay-15`。

## 明確不宣稱

- 沒有宣稱 `DS:8B56h` 是什麼（只知道它與 `8B69h` 並列，非 0 就去打）。
- 沒有宣稱 `sub_1956`（打）的內容。
