# 第 558 輪：PC-98 ECL `TREASURE → COMBAT` 邊界勘誤

狀態：`READY`（限三個靜態候選、interpreter dispatch、VM pause／resume 與既有
State transaction；不代表全域 ordered-effect log 已完成）

## 結論

第 557 輪把三組 `TREASURE → COMBAT` 列為第一批待閉合項目，但沒有先連結既有
第 255／257／258 輪 READY 規格。重新稽核後，這三組不是尚未實作的玩家 blocker：

- `TREASURE (27h)` 先消耗八個 operand，建立 pooled money／item pending state；
- 後續 `COMBAT (24h)` 才依當時 engine state 派送商店、神殿、一般戰後服務或
  真正戰鬥；
- 有生怪時，remake 將同一份 raw treasure request 掛到 encounter，勝利後才解析，
  再由 `COMBAT` 下一條 PC 恢復原 ECL；
- 無生怪時，`COMBAT` 是立即 treasure／service boundary，而不是零怪物戰鬥。

因此不應為這三組另建一套 transaction 或改寫現有 State。仍未完成的是其餘候選的
跨類型全域 ordered-effect log、動態 branch 與正常玩家事件 coverage。

## 原始輸入、工具與位址空間

| 輸入／工具 | SHA-256／版本 | 位址空間 |
|---|---|---|
| `curseoftheazurebonds.zip` | `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d` | DOS DAX decoded payload offset；code address=`8000h+offset` |
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland symbol `segment:handler offset`／resident control stubs |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | TPOV overlay-local code offset |
| `overlay-02.bin` | `49a35a10c3a08def6fcf7d17cfa48fb58e6ce70d88d754163c54c1d074d338a4` | `INTERPET` overlay-local code offset |
| `overlay-05.bin` | `4136fc40ec5647cac6ef7cbc762ecb8a2d4ada045020bff4646290d267864cdf` | `POSTCOM` overlay-local code offset |
| IDA Pro | `ida-pro-9.4-ver3:latest`／IDA Pro 9.4；image ID `8ca4202a…a72fa` | disposable raw 8086 database |
| IDA 稽核腳本 | [`pc98_ecl_treasure_combat_audit.idc`](../../scripts/ida/pc98_ecl_treasure_combat_audit.idc)，SHA-256 `8e08f806cb9e3e7155fdd0c2271a4ab3d9e17957cfd5f8cee274e66961fa27ba` | 只輸出原始 local offset、bytes、disassembly；不 rename |

IDA 的 IDAPython 仍受主機 Python 路徑影響；本輪使用已驗證可用的 IDC，並以明確
報告內容而非 exit code 判斷成功。原始 binary、overlay 與基準 `.i64` 均未修改。

## 位址基準勘誤

Borland symbol table 將 `INTERPET.GOECL` 列為 `0037:3A21`、
`POSTCOM.DOPOSTCOMBAT` 列為 `0057:1775`。這些是各 module 的 handler offset，
不能把相同數字誤當 resident stub offset。typed resolver 的實例是：

```text
overlay 5 resident stub 0025h → handler-local 1775h
overlay 7 resident stub 0025h → handler-local 0296h (ADDRESSVALUE)
overlay 7 resident stub 002Ah → handler-local 008Eh (READVAR)
```

這保留了 `segment:offset`、resident stub 與 overlay-local 三種不同位址空間；第 558
輪一度將 symbol offset 直接當 stub 查詢而無法解析，現已明確訂正，未把失敗命中
寫入正式語意。

## PC-98 interpreter 的 exact bytes

`INTERPET` local `388Ah` 比較 opcode `24h`，`388Fh` 呼叫 local `1820h`；
local `38A4h` 比較 opcode `27h`，`38A9h` 呼叫 local `1BEAh`。這是原始 dispatcher
到 handler 的 `exact` control-flow。

`27h` handler local `1BEAh`：

- `1BF1h..1BF4h` 以值 `08h` 呼叫 ECL2 `READVAR`；
- `1C02h..1C2Ah` 逐一取得 operand 1..7，依四 byte stride 寫入 `A00Ah` 起的
  七個 pooled value 槽；
- `1C2Ch..1C34h` 再取得第八個 operand；
- `<80h` 與 `>=80h` 分別進入資料 block／random item 路徑；
- handler 最終於 `2121h..2124h` 返回。

`24h` handler local `1820h`：

- `1826h` 先遞增 `DS:7F21h`，即 interpreter cursor 已越過這個無 operand opcode；
- `183Eh..1963h` 依 party／area 欄位選擇 shop、temple 或其他 service；
- `19DCh..1A71h` 是實際 combat preparation／loop／postcombat 呼叫路徑；
- `1ADCh..1B27h` 依結果更新 mode、遮罩 state，最後返回 dispatcher。

`POSTCOM.DOPOSTCOMBAT` local `1775h..19A2h` 會清理戰鬥狀態，並在
`1963h..199Ah` 清除多個 area/combat modifier 欄位。這支持戰後處理是獨立階段；
本輪沒有僅憑 offset 猜測各未具名欄位的產品語意。

以上只證明 PC-98 同系 engine 的 handler ordering 與 service/combat dispatch；
DOS ECL candidate 的具體 operand、分支與 remake continuation 仍由下節 raw DAX
與實際 VM 回歸交叉驗證，沒有跨平台把所有欄位地址視為相同。

## 三個 DOS 真實候選

| 候選 | raw sequence | treasure | COMBAT 後 PC／續跑 |
|---|---|---|---|
| `ECL3.DAX/15h +050A..+0578` | `+0536 CLEARMONSTERS`、動態 `LOAD MONSTER`、`+055A TREASURE`、`+056D COMBAT` | 4 gems、動態 jewelry／item source | pause=`+056E`；續跑 `4C06+1`、EXIT=`+0578` |
| `ECL4.DAX/25h +1271..+12A7` | `+1288 CLEARMONSTERS`、`+1289 LOAD MONSTER 23h`、`+1291 TREASURE`、`+12A2 COMBAT` | 30 platinum、`ItemBlock=FFh` | pause=`+12A3`；GOTO `+1529`、清圖、`CALL 2E10h`、EXIT=`+1534` |
| `ECL6.DAX/45h +04F6..+0575` | `+0522 CLEARMONSTERS`、動態 `LOAD MONSTER`、`+0557 TREASURE`、`+056A COMBAT` | 4 gems、動態 jewelry／item source | pause=`+056B`；續跑 `4C06+1`、EXIT=`+0575` |

新增 real-image VM regression 以正式 DAX bytes 設定可觀察的動態 operands，逐項
斷言 monster descriptor、raw treasure、`CombatRequested`、pause PC 與 resume PC。
ECL4 的 `PICTURE FFh` 是清除目前圖片，不是 `PictureRequested=true`；測試第一次
把它誤寫成圖片 request，已由實際結果與 raw opcode 契約更正。

## Remake transaction 與證據等級

第 255 輪已將 `TREASURE` 八 operand 建成 raw request；第 257 輪完成 item/random
解析與 pickup；第 258 輪完成：

1. `MonsterSpawns>0` 時把 treasure 延後到 party victory；
2. 勝利後解析／顯示 loot，再由保存的 `treasureResumeECL` 恢復下一條 PC；
3. `MonsterSpawns==0` 時立即派送 treasure service；
4. headless 缺 ITEM DAX 時保留 raw request 與 continuation，失敗即關閉。

三組候選的 R1–R4 可標為 `exact／covered`。本輪的 direct-entry VM regression 是
R4，不冒稱三個 camp-interrupted 事件皆有完整正常玩家 R5。另一方面，第 258 輪
已有火刀首領的正常玩家戰鬥→勝利→loot→夢境→`NEWECL`／世界選單路徑，因此共用
transaction 本身已有 R5 代表案例。

## 清冊與後續工作

`ecl-event-catalog` 格式升至 v2，每個候選帶穩定 ID；受版控的
[`ecl-ordered-effect-reviews.json`](../audit/ecl-ordered-effect-reviews.json) 只按完整
ID 附加 status、confidence 與 spec refs。不存在的 ID 會使生成失敗，避免原始 offset
漂移後把舊結論套到新候選。生成 JSON／Markdown 仍保留原始定位，不讓 review 覆蓋
靜態證據。

下一個 P0 改為：從剩餘 30 個未審查候選選取 `COMBAT → text`，閉合真正的戰後文字
pause／resume，再擴充動態 edge／ordered-effect trace；不再重做這三組。
