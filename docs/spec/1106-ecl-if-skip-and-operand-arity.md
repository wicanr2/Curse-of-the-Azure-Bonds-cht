# 1106 — `IF` 跳過下一條指令：靜態清冊先前漏掉三分之二的程式碼

- 狀態：`READY`
- 證據等級：`exact`（DOS `overlay-02:0B3Bh` 43 條、`overlay-07 entry#29`（局部 `1FF5h`）139 條逐條讀完）
- 通則與 operand 解碼見 spec 1104；清冊契約見 spec 557

## 一、`16h`..`1Bh IF` 的實際語意

`0B3Bh`（43 條）：

```pascal
inc DS:4FB4h;                       { IF 是單 byte 指令 }
case DS:75FFh of                    { 重讀目前 opcode 分流，同 11h/12h 的作法 }
  16h: if DS:75F8h = 0 then 跳過下一條;    { =  }
  17h: if DS:75F9h = 0 then 跳過下一條;    { <> }
  18h: if DS:75FAh = 0 then 跳過下一條;    { <  }
  19h: if DS:75FBh = 0 then 跳過下一條;    { >  }
  1Ah: if DS:75FCh = 0 then 跳過下一條;    { <= }
  1Bh: if DS:75FDh = 0 then 跳過下一條;    { >= }
end;
```

★ 六個比較結果各佔一格（`DS:75F8h`..`75FDh`），由 `03h COMPARE`／`14h COMPARE AND`
之類的指令寫入。**`IF` 本身不比較**，它只看旗標。

★★★ **條件成立就往下執行，不成立就跳過下一條指令**——沒有分支目的地，
所以 `IF` 後面接的通常是 `01h GOTO`：條件成立跳走，不成立就略過那個 `GOTO`
繼續往下。⇒ `IF ...; GOTO x; <else 路徑>` 是這套 bytecode 表達 if/else 的方式。

## ★★★ 二、`overlay-07 entry#29` 是「跳過下一條」，順便是一張 arity 表

`1FF5h`（139 條）做的事只有一件：

```pascal
DS:75FFh := byte[DS:4FA5h^ + DS:4FB4h − 8000h];   { 讀下一條的 opcode }
case DS:75FFh of
    …依 opcode 查出操作元個數 n…
    entry#2(n);                                   { 解 n 個操作元並推進 PC }
else
    inc DS:4FB4h;                                 { 沒有操作元的指令 }
end;
```

那張 case 表就是**每個 opcode 的操作元個數**：

| 個數 | opcode |
|---:|---|
| 1 | `01` `02` `0A` `0E` `11` `12` `1D` `20` `2D` `32` `34` `36` `38` `39` `3C` `3F` `40` |
| 2 | `03` `08` `09` `0F` `10` `1F` `22` |
| 3 | `04`..`07` `0B` `0C` `21` `28` `2A` `2F` `30` `35` `37` `3B` |
| 4 | `14` `23` |
| 5 | `2E` |
| 6 | `1E` `2C` |
| 8 | `27` |
| 14 | `29` |
| 0 | 其餘（走 `else` 的 `inc`） |

### 二之一、原作瑕疵：跳過表少算兩個 opcode

拿這張表逐格對 remake 的解碼器，只有兩處不一致，而**兩處都是原作的跳過表錯**：

| opcode | 跳過表 | handler 實際 | 依據 |
|---|---:|---:|---|
| `34h ECL CLOCK` | 1 | **2** | `overlay-02:2CB5h` 開頭 `entry#2(2)` |
| `36h ADD NPC` | 1 | **2** | `overlay-02:2DA9h` 開頭 `entry#2(2)` |

⇒ 被 `IF` 守衛的 `ECL CLOCK`／`ADD NPC` 一旦條件不成立，PC 會停在最後一個
操作元上，之後整段解碼錯位。

★ **CoAB 踩不到**：全 corpus 沒有任何 `IF` 後面接 `34h` 或 `36h`（實測 0 處）。
⇒ 列入「原作瑕疵」清單但不需要決定照不照抄。
★ **以 handler 為準，不是以跳過表為準**：remake 的解碼器兩處都寫 2，是對的。

## ★★★ 三、對靜態清冊的影響：可達指令 1,355 → 4,222

`ecl.TraceGraph` 原本在 `GOTO` 就中斷線性走訪。配合 §一，那等於**每個 if/else
都只走 then 那一半**，else 路徑整段看不到——而劇情文字大多在 else 那邊。

修正：被 `IF` 守衛的那條指令，除了照常處理之外，它的 `Next` 也要進佇列。

| 指標 | 修正前 | 修正後 |
|---|---:|---:|
| 不重複靜態可達 instruction | 1,355 | **4,222** |
| 跨 effect-kind 直線候選 | 32 | **154** |
| corpus 出現的 opcode 種類 | 46 | **55** |
| 靜態可達的文字段落 | 43 | **221** |

新出現的 9 個 opcode（`0Dh` `10h` `15h` `28h` `2Eh` `39h` `3Ch` `3Eh` `3Fh`）
先前完全不在清冊裡。它們在 opcode 相位台帳中一律標 `unknown`——本輪沒有讀它們的
handler，只是讓它們變成可盤點的對象。

★ 這是本專案 fail-closed 設計的一次實績：擴大可達性之後，
`VerifyPhaseCoverage` 立刻指出台帳缺了 9 個 opcode，而不是靜默漏掉。

### 三之一、`38h PROGRAM` 現在依值切斷直線區段

spec 1104 §九當時保守不依值終止，理由是「運算元可能是記憶體讀取」。擴大可達性
之後可以逐一檢查：corpus 的三個 `PROGRAM` **全是立即值**，且全是 3 或 9
（`ECL1/0x52 +0391h` ＝ 3、`ECL3/0x12 +1181h` ＝ 9、`ECL4/0x23 +012D h` ＝ 9），
兩者都轉呼叫 `00h` 的 handler。⇒ 立即值 3 與 9 切斷直線區段；值 0、值 8 與
非立即值不切（後者靜態上判不出來）。

## 四、審查結論的搬移

候選 ID 是 `member/block/起點-終點`，區段邊界改變會讓 ID 位移。33 筆已審結論的
處置是**逐筆比對效果序列**（`(offset, opcode)` 的序列），完全相同才沿用：

- 31 筆在新清冊有唯一的相同序列 ⇒ 直接沿用。
- 1 筆（`ECL3.DAX/0x11` 的 `PICTURE → PRINTCLEAR → CALL`）對到**兩個**候選：
  同樣三條指令，只是從不同 lifecycle entry 進來使區段起點不同 ⇒ 兩個都套用。
- 0 筆需要重新審查。

⚠ **不要靠起點對齊搬移結論。** 起點會因為前面多了幾條可達指令而移動，
但效果序列不會；用序列比對才不會把結論套到不同的程式碼上。

## 五、明確不宣稱

- 沒有宣稱 `DS:75F8h`..`75FDh` 六格分別由哪些指令寫入（`03h`／`14h` 的 handler 未讀）。
- 沒有宣稱新出現那 9 個 opcode 的任何副作用或次序。
- 沒有宣稱 4,222 就是完整可達集合：`ON GOTO`／`ON GOSUB` 的動態目的地、選單分支與
  `CALL` 之後的控制流仍未納入。**這個數字只會再往上。**

## 六、回歸

| 測試 | 釘住什麼 |
|---|---|
| `eclcatalog.TestProgramEndsStraightLineOnlyForTerminalImmediates` | 立即值 3／9 切、0／8 與非立即值不切 |
| `eclcatalog.TestNewEclIsTerminalAndEndsStraightLine` | `20h` 仍然切 |
| `eclcatalog.VerifyPhaseCoverage` | corpus 的 opcode 與相位台帳互為子集 |
| `ecl-event-catalog -check` | 清冊與提交版本逐位元組相同 |
