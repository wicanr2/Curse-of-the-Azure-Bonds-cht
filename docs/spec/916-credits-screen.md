# 916 — 製作群畫面：完整版面重建，以及 `0634h` 的呼叫慣例

- 證據等級：`exact`（兩平台都逐條讀完，指令數與行數逐一對得上）
- 作法見 spec 783

## `overlay-01:001DFh`（DOS 551 條）↔ `overlay-01:00321h`（PC-98 775 條）

`retf` ＝ 無參數。由 spec 915 的片頭序列近呼叫。

**整支函式就是製作群畫面，沒有別的邏輯。**
每一行固定 16 條指令（4 個 byte 參數各 2 條、目的字串 3 條、來源字串 3 條、
建字串 1 條、印字 1 條、`push ss`／`push cs` 各 1 條）：

- DOS：34 行 × 16 ＋ 7（prologue／首呼叫／epilogue）＝ **551** ✓
- PC-98：48 行 × 16 ＋ 7 ＝ **775** ✓

指令數完全對上，所以**沒有任何一行漏讀**。

## ★ `0A54h:0634h`（PC-98 `0A65h:062Fh`）只彈掉自己的來源參數

本支給出決定性證據：

```
mov al, 2   ; push ax        ← 欄
mov al, 1   ; push ax        ← 列
mov al, 0Ah ; push ax        ← 色
mov al, 0   ; push ax        ← 第四個參數（全部呼叫點都是 0）
lea di, [bp+var_25] ; push ss ; push di    ← @目的字串
mov di, offset loc_51 ; push cs ; push di  ← @字串常數
call far ptr 0A54h:634h       ← 只彈掉 @字串常數，把 @目的字串留在堆疊上
call far ptr 542h:352h        ← 沒有再推任何東西就呼叫
```

印字呼叫**一條參數都沒有自己推**——四個 byte 與 `@目的字串`
全部是建字串之前推的。所以：

> **建字串助手的簽章是 `建字串(@目的, @來源)`，但它只彈掉 `@來源`；
> `@目的` 留在堆疊上，直接當下一個呼叫的引數。**

這解釋了整個專案裡「建字串之後緊接一個看似少參數的呼叫」的形狀，
也回頭確認 spec 915 的載入呼叫第一個參數就是檔名 `'title'`。

印字簽章因此是 **`印字(欄, 列, 色, 0, 字串)`**（`retf 0Ch`）。

## DOS 版面（40 欄）

| 列 | 欄 | 色 | 內容 |
|---|---|---|---|
| 1 | 2 | 10 | `based on the tsr novel "azure bonds"` |
| 2 | 6 / 9 / 20 / 24 | 10 / 11 / 10 / 11 | `by:` `kate novak` `and` `jeff grubb` |
| 4 | 10 | 10 | `scenario created by:` |
| 5 | 11 / 21 / 25 | 14 / 10 / 14 | `tsr, inc.` `and` `ssi` |
| 6 | 14 | 11 | `jeff grubb` |
| 7 | 11 | 11 | `george mac donald` |
| 9 | 1 / 18 | 10 / 14 | `game created by:` `ssi special projects` |
| 11 | 2 | 14 | `project leader:` |
| 12 | 2 | 11 | `george mac donald` |
| 14 | 2 | 14 | `programming:` |
| 15–17 | 2 | 11 | `scot bayless` `russ brown` `michael mancuso` |
| 19 | 2 | 14 | `development:` |
| 20–22 | 2 | 11 | `david shelley` `michael mancuso` `oran kangas` |
| 11 | 22 | 14 | `graphic arts:` |
| 12–16 | 22 | 11 | `tom wahl` `fred butts` `susan manley` `mark johnson` `cyrus lum` |
| 18 | 22 | 14 | `playtesting:` |
| 19–22 | 22 | 11 | `jim jennings` `james kucera` `rick white` `robert daly` |

色碼：**10 ＝ 說明文字、11 ＝ 人名、14 ＝ 分類標題**。
DOS 下半段排成**兩欄**（欄 2 與欄 22）。

## PC-98 版面（80 欄）

同樣的名單排成**四欄**（欄 2 / 13 / 23 / 32），列數從 22 壓到 14，
騰出下半頁放移植組：

| 列 | 欄 | 內容 |
|---|---|---|
| 16 | 6 | `GAME CONVERTED BY PONYCANYON` |
| 17 | 6 | ` ＳＰＥＣＩＡＬ　ＰＲＯＪＥＣＴＳ　ＶＥＲＳＩＯＮ１．２` |
| 19–21 | 1 | ` PONYCANYON INC.` ` Kunihiko Kagawa` ` Yoshiaki Matsumoto` |
| 19–21 | 11 | ` Group SNE` ` Hitoshi Yasuda` ` Miyuki Kiyomatsu` |
| 19–22 | 20 | ` S.R.S.` ` Seishi Yokota` ` MUSIC COMPOSED BY` ` Takeshi Yasuda` |
| 19–21 | 30 | `Marionette Inc.` `Yoshiaki Sakaguchi` `Masato Kobayashi` |

**`ＶＥＲＳＩＯＮ１．２` 是唯一一處全形字**——PC-98 版是 1.2 版。

## 四處 PC-98 對原文的改動

| DOS | PC-98 |
|---|---|
| 全小寫 | **標題全大寫、人名首字大寫** |
| `tsr, inc.` | `TSR,INC.`（逗號後**沒有空格**） |
| 列 6 `jeff grubb` ＋ 列 7 `george mac donald`（兩列兩個字串） | 列 5 `Jeff Grubb,George Mac Donald`（**一列一個字串**） |
| — | 多 14 行移植組名單 |

第三項是**版面因素造成的字串合併**，與 spec 913（文法）／spec 893（單複數）
不同因：那兩處是語言差異，這一處純粹是 80 欄放得下。

## PC-98 有兩支印字助手

| 助手 | 次數 | 用在 |
|---|---|---|
| `418h:0D17h` | 26 | 上半段（列 1–7）＋ `GAME CONVERTED BY PONYCANYON` |
| `418h:0DA5h` | 70 | 其餘全部 |

DOS 只有一支 `542h:352h`。兩支各印哪些行已完整列出，**但為什麼分兩支沒有結論**
——用 `0DA5h` 印的既有純 ASCII（`PROJECT LEADER:`、`Marionette Inc.`）
也有全形（`ＶＥＲＳＩＯＮ１．２`），所以不是單純的半形／全形分流。

## 中文化

- **人名一律不譯**（`Kate Novak`、`Jeff Grubb`、`George Mac Donald` …）。
- 要譯的只有 8 個分類標題：`based on the tsr novel "azure bonds"`、`by:`、`and`、
  `scenario created by:`、`game created by:`、`project leader:`、`programming:`、
  `development:`、`graphic arts:`、`playtesting:`。
- 欄位座標**寫死在程式碼裡**，中譯後字寬改變必須連座標一起改
  ——這 34（DOS）／48（PC-98）行每一行的欄列都在上表。
- PC-98 那 14 行移植組名單與繁中版無關，不必保留。

## 明確不宣稱

- 沒有宣稱 `1A0h:00D3h`（PC-98 `19Eh:00C4h`）做什麼（形狀上是清畫面或切文字模式）。
- 沒有宣稱印字的第四個參數（所有 34／48 個呼叫點都是常數 0）是什麼。
- 沒有宣稱 PC-98 為什麼要兩支印字助手。
