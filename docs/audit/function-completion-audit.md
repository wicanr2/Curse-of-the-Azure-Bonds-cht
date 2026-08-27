# CoAB 函式 RE 完成與無直接 caller 分類

由 `scripts/function_completion_audit.py` 從 `docs/audit/coab-function-index.json` 產生；不要手改。

## 結論

- 函式台帳：**2874** 筆。
- 真正尚待 RE 的函式（`待解讀`）：**0**。
- `callers = 0`：**923**；排除 entry、RTL、空函式與錯切邊界後，未解釋 **0**。
- Borland／Turbo Pascal RTL 或編譯器輔助：**169**；library 身分不自動等於玩家可見語意可忽略。
- IDA `邊界碎片`：**575**；它們不是可獨立宣稱『尚未 RE』的函式。

## 狀態

| 平台 | 已解讀 | 不阻塞 | 邊界碎片 | 待解讀 |
|---|---:|---:|---:|---:|
| dos | 1016 | 133 | 237 | 0 |
| pc98 | 1121 | 29 | 338 | 0 |

## `callers = 0` 互斥分類

| 類別 | 數量 |
|---|---:|
| overlay entry（entry table／manager 間接分派） | 855 |
| Borland RTL／編譯器輔助 | 63 |
| 空函式／空程序 | 2 |
| IDA 錯切邊界碎片 | 3 |
| 仍無解釋 | 0 |

直接 code xref 是下界；overlay entry、callback、relocation 與 far pointer 必須先排除。
Borland pattern 的證據與限制見 Codex 知識庫 `local/borland-turbo-pascal-runtime-patterns.md`。

## 邊界碎片分布

這一節是 IDA 資料品質台帳，不是待實作玩法清單。

| 平台／模組 | 筆數 |
|---|---:|
| dos/START.EXE | 20 |
| dos/overlay-02 | 29 |
| dos/overlay-03 | 1 |
| dos/overlay-04 | 1 |
| dos/overlay-05 | 3 |
| dos/overlay-07 | 8 |
| dos/overlay-08 | 7 |
| dos/overlay-09 | 22 |
| dos/overlay-10 | 13 |
| dos/overlay-11 | 1 |
| dos/overlay-12 | 18 |
| dos/overlay-13 | 15 |
| dos/overlay-15 | 10 |
| dos/overlay-16 | 8 |
| dos/overlay-17 | 21 |
| dos/overlay-19 | 15 |
| dos/overlay-21 | 4 |
| dos/overlay-22 | 25 |
| dos/overlay-23 | 12 |
| dos/overlay-24 | 3 |
| dos/overlay-29 | 1 |
| pc98/PC98-GAME.EXE | 25 |
| pc98/overlay-00 | 1 |
| pc98/overlay-01 | 1 |
| pc98/overlay-02 | 28 |
| pc98/overlay-04 | 10 |
| pc98/overlay-05 | 4 |
| pc98/overlay-06 | 2 |
| pc98/overlay-07 | 16 |
| pc98/overlay-08 | 12 |
| pc98/overlay-09 | 20 |
| pc98/overlay-10 | 12 |
| pc98/overlay-12 | 18 |
| pc98/overlay-13 | 27 |
| pc98/overlay-15 | 15 |
| pc98/overlay-16 | 14 |
| pc98/overlay-17 | 34 |
| pc98/overlay-18 | 4 |
| pc98/overlay-19 | 29 |
| pc98/overlay-21 | 8 |
| pc98/overlay-22 | 27 |
| pc98/overlay-23 | 12 |
| pc98/overlay-24 | 17 |
| pc98/overlay-29 | 1 |
| pc98/overlay-35 | 1 |
