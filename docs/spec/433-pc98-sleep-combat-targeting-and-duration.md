# 第 433 輪：PC-98 睡眠術戰鬥目標入口與持續時間

狀態：`READY`（限 combat targeting dispatch、`AOECOMBAT=9` 分支、無豁免與
`5 × caster level` 持續時間；魔法抗性與 `PUTEFFECT` 已由 spec 434 接續，
畫面演出仍待續；`SCAN` 三欄與排序已由 spec 435 訂正並接續）

## 結論

第 432 輪把 `DS:A38Ch DOSPELLTARGETING` 誤連到 overlay 12，是位址空間混用。
重新依 resident segment、TPOV entry 與 stack cleanup 驗證後，正確資料流是：

1. `INITSPELLS` 的預設值 `0117:0034h` 是 overlay 22 entry 4、
   `GETSPELLTARGETS 112Ch`；它讀 `SPELLREC.AOENONCOMBAT(+07h)`。
2. 進入戰鬥時 overlay 08 `0154h..0157h` 把同一 function pointer 覆寫為
   `00B8:007Ah`；typed resolver 證明它是 overlay 13 entry 18、local
   `225Fh` 的戰鬥目標處理器。
3. 戰鬥處理器讀 `SPELLREC.AOECOMBAT(+06h)`。Sleep `15h` 的值是 `09h`，
   因而走 `08h..0Eh` 選格分支，以低三 bits `1` 呼叫 LOS `SCAN`，再把
   `COMBATMAP` 產生的 object ID 順序投影成 `SPELLTARGET[1..N]`。
4. overlay 31 的 `SCAN 08D8h` 先清 `LASTSIGHT 9F30h`，依原始 combat-object
   traversal 建立三 byte candidate records，再由 local `0035h` 原地排序。
   Sleep handler 因此不是自行按 HP、HD 或 stable ID 排序；第 432 輪的
   `4d4` 容量篩選必須保留這份上游順序。
5. 共用 effect writer 對 Sleep 讀到 `SAVERESULT=0`，不呼叫 saving-throw
   分支。持續時間 helper 走一般公式：
   `DURATIONFIXED(0) + FIGCASTERLEVEL × DURATIONPERLEVEL(5)`。

本輪仍未證明 `PUTEFFECT` 內的 magic-resistance gate、raw effect record
全部欄位、解除時點及動畫／聲音。因此 remake 仍維持失敗即關閉，不把 Sleep
加入可選戰鬥法術。

## 非破壞性輸入

| 輸入 | SHA-256 | 用途 |
|---|---|---|
| `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland symbols、resident pointers、TPOV stubs |
| `PC98-GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | typed overlay resolution |
| overlay 08 | `d39b6aad76af8d3ccc182b4b4d95dd9195160c9c21b36bfa5b0cd4edc1942788` | combat init writer |
| overlay 13 | `db1c03daaa984b94056e67d7b9e10fdc6bd3f2393f8a9e5ec720703f93835d4a` | combat targeting entry 18 |
| overlay 22 | `c54729525d576c11d731d64a1b06ee2547b2562b73e3708a1beaafc535cabbe8` | default targeting、Sleep dispatch、common writer |
| overlay 31 | `6cd5e38dddeb1ea5ddc44dd7f3af68c49bd9b67198e7c42247f1d08743050081` | LOS `SCAN` candidate list 與排序 |

所有來源均唯讀掛載。IDA 只在 `/tmp/coab-ida-433` 的副本建立 database 與
ledger；版本控制只保存 IDC、hash、位址與結論，不保存或 rename database。

## Targeting dispatch 訂正

overlay 22 `INITSPELLS 66EBh`：

```text
66EE B8 34 00       mov ax,0034h
66F1 BA 17 01       mov dx,0117h
66F4 A3 8C A3       mov [A38Ch],ax
66F7 89 16 8E A3    mov [A38Eh],dx
```

`22:0034h` 解析為 entry 4／local `112Ch GETSPELLTARGETS`，不是
`12:0034h`。後者是另一個 overlay 的 entry 4，雖然數字相同，卻屬不同
resident control segment；把兩者合併正是本輪訂正的錯誤來源。

overlay 08 戰鬥初始化：

```text
0154 B8 7A 00       mov ax,007Ah
0157 BA B8 00       mov dx,00B8h
015A A3 8C A3       mov [A38Ch],ax
015D 89 16 8E A3    mov [A38Eh],dx
```

`13:007Ah → entry 18 → 225Fh`。退出／回復一般流程時，overlay 08
`00E2h..00EBh` 又寫回 `0117:0034h`。因此 `GETSPELLTARGETS` 的
`AOENONCOMBAT` 分支不能用來推論 Sleep 戰鬥幾何。

## Sleep 的 `AOECOMBAT=9` 分支

Borland `SPELLREC` 已證明 `+06h=AOECOMBAT`。戰鬥 handler `225Fh`：

- `229Ch..22ABh` 讀 `SPELLINFO[spell].AOECOMBAT & 0Fh`；
- `2561h..256Ah` 接受 `08h..0Eh`；Sleep 的 `09h` 命中；
- `256Fh..2583h` 經 local `2049h` 取得玩家／Quick 選定的合法 target；
- `2592h..25B4h` 把 `AOECOMBAT & 07h`（Sleep 為 `1`）傳入
  `0184:003Eh SCAN`，並使用 `COMBATMAP 9F2Ch`；
- `25BFh..260Ah` 依 `LASTSIGHT 9F30h` 的 `1..N` 順序，從每筆三 byte
  record 的 object ID 取 `COMPOBJ` far pointer，原樣寫進
  `SPELLTARGET[1..N]`；沒有第二次排序。

overlay 31 `31:003Eh → entry 6 → 08D8h SCAN` 的輸出 record 是三 bytes。
`0945h` 清 count，`0B04h` 每接受一名 candidate 加一，`0B17h／0B2Ah／
0B8Ch` 依序寫 object ID 與兩個排序欄位，最後呼叫 local `0035h` 排序。
本輪原先把等距時的奇偶邏輯歸到第三欄，已被 spec 435 的完整連續指令推翻：
排序以第二欄為主要遞增鍵，等值時只比較第一欄 object ID 與其奇偶；第三欄
是方向 payload，排序器完全不讀。欄位 producer、footprint 展開與 exact
巢狀排序見 spec 435；本節只保留 dispatch 與 copy order。

### 推論等級

- `exact`：兩次 `DOSPELLTARGETING` writer、兩個 typed stub resolution、
  `AOECOMBAT=09h` 分派、`SCAN` call、三 byte list 建立／排序、
  `SPELLTARGET` copy order、`SAVERESULT=0` 跳過 save，以及 duration 公式。
- `exact`（由 spec 435 接續）：三欄依序是 object ID、最小成功 LOS 加權
  距離低 byte、方向 sector；第三欄不參與排序。
- `unknown`：terrain／large-footprint 對 Sleep list 的所有邊界、
  `PUTEFFECT` magic-resistance 行為、效果解除、動畫與音訊。

## 無豁免與持續時間

Sleep handler `2656h..2676h` 以 spell `15h` 和四個零參數進入 common writer
`0F62h`。對此固定 record：

- `SAVERESULT(+08h)=0`，故 `1004h..100Fh` 直接令 save result 為零，
  不呼叫後續 saving throw helper；`SAVEVS(+09h)=4` 在這條路徑不消耗。
- effect `35h` 的 duration helper `0E75h` 走一般分支；
  `014A:00D4h` typed 解析為 `FIGCASTERLEVEL 2AEAh`，乘
  `DURATIONPERLEVEL(+05h)=5`，再加 `DURATIONFIXED(+04h)=0`。

這只能證明原始 effect writer 收到的 duration 數值，不等於 remake 已完成
每回合 decrement、喚醒、受傷解除或 save round-trip。

## 實作門檻與下一步

1. 追 `PUTEFFECT 013E:2325h` 的正確 overlay/module projection，閉合 magic
   resistance、effect record 欄位與 raw `35h` writer。
2. 以 DOSBox／PC-98 runtime 固定場面取得 Sleep 候選 list，驗證 terrain
   property、large footprint 與畫面 target cursor；三欄 producer／sort 已由
   spec 435 靜態閉合。
3. 接續 spec 435 的 target-order adapter，再完成 Battle effect lifecycle、
   手動／Quick delayed cast、法術格消耗、繁中訊息、動畫與聲音。
