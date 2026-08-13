# 第五百七十五輪：原版亂數完整解出、PC-98 VRAM 與 EMS

狀態：`READY`。等級：`exact`。日期：2026-08-14

## 原版的亂數：三塊全部到齊

```text
Randomize:                                  ← DOS 1B6CCh
    RandSeed := 系統時間（INT 21h AH=2Ch）

LCG（DOS sub_1B694／PC-98 sub_1B7F4）:
    RandSeed := (RandSeed * 134775813 + 1) mod 2^32

Random(n)（DOS @Random$q4Word 1B645h／PC-98 1B7A5h）:
    呼叫 LCG
    if n = 0 then return 0
    return (RandSeed shr 16) mod n          ← 只用高 16 位
```

### 乘數怎麼確定的

程式裡沒有 `08088405h` 這個常數。`mul cs:word_1B6CA` 只乘了低位
`8405h`，高位 `0808h` 是用一串 shift／`add ch, cl`／`add dh, bl` 湊出來的
（Turbo Pascal RTL 避免 8086 沒有的 32-bit 乘法）：

| 指令 | 效果 |
|---|---|
| `shl cx,1` ×3 → `add ch, cl` | `cx := L × 0808h mod 2^16` |
| `add dx,bx` | `+ H × 0001h` |
| `shl bx,1` ×2 → `add dx,bx` | `+ H × 0004h` |
| `add dh, bl` | `+ H × 0400h` |
| `shl bx,5` → `add dh, bl` | `+ H × 8000h` |

後四項合計 `H × 8405h`。整條鏈已用 20 萬組隨機種子加邊界值
（`0`／`1`／`FFFFFFFFh`）與 `seed × 134775813 + 1` 逐一比對，**全部相符**。

### `Random(n)` 只用高 16 位

`call` 之後緊接 `xor ax, ax`——低位字被丟掉，`xchg ax, dx` 把高位字搬進
`ax`、`dx` 清零，`div bx` 於是是「高位字 ÷ n」，餘數即結果。

這有兩個後果，remake 必須照做才會得到同一串數字：

1. 序列的實際狀態是 32 位，但**每次只有高 16 位參與取值**。
2. `(RandSeed shr 16) mod n` 對非 2 冪的 `n` 有取模偏差（例如 `Random(6)`
   的六個結果不等機率）。這是原版行為，不是 bug，不要「修正」它。

種子 0 起算，`ROLLDICE(1, 6)`（＝ `Random(6)+1`）的前十次是
`1, 5, 6, 5, 1, 2, 6, 2, 6, 3`——可直接當 remake 的回歸測試向量。

兩平台逐指令相同（種子變數 DOS `dword_20998`／PC-98 `dword_23B0A`）。

## PC-98 圖形 VRAM 與 EMS

- `17239h`：把回傳的 far pointer 設成 **`A800:0000`**——PC-98 圖形 VRAM
  起始段。
- `1A485h`：以 `INT 21h AX=3567h` 取 `INT 67h` 向量，把該處理常式起始 8
  bytes 與 `CS:558h` 的簽章比對 ⇒ **EMS 驅動偵測**。
- `overlay-18:1908h`：`si` 從 0 起、每次加 `2B67h` 再 `and 7FFFh`，重複
  `8000h` 次。`2B67h` 與 `8000h` 互質，故會走遍 32 KB 平面的每一格而順序
  打散 ⇒ 常見的「隨機化淡入／淡出」實作。

## 明確不宣稱

- `overlay-18:1908h` 每格做什麼（`sub_191E` 未讀）。
- VRAM 平面配置與屬性平面位址。

## 修訂

本文件初版（commit `b63805a`）寫成「以 32-bit 亂數對 n 取餘數」，漏看了
`xor ax, ax`；正確的是**只取高 16 位**。以該版本推導的序列不會對上原版。
