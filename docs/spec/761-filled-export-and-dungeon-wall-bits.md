# 761 — 補洞重匯出，以及被它救回來的迷宮牆面位元

- 工具：`tools/ida/fill_gaps_and_export.py`、`tools/fill-gaps-export.sh`、
  `scripts/filled_export_audit.py`
- 輸出：`workplace/re-sweep/<平台>/overlays/filled/<平台>-<模組>.json`

## 問題

`export_by_prologue.py` 只匯出 `is_code` 為真的位元組。IDA 從頭到尾沒認成指令
的那幾個 byte 會整段消失，而**逐條讀的時候看不出缺口**——相鄰兩條指令的 `ea`
中間少了幾格，讀起來像是連續的。

spec 736／752 記錄過它最輕微的形式（`89 EC` 掉一個 byte 變成 `in al, dx`）。
這一輪發現它也有最嚴重的形式：**整個函式本體只剩開頭幾條**。

## 作法與副作用

對每個缺口 `del_items` 後逐條 `create_insn`，一條都建不起來就跳過。全 72 個
overlay 模組加上 `START.EXE`／`PC98-GAME.EXE` 兩個常駐檔跑完。

`pc98/overlay-19` 中途出過一次意外：它的 `.i64` 損壞，`idat` 回
「Failed to initialize IDA as library (error code 4)」。**第一次用
`tools/ida.sh binary16` 重建是錯的**——那條路徑少了 `analyze_overlay.py`，
建出來的資料庫是 64-bit `metapc`，匯出裡出現 `push rbp`／`qword ptr [rsi]`
這種在 16-bit real mode 不可能存在的東西。重建一定要照 `tools/re-sweep.sh`
的原樣：`idat -A -p8086 -b0 -S/work-tools/analyze_overlay.py …`。
**驗收方式是看匯出裡有沒有 16-bit 的指令形狀，不是看檔案有沒有產生出來。**
（順帶：`tools/ida.sh binary16` 引用的 `seed_binary16.py` 根本不存在於
`tools/ida/`，那條子命令目前是壞的。）

**副作用要講清楚**：這個作法會把函式尾巴後面的字串常數也解成一堆無意義的
指令。全庫 534 支有新增指令，其中 **497 支的新增全部落在原本最後一條 `ret`
之後**——那是資料被誤解，對本體判讀沒有影響。所以補洞的輸出**不能整份照單
全收**，要用 `scripts/filled_export_audit.py` 的判準分類：

| 分類 | 判準 | 數量 | 意義 |
|---|---|---|---|
| 尾巴誤解 | 新增指令全在最後一條 `ret` 之後 | 561 | 忽略 |
| 本體截斷 | 有新增指令在最後一條 `ret` 之前 | 50 | 要逐支看 |

50 支裡面 **9 支先前已判為「已解讀」**。逐支查完之後，真正需要更正的只有兩支
——「本體截斷」這個判準會把兩種東西一起抓進來：

- **函式中段夾著資料**（VROOMM stub 陣列、跳躍表）。那些位元組本來就不是指令，
  被 `create_insn` 解成 `daa`／`verr`／`aas` 之類的東西，落在最後一條 `ret`
  之前純粹是因為它們夾在中間。這是**假陽性**。
- **函式真的被截斷**。

## 九支的逐支結果

| 函式 | 指令數 | 結果 |
|---|---|---|
| `dos overlay-14:003Eh` | 6 → 115 | **判讀完全錯誤**，見下節 |
| `pc98 overlay-14:003Eh` | 6 → 114 | 同上 |
| `pc98 overlay-22:2776h` | 62 → 78 | **少了最後一個呼叫**，見下節 |
| `dos overlay-16:351Ch` | 29 → 43 | 判讀正確——spec 756 已由原始 bytes 補讀過那 `21h` bytes |
| `pc98 PC98-GAME.EXE:10D00h` | 11 → 45 | 假陽性：本來就判為 VROOMM stub 陣列，新增的是描述子位元組 |
| `pc98 PC98-GAME.EXE:10E90h` | 12 → 14 | 同上 |
| `pc98 PC98-GAME.EXE:1B756h` | 202 → 211 | 假陽性：函式中段的 6-byte real 常數表 |
| `pc98 PC98-GAME.EXE:18EE0h` | 491 → 496 | 假陽性：新增的兩條都是 `nop` 對齊填充 |
| `pc98 PC98-GAME.EXE:17DD5h` | 43 → 44 | 同上，一條 `nop` |

`overlay-14:003Eh` 兩支原本的台帳註記寫「整個 body 只準備參數並執行
`jmp loc_14D`（body 共 277 bytes，**已逐條讀完**）」——匯出裡只有 6 條指令，
`jmp` 的目標是函式自己的 epilogue，看起來像一支轉呼叫的空殼。「已逐條讀完」
這句話當時就不成立。

## `overlay-14:003Eh` 的真正內容：清掉一個方向的牆面位元

`retf 6`，三個參數：

```pascal
procedure 清牆(dir, row, col: byte);
var 格 : byte absolute DS:7206h^[300h + (row shl 4) + col];
begin
    case dir of
        6: 格 := 格 and 3Fh;      { 清 bit 6-7 }
        4: 格 := 格 and 0CFh;     { 清 bit 4-5 }
        2: 格 := 格 and 0F3h;     { 清 bit 2-3 }
        0: 格 := 格 and 0FCh;     { 清 bit 0-1 }
    end;
end;
```

四件事一次到位：

- 地圖在 `DS:7206h` 這個遠指標的 `+300h` 之後（PC-98 是 `DS:0A2A0h`）。
- 索引是 `(row shl 4) + col`，也就是**每列 16 格**——與 spec 754 的
  `overlay-30:0556h`「兩個參數都要落在 0..15」對得上，那支就是這張地圖的
  範圍檢查。
- **每格一個 byte，四個方向各佔 2 個 bit**：方向碼 `0`／`2`／`4`／`6` 對應
  bit pair 0-1／2-3／4-5／6-7。方向碼是 2 的倍數，中間的奇數（1、3、5、7）
  在這支裡沒有分支——`case` 沒有 `else`，落到別的值就整支不做事。
- 兩個平台唯一的差別是 PC-98 用 `cmp al, 6` 取代 DOS 的
  `xor ah, ah` ＋ `cmp ax, 6`（少一條指令），邏輯相同。

這與 spec 749／756／759 那條「戰場是 50 × 25」的線是**兩張不同的地圖**：
戰場格陣在 `DS:6E92h`、每列 50 格；這張 16 × 16 的在 `DS:7206h`，位元編碼是
牆面而不是圖塊。

## `pc98 overlay-22:2776h` 補上的最後一步

spec 629 的判讀停在
`<far 013E:002A>(4Eh, t, 0, 1)`。它後面還有一個呼叫：

```pascal
<far loc_1434+3>(1, 0FFh, 0Ah, 0Fh, t);
```

（推入順序：`t` 的 4 bytes、`0Fh`、`0Ah`、`0FFh`、`1`。）其餘判讀不受影響。

## 對後續判讀的影響

從這一輪起，逐條讀一律以 `filled/` 的匯出為準，並且**只讀到最後一條 `ret`
為止**——之後的都是被誤解的資料。`prologue/` 的匯出保留不動，兩者相減就是
這份稽核。

47 支被截斷的其餘 43 支目前都還是待解讀，之後會用完整的本體讀。

## 明確不宣稱

- 沒有宣稱補洞後的匯出就沒有缺口了。`filled/` 裡仍會留 `gaps` 欄位，補不起來
  的區間照樣列出。
- 沒有宣稱 `DS:7206h^[300h]` 這張地圖的其他欄位（`+000h`..`+2FFh` 是什麼、
  2 個 bit 的四種值各代表什麼牆面）。只確定索引式與位元位置。
- 沒有宣稱方向碼 `0/2/4/6` 對應哪四個方位。
