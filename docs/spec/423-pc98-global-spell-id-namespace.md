# 第 423 輪：PC-98 全域法術 ID 命名空間

狀態：`READY`（限 Player spell bytes、全域 spell table identity 與既有法術分派）

## 結論

DOS／PC-98 Player record 的已記憶法術 byte 是全域 spell table ID，不是每個
職業各自從 1 起算的 class-local 序號。牧師 Protection From Good 是 `07h`；
魔法師 Magic Missile 是 `0Fh`。舊 spec 134／142 對玩家 Magic Missile `07h`
與「依施法職業解釋重疊 ID」的斷言由本規格 supersede。

## 原始證據

輸入：`workplace/ida406/PC98-GAME.EXE`

- 長度：151,230 bytes。
- SHA-256：
  `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`。
- MZ file offset 公式：
  `09A0h + 0C29h*10h + 61B4h + spellID*10h`。
- `07h @ 012E54h`：
  `00 01 00 00 00 03 04 02 00 04 09 02 04 01 00 00`。
- `0Fh @ 012ED4h`：
  `02 01 06 04 00 00 04 00 00 04 00 01 01 04 01 00`。

IDA overlay 09 的 Quick selector 直接以 Player spell list byte 乘 `10h`，再
讀取 `DS:61C1h`（record `+0Dh` priority）；中間沒有減去或加上職業表基底。
同一 consumer 在 `03D3h..04C9h` 讀取 record `+0Eh／+0Fh` 作 suitability，
在 `0627h..0754h` 從 Player list 隨機選取同一 ID。既有 DOS parser 也從
record `+01Eh` 原樣保存 memorized bytes，known flags 則以全域一基底索引保存。

### 推論等級

- **exact**：上述 file offsets、16-byte records、IDA 指令中的 `ID*16` 與
  `+0Dh／+0Eh／+0Fh` consumer。
- **material-exact**：DOS parser 原樣保存 record spell byte，與 PC-98
  consumer 使用方式一致；兩平台資料結構可交叉支持全域 identity。
- **hypothesis**：尚未逐一命名全部 100 筆 spell records；未命名欄位不得因
  secondary source 名稱直接升格成正式規則。

## 實作邊界

- 玩家 Magic Missile 改用 `0Fh`，與怪物端既有 `0Fh` 一致。
- camp 一級法術名稱以全域起點解析：牧師 `01h`、目前已翻譯的魔法師表
  `09h` 起算。跨職業 ID 不再被錯誤套用另一職業名稱。
- Protection From Good 與 Magic Missile 的 targeting／availability／cast
  分派不再依相同 `07h` 加 caster class 猜測。
- 原始未知 ID 仍顯示十六進位，不自行重編號。

本輪不宣稱 ALT+M Quick spell AI 已完成。雖已 exact 解出三次隨機候選、
priority tier 與 `+0Eh／+0Fh` suitability，完整 spell cast consumer、所有法術
效果及未支援法術的 fail-closed handoff 尚未完成，因此 UI 繼續不開放 ALT+M。
