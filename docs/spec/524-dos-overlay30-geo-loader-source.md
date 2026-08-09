# 第五百二十四輪：DOS overlay-30 `GEO<區域>.dax` loader 與四平面來源

狀態：`READY`（限 GEO resource source／`0402h` decoded-size gate／四平面格式對應；
不含普通鍵盤、地圖位置 handoff 或 secret-door writer）
日期：2026-08-09

## 結論先行

第 522 輪已證明 overlay-30 把暫存輸出分四段搬到 `DS:7206h`，但仍把
「`.dax` record 來源、`0402h` 的正式格式與四個 plane 語意」列在未知。本輪把
overlay-30 local `1310h..139Eh` 的來源檔名片段、區域值、loader call 與既有
GEO DAX corpus 對齊：

```text
overlay-30 local 1310h
  03 'GEO' + 04 '.dax'                 （raw Pascal string fragments）
        ↑
DS:5BEEh → 0A54:12ABh，width=1       （插入區域數字的 builder candidate）
        ↓
0636:08DEh，selector = [bp+6]         （GEO block selector）
        ↓
decoded-size gate = 0402h
        ↓
output + 002h → DS:7206h + 000/100/200/300，各 0100h
```

原始 DAX container 的 GEO2–GEO6 corpus 每個 decoded block 都是 `0x402` bytes；
前 2 bytes 是 block prefix，後面正好是四個 `0x100` plane。既有 GEO parser／
spec 56 對四平面格式的語意已是 `READY`：前兩個 plane 是兩組 wall direction
nibble，第三個是 terrain／background byte，第四個是四組 2-bit direction detail。
因此本輪可以移除「這條 loader 的 DAX record／四平面格式完全未知」的現行斷言。

仍須保留的邊界是：完整 Pascal 字串操作與 runtime file-open trace 尚未取得，
所以 `GEO<DS:5BEEh 十進位值>.dax` 是 `strong inference`；`DS:5BEEh` 的正式
全域欄位名稱、block selector 的上游 producer、`DS:7206h` 對應的地圖座標與
普通輸入仍不能只靠本輪命名。這一輪不修改 engine、CoAB JSON 或 movement 規則。

## 1. Provenance 與位址空間

| 輸入／工具 | SHA-256／版本 | 位址／用途 |
|---|---|---|
| extracted DOS `overlay-30.bin` | `444893f6d239cc57f555287786e3704bc801dd63f3d1f1f4d4ac14e8742652d468` | overlay-30 local offset |
| DOS archive `curseoftheazurebonds.zip` | `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d` | ZIP member source |
| archive member `GEO2.DAX` | `1d4fe936f9d78b6f7d7ef689c78ebb8f86c0e68a9e1330b0a371839f9fea1862` | DAX container／decoded blocks |
| IDA | `IDA Pro 9.4`，`ida-pro-9.4-ver2:uidfix-v1` | disposable overlay `.i64` |
| IDA audit script | `2ceec4bd897fb001b468c4a074fce30793a647caf37abc2984be445f150ca76b` | `scripts/ida/dos_overlay30_geo_loader_audit.idc` |
| IDA audit report | `905c9835ac5b2123a6af902e64c8f409ba53343037ec9ff5dcf2253eecf437ad` | overlay-local raw／code evidence |
| DAX audit tool | Python `3.12.3`; `scripts/dax_dump.py` SHA `3a70ecc25f9d2a94f1093f027cc26394ea49167a14bfa47eff6d96c3faecfc55` | archive member／RLE decode |
| GEO format spec | `docs/spec/round-56-geo-map-geometry.md` SHA `8bc10f5b38b32cf9749ce2242827b0ee270b2cd56b9c7265b4353e15a06042dd` | existing parser／plane semantics |

原始 `START.EXE`、`GAME.OVR` 與 baseline `.i64` 仍是唯讀輸入；本輪 IDA 只在
Docker disposable overlay database 上執行。`overlay-30` local offset、resident
`0636:08DEh`／`0A54:12ABh` 與 `DS:7206h` work address 不在同一位址空間，本文
刻意並列而不合併。

## 2. overlay-30 的原始 bytes 與呼叫資料流

### 2.1 資源名稱片段

local `0x1310..0x1339` 的連續 raw bytes 是：

```text
03 47 45 4F 04 2E 64 61 78
20 55 6E 61 62 6C 65 20 74 6F 20 6C 6F 61 64
20 67 65 6F 20 69 6E 20 4C 6F 61 64 33 44 4D 61 70 2E
```

以 Pascal short-string 格式讀取開頭兩個欄位：

```text
0x1310: 03 47 45 4F       → "GEO"
0x1314: 04 2E 64 61 78    → ".dax"
```

`0x1318` 之後是相鄰的錯誤訊息資料，不能把它和檔名字串拼成一個更長的檔名。
local `0x1361` 把 `0x1310` 當作 source，local `0x137B` 把 `0x1314` 當作另一個
source，兩者分別經 resident `0A54:0634h` `Store string`，中間由
`0A54:06C1h` `Concat` 組合；這是 raw call sequence，不是自訂 rename。

### 2.2 區域值與 loader call

local `0x1341..0x1356` 的連續指令是：

```text
1341  A0 EE 5B             mov al,ds:5BEEh
1344  30 E4                xor ah,ah
1346  31 D2                xor dx,dx
1348  52                   push dx
1349  50                   push ax
134A  31 C0                xor ax,ax
134C  50                   push ax
134D  8D 7E F6             lea di,[bp-0Ah]
1350  16 57                push ss:di
1352  B8 01 00             mov ax,1
1355  50                   push ax
1356  9A AB 12 54 0A       call far 0A54:12ABh
```

這可以精確描述為：從 `DS:5BEEh` 讀一個 byte、清除高位、以 width `1` 呼叫
resident helper，並把結果放入 local string variable。結合前後的 `GEO`／`.dax`
片段，完成檔名是 `GEO<區域值>.dax` 的 `strong inference`；正式全域欄位名稱
與 runtime `Assign／Reset` 的 file-open trace 仍待補。

local `0x1385..0x1393` 將 block selector 與輸出位置傳給 loader：

```text
1385  8A 46 06             mov al,[bp+6]
1388  50                   push ax
1389  8D 7E FE             lea di,[bp-2]     ; returned decoded size word
138C  16 57                push ss:di
138E  8D 7E FA             lea di,[bp-6]     ; returned far pointer variable
1391  16 57                push ss:di
1393  9A DE 08 36 06       call far 0636:08DEh
```

`[bp+6]` 是此 overlay function 的 selector input；其作品中立的正式名稱仍由
上游 caller／map service 決定，不能只叫作座標或 map block。

### 2.3 `0402h` gate 與四平面 handoff

local `0x1398..0x13A3`：

```text
1398  83 7E FE 00          cmp word [bp-2],0
139C  74 07                jz  13A5h
139E  81 7E FE 02 04       cmp word [bp-2],0402h
13A3  74 2E                jz  13D3h
```

非零但不是 `0402h` 會走錯誤／清理路徑；這不是「所有 DAX block 都可送入此
map buffer」的通用規則。成功路徑從 output pointer `+002h` 開始，四次
`@Move` 的 destination／count 是：

| source | destination | count |
|---|---|---:|
| output `+002h` | `DS:7206h +000h` | `0100h` |
| output `+102h` | `DS:7206h +100h` | `0100h` |
| output `+202h` | `DS:7206h +200h` | `0100h` |
| output `+302h` | `DS:7206h +300h` | `0100h` |

同一 output pointer 與返回 word 最後傳給 resident `0A54:0364h @FreeMem`。這與
第 522 輪 writer spec 完全一致，這輪新增的是「來源確實是 GEO DAX、`0402h`
是其 decoded block size、`+002h` 是四平面 payload 起點」的資料來源閉合。

### 2.4 已丟棄的 caller probe

本輪曾在 disposable overlay database 以整段線性 `create_insn` 後查 local `133Ah`
的 direct cref；輸出把資料字串尾端的 `70 2E`（`"p."`）當成假條件跳躍，形成
`0x1338 → 0x133Ah` 的 data/code 混讀假 xref。該 probe 與報告沒有版本化，也不
支持任何 caller 結論；已刪除 probe 檔，後續 caller 追查必須使用原始函式邊界、
control vector／loader trace 或其他能區分 data 與 code 的證據。這筆勘誤保留在
spec，避免 compact 後把假 xref 當成新進度。

## 3. 原始 GEO DAX corpus 對照

Python DAX audit 對 ZIP member `GEO2.DAX` 的輸出：

```text
data_offset=29 blocks=3
block=1 raw=1026 packed=968
block=3 raw=1026 packed=900
block=4 raw=1026 packed=982
```

`1026 = 0x402`；以 little-endian 讀取 decoded block：

```text
prefix       = bytes[0:2]       = 00 04  （0x0400 payload size）
plane 0      = bytes[2:0x102]   = 0x100 bytes
plane 1      = bytes[0x102:0x202] = 0x100 bytes
plane 2      = bytes[0x202:0x302] = 0x100 bytes
plane 3      = bytes[0x302:0x402] = 0x100 bytes
```

這個 offset 對照與 overlay-30 的 `output+002h`／四次 `0100h` copy 是逐 byte
區段一致的。`GEO3.DAX`～`GEO6.DAX` 的所有 decoded blocks 也同樣是 `0x402`；
完整 plane 語意沿用 [`round 56`](./round-56-geo-map-geometry.md)，不在本 spec
重複發明另一份 parser 命名。

## 4. 推論等級與清理結果

| 結論 | 等級 | 目前界線 |
|---|---|---|
| local `1310h`／`1314h` 是 `GEO`／`.dax` Pascal string fragments | `exact` | raw bytes、Store／Concat call-site |
| `DS:5BEEh` 的 byte 參與區域數字字串建構 | `exact` | `mov al,ds:5BEEh`、`width=1`、resident helper call |
| overlay-30 loader 只接受 decoded size `0402h` 才進四平面 copy | `exact` | `1398..13A3` raw branch |
| 此 loader 的資源路徑是 `GEO<DS:5BEEh 十進位值>.dax` | `strong inference` | 字串片段／concat／區域值已閉合；仍缺 file-open trace |
| `0402h` 對應 GEO DAX block 的 2-byte prefix＋四個 `0x100` payload planes | `exact` | overlay bytes＋GEO2–GEO6 corpus／round 56 parser evidence |
| 四個 plane 的現有格式語意 | `exact`（限 GEO parser） | 兩組 wall nibble、terrain／background、direction detail；不等於 UI／碰撞規則 |
| `DS:5BEEh` 正式欄位名、selector 上游 producer、座標／方向投影 | `unknown／strong inference` | 仍需 consumer／runtime trace |

因此工作清單中「`.dax` record／四平面格式完全未知」應刪除；改成「GEO source／
payload layout exact，區域 builder 的正式全域欄位名、上游 selector、map projection
與 DOSBox handoff 未完成」。本輪不把 static resource closure 擴大成完整地圖或
正常移動完成。

## 5. 下一個最小工作

1. 以同一 `GEO2.DAX`／`GEO6.DAX` source 追 `[bp+6]` 的 selector producer，確認
   它如何由 ECL／map service 選出 block；保留 ECL block ID、DAX block ID、
   overlay-local offset 三個欄位。
2. 追 `DS:7206h` 四平面各自的正式 consumer 與 `DS:7212／7213`／`7Fh` sentinel
   讀取；不要因 round 56 的 parser 語意直接命名所有 runtime state。
3. 再做 DOSBox 正常鍵盤／目的地／座標／方向／重繪 trace；若沒有可用 DOSBox
   image，先建立與既有 IDA image 分離且可重現的工具鏈，不在主機執行。
