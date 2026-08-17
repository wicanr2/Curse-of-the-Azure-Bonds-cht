# 1121 — 原版存檔匯入：版面相同、編碼不同，所以要兩條路

- 證據等級：`exact`（版面取自 spec 1072／1076／1115；編碼分流的必要性由測試
  直接示範，含正對照）
- 對應工作項：`ENG-10`（存檔完整實作）與 `CHT-10`（原版位元組的解碼邊界）

## 問題不在版面，在來源

`CHRDAT?{1..6}.sav`／`.swg` 的版面**同時**被兩種東西使用：

| 來源 | 名字的編碼 | 誰寫的 |
|---|---|---|
| 原版存檔 | Big5（中文版）／ASCII（英文版） | 原作 |
| remake 自己的槽 | UTF-8 | `SaveSAVGAMSlot` → `PatchDOSPlayerRecord` |

`SaveSAVGAMSlot` 保留原始位元組再覆寫已知欄位，所以一個槽可能是「原版的骨架
＋ remake 的名字」。**光看檔案分不出來源**——沒有版本欄位，沒有魔數，名字那幾
個位元組在兩種編碼下都是合法的。

## 猜錯的代價不對稱

| 走錯的方向 | 症狀 |
|---|---|
| 把 remake 的槽當原版讀 | UTF-8 位元組被當 Big5 解 → 亂碼 |
| 把原版的槽當 remake 讀 | Big5 位元組原樣塞進 UTF-8 字串 → 亂碼 |
| **英文原版，兩個方向** | **完全一樣**（`origtext.Decode` 對純 ASCII 直接回傳） |

第三列是這件事最容易被漏掉的地方：**拿英文樣本測分流等於沒測**。所以
`TestBothPathsAgreeOnASCIINames` 是正對照，證明兩條路在 ASCII 上一致；
另外兩條測試才用中文樣本證明它們真的不同。

## 作法：由呼叫端明講

| 入口 | 用途 |
|---|---|
| `party.ParseDOSPlayerFiles` ／ `State.LoadSAVGAMSlot` | remake 自己的槽 |
| `party.ParseOriginalDOSPlayerFiles` ／ `State.LoadOriginalSAVGAMSlot` | 原版存檔 |
| `monster.ParseItems` ／ `ParseOriginalItems` | 同理，`.SWG` 的物品名 |
| `party.ParseDOSPlayerRecord` ／ `ParseOriginalDOSPlayerRecord` | 角色名 |

`.FX`（效果記錄）沒有字串，兩條路共用。

CLI 這一側是 `-savgam-import`：**匯入之後不寫回同一個槽**。原版檔案是唯讀輸入
（決策四：remake 讀舊版 DAT、存成自己的格式，不必互通），寫回去只會把原版資料
換成 UTF-8 的混合體。

## `ITEM*.DAX` 也走同一條線

`ParseTreasureItemBlocks` 與兩個 `cmd/` 的 ITEM 區塊載入都改走
`ParseOriginalItems`——那些位元組全部來自原版。

`TestEveryDAXItemLoadUsesTheOriginalDecoder` 掃 `internal/` 與 `cmd/` 的原始碼，
只要有人拿 DAX 區塊去餵 `monster.ParseItems` 就當場紅，並附一條正對照確認掃描
真的看得到檔案。**這種錯不會 panic 也不會被既有測試碰到**，它只是把中文名字讀成
亂碼，所以只能用「每次都把載入點列出來比對」的方式守。

`ParseBaseItems`（ITEMS 成員）不受影響：那筆記錄裡沒有字串。

## 為什麼不自動偵測

「無效 UTF-8 就當 Big5」這種啟發式在這裡剛好最危險：Big5 的雙位元組序列有相當
比例**同時是合法的 UTF-8**，反過來也一樣。偵測會在大部分名字上答對、在少數名
字上安靜地答錯，而錯的那幾個沒有任何訊號。旗標會被使用者忘記，但忘記的後果是
一眼看得見的亂碼——這比安靜的部分錯誤好。

## 明確不宣稱

- 沒有宣稱原版中文版用的一定是 Big5 全字集；`origtext.Decode` 解不出來時原樣
  回傳位元組，不假裝解碼成功。
- 沒有宣稱 remake 寫出來的槽能被原版遊戲讀回去——決策四已經移除這個要求。
