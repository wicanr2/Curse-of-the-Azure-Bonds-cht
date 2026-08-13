# 第五百五十九輪：全模組反組譯全掃與函式覆蓋台帳

狀態：`READY`（限工具鏈、模組清冊與函式全集）；本輪**不宣稱**任何新的
runtime 語意。日期：2026-08-13

## 結論先行

先前的反組譯是問答式的：`scripts/ida/` 的 64 支腳本每支只回答當時那一題，
沒有任何全模組清冊，因此「還剩多少沒看」無法回答。本輪把它換成可重生的
全掃流程，得到兩個可被推翻的數字：

| 平台 | 模組 | 函式 | 指令 | 已定義程式碼 | segment 總量 | 未定義 |
|---|---:|---:|---:|---:|---:|---:|
| DOS | 37 | 1,344 | 92,495 | 238,602 | 357,156 | 16,044 |
| PC-98 | 37 | 1,481 | 103,759 | 269,027 | 389,469 | 20,319 |

全部 2,825 個函式目前一律標 `待解讀`。這是刻意的：狀態只能來自
`docs/audit/re-function-ledger.json` 的明確記錄，不由關鍵字比對產生。

**最重要的發現是先前的盲區**：raw TPOV overlay 直接載入 IDA 只會得到
**1 個函式**（DOS `GAME.OVR` 實測）。overlay 沒有 MZ header 也沒有 entry
point，自動分析無從開始，所以 260 KB（DOS）／274 KB（PC-98）的 overlay
程式碼——也就是這款遊戲的絕大部分——在本輪之前從未被系統性分析過。
既有規格對個別 overlay 位址的結論仍然成立（它們是用逐段 `create_insn` 的
一次性 IDC 取得的），但那條路徑無法產生全集。

## 輸入與雜湊

| 輸入 | SHA-256 |
|---|---|
| DOS `START.EXE` | `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5` |
| DOS `GAME.OVR` | `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215` |
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` |

IDA image `ida-pro-9.4-idapython:py312-v1`（`9.4.0.260610`）。原始檔唯讀，
所有 `.i64`／`.asm`／JSON 只存在 `workplace/re-sweep/`，不進版控。
DOS overlay-02 重新切出的 code span 雜湊
`ba41e4f437d5a86b09078f65a09197ebd6728e1d0b062edc13309174c48aa201` 與第 519 輪
保存的值相同，證明本輪的抽取與既有規格對同一段 bytes。

## 流程（可重生）

```sh
tools/re-sweep.sh dos  workplace/ida-dos-probe/START.EXE  workplace/ida-dos-probe/GAME.OVR
tools/re-sweep.sh pc98 workplace/ida406/PC98-GAME.EXE     workplace/ida406/PC98-GAME.OVR
tools/go.sh build -o workplace/re-sweep/re-ledger ./cmd/re-ledger && workplace/re-sweep/re-ledger
```

1. `cmd/ovr-manifest` 用既有 `internal/pc98ovr.Decode` 解析 TPOV 容器，輸出每段
   的 file offset、code size、relocation、entry stub 與 code SHA-256。
   DOS 36 段／871 entry／260,948 bytes；PC-98 36 段／897 entry／273,917 bytes。
2. resident executable 以 `idat -A -B` 建庫。
3. 每段 overlay 以 `idat -A -p8086 -b0` 建庫，再由
   `tools/ida/analyze_overlay.py` 強制 16-bit segment、把 entry stub 的
   handler-local offset 與 code offset 0（unit 初始化，不在 stub 表內）標成
   函式，最後 `auto_wait` 讓 near call 自然傳染。
4. `tools/ida/export_module.py` 匯出函式、chunk、xref、字串、具名資料與
   **未定義區段**。未定義位元組是誠實的殘量指標，不做線性掃描硬湊函式。
5. `cmd/re-ledger` join 三份來源產生
   [`docs/audit/coab-function-index.md`](../audit/coab-function-index.md)
   與每模組明細 `docs/audit/function-index/<平台>-<模組>.md`。

種子失敗數：DOS 107／PC-98 75（entry 指到已屬於其他函式的位址或無法建立
指令處）。這些是已知殘量，列在各模組明細，不得當成不存在。

## PC-98 的 Borland 除錯符號

PC-98 `GAME.EXE` 保留完整 Turbo Pascal 除錯表：1,725 symbols、53 modules、
1,531 types、641 members（`cmd/borland-symbols` 匯出）。DOS `START.EXE`
**沒有**這張表（`no legacy Borland 0x52FB header`）。

符號位址的 segment 等於 overlay control segment、offset 等於 overlay-local
code offset：332 個 overlay-code 符號中有 297 個直接落在 IDA 函式起點，其餘
35 個多為 code offset 0 的 unit 初始化（已補為種子）。

原始 Turbo Pascal 單元 → overlay 對應（32 個由 code offset 0 的 `LOADxxx`
符號直接證明，等級 `exact`）：

| overlay | 單元 | overlay | 單元 | overlay | 單元 |
|---|---|---|---|---|---|
| 00 | MEMORY※ | 12 | EFFPROCS | 24 | GENERIC |
| 01 | INTRO※ | 13 | COMSTUFF※ | 25 | TRAINING |
| 02 | INTERPET | 14 | MOVEMENT | 26 | MENUS |
| 03 | PROTECT | 15 | CAMP | 27 | OVERLAND |
| 04 | TEMPLE | 16 | LOADSAVE | 28 | DRAWWIN |
| 05 | POSTCOM | 17 | GEN | 29 | PORTRAIT |
| 06 | SHOP | 18 | ENDSTUFF※ | 30 | THREED |
| 07 | ECL2 | 19 | LIBRARY | 31 | LOS |
| 08 | COMBAT | 20 | CLOCK_ | 32 | TACMAP |
| 09 | COMPTACT | 21 | MONEY | 33 | SQRPAK24 |
| 10 | COMPREP | 22 | SPELLS | 34 | BUG |
| 11 | INIT | 23 | EFFECTS | 35 | SQRPAK8 |

※ 沒有 `LOADxxx` 符號，由該 overlay 內的符號名（`HEAPERRORHANDLER`、
`DOINTRO`、`INITDUDE`／`FIGMOVE`／`CALCWEAPDAMAGE`、`FINAL`）與 Borland
module 表順序推得，等級 `strong inference`。

## DOS 與 PC-98 的對應

36 段中有 **29 段的 entry_count 完全相同**，順序一致。不同的 7 段是
overlay-11 `INIT`（+1）、16 `LOADSAVE`（+14）、17 `GEN`（+4）、18 `ENDSTUFF`
（+4）、22 `SPELLS`（+1）、26 `MENUS`（+1）、32 `TACMAP`（+1），方向都是
PC-98 較多，與該版另加的日文文字、音源與 `MOVEPARTY` 功能一致。
overlay-34 `BUG` 在 PC-98 只剩 14 bytes（DOS 2,318）。

**這只建立 module 級對應假說**：可以用 PC-98 的單元名理解 DOS 同編號 overlay
的職責，但個別函式位址必須逐一證明，不得因偏移相近就把 PC-98 的符號名寫成
DOS 的事實。

## 已定位但尚未解讀：ECL opcode dispatcher

兩平台的 `INTERPET`（overlay-02）都有一個呼叫 53 個內部函式的分派函式：

| 平台 | 位址 | 大小 | opcode 來源 | 比較寬度 |
|---|---|---|---|---|
| PC-98 | overlay-02 local `373Eh` | 615 bytes | `ds:0A891h` | `cmp al, imm8` |
| DOS | overlay-02 local `3377h` | 682 bytes | `ds:75FFh` | `cmp ax, imm16`（先 `xor ah, ah`） |

兩邊都是線性 `cmp／jz／push cs／call near／jmp 收斂點` 的鏈；開頭四個 opcode
的 handler 位址在兩平台**完全相同**（`0052h`／`00E8h`／`0107h`／`011Eh`），
可作為同源建置的佐證。

`ds:0A891h`（PC-98）／`ds:75FFh`（DOS）是本次分派讀取的 opcode 位元組，等級
`exact`（raw bytes）；把它命名成「ECL 目前指令暫存」之前還需要 writer 端與
runtime trace，目前維持 `strong inference`。

完整 opcode → handler 表**本輪未完成**：dispatcher 的 53 個 call 中，只有 31
個 opcode 能由單層 `cmp` 直接對上，其餘落在巢狀條件內，必須逐分支讀完控制流
才能定案。以文字比對硬湊會產生假表——第 557 輪已有「靜態候選被當成 runtime
順序」的前例。這是下一批的第一項工作，工具（`tools/ida/dump_function.py`）
已就緒。

## 對既有文件的修正

- `AGENTS.md` 原本寫「本專案已驗證 IDC 可用，IDAPython 受主機 Python 影響」。
  該斷言已被推翻：IDAPython 在 `ida-pro-9.4-idapython:py312-v1` 可用，新腳本
  一律優先寫 IDAPython。舊 IDC 腳本仍有效，不必回頭改寫。
- Git 目錄由 `/tmp/azure-bonds-git` 移到 `workplace/azure-bonds-git`；原位置
  一次重開機就會連同未推送的 commit 消失。
- `CONTEXT.md` 由 5,073 行分冊到 `docs/context/`，只留現況與真相來源優先序。

## 尚未完成（本規格明確不宣稱）

- 任何 opcode、handler 或欄位的 runtime 語意。
- DOS 函式的命名：DOS 沒有符號表，PC-98 的名字不自動適用。
- 未定義區段（DOS 16,044／PC-98 20,319 bytes）的內容判定；它們可能是字串表、
  常數或未被種子觸及的程式碼，本輪只如實計數。
- 2,825 個函式的任何一項語意升級。台帳目前全為 `待解讀` 就是現況。
