# 第 431 輪：PC-98 held effects 清除 Action 但不消耗法術格

狀態：`READY`（限 effects `1Fh／33h／34h／35h` 的 Action consumer）

## 問題與結論

第 429／430 輪分別證明正傷害與毒雲術 effect `44h` 會取消 pending spell，
並呼叫 memorized-slot consumer。不能因此推論所有「無法行動」狀態也會消耗
法術格。

PC-98 effect table 現在證明 `1Fh／33h／34h／35h` 共用同一 handler。該
handler 呼叫完整 `CLEARACTION`，清除 pending spell、delay、guard 與另一個
尚未命名的 Action byte，但整條函式沒有呼叫第 429 輪的 memorized-slot
consumer `014A:0070h`。因此四種 held effects 的契約是：取消待結算動作，
不消耗 memorized spell slot，也不顯示正傷害／毒雲術的施法中斷訊息。

## 非破壞性輸入

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | typed resident stubs | `exact` |
| `PC98-GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | overlay corpus | `exact` |
| overlay 12 | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` | effect table 與共用 handler | `exact` |
| overlay 24 | `0b5d1bebb367414fac93287449a230a36355c075efdc30640a4994bc11d5cd5f` | `CLEARACTION` consumer | `exact` |
| `pc98_held_action_clear_audit.idc` | `5e3463596134b50053f0906a9dc21f6ade969d9ce150472f7e7de3c5393aedf6` | 連續 bytes／指令 ledger | 工具 |

原始 executable／overlays 唯讀掛載，IDA database 與 29＋21 行 ledger 只建立
在 `/tmp/coab-ida-431`。IDC 不 rename、不 patch；每筆輸出保留原始
overlay-local offset、bytes 與 disassembly，並以非空檔案檢查成功。

## Exact table → handler → consumer 鏈

effect table base 是 `DS:A044h`，每項四 bytes，effect ID 使用 one-based
索引。`INITEFFPROX` 的直接 writes：

| raw effect | table slot | resident pointer |
|---|---|---|
| `1Fh` | `A0BCh／A0BEh` | `008B:0039h` |
| `33h` | `A10Ch／A10Eh` | `008B:0039h` |
| `34h` | `A110h／A112h` | `008B:0039h` |
| `35h` | `A114h／A116h` | `008B:0039h` |

typed TPOV resolver 將 `overlay 12 resident stub 0039h` 解析為 entry 5、
handler local `0075h`。local `0078h..0082h` 把 Player far pointer 傳給
`014A:00CAh`；後面的 `or al,al／jz $+2` 沒有額外分支副作用。

`014A:00CAh` 經 typed resolver 落到 overlay 24 entry 34、local
`2A5Bh`。完整函式 `2A5Bh..2A9Eh`：

- 由 `Player+18Eh` 取得 Action；
- `2A69h` 清 `Action+03h`，第 421／425 輪已證明是 delay；
- `2A76h` 清 `Action+00h`，第 425／429 輪已證明是 pending spell ID；
- `2A82h` 清 `Action+07h`，第 421 輪已證明是 guarding；
- `2A8Fh` 清 `Action+06h`，完整語意仍是 `unknown`；
- 回傳 1，函式內沒有 far call，也沒有抵達 `014A:0070h`。

因此「四種 effect → 同一 handler → 完整 Action clear」及「不呼叫
memorized-slot consumer」均為原始 bytes 與 typed control flow 的 `exact`。
effect 名稱沿用既有 reference `IsHeld` mapping：`1Fh` helpless、`33h`
snake charm、`34h` paralyze、`35h` sleep；本輪沒有以名稱反推控制流。

## Remake contract

- engine 既有作品中立 `combat/action.State.Clear` 已符合完整 typed Action
  clear，本輪不需修改共用 engine。它一併清除 stable target ID／point 是
  remake transaction invariant；不能把這項 typed projection 反寫成 raw
  `Action+06h` 已證明是目標欄位。
- CoAB scheduler 遇到 operational held fighter 時，先呼叫 Battle
  `ClearAction`，同步清除 UI-owned cast／move／view，再顯示資料化「無法行動」
  訊息並跳過回合。
- 這條路徑不得建立 `SpellInterruption` event；State 因此不會移除 roster 的
  memorized slot，也不會誤用 `combat_spell_interrupted` 訊息。
- 四個 raw IDs 共用同一機制，不按怪物名、sprite、章節或顯示譯名特判。

## 驗收與邊界

- 正常 mutable scheduler 對 `1Fh／33h／34h／35h` 各走一次 held turn：
  pending spell／target／delay 全清，下一名玩家角色取得控制。
- 四例皆不建立 interruption event，正式 roster 法術格保持不變。
- 既有 held target 必中、held guard suppression 與 effect duration regression
  必須保持通過。
- Docker／Xvfb、`--network none`、本機 engine replace 的正式
  `./cmd/... ./gamepack ./internal/...` gate 共 31 套件以 `-count=1` 通過。

六章 shipped `MON*SPC` corpus 沒有這四種 innate effect；目前也尚未實作
Sleep／Hold Person 等動態 writer。因此本輪只宣稱 consumer 與正常 scheduler
handoff 完成，不宣稱由正式法術選單施放到 effect 套用的玩家路徑已完成。
effect 豁免、施法動畫、解除／喚醒、角色 `.FX` 寫回、原版訊息與聲音仍是
後續工作；石化與沉默也不在這四個 IDs 內，不能套用本規則。
