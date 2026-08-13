# 第五百六十七輪：packed text operand 與 bank 1 欄位對照

狀態：`READY`（限本文件列出的 operand 路徑與已抽出的對照列）。
日期：2026-08-14

## packed text（operand code `80h`）

`READVAR` 對 `80h` 走獨立路徑（PC-98 `overlay-07:012Ah` 起）：

1. 維護一個**字串槽計數器**，每遇到一個 `80h`／`81h` operand 就加一。
2. 讀 low byte 當**長度／索引**。
3. 長度 > 0：呼叫 `FINDSTR(槽, 長度)`（PC-98 `overlay-07:1364h`，Borland 符號）。
4. 長度 = 0：把字串槽首位元組清成 0（空字串）。

字串槽是 **`DS:A8DAh` 起、每槽 `100h` bytes** 的陣列（`槽 << 8` 定址）。

`81h`（string-memory word）另走一條：多讀一個 byte 組成 word，再以
`ADDFNC` 併起來當 ECL 位址；**對 `7C00h` 有特例分支**（`7C00h` 正好是
bank 1 的起點），走一段以 `DS:9594h` 的 record 為基底的處理。

⚠ 尚未解出：`FINDSTR` 的實際查表方式、`80h` 在 bytecode 裡佔幾個 byte
（步進由 `FINDSTR` 內部推進，不在 `READVAR` 主迴圈），以及 `7C00h` 特例做什麼。
因此**不能宣稱 packed text 的長度規則已完成**。

## bank 1 的計算值對照

第 565 輪確認 bank 1（`7C00h..7FFFh`）讀取時先呼叫計算 routine
（PC-98 `overlay-07:08BDh`）。本輪把它機械化抽出：那是一條 `cmp ax, N` 鏈，
把 `7C00h + N` 對映到 `DS:9594h` 所指 record 的欄位偏移。

產物：[`docs/audit/ecl-bank1-field-map.md`](../audit/ecl-bank1-field-map.md)，
目前 **32 個比較中抽出 22 筆**單純的欄位讀取，例如：

| ECL 位址 | record 偏移 | 寬度 | 延伸 |
|---|---|---|---|
| `7C15h` | `+013h` | byte | 零 |
| `7CA0h` | `+0E5h` | byte | 符號 |
| `7CBBh` | `+0FBh` | word | 零 |
| `7CC9h` | `+116h` | byte | 符號 |
| `7F3Eh` | `+67Ch` | word | 零 |

**其餘 10 個分支不是單純欄位讀取**（有額外運算或多重來源），必須逐一人工
確認；產生器會把數量差如實印出來，不得假設漏掉的那幾筆不存在。

本表只證明「ECL 位址 → record 偏移 → 存取寬度與延伸方式」，**不證明欄位語意**。
`DS:9594h` 指向哪一種 record（隊伍、目前角色、或戰鬥暫存）也尚未證明——
它同時出現在 `81h` 的 `7C00h` 特例與 `STOREVALUE` 的部分分支，是下一個
值得追的中心結構。

## 對 remake 的意義

bank 1 的位址**不是可自由讀寫的 work memory**：讀取會被導向 record 欄位，
寫入才落到陣列。remake 若把 `7C00h..7FFFh` 當成單純的記憶體，讀回來的值會與
原版不同。這與第 565 輪的 bank 4 讀寫不對稱是同一類問題。

## 這份規格明確不宣稱

- `FINDSTR`／`GETSTR`／`STORESTRING`／`CHECKSTRING` 的實作。
- 字串槽的總數與生命週期（何時重置、跨 block 是否保留）。
- `DS:9594h` record 的身分與版面。
- 漏抽的 10 個 bank 1 分支。
