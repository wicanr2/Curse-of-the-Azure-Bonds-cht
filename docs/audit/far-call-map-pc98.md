# Overlay far call → `overlay-NN entry#K` 對照

由 `cmd/far-call-map` 產生，不要手改。

IDA 把每個 overlay 的 `.bin` 從位址 0 載入，所以 `call far ptr` 的目標會被「解析」成同一個 `.bin` 裡的標籤（`loc_1584+2`、`sub_1590` 之類）。**那些標籤與真正的目標無關**，規格裡的 `<far 1590h>` 就是這樣來的。

真正的目標是常駐段裡的 TPOV 控制記錄：段選 overlay、位移選 entry（`(位移 − 20h) / 5`，stub 每筆 5 bytes）。段與檔案位移差 **0x9a 個段**（＝ EXE header 的長度），這個常數是**量出來的**：掃過全部 far call、對每個候選值數有幾個目標正好落在 stub 邊界，取命中最多的那個。

| 指標 | 數字 |
|---|---:|
| far call 總數 | 5262 |
| 解出 `overlay-NN entry#K` | 2291 |
| 目標是常駐程式碼（段不是任何控制記錄）| 2971 |
| 段對得上但位移不在 stub 邊界 | 0 |

## 每個 overlay 的 entry 被叫了幾次

`0` 代表沒有任何 overlay 呼叫它——那種 entry 只會被常駐程式碼或 ECL 叫到。

| overlay | entry 數 | 有呼叫端 | 沒有呼叫端 |
|---|---:|---:|---:|
| `overlay-00` | 3 | 0 | 3 |
| `overlay-01` | 5 | 0 | 5 |
| `overlay-02` | 60 | 0 | 60 |
| `overlay-03` | 4 | 3 | 1 |
| `overlay-04` | 13 | 2 | 11 |
| `overlay-05` | 13 | 2 | 11 |
| `overlay-06` | 6 | 3 | 3 |
| `overlay-07` | 34 | 23 | 11 |
| `overlay-08` | 14 | 2 | 12 |
| `overlay-09` | 16 | 2 | 14 |
| `overlay-10` | 25 | 2 | 23 |
| `overlay-11` | 5 | 2 | 3 |
| `overlay-12` | 136 | 2 | 134 |
| `overlay-13` | 45 | 20 | 25 |
| `overlay-14` | 15 | 3 | 12 |
| `overlay-15` | 31 | 3 | 28 |
| `overlay-16` | 35 | 14 | 21 |
| `overlay-17` | 24 | 6 | 18 |
| `overlay-18` | 16 | 1 | 15 |
| `overlay-19` | 28 | 13 | 15 |
| `overlay-20` | 17 | 3 | 14 |
| `overlay-21` | 23 | 16 | 7 |
| `overlay-22` | 117 | 11 | 106 |
| `overlay-23` | 30 | 23 | 7 |
| `overlay-24` | 53 | 44 | 9 |
| `overlay-25` | 14 | 11 | 3 |
| `overlay-26` | 22 | 10 | 12 |
| `overlay-27` | 8 | 4 | 4 |
| `overlay-28` | 6 | 2 | 4 |
| `overlay-29` | 12 | 9 | 3 |
| `overlay-30` | 15 | 8 | 7 |
| `overlay-31` | 9 | 7 | 2 |
| `overlay-32` | 26 | 20 | 6 |
| `overlay-33` | 8 | 6 | 2 |
| `overlay-34` | 4 | 1 | 3 |
| `overlay-35` | 5 | 3 | 2 |

## 每個 entry 的呼叫端

讀某一支 overlay 函式時要問的是「這個 far call 打到哪」——那個方向逐筆列在 JSON 裡（每個呼叫點一列）。這裡是**反向**：某個 entry 是被誰叫的，用來回答「改這支會影響誰」。

| entry | code | 呼叫端 |
|---|---|---|
| `overlay-03 entry#1` | `0131h` | overlay-02 sub_3520 |
| `overlay-03 entry#2` | `0000h` | overlay-02 sub_0 |
| `overlay-03 entry#3` | `000Ch` | overlay-02 sub_3520 |
| `overlay-04 entry#1` | `0F42h` | overlay-02 sub_1820 |
| `overlay-04 entry#2` | `0000h` | overlay-02 sub_1820 |
| `overlay-05 entry#1` | `1775h` | overlay-02 sub_1820 |
| `overlay-05 entry#2` | `0000h` | overlay-02 sub_1820 |
| `overlay-06 entry#1` | `0778h` | overlay-02 sub_1820 |
| `overlay-06 entry#2` | `03A0h` | overlay-05 sub_DCC |
| `overlay-06 entry#3` | `0000h` | overlay-02 sub_1820、overlay-05 sub_0 |
| `overlay-07 entry#1` | `0296h` | overlay-02 sub_10C2、overlay-02 sub_11E、overlay-02 sub_149C、overlay-02 sub_1758、overlay-02 sub_1B7、overlay-02 sub_1BEA、overlay-02 sub_2125、overlay-02 sub_222C、overlay-02 sub_25A、overlay-02 sub_2940、overlay-02 sub_29FB、overlay-02 sub_2ACE、overlay-02 sub_2B8、overlay-02 sub_2E2C、overlay-02 sub_2F5F、overlay-02 sub_2FDA、overlay-02 sub_306、overlay-02 sub_3393、overlay-02 sub_364B、overlay-02 sub_369F、overlay-02 sub_3E0、overlay-02 sub_49C、overlay-02 sub_8A7、overlay-02 sub_A5D、overlay-02 sub_B49、overlay-02 sub_C26、overlay-02 sub_C81、overlay-02 sub_E10、overlay-02 sub_E7F、overlay-02 sub_EDD |
| `overlay-07 entry#2` | `008Eh` | overlay-02 sub_107、overlay-02 sub_10C2、overlay-02 sub_11E、overlay-02 sub_12F7、overlay-02 sub_149C、overlay-02 sub_16BC、overlay-02 sub_1758、overlay-02 sub_1B7、overlay-02 sub_1BEA、overlay-02 sub_2125、overlay-02 sub_222C、overlay-02 sub_25A、overlay-02 sub_2940、overlay-02 sub_29FB、overlay-02 sub_2ACE、overlay-02 sub_2B8、overlay-02 sub_2E2C、overlay-02 sub_2EDA、overlay-02 sub_2F5F、overlay-02 sub_2FDA、overlay-02 sub_306、overlay-02 sub_30C6、overlay-02 sub_3393、overlay-02 sub_3520、overlay-02 sub_364B、overlay-02 sub_369F、overlay-02 sub_3E0、overlay-02 sub_49C、overlay-02 sub_8A7、overlay-02 sub_992、overlay-02 sub_9DD、overlay-02 sub_A5D、overlay-02 sub_B49、overlay-02 sub_C26、overlay-02 sub_C81、overlay-02 sub_E10、overlay-02 sub_E7F、overlay-02 sub_E8、overlay-02 sub_EDD、overlay-02 sub_F29 |
| `overlay-07 entry#3` | `0317h` | overlay-02 sub_3A21、overlay-02 sub_C26 |
| `overlay-07 entry#4` | `0499h` | overlay-02 sub_3A21、overlay-02 sub_C26 |
| `overlay-07 entry#5` | `0556h` | overlay-02 sub_49C |
| `overlay-07 entry#6` | `05BDh` | overlay-02 sub_1820、overlay-02 sub_222C、overlay-02 sub_3E0 |
| `overlay-07 entry#7` | `064Ch` | overlay-02 sub_8A7 |
| `overlay-07 entry#8` | `068Bh` | overlay-02 sub_222C、overlay-02 sub_3E0、overlay-02 sub_858 |
| `overlay-07 entry#9` | `07DCh` | overlay-02 sub_10C2、overlay-02 sub_12F7、overlay-02 sub_149C、overlay-02 sub_14ED、overlay-02 sub_16BC、overlay-02 sub_1B7、overlay-02 sub_222C、overlay-02 sub_25A、overlay-02 sub_2940、overlay-02 sub_2B8、overlay-02 sub_2FDA、overlay-02 sub_30C6、overlay-02 sub_992、overlay-02 sub_9DD、overlay-02 sub_E7F、overlay-02 sub_E8、overlay-02 sub_EDD、overlay-02 sub_F29 |
| `overlay-07 entry#15` | `0E2Bh` | overlay-02 sub_107A、overlay-02 sub_10C2、overlay-02 sub_12F7、overlay-02 sub_143C、overlay-02 sub_16BC、overlay-02 sub_1817、overlay-02 sub_1B7、overlay-02 sub_222C、overlay-02 sub_25A、overlay-02 sub_2940、overlay-02 sub_2B8、overlay-02 sub_2FDA、overlay-02 sub_992、overlay-02 sub_EDD |
| `overlay-07 entry#17` | `1148h` | overlay-02 sub_2B8、overlay-02 sub_9DD |
| `overlay-07 entry#20` | `1780h` | overlay-02 sub_10C2、overlay-02 sub_222C、overlay-02 sub_2940、overlay-04 sub_114F、overlay-06 sub_778 |
| `overlay-07 entry#22` | `1928h` | overlay-02 sub_11E |
| `overlay-07 entry#23` | `19FEh` | overlay-02 sub_11E |
| `overlay-07 entry#24` | `1A66h` | overlay-02 sub_107 |
| `overlay-07 entry#25` | `1B89h` | overlay-02 sub_30C6 |
| `overlay-07 entry#26` | `1AECh` | overlay-02 sub_30C6 |
| `overlay-07 entry#27` | `1DD1h` | overlay-02 sub_2125 |
| `overlay-07 entry#28` | `1F00h` | overlay-02 sub_2125 |
| `overlay-07 entry#29` | `1FB0h` | overlay-02 sub_BAE |
| `overlay-07 entry#30` | `20C5h` | overlay-02 sub_222C |
| `overlay-07 entry#31` | `2235h` | overlay-02 sub_2ACE |
| `overlay-07 entry#32` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0 |
| `overlay-08 entry#1` | `0143h` | overlay-02 sub_1820 |
| `overlay-08 entry#2` | `0000h` | overlay-02 sub_1820 |
| `overlay-09 entry#1` | `005Dh` | overlay-08 sub_2BB、overlay-08 sub_414 |
| `overlay-09 entry#2` | `0000h` | overlay-02 sub_1820、overlay-08 sub_0 |
| `overlay-10 entry#1` | `1C28h` | overlay-08 sub_143 |
| `overlay-10 entry#2` | `0000h` | overlay-02 sub_1820、overlay-08 sub_0 |
| `overlay-11 entry#2` | `0508h` | overlay-02 sub_3393 |
| `overlay-11 entry#3` | `0000h` | overlay-02 sub_3393 |
| `overlay-12 entry#1` | `2ED4h` | overlay-11 sub_48 |
| `overlay-12 entry#2` | `0000h` | overlay-11 sub_0 |
| `overlay-13 entry#1` | `0000h` | overlay-08 sub_143 |
| `overlay-13 entry#5` | `078Eh` | overlay-08 sub_C1D、overlay-09 sub_C4A |
| `overlay-13 entry#6` | `0999h` | overlay-08 sub_C1D、overlay-09 sub_C4A |
| `overlay-13 entry#7` | `0D81h` | overlay-08 sub_C1D、overlay-09 sub_9E9 |
| `overlay-13 entry#8` | `0E3Eh` | overlay-08 sub_414、overlay-09 sub_1923 |
| `overlay-13 entry#10` | `0FB1h` | overlay-08 sub_F17、overlay-09 sub_FE8 |
| `overlay-13 entry#11` | `11AFh` | overlay-09 sub_DD3 |
| `overlay-13 entry#12` | `12A8h` | overlay-08 sub_414、overlay-09 sub_26F |
| `overlay-13 entry#13` | `14B4h` | overlay-09 sub_26F |
| `overlay-13 entry#14` | `19CBh` | overlay-08 sub_F17、overlay-09 sub_FE8 |
| `overlay-13 entry#15` | `1A59h` | overlay-08 sub_F17、overlay-09 sub_FE8 |
| `overlay-13 entry#17` | `1E30h` | overlay-09 sub_3D3 |
| `overlay-13 entry#19` | `27A1h` | overlay-08 sub_414、overlay-09 sub_627 |
| `overlay-13 entry#20` | `2DB9h` | overlay-08 sub_F17 |
| `overlay-13 entry#21` | `3A74h` | overlay-08 sub_414 |
| `overlay-13 entry#23` | `29F4h` | overlay-09 sub_9E9、overlay-09 sub_FE8 |
| `overlay-13 entry#25` | `2C94h` | overlay-08 sub_9A5 |
| `overlay-13 entry#27` | `3D7Fh` | overlay-09 sub_5D、overlay-09 sub_C4A |
| `overlay-13 entry#33` | `4876h` | overlay-08 sub_11C7 |
| `overlay-13 entry#34` | `1DD7h` | overlay-02 sub_1820、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0 |
| `overlay-14 entry#1` | `0BCCh` | overlay-02 sub_3A21 |
| `overlay-14 entry#2` | `08E4h` | overlay-02 sub_3A21 |
| `overlay-14 entry#3` | `0000h` | overlay-02 sub_0、overlay-11 sub_0 |
| `overlay-15 entry#1` | `2457h` | overlay-02 sub_3237 |
| `overlay-15 entry#2` | `04C2h` | overlay-14 sub_8E4 |
| `overlay-15 entry#3` | `0000h` | overlay-02 sub_3237、overlay-02 sub_3A21 |
| `overlay-16 entry#1` | `0614h` | overlay-17 sub_3BF4 |
| `overlay-16 entry#2` | `0B4Fh` | overlay-17 sub_2CB、overlay-17 sub_3BF4 |
| `overlay-16 entry#4` | `1061h` | overlay-17 sub_3BF4、overlay-17 sub_45D0 |
| `overlay-16 entry#5` | `1261h` | overlay-17 sub_28E1 |
| `overlay-16 entry#6` | `14C9h` | overlay-17 sub_1627、overlay-17 sub_2CB |
| `overlay-16 entry#7` | `2A6Dh` | overlay-17 sub_3BF4 |
| `overlay-16 entry#8` | `3AA8h` | overlay-07 sub_556 |
| `overlay-16 entry#9` | `3DFFh` | overlay-02 sub_2F5F |
| `overlay-16 entry#10` | `42FEh` | overlay-17 sub_2CB |
| `overlay-16 entry#11` | `46CEh` | overlay-00 sub_1B、overlay-15 sub_2457、overlay-17 sub_2CB |
| `overlay-16 entry#12` | `3E8Bh` | overlay-17 sub_3BF4 |
| `overlay-16 entry#14` | `0000h` | overlay-02 sub_0、overlay-07 sub_0、overlay-10 sub_0、overlay-14 sub_0、overlay-15 sub_0、overlay-17 sub_0 |
| `overlay-16 entry#15` | `4961h` | overlay-02 sub_3393 |
| `overlay-16 entry#17` | `4E65h` | overlay-02 sub_3393、overlay-02 sub_3A21 |
| `overlay-17 entry#1` | `02CBh` | overlay-02 sub_3393 |
| `overlay-17 entry#3` | `4122h` | overlay-02 sub_306、overlay-02 sub_35AF、overlay-05 sub_52F、overlay-10 sub_1900、overlay-15 sub_19A4 |
| `overlay-17 entry#4` | `45D0h` | overlay-15 sub_1D4C |
| `overlay-17 entry#5` | `0000h` | overlay-02 sub_3393、overlay-10 sub_0 |
| `overlay-17 entry#6` | `5420h` | overlay-16 sub_26CF、overlay-16 sub_2A6D |
| `overlay-17 entry#7` | `4CBBh` | overlay-16 sub_26CF、overlay-16 sub_2A6D |
| `overlay-18 entry#1` | `1213h` | overlay-02 sub_3393 |
| `overlay-19 entry#1` | `0098h` | overlay-17 sub_1627、overlay-17 sub_2E9B |
| `overlay-19 entry#2` | `069Eh` | overlay-17 sub_1627、overlay-17 sub_2E9B |
| `overlay-19 entry#3` | `09BCh` | overlay-17 sub_1627、overlay-17 sub_2DB7 |
| `overlay-19 entry#4` | `05B1h` | overlay-17 sub_1627 |
| `overlay-19 entry#5` | `0B82h` | overlay-04 sub_101A、overlay-06 sub_778、overlay-08 sub_414、overlay-14 sub_8E4、overlay-15 sub_2457、overlay-17 sub_2CB |
| `overlay-19 entry#6` | `1578h` | overlay-08 sub_414 |
| `overlay-19 entry#7` | `1EAEh` | overlay-09 sub_1923、overlay-12 sub_5BD |
| `overlay-19 entry#8` | `246Fh` | overlay-09 sub_4CC |
| `overlay-19 entry#9` | `32B7h` | overlay-06 sub_3A0 |
| `overlay-19 entry#10` | `333Bh` | overlay-04 sub_101A、overlay-04 sub_F42、overlay-05 sub_F0F、overlay-06 sub_778、overlay-07 sub_1808、overlay-14 sub_8E4、overlay-15 sub_14CA、overlay-15 sub_17F6、overlay-15 sub_1D4C、overlay-15 sub_2457 |
| `overlay-19 entry#11` | `340Ah` | overlay-04 sub_103、overlay-06 sub_4BE、overlay-16 sub_26CF、overlay-17 sub_5420 |
| `overlay-19 entry#12` | `34E4h` | overlay-13 sub_27A1、overlay-15 sub_4C2、overlay-15 sub_92D、overlay-15 sub_B86、overlay-17 sub_546F |
| `overlay-19 entry#18` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-13 sub_1DD7、overlay-14 sub_0、overlay-15 sub_0、overlay-17 sub_0 |
| `overlay-20 entry#2` | `03B9h` | overlay-02 sub_2E2C、overlay-08 sub_9A5、overlay-14 sub_807、overlay-14 sub_8E4 |
| `overlay-20 entry#3` | `0CD0h` | overlay-15 sub_2370、overlay-15 sub_7B4 |
| `overlay-20 entry#4` | `0000h` | overlay-02 sub_0、overlay-08 sub_0、overlay-14 sub_0、overlay-15 sub_0、overlay-19 sub_0 |
| `overlay-21 entry#4` | `0073h` | overlay-19 sub_27DC |
| `overlay-21 entry#5` | `0541h` | overlay-04 sub_101A、overlay-06 sub_778 |
| `overlay-21 entry#7` | `0669h` | overlay-04 sub_101A、overlay-06 sub_778 |
| `overlay-21 entry#8` | `0DE8h` | overlay-04 sub_101A、overlay-05 sub_F0F、overlay-06 sub_778 |
| `overlay-21 entry#9` | `0218h` | overlay-16 sub_26CF、overlay-16 sub_2A6D |
| `overlay-21 entry#10` | `0B9Bh` | overlay-19 sub_2C69、overlay-19 sub_2FE5 |
| `overlay-21 entry#11` | `02ABh` | overlay-19 sub_2C69、overlay-19 sub_2FE5 |
| `overlay-21 entry#12` | `0485h` | overlay-19 sub_2C69 |
| `overlay-21 entry#13` | `0A1Dh` | overlay-16 sub_26CF、overlay-19 sub_2FE5 |
| `overlay-21 entry#14` | `1070h` | overlay-04 sub_101A、overlay-04 sub_F42、overlay-05 sub_1072、overlay-05 sub_F0F、overlay-06 sub_778 |
| `overlay-21 entry#15` | `0148h` | overlay-04 sub_103、overlay-06 sub_4BE、overlay-19 sub_2A32 |
| `overlay-21 entry#16` | `019Dh` | overlay-04 sub_103、overlay-06 sub_4BE、overlay-19 sub_2A32 |
| `overlay-21 entry#17` | `00C7h` | overlay-04 sub_103、overlay-06 sub_4BE、overlay-19 sub_2A32 |
| `overlay-21 entry#18` | `10FCh` | overlay-02 sub_1C48 |
| `overlay-21 entry#19` | `1A3Eh` | overlay-04 sub_101A、overlay-06 sub_778 |
| `overlay-21 entry#20` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0、overlay-19 sub_0 |
| `overlay-22 entry#1` | `0227h` | overlay-19 sub_34E4 |
| `overlay-22 entry#2` | `0A69h` | overlay-19 sub_34E4 |
| `overlay-22 entry#3` | `0DCBh` | overlay-09 sub_3D3、overlay-13 sub_2049 |
| `overlay-22 entry#5` | `143Dh` | overlay-08 sub_414、overlay-09 sub_5D、overlay-13 sub_27A1、overlay-13 sub_426B、overlay-15 sub_4C2、overlay-19 sub_246F |
| `overlay-22 entry#6` | `64E8h` | overlay-09 sub_4CC、overlay-15 sub_1CB、overlay-15 sub_34、overlay-15 sub_B86、overlay-19 sub_17B7、overlay-19 sub_246F、overlay-19 sub_E8E、overlay-20 sub_A11 |
| `overlay-22 entry#7` | `6535h` | overlay-19 sub_246F、overlay-20 sub_A11 |
| `overlay-22 entry#8` | `65C9h` | overlay-20 sub_94E、overlay-20 sub_A11 |
| `overlay-22 entry#9` | `3879h` | overlay-04 sub_A33 |
| `overlay-22 entry#10` | `66EBh` | overlay-11 sub_48 |
| `overlay-22 entry#11` | `0000h` | overlay-04 sub_0、overlay-05 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-11 sub_0、overlay-13 sub_1DD7、overlay-15 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-20 sub_0 |
| `overlay-22 entry#12` | `1247h` | overlay-13 sub_29F4 |
| `overlay-23 entry#1` | `0016h` | overlay-12 sub_1ABC、overlay-12 sub_20AA、overlay-12 sub_2E15、overlay-12 sub_397、overlay-12 sub_54C、overlay-12 sub_8DF、overlay-13 sub_426B |
| `overlay-23 entry#2` | `00C9h` | overlay-12 sub_2AA4、overlay-12 sub_9B9、overlay-12 sub_A32、overlay-13 sub_3F71、overlay-13 sub_478B、overlay-19 sub_1A8F、overlay-22 sub_2147、overlay-22 sub_2776、overlay-22 sub_28DC、overlay-22 sub_3879、overlay-22 sub_45F1、overlay-22 sub_4A9A、overlay-22 sub_5147、overlay-22 sub_592D |
| `overlay-23 entry#3` | `010Eh` | overlay-04 sub_2A3、overlay-04 sub_35B、overlay-04 sub_48E、overlay-04 sub_6A8、overlay-04 sub_93E、overlay-05 sub_52F、overlay-09 sub_14F2、overlay-12 sub_218E、overlay-12 sub_26BD、overlay-12 sub_2AA4、overlay-12 sub_2E15、overlay-12 sub_54C、overlay-12 sub_773、overlay-12 sub_A32、overlay-12 sub_E3、overlay-13 sub_453D、overlay-13 sub_4671、overlay-19 sub_1A8F、overlay-19 sub_39CA、overlay-20 sub_20、overlay-22 sub_235E、overlay-22 sub_32A0、overlay-22 sub_345B、overlay-22 sub_4477、overlay-22 sub_47BD、overlay-22 sub_4BF7、overlay-22 sub_5B90 |
| `overlay-23 entry#4` | `03FEh` | overlay-08 sub_2BB、overlay-08 sub_9A5、overlay-09 sub_DD3、overlay-12 sub_2A14、overlay-13 sub_0、overlay-13 sub_11CC、overlay-13 sub_124、overlay-13 sub_1697、overlay-13 sub_1883、overlay-13 sub_192、overlay-13 sub_358、overlay-13 sub_E3E、overlay-22 sub_456C、overlay-22 sub_48AA、overlay-22 sub_F62 |
| `overlay-23 entry#5` | `0E32h` | overlay-08 sub_9A5、overlay-08 sub_C1D、overlay-09 sub_C4A、overlay-22 sub_2A52、overlay-22 sub_52E0 |
| `overlay-23 entry#6` | `11C4h` | overlay-02 sub_2ACE |
| `overlay-23 entry#7` | `122Ch` | overlay-13 sub_1697 |
| `overlay-23 entry#8` | `12D8h` | overlay-02 sub_2ACE、overlay-09 sub_2D3、overlay-09 sub_757、overlay-12 sub_147A、overlay-12 sub_1547、overlay-12 sub_1A34、overlay-12 sub_1ABC、overlay-12 sub_26BD、overlay-12 sub_27F0、overlay-12 sub_290C、overlay-12 sub_2E15、overlay-12 sub_A32、overlay-13 sub_426B、overlay-22 sub_1D0B、overlay-22 sub_235E、overlay-22 sub_3BF5、overlay-22 sub_496B、overlay-22 sub_4B2C、overlay-22 sub_4DB9、overlay-22 sub_5147、overlay-22 sub_5D17、overlay-22 sub_5E32、overlay-22 sub_6019、overlay-22 sub_61AD、overlay-22 sub_63D4、overlay-22 sub_F62 |
| `overlay-23 entry#9` | `1368h` | overlay-02 sub_1C48、overlay-02 sub_2ACE、overlay-04 sub_48E、overlay-07 sub_1F00、overlay-08 sub_1FB、overlay-09 sub_4CC、overlay-09 sub_5D、overlay-09 sub_627、overlay-09 sub_9E9、overlay-10 sub_9F2、overlay-10 sub_AFE、overlay-10 sub_C4A、overlay-10 sub_C95、overlay-10 sub_CC、overlay-12 sub_15D1、overlay-12 sub_1726、overlay-12 sub_17B5、overlay-12 sub_1C86、overlay-12 sub_218E、overlay-12 sub_2237、overlay-12 sub_2305、overlay-12 sub_2396、overlay-12 sub_2414、overlay-12 sub_26BD、overlay-12 sub_289D、overlay-12 sub_773、overlay-12 sub_A32、overlay-12 sub_F5B、overlay-13 sub_0、overlay-13 sub_12A8、overlay-13 sub_225F、overlay-13 sub_3D7F、overlay-13 sub_3F71、overlay-13 sub_D81、overlay-14 sub_2F5、overlay-14 sub_5B4、overlay-15 sub_1F69、overlay-15 sub_2020、overlay-15 sub_2095、overlay-17 sub_1627、overlay-17 sub_1683、overlay-17 sub_50DA、overlay-19 sub_246F、overlay-20 sub_CD0、overlay-21 sub_10C4、overlay-21 sub_10FC、overlay-22 sub_1FB6、overlay-22 sub_2404、overlay-22 sub_2547、overlay-22 sub_29A3、overlay-22 sub_2EB2、overlay-22 sub_3396、overlay-22 sub_345B、overlay-22 sub_39EF、overlay-22 sub_3F4E、overlay-22 sub_41F0、overlay-22 sub_423F、overlay-22 sub_42C4、overlay-22 sub_434D、overlay-22 sub_4687、overlay-22 sub_4B2C、overlay-22 sub_4F5F、overlay-22 sub_5B2B、overlay-22 sub_5B90、overlay-22 sub_5D17、overlay-22 sub_61AD、overlay-22 sub_E75 |
| `overlay-23 entry#10` | `13B3h` | overlay-12 sub_19B3、overlay-12 sub_19E8、overlay-12 sub_25BC、overlay-12 sub_26BD、overlay-12 sub_2868、overlay-13 sub_192、overlay-13 sub_426B、overlay-22 sub_1FF4、overlay-22 sub_244B、overlay-22 sub_24EF、overlay-22 sub_39EF、overlay-22 sub_43DA、overlay-22 sub_4416、overlay-22 sub_46CA、overlay-22 sub_477A、overlay-22 sub_525B、overlay-22 sub_5888、overlay-22 sub_5A61、overlay-22 sub_62D7 |
| `overlay-23 entry#11` | `13D7h` | overlay-12 sub_1091、overlay-12 sub_111B、overlay-12 sub_12DC、overlay-12 sub_131F、overlay-12 sub_1574、overlay-12 sub_1726、overlay-12 sub_218E、overlay-12 sub_2237、overlay-12 sub_2274、overlay-12 sub_27F0、overlay-12 sub_290C、overlay-12 sub_2AA4、overlay-12 sub_2CA6、overlay-12 sub_3C、overlay-13 sub_3F71、overlay-13 sub_478B、overlay-16 sub_2A6D、overlay-17 sub_AFB、overlay-17 sub_DEA、overlay-17 sub_E26、overlay-19 sub_37AB、overlay-19 sub_39CA、overlay-22 sub_222A、overlay-22 sub_2A52、overlay-22 sub_2EB2、overlay-22 sub_423F、overlay-22 sub_52E0、overlay-22 sub_63D4 |
| `overlay-23 entry#12` | `1486h` | overlay-09 sub_1529、overlay-13 sub_D81 |
| `overlay-23 entry#13` | `1552h` | overlay-22 sub_5B90、overlay-22 sub_62D7 |
| `overlay-23 entry#14` | `158Ah` | overlay-12 sub_2A14、overlay-13 sub_358 |
| `overlay-23 entry#15` | `15ECh` | overlay-09 sub_13A5 |
| `overlay-23 entry#16` | `1630h` | overlay-22 sub_3223、overlay-22 sub_32A0、overlay-22 sub_3879、overlay-22 sub_3AE5、overlay-22 sub_41AF |
| `overlay-23 entry#19` | `170Dh` | overlay-22 sub_222A、overlay-22 sub_2EB2、overlay-22 sub_423F |
| `overlay-23 entry#20` | `1FFDh` | overlay-12 sub_111B、overlay-12 sub_19B3、overlay-12 sub_19E8、overlay-12 sub_26BD、overlay-12 sub_2868、overlay-12 sub_2D56、overlay-12 sub_3DC、overlay-13 sub_426B、overlay-19 sub_1A8F、overlay-22 sub_3BF5、overlay-22 sub_5A61、overlay-22 sub_5D17、overlay-22 sub_5E32、overlay-22 sub_6019、overlay-22 sub_61AD、overlay-22 sub_F62 |
| `overlay-23 entry#21` | `2325h` | overlay-12 sub_A32、overlay-22 sub_1D0B、overlay-22 sub_3072、overlay-22 sub_4719、overlay-22 sub_496B、overlay-22 sub_4B2C、overlay-22 sub_4DB9、overlay-22 sub_4F5F、overlay-22 sub_5147、overlay-22 sub_F62 |
| `overlay-23 entry#22` | `2419h` | overlay-04 sub_48E、overlay-12 sub_1414、overlay-15 sub_22BC、overlay-19 sub_37AB、overlay-20 sub_898、overlay-22 sub_1FB6、overlay-22 sub_41F0、overlay-22 sub_434D、overlay-22 sub_4687、overlay-22 sub_5B2B |
| `overlay-23 entry#23` | `251Ah` | overlay-12 sub_1973、overlay-12 sub_218E、overlay-12 sub_22C6、overlay-22 sub_3072 |
| `overlay-23 entry#24` | `18F3h` | overlay-04 sub_48E、overlay-19 sub_1A8F、overlay-22 sub_222A、overlay-22 sub_2404、overlay-22 sub_2EB2、overlay-22 sub_3879、overlay-22 sub_423F |
| `overlay-23 entry#25` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-12 sub_0、overlay-13 sub_1DD7、overlay-14 sub_0、overlay-15 sub_0、overlay-16 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-20 sub_0、overlay-21 sub_0、overlay-22 sub_0 |
| `overlay-24 entry#1` | `046Ch` | overlay-02 sub_1C48、overlay-05 sub_39、overlay-05 sub_CE9、overlay-06 sub_37、overlay-06 sub_778、overlay-19 sub_174A、overlay-19 sub_17B7、overlay-19 sub_2055、overlay-19 sub_246F、overlay-19 sub_27DC、overlay-19 sub_2A32、overlay-19 sub_98 |
| `overlay-24 entry#2` | `0917h` | overlay-02 sub_2E8C、overlay-02 sub_2F5F、overlay-02 sub_306、overlay-02 sub_327E、overlay-02 sub_3393、overlay-02 sub_35AF、overlay-02 sub_3A21、overlay-02 sub_D2F、overlay-04 sub_11BD、overlay-04 sub_BFF、overlay-04 sub_F42、overlay-05 sub_F0F、overlay-06 sub_778、overlay-07 sub_1808、overlay-07 sub_2235、overlay-12 sub_111B、overlay-12 sub_3DC、overlay-14 sub_8E4、overlay-15 sub_14CA、overlay-15 sub_17F6、overlay-15 sub_19A4、overlay-15 sub_1D4C、overlay-15 sub_2370、overlay-15 sub_2457、overlay-17 sub_115、overlay-17 sub_28E1、overlay-17 sub_2CB、overlay-20 sub_898、overlay-23 sub_16 |
| `overlay-24 entry#3` | `0BD5h` | overlay-17 sub_2DB7、overlay-19 sub_69E |
| `overlay-24 entry#4` | `0B41h` | overlay-19 sub_69E |
| `overlay-24 entry#5` | `0C53h` | overlay-08 sub_2BB、overlay-08 sub_414、overlay-09 sub_1923、overlay-12 sub_829、overlay-13 sub_12A8、overlay-13 sub_1A59、overlay-13 sub_27A1、overlay-13 sub_2EBB、overlay-13 sub_329F、overlay-13 sub_999 |
| `overlay-24 entry#6` | `0E02h` | overlay-08 sub_C1D、overlay-09 sub_1223、overlay-13 sub_1697、overlay-13 sub_2049、overlay-13 sub_684、overlay-13 sub_78E、overlay-13 sub_999 |
| `overlay-24 entry#7` | `0E47h` | overlay-02 sub_2F5F、overlay-02 sub_369F、overlay-05 sub_1775、overlay-06 sub_3A0、overlay-08 sub_2BB、overlay-09 sub_1923、overlay-10 sub_10D1、overlay-12 sub_5BD、overlay-12 sub_829、overlay-13 sub_1697、overlay-13 sub_1A59、overlay-16 sub_36B2、overlay-17 sub_2E9B、overlay-19 sub_17B7、overlay-19 sub_2102、overlay-19 sub_32B7、overlay-19 sub_69E、overlay-22 sub_F62 |
| `overlay-24 entry#8` | `135Bh` | overlay-19 sub_98 |
| `overlay-24 entry#9` | `1321h` | overlay-02 sub_A5D、overlay-04 sub_103、overlay-13 sub_329F、overlay-19 sub_27DC、overlay-19 sub_5B1、overlay-19 sub_69E、overlay-19 sub_98 |
| `overlay-24 entry#10` | `12E5h` | overlay-08 sub_124D、overlay-09 sub_9E9、overlay-13 sub_2EBB、overlay-13 sub_358、overlay-15 sub_1084、overlay-15 sub_1B94、overlay-15 sub_5CC、overlay-16 sub_1627、overlay-16 sub_5008、overlay-16 sub_B4F、overlay-16 sub_ED3、overlay-17 sub_1627、overlay-17 sub_546F、overlay-19 sub_142D、overlay-19 sub_1437、overlay-19 sub_147D、overlay-19 sub_69E、overlay-19 sub_98、overlay-19 sub_9BC、overlay-20 sub_549、overlay-22 sub_621、overlay-23 sub_1FFD |
| `overlay-24 entry#11` | `1416h` | overlay-13 sub_0 |
| `overlay-24 entry#14` | `15D3h` | overlay-12 sub_17B5 |
| `overlay-24 entry#15` | `1659h` | overlay-19 sub_32B7、overlay-21 sub_1B |
| `overlay-24 entry#16` | `1739h` | overlay-12 sub_1627、overlay-13 sub_358、overlay-23 sub_1FFD |
| `overlay-24 entry#17` | `17D1h` | overlay-02 sub_369F、overlay-07 sub_1F00、overlay-12 sub_5BD、overlay-13 sub_1A59、overlay-19 sub_2102、overlay-19 sub_2275、overlay-19 sub_246F、overlay-19 sub_27DC、overlay-22 sub_6535 |
| `overlay-24 entry#18` | `18B8h` | overlay-12 sub_5BD、overlay-19 sub_2102 |
| `overlay-24 entry#19` | `194Ch` | overlay-04 sub_103、overlay-06 sub_3A0、overlay-06 sub_4BE、overlay-08 sub_11C7、overlay-08 sub_414、overlay-08 sub_9A5、overlay-08 sub_C1D、overlay-08 sub_F17、overlay-09 sub_1296、overlay-12 sub_1D5E、overlay-12 sub_C82、overlay-13 sub_12A8、overlay-13 sub_225F、overlay-13 sub_27A1、overlay-13 sub_4876、overlay-13 sub_D81、overlay-15 sub_B86、overlay-17 sub_1627、overlay-17 sub_28E1、overlay-17 sub_2C1C、overlay-17 sub_3BF4、overlay-17 sub_546F、overlay-19 sub_1EAE、overlay-19 sub_2095、overlay-19 sub_2102、overlay-19 sub_21C5、overlay-19 sub_27DC、overlay-19 sub_2A32、overlay-19 sub_37AB、overlay-19 sub_39CA、overlay-19 sub_E8E、overlay-21 sub_1A3E、overlay-21 sub_218、overlay-21 sub_485、overlay-21 sub_AA6 |
| `overlay-24 entry#20` | `19A8h` | overlay-04 sub_103、overlay-04 sub_51、overlay-09 sub_13A5、overlay-09 sub_5D、overlay-12 sub_1091、overlay-12 sub_14CA、overlay-12 sub_1627、overlay-12 sub_176D、overlay-12 sub_1ABC、overlay-12 sub_26BD、overlay-12 sub_27F0、overlay-12 sub_290C、overlay-12 sub_2D56、overlay-12 sub_2E15、overlay-12 sub_4E2、overlay-12 sub_5BD、overlay-12 sub_773、overlay-12 sub_829、overlay-12 sub_C0C、overlay-12 sub_F5B、overlay-13 sub_101A、overlay-13 sub_12A8、overlay-13 sub_1428、overlay-13 sub_27A1、overlay-13 sub_358、overlay-13 sub_3F71、overlay-13 sub_426B、overlay-13 sub_478B、overlay-15 sub_17F6、overlay-15 sub_19A4、overlay-15 sub_1D4C、overlay-15 sub_386、overlay-15 sub_4C2、overlay-15 sub_5CC、overlay-15 sub_92D、overlay-15 sub_B86、overlay-19 sub_246F、overlay-19 sub_39CA、overlay-22 sub_2147、overlay-22 sub_222A、overlay-22 sub_2A52、overlay-22 sub_3FEB、overlay-22 sub_423F、overlay-22 sub_4477、overlay-22 sub_45F1、overlay-22 sub_4BF7、overlay-22 sub_4DB9、overlay-22 sub_52E0、overlay-22 sub_5B90、overlay-22 sub_5D17、overlay-22 sub_5E32、overlay-22 sub_6019、overlay-22 sub_61AD、overlay-22 sub_62D7、overlay-22 sub_63D4、overlay-22 sub_65C9、overlay-23 sub_1486、overlay-23 sub_16、overlay-23 sub_1630、overlay-23 sub_1FFD、overlay-23 sub_2325、overlay-23 sub_251A、overlay-23 sub_E32 |
| `overlay-24 entry#21` | `1A90h` | overlay-04 sub_103、overlay-04 sub_51、overlay-09 sub_1223、overlay-09 sub_5D、overlay-12 sub_A32、overlay-12 sub_E3、overlay-13 sub_12A8、overlay-13 sub_358、overlay-15 sub_17F6、overlay-15 sub_2457、overlay-19 sub_246F、overlay-19 sub_39CA、overlay-20 sub_898、overlay-20 sub_CD0、overlay-22 sub_15A1、overlay-22 sub_2A52、overlay-22 sub_52E0、overlay-22 sub_65C9、overlay-23 sub_16、overlay-23 sub_1FFD、overlay-23 sub_2325 |
| `overlay-24 entry#22` | `1AC2h` | overlay-13 sub_358、overlay-17 sub_546F、overlay-19 sub_17B7、overlay-19 sub_34E4、overlay-19 sub_98、overlay-19 sub_E8E、overlay-22 sub_65C9 |
| `overlay-24 entry#23` | `1B2Fh` | overlay-13 sub_2A72、overlay-13 sub_3A74 |
| `overlay-24 entry#24` | `1C76h` | overlay-12 sub_1ABC、overlay-12 sub_26BD、overlay-12 sub_290C、overlay-13 sub_2A72、overlay-13 sub_40B8、overlay-22 sub_1D0B、overlay-22 sub_3BF5、overlay-22 sub_3CB1、overlay-22 sub_5B90、overlay-22 sub_5D17、overlay-22 sub_5E32、overlay-22 sub_6019、overlay-22 sub_61AD、overlay-22 sub_62D7、overlay-22 sub_63D4 |
| `overlay-24 entry#25` | `1CC3h` | overlay-12 sub_1ABC、overlay-12 sub_26BD、overlay-12 sub_290C、overlay-13 sub_2A72、overlay-13 sub_3889、overlay-13 sub_40B8、overlay-22 sub_1D0B、overlay-22 sub_3CB1、overlay-22 sub_5B90、overlay-22 sub_5D17、overlay-22 sub_5E32、overlay-22 sub_6019、overlay-22 sub_61AD、overlay-22 sub_62D7、overlay-22 sub_63D4 |
| `overlay-24 entry#26` | `240Ah` | overlay-12 sub_A32、overlay-12 sub_E3、overlay-13 sub_12A8、overlay-22 sub_3223、overlay-22 sub_345B、overlay-22 sub_3879、overlay-22 sub_434D、overlay-22 sub_5B2B、overlay-23 sub_1FFD、overlay-23 sub_2325 |
| `overlay-24 entry#27` | `25CDh` | overlay-02 sub_1547、overlay-02 sub_364B、overlay-04 sub_2A3、overlay-04 sub_35B、overlay-04 sub_93E、overlay-04 sub_A33、overlay-07 sub_20C5、overlay-09 sub_757、overlay-12 sub_1091、overlay-12 sub_111B、overlay-12 sub_16C2、overlay-12 sub_1ABC、overlay-12 sub_2274、overlay-12 sub_2305、overlay-12 sub_54C、overlay-12 sub_6F9、overlay-13 sub_3F71、overlay-13 sub_999、overlay-19 sub_3658、overlay-19 sub_39CA、overlay-19 sub_69E、overlay-22 sub_156A、overlay-22 sub_2147、overlay-22 sub_235E、overlay-22 sub_2547、overlay-22 sub_2776、overlay-22 sub_4477、overlay-22 sub_45F1、overlay-22 sub_4A9A、overlay-22 sub_4BF7、overlay-22 sub_5147、overlay-22 sub_592D、overlay-22 sub_90E、overlay-23 sub_155B、overlay-23 sub_158A、overlay-23 sub_1630、overlay-23 sub_19CA、overlay-23 sub_2325、overlay-23 sub_269、overlay-23 sub_E32 |
| `overlay-24 entry#28` | `2658h` | overlay-07 sub_2235、overlay-13 sub_358、overlay-23 sub_1FFD |
| `overlay-24 entry#29` | `2776h` | overlay-22 sub_1FB6、overlay-22 sub_41F0、overlay-22 sub_4687 |
| `overlay-24 entry#30` | `27E3h` | overlay-09 sub_2D3、overlay-09 sub_4CC、overlay-09 sub_627、overlay-13 sub_2D1E、overlay-13 sub_2DB9、overlay-22 sub_1F8B、overlay-22 sub_3FA8、overlay-22 sub_5E32 |
| `overlay-24 entry#31` | `2807h` | overlay-08 sub_143、overlay-08 sub_9A5、overlay-10 sub_17F4、overlay-12 sub_2C0、overlay-13 sub_12A8、overlay-13 sub_2DB9、overlay-23 sub_251A |
| `overlay-24 entry#32` | `285Bh` | overlay-09 sub_3D3、overlay-13 sub_14B4、overlay-13 sub_2EBB、overlay-13 sub_329F、overlay-13 sub_3D7F、overlay-13 sub_684、overlay-13 sub_999、overlay-13 sub_D81、overlay-13 sub_FB1、overlay-22 sub_1E8D |
| `overlay-24 entry#33` | `297Fh` | overlay-12 sub_26BD、overlay-12 sub_2D56、overlay-12 sub_F5B、overlay-13 sub_1D1C、overlay-13 sub_2EBB、overlay-13 sub_30D8、overlay-13 sub_426B、overlay-13 sub_FB1、overlay-22 sub_5D17、overlay-22 sub_61AD |
| `overlay-24 entry#34` | `2A5Bh` | overlay-08 sub_11C7、overlay-08 sub_414、overlay-08 sub_C1D、overlay-09 sub_1223、overlay-09 sub_5D、overlay-09 sub_C4A、overlay-09 sub_FE8、overlay-12 sub_26BD、overlay-12 sub_75、overlay-12 sub_E3、overlay-13 sub_12A8、overlay-13 sub_1923、overlay-13 sub_1A59、overlay-13 sub_27A1、overlay-13 sub_30D8、overlay-13 sub_453D、overlay-13 sub_4671、overlay-13 sub_4876、overlay-13 sub_D81、overlay-19 sub_246F、overlay-22 sub_5B90、overlay-22 sub_5E32、overlay-22 sub_6019、overlay-22 sub_61AD、overlay-22 sub_62D7、overlay-23 sub_1529 |
| `overlay-24 entry#35` | `2AAEh` | overlay-08 sub_11C7、overlay-09 sub_1223 |
| `overlay-24 entry#36` | `2AEAh` | overlay-12 sub_2396、overlay-13 sub_225F、overlay-22 sub_1D0B、overlay-22 sub_20ED、overlay-22 sub_2147、overlay-22 sub_222A、overlay-22 sub_244B、overlay-22 sub_24EF、overlay-22 sub_29A3、overlay-22 sub_2A52、overlay-22 sub_3072、overlay-22 sub_345B、overlay-22 sub_3804、overlay-22 sub_39EF、overlay-22 sub_3AE5、overlay-22 sub_3F4E、overlay-22 sub_45F1、overlay-22 sub_52E0、overlay-22 sub_5888、overlay-22 sub_DCB、overlay-22 sub_E75、overlay-22 sub_F62 |
| `overlay-24 entry#37` | `2CA4h` | overlay-02 sub_1820、overlay-02 sub_3237、overlay-02 sub_3393、overlay-02 sub_3520、overlay-04 sub_101A、overlay-04 sub_BFF、overlay-04 sub_F42、overlay-05 sub_1072、overlay-05 sub_DCC、overlay-05 sub_F0F、overlay-06 sub_778、overlay-15 sub_11C7、overlay-15 sub_1D4C、overlay-15 sub_2457、overlay-15 sub_4C2、overlay-15 sub_92D、overlay-15 sub_B86、overlay-18 sub_167E、overlay-19 sub_2102、overlay-19 sub_2C69、overlay-19 sub_37AB、overlay-19 sub_39CA、overlay-19 sub_B82、overlay-22 sub_112C |
| `overlay-24 entry#38` | `2E8Ch` | overlay-02 sub_2E8C、overlay-02 sub_30C6、overlay-02 sub_3393、overlay-02 sub_D2F、overlay-06 sub_778、overlay-14 sub_8E4、overlay-14 sub_BCC、overlay-15 sub_14CA、overlay-15 sub_2370、overlay-15 sub_2457、overlay-15 sub_7B4、overlay-17 sub_115 |
| `overlay-24 entry#39` | `30E7h` | overlay-08 sub_414、overlay-13 sub_27A1、overlay-19 sub_1A8F、overlay-19 sub_246F |
| `overlay-24 entry#40` | `314Bh` | overlay-02 sub_2EDA、overlay-17 sub_2CB、overlay-19 sub_2102、overlay-19 sub_2C69、overlay-19 sub_37AB、overlay-19 sub_39CA、overlay-22 sub_112C |
| `overlay-24 entry#41` | `3334h` | overlay-08 sub_1028、overlay-08 sub_F17、overlay-09 sub_1223、overlay-09 sub_FE8、overlay-13 sub_1D1C、overlay-13 sub_2EBB、overlay-13 sub_30D8、overlay-13 sub_329F、overlay-13 sub_684、overlay-13 sub_999、overlay-13 sub_E3E |
| `overlay-24 entry#42` | `3377h` | overlay-08 sub_1028、overlay-08 sub_F17、overlay-09 sub_119F、overlay-13 sub_1A59、overlay-13 sub_2EBB、overlay-13 sub_30D8、overlay-13 sub_329F、overlay-13 sub_684、overlay-13 sub_999 |
| `overlay-24 entry#43` | `33BCh` | overlay-09 sub_FE8、overlay-12 sub_2A14、overlay-13 sub_2EBB、overlay-13 sub_30D8、overlay-13 sub_329F、overlay-13 sub_E3E |
| `overlay-24 entry#44` | `347Eh` | overlay-22 sub_3FEB |
| `overlay-24 entry#45` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-12 sub_0、overlay-13 sub_1DD7、overlay-14 sub_0、overlay-15 sub_0、overlay-16 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-20 sub_0、overlay-21 sub_0、overlay-22 sub_0、overlay-23 sub_0 |
| `overlay-24 entry#47` | `35D8h` | overlay-08 sub_11C7、overlay-08 sub_9A5、overlay-09 sub_DD3 |
| `overlay-25 entry#1` | `03B1h` | overlay-16 sub_36B2、overlay-16 sub_3E8B、overlay-17 sub_546F、overlay-22 sub_3FEB |
| `overlay-25 entry#2` | `0671h` | overlay-17 sub_1627、overlay-17 sub_2E9B |
| `overlay-25 entry#3` | `07E3h` | overlay-17 sub_11C7 |
| `overlay-25 entry#4` | `0AC4h` | overlay-17 sub_11C7、overlay-19 sub_1A8F |
| `overlay-25 entry#5` | `0F30h` | overlay-17 sub_2CB |
| `overlay-25 entry#6` | `14A2h` | overlay-17 sub_2CB |
| `overlay-25 entry#7` | `139Ah` | overlay-17 sub_2CB、overlay-19 sub_246F |
| `overlay-25 entry#8` | `13F6h` | overlay-19 sub_246F |
| `overlay-25 entry#9` | `1452h` | overlay-17 sub_1627、overlay-19 sub_98 |
| `overlay-25 entry#10` | `14C4h` | overlay-02 sub_12F7、overlay-07 sub_8BD、overlay-12 sub_2CC7、overlay-12 sub_8DF、overlay-13 sub_12A8、overlay-13 sub_192、overlay-13 sub_2915、overlay-14 sub_27B、overlay-19 sub_1A8F、overlay-19 sub_3658、overlay-19 sub_36BD、overlay-19 sub_37AB、overlay-22 sub_20、overlay-22 sub_3396、overlay-24 sub_2AEA、overlay-24 sub_E47 |
| `overlay-25 entry#11` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-12 sub_0、overlay-13 sub_1DD7、overlay-14 sub_0、overlay-16 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-22 sub_0、overlay-23 sub_0、overlay-24 sub_0 |
| `overlay-26 entry#1` | `0011h` | overlay-17 sub_152E、overlay-17 sub_3BF4、overlay-17 sub_AFB |
| `overlay-26 entry#2` | `08AEh` | overlay-07 sub_17A8、overlay-08 sub_124D、overlay-13 sub_2EBB、overlay-13 sub_329F、overlay-14 sub_8E4、overlay-14 sub_BCC、overlay-15 sub_1D4C、overlay-16 sub_42FE、overlay-17 sub_45D0、overlay-19 sub_B82、overlay-24 sub_314B |
| `overlay-26 entry#3` | `0858h` | overlay-02 sub_10C2、overlay-04 sub_BFF、overlay-05 sub_1072、overlay-05 sub_CE9、overlay-06 sub_37、overlay-06 sub_778、overlay-08 sub_1028、overlay-08 sub_124D、overlay-08 sub_795、overlay-08 sub_B20、overlay-13 sub_2EBB、overlay-13 sub_329F、overlay-14 sub_8E4、overlay-14 sub_BCC、overlay-15 sub_1B94、overlay-15 sub_1D4C、overlay-16 sub_42FE、overlay-17 sub_15A1、overlay-17 sub_2CB、overlay-17 sub_3BF4、overlay-17 sub_45D0、overlay-17 sub_AFB、overlay-19 sub_15A1、overlay-19 sub_2C69、overlay-19 sub_2FE5、overlay-19 sub_B82、overlay-22 sub_227、overlay-24 sub_314B、overlay-25 sub_F30 |
| `overlay-26 entry#5` | `0133h` | overlay-04 sub_101A、overlay-05 sub_1506、overlay-05 sub_A5C、overlay-05 sub_F0F、overlay-06 sub_778、overlay-07 sub_1808、overlay-08 sub_11C7、overlay-08 sub_124D、overlay-08 sub_C1D、overlay-13 sub_2EBB、overlay-13 sub_329F、overlay-14 sub_8E4、overlay-14 sub_BCC、overlay-15 sub_14CA、overlay-15 sub_17F6、overlay-15 sub_1D4C、overlay-15 sub_2457、overlay-16 sub_1627、overlay-16 sub_42FE、overlay-16 sub_46CE、overlay-16 sub_8E4、overlay-17 sub_2B27、overlay-17 sub_2E9B、overlay-17 sub_3BF4、overlay-17 sub_45D0、overlay-19 sub_B82、overlay-20 sub_73A、overlay-22 sub_4F5F、overlay-24 sub_314B |
| `overlay-26 entry#6` | `081Eh` | overlay-02 sub_222C、overlay-02 sub_2940、overlay-02 sub_2ACE、overlay-02 sub_3E0、overlay-02 sub_992、overlay-02 sub_9DD、overlay-07 sub_499、overlay-08 sub_414、overlay-09 sub_5D、overlay-10 sub_1C28、overlay-11 sub_508、overlay-13 sub_12A8、overlay-15 sub_2457、overlay-16 sub_1627、overlay-16 sub_42FE、overlay-16 sub_46CE、overlay-16 sub_49C6、overlay-16 sub_5008、overlay-17 sub_2CB、overlay-17 sub_3BF4、overlay-17 sub_45D0、overlay-18 sub_1213、overlay-18 sub_167E、overlay-20 sub_CD0、overlay-24 sub_314B |
| `overlay-26 entry#7` | `0ED3h` | overlay-04 sub_BFF、overlay-05 sub_CE9、overlay-06 sub_37、overlay-07 sub_188A、overlay-15 sub_11C7、overlay-17 sub_15A1、overlay-17 sub_2CB、overlay-17 sub_3BF4、overlay-17 sub_AFB、overlay-19 sub_17B7、overlay-19 sub_2C69、overlay-19 sub_2FE5、overlay-22 sub_227、overlay-25 sub_F30 |
| `overlay-26 entry#8` | `129Dh` | overlay-04 sub_103、overlay-04 sub_51、overlay-08 sub_9A5、overlay-08 sub_C1D、overlay-13 sub_2DB9、overlay-15 sub_19A4、overlay-15 sub_92D、overlay-15 sub_B86、overlay-16 sub_1627、overlay-16 sub_42FE、overlay-16 sub_46CE、overlay-17 sub_1627、overlay-17 sub_28E1、overlay-17 sub_2C1C、overlay-17 sub_2CB、overlay-17 sub_45D0、overlay-17 sub_546F、overlay-19 sub_19CA、overlay-19 sub_27DC、overlay-19 sub_2A32、overlay-19 sub_39CA、overlay-19 sub_E8E、overlay-20 sub_CD0、overlay-22 sub_1542 |
| `overlay-26 entry#9` | `131Fh` | overlay-04 sub_BFF、overlay-17 sub_14CA、overlay-17 sub_AFB |
| `overlay-26 entry#10` | `13C7h` | overlay-04 sub_BFF、overlay-15 sub_11C7、overlay-17 sub_11C7、overlay-17 sub_15F5、overlay-17 sub_1627、overlay-17 sub_2CB、overlay-17 sub_3BF4、overlay-17 sub_AFB、overlay-17 sub_FCA、overlay-19 sub_2C69、overlay-19 sub_2FE5、overlay-22 sub_227、overlay-25 sub_F30 |
| `overlay-26 entry#11` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-11 sub_0、overlay-13 sub_1DD7、overlay-14 sub_0、overlay-15 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-20 sub_0、overlay-21 sub_0、overlay-22 sub_0、overlay-24 sub_0、overlay-25 sub_0 |
| `overlay-27 entry#3` | `0007h` | overlay-26 sub_133 |
| `overlay-27 entry#4` | `0032h` | overlay-26 sub_133 |
| `overlay-27 entry#5` | `0063h` | overlay-26 sub_133 |
| `overlay-27 entry#6` | `0000h` | overlay-02 sub_1820 |
| `overlay-28 entry#1` | `0016h` | overlay-02 sub_2E06、overlay-02 sub_30C6、overlay-02 sub_3237、overlay-02 sub_327E、overlay-02 sub_3A21、overlay-02 sub_8A7、overlay-07 sub_68B、overlay-24 sub_2CA4 |
| `overlay-28 entry#2` | `0000h` | overlay-02 sub_0、overlay-07 sub_0、overlay-14 sub_0、overlay-15 sub_0、overlay-19 sub_0、overlay-24 sub_0 |
| `overlay-29 entry#1` | `000Ch` | overlay-02 sub_2E8C、overlay-02 sub_30C6、overlay-02 sub_8A7、overlay-07 sub_68B、overlay-18 sub_A6F、overlay-24 sub_2CA4、overlay-26 sub_133 |
| `overlay-29 entry#4` | `010Bh` | overlay-02 sub_8A7、overlay-07 sub_68B、overlay-15 sub_2457、overlay-18 sub_A6F、overlay-24 sub_2CA4 |
| `overlay-29 entry#5` | `045Ah` | overlay-02 sub_327E、overlay-10 sub_1C28、overlay-14 sub_BCC、overlay-15 sub_2457、overlay-18 sub_A6F |
| `overlay-29 entry#6` | `05F9h` | overlay-07 sub_64C、overlay-24 sub_2CA4 |
| `overlay-29 entry#7` | `04E0h` | overlay-07 sub_64C、overlay-24 sub_2CA4 |
| `overlay-29 entry#8` | `066Ch` | overlay-07 sub_68B |
| `overlay-29 entry#9` | `072Bh` | overlay-02 sub_1820、overlay-02 sub_8A7、overlay-02 sub_CBF、overlay-16 sub_49C6、overlay-18 sub_167E、overlay-28 sub_16 |
| `overlay-29 entry#10` | `0777h` | overlay-02 sub_8A7、overlay-18 sub_167E、overlay-28 sub_16 |
| `overlay-29 entry#11` | `0000h` | overlay-02 sub_0、overlay-07 sub_0、overlay-10 sub_0、overlay-11 sub_0、overlay-14 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-24 sub_0、overlay-26 sub_0、overlay-28 sub_0 |
| `overlay-30 entry#2` | `016Bh` | overlay-02 sub_3393 |
| `overlay-30 entry#4` | `060Dh` | overlay-02 sub_30C6、overlay-07 sub_1AEC、overlay-07 sub_5BD、overlay-10 sub_2F6、overlay-14 sub_807、overlay-14 sub_8E4 |
| `overlay-30 entry#5` | `04DEh` | overlay-10 sub_2F6、overlay-14 sub_2F5、overlay-14 sub_759、overlay-14 sub_BCC |
| `overlay-30 entry#6` | `0710h` | overlay-02 sub_30C6、overlay-02 sub_327E、overlay-07 sub_1AEC、overlay-10 sub_8BD、overlay-14 sub_807、overlay-28 sub_16 |
| `overlay-30 entry#7` | `078Bh` | overlay-14 sub_8E4、overlay-28 sub_16 |
| `overlay-30 entry#8` | `0F8Fh` | overlay-02 sub_D2F、overlay-07 sub_C28、overlay-16 sub_49C6 |
| `overlay-30 entry#9` | `1253h` | overlay-02 sub_CBF、overlay-16 sub_49C6 |
| `overlay-30 entry#11` | `0000h` | overlay-02 sub_1820、overlay-02 sub_3393、overlay-09 sub_0、overlay-10 sub_0 |
| `overlay-31 entry#1` | `0007h` | overlay-22 sub_5B90、overlay-24 sub_1CC3、overlay-24 sub_2095、overlay-24 sub_224A |
| `overlay-31 entry#2` | `019Dh` | overlay-22 sub_3CB1、overlay-24 sub_1CC3 |
| `overlay-31 entry#3` | `0246h` | overlay-22 sub_18E4、overlay-22 sub_19B4、overlay-22 sub_3CB1、overlay-24 sub_1CC3 |
| `overlay-31 entry#4` | `054Ah` | overlay-13 sub_999 |
| `overlay-31 entry#5` | `03EAh` | overlay-09 sub_FE8、overlay-13 sub_329F、overlay-13 sub_412A、overlay-22 sub_19B4 |
| `overlay-31 entry#6` | `08D8h` | overlay-09 sub_2D3、overlay-12 sub_184F、overlay-12 sub_2BB6、overlay-13 sub_225F、overlay-13 sub_37E8、overlay-22 sub_39EF、overlay-22 sub_4BF7、overlay-23 sub_269、overlay-24 sub_285B、overlay-24 sub_297F |
| `overlay-31 entry#7` | `0000h` | overlay-02 sub_1820、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-13 sub_1DD7 |
| `overlay-32 entry#2` | `0011h` | overlay-31 sub_8D8 |
| `overlay-32 entry#4` | `00F4h` | overlay-17 sub_45D0、overlay-19 sub_B82、overlay-24 sub_30E7 |
| `overlay-32 entry#5` | `0111h` | overlay-08 sub_4D、overlay-17 sub_45D0、overlay-19 sub_B82 |
| `overlay-32 entry#7` | `0473h` | overlay-08 sub_2BB、overlay-13 sub_225F、overlay-13 sub_329F、overlay-13 sub_3889、overlay-13 sub_3A74 |
| `overlay-32 entry#8` | `06D7h` | overlay-10 sub_1900、overlay-13 sub_78E、overlay-23 sub_1529 |
| `overlay-32 entry#9` | `07F5h` | overlay-13 sub_1E30、overlay-13 sub_329F、overlay-22 sub_18E4、overlay-22 sub_19B4、overlay-22 sub_2A52、overlay-22 sub_345B、overlay-22 sub_3BF5、overlay-22 sub_3CB1、overlay-22 sub_52E0 |
| `overlay-32 entry#10` | `0864h` | overlay-13 sub_78E、overlay-22 sub_4BF7、overlay-23 sub_1510 |
| `overlay-32 entry#11` | `0A1Eh` | overlay-13 sub_78E、overlay-24 sub_1CC3、overlay-24 sub_224A |
| `overlay-32 entry#12` | `0A4Fh` | overlay-08 sub_2BB、overlay-09 sub_C4A、overlay-13 sub_1A59、overlay-24 sub_240A |
| `overlay-32 entry#13` | `0CCBh` | overlay-08 sub_9A5、overlay-09 sub_FE8、overlay-13 sub_30D8、overlay-13 sub_329F、overlay-13 sub_3A74、overlay-13 sub_4876、overlay-13 sub_684、overlay-13 sub_78E、overlay-22 sub_2A52、overlay-22 sub_39EF、overlay-22 sub_4BF7、overlay-22 sub_52E0、overlay-24 sub_1CC3、overlay-24 sub_2095、overlay-24 sub_224A、overlay-24 sub_240A、overlay-24 sub_30E7 |
| `overlay-32 entry#14` | `11F4h` | overlay-08 sub_C1D、overlay-09 sub_C4A、overlay-13 sub_1A59 |
| `overlay-32 entry#15` | `12EBh` | overlay-08 sub_2BB、overlay-08 sub_B20、overlay-09 sub_DD3、overlay-09 sub_FE8、overlay-10 sub_1C28、overlay-12 sub_184F、overlay-12 sub_1ABC、overlay-12 sub_26BD、overlay-12 sub_290C、overlay-12 sub_2BB6、overlay-13 sub_1E30、overlay-13 sub_2049、overlay-13 sub_225F、overlay-13 sub_29F4、overlay-13 sub_2A72、overlay-13 sub_30D8、overlay-13 sub_329F、overlay-13 sub_37E8、overlay-13 sub_3A74、overlay-13 sub_40B8、overlay-13 sub_412A、overlay-13 sub_684、overlay-13 sub_999、overlay-22 sub_1D0B、overlay-22 sub_3072、overlay-22 sub_3CB1、overlay-22 sub_4BF7、overlay-22 sub_4DB9、overlay-22 sub_5888、overlay-22 sub_5A61、overlay-22 sub_5B90、overlay-22 sub_5D17、overlay-22 sub_5E32、overlay-22 sub_6019、overlay-22 sub_61AD、overlay-22 sub_62D7、overlay-22 sub_63D4、overlay-23 sub_251A、overlay-23 sub_269、overlay-23 sub_E32、overlay-24 sub_240A、overlay-24 sub_285B、overlay-24 sub_297F |
| `overlay-32 entry#16` | `1313h` | overlay-08 sub_2BB、overlay-08 sub_B20、overlay-09 sub_DD3、overlay-09 sub_FE8、overlay-12 sub_184F、overlay-12 sub_1ABC、overlay-12 sub_26BD、overlay-12 sub_290C、overlay-12 sub_2BB6、overlay-13 sub_1E30、overlay-13 sub_2049、overlay-13 sub_225F、overlay-13 sub_29F4、overlay-13 sub_2A72、overlay-13 sub_30D8、overlay-13 sub_329F、overlay-13 sub_37E8、overlay-13 sub_3A74、overlay-13 sub_40B8、overlay-13 sub_412A、overlay-13 sub_684、overlay-13 sub_999、overlay-22 sub_1679、overlay-22 sub_1D0B、overlay-22 sub_3072、overlay-22 sub_3CB1、overlay-22 sub_4BF7、overlay-22 sub_4DB9、overlay-22 sub_5888、overlay-22 sub_5A61、overlay-22 sub_5B90、overlay-22 sub_5D17、overlay-22 sub_5E32、overlay-22 sub_6019、overlay-22 sub_61AD、overlay-22 sub_62D7、overlay-22 sub_63D4、overlay-23 sub_251A、overlay-23 sub_269、overlay-23 sub_E32、overlay-24 sub_240A、overlay-24 sub_285B、overlay-24 sub_297F |
| `overlay-32 entry#17` | `133Bh` | overlay-12 sub_184F、overlay-12 sub_2BB6、overlay-13 sub_37E8、overlay-23 sub_269、overlay-24 sub_285B、overlay-24 sub_297F |
| `overlay-32 entry#18` | `1363h` | overlay-13 sub_101A、overlay-13 sub_3F71、overlay-13 sub_478B、overlay-13 sub_4876、overlay-13 sub_78E、overlay-13 sub_999、overlay-22 sub_4BF7、overlay-23 sub_1486、overlay-23 sub_269、overlay-24 sub_240A |
| `overlay-32 entry#19` | `13BBh` | overlay-08 sub_C1D、overlay-09 sub_757、overlay-10 sub_1220、overlay-23 sub_E32 |
| `overlay-32 entry#20` | `1526h` | overlay-12 sub_2A14、overlay-13 sub_358、overlay-23 sub_16、overlay-23 sub_1FFD |
| `overlay-32 entry#21` | `1813h` | overlay-22 sub_3072、overlay-22 sub_4BF7、overlay-23 sub_251A |
| `overlay-32 entry#22` | `1A0Dh` | overlay-08 sub_11C7、overlay-08 sub_2BB、overlay-08 sub_414、overlay-13 sub_12A8、overlay-13 sub_27A1、overlay-13 sub_2EBB、overlay-23 sub_1486、overlay-23 sub_251A |
| `overlay-32 entry#23` | `0000h` | overlay-02 sub_1820、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-13 sub_1DD7、overlay-17 sub_0 |
| `overlay-33 entry#2` | `004Ch` | overlay-03 sub_131、overlay-10 sub_1023 |
| `overlay-33 entry#3` | `0267h` | overlay-03 sub_131、overlay-32 sub_12E、overlay-32 sub_2DA、overlay-32 sub_864、overlay-32 sub_B09 |
| `overlay-33 entry#4` | `02E8h` | overlay-16 sub_1061、overlay-17 sub_4122、overlay-17 sub_45D0 |
| `overlay-33 entry#5` | `038Fh` | overlay-02 sub_49C、overlay-07 sub_1B89、overlay-11 sub_508、overlay-16 sub_1061、overlay-16 sub_3DFF、overlay-16 sub_49C6 |
| `overlay-33 entry#6` | `07A9h` | overlay-32 sub_11F4、overlay-32 sub_12E、overlay-32 sub_2DA、overlay-32 sub_473、overlay-32 sub_CCB |
| `overlay-33 entry#7` | `0000h` | overlay-02 sub_0、overlay-05 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-10 sub_0、overlay-11 sub_0、overlay-13 sub_1DD7、overlay-14 sub_0、overlay-16 sub_0、overlay-17 sub_0、overlay-24 sub_0、overlay-28 sub_0、overlay-32 sub_0 |
| `overlay-34 entry#2` | `0000h` | overlay-02 sub_0、overlay-03 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-11 sub_0、overlay-12 sub_0、overlay-13 sub_1DD7、overlay-14 sub_0、overlay-15 sub_0、overlay-16 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-20 sub_0、overlay-21 sub_0、overlay-22 sub_0、overlay-23 sub_0、overlay-24 sub_0、overlay-25 sub_0、overlay-26 sub_0、overlay-28 sub_0、overlay-29 sub_0、overlay-30 sub_0、overlay-32 sub_0、overlay-33 sub_0 |
| `overlay-35 entry#2` | `004Ch` | overlay-11 sub_48、overlay-30 sub_F8F |
| `overlay-35 entry#3` | `032Ah` | overlay-30 sub_11、overlay-30 sub_39F |
| `overlay-35 entry#4` | `0000h` | overlay-11 sub_0、overlay-30 sub_0 |
