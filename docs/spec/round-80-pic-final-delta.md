# 第八十輪：PIC／FINAL animation XOR delta

狀態：`READY`（限 PIC frame stream decode；SPRIT 不使用 delta）

## 已確認

- reference `load_pic_final` 對 `file_name == PIC || FINAL`：保存第一幀 packed EGA bytes，後續每幀先與第一幀逐 byte XOR，再解碼成 indexed pixels。
- `SPRIT` 使用相同的 frame header，但其 frame payload 是完整 packed pixels，不應套用 XOR。
- 每幀的 delay、dimensions、x/y、reserved byte 與 metadata 仍照共用 header 解析。

## 本輪成果

- `gfx.ParseAnimationWithDelta(..., xorFromFirst)` 提供明確 bounded API；原本 `ParseAnimation` 預設 false，保持 SPRIT 正確。
- `scripts/render_previews/` 對 PIC1–PIC6 使用 delta mode，對 SPRIT1–SPRIT6 使用 full-frame mode。
- 從本地原始映像重新產生 152 張 PIC frame PNG，並納入 animation manifest／sprite sheet 證據。
- regression test 覆蓋第二幀 XOR 第一幀後的 indexed pixel 還原。

## 邊界

本輪尚未載入 FINAL member（原始 ZIP 沒有 FINAL.DAX），也尚未將 PIC animation 對應到完整 ECL story／combat event；PNG 仍是由本地原始映像重建的衍生素材。

## 驗證

```sh
go test ./...
go run ./scripts
```
