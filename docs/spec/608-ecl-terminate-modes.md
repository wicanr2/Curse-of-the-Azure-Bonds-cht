# 第六百零八輪：`38h`（結束 script 並指定方式）

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:3393h`（397 bytes，126 條指令）。

```text
if DS:7898h <> 0 then
    DS:9594h := DS:789Dh ; DS:7898h := 0      ← 先還原目前目標
READVAR(1)
case ADDRESSVALUE(1) of
    0: <重繪>
       if DS:A325h <> 50h and bank0^[1CCh] = 0 then <far 014A:00D9>
    8: if <far 081F:034E>(DS:7F38h) then       ← 檔案存在
           <顯示>「データの整理をします リターン・キーを押してください」
           <一連串存檔／載入動作>
       <far 00F3:0025>
       DS:BE01h := 1 ; bank0^[3FAh] := FFh ; bank1^[550h] := FFh
       <走遍全隊鏈>
    9: saved := ECL_PC
       <sub_3237>()
       ECL_PC := saved                        ← 呼叫前後保住 PC
       <opcode 00h 的 handler>                ← 結束 script
    3: DS:7F34h := 1
       <opcode 00h 的 handler>                ← 結束 script
    其他: 什麼都不做
```

## 三件事

1. **`9` 與 `3` 都轉去執行 opcode `00h` 的 handler**——也就是結束整個 script
   並清空 GOSUB 堆疊（[spec 588](588-ecl-return-and-gosub-stack.md)）。加上
   `RETURN` 在堆疊空時、`39h` 選單取消時
   （[spec 595](595-ecl-target-selection-and-effect-query.md)），目前已知有
   **四條路**會走到 `00h`。
2. **`9` 在呼叫 `sub_3237` 前後把 `ECL_PC` 存起來再還原**——那支會改動 PC，
   而這裡不要它的副作用。
3. `3` 設的 `DS:7F34h` 就是 `3Eh` 用來表示「全隊都不能行動」的旗標
   （[spec 596](596-ecl-party-item-sweep.md)）。所以 `38h(3)` 是「以全隊失去
   行動能力的狀態結束」。

## `8`：存檔／資料整理

顯示的訊息是**「データの整理をします リターン・キーを押してください」**
（要整理資料，請按 Enter 鍵）。之後把 `bank0^[3FAh]` 與 `bank1^[550h]` 都設成
`FFh`，並走遍全隊鏈——`FFh` 在這裡又是哨兵（清空／無效標記）。

## 明確不宣稱

- `sub_3237` 做什麼。
- `8` 那一路的存檔格式與 `DS:7F38h` 指向什麼。
- `bank0^[3FAh]`／`bank1^[550h]` 的語意。
