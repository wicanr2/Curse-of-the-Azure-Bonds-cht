# 1123 — 效果的修正表：`CHECKFX` 給清單，handler 給數字

- 證據等級：`exact`（141 個 handler 的本體由 `scripts/effect_modifier_table.py`
  線性反組譯後機械分類；timing 清單取自既有的
  [`checkfx-timing-table.md`](../audit/checkfx-timing-table.md)）
- 對應工作項：`ENG-09` 的「效果碼語意」半邊
- 上游 spec 577（`ATTEMPTTOHIT` 與效果鏈）、spec 1005（效果分派表）、
  spec 581／582（`PUTDAMAGE`／`MAKESAVE` 的呼叫點）

## 原作沒有「一個效果一段 if」

規則各處問的都是同一句話：`CHECKFX(timing, 對象)`。它是一張表——每個 timing 對
應一組效果碼，逐一交給效果鏈遍歷去看對象身上有沒有，有就 `CALLEFFECT` 分派。

handler 本體則是一批**極短**的函式，把修正寫進一組固定的全域：

```asm
; 效果 01h（祝福）
add   BYTE PTR ds:6FA2h, 5      ; 士氣 ＋5
inc   BYTE PTR ds:6F9Fh         ; 共用修正 ＋1
retf  0Ah

; 效果 2Ah（緩速）
mov   al, ds:6F96h
mov   cx, 2
idiv  cx
mov   ds:6F96h, al              ; 移動率折半
```

呼叫端先把基準值寫進自己那個全域，跑一次 `CHECKFX`，再讀回來。所以
**「效果碼 → 修正」是資料**，remake 這一側不需要一個效果一段程式碼。

## 暫存全域

| 位址 | 名字 | 佐證 |
|---|---|---|
| `DS:6F9Fh` | `modifier` | 共用修正；**語意由讀它的 timing 決定** |
| `DS:6F92h` | `saving_throw` | 豁免總和（`overlay-23 entry#8` 寫它，spec 1119）|
| `DS:6FA2h` | `morale` | 士氣（`overlay-09:01388h` 寫它，spec 1122）|
| `DS:6F96h` | `movement` | 移動率（`overlay-13 entry#2` 寫它，spec 1122）|
| `DS:6F9Bh` | `attack_forced_miss` | 攻擊必失手旗標 |
| `DS:6F94h`／`6F95h`／`6F9Ch` | — | 未定 |

★ `6F9Fh` **不叫「命中修正」**。同一支 handler 無條件寫它，而它在 `0Ah`／`10h`
是命中、在 `0Bh` 是護甲、在 `0Dh` 是死亡後。取名叫命中會在護甲那條路上變成錯的
斷言，所以名字保持中性。

## 分類的四級，以及為什麼要有 `unread`

| 狀態 | 意思 | 數量 |
|---|---|---:|
| `decoded` | 整支都是對全域的加減／設定 | 6 |
| `partial` | 有數字，但還有沒解析的指令（動角色記錄、呼叫別支）| 20 |
| `inert` | 只有序幕與 `retf`，什麼都不做 | 12 |
| `unread` | 有內容但沒有可辨識的無條件加減 | 103 |

`CheckFX` 只套 `decoded` 與 `partial` 的數字，並把其餘的列進 `Unread`。
**回傳 0 有兩種意思**——沒有效果介入，或介入的 handler 還沒解讀——分不開就會
讓覆蓋報告把「還沒讀」算成「沒有影響」。

## ★ 條件分支裡的指令不算

分類器先算出「哪些指令位在條件跳躍與它的目標之間」，那些**一律不收**。
兩個例外是認得出來的慣用法：

- **夾底減法**：`cmp ds:G,K / jae L1 / mov ds:G,0 / jmp L2 / L1: sub ds:G,K`
  ⇒ 一個 `sub_clamped` 動作。拆成兩個獨立動作的話，詛咒會把士氣直接歸零而不是減 5。
- **除以常數**：`mov al,ds:G / … / mov cx,K / idiv cx / mov ds:G,al` ⇒ `div`。

這條規則讓防護邪惡（`08h`）從「豁免 ＋2、修正 −2」退回 `unread`：原作那兩個數字
**在比過陣營之後才生效**。照字面收下來會得到「看起來完整但對所有目標都生效」的
錯規則——寧可標成不知道。

## 已知的 timing

| timing | 呼叫點 | 出處 |
|---|---|---|
| `06h`／`14h` | `PUTDAMAGE` 進入時／無豁免時 | spec 581 |
| `09h` | `PUTEFFECT` | 同上 |
| `0Ah`／`10h` | `ATTEMPTTOHIT` 對目標／對攻擊者 | spec 577 |
| `0Ch` | `MAKESAVE` | spec 582 |
| `0Dh` | `KILLDUDE`／死亡後 | spec 579 |
| `11h` | 士氣檢定 | spec 1122 |
| `12h` | 移動率換算 | spec 1122 |
| `15h` | 士氣崩潰的四段表（效果 `23h`）| spec 831 |

## remake 這一側

- `gamepack/rules/effect-modifiers.json`（由腳本產生）＋ `gamepack.EffectModifiers()`。
- `combat.CheckFX(戰鬥員, timing, 基準)` 與 `CheckFXValue`。
- 已接的三處：豁免（`RollSavingThrow`）、士氣（`CheckMorale`）、
  移動（`MonsterApproach`）。
- `combat.AffectKindIsInterpreted` 把「表裡有數字」也算進判讀範圍，
  所以法術覆蓋台帳的分母跟著動。

## 明確不宣稱

- 沒有宣稱 `6F94h`／`6F95h`／`6F9Ch` 是什麼。
- 沒有宣稱 `unread` 那 103 個 handler 各自做什麼；它們多半是傷害、狀態、區域
  那一類，不是加減型。
- 沒有宣稱 `partial` 那 20 個「沒解析的部分」不重要——致盲除了三個全域之外還動
  角色記錄的 `+19Ah`／`+19Bh`，那一段沒有進表。
- 沒有宣稱命中那條路上 `modifier` 的正負號怎麼套進公式；本輪只接豁免、士氣、
  移動三處，那三處的方向由各自的呼叫端證實。
- 沒有宣稱 PC-98 側的 handler 逐條相同（那一側的全域是 `A0xxh` 一組）。
