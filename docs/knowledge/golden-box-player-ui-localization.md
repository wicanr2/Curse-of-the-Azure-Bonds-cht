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

## 冒險 chrome 與地圖

同一句「繼續」可能出現在純文字事件、HEAD／BODY 人物、一般 PIC 與 BIGPIC，
但這不代表 renderer 應保存四份譯文。應共用 stable identity，由各 layout 選擇
自己的座標、字型與顏色。只有文字內容本身不同，例如 BIGPIC 把畫面類型與操作
合成一句時，才建立另一個 ID。

世界地圖另有一個常見陷阱：renderer 直接以固定 locale 查 game-pack 地名。
這會讓主 State 已切換語系、地圖卻仍顯示舊語系。正確做法是讓 State catalog
持有語系身分，所有 game-pack overlay 使用同一語系查詢。若目的地沒有翻譯，
才保留既有 State location fallback；不要在 renderer 補一份作品地名表。

角色欄標題、倒地標記與缺素材診斷也屬玩家可見文字。即使缺素材只在異常狀態
出現，仍應通過正式 locale coverage；診斷狀態不能成為前端硬編碼的豁免區。

## 地城門鎖結果

門鎖流程會同時碰到門面 detail、角色能力、一次性撬鎖機會、Knock spell slot、
撞門結果與雙側 geometry mutation。資料化時不可把這些規則搬成 locale key
判斷，也不可讓 renderer 根據中文訊息反推門是否打開。

正確方向是：規則先產生 typed result，作品 adapter 再選 stable message ID，
renderer 最後只顯示 State 解析出的文字。動態錯誤與 spell ID 是 format argument；
門旗標、法術消耗及 geometry mutation 則永遠留在 typed state。這個分層方法可
沿用其他 SSI RPG，但 `flags=2／3`、spell ID 與雙側牆面表示法仍是作品證據，
未經第二款遊戲驗證前不能抽成 Golden Box 共用格式。

## 研究預覽器與診斷

素材 atlas、AREA、GEO 與地城 preview 常被視為「開發工具」，因而留下大量
硬編碼文字。這會讓研究截圖無法切換語系，也容易把數值格式、錯誤與 renderer
綁死。診斷 UI 應遵守和正式遊戲相同的 stable ID／format argument 原則，但要
明確標為 preview，不得讓玩家誤認為原版畫面。

動態 selector 必須保留來源型別。CoAB `LOAD PIECES` selector 是 `uint16`；即使
已知值目前都很小，也不能為了配合顯示 helper 改成 `uint8`。另外，已翻譯的時間
全文不是資料 API：世界地圖若只需日期，應從 typed clock 欄位重新格式化，而非
依全形空白或中文字切割 `GameTimeText`。

門選項等組合式診斷應由 typed booleans 選完整 stable ID，不在 renderer 拼接
「撬鎖」「Knock」「離開」等局部片段。這能讓每個語系自行決定順序與標點。

## 自動路徑不能比較譯文

截圖工具、smoke test 與 preview automation 若用「等待」「攻擊」等目前譯文
尋找選項，翻譯審校就會改變控制流。正確身分是原始 ECL option token 或正式
stable option ID；顯示文字只用於畫面。

劇情訊息也不能在 Go 保存一小段中文再 `Contains`。若 runtime 尚未公開目前
message ID，至少要讓 adapter 以 stable game-pack message ID 在當下語系解析
全文，再比對目前訊息。如此修改 JSON 譯文時，producer 與 verifier 會讀同一份
資料。長期更理想的 contract 是 State 直接保存目前 message ID，顯示全文只是
projection；在該 identity 完成續存前，不可假稱所有訊息已完全脫離文字比對。

Demo／fixture 名稱也屬顯示資料。Technical fighter ID 可固定在程式，名稱則由
stable locale ID 取得；否則 README 截圖或跨語系 smoke test 會悄悄保存繁中
作為戰鬥身分。
