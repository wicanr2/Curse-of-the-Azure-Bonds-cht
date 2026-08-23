# DOS 側的音效格子：不靠符號，靠形狀認出來

由 `cmd/dos-sound-map` 產生，不要手改。方法與判讀見 spec 1187。

音效的語意全部來自 PC-98 的 Borland 除錯符號，而**行為 oracle 是 DOS**。DOS 沒有符號，所以要先在 DOS 這一側把同一組格子認出來，spec 1186 的結論才算在 oracle 上成立。

## 一、`SOUNDFX` 在 DOS 的哪裡

指紋是 **PC-98 `SOUNDFX` 的 overlay 呼叫點分佈**（跨 8 個模組、共 42 處）：

> overlay-01×1、overlay-02×18、overlay-03×2、overlay-13×10、overlay-14×3、overlay-22×4、overlay-24×2、overlay-32×2

拿它去比 DOS 全部 432 個相異 far-call 目標，前五名：

| 名次 | DOS 目標 | 差 | 呼叫點 | 分佈 |
|---:|---|---:|---:|---|
| 1 | `0713:0020` | 1 | 41 | overlay-01×1、overlay-02×17、overlay-03×2、overlay-13×10、overlay-14×3、overlay-22×4、overlay-24… |
| 2 | `0713:00D6` | 29 | 13 | overlay-02×13 |
| 3 | `006B:004D` | 32 | 26 | overlay-02×26 |
| 4 | `014D:0093` | 33 | 11 | overlay-13×11 |
| 5 | `018C:0061` | 34 | 20 | overlay-08×2、overlay-09×1、overlay-13×8、overlay-22×4、overlay-24×5 |

第一名 `0713:0020` 的差是 **1**，第二名是 **29**。★ **差距本身就是證據**：如果前兩名咬得很近，那就不是指紋而是巧合。

## 二、描述子表逐格對照

兩版的描述子都是 `基底 ＋ 選擇子×2`，順序相同。DOS 基底 `25AAh`、PC-98 基底 `4838h`（差 228Eh）。

★ 基底不是用猜的：**每一個觀察到的描述子都試過一次**當錨點，用「有幾格的 overlay 分佈與 PC-98 逐模組相同」計分。最佳基底 `25AAh` 對上 **14** 格，第二名只對上 **2** 格。整張表平移一格的話位址仍然連續、名字仍然一一對應，**只有分佈會整排錯開**——所以這個差距就是「沒有偏移」的證據。

`分佈相同` 那一欄是**獨立的第二個訊號**：位置對上之後，兩邊那一格的 overlay 呼叫點分佈還必須逐模組一致。整張表平移一格的話位置全部說得通，但分佈會整排錯開。

| PC-98 | DOS | 名稱 | DOS overlay 分佈 | PC-98 overlay 分佈 | 分佈相同 |
|---|---|---|---|---|---|
| `4838h` | `25AAh` | SOUNDHALT | overlay-02×13 | overlay-02×13 | ✅ |
| `483Ah` | `25ACh` | SOUNDOFF | — | — | —（兩邊都沒有呼叫點） |
| `483Ch` | `25AEh` | SOUNDON | overlay-03×1 | overlay-03×1 | ✅ |
| `483Eh` | `25B0h` | CASTFX | overlay-22×2 | overlay-22×2 | ✅ |
| `4840h` | `25B2h` | MISSFX | overlay-24×1 | overlay-24×1 | ✅ |
| `4842h` | `25B4h` | SPELLHITFX | overlay-24×1 | overlay-24×1 | ✅ |
| `4844h` | `25B6h` | DEADFX | overlay-03×1、overlay-32×2 | overlay-03×1、overlay-32×2 | ✅ |
| `4846h` | `25B8h` | WHISTLEFX | overlay-13×2 | overlay-13×2 | ✅ |
| `4848h` | `25BAh` | HITFX | overlay-13×2 | overlay-13×2 | ✅ |
| `484Ah` | `25BCh` | LIGHTNINGFX | overlay-22×1 | overlay-22×1 | ✅ |
| `484Ch` | `25BEh` | SWISHFX | overlay-13×3 | overlay-13×3 | ✅ |
| `484Eh` | `25C0h` | PADFX | overlay-02×3、overlay-13×1、overlay-14×3 | overlay-02×3、overlay-13×1、overlay-14×3 | ✅ |
| `4850h` | `25C2h` | FIREBALLFX | overlay-02×1、overlay-22×1 | overlay-02×1、overlay-22×1 | ✅ |
| `4852h` | `25C4h` | ARROWFX | overlay-13×2 | overlay-13×2 | ✅ |
| `4854h` | `25C6h` | OVERTUREFX | overlay-01×1 | overlay-01×1 | ✅ |
| `4856h` | `25C8h` | COMBATFX | — | — | —（兩邊都沒有呼叫點） |
| `4858h` | `25CAh` | CRASHFX | — | overlay-02×1 | 已判定是平台差異：PC-98 在 `INTERPET` 有一處，DOS 整支執行檔 **0 處**——而那個 0 做過正對照（同一種掃法在 DOS 找得到 `PADFX` 7 處、`SOUNDOFF` 4 處），所以是真的沒有。DOS 走 PC 喇叭、PC-98 走軟體發聲，兩版的音效編制本來就不一樣。 |

| 指標 | 數字 |
|---|---:|
| 兩版分佈**逐模組相同**的格子 | 14 |
| 對不上的格子（還沒判的）| 0 |
| 對不上、但**已判定是平台差異** | 1 |
| DOS overlay 呼叫點 | 41 |
| DOS 常駐呼叫點 | 14 |
| DOS 推不出描述子的 | 0 |

## 三、常駐那一半

| 名稱 | DOS 常駐 | PC-98 常駐 |
|---|---:|---:|
| SOUNDHALT | 6 | 5 |
| SOUNDOFF | 4 | 3 |
| SOUNDON | 4 | 4 |

★ 兩版的常駐都**只**放 `SOUNDHALT`／`SOUNDOFF`／`SOUNDON`——沒有任何一個玩法音效。這是 spec 1186「`SoundStop` 是引擎內務、不對應玩法事件」那句話的**第二次獨立印證**：換一個執行檔、換一組編譯結果，結論一樣。
