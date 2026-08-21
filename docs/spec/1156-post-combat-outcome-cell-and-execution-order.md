# 1156 — `7EC7h` 是戰後結果碼，以及 remake 拿 PC 當執行順序的那個代理

- 證據等級：`exact`（`bank1^[58Eh]` 的四處寫入逐處讀出；`ECL2/0x04` 的
  三處座標寫入逐處列出）
- 分歧的量測見 [spec 1155](1155-redraw-call-coordinate-divergence.md)
- 換算依 [spec 1096](1096-ecl-bank1-address-formula.md)：`bank1 + (位址 − 7C00h) × 2`

## `7EC7h` ＝ `bank1^[58Eh]`：戰後結果碼

寫入端全部在 `overlay-05`（POSTCOM），四處：

| 位址 | 寫入值 | 所在函式 |
|---|---:|---|
| `0817h` | **`81h`** | `sub_53C`（走隊伍鏈修 `+195h`／`+196h`／`+1A4h`）|
| `0B52h` | `80h` | `sub_A7E` |
| `173Fh` | `0` | `sub_1736` |
| `185Dh` | `80h` | `sub_1736` |

腳本側一律是 `COMPARE 7EC7 80h / IF >`，**只有 `81h` 會成立**。三份既有規格
各自從別的方向用過這道閘：

- [spec 303](303-fire-knife-leader-bond.md)：`[7EC7] <= 80h` 進入**勝利線**。
- [spec 1145](1145-clearmonsters-drops-the-treasure-pile.md)：`> 80h` 跳回
  火刀首領的**重打迴圈**。
- [spec 318](318-wizard-tower-black-dragons-heart.md)：戰後 `> 80h` 回到前一步，
  並刻意不把它猜成士氣或逃跑。

⇒ **`81h` ＝ 這一場沒有打贏**（三處用法一致：重打、退回、不走勝利線）。
`sub_53C` 在寫 `81h` 之後才走隊伍鏈把角色狀態修回來，形狀上是「收拾殘局」。

⚠ 仍不宣稱 `80h` 與 `0` 的差別，也不宣稱 `81h` 具體是逃跑還是全滅
——三處呼叫端只分「> 80h」與「<= 80h」兩邊。

## `ECL2/0x04` 的三處座標寫入（全部列出）

| 位移 | 指令 | 情境 |
|---|---|---|
| `0060h` | `ADD 02 C04B C04B`；`SAVE 0F C04C` | 離開巢穴回提爾佛頓，之後接 `NEWECL 03` |
| `02D7h` | `SAVE 02 C04B`；`SAVE 0C C04C` | 「YOU ARE BEING TAKEN TO THEIR LEADER.」被押到 `(2, 12)` |
| `160Ch` | `SAVE 4BF0 C04B`；`SAVE 4BF1 C04C` | 戰後 `7EC7 > 80h` ⇒ 退回上一格 |

★ `02D7h` 那一處是**沒有 `C04D` 的真實移動**，remake 現在整個跳過
——玩家看得見：被押去見首領時人不會被移過去。

## ★★★ remake 拿 PC 當執行順序的代理

`projectFreshDungeonCoordinatesBeforeCall` 原本用
`write.BlockID != call.BlockID || write.PC >= call.PC` 篩「這一條寫入在
`CALL` 之前嗎」。

⚠ **PC 的大小不是執行順序。** 一次執行裡有迴圈與反向跳躍：`CALL` 之後腳本
跳回較小的 PC 再寫一輪座標時，那一輪的 `PC` 仍然小於 `CALL` 的 `PC`，於是被
當成「先發生」算進來。同一個位址在一次執行裡被寫好幾次也是常態
（`ECL2/0x01` 的 `C04B` 就有 6 處寫入）。

修法是讓 VM 記**執行序**：`MemoryWrite.Sequence` 與 `CallRequest.Sequence`
共用同一條計數，`BlockSession` 聚合多段 sub-run 時再加上前面幾段的總量。

★ 那條計數是共用的，所以序號會跳號——`ECL4` 死精靈那一段的
`SaveWrites` 是 1、2、3、**5**，中間的 4 是那條 `CALL`。這一點本身就是
「兩份清單真的在同一條時間軸上」的證據，已經釘進測試。

⚠ 這個修法**單靠既有測試抓不到**：第一版的回歸測試把「後執行」那一輪放在
比 `CALL` 小的 PC，但**兩種比法給的答案相同**，所以改回 PC 版仍然全綠。
要區分開，那一輪必須排在 `CALL` **之後**（執行序大）而 PC **仍然小於** `CALL`
——也就是真正的反向跳躍。

## remake

- VM 記執行序，adapter 改用執行序篩選。行為在真實 corpus 上沒有改變
  （直線碼上兩種比法一致），但那個代理不再是隱形的假設。
- `C04D` 條件仍在。它擋的是 spec 1155 量到的 23 處，而其中
  `02D7h` 這種「沒有 `C04D` 的真實移動」是要接的，`160Ch` 那種
  「戰後退回」則要等 `7EC7h` 有 producer 才會正確觸發。

## 明確不宣稱

- 沒有宣稱 `80h` 與 `0` 兩個值的差別。
- 沒有宣稱 `sub_53C`／`sub_A7E`／`sub_1736` 各自是 POSTCOM 的哪一段。
- 沒有宣稱 remake 該怎麼產生 `7EC7h`——本輪只指出它目前沒有 producer。
- 沒有宣稱 spec 1155 那 23 處裡除了 `ECL2/0x04` 以外各是什麼情境。
