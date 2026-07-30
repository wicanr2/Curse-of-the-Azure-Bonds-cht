# 第 386 輪：Burial Glen 紅網與 ECL `INPUT STRING`

狀態：`READY`（限可續跑字串輸入、terrain `82h` 四分支與兩戰
continuation；原版逐鍵輸入細節及完整 Burial Glen 仍未完成）

## 1. 原始資料證據

DOS `ECL6.DAX` block `40h` 的 terrain `82h` dispatcher 位於 decoded payload
`+0768h`：

- `+0770h`／`+0791h` 顯示紅網橫跨通道並散發黯淡光芒。
- `+07A7h` 的 HORIZONTAL MENU 將四項寫入 `7F79h`：
  `ENTER IT`、`SPEAK`、`HACK IT`、`RETREAT`。
- `SPEAK` 於 `+0895h` 詢問要說什麼；`+08A8h` 是
  `INPUT STRING 8, [7F79h]`。接著不做字串比較，無條件顯示
  `THE WEB GLOWS MORE BRIGHTLY.`，等待按鍵後回到 `+0768h`。
- 因此攻略／Journal 25 的 `Krrkik` 是玩家取得的敘事提示，但這個紅網
  branch 並未驗證答案。不得在 remake 另加「只有輸入 Krrkik 才成功」的
  分支，也不得把字串輸入假造成只有一個正確答案的選單。

另一個有效用例交叉證明字串記憶體用途：

- block `40h:+0425h`：`INPUT STRING 12, [7B90h]`。
- `+042Bh` 立即以字串 operand 比較 `[7B90h]` 與 packed
  `TYRANTHRAXUS`，錯誤答案進入敵對分支。
- block `42h:+040Dh` 有同一組 12 字元輸入與比較。

這證明第二個 operand 是可由後續 `COMPARE`／`PRINT` 讀取的字串記憶體
destination，而不是 numeric menu location。

## 2. 紅網分支與戰鬥續跑

`ENTER IT`：

1. `+07E3h` 顯示隊伍被網牢牢黏住。
2. `+07FDh..+080Ch` 載入四隻 monster `42h`、icon block `41h`，開始第一戰。
3. 戰後從 `+080Dh` 續跑；正常精靈幽魂路徑已設定 `4CBEh`，所以顯示
   PICTURE 72 幽魂大笑並揭露 rakshasa。
4. `+085Eh..+086Dh` 載入一隻 monster `43h`、icon block `43h`，開始第二戰。
5. 第二戰後顯示隊伍終於脫困，並把 `4CBFh` 設為 1；再次踏入此 terrain
   會直接退出事件。

`HACK IT`：

- 紅光消退並露出金屬套索；等待按鍵後顯示蜘蛛循聲而來。
- 載入四隻 monster `42h` 開戰，沒有 rakshasa 揭露分支。

`RETREAT` 直接離開，沒有戰鬥。上述 monster ID 是作品 ECL 證據，不可搬入
作品中立 frontend。

## 3. Remake contract

- `RunResult` 新增 `WaitingForString`、`StringInputRequests` 與消耗計數。
- `StringInputRequest` 保存 `MaxLength`、`Destination`、提交值與是否已提供。
- `BlockSession` 與 menu／WHO 相同，保存獨立的累積 input offset；暫停時
  PC 留在 `INPUT STRING`，提交後只消耗新值一次，不重播前置文字或旗標。
- runtime 依 rune 截到 ECL 最大長度，並轉成大寫以對應已驗證的 uppercase
  ECL literal vocabulary。這是 remake 輸入便利層；原 DOS 是否在每個鍵入
  路徑立即轉大寫，仍待 overlay handler／DOSBox 逐鍵 trace。
- State 與 Ebiten 提供 Unicode 輸入、Backspace 與 Enter；原作 prompt、
  密語及後續結果仍由 ECL／CoAB text rules 驅動。
- text-only dungeon input 仍使用原版 cracked-stone adventure frame、
  640×480 canvas、nearest-neighbour 第一人稱素材與 16×15 倚天粗體，不退回
  generic 全螢幕文字頁。

## 4. 驗證

- synthetic VM regression：暫停於 opcode、destination／length、resume、
  uppercase、rune truncation、字串記憶體與後續 PRINT。
- real-image regression：紅網四選單、任意 8 字元輸入的回返、蜘蛛第一戰、
  rakshasa 第二戰、`4CBFh` completion、HACK 與 RETREAT。
- normal player path：Standing Stone→Myth Drannor→Burial Glen `(2,15,E)`，
  經 GEO `(3,15)→(3,14)` 精靈幽魂，再沿可通行
  `(4,14)→(5,14)→(6,14)` 抵達 terrain `82h`，輸入 `Krrkik` 後返回紅網。
- deterministic 640×480 畫面：
  `docs/screenshots/burial-glen-red-web-input.png`。

## 5. 尚未完成

- DOS／PC-98 `INPUT STRING` 編輯鍵、游標、空字串、ASCII case conversion
  與畫面逐像素核對；指定 IDA Pro overlay handler 尚未定位。
- 兩場戰鬥的 State 勝利續跑與「great strength」負面證據已由
  [第 387 輪規格](./387-burial-glen-red-web-two-battle-state.md)補完；
  戰敗路徑與完整蜘蛛／羅剎妖規則仍未完成。
- Burial Glen 其餘 terrain、Tyranthraxus 最終遭遇與結局。
