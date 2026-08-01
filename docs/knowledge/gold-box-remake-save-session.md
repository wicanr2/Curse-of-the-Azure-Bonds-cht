# Gold Box 重製的 ECL session 存檔原則

Gold Box 存檔不能只保存隊伍與座標。ECL interpreter 可能正停在文字、選單、
戰鬥、財寶或外部 routine boundary；共享 work memory、resume PC 與亂數串流
共同決定下一個結果。

## 最小一致 transaction

```text
原始 game image
  └─ ECL code bytes（玩家自備，載入時重建）

remake save
  ├─ current block + resume PC + stack
  ├─ mutable work/string/compare state
  ├─ input cursors + pending monster descriptors
  └─ PRNG seed + underlying-source draw count
```

只保存 seed 會讓讀檔後亂數回到開頭；只保存 RNG 而不保存 ECL memory，則會
讓同一亂數落入不同劇情條件。兩者都不是 faithful continuation。

## 商業資料與向下相容

code window 只保存相對於玩家自備 block 的 runtime differences，不複製完整
ECL bytes。舊 save 沒有某欄位時應採版本化相容預設，文件必須如實標成「未曾
保存」，不能根據目前 seed 猜造過去的 draw position。

## 仍需第二款作品驗證

目前 contract 已由 CoAB 真實 Burial Glen 隨機事件驗證；在第二款 Gold Box
採用前仍是 cross-title candidate。下一款需交叉驗證跨 block、選單 pause、
戰鬥 handoff、存讀檔後的下一批 RANDOM，以及長局 replay 上限。
