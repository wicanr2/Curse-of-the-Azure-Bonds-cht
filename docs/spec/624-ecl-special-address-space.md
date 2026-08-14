# 第六百二十四輪：ECL 特殊位址空間 `7C00h+` 的讀寫對照表

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-07:08BDh`（`CHECKSPECIALS`，875 bytes）、
`overlay-07:0C28h`（`STORESPECIALS`，515 bytes）。

ECL 位址 `7C00h` 以上不是真的記憶體，而是**接到角色紀錄欄位與全域狀態的偽變
數**。兩支函式各自是一長串比對鏈：讀走 `CHECKSPECIALS`，寫走 `STORESPECIALS`。

兩支都先做 `addr := addr − 7C00h` 再比對，所以下表的「偏移」欄就是鏈上的比較值。

## 讀（`CHECKSPECIALS`）

`arg_0` 是一個 **byte 旗標指標**：進函式先設 `1`，鏈全部落空才在最後設 `0`。
呼叫端靠它分辨「這是特殊位址」與「這是普通記憶體」。回傳值在 `AX`。

角色紀錄一律取自 `DS:9594h` 指的那一筆。

| ECL 位址 | 來源 | 寬度 |
|---|---|---|
| `7C15h` | `+13h` | byte |
| `7C18h` | `+19h` | byte |
| `7C72h` `7C73h` | `+74h` `+75h` | byte，`cbw`（**有號**）|
| `7C9Bh` | `+0E0h` | byte |
| `7CA0h` | `+0E5h` | byte，有號 |
| **`7CA5h`..`7CACh`** | **`+0E9h + (addr − 0A4h)`** | byte（**區間**，8 個連號）|
| `7CB8h` | `+0F7h`（劇情 NPC 旗標，見 [spec 623](623-killthedude-damage-message.md)）| byte |
| `7CBBh` `7CBDh` `7CBFh` `7CC1h` `7CC3h` | `+0FBh` `+0FDh` `+0FFh` `+101h` `+103h`（**五種硬幣**，見 [spec 622](622-character-money-block.md)）| word |
| `7CC9h` | `<overlay-25 entry#10>() × +116h ＋ +10Eh`（spec 1040 補上公式） | byte，有號 |
| `7CD6h` `7CD8h` | `+119h` `+11Bh` | byte，有號 |
| `7CE4h` | `+193h` **`and 1`**（只取最低位元，spec 1040 補） | byte |
| `7CF7h` `7CF9h` | `+13Ch` `+13Eh` | word／byte |
| `7D1Bh` | `+1A6h` | byte |
| `7ECFh` | `+1Bh`（魅力）**再查一張十二段換算表**，見 spec 1040 | byte |
| `7F12h` | `DS:8BF4h` ＝ **資料片編號**（spec 1040 由 DOS 對側關掉） | byte |
| `7F3Eh` | bank 1（`DS:7F09h`）`+67Ch` | word |

五個硬幣在讀寫兩側都是**五個連號位址**，這是 spec 622「`+FBh` 起是陣列」的獨立
佐證——那邊的證據是迴圈索引算式，這邊是位址編號本身。

### 三個要算的

```text
7D00h:  if char^[197h] <> 0 then 1 else 80h
        if DS:0BDFFh = 1 then 結果 := 0
        DS:0BDFFh := 0                        ← 讀一次就清掉

7D0Ch:  if char^[198h] = 0 and char^[199h] <> 0 then 80h
        else if char^[198h] = 1               then 81h
        else                                       0

7EB1h、7EB4h: 兩者 body 完全相同，都是 <near 085Ch>(DS:9594h) 再零延伸
```

`DS:0BDFFh` 是**一次性旗標**：讀 `7D00h` 會順手清掉它，所以連讀兩次結果不同。

### `7D0Dh` 讀出未初始化的值

```asm
cmp  ax, 10Dh
jnz  short loc_B11
jmp  loc_C1F          ; ← 直接跳到出口
```

命中 `7D0Dh` 什麼都不做就跳到 `mov ax, [bp+var_2]`，而 `var_2` **在這支函式裡從
未初始化**。回傳的是堆疊殘值。旗標仍然是 `1`（表示「這是特殊位址」），呼叫端不會
察覺。這是原作行為，remake 要照抄。

## 寫（`STORESPECIALS`）

`arg_0` 是要寫入的值，`arg_2` 是位址。**寫的表比讀的小**——讀有 27 個入口，寫只有
18 個。多數欄位是唯讀的（延續 [spec 565](565-ecl-memory-read-path-and-asymmetry.md)
記錄的讀寫不對稱）。

| ECL 位址 | 動作 |
|---|---|
| `7C00h` | **只在值為 0 時**設 `DS:0BDE5h := 1`；非 0 什麼都不做 |
| **`7C20h`..`7C70h`** | `char^[1Eh + (addr − 1Fh)] := value`（byte，**81 個連號位址**）|
| `7CB8h` | `if value > 0B2h then value := value − 32h`；再寫 `+0F7h` |
| `7CBBh` `7CBDh` `7CBFh` `7CC1h` `7CC3h` | 五種硬幣，word |
| `7CF7h` `7CF9h` | `+13Ch`（word）、`+13Eh`（byte）|
| `7D00h` | `if value >= 80h then +197h := 0`；`if value = 87h then +196h := 7`；`if value = 0 then DS:0BDE4h := 1` |
| `7D0Ch` | `value=0` → `+198h,+199h := 0,0`；`80h` → `0,1`；`81h` → `1,1`；其他值不動 |
| `7F12h` | `<far 0723h:0A85h>(byte(value))` |
| `7F22h` `7F24h` `7F26h` | **值 > 80h 才動作**：`value := value and 7Fh`，再 `<sub_1808>(1／2／3, value)` |

`7F22h`／`7F24h`／`7F26h` 三者除了傳給 `sub_1808` 的第一個參數（1、2、3）以外
完全相同——是同一個動作的三個編號。值 `<= 80h` 時**靜默不做事**。

## 明確不宣稱

- 各欄位在遊戲規則上的名稱（哪個是力量、哪個是護甲）。只確定位址對到哪個
  offset、寬度與有號性。
- `DS:0BDE4h`／`DS:0BDE5h`／`DS:0BDFFh` 三個旗標由誰消費。
- `+196h`／`+197h`／`+198h`／`+199h` 的完整狀態列舉。
- `7CC9h` 為什麼碰兩個欄位（`+116h` 與 `+10Eh`）——需要另外讀那段的算式。
