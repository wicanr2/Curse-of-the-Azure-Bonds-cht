# 1187 — DOS 側的音效格子：沒有符號要怎麼認

- 證據等級：`exact`（兩版執行檔的原始位元組）
- 上游：spec 1186（音效逐處語意，PC-98 側）、spec 1182（overlay 單元名）
- 產物：`cmd/dos-sound-map`、`docs/audit/dos-sound-map.md`
- 狀態：`READY`

## 問題

spec 1186 把 54 處 `SOUNDFX` 逐處定位到所在常式與觸發條件——**全部來自 PC-98**，
因為只有 PC-98 帶 Borland 除錯符號。而本專案的**行為 oracle 是 DOS**。
DOS 沒有符號，所以 spec 1186 的每一條結論在 oracle 上都還沒有落點：
不知道 DOS 的音效常式在哪、不知道那 17 個描述子在 DOS 對到哪些位址。

「PC-98 解出來的東西不能直接套到 DOS」這句話一直寫在各份規格的警告裡，
但**沒有人做過那個「另外認一次」的動作**。這一份做它。

## 認法：兩個互相獨立的訊號

不靠符號，靠形狀。

### 訊號一：呼叫點分佈

PC-98 的 `SOUNDFX`（段 `0893h` 位移 `0000h`）在 overlay 側有 42 處呼叫點，
跨 8 個模組：

```
overlay-01×1、overlay-02×18、overlay-03×2、overlay-13×10、
overlay-14×3、overlay-22×4、overlay-24×2、overlay-32×2
```

拿這組分佈去比 DOS 全部 **432** 個相異 far-call 目標：

| 名次 | DOS 目標 | 差 | 呼叫點 |
|---:|---|---:|---:|
| 1 | **`0713:0020`** | **1** | 41 |
| 2 | `0713:00D6` | 29 | 13 |
| 3 | `006B:004D` | 32 | 26 |

★ **差距本身就是證據**。第一名差 1、第二名差 29——不是「最接近的那個」，
是「唯一接近的那個」。前兩名咬得很近的話，這就不是指紋而是巧合，
報表會把兩個分數都印出來讓人自己判斷。

唯一的差在 `overlay-02`：PC-98 18 處、DOS 17 處。成因見下面的 `CRASHFX`。

順帶認出第二個入口 `0713:00D6h`（overlay-02×13 ＋ 常駐 7 處，多數不推描述子），
對應 PC-98 的 `INITSOUND`。

### 訊號二：表的位置

兩版的描述子都是 `基底 ＋ 選擇子×2`，順序相同。問題是**平移一格之後每一列
還是說得通**：位址仍然連續、名字仍然一一對應。所以基底不能用猜的，也不能
「取最低的那個」——最低那一格剛好沒有 DOS 呼叫點時會整排偏掉，而且看不出來。

`cmd/dos-sound-map` 把**每一個觀察到的描述子都試一次**當錨點，
用「有幾格的 overlay 分佈與 PC-98 逐模組相同」計分：

| 錨點 | 對上的格子 |
|---|---:|
| `25AAh` | **14** |
| 第二名 | 2 |

分佈**不參與**位置的推導，位置也不參與分佈的比對——兩個訊號各自獨立，
所以「都指到同一個答案」才算證據。

## 結果：描述子逐格對照

| PC-98 | DOS | 名稱 | overlay 分佈（兩版相同） |
|---|---|---|---|
| `4838h` | `25AAh` | SOUNDHALT | overlay-02×13 |
| `483Ah` | `25ACh` | SOUNDOFF | （overlay 側沒有）|
| `483Ch` | `25AEh` | SOUNDON | overlay-03×1 |
| `483Eh` | `25B0h` | CASTFX | overlay-22×2 |
| `4840h` | `25B2h` | MISSFX | overlay-24×1 |
| `4842h` | `25B4h` | SPELLHITFX | overlay-24×1 |
| `4844h` | `25B6h` | DEADFX | overlay-03×1、overlay-32×2 |
| `4846h` | `25B8h` | WHISTLEFX | overlay-13×2 |
| `4848h` | `25BAh` | HITFX | overlay-13×2 |
| `484Ah` | `25BCh` | LIGHTNINGFX | overlay-22×1 |
| `484Ch` | `25BEh` | SWISHFX | overlay-13×3 |
| `484Eh` | `25C0h` | PADFX | overlay-02×3、overlay-13×1、overlay-14×3 |
| `4850h` | `25C2h` | FIREBALLFX | overlay-02×1、overlay-22×1 |
| `4852h` | `25C4h` | ARROWFX | overlay-13×2 |
| `4854h` | `25C6h` | OVERTUREFX | overlay-01×1 |
| `4856h` | `25C8h` | COMBATFX | （兩版都沒有呼叫點）|
| `4858h` | `25CAh` | CRASHFX | **只有 PC-98 有** |

**14 格逐模組完全相同**，1 格不同。兩支執行檔各自編譯、沒有共用符號、
位址差 `228Eh`，而每一格的呼叫點落在哪些模組、各幾次，全部一致。

## 唯一的差異：`CRASHFX`

PC-98 在 `INTERPET`（`LOADINTERPET＋35A1h`）有一處 `CRASHFX`；
**DOS 一處都沒有**。逐位元組確認過：整支 DOS（常駐 ＋ 36 個 overlay）裡
`push word [25CAh]` 出現 **0** 次。

正對照做過才讓這個 0 算數：同一種掃法在 DOS 找得到 `PADFX` 7 處、
`SOUNDOFF` 4 處。**掃描面是好的，所以 `CRASHFX` 的 0 是真的 0。**

remake 這一側在 `PROGRAM 3`（全隊陣亡）發 `SoundCrash`，而 `PROGRAM` 正是
`INTERPET` 的 opcode——與 PC-98 對得上。⚠ **DOS 版沒有這一聲**；DOS 的資源
對照表裡 `Crash` 也沒有對應的 WAV，所以 remake 在 DOS 資源下本來就是靜音。
兩邊因此沒有衝突，但這是**巧合不是設計**，記在這裡免得日後有人「補上」
DOS 的 crash 音效。

## 常駐那一半

| 名稱 | DOS 常駐 | PC-98 常駐 |
|---|---:|---:|
| SOUNDHALT | 6 | 5 |
| SOUNDOFF | 4 | 3 |
| SOUNDON | 4 | 4 |

★ 兩版的常駐都**只**放這三個，一個玩法音效都沒有。這是 spec 1186
「`SoundStop` 是引擎內務、不對應任何玩法事件」的**第二次獨立印證**：
換一支執行檔、換一組編譯結果，結論一樣。

## 這一份不宣稱什麼

- **沒有**宣稱 DOS 的音效常式內部實作與 PC-98 相同。認出來的是**呼叫介面**
  （入口位址與描述子表），不是波形產生方式。DOS 走 PC 喇叭、PC-98 走
  軟體發聲，兩者的 `Effect` 資料本來就不一樣。
- **沒有**宣稱 DOS 的描述子內容（那 20 個 word）與 PC-98 相同。這一份只對到
  「哪一格是哪個音效」。
- **沒有**逐處比對 DOS 的呼叫點落在哪一支常式。DOS 沒有符號，
  `docs/audit/pc98-music-triggers.md` 那張「哪一支常式在放」的表在 DOS
  這一側目前只到模組層級。
