# 第五百二十八輪：PC-98 MOVEPARTY result／action transaction 邊界

狀態：READY（靜態 raw bytes／控制流；不宣稱秘密門、開門語意或正常路徑）

日期：2026-08-10

## 本輪目的與結論

第 527 輪已閉合 `MOVEPARTY` 的 B／P／K token 與第三平面 raw writer。本輪再
沿同一段連續 bytes 追到 result branch、action input、共用續跑點與第二次
map-cell service call，讓下一輪 runtime probe 有固定的觀測邊界。

可由 raw bytes 精確證明：

1. `017C:0039` 在 overlay-14 local `0x0C06` 被呼叫，回傳 `AL=1` 時直接把
   local success byte 寫成 `1`。
2. `AL=2` 或 `AL=3` 都進入同一段後續流程；它們不是「已證明的門型」名稱。
3. action input 經 `0164:0039` 取得後，程式直接比較 `B`、`P`、`K`，分別呼叫
   local `0x02F5`、`0x05B4`、`0x0714`。
4. `P` 路徑在 local `0x0D37` 再呼叫一次 `017C:0039`，只有回傳 `AL=2` 才
   呼叫 `0x05B4`；其他結果清除 `DS:0xA81B`。
5. 三種 helper 的結果共用一個 success byte；非零才呼叫 local `0x0807`，
   然後所有分支都到 `014A:00DE`。

這些是 exact 的控制流與 call-site，不足以安全命名 `017C:0039` 的 map-cell
service、`0x0807`、`014A:00DE`、第三平面 raw `01`、`AL=1/2/3` 或 B/P/K
為秘密門、開門、技能、ECL flag、另一側 cell 或劇情結果。`017C:0039` 也不
得直接等同 TPOV resolver 的 overlay-30 local `0x0039`，更不能因此等同
`BLOCKCODE=017C:04DEh`。

## 輸入與位址基準

| 證據 | SHA-256／版本 | 位址基準 | 等級 |
|---|---|---|---|
| `workplace/ida406/pc98-overlays/overlay-14.bin` | `a8e03ba9a5381c3a9f7ab411ced3262b21e0b65b948160d614386d677610e7b9` | overlay-local code offset | exact raw |
| PC-98 `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland symbol／resident linear；承接第 526 輪 | provenance |
| `scripts/research/pc98_overlay14_action_transaction_audit.py` | `2eab1dfcf0826afcba98a1bce60d36b1f9bf4fc0296467c43ca9366861adf65d` | raw bytes；不回寫 binary／IDA | exact procedure |
| GNU `objdump` | Docker `coab-go-test:20260729` | binary file VMA = overlay-local | 交叉驗證 |
| IDA Pro | 既有 9.4 baseline database／assembly；本輪未重新寫入 | 既有函式／symbol 邊界 | provenance |

原始 FDD 只作唯讀 runtime 嘗試；本輪沒有修改 FDD、executable 或 baseline
IDA database。NP2kai 已能顯示 PC-98 BIOS、NP2 menu 與 FDD selector，但開啟
Disk 1 後沒有取得可用遊戲 loader／地城 trace，因此不能把畫面嘗試寫成
runtime evidence。

## Exact control-flow boundary

### 第一次 map-cell service 與 result 分流

```text
overlay14:0x0C06  call far 017C:0039
overlay14:0x0C0B  cmp AL,01h
overlay14:0x0C0F  mov [BP-01h],01h       ; AL=1 branch
overlay14:0x0C16  cmp AL,02h
overlay14:0x0C1A  cmp AL,03h
overlay14:0x0C21  call far 0164:002F       ; AL=2/3 common path
```

`AL=1` 的 local success assignment、`AL=2/3` 的共同分流與 call-site 是
exact。`AL` 的正式欄位名稱、比較值代表的 map 狀態，以及 `0164:002F` 的
產品語意仍是 unknown。

### Action input 與 B/P/K dispatch

```text
overlay14:0x0D0F  call far 0164:0039
overlay14:0x0D1A  cmp AL,42h             ; B token
overlay14:0x0D1F  call near 0x02F5
overlay14:0x0D27  cmp AL,50h             ; P token
overlay14:0x0D37  call far 017C:0039
overlay14:0x0D3C  cmp AL,02h
overlay14:0x0D41  call near 0x05B4
overlay14:0x0D50  cmp AL,4Bh             ; K token
overlay14:0x0D55  call near 0x0714
```

輸入 token 與 helper target 為 exact。B／P／K 可以暫作 UI／手冊的工作標籤，
但 helper 的能力條件、物品／技能來源、地圖效果與成功訊息不可由這段 bytes
單獨命名。

### 共用 success continuation

```text
overlay14:0x0D5B  cmp byte [BP-01h],00h
overlay14:0x0D62  call near 0x0807       ; 只有非零結果會呼叫
overlay14:0x0D65  call far 014A:00DE     ; common tail
```

這只能證明「非零 success byte 會多一次 local call，接著所有分支都有共同 far
call」。不能把 `0x0807` 命名成移動、開門、更新座標，也不能把 `014A:00DE`
命名成重繪、事件續跑或地圖服務；這些要由 consumer／runtime trace 閉合。

> ★ **記 call site 位址時要指向 opcode，不是 displacement。**
> near call 是 `E8 <disp16>`：`0x05A4` 是那條指令，`0x05A5` 只是 displacement
> 的第一個 byte。從反組譯輸出裡抄行號很容易差這一格，而差一格之後
> xref 查不到、consumer 也對不上。

## 對 remake 的限制

本輪維持 fail-closed：

- 不新增 `secret_door`、`map_action`、movement predicate 或 ECL flag JSON。
- 不讓 `AL=1/2/3` 直接轉成可通行／開門結果。
- 不把 `0x0807` 或 `014A:00DE` 寫成 engine API 的作品中立語意。
- 不把 PC-98 FDD 操作停在畫面 selector 就當成原版動態驗證。

下一個真正能解鎖 runtime 實作的最小邊界，是從可啟動同版本 oracle 取得：

1. `(13,10)` 各 token 的輸入、map-cell service return 與 helper return；
2. action 前後 `THE3DMAP+300h` selected／相鄰 cell 的 bytes；
3. `BLOCKCODE`／`WALLCODE` 的實際回傳與下一步方向；
4. 離開、重訪、save-load／重新載入 GEO 後的持久性；
5. 若有 ECL continuation，記錄 flag／work address 與下一個 block。

在這些證據出現前，正常玩家路徑仍應把 `wall=09/detail=0` 視為 blocked，
不能用 direct-entry 或 BFS probe 代替玩家驗收。

## 可重現驗證

```text
docker run --rm --network none --cpus 1 --memory 256m --pids-limit 64 \
  --user <current-uid>:<current-gid> \
  -v <repo>:/repo:ro coab-go-test:20260729 \
  python3 /repo/scripts/research/pc98_overlay14_action_transaction_audit.py \
  /repo/workplace/ida406/pc98-overlays/overlay-14.bin
```

預期輸出必須同時列出原始位址、far／near target、result 分支、B/P/K token、
共同續跑與 `action_success_predicate=unknown`。本規格是 READY 的靜態證據
邊界，不是完整秘密門、完整戰鬥、完整中文化或三平台 release 證據。
