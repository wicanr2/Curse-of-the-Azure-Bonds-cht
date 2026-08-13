# 第六百八十二輪：兩支法術強度計算 —— 多推一個參數當加項

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-22` 的 `21C7h`、`2260h`。

## `21C7h`

```text
n := <far 013Eh:05A4h>(DS:6F97h) + 1
half := n div 2                                  ← cwd ＋ idiv 2（有號）
v := <sub_1462>(DS:6F97h, 0, 0, half, 4)
half := n div 2                                  ← ★ 再算一次
<sub_F06>(8, v + half, 空字串)
```

`n div 2` **算了兩遍**——第一次的結果推進參數，第二次重算來當加項。編譯器沒有把
它存起來重用。同樣的重複除法也出現在
[spec 672](672-numeric-menu-input-and-nibble-array.md) 的 `129EBh`。

## `2260h`：多推一個參數，事後 `pop` 回來當加項

```asm
push ds:6F97h                    ; ← 多推的那一個
push 0
push 0
push ds:6F97h
call far ptr loc_15A3+1           ; 巢狀呼叫，吃掉剛才那個
push ax                           ; 巢狀呼叫的結果
push 1
push 8
call far ptr sub_1462             ; 只吃 5 個 word
xor  ah, ah
pop  dx                           ; ← 把最前面那個推入取回來
add  ax, dx
```

`sub_1462` 的 `retf` 只清 **5 個 word**（由 `21C7h` 的五個推入確認），所以 `2260h`
推的六個裡**最前面那個留在堆疊上**，被呼叫後的 `pop dx` 取回來當加項。

這不是溢出或錯誤——**是刻意把「要傳的參數」與「事後要用的值」用同一個推入完成**。
判讀時若照「呼叫前的推入都是參數」去數，就會把 `sub_1462` 算成 6 個參數。

兩支的加項來源不同：`21C7h` 是重算的 `n div 2`、`2260h` 是這個留在堆疊上的
`DS:6F97h`。

## 兩支傳給 `sub_F06` 的字串都是空的

`21C6h` 與 `225Fh` 的長度位元組都是 **0**——與
[spec 627](627-spell-cure-family.md) 的 `43DAh`／`46CAh` 一樣，訊息由 `sub_F06`
自己組，呼叫端只給空字串。

## 明確不宣稱

- `sub_1462`／`sub_F06`／`013Eh:05A4h` 的行為。
- `DS:6F97h` 的語意（在兩支裡既當參數又當加項）。
- 常數 `4`／`8`／`0Ch` 的意義。
