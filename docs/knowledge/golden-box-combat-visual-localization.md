# Gold Box 戰鬥動畫訊息的本地化邊界

戰鬥動畫通常同時存在三種狀態：規則結果、動畫 phase 與玩家可見文字。重製時
不可讓 renderer 重新推論規則，也不可在規則核心內決定字型與位置。

```text
戰鬥規則結果（damage／save／death）
             ↓
VisualEvent + ordered impacts
             ↓
作品 adapter：phase + locale message ID
             ↓
renderer：frame、sprite、座標、顏色
```

實作原則：

- `impact`、`commit`、`death` 必須使用 typed phase，不以經過毫秒數猜文字。
- 動態目標名稱從 stable fighter ID 解析目前 locale 顯示值。
- protected、找不到 impact 或非適用 phase 時保留既有訊息，避免動畫覆蓋規則結果。
- format string 放 locale catalog；damage、duration 與 saved flag 留在事件資料。
- 搬移文字不得改變 travel→impact→commit→death→handoff 次序，也不能省略聲音。

其他 SSI RPG 可沿用這個分層方法；effect 名稱、phase 意義、動畫素材與時間仍須
由各作品 runtime／影片／反組譯證據確認。
