# 第五百六十九輪：小函式批次逐條判讀

狀態：`READY`（限本輪五種可判定形狀的判準與結果）。日期：2026-08-14

## 結論先行

`≤ 48 bytes` 的函式全部匯出**完整 body**（1,128 個），依「整個 body 都被解釋掉」
的原則逐條判讀，**257 個可完全判定並標為 `已解讀`／`exact`**：

| 形狀 | 數量 | 判準 |
|---|---:|---|
| overlay stub | 99 | 首指令是 `int 3Fh`（Borland overlay manager 中斷）＋ control 資料 |
| 空函式 | 99 | prologue／epilogue 之外沒有任何指令，呼叫即返回 |
| 轉呼叫 | 14 | 唯一實質動作是一個 `call`／`jmp`，參數原樣傳遞 |
| 常數字串 | 45 | 把一段 Pascal 字串常數複製進區域緩衝，再呼叫訊息 routine |

「常數字串」這一類的字串內容可以直接從同一段 overlay 解出來（首位元組是
長度），所以字串是**函式行為的一部分**，不是推測。DOS 側解出
`is Blessed`、`is Cursed`、`is fire resistant`、`is paralyzed`、`Knock-Knock`
等法術狀態訊息，PC-98 同單元是對應的日文（Shift-JIS）。清單見
[`docs/audit/spell-status-message-strings.md`](../audit/spell-status-message-strings.md)。
⚠ 兩平台**尚未證明逐筆對應**，不得依表格列序建立 en／ja 配對。

其餘 871 個維持 `待解讀`：它們有真正的控制流或資料處理，形狀不在本輪的
可判定集合內。

## 判準與兩個防呆

判定的前提是**整個 body 都被解釋掉**。扣掉 prologue／epilogue 後，剩下的
指令必須全部落在該形狀允許的集合裡；只要有一條殘留就不判定。這是為了避免
「前兩條看起來像 accessor 就當成 accessor」而漏掉後面真正做事的指令。

**防呆一：沒有 `ret` 就不判定。** IDA 的函式邊界會建錯，被切短的函式看起來
就像空函式（第 564 輪實測 `ON GOTO` handler 被切成 3 bytes）。因此要求
body 內必須真的出現 `retf`／`retn`，否則一律留 `待解讀`。加上這道防呆後，
「轉呼叫」從 33 降到 14——那 19 個是邊界建錯的假象，不是真的 thunk。

**防呆二：位址語意不在判定範圍。** 「回傳 `DS:A2A9h` 的 byte」是對函式本體的
完整描述，但 `DS:A2A9h` 代表什麼是另一條證據線。台帳條目會如實這樣寫，
不得把 accessor 的解讀當成該位址的語意已知。

抽驗（人工核對逐條反組譯，非只看分類結果）：

- `START.EXE:103B0h` — `int 3Fh`，IDA 自身也標註為 Overlay manager interrupt。
- `START.EXE:11860h` — `push bp／mov bp,sp／mov sp,bp／pop bp／retf`，確為空函式。
- `overlay-03:0000h` — 單一 `call far ptr 19Ch:2Ah` 後返回，確為轉呼叫
  （該位址是 `LOADPROTECT`，unit 初始化呼叫 overlay 載入器）。

## 過程中的環境事故

DOS `overlay-12`／`13`／`14` 的 `.i64` 在批次執行中崩壞（`Failed to initialize
IDA as library (error code 4)`）。依既有處置：先用其他 `.i64` 跑同一支腳本做
正對照確認工具正常，再刪除重建三份資料庫。重建後函式數與崩壞前一致
（全庫仍是 2,825），沒有因事故改變分母。

## 重生

```sh
tools/ida.sh py <module>.i64 export_small_functions.py /work/small/<module>.json 48
python3 scripts/small_function_reader.py            # 預覽
python3 scripts/small_function_reader.py --write    # 寫入台帳
```

## 這份規格明確不宣稱

- **99 個空函式的存在理由**。它們可能是被條件編譯掉的除錯樁、佔位的
  overlay entry，或原始碼裡真的空的程序。本輪只證明「執行時不做事」。
- **overlay stub 指向哪裡**。stub 的 control 資料尚未逐一解析回目標 entry；
  要那份對應請用 `cmd/ovr-manifest` 的 entry 表。
- **916 個未判定小函式的任何性質**。它們只是「不屬於本輪五種形狀」，
  不代表複雜或不重要。
