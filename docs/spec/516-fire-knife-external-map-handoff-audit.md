# 第 516 輪：火刀據點外部地圖／秘密門 handoff 反組譯盤點

狀態：`DRAFT`（證據盤點；沒有新增遊戲規則）
日期：2026-08-09

## 目的

第 515 輪已把火刀戰後第一個 `(1,8)`→`(13,10)` 位置轉移改成資料包契約，但
來源仍是 `strong inference`。本輪不直接猜測第二個 handoff，而是盤點 DOS／PC-98
目前已能證明的資料流，確認下一個反組譯工作的範圍。

## 輸入與工具

| 輸入／工具 | 雜湊／版本 | 位址基準 |
|---|---|---|
| DOS `ECL2.DAX` | SHA-256 `ec2957d51c53d04a419f47453d345b25a5013f3cab483637acd8986739353338` | ECL block payload offset |
| DOS `START.EXE` | SHA-256 `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5` | DOS executable／runtime |
| DOS `GAME.OVR` | SHA-256 `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215` | GAME.OVR overlay-local |
| PC-98 `PC98-GAME.EXE` | SHA-256 `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland `segment:offset`／overlay-local |
| PC-98 `PC98-GAME.OVR` | SHA-256 `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | TPOV overlay index／code offset |
| IDA | `ida-pro-9.4-ver2:uidfix-v1`，IDA Pro 9.4 | 只操作 Docker 內分析副本 |
| TPOV resolver | `cmd/pc98-ovr-audit` | overlay index、stub／code offset，不與 resident linear EA 混用 |

## 已閉合的局部證據

### DOS block 3 戰後 branch

ECL2 block 3 `+1B5Bh` 的清理 branch 會呼叫 `2E10h`；同一 remake runtime trace
在火刀戰勝利後 PRESS 續跑到該 PC 時，沒有 `C04Bh／C04Ch` `SAVE`，狀態仍在
`(1,8,S)`。`GEO2.DAX` block 3 的 `(1,8)` 與 `(13,10)` 不在同一合法 wall
component，因此「直接沿同一 GEO 走過去」不是解釋。

DOS extracted overlays 的 raw little-endian candidate scan 沒找到 literal `4C28h`。
這只證明沒有直接 literal 命中；指標、通用 interpreter 或資料表間接使用仍然可能，
不能寫成 `4C28h` 未使用，也不能由負面掃描命名 writer。

### PC-98 `S` 與 `SHOWLOCATION`

在 PC-98 `MOVEMENT` overlay 14、Borland segment `00C9h`、overlay-local
`PREMOVEPARTY` `0x0A0F..0x0A28`，連續 bytes 證明：

```text
cmp AL,'S'
load current object +594h
xor AX,1
store current object +594h
call far 014A:00DEh
```

TPOV resolver 的可重現結果是：

```text
stub_resolution overlay=24 stub=0x00DE entry=38 code=0x2E8C flags=0x00 exe=0x1F1E
```

因此 `014A:00DEh` 是 stub，不是可以直接拿來當 overlay-24 code `0x00DE` 的
函式位址；實際 handler 是 overlay 24 `SHOWLOCATION` `0x2E8Ch`。該 handler 會
讀 `+594h` bit 0 後組合位置／狀態文字與畫面，dump 範圍內沒有直接把秘密門的
第三平面寫成 detail 1 的指令。這只證明顯示／搜尋狀態資料流，尚不能宣稱搜尋
沒有其他間接 writer。

### PC-98 wall／door movement

PC-98 overlay 30、Borland segment `017Ch` 的 `BLOCKCODE` `0x04DE` 以目前方向
讀取第三平面 `es:[di+300h]` 的 2-bit detail。其控制流可精確分出：

- wall type 為 0 時進入無牆／特殊判定；
- wall type 非零時 detail `0..3` 由方向對應的 2-bit 欄位取得；
- `MOVEPARTY` 將結果 `1` 視為可走，`2／3` 交給 BASH／PICK／KNOCK／ENTER
  action；
- P 成功的 writer 會把選定門的 detail 設為 `1`，並對另一側再寫一次。

所以候選 `wall=09h／detail=0` 目前可證明是普通 movement blocked；它可能是
秘密門候選，但尚未證明 `S` 會把它改成 detail `1`。

### `BDF0/BDF1` 狀態橋接

PC-98 overlay 2（INTERPET，segment `0037h`）`0x3BB8..0x3BFD` 可見：在
`+594h` 條件成立的分支中，先將 `+594h & FFFDh` 寫入 `BDF1`，暫時把目前
record `+594h` 寫成 `1`；稍後又把 `BDF1` 載回 `+594h`。overlay 11（INIT，
segment `004Ah`）初始化 `BDF0=0`。這是可回查的 record／shared-state 流程，
但目前沒有從這段流追到 `es:[di+300h]` 的 map writer，因此等級仍是
`strong inference／unknown`，不能命名成「秘密門開關」。

補充抽樣：overlay 11（INIT）`0x06BE..0x06CD` 連續 bytes 是：

```text
mov byte ds:0BDF1h,0
mov byte ds:0BDF2h,0
mov byte ds:0BDF3h,0
mov byte ds:0BDF4h,0
```

`cmd/pc98-ovr-audit -word BDF1` 顯示的 overlay-local `0x06C0` 只是上述第一個
指令內的 little-endian 位址運算元（raw bytes `C6 06 F1 BD 00`），不是獨立的
`BDF1` consumer 或秘密門 writer。這是「raw word 命中」與「可執行資料流」的
明確區分；目前不升級 `BDF1` 語意。

同樣地，overlay 2 `0x3BFD` 的 far call 會回到 TPOV `014A:00DEh`，解析為
overlay 24 `SHOWLOCATION` `0x2E8Ch`。它再次支持顯示／位置狀態回饋，但該段抽樣
仍未看到 `es:[di+300h]` 的 writer，不能把畫面刷新誤當成地圖第三平面變更。

## 目前不能保留的斷言

- `docs/spec/297-fire-knife-hideout-transition.md` 舊版所稱「formal new-game
  regression crosses `(8,15)`」不成立：測試當時直接寫入座標再呼叫 exit lifecycle。
  現檔已標 `SUPERSEDED`，保留 E2／`NEWECL 4` 的歷史 bytes 與勘誤。
- `S` 呼叫 `SHOWLOCATION` 不等於 `S` 已開啟秘密門。
- PC-98 的 `BDF0/BDF1`、`+594h`、`+300h` 不可因數值相似而直接映射成 DOS
  `4C28h`、`4BF0h／4BF1h` 或 CoAB GEO 欄位。

## 下一個窄工作

1. 以 `BDF1` overlay-2 consumer 的 caller／前後資料流為入口，找是否有 map
   service 的 address-taken writer；先查 writer，再查名字。
2. 若 PC-98 仍只有顯示／record state，轉回 DOS `GAME.OVR` 的 `2E10h` caller、
   `4BF0h／4BF1h` 間接取址與 runtime DOSBox trace。
3. 只有找到 writer→projection→movement consumer 後，才新增資料驅動
   `secret_door`／`search` contract；在此之前保持 `MoveDungeon` fail-closed。

本規格不支援任何新的正常玩家完成聲明；正式 worklist 見
`docs/knowledge/golden-box-reverse-engineering-worklist.md`。
