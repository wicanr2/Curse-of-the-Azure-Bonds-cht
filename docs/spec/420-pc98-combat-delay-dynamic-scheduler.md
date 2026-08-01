# 第四百二十輪：PC-98 戰鬥延後與動態排程（READY）

狀態：`READY`（限人工角色的戰鬥 `DONE` 子選單、`DELAY` 寫入與同輪重新入列）

## 1. 結論

PC-98 版不是把頂層 `D` 直接當成「結束回合」。頂層 `D` 先開啟第二層選單；
該選單包含 Guard、Delay、Quit、Bandage、Speed、Exit。玩家在第二層再選
`D` 時，原程式把目前角色的 `Action.delay @ Player+18Eh+03h` 寫成 `01h`，
並結束這次人工操作。角色沒有離開本輪；下一次 TeamList 全掃描會先處理仍高於
一的角色，之後再讓延後者以 delay tier 1 重新取得行動。

先前 spec 419 把 `20→19` 列為一般 `DELAY` 待辦，這項假說已被本輪推翻：
`14h→13h` 是 `ALT+Q` 將全員切為 Quick 時，讓目前角色在同輪由自動控制接手
的特殊 handoff；一般 Delay 寫的是 `01h`。

## 2. 原始證據

分析輸入是本機 PC-98 `GAME.OVR` 解出的 overlay 08 暫存唯讀副本；IDA Pro
9.4 在 Docker 內執行，稽核腳本是
`scripts/ida/pc98_initiative_scheduler_audit.idc`。腳本只重建暫存 IDB／輸出
disassembly，不改原始映像。

### 2.1 頂層 D 是子選單入口

overlay 08 local `05A1h–05B4h`：

```text
05A1  3C44              cmp al,44h        ; 'D'
05A3  7511              jnz 05B6h
05A5  FF7608 FF7606     push current Player far pointer
05AB  8D7EFE 16 57      push address of turn-ended result
05B0  0E E8740A         call near 1028h
```

因此 `D` 本身不是直接清除 action，而是呼叫 local `1028h` 的第二層處理器。

### 2.2 第二層 Delay exact writer

overlay 08 local `116Fh–1187h`：

```text
116F  3C44              cmp al,44h        ; 'D'
1171  7516              jnz 1189h
1173  C47E0A            les di,[bp+0Ah]   ; Player
1176  26C4BD8E01        les di,es:[di+18Eh]
117B  26C6450301        mov byte ptr es:[di+03h],01h
1180  C47E06
1183  26C60501          mov byte ptr es:[di],01h ; turn ended
```

這條 writer 與 spec 419 已證實的 selector consumer 形成完整橋接：selector 每次
掃描目前 Action.delay，最大 delay 優先，故 `01h` 會保留到本輪後段，而不是
開始新一輪。

### 2.3 `14h→13h` 屬於 Quick handoff

overlay 08 local `0677h–06B8h` 的內部控制碼 `10h` 分支寫 `14h`，接著逐一
處理 TeamList；local `0310h–031Fh` 在角色再次進入 turn processor 時將
`14h` 改成 `13h`。本機 DOS Reference Card 另明列 `ALT Q` 為「所有角色切到
Quick」、空白鍵恢復 manual。次級 C 轉寫只用來協助定位，亦把該 `10h` 分支
對應為全員 Quick；主結論仍由上述 bytes、TeamList 迴圈與手冊交叉支持。

推論等級：

- `D → local 1028h`、第二層 `D → delay 01h`：`exact`。
- `01h` 由同一 selector 在同輪重選：`exact`。
- `14h→13h` 是 Quick handoff、不是一般 Delay：`strong inference`；bytes、
  手冊與獨立次級轉寫一致，尚缺 PC-98 鍵盤 runtime trace。

## 3. Remake mapping

- engine `combat/initiative.Scheduler` 保存 stable TeamList 工作副本，逐次全掃描，
  並提供 `SetDelay`／`Complete`；不預抽未來 action 的 d100。
- CoAB `Battle.BeginScheduledRound` 先把全部初始化 delay 投影回 fighter，再由
  `NextScheduledTurn` 每次選一人。這項全表投影是 Blink visibility 的必要
  consumer bridge。
- `Battle.DelayAction` 只接受目前 selection，寫 delay 1 並釋放 selection；
  State 保存該 turn 未完成，之後同輪重新出現。
- Ebiten 的頂層 `D` 開啟結束子選單；子選單 `D` 是延後，`Q` 是目前 remake
  已有的 no-attack 結束回合。繁中文字串由 locale stable ID 解析。

## 4. 驗證

- engine 測試驗證 `7 → 1`、另一個 delay 4 行動、延後者同輪再次被選，以及
  每次 selection／最終零表掃描的 d100 draw 數。
- Battle 測試驗證 delay 20 的角色延後後，delay 10 角色先行，前者再以 1 回來。
- State 正常 `StartCombat` 玩家路徑驗證 hero → Delay → ally → Done → hero，
  並確認重新出現時 fighter Action.delay 為 1。
- Standing Stone → Burial Glen 真實資料玩家路徑覆蓋 Phase Spider Blink；
  全表初始 delay 必須先投影，否則尚未行動的蜘蛛會被誤判成 delay 0 隱形。

## 5. 未完成

- `ALT+Q`／空白鍵 Quick 切換尚未接入 remake UI，也尚缺 PC-98 runtime 鍵盤 trace。
- Guard、Bandage、Speed 的完整原版行為與子選單可用條件尚未實作。
- DOS 版是否逐 byte 等價及底層原版 PRNG 仍待關閉。
- `area.field_596` surprise writer 仍未知。
