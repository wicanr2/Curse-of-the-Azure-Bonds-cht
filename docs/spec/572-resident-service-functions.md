# 第五百七十二輪：resident 服務函式逐一解讀（第一、二批）

狀態：`READY`（限本文件列出的 24 個函式）。日期：2026-08-14

## 範圍

DOS `START.EXE` 內、形狀不屬第 569 輪可機械判定集合的小函式，改為逐條人工
閱讀。本輪完成 24 個：21 個 `已解讀`／`exact`、3 個 `不阻塞`（RTL 日期時間）。

## 玩家可見的 RTL：`SOUND`／`NOSOUND`／`DELAY` 已讀

第 566 輪刻意把這幾支排除在「不阻塞」之外，本輪把它們讀完：

| 位址 | 內容 |
|---|---|
| `19D76h` | `@SOUND`：`divisor := 1234DDh ÷ 頻率`（頻率 `<= 12h` 直接返回）；`in 61h` 開喇叭閘、`out 43h` 送 `0B6h`（PIT channel 2 方波模式）、分兩次 `out 42h` 送 divisor 高低位元組 |
| `19DA3h` | `@NOSOUND`：`in 61h` → `and 0FCh` → `out 61h`，同時關閉閘門與計時器輸出 |
| `19D4Eh` | `@DELAY`：反覆讀 `0000:0000` 的 BIOS timer 位元組，每次變動遞減計數 |
| `19D6Eh` | `@DELAY` 的忙碌等待迴圈：`cmp al, es:[di]` 相同就 `loop` |

`1234DDh` ＝ 1,193,181，是 PIT 的輸入頻率。這是 PC 喇叭發聲的標準作法，
**remake 若要重現原版音高，除數公式就是這一條**；`DELAY` 綁 BIOS timer tick
（約 18.2 Hz）而不是固定迴圈，所以原版的等待時間與 CPU 速度無關。

## 顯示與 BIOS

| 位址 | 內容 |
|---|---|
| `1A0B8h` | BIOS 視訊呼叫的統一入口：保存 `si/di/es` → `int 10h` → 還原。被呼叫 18 次 |
| `12A47h` | 設定顯示模式：`Registers` record 填 `AH=0`／`AL=arg`，呼叫 `INTR(10h)`；呼叫前把舊模式存進 `byte_211A4`，呼叫後把新模式記進 `byte_211A5` |
| `12A77h` | 以 `byte_211A4` 還原先前的顯示模式 |
| `154C4h` | `word_211A0 := (arg << 9) + 0A000h` — 由列號算出顯示記憶體段位址 |
| `19F84h` | 捲動一行：超出視窗下界時以 `INT 10h AX=0601h` 上捲 |

## 記憶體、檔案與格式化

| 位址 | 內容 |
|---|---|
| `1637Dh` / `16360h` | `@GetMem` / `@FreeMem`，大小先 `(size+7) and 0FFF8h` 對齊到 8 bytes |
| `16C17h` | 依序 `@Close` 兩個 `File` 參數 |
| `1A1F1h` | 開檔：Pascal 字串轉 ASCIIZ 後 `INT 21h AX=3D00h` 唯讀開啟 |
| `1A42Ch` | `(es:[8] + 0Fh) >> 4`，位元組數無條件進位換算成 paragraph |
| `1A6D4h` | 輸出 `CS` 內的 ASCIIZ 字串，逐字元呼叫 `sub_1A716` |
| `1A6E2h` / `1A6EEh` | 十進位輸出：以除數 100／10 印高位，`sub_1A6EE` 印一位並保留餘數 |
| `17240h` | 把 `dword_21154` 所指的 word 清 0 |
| `197B4h` | 跨段寫入：切到 `seg045` 寫 `byte_17286`，前後保存 `AX`／`DS` |
| `15F53h` | `@DELAY(byte_21169 × 100)` |

`1985Ch`／`1987Eh`／`19892h`／`198B7h` 是 Turbo Pascal `Dos` 單元的
`GetDate`／`SetDate`／`GetTime`／`SetTime` 實作（`INT 21h AH=2Ah..2Dh`），
標 `不阻塞`。

## 這份規格明確不宣稱

- 各全域（`byte_211A4`／`word_211A0`／`byte_24E60..64` 等）在遊戲層的語意。
  本輪只描述這些函式對它們做了什麼。
- `sub_1A716`（字元輸出）與 `sub_16FAD`（按鍵取值）本體。
- PC-98 的對應函式：該平台 resident 沒有符號，也尚未逐一核對。
