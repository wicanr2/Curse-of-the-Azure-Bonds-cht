# 第七十五輪：SPRIT runtime playback

狀態：`READY`（限目前 combat renderer 的 SPRIT playback）

## 本輪成果

- ECL `SETUP MONSTER` 的 `SpriteID` 會保存為 fighter 的 `AnimationBlock`，不再只保留 CPIC icon block。
- renderer 載入 `assets/sprites/animation.json` 與逐幀 PNG，依原始 delay 單位（每單位 0.1 秒）循環選擇 frame。
- CPIC 與 SPRIT 來源欄位分離：CPIC 負責靜態／攻擊 icon，SPRIT 負責 encounter animation。
- 若動畫 block 不存在，仍會退回 CPIC 或 deterministic static sprite，保持戰鬥畫面可用。

## 邊界

目前沒有實作 reference 的方向 flip、PIC/FINAL XOR delta 或完整八方向 combat placement；第 79 輪已補上 SPRIT frame position offset，CHEAD/CBODY merge 則見第 76 輪。

## 驗證

```sh
go test ./...
go run ./scripts
```
