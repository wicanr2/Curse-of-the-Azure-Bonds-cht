# 928 — `START.EXE` 裡那 36 支「函式」是 overlay entry stub，不是程式碼

- 證據等級：`exact`（控制記錄逐筆解出並用鏈驗證，最後一段結束於 `.OVR` 檔案大小）
- 作法見 spec 783；版面依據
  `~/.claude/knowledge-base/retro/borland-tpov-overlay-re.md`

## 結論

DOS `START.EXE` 的反組譯裡有 **36 支以 `int 3Fh` 開頭的「函式」**
（它們不在 `docs/audit/function-index/dos-START.EXE.md` 裡，
`cmd/re-ledger` 沒有把它們當函式，所以**不影響台帳數字**）。
逐一定位之後，它們**全部落在 36 段 overlay 控制記錄的 `stub offset 2Ah`**，
也就是每一段的 **entry index 2**：

```
IDA 10390h → 控制記錄 #00（IDA 10366h）＋ 2Ah → entry 2
IDA 103C0h → 控制記錄 #01（IDA 10396h）＋ 2Ah → entry 2
…（36 筆，全部落在 index 2）
```

entry stub 是 5 bytes 的 `CD 3F, handler_offset:u16le, flags:u8`——
**IDA 把 `CD 3F` 後面的資料當成指令反組譯**，才生出
`push cx`／`aas`／`rcl word ptr [bp+di], 0` 這類看起來荒謬的序列。
這 36 支沒有任何遊戲語意，是 overlay 派送表的一部分。
**函式索引把它們排除掉是對的**——本規格記錄的是「為什麼它們看起來像函式」，
以及由此解出的控制表。

## 控制表已經完整解出

`docs/audit/tpov-overlay-control-table.md` 收錄 36 段的
`code offset`／`code size`／`relocation size`／`entry count`
以及**全部 871 筆 entry 的 stub offset 與 handler offset**。

驗證方式（知識庫指定）：第 0 段的 `code offset` 是 `000004h`，
之後每一段等於前一段的 `code offset ＋ code size ＋ reloc size`，
最後一段結束於 **`04275Ah`** ＝ `GAME.OVR` 的檔案大小，逐位元組吻合。
只靠「找到 `CD 3F`」會把常駐碼裡剛好出現的位元組誤判——`START.EXE` 全檔
共有 **908 個 `CD 3F`**，其中 36 個是控制記錄開頭、871 個是 entry stub，
剩下 1 個不在鏈上。

換算：**IDA linear ＝ 檔案位移 ＋ `0F826h`**（36/36 驗證）。

## ★ 這解釋了 spec 927 的撞號

spec 927 記到 PC-98 `overlay-13:029F4h` 有一條 IDA 顯示成 `call sub_11CC`、
實際是 `9A 5C 00 17 01` ＝ `call far ptr 0117h:005Ch` 的指令。
套上 stub 版面：

```
stub offset 5Ch  →  entry index ＝ (5Ch − 20h) ÷ 5 ＝ 12
```

`5Ch` 是合法的 `20h ＋ 5i`，所以那是**段 `0117h` 的 entry 12**，
不是 `overlay-13` 自己的偏移 `11CCh`。
知識庫把這一條列為第一號陷阱：**「只比對 stub offset、不比對 segment」**——
`20h ＋ 5i` 這種值撞號是常態不是意外。

## 引用近位址前先看第一個位元組

- `E8 <disp:u16>` ＝ near call，目標是**本模組偏移**。
- `9A <off:u16> <seg:u16>` ＝ far call，`off` 是**別的段的 stub offset**，
  必須先解 segment 才知道是哪一段。

IDA 產生的符號名（`sub_XXXX`）在第二種情況下會誤導，
函式索引裡因此會出現「落在別支函式 body 內部」的幻影邊界碎片。

## 明確不宣稱

- 沒有宣稱 segment 值（`0117h` 等）對應哪一段 overlay——
  那需要另外建「segment → 控制記錄」的對照，本規格沒有做。
- 沒有宣稱 entry stub 的 `flags` byte 是什麼。
- 沒有宣稱 PC-98 側的控制表（本規格只解 DOS `START.EXE`）。
- 沒有宣稱那 1 個不在鏈上的 `CD 3F` 是什麼。
