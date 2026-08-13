# 第六百七十三輪：DOS 版的 EGA/VGA 暫存器重設，與一組雙緩衝的釋放

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `START.EXE` 的 `15420h`、`14689h`。

## `15420h`：把 EGA/VGA 繪圖暫存器設回預設

用 `out` 逐一寫入，索引埠與資料埠成對：

| 埠 | 索引 | 值 | EGA/VGA 暫存器 |
|---|---|---|---|
| `3CEh`／`3CFh` | 0 | `00h` | Set/Reset |
| | 1 | `00h` | Enable Set/Reset |
| | 2 | `00h` | Color Compare |
| | 3 | `00h` | Data Rotate |
| | 4 | `00h` | Read Map Select |
| | 5 | `00h` | Graphics Mode |
| | 7 | `0Fh` | Color Don't Care |
| | 8 | `0FFh` | Bit Mask |
| `3C4h`／`3C5h` | 2 | `0Fh` | Map Mask（四個平面全開） |

`3CEh`／`3CFh` 是 **Graphics Controller** 的索引／資料埠，`3C4h`／`3C5h` 是
**Sequencer**。這一整串就是「把繪圖狀態設回不做任何特殊處理」——寫入直接落到
記憶體、四個平面都可寫、bit mask 全開。

**索引 6 被跳過**（Miscellaneous，控制記憶體對映與奇偶模式）——那支不重設，
表示模式本身由別處設定，這裡只清理繪圖行為。

### 對 remake 的意含

DOS 版走的是 **EGA/VGA 平面圖形**（四個平面、Map Mask 選平面），與 PC-98 版的
文字 VRAM ＋ 三平面圖形（[spec 645](645-pc98-text-layer-primitives.md)）是**兩套
不同的顯示模型**。兩平台的繪圖程式碼不可能共用，遊戲邏輯層才有共用的可能。

## `14689h`：釋放主記錄與附帶緩衝

```text
if arg_0^ = nil then return
p := arg_0^
size := byte(p^[8]) × word(p^[11h])
if p^[13h]:p^[15h] <> nil then
    <FreeMem>(@p^[13h], size)                ← 先放附帶緩衝
<FreeMem>(arg_0^, size + 17h)                 ← 再放主記錄
arg_0^ := nil
```

`size` 由**兩個欄位相乘**得出（一個 byte × 一個 word），所以記錄自己記著「幾筆 ×
每筆多大」。

主記錄的大小是 `size + 17h`——前 `17h`（23）bytes 是標頭，其中 `+13h`..`+16h` 是
指向附帶緩衝的 far pointer。**兩塊用同一個 `size`**。

釋放順序是「先附帶、後主體」，最後把呼叫端的指標清成 nil——避免懸空指標。

`@FreeMem$qm7Pointer4Word` 是 IDA 由 Borland 簽章還原的名稱，所以這裡是 Turbo
Pascal 的 `FreeMem`，不是自製的配置器（那是
[spec 655](655-heap-block-header.md) 的另一套）。

## 明確不宣稱

- `+8`／`+11h` 兩個欄位在遊戲語意上代表什麼。
- 為什麼 Graphics Controller 索引 6 不重設。
