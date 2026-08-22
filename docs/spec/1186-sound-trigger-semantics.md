# 1186 — 音效什麼時候響：原作 54 處 `SOUNDFX` 的逐處語意

- 證據等級：`exact`（PC-98 執行檔與 overlay 的原始位元組 ＋ Borland 除錯符號）
- 上游：spec 1182（overlay 單元名）、`docs/audit/pc98-music-triggers.md`
- 產物：`cmd/pc98-music-triggers`（逐處表）、`cmd/sound-trigger-compare`（對照）
- 狀態：`READY`

## 問題

`docs/audit/pc98-music-triggers.md` 已經解出「每個音效有幾處呼叫點」，但**處數
回答不了「什麼時候響」**，而 remake 要接的是時機。這一份把 54 處逐一定位到
所在常式與觸發條件。

## 先修兩個把答案藏起來的洞

### 洞一：描述子對照表漏了兩格

`SOUNDFX` 的引數是**音效描述子變數的位址**（`push word [位址]`）。那些變數在
PC-98 資料段 3113 連續排列：

```
4838h  SOUNDHALT      ← 不是選擇子，是「停掉手上那一個」
483Ah  選擇子 0（SOUNDOFF）
483Ch  選擇子 1（SOUNDON）
…
4858h  選擇子 15（CRASHFX）
```

工具原本用一張**手打的位址→名字表**，而那張表漏了 `4840h` 與 `4842h`。漏掉
沒有任何徵兆：報表把那兩處印成「符號表沒有」，讀起來像原作真的少了兩個名字。
`cmd/sound-trigger-compare` 於是也少了兩列——正好是 remake 用得最兇的
`SoundSpellHit`（29 處）與 `SoundMiss`（2 處）。

改成 `描述子 ＝ 483Ah ＋ 選擇子×2` 之後，「漏一格」在結構上不可能發生。
`4840h` ＝ **MISSFX**、`4842h` ＝ **SPELLHITFX**。

### 洞二：far-call 表比實際少 12 處

far-call 對照表只收得到 IDA 認成程式碼的呼叫點。改用**位元組直掃**
（`9A 00 00 93 08` ＝ `call far 0893:0000`）：

| 來源 | far-call 表 | 位元組直掃 |
|---|---:|---:|
| overlay | 36 | **42** |
| 常駐 `PC98-GAME.EXE` | 0（表不涵蓋常駐） | **12** |
| 合計 | 36 | **54** |

表裡那 36 處是直掃 42 處的**真子集**（逐處位址比對過，沒有一處只在表裡）。
多出來的 6 處是 `overlay-02×2`（SOUNDHALT）、`overlay-13×1`（HITFX）、
`overlay-22×3`（FIREBALLFX／LIGHTNINGFX／CASTFX）。

⚠ **這 6 處裡有一處會改結論**：`LIGHTNINGFX` 在 far-call 表裡是 **0 處**，
於是對照報表寫著「remake 有、原版沒有」。實際上原作有，在 `CASTSPELL` 裡。
**假零的來源是掃描面，不是原作。**

正對照：同一個位元組樣式在常駐裡掃得到 12 處 `SOUNDFX`、3 處 `INITSOUND`、
1 處 `BGMPLAY`——樣式抓得到東西，所以某個音效的 0 才算數。

## 逐處語意

### 戰鬥：近戰（COMSTUFF ＝ overlay-13）

| 位移 | 音效 | 條件 |
|---|---|---|
| `1863h`、`15C0h` | **HITFX** | `call far 014A:003E` 回非零（命中判定過）之後 |
| `193Bh` | **SWISHFX** | 揮擊，緊接著呼叫 `DESCRIBEWEAPONATTACK` |

`193Bh` 那一段的形狀：

```
cmp byte [bp-15h], 2 ; jnz …
cmp byte [bp-11h], 0 ; jnz …
push word [484Ch]    ; SWISHFX
call far 0893:0000   ; SOUNDFX
push ×4 / push …
call DESCRIBEWEAPONATTACK
```

**揮擊聲在描述文字之前放，而且不看中不中**。命中另外補一聲 HITFX
（`1863h` 那一段在命中判定通過之後才走到）。

⚠ **原作的近戰揮空沒有專屬音效**。`MISSFX` 整支執行檔只有一處呼叫，在
`GENERIC` 的 `TWINKLE` 裡，見下。

### 戰鬥：投射武器（`SHOWARROW`，唯一呼叫者是 `ATTACKE`）

進場**無條件**放 `ARROWFX`（`SHOWARROW＋Ah`），之後依 `es:[di+56h]`
（武器類別）分歧，在飛行動畫尾端再放一聲：

| 武器類別 | 第二聲 | 位移 |
|---|---|---|
| `09h`、`15h`、`64h`、`1Ch`、`1Fh`、`49h` | ARROWFX | `2B4Ah` |
| `02h`、`07h`、`14h` | SWISHFX | `2B81h` |
| `55h`、`56h` | WHISTLEFX | `2BB4h` |
| `65h`、`2Fh`、`62h` | WHISTLEFX | `2C01h` |
| 其餘 | SWISHFX | `2C48h` |

分歧鏈是六個 `cmp al,imm / jz` 串起來的，六個 `jz` 全部落在同一個目標
（`2AC4h`），落空的 `jmp` 才進第二層。

### 戰鬥：法術

**放法術的當下**（`CASTSPELL＋2AAh`，依法術編號）：

```
mov al, [bp+0Eh]
cmp al, 2Fh ; jnz →   push [4850] FIREBALLFX
cmp al, 33h ; jnz →   push [484A] LIGHTNINGFX
                      push [483Eh] CASTFX     ← 其餘全部
```

`2Fh` ＝ 47 ＝ **Fireball**、`33h` ＝ 51 ＝ **Lightning Bolt**，與
`docs/audit/spell-damage-table.md`／`spell-visual-table.md` 早先由另一條路
解出的編號一致——兩次獨立推導指到同一組編號。

**命中結算**（`TWINKLE＋7Fh`／`＋8Ah`，在 `GENERIC`）：

```
cmp byte [bp+0Ah], 0
jz  →  push [4840h] MISSFX      ; 沒中
       push [4842h] SPELLHITFX  ; 中了
```

⚠ **`MISSFX` 是法術沒中的聲音，不是近戰揮空的聲音**。整支執行檔只有這一處
呼叫它，而它與 `SPELLHITFX` 共用同一個 `if`。

### 死亡

`SUBTRACTDUDE`（TACMAP）`＋11h`／`＋AEh` 放 **DEADFX** ——戰鬥員從戰場移除時。
另有一處在 `DOPROTECT＋294h`（PROTECT）。

### 移動

**PADFX** 7 處：`REALMOVE＋1CFh`（COMSTUFF，戰鬥中走一格）、
`PREMOVEPARTY＋1EBh`／`＋236h` 與 MOVEMENT 內一支無名靜態（overlay-14）、
INTERPET 3 處。

### 停音

**SOUNDHALT** 18 處（overlay-02 ＝ INTERPET 13 處、常駐 5 處），
每一處都是同一個守門：

```
cmp byte [8B5Ah], 0   ; SOUNDUP
jz  +9
push word [4838h]     ; SOUNDHALT
call far 0893:0000
```

`8B5Ah` 由 Borland 符號表讀出是 **`SOUNDUP`**（同一段還有 `8B58h` ＝
`SOUNDTYPE`、`8B59h` ＝ `OLDSOUND`）。語意是「手上還有聲音在響就先停掉」，
而不是某個玩法事件。常駐另外有 SOUNDOFF×3、SOUNDON×4。

## remake 這一側該怎麼接

| 時機 | 原作 | remake 現況 | 動作 |
|---|---|---|---|
| 近戰揮擊 | SWISHFX（每次） | 沒發過 | **補**：`VisualMelee` 的 travel 階段發 `SoundSwish` |
| 近戰命中 | HITFX | `SoundHit` | 不動 |
| 近戰揮空 | **沒有音效** | `SoundMiss` | **移除**：那是法術沒中的聲音 |
| 法術沒中 | MISSFX | 一律 `SoundSpellHit` | **改**：沒中發 `SoundMiss` |
| 法術命中 | SPELLHITFX | `SoundSpellHit` | 不動 |
| 放 Fireball | FIREBALLFX | 靠視覺種類推 | 保留，另記編號規則 |
| 放 Lightning Bolt | LIGHTNINGFX | 靠視覺種類推 | 保留 |
| 其餘放法術 | CASTFX | `SoundCast` | 不動 |
| 戰鬥員離場 | DEADFX | `SoundDead` | 不動 |
| 走一格 | PADFX | `SoundStep` | 不動 |
| 投射武器飛行尾端 | 依武器類別（箭／哨音／揮擊）| 沒發過 | **補**：`missileImpactSound`，見下 |
| 手上有聲音要停 | SOUNDHALT（看 `SOUNDUP`） | 沒發過 | 見下 |

`SoundStop` 的 18 處全部是引擎內務（`SOUNDUP` 守門），不對應任何玩法事件。
remake 的 `sound.Player` 是逐事件的一次性播放器，換場景時本來就不會殘留，
所以**不補**——這一格是「原作需要而 remake 的架構不需要」，不是缺漏。
理由寫在這裡，免得下一輪又把它當成待辦。

## `+56h` 是什麼：`SHOWARROW` 的武器類別碼

`+56h` 是 **`CHARITEMREC.ITEMPTR`**——物品類別，remake 的 `ItemRecord.Type`。
兩份獨立證據：

1. `docs/audit/pc98-record-layouts.md` 的 `CHARITEMREC`（103 bytes）在 `+056h`
   列著 `ITEMPTR`；同一欄位在檔案版的 `CHARITEMFILREC`（63 bytes）是 `+02Eh`，
   而 `+02Eh` 正是 remake 讀 `Type` 的位移。
2. `USINGMISSLEWEAPON` 走的是雙重間接：角色 `+152h` 取出 `CHARITEMRECPTR`，
   再讀那一份的 `+56h`。與 `SHOWARROW` 讀的是同一個欄位。

類別碼對回武器名稱的表由 `cmd/missile-sound-classes` 產生
（`docs/audit/missile-sound-classes.md`）。分歧鏈點名 14 個類別，其中
**12 個遊戲裡真的有物品**，2 個一件都沒有（`55h`、`62h` ⇒ 那兩條分支走不到）。

### 為什麼 remake 不必知道「架著的是哪一件彈藥」

原作分歧的是**飛出去那一件**的類別，而那不一定是武器本身：

| 武器種類 | 飛出去的 | 判斷式 |
|---|---|---|
| 弓、弩 | 另外的彈藥（箭 `49h`／弩矢 `1Ch`）| `+0Eh` bit 3 且非 bit 4 且 ≠ `0Ah` |
| 投石索類 | 自給自足（`+0Eh` ＝ `0Ah`）| 武器自己 |
| 投擲武器 | 自己（`+0Eh` bit 4）| 武器自己 |

★ **關鍵**：`49h` 與 `1Ch` **都落在同一個 ARROWFX 分支**。所以「要另外彈藥」
的武器一律得到 ARROWFX，**不必知道架著的是哪一件**。remake 因此只需要武器類別
（`Fighter.WeaponItemType`）加上 `UsesSeparateAmmunition()` 就能重現整張表。

結果（`internal/game.missileImpactSound`，逐條有測試，含預設分支的負對照）：

| 武器 | 類別 | 第二聲 |
|---|---|---|
| 長弓／短弓／複合弓／輕弩 | 要另外彈藥 | `SoundArrow` |
| 投石索、小筏投石索、油瓶 | `2Fh`／`65h`／`56h` | `SoundWhistle` |
| 飛鏢、標槍、矛、碗之飛鏢 | `09h`／`15h`／`1Fh`／`64h` | `SoundArrow` |
| 擲斧、棍、鎚 | `02h`／`07h`／`14h` | `SoundSwish` |
| 其餘 | — | `SoundSwish`（`2C48h` 的預設分支）|

⚠ `ATTACKE` 對 `SHOWARROW` 有**兩處**呼叫：`1B47h` 推它自己收到的一個
`CHARITEMRECPTR` 參數（先檢查非 NIL），`1B82h` **只在架著的武器類別 ＝ `65h`
時**成立，推的是武器本身。第二處是逐位元組讀出來的；第一處推的到底是彈藥還是
武器**沒有直接證據**，是從「分歧鏈同時點名了純彈藥類別（`49h`／`1Ch`，槽 10、
射程 0、不可能被當武器架著）與槽 0 的投擲武器」反推的。上表的結論不依賴這個
推論——兩種讀法都得到同一張表。

## 還沒解的

- `CASTSPELL` 依編號選音效這條規則已解出，但 remake 目前是依**視覺種類**選。
  兩者在 Fireball／Lightning Bolt 上結論相同，其他法術也都落到 `SoundCast`，
  所以行為一致；要不要改成照編號查是實作口味，不是正確性問題。

## 順手撞到的一個**與音效無關**的缺陷

查武器類別時對照過類別表，發現 `internal/party/character.go` 的
`AmmunitionCount` 挑的是「裝備中、類別表槽 ＝ 11 或 12」的物品，而這一款遊戲裡
**槽 11／12 是卷軸**（`3Dh` 法師卷軸、`3Eh` 牧師卷軸）；箭與弩矢是**槽 10**
（`49h` Arrow、`1Ch` Quarrel，槽 10 還混著藥水、飾品、`3Ch` Scroll）。

來源是 spec 1120 **自己的一處自相矛盾**：

- 該份 §159–161 由 `+17Dh`／`+181h` 推出「裝備槽區塊第 11、12 格」，
  再推成「類別表 `+0` 是 11 或 12」。
- 同一份 §199 卻寫著「第一件裝備中的物品常常是弓或**彈藥（slot 10）**」。

後果：`capByAmmunition` 的 `count <= 0` 是「不設限」，所以
**箭的數量從來不會限制射擊次數**；反過來，架著卷軸的角色會被卷軸張數限制遠程
攻擊次數。兩者都不會噴錯，也不會讓測試變紅。

⚠ **這一輪不修**：正確的判斷式要先解出角色記錄裡那兩個彈藥指標（`+17Dh`／
`+181h`）到底由什麼決定——槽 10 是個混裝格，光靠槽號認不出彈藥。
先把事實記在這裡，免得下次又從 spec 1120 那句錯的結論出發。
