# 第七百零一輪：治療／致傷的對稱階梯

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-22` 的 `407Ah`、`3F23h`、`43A7h`、`4103h`、`413Fh`、`43EAh`。

## 六支排在一起就看得出來

| 位址 | 骰子 | 走向 | 訊息 |
|---|---|---|---|
| `407Ah` | `2d4+2` | `HEALDUDE` | `is Healed`（`CS:4070h`） |
| `3F23h` | `2d8+1` | `HEALDUDE` | `is fully/partially healed` |
| `43A7h` | `3d8+3` | `HEALDUDE` | 同上 |
| `4103h` | `2d4+2` | `sub_F06h` | 空字串（`CS:4102h`，長度 0） |
| `413Fh` | `2d8+1` | `sub_F06h` | 空字串（`CS:413Eh`） |
| `43EAh` | `3d8+3` | `sub_F06h` | 空字串（`CS:43E9h`） |

上下三支的骰子完全一樣，差別只在走 `HEALDUDE`（治療）還是 `sub_F06h`（致傷）。
這就是 AD&D 的 Cure／Cause Wounds 兩條平行階梯。

治療側用 `ROLLDICE`（`overlay-23` entry#9），致傷側用 `ROLLDAMAGEDICE`
（entry#10，會先把骰數寫進一個全域再委派 `ROLLDICE`）。

三支致傷的訊息字串長度都是 **0**——訊息由 `sub_F06h` 或更下層決定，不是由這裡
給。中文化時不要以為這裡沒有文字。

## `sub_F06h` 的 7 個 word 正好驗證了殘留規則

`413Fh` 在呼叫 `sub_F06h` 之前明確推入的只有 5 個 word：

```text
push DS:6F97h ; push 0 ; push 0        ← 3 個
call ROLLDAMAGEDICE(2, 8) ; inc ax ; push ax   ← 傷害量
push 8
lea di, [bp+var_1] ; push ss ; push di
mov di, offset CS:413Eh ; push cs ; push di
call far 0A54h:0634h                   ← 消耗後面兩個 far 指標，留下結果
push cs ; call near sub_F06h
```

而 `sub_F06h` 的結尾是 **`retf 0Eh`** ＝ 14 bytes ＝ 7 個 word。
`5 + 2 = 7`，差的兩個正是 `0A54:0634h` 留在堆疊上的字串結果（spec 690）。

**這是那條規則第一次被 `retf N` 獨立驗證**，不再只是從語意推的。

`sub_F06h` 進來第一件事是把 `[bp+6]`（最後一個參數）複製到 40 字元的緩衝，
與「最後推入的是訊息字串」一致。

## 治療側的兩條訊息路徑

`407Ah` 走 `overlay-24 entry#26`，訊息是常數 `is Healed`；
`3F23h`／`43A7h` 走 `overlay-24 entry#29`，那支自己依「HP 是否等於上限」在
`is fully healed` 與 `is partially healed` 之間選（spec 693）。

所以同樣是治療，玩家看到的句子有兩套來源。中文化時兩邊都要處理。

## `HEALDUDE` 的第三個參數固定 0

六支裡三支治療的呼叫都是 `HEALDUDE(目標, 量, 0)`——**滿血的人也照治**。
擋下來的只有 `HEALDUDE` 內部那個「狀態要在 `{0,1,4,5}` 裡」的集合測試。

## 明確不宣稱

- `sub_F06h` 前五個參數（`DS:6F97h`、`0`、`0`、傷害量、`8`）各自的意義。
- `DS:6F97h` 是什麼。
- 這六支各自對應哪一個法術編號。
