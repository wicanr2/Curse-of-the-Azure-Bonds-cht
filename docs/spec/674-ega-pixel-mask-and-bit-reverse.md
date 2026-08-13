# 第六百七十四輪：EGA 的顏色比對遮罩與 byte 位元反轉

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `START.EXE` 的 `12970h`、`133E2h`。

## `12970h`：算出「哪幾個像素等於某個顏色」

```text
mask := 0
target_hi := arg_4 shl 4                    ← 拿來比高 nibble
bit := 80h
for i := 0 to 3 do
    if (arg_0[i] and 0F0h) = target_hi then mask := mask + bit
    bit := bit shr 1
    if (arg_0[i] and 0Fh) = arg_4 then mask := mask + bit
    bit := bit shr 1
return mask
```

**4 bytes ＝ 8 個 4-bit 像素**，掃描順序是「每個 byte 先高 nibble 後低 nibble」，
產出的 8-bit 遮罩由 `bit 7` 對應第 0 個像素、`bit 0` 對應第 7 個。

這正好餵給 EGA 的 **Bit Mask 暫存器**（[spec 673](673-ega-graphics-controller-reset.md)
的 `3CEh` 索引 8）：先算出「這 8 個像素裡哪幾個是目標顏色」，再用遮罩只寫那幾個。

高 nibble 在前的順序與 [spec 672](672-numeric-menu-input-and-nibble-array.md) 的
`129EBh` 一致——同一份程式裡的 nibble 打包順序是統一的。

`bit` 用 `shr` 遞減，兩次除以 2；`mask` 用 `add` 累加（各 bit 不重疊，等同 `or`）。

## `133E2h`：byte 位元反轉

八段固定的 `and` ＋ `shl`／`shr`：

| 來源 bit | 移動 | 目的 bit |
|---|---|---|
| 0 | `shl 7` | 7 |
| 1 | `shl 5` | 6 |
| 2 | `shl 3` | 5 |
| 3 | `shl 1` | 4 |
| 4 | `shr 1` | 3 |
| 5 | `shr 3` | 2 |
| 6 | `shr 5` | 1 |
| 7 | `shr 7` | 0 |

**逐值驗算**：把這八段照樣實作一遍，對 `0..255` 全部比對「位元字串反轉」，
**不一致 0 筆**。所以這支確定是位元反轉，不是別的位元重排。

第一段沒有先 `and 1`——直接 `shl ax, 7`，靠「結果只存回 byte」把其餘位元擠掉。
其他七段都有遮罩。

位元反轉在 EGA 繪圖裡用於**左右鏡射**或**位元順序轉換**（VRAM 裡 bit 7 是最左邊的
像素，而一般陣列索引是由左到右遞增）。

## 明確不宣稱

- `133E2h` 的呼叫端拿反轉結果做什麼（鏡射、還是像素順序轉換）。
- `12970h` 產出的遮罩是直接寫進 Bit Mask 暫存器，還是先經過別的處理。
