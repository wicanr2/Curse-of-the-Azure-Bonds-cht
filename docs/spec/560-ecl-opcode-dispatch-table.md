# 第五百六十輪：ECL opcode → handler 全表（兩平台）

狀態：`READY`（限 opcode 與 handler 的**綁定**）；handler 本身的 runtime 語意
一律仍是 `待解讀`。日期：2026-08-13

## 結論先行

`INTERPET`（overlay-02）的 ECL 指令分派表已在兩平台完整解出，沒有殘留待人工
判讀的分支：

| 平台 | dispatcher | opcode 來源 | handler | 涵蓋 opcode | 需人工讀 |
|---|---|---|---:|---:|---:|
| DOS | overlay-02 `3377h..3621h` | `ds:75FFh`（`cmp ax`，先 `xor ah, ah`） | 52 | 64 | 0 |
| PC-98 | overlay-02 `373Eh..39A5h` | `ds:0A891h`（`cmp al`） | 52 | 64 | 0 |

兩平台的**指令集完全相同**：`00h..40h` 之中只有 `1Fh` 沒有 handler，其餘 64 個
opcode 都有；`41h` 以上一律走到函式 epilogue。共用 handler 的分組也一致
（例如 `04h..07h` 共用一個、`16h..19h` 共用一個）。涵蓋範圍與
`internal/ecl/operand.go` 的 `KnownCommands`（`00h..40h`）一致，這是本表的第一個
獨立佐證。

完整對照表：[`docs/audit/ecl-opcode-dispatch.md`](../audit/ecl-opcode-dispatch.md)。

## 方法（可重生，且不靠文字比對）

先前用「掃 `cmp al, N` 後面接哪個 `call`」的字串比對只能對上 31 個 opcode，
其餘落在巢狀條件裡——那正是會生出假表的做法。本輪改為**逐值符號執行**：

1. `tools/ida/dump_function.py` 匯出 dispatcher 的逐指令 raw bytes、反組譯與
   每條指令的 IDA code ref。
2. `scripts/ecl_dispatch_table.py` 對 opcode `00h..FFh` 每個值各跑一次：從函式
   起點開始，只認 `mov {al,ax}, ds:XXXX`（載入被追蹤值）、`cmp`＋條件跳躍、
   無條件跳躍、`push`／`nop` 等不影響結果的指令，走到第一個 `call` 就停。
3. **分支目標一律取 IDA 的 code ref**，不從助憶碼字串解析標籤名。
4. 遇到任何不在支援集合內、且可能影響結果的指令就停下並標記為「需人工讀」。
   本輪兩平台都是 0；若日後改版出現非 0，表就不完整，不得當成全表使用。
5. 走到 epilogue（`mov sp, bp`／`pop bp`／`retn`）記為「本 dispatcher 無
   handler」，與「解不出來」分開記錄。

重生：

```sh
tools/ida.sh py workplace/re-sweep/dos/overlays/overlay-02.bin.i64 \
  dump_function.py /work/dispatch.json 3377
tools/ida.sh py workplace/re-sweep/pc98/overlays/overlay-02.bin.i64 \
  dump_function.py /work/dispatch.json 373E
python3 scripts/ecl_dispatch_table.py --merge \
  dos=workplace/re-sweep/dos/overlays/dispatch.json \
  pc98=workplace/re-sweep/pc98/overlays/dispatch.json \
  > docs/audit/ecl-opcode-dispatch.md
```

## 兩平台位址關係

只有 `00h..03h` 四個 opcode 的 handler 位址在兩平台相同
（`0052h`／`00E8h`／`0107h`／`011Eh`）。之後 PC-98 的位址逐步大於 DOS，起始
差約 `16h..1Ch` 並隨模組增長——與 PC-98 版在同一份原始碼上另加內容一致。

**因此不得用位址加固定偏移在兩平台之間換算。** 對照的依據是 opcode 值：
同一個 opcode 在兩平台各自的 handler 才是同一件事的兩個實作，而這一點已由
本表建立。

## 勘誤（2026-08-14）

本規格第一版寫 44 handler／53 opcode，且宣稱 `29h`、`30h..39h`、`40h` 沒有
handler。那是錯的，原因是 `scripts/ecl_dispatch_table.py` 的正規表示式
`([0-9A-Fa-f]+)h?` 把立即數的 `h` 後綴吃掉，`cmp al, 10h` 於是被當成十進位
`10`（＝`0Ah`）。凡是全為數字的十六進位立即數都算錯，錯出來的表看起來仍然
自洽——「需人工讀 0 個」這個檢查抓不到它。

修正後後綴由 `parse_immediate()` 判斷，並以逐條反組譯手動抽驗
（`0Ah → 0306h`、`10h → 09DDh`，PC-98）。上表與
`docs/audit/ecl-opcode-dispatch.md` 均為修正後的值。

教訓寫成規則：**符號執行的自洽性檢查不能證明輸入解析是對的**。凡是把反組譯
文字轉成數值的地方，都要另外用原始 bytes 抽驗至少兩筆。

## 這份規格明確不宣稱

- **任何 handler 的語意**。表建立的是「opcode X 由這個位址的函式處理」，
  等級 `exact`（raw bytes ＋ IDA xref ＋ 可重跑的逐值符號執行）。handler 內部
  做什麼、寫哪些欄位、有哪些副作用，全部仍是 `待解讀`，已逐筆記入
  `docs/audit/re-function-ledger.json` 並顯示在函式索引的明細。
- **opcode 來源位址的角色**。`ds:75FFh`／`ds:0A891h` 是本次分派讀取的位元組，
  這一點是 `exact`；但把它命名成「目前 ECL 指令暫存」需要 writer 端與 runtime
  trace，目前維持 `strong inference`。
- **與既有 ECL 指令表的對應**。`docs/knowledge/gold-box-ecl-command-set.md` 的
  命令語意是另一條證據線；把本表的 handler 綁到那些名稱之前，必須逐一證明，
  不得因 opcode 編號相同就直接套用。
- **`1Fh` 的意義**。它在 `00h..40h` 之中是唯一沒有 handler 的 opcode，這是
  事實；它是否由別處處理、是保留值、還是根本不會出現在合法 bytecode 裡，
  尚未證明。獨立佐證是 `docs/audit/ecl-event-catalog.json`：CoAB 的 ECL1–ECL6
  corpus 從未出現 `1Fh`（corpus 共用到 46 個相異 opcode）。兩者一致，但這是
  相關性不是證明——「沒有 handler」與「corpus 沒用到」可以有共同原因。

## 對台帳的影響

- DOS `overlay-02:3377h` 與 PC-98 `overlay-02:373Eh` 由 `待解讀` 升為
  `已解讀`／`exact`（角色：ECL opcode dispatcher）。
- 104 個 handler（兩平台各 52 個）維持 `待解讀`，但都帶上「服務哪些 opcode」
  的註記，作為下一批解讀的排序依據。
