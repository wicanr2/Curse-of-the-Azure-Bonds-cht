# 第七十二輪：戰鬥小人 sprite extraction

狀態：`READY`（限 CPIC masked picture extraction）

## 本輪成果

- 由原始 `CPIC1.DAX`–`CPIC6.DAX` 解析出 156 張 masked 4-bit sprite PNG，放在 [`assets/sprites/`](../../assets/sprites/)。
- `maskColor=0` 轉為透明，保留 EGA16 原始色盤；每張輸出依 source／block／item 命名。
- [`assets/sprites/README.md`](../../assets/sprites/README.md) 提供完整 manifest 與可重現來源。
- [`combat-sprites.png`](../screenshots/combat-sprites.png) 是方便瀏覽的 sprite sheet；其中可看到熟悉的戰士、法師、怪物與大型生物小人。
- `COMSPR.DAX` 同步抽出作為 24×24 combat effect/icon 參考。

## 反組譯邊界

reference `ovr034.chead_cbody_comspr_icon` 明確將 `CPIC`、`COMSPR` 與 `CHEAD/CBODY` 分成不同載入路徑；因此本輪不把 `SPRIT*.DAX` 的動畫資料誤當成 CPIC 小人。`SPRIT*.DAX` 的 custom animation codec 與 Ebiten combat renderer 仍待實作。

原始 ZIP 仍受 `.gitignore` 保護；repo 只提交使用者要求的 derived PNG、manifest、preview 與 extractor。

## 驗證

```sh
go run ./scripts
go test ./...
```
