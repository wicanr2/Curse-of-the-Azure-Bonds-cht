# 音效觸發點：原版 vs remake

由 `cmd/sound-trigger-compare` 產生，不要手改。

⚠ remake 那一側數的是 `requestSound(SoundXxx)` **與** `return SoundXxx`（由挑選器回傳的也算一條路）。只數前者會讓 `SoundWhistle` 印成 0——它是 `missileImpactSound` 依武器類別挑出來的，一處直接呼叫都沒有。

⚠ **處數不等於播放次數**，兩邊都是：一處在迴圈裡可以響很多次，而同一個音效在原版可能由一支共用常式依武器種類分歧（`SHOWARROW` 就同時放箭／揮擊／哨音）。這張表回答的是「**有沒有這條路**」，不是「響幾次」。逐處的時機在 spec 1186。

⚠ 對照關係由 `internal/pc98sfx` 的選擇子表推出來（`SWISHFX` 帶著 `swish`，`swish` 對應 `SoundSwish`），不是另外抄一份。名字**判讀**仍然是人做的——`MISSFX` 到底是揮空還是法術沒中，要看呼叫端；那一題的答案在 spec 1186。

⚠ 原版處數來自**位元組直掃**（見 `pc98-music-triggers.md`），不是 far-call 對照表——表比實際少 12 處，而且會把 `LIGHTNINGFX` 印成 0。

| 原版音效 | 原版處數 | remake 事件 | remake 處數 | 落差 |
|---|---:|---|---:|---|
| ARROWFX（箭） | 2 | `SoundArrow` | 4 |  |
| CASTFX（施法） | 2 | `SoundCast` | 19 |  |
| COMBATFX（戰鬥） | 0 | `SoundCombat` | 0 |  |
| CRASHFX（撞擊） | 1 | `SoundCrash` | 1 |  |
| DEADFX（死亡） | 3 | `SoundDead` | 5 |  |
| FIREBALLFX（火球） | 2 | `SoundFireball` | 2 |  |
| HITFX（命中） | 2 | `SoundHit` | 2 |  |
| LIGHTNINGFX（閃電） | 1 | `SoundLightning` | 4 |  |
| MISSFX（揮空） | 1 | `SoundMiss` | 1 |  |
| OVERTUREFX（序曲） | 1 | `SoundOverture` | 1 |  |
| PADFX（腳步） | 7 | `SoundStep` | 4 |  |
| SOUNDHALT（停止） | 18 | `SoundStop` | 0 | **原版有、remake 從沒發過** |
| SPELLHITFX（法術命中） | 1 | `SoundSpellHit` | 30 |  |
| SWISHFX（揮擊） | 3 | `SoundSwish` | 4 |  |
| WHISTLEFX（哨音） | 2 | `SoundWhistle` | 1 |  |

## 結論

- 對照的音效種類：15 種。
- **原版有、remake 從沒發過**：1 種。
- remake 有、原版那一支沒出現在 `SOUNDFX` 的立即呼叫裡：0 種。

⚠ 第二類**先當成掃描面的問題**：上一版把 `LIGHTNINGFX` 列在這一類，而它其實在 `CASTSPELL` 裡——是 far-call 對照表看不到，不是原版沒有。現在原版那一側改成位元組直掃並且涵蓋常駐，這一類要是還有東西，要先問「是不是又有一個面沒掃到」再問「是不是 remake 多做」。

⚠ 第一類才是可以動手的：那幾個 `SoundEvent` 常數**宣告了卻從來沒有人送出**——編譯得過、測試全綠、玩起來就是少了那個聲音。
