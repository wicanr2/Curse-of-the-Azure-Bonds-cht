# 1145 — `1Ch CLEARMONSTERS` 連同還沒領走的戰利品一起丟掉

- 證據等級：`exact`（DOS `overlay-02:120Eh`..`1270h` 逐條讀完，37 條）
- 作法見 spec 783；池與鏈的語意見 spec 1059（七種貨幣）與 spec 1087（`27h TREASURE`）
- 查詢工具：`cmd/ecl-treasure-clear`

## 結論

`1Ch` 的名字只講了一半。它除了釋放怪物相關狀態，還把**這一場遭遇累積、玩家還沒
領走的戰利品**整堆丟掉：`DS:6F70h` 起 28 個位元組歸零，`DS:6F8Ch` 鏈逐節點釋放。

## 一手證據（全函式 37 條）

```asm
1214  inc  word ptr ds:4FB4h     ; ECL 程式計數器 ＋1（每個 handler 都做，spec 1104）
1218  mov  byte ptr ds:47E6h, 0  ; 已放置怪物數
121D  mov  byte ptr ds:8B69h, 0  ; 「有怪要打」旗標（spec 1095）
1222  mov  byte ptr ds:7603h, 8  ; ★ 設成 8，不是清成 0
1227  di := 6F70h
1233  call far 0A54h:1AE0h       ; FillChar(DS:6F70h, 1Ch＝28, 0)
1238  while (DS:6F8Ch <> nil) do begin
1245    next := 節點^[2Ah];
125A    FreeMem(DS:6F8Ch, 3Fh);  ; 63 bytes 一筆
1264    DS:6F8Ch := next;
      end;
```

## 那兩塊是什麼

| 位址 | 是什麼 | 出處 |
|---|---|---|
| `DS:6F70h ＋ i × 4`（`i = 0..6`）| 七種貨幣／寶石／珠寶的**戰利品池**，`4 × 7 ＝ 28` bytes | spec 1059 的「從池中取用」選單逐格讀它 |
| `DS:6F8Ch` | `27h TREASURE` 串進去的**物品節點鏈**，一筆 `3Fh ＝ 63` bytes、next 在 `+2Ah` | spec 1087 §`27h` |

兩份規格各自從別的方向讀出同一組位址，互相印證。

## corpus 的慣用法：先清再擺

`ECL2/0x04` 兩處：

```
0x036c  CLEARMONSTERS          0x07c8  CLEARMONSTERS
0x036d  LOAD MONSTER           0x07c9  TREASURE
0x0374  LOAD MONSTER           0x07da  COMBAT
0x037b  TREASURE
0x038e  COMBAT
```

⇒ `1Ch` 是「把上一場的殘留清乾淨、再擺下一場」的重置指令。

## 唯一走得到「發了又清」的地方

`cmd/ecl-treasure-clear` 前向走訪（跳躍邊 ＋ 循序後繼）全 corpus 的 **63 處 `27h`**，
走得到 `1Ch` 的只有 **1 組**：`ECL2/0x04` 的 `0x037b → 0x036c`，路上經過 `24h`。
那是提爾佛頓火刀首領的**重打迴圈**——`COMPARE 7EC7 80 / IF > / GOTO 0x8359` 跳回
擺場的開頭，`1Ch` 在那裡把上一輪的戰利品丟掉再擺新的一堆。

⚠ 前向走訪只用 `TraceGraph.Edges` 會得到 0 組（那些邊只記跳躍），看起來像「原作
這個行為在本作用不到」。

## remake

`1Ch` 現在清三樣東西：怪物鏈（`MonsterSpawns`）、**這一次執行裡排在它前面的**
戰利品請求（順序由執行本身決定，不必事後猜），以及 State 那一側**跨執行累積**的
`pendingTreasure`／`pendingTreasureItems` 與寶石、珠寶池。

⚠ **金幣不動**：remake 的金幣是直接入帳的，清它等於沒收玩家的錢。原作那一側金幣
還在池裡沒發，兩者的分界不同，這一點刻意不對齊。

## 明確不宣稱

- 沒有宣稱 `DS:7603h` 設成 `8` 是什麼意思（它與 `SETUP MONSTER` 寫的 `7601h`／
  `7602h` 相鄰，但語意還沒讀）。
- 沒有宣稱 `DS:4FB4h` 以外的計數器語意——那一條是每個 handler 都有的 PC 推進。
