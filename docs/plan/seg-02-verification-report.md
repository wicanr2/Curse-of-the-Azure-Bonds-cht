# 驗證報告：`SEG-02` block ↔ 地圖 ↔ 區域對照

- 日期：2026-08-20
- 範圍：主線分段驗證計畫的 `SEG-02`
- 結論：**大致通過**。25 個 block 的地圖歸屬都由資料算得出來，段落清單已產生；
  剩三張地圖沒有 `script_block` 欄位、一組宣告與原作資料不一致，逐項列在第五節。

## 一、對照關係是什麼

`LOAD FILES`（`21h`）的**第一個運算元就是那一段要載的地圖區塊**，
第二個固定是 `02`，第三個固定是 `FF`。世界地圖 hub 與開場用 `7F/7F/7F`
——不載 3D 地圖。

game pack 那一側：**`area_id` 就是 ECL 成員編號、`script_block` 就是 block 編號**。
兩邊 join 得起來，段落清單見
[`docs/audit/ecl-block-graph.md`](../audit/ecl-block-graph.md)。

## 二、block 與地圖區塊的關係

多數 block 載**與自己同號**的地圖區塊。四個例外與四個「不載」：

| block | `LOAD FILES` | 意思 |
|---|---|---|
| `ECL2/0x02` | `01` | 與 `0x01` 共用提爾佛頓城區 |
| `ECL4/0x22` | `21` | ⚠ 見第五節 |
| `ECL5/0x31` | `32` | 與 `0x32` 共用地圖（pack 的 `geometry_block 50` 也是這樣宣告的）|
| `ECL3/0x12`、`ECL4/0x21`、`ECL4/0x23`、`ECL5/0x30` | 沒有 | **沿用上一段載好的檔案** |

★ `ECL3/0x12` 沒有 `LOAD FILES`、而且只與 `0x11` 互相往返——
game pack 早就把它宣告成 `original.geo3.block-11-level-2`（**同一張地圖的第二層**）。
兩邊互相印證。

## 三、25 個 block 的歸屬

| 章 | block | 段 |
|---|---|---|
| ECL1 | `0x50`、`0x51` | 世界地圖 hub（不載 3D 地圖）|
| ECL1 | `0x52` | 開場 |
| ECL2 | `0x01`、`0x02` | 提爾佛頓城區 |
| ECL2 | `0x03`、`0x04` | 下水道、火刀據點 |
| ECL3 | `0x10`、`0x11`、`0x12`、`0x15` | 佔位名（`original.geo3.*`）|
| ECL4 | `0x20`、`0x21`、`0x22`、`0x23`、`0x25` | 散提爾堡內城、暗黑神殿、眼魔洞穴 ＋ 兩個未命名 |
| ECL5 | `0x30`、`0x31`、`0x32`、`0x33`、`0x35` | 佔位名（`original.geo5.*`）＋ 一個未命名 |
| ECL6 | `0x40`、`0x42`、`0x43` | 密斯卓諾：墓園、外城遺跡、內城遺跡 |
| ECL6 | `0x45` | 佔位名（`original.geo6.block-45`）|

★ **開場 `0x52` 在 `NEWECL` 圖裡是孤立的，那不是漏算**：它由引擎選中
（`DS:4FBCh ≠ 0` 時主迴圈直接載它，spec 1141），不是被 `NEWECL` 指過去的。
回歸測試特別把它排除並寫明理由。

## 四、產出

- `cmd/ecl-block-graph` 現在同時輸出轉移圖、`LOAD FILES`、`InDungeon`
  與**段落清單**（進入自／離開到／game pack 地圖）。
- 回歸測試新增三條：沒有孤立 block（開場除外）、hub 的 `LOAD FILES` 是 `7F/7F/7F`、
  三個代表性地城的 `LOAD FILES` 值。

## 五、還沒收乾淨的

1. **三張地圖沒有 `script_block`**：`tilverton.first-person`、
   `zhentil-keep.inner-city`、`zhentil-keep.dark-shrine`。它們的 `geometry_block`
   分別是 `01`／`20`／`21`，與 `LOAD FILES` 對得上，但**沒有直接證據說哪一個
   block 是它的 script**（`0x01` 與 `0x02` 都載 `01`），所以這一輪不填。
2. ⚠ **`zhentil-keep.beholder-cave` 的宣告與原作資料不一致**：pack 宣告
   `script_block 0x22` ＋ `geometry_block 0x25`，但 `ECL4/0x22` 的 `LOAD FILES`
   是 `21`，而 `ECL4/0x25` 才載 `25`。兩者只能有一個對。要回到原始 bytes 或
   實機路徑判。
3. **九個佔位名還沒定**（`original.geo3.*` 四個、`original.geo5.*` 四個、
   `original.geo6.block-45`）。命名要靠那些 block 的劇情文字，屬於下一輪。
   已知線索：`gamepack` 既有測試把 `original.geo5.block-31` 稱作 Hap。
