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
| 手上有聲音要停 | SOUNDHALT（看 `SOUNDUP`） | 沒發過 | 見下 |

`SoundStop` 的 18 處全部是引擎內務（`SOUNDUP` 守門），不對應任何玩法事件。
remake 的 `sound.Player` 是逐事件的一次性播放器，換場景時本來就不會殘留，
所以**不補**——這一格是「原作需要而 remake 的架構不需要」，不是缺漏。
理由寫在這裡，免得下一輪又把它當成待辦。

## 還沒解的

- `SHOWARROW` 的武器類別碼（`+56h` 那 14 個值）還沒對回 remake 的武器分類，
  所以第二聲還接不上；目前 remake 的投射武器只發一次 `SoundArrow`。
- `CASTSPELL` 依編號選音效這條規則已解出，但 remake 目前是依**視覺種類**選。
  兩者在 Fireball／Lightning Bolt 上結論相同，其他法術也都落到 `SoundCast`，
  所以行為一致；要不要改成照編號查是實作口味，不是正確性問題。
