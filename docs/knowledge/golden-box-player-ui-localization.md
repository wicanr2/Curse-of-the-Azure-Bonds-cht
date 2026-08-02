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

## 中文 fallback 仍是重複真相

呼叫 `catalog.Text("key", "目前中文")` 並不等於完成資料分離：第二個參數仍是
會隨 JSON 審校而漂移的另一份真相來源。正式玩家流程應以 stable ID 作缺鍵
fallback，並由 catalog coverage 測試保證正式資料完整；缺鍵時顯示 ID 也比
悄悄顯示過期中文更容易被驗收發現。

產品行為測試應載入正式 game pack／catalog，再於執行當下由 stable ID 取得
期望文字。最小合成 catalog 只適合 parser 或缺鍵結構測試，不能拿來證明正式
世界地名、NPC 名稱與選單內容。控制流仍應使用原始 ECL token、stable option
ID 或 typed result；即使顯示 choice 被換成任意文字，行為也不應改變。

## 財寶服務邊界

Gold Box 的 TREASURE 不只是顯示一句文字：它可能先累積 coins／gems／jewelry，
解析固定或隨機 ITEM block，再於戰後或零怪物服務 boundary 暫停 ECL。資料化
提示時，不能把「看到財寶 prompt」誤當成財寶規則完成，也不能因翻譯調整而
改變戰後延遲、take／cancel／skip、角色裝備 mutation 或 continuation PC。

顯示選項與控制 identity 必須平行保存。物品列可顯示 locale item name，但選擇
應使用 `TREASURE_ITEM_n`；角色列可顯示玩家姓名，但選擇應使用
`TREASURE_CHARACTER_n`。缺 ITEM 素材時，錯誤物件只是格式參數；正式訊息
模板放 catalog。正常玩家路徑要同時驗證原始財寶數量、列表 identity、收取或
略過後的回返模式，以及再次搜索不複製獎勵。

## 遭遇談判策略

PARLAY 的「傲慢、狡猾、謙卑、友善、威嚇」是顯示投影，不是 branch key。
ECL 或作品 adapter 應保存來源 tactic identity；locale 只負責名稱與提示。
測試若以中文陣列找索引，改譯、同義詞或語序調整就會改變劇情分支。

驗收至少要把入口顯示 choice 換成無關文字，證明 `PARLAY` identity 仍能開啟
策略選單；再於一條真實 ECL 玩家路徑同時斷言 tactic identity 與當前 locale
投影，最後追到劇情文字、Journal／旗標、戰鬥或返回地圖。Generic「對方反應
尚待 script」只可用於沒有 ECL continuation 的暫時 encounter，不能掩蓋作品
原本存在的 tactic branch。

## 戰鬥訊息與 typed result

戰鬥文字往往同時含角色名、目標名、座標、傷害、命中次數、豁免與元素防護。
正確邊界是規則先產生 typed result，再由 adapter 選 message ID 與參數；locale
只決定語序與措辭。不可從「未受傷」「抵消」等中文反推 protected flag，也不可
因多數法術共用一句格式就省略各自的 visual／sound／save 流程。

未知法術名稱應走共用 spell label resolver；戰鬥、紮營、商店各自保存
「法術 0xNN」會形成三份 fallback。測試要載入正式 catalog，使用實際 damage、
fighter name 與 typed result 動態組期望全文。資料化文字的低風險 milestone 可
抽樣既有近戰、移動、法術、快速戰鬥與勝敗路徑；但任何規則、PRNG、timeline、
聲音、ECL continuation 或 renderer 變更仍須提升為完整相關 gate。
