# 1097 — ECL↔引擎共用格子清冊，與 remake 現行用法的逐格比對

- 狀態：`READY`
- 證據等級：`exact`（映射依 spec 1096；引擎側存取點由 `scripts/ecl_shared_cells.py`
  對 DOS 全 36 個 overlay 的 `filled` 匯出做保守符號執行掃出，可重生）
- 產物：[`../audit/ecl-shared-cells.md`](../audit/ecl-shared-cells.md)
- 對應 TODO `RE-14`

## 方法

1. ECL 側：`ecl-event-catalog.json` 的 1,355 條指令 operand → 依 spec 1096
   換算 bank 內位移 `(位址 − 區基底) × 2`。
2. 引擎側：對每個函式追蹤 `les di, ds:4F99h／4F9Dh／4FA1h` 把 `di` 綁到哪一塊 bank，
   之後的 `es:[di+XXXXh]` 歸屬該 bank 的該位移。
   **`di` 一旦被無法靜態求值的指令改動就停止歸屬**——寧可漏報，不要報錯。
3. 兩邊取交集。

⚠ 因為第 2 步是保守的，「引擎側查無存取」代表**這個方法沒找到**，不代表不存在。

## 結果

| | 數量 | ECL 存取次數 |
|---|---|---|
| ECL 側變數位址 | 81 | 674 |
| **共用格子**（引擎側也存取同一位移） | **24** | 184 |
| ECL 私有（引擎側查無存取） | 57 | 490 |

★ 最熱的三格 `7F79h`／`7F7Ah`／`7F7Bh`（167／42／20 次）落在 ECL 私有那一側
——與 spec 1096 §四「形狀上是通用工作暫存器」一致：**工作暫存器本來就不該被引擎碰**。
這是分區表與掃描結果互相支持的一個旁證。

## ★★★ remake 用到的 20 個位址：逐格比對

| ECL 位址 | bank 位移 | ECL | 引擎 | remake 用法 | 判定 |
|---|---|---|---|---|---|
| `4BF2h` | `bank0^[1E4h]` | 33 | 5 | `previousBlockID`（`state.go:5112`） | ★ **正確**，見下 §一 |
| `7EE1h` | `bank1^[5C2h]` | 30 | 3 | `head`（`runtime.go:1443`） | ★ 形狀相符，見 §二 |
| `7ED2h` | `bank1^[5A4h]` | 23 | 2 | 遭遇週期 | ⚠ 待引擎側佐證 |
| `7ED3h` | `bank1^[5A6h]` | 21 | 2 | 遭遇百分比 | ⚠ 待引擎側佐證 |
| `7ED5h` | `bank1^[5AAh]` | 8 | 2 | 3D 移動迴圈退出狀態 | ⚠ 待引擎側佐證 |
| `7ECAh` | `bank1^[594h]` | 2 | 5 | SearchLocation 期間設 1、結束設 0 | ⚠ **有落差**，見 §三 |
| `7EC9h` | `bank1^[592h]` | 1 | 3 | 一律寫 0 | ★ 形狀相符（計數重置） |
| `4BC6h` | `bank0^[18Ch]` | 2 | — | ECL 時鐘基底 | ECL 私有，自洽即可 |
| `4C02h` | `bank0^[204h]` | 12 | — | 世界路線點 | ECL 私有 |
| `4C9Ch` | `bank0^[338h]` | 5 | — | 世界目的地 | ECL 私有 |
| `7CB8h` | `bank1^[170h]` | 1 | — | 控制士氣 | ECL 私有 |
| `7D00h` | `bank1^[200h]` | 1 | — | 旗標 | ECL 私有 |
| `7F81h` | `bank1^[702h]` | 5 | — | 旗標 | ECL 私有 |
| `7F6Ch` | `bank1^[6D8h]` | **0** | 1 | `ShopRequested` | ★ **正確**，見 §四 |
| `7F6Dh` | `bank1^[6DAh]` | **0** | 2 | `ShopPriceScale` | ★ 位址相符 |
| `7EE2h` | `bank1^[5C4h]` | **0** | 1 | `TempleRequested` | ⚠ **語意不符**，見 §四 |
| `4C9Bh`／`7C00h`／`7D0Ch`／`7FFFh` | — | 0 | 0 | remake 自訂 | 兩側都沒用，安全 |

## ★ 一、`4BF2h` ＝ 上一個 block／scene 編號（remake 正確）

五份既有規格獨立指向同一個語意：

| 出處 | 內容 |
|---|---|
| spec 892 | 新遊戲初始化 `bank0^[1E4h] := 0` |
| spec 590／612 | `bank0^[1E4h] := DS:BDF0h`——**把舊值收起來** |
| spec 613 | `if bank0^[1E4h] = 0 then DS:BDF0h := 1 else DS:BDF0h := bank0^[1E4h]` |
| spec 1045 | 「上次離開時記下的狀態：非 0 就沿用，0 才重新初始化」 |
| spec 1085 | `啟用[8] := ord(bank0^[1E4h] = 0)`——**只有 `= 0` 才給主選單的 `L`（載入）** |
| spec 1092 | `if bank0^[1E4h] <= 1 then 直接進入編輯`（角色數值調整） |

> ★★★ **`bank0^[1E4h]` 存的是「上一個 block／scene 編號」，`DS:BDF0h` 是目前的。**
> ⇒ remake 寫 `previousBlockID` **與原作一致**。
> ★ 這同時解釋了另外兩條原本標為「不宣稱」的規則：
> `= 0`（還沒進過任何 block ＝ 新遊戲）才給載入；
> `<= 1`（還在第一個 block ＝ 剛建完角）才准改數值。
> ⇒ **spec 1045／1085 的「沒有宣稱 `bank0^[1E4h]` 語意」可以解除。**

## 二、`7EE1h` ＝ 目前顯示的圖片（形狀相符）

既有規格（`0Eh PICTURE` 相關）顯示：`bank1^[5C2h] = 0FFh` 走一條路、
`<> 0FFh` 時改呼叫 `0062:0043`（另一種顯示方式），`0Eh` 會把它設成 `0FFh`。
⇒ 形狀是「目前顯示的圖片 ID，`0FFh` ＝ 沒有」。remake 用它當 `head`（頭像）**形狀相符**。
⚠ 本規格不宣稱 remake 的 `head` 與原作的圖片 ID 值域是否一致。

## ⚠ 三、`7ECAh` 的還原方式與原作不同

spec 1045 的地城主迴圈：

```pascal
while (bank1^[594h] > 1) or (UpCase(鍵) = 'E') do begin
    DS:8B5Fh := bank1^[594h] and 1;      { ★ 保存最低位 }
    bank1^[594h] := 1;
    …跑 ECL…
    bank1^[594h] := DS:8B5Fh;            { ★ 還原成 and 1 的值 }
end;
```

spec 1095 的 `COMBAT` 收尾也做同一個動作：`bank1^[594h] := bank1^[594h] and 1`。

> ★★ **`bank1^[594h]` 是「這一格還有沒有事件」的計數**：`> 1` 才進迴圈跑 ECL，
> 進 ECL 前壓成 1，跑完**還原成進入前的值 `and 1`**（也就是 0 或 1）。
>
> ⚠ remake（`state.go:4958`／`5013`）是 `SetMemoryValue(0x7ECA, 1)` ＋ **一律設 0**。
> **進入前若是奇數，原作跑完仍是 1，remake 會變成 0。**
>
> ⚠ 本規格不宣稱這在 remake 現行流程會不會實際造成差異
> ——remake 用它的時機是 `SearchLocation`，與原作這段主迴圈**是否對應同一個時機尚未確認**。
> ⇒ **修正前要先確認時機對應**，不可直接照抄 `and 1`。

## ⚠ 四、`24h` 的三選一：商店正確，第二格語意不符

| remake | 換算 | 原作 `24h` 分支（spec 1095） |
|---|---|---|
| `memory[0x7F6C]` → `ShopRequested` | `bank1^[6D8h]` | **商店** ✓ |
| `memory[0x7F6D]` → `ShopPriceScale` | `bank1^[6DAh]` | 商店那支的相鄰格 ✓ |
| `memory[0x7EE2]` → `TempleRequested` | `bank1^[5C4h]` | **營地（Camp）**，呼叫 `overlay-04 entry#1` |

★ **三個位址都與原作對得上**（remake 不是隨便選的）。
⚠ **但第三格的語意標成 Temple 與原作不符**，它是營地。

★★★ 這三格在 1,355 條 ECL 指令裡**一次都沒被寫過**（已全掃確認）
⇒ 它們是**引擎寫入、`24h` 讀取並清零**的請求旗標。

> ⚠ **這不是單純改個名字就好。** remake 的營地是獨立的引擎功能
> （`State.Camp()`／`EnterDungeonCamp()`），不經 `24h`；
> 原作的營地是**從 ECL 的 `24h` 觸發，結束後回到同一個 ECL 迴圈續跑**（spec 1095）。
> ⇒ 真正缺的是「營地結束後 ECL 續跑」這個語意，不只是命名。
> **本規格只記錄差異，修正屬於 `ENG-07`／營地流程的範圍。**

## 明確不宣稱

- 沒有宣稱 `7ED2h`／`7ED3h`／`7ED5h` 的原作語意（remake 的用法目前只有形狀支持，
  引擎側存取點 `overlay-07:01FC`／`overlay-20:0C9C`／`overlay-14:078E` 尚未逐條讀）。
- 沒有宣稱 57 個「ECL 私有」格子的語意——它們屬於 spec 1096 §五第 1 點的自洽情形。
- 沒有宣稱掃描漏報的範圍（`di` 被動態改動的存取點一律未計入）。
- 沒有宣稱 PC-98 側的共用格子（`scripts/ecl_shared_cells.py` 的 PC-98 bank 指標
  只有 bank0／bank1 有證據，第三塊未確認）。
- 沒有宣稱 remake 的 `head` 與原作圖片 ID 的值域對應。
- 沒有宣稱 remake 的 `SearchLocation` 對應原作哪一段流程。
