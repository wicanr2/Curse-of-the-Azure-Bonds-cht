# 第五百二十七輪：PC-98 MOVEPARTY action 與 map-field writer 邊界

狀態：READY（靜態 raw bytes／控制流邊界；不宣稱秘密門、成功規則或正常路徑）

日期：2026-08-10

## 本輪結論

本輪只追查目前正常玩家路徑的 P0-2 阻塞點：騎士事件後，隊伍在
PC-98 GEO2 的 (13,10) 遇到 wall=09/detail=0，需要知道 MOVEPARTY 的
action 是否會改寫 loaded map 的第三平面。

在既有 PC-98 overlay-14 raw 副本中，可以閉合兩種不同的第三平面 field
操作，以及它們和 MOVEPARTY action helper 的靜態呼叫關係：

1. overlay-local 0x003E 會保留 selected 2-bit field 以外的位元，將選定
   field 清為 raw 00。
2. overlay-local 0x014C 會保留其他位元，將選定 field 寫成 raw 01。
3. MOVEPARTY 的 B、P、K 輸入比較 bytes 可直接回查，分別呼叫
   local 0x02F5、0x05B4、0x0714。
4. B helper 與 P helper 各有兩個直接呼叫 0x014C 的位置；MOVEPARTY
   的另一條 movement-result 分支直接呼叫 0x003E。

這把 P0-2 的靜態範圍從「只知道有一個清除 writer」縮小為
「B/P helper 的 set writer、movement-result 的 clear writer 與 action
dispatch 已知」。但仍不能由 AND／OR 的 raw 值命名成「開門」或「關門」：
selected cell 的 detail 意義、action 成功條件、另一側 cell 的正式對稱規則、
BLOCKCODE 回傳值、ECL flag、離場／重載後持久性與 (13,10)→(8,15) 的
正常 runtime 路徑都尚未閉合。

因此本輪不新增 secret_door、search 或 movement JSON，也不把 (8,15) 寫入
正常玩家 regression。

## 輸入、工具與位址基準

| 證據 | SHA-256／版本 | 位址基準 |
|---|---|---|
| PC-98 overlay-14 raw 副本 workplace/ida406/pc98-overlays/overlay-14.bin | a8e03ba9a5381c3a9f7ab411ced3262b21e0b65b948160d614386d677610e7b9 | overlay-local code offset |
| PC-98 PC98-GAME.EXE | 8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0 | Borland symbol／resident linear；沿用第 526 輪輸入 |
| scripts/research/pc98_overlay14_action_writer_audit.py | 9fe8c69be1e39b8cc1cf173633a57a6ce7c4d9e4c4a6c75edd26f0016ead1573 | raw bytes；不回寫 IDA／binary |
| Python raw audit | Python 3.11.2（Docker coab-go-test:20260729） | 檔案 offset／overlay-local |
| objdump 交叉驗證 | GNU objdump（同一 Docker image；僅作連續 bytes 可讀化） | binary file VMA = overlay-local |
| IDA Pro | 既有 9.4 baseline assembly／database provenance；本輪未重新寫入 database | 只作既有函式／symbol 邊界交叉查詢 |

本輪沒有在主機執行分析 workload，也沒有修改原始 FDD、executable 或 baseline
IDA database。raw audit 只讀取輸入；推論語意另外寫在本規格，不回寫工具內
原始名稱。

## MOVEPARTY action dispatch

### 可直接證明的 bytes

overlay-14 local 0x0BCC 是 MOVEPARTY 原始 entry bytes。其 action branch
保留下列比較：

~~~text
0D1A  cmp al,42h   ; token B
0D27  cmp al,50h   ; token P
0D50  cmp al,4Bh   ; token K
~~~

連續 near call 的 target 是：

| caller | target | 可直接證明的內容 | 推論等級 |
|---:|---:|---|---|
| 0x0D1F | 0x02F5 | B branch 呼叫 local helper | exact |
| 0x0D41 | 0x05B4 | P branch 呼叫 local helper | exact |
| 0x0D55 | 0x0714 | K branch 呼叫 local helper | exact |
| 0x0CD4 | 0x003E | movement-result flag 通過後呼叫 clear writer | exact；flag 語意未知 |

B、P、K 只代表程式實際比較的輸入 token；它們和玩家介面中的 Bash、Pick、
Knock 對應可作為目前 UI／手冊的工作標籤，但不把 helper 的全部成功語意當成
已由本輪 raw bytes 證明。

### set writer 的直接呼叫

local 0x014C 的四個方向分支會讀取 THE3DMAP 的 +300h，使用方向對應 mask
保留其他 6 bits，再 OR 入該方向的 raw 01 field：

| direction selector | 保留 mask | OR 值 | selected field 的 raw 結果 |
|---:|---:|---:|---:|
| 06h | 3Fh | 40h | 01 |
| 04h | CFh | 10h | 01 |
| 02h | F3h | 04h | 01 |
| 00h | FCh | 01h | 01 |

這個 helper 的直接呼叫位置是：

| caller | 所在 action helper | 目前可證明的內容 | 推論等級 |
|---:|---:|---|---|
| 0x0566 | 0x02F5 | 第一次 set-field call | exact |
| 0x05A4 | 0x02F5 | 第二次 set-field call | exact |
| 0x062F | 0x05B4 | 第一次 set-field call | exact |
| 0x066D | 0x05B4 | 第二次 set-field call | exact |

第二次呼叫前的 bytes 會以 table 值加到目前座標，並形成另一個方向參數；
這是 raw arithmetic／call-site 事實，但該座標是否就是門的另一側、是否要
同步其他 GEO cell，以及 field 01 在 BLOCKCODE 中的正式名稱，仍是
strong inference／unknown。

## clear writer 與 set writer 的對照

兩個 helper 都以同一個位址公式取 selected cell：

~~~text
THE3DMAP + 0x300 + (arg_0A << 4) + arg_08
~~~

local 0x003E 的方向分支是：

~~~text
direction 06h: old & 3Fh
direction 04h: old & CFh
direction 02h: old & F3h
direction 00h: old & FCh
~~~

local 0x014C 則在相同保留 mask 後 OR 入 40h／10h／04h／01h。因此
「selected field 被清為 00」與「selected field 被寫成 01」是 exact raw
operation；「無牆／有門／開門／關門」都不是本輪可合法宣稱的語意。

0x003E 本身只寫入一個由參數選定的 cell。0x014C 的兩次 call 以及
呼叫端的座標算術顯示可能存在 pair／opposite-cell transaction，但
是否是正式雙側門狀態仍需 runtime before／after 及重載證據。

## 對 remake 的影響

本輪維持 fail-closed：

- 不新增 secret_door、search、map_mutation 或 movement predicate。
- 不把 B、P、K 的 helper return 直接轉成 wall=09 可通行。
- 不把 local 0x014C 的 raw 01 命名成 detail=1、door_open 或劇情旗標。
- 不把 direct-entry／允許 wall 09 的 BFS probe 算作正常玩家路徑。

下一個最小 runtime boundary 是在同一版本 PC-98／DOS 可啟動 oracle 中記錄：

1. (13,10) 每個 action token 的輸入與 helper return；
2. action 前後 THE3DMAP+300h 的 selected cell 與相鄰 cell bytes；
3. BLOCKCODE／WALLCODE 的結果與再次前進的方向；
4. action 後離開、重訪、save-load／重新載入 GEO 的狀態；
5. 若有 ECL continuation，記錄 flag／work address 與下一個 block。

目前使用者提供的 PC-98 FDD 在既有 NP2kai probe 中尚未通過完整遊戲 loader，
所以本輪沒有捏造 runtime trace。DOSBox 原版若能以正常玩家輸入抵達同一格，
仍是首選 oracle。

## 可重現驗證

~~~text
docker run --rm --network none --cpus 1 --memory 256m --pids-limit 64 \
  --user <current-uid>:<current-gid> \
  -v <repo>:/repo:ro coab-go-test:20260729 \
  python3 /repo/scripts/research/pc98_overlay14_action_writer_audit.py \
  /repo/workplace/ida406/pc98-overlays/overlay-14.bin
~~~

本輪 audit 必須回報：

- exact_file_hash
- clear writer 四個方向 mask
- set writer 四個方向 mask／OR 值
- B/P/K helper target
- 四個 set-writer call site
- movement-result clear-writer call site
- action_success_predicate=unknown

本規格是 READY 的靜態證據邊界，不是完整秘密門、完整戰鬥、完整中文化或
三平台 release 證據。
