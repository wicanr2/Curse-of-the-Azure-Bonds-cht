# Gold Box 玩家操作介面的資料化邊界

存檔、讀檔、重新命名與劇情文字輸入看似只是按鍵提示，但它們同時跨越平台
輸入、遊戲狀態、檔案系統與翻譯。若直接在 renderer 組合文字，後續增加語系、
更換前端或調整 JSON 時會出現多份真相來源。

建議分層：

```text
平台輸入（F5／F9／Enter／Backspace）
                ↓
typed operation／result／editor state
                ↓
作品 adapter：stable locale ID + runtime arguments
                ↓
renderer：字型、座標、顏色、裁切
```

實作時應遵守：

- 錯誤物件與檔案路徑是動態參數，不是翻譯字串的一部分。
- 同一 editor 在地城、事件或不同解析度出現時，共用文字 contract，只交換 layout。
- 存檔保存 stable identity 與 typed state，不保存目前語系解析出的全文。
- 測試從正式 catalog 解析期望文字；不要把 JSON 當時的中文複製到 Go 測試。
- fallback 只用於缺 catalog 的 embedders／診斷，正式 catalog coverage 必須逐鍵驗證。
- 資料化不能改變 continuation、輸入長度、確認／取消語意或存讀檔副作用。

這套方法可沿用於其他 Gold Box、冬之魔與後續《Wasteland》中文化；但按鍵、
存檔格式、字數限制與 editor 生命週期仍須由各作品 executable／runtime 證明。
