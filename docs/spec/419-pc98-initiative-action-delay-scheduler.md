# 第四百一十九輪：PC-98 先攻、action delay 與回合選人（READY）

## 範圍與結論

本輪補完 spec 418 明列未完成的 initiative 數值與 tie order。PC-98 原版不是
`d20 + (DEX-10)/2` 後一次排序；它把先攻存在每名 combatant 的 `Action +3`
delay，並反覆掃描 linked TeamList 選出下一名：

1. 戰鬥中每名角色先擲 `1d6 + DEX reaction adjustment`。
2. 小於 1 先夾到 1；若該 combat team 被 `area.field_596` 標記，再減 6。
3. 最終 signed delay 小於 0 或大於 20 時寫 0；非戰鬥狀態也寫 0。
4. 每次選人都依 TeamList 原順序為每個節點擲一次 `1d100`，包含 delay 0 節點。
5. delay 較大者優先；delay 相同時 d100 較大者優先；d100 也相同時後掃到者
   取代先前候選。
6. 全表最大 delay 為 0 時回傳 null，該輪沒有下一名 combatant。

一般行動完成後由各 action handler 消耗其 delay；原版 `DELAY` 命令則把 delay
設為 20，使角色重新加入同一輪的後續選人。因而 remake 不得把它簡化成一張只
包含每人一次的靜態排序表。

## 輸入與非破壞性證據

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland symbols、resident entry stubs | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | overlay 8／13／24 code 與 fixups | `exact` |

`scripts/ida/pc98_initiative_scheduler_audit.idc` 只在 Docker `/tmp` 的抽取副本上
執行，保留原始位址與 raw bytes，沒有 rename Borland symbols 或修改基準
database。以下位址一律是 overlay-local；resident stub 與 file offset 另列，
不得混成同一位址空間。

## DEX reaction adjustment

Borland symbol `DEXRABONUS 014A:1416` 經 typed resolver 關閉為：

```text
overlay 24:1416 -> entry 11 -> resident stub 014A:0057
GAME.EXE file offset 1E97h
```

函式讀 shared Player record `+17h`，回傳 signed byte：

| DEX | adjustment |
|---:|---:|
| 0–2 | -4 |
| 3–5 | `DEX-6`（-3…-1） |
| 6–15 | 0 |
| 16–18 | `DEX-15`（+1…+3） |
| 19–20 | +3 |
| 21–23 | +4 |
| 24–25 | +5 |
| 其他 | 0 |

這是原版 table／branch 語意，不是現代 D&D ability modifier。

## initiative writer

36 個 overlay 的 resident far-call 掃描只有 `overlay 13:0093` 呼叫
`014A:0057`。typed resolver 另將函式入口解析為：

```text
overlay 13:0000 -> entry 1 -> resident stub 0025h
```

關鍵資料流：

```text
0085 cmp player[+197h],0       ; 是否在戰鬥
0093 call DEXRABONUS
009A roll_dice(6,1)
00A8 add ax,dx
00AD mov Action[+3],al
00B4 cmp Action[+3],1
00BE mov Action[+3],1          ; 第一段下限
00C6 mov al,player[+198h]      ; combat_team
00CC inc ax
00D1 and ax,area[+596h]
00DD sub Action[+3],6          ; 被標記的 team
00E5 cmp Action[+3],0          ; signed
00EF cmp Action[+3],20
00F9 mov Action[+3],0          ; <0 或 >20
0103 mov Action[+3],0          ; 非戰鬥
```

`area.field_596` 的完整 writer／遭遇 surprise 投影尚未在本輪關閉；作品中立
scheduler 必須接受 explicit team mask，CoAB 尚未取得原始 mask 時只能傳 0，
並把 surprise fidelity 保持未完成，不能自行由敵我 side 猜測。

## 下一名 combatant

primary 結構掃描定位 `overlay 08:01FB`；typed resolver：

```text
overlay 8:01FB -> entry 3 -> resident stub 002Fh
GAME.EXE file offset 10BFh
```

函式以 `DS:9598h` 為 TeamList head，以節點／Player `+18Ah` 走 next pointer。
local 變數 `bp-1=maxDelay`、`bp-2=maxRoll`、`bp-3=currentRoll`：

```text
0201 maxDelay=0
0205 maxRoll=0
0220 currentRoll=roll_dice(100,1)
023D currentDelay=Action[+3]
0242 if currentDelay > maxDelay: maxRoll=currentRoll
025B if currentDelay < maxDelay: skip
0264 if currentRoll < maxRoll: skip
026F maxRoll=currentRoll
027A maxDelay=currentDelay
0281 selected=current TeamList node
0290 current=current[+18Ah]
02A3 if maxDelay==0: selected=null
```

因此不能用 fighter ID 排 tie，也不能只為 delay 非零者抽 d100；兩者都會改變
原版 PRNG continuation。

## remake 契約

- 可重用 engine 提供純資料的 DEX table、delay writer、一次 TeamList selection
  與完整 round order helper；亂數由 caller 注入，不認識 CoAB 名稱、ECL 或 UI。
- `Battle` 保存建構時的 combatant order，不能從 Go map 或字串 ID 重建順序。
- party 與 MON*CHA 都投影 shared Player `+17h` Dexterity；舊 `+1A5h` 值不能再
  當成 initiative bonus。該欄位真正語意仍是 `unknown`，保留 raw 時不得改名。
- `StartRound` 依序完成所有 `1d6` writer，再以同一 RNG 串流反覆選人；每選一名
  後把工作副本 delay 歸零，產生本輪每人一次的基線 turn list。日後 UI 的
  `DELAY` 指令須改用動態 scheduler 重新入列，不能把這份基線 helper 當成已完成
  DELAY fidelity。
- `Action +3` 在尚未行動時非零、一般行動完成後為零，繼續供 Blink `25h`
  visibility consumer 使用。

## 未完成邊界

- `area.field_596` 的所有 writers、ECL surprise opcode 與 party ranger 分支。
- 玩家 `DELAY` 命令的 UI、20→19 handoff、重新抽 d100 與同輪第二次行動。
- DOS 版本是否 byte-for-byte 使用相同 routine／PRNG；本輪 exact 只限 PC-98。
- PC-98 `roll_dice` 底層 PRNG 演算法與 save/load continuation。

上述項目未完成前，不得宣稱完整戰鬥 scheduler 或原版 RNG fidelity。
