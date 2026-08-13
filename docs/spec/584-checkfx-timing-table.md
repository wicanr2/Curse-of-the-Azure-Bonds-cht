# 第五百八十四輪：`CHECKFX` 是一張時機分派表

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98／DOS 都在 `overlay-23:03FEh`（2,287／2,313 bytes，6 個呼叫者）。

## 它不是一支做很多事的大函式

`CHECKFX(timing, subject)` 的 2,287 bytes 裡**沒有任何算術、沒有任何記憶體
存取**，只有：

| 助憶碼 | 條數 | 用途 |
|---|---:|---|
| `cmp al, N` | 24 | 24 個 timing 的分派 |
| `mov al, N` | 161 | 要檢查的 effect id |
| `call` | 161 | 交給 `sub_269`（effect 鏈遍歷） |
| `push` | 645 | 參數 |
| `jz`／`jnz`／`jmp` | 53 | 分支 |
| prologue／epilogue | 4 | — |

**161 個 `mov al, N` 對上 161 個 `call`，也對上解析出來的 161 個 effect id
——三個數字相等，這就是「整支讀完了」的證明。** 兩平台指令數 1,050 對 1,051
（DOS 多一條 `xor`）。

所以規則各處的 `CHECKFX(06h)`／`CHECKFX(09h)`／`CHECKFX(0Ah)`／`CHECKFX(0Ch)`
問的都是同一件事：**「這個時機有哪些效果要介入？」**

完整的表：[`../audit/checkfx-timing-table.md`](../audit/checkfx-timing-table.md)。

## 已知的呼叫時機

| timing | 呼叫處 | effect 數 |
|---:|---|---:|
| `06h` | `PUTDAMAGE` 進入時 | 21 |
| `09h` | `PUTEFFECT` | 12 |
| `0Ah` | `ATTEMPTTOHIT` 對目標 | 10 |
| `0Ch` | `MAKESAVE` | 16 |
| `0Dh` | `KILLDUDE`／`PUTDAMAGE` 死亡後 | 3 |
| `10h` | `ATTEMPTTOHIT` 對攻擊者 | 7 |
| `14h` | `PUTDAMAGE` 無豁免時 | 2 |

`00h` 這個 timing **什麼都不做**（直接跳到 epilogue）——呼叫端傳 0 是合法的
no-op，不是錯誤。

## 對 remake 的意義

effect 系統的介入點是**資料**不是程式碼：24 個時機各自帶一張 effect id 清單。
要正確重現原版，這 161 筆對應必須照抄，不能靠「哪些效果聽起來該在命中時生效」
去猜。例如 `ATTEMPTTOHIT` 對目標檢查的是
`01h, 02h, 21h, 24h, 31h, 06h, 12h, 1Ah, 4Bh, 4Ch` 這 10 個，多一個少一個都
會改變命中率。

## 怎麼解出來的

`scripts/checkfx_timing_table.py` 解析 IDAPython 匯出的逐指令序列。判準純結構：

```text
cmp al, N / jnz NEXT                  → body 在後，下一個 case 在 NEXT
cmp al, N / jz BODY / jmp NEXT        → body 在 BODY，下一個 case 在 NEXT
```

**兩種形狀都要處理。** body 超過 short jump 的 127 bytes 時編譯器就換成第二
種；只認第一種會在第一個大 case（這裡是 `05h`）停住，24 個 case 只解出 6 個
——而且**不會有任何錯誤訊息**，看起來像「這函式只有 6 個 case」。

## 明確不宣稱

- 161 個 effect id 各自是什麼效果。
- 24 個 timing 中還沒找到呼叫點的那 17 個分別在哪裡被呼叫。
