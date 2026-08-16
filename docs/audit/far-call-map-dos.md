# Overlay far call → `overlay-NN entry#K` 對照

由 `cmd/far-call-map` 產生，不要手改。

IDA 把每個 overlay 的 `.bin` 從位址 0 載入，所以 `call far ptr` 的目標會被「解析」成同一個 `.bin` 裡的標籤（`loc_1584+2`、`sub_1590` 之類）。**那些標籤與真正的目標無關**，規格裡的 `<far 1590h>` 就是這樣來的。

真正的目標是常駐段裡的 TPOV 控制記錄：段選 overlay、位移選 entry（`(位移 − 20h) / 5`，stub 每筆 5 bytes）。段與檔案位移差 **0x7b 個段**（＝ EXE header 的長度），這個常數是**量出來的**：掃過全部 far call、對每個候選值數有幾個目標正好落在 stub 邊界，取命中最多的那個。

| 指標 | 數字 |
|---|---:|
| far call 總數 | 5079 |
| 解出 `overlay-NN entry#K` | 2267 |
| 目標是常駐程式碼（段不是任何控制記錄）| 2812 |
| 段對得上但位移不在 stub 邊界 | 0 |

## 每個 overlay 的 entry 被叫了幾次

`0` 代表沒有任何 overlay 呼叫它——那種 entry 只會被常駐程式碼或 ECL 叫到。

| overlay | entry 數 | 有呼叫端 | 沒有呼叫端 |
|---|---:|---:|---:|
| `overlay-00` | 3 | 0 | 3 |
| `overlay-01` | 5 | 0 | 5 |
| `overlay-02` | 60 | 0 | 60 |
| `overlay-03` | 4 | 2 | 2 |
| `overlay-04` | 13 | 2 | 11 |
| `overlay-05` | 13 | 2 | 11 |
| `overlay-06` | 6 | 3 | 3 |
| `overlay-07` | 34 | 24 | 10 |
| `overlay-08` | 14 | 2 | 12 |
| `overlay-09` | 16 | 2 | 14 |
| `overlay-10` | 25 | 2 | 23 |
| `overlay-11` | 4 | 0 | 4 |
| `overlay-12` | 136 | 1 | 135 |
| `overlay-13` | 45 | 20 | 25 |
| `overlay-14` | 15 | 3 | 12 |
| `overlay-15` | 31 | 3 | 28 |
| `overlay-16` | 21 | 11 | 10 |
| `overlay-17` | 20 | 6 | 14 |
| `overlay-18` | 12 | 1 | 11 |
| `overlay-19` | 28 | 13 | 15 |
| `overlay-20` | 17 | 3 | 14 |
| `overlay-21` | 23 | 16 | 7 |
| `overlay-22` | 116 | 9 | 107 |
| `overlay-23` | 30 | 23 | 7 |
| `overlay-24` | 53 | 45 | 8 |
| `overlay-25` | 14 | 11 | 3 |
| `overlay-26` | 21 | 8 | 13 |
| `overlay-27` | 8 | 4 | 4 |
| `overlay-28` | 6 | 2 | 4 |
| `overlay-29` | 12 | 9 | 3 |
| `overlay-30` | 15 | 8 | 7 |
| `overlay-31` | 9 | 7 | 2 |
| `overlay-32` | 25 | 20 | 5 |
| `overlay-33` | 8 | 6 | 2 |
| `overlay-34` | 4 | 2 | 2 |
| `overlay-35` | 5 | 3 | 2 |

## 每個 entry 的呼叫端

讀某一支 overlay 函式時要問的是「這個 far call 打到哪」——那個方向逐筆列在 JSON 裡（每個呼叫點一列）。這裡是**反向**：某個 entry 是被誰叫的，用來回答「改這支會影響誰」。

| entry | code | 呼叫端 |
|---|---|---|
| `overlay-03 entry#1` | `011Ah` | overlay-02 sub_321F |
| `overlay-03 entry#2` | `0000h` | overlay-02 sub_0 |
| `overlay-04 entry#1` | `0DD6h` | overlay-02 sub_184B |
| `overlay-04 entry#2` | `0000h` | overlay-02 sub_184B |
| `overlay-05 entry#1` | `1736h` | overlay-02 sub_184B、overlay-02 sub_19A3 |
| `overlay-05 entry#2` | `0000h` | overlay-02 sub_184B、overlay-02 sub_19A3 |
| `overlay-06 entry#1` | `06F6h` | overlay-02 sub_17E7 |
| `overlay-06 entry#2` | `0355h` | overlay-05 sub_DB1 |
| `overlay-06 entry#3` | `0000h` | overlay-02 sub_17E2、overlay-05 sub_0 |
| `overlay-07 entry#1` | `0173h` | overlay-02 sub_1082、overlay-02 sub_11E、overlay-02 sub_1416、overlay-02 sub_16D2、overlay-02 sub_19B、overlay-02 sub_1A9B、overlay-02 sub_1B53、overlay-02 sub_1F46、overlay-02 sub_2086、overlay-02 sub_244、overlay-02 sub_27A8、overlay-02 sub_2847、overlay-02 sub_2942、overlay-02 sub_2A2、overlay-02 sub_2CB5、overlay-02 sub_2DA9、overlay-02 sub_2E16、overlay-02 sub_2F0、overlay-02 sub_30DD、overlay-02 sub_3284、overlay-02 sub_32D8、overlay-02 sub_3CA、overlay-02 sub_466、overlay-02 sub_841、overlay-02 sub_9EA、overlay-02 sub_AD6、overlay-02 sub_BBB、overlay-02 sub_C15、overlay-02 sub_DA4、overlay-02 sub_DBF、overlay-02 sub_E13、overlay-02 sub_E71、overlay-02 sub_EBD |
| `overlay-07 entry#2` | `0034h` | overlay-02 sub_107、overlay-02 sub_1082、overlay-02 sub_11E、overlay-02 sub_1271、overlay-02 sub_1416、overlay-02 sub_1636、overlay-02 sub_16D2、overlay-02 sub_19B、overlay-02 sub_1A9B、overlay-02 sub_1B53、overlay-02 sub_1F46、overlay-02 sub_2086、overlay-02 sub_244、overlay-02 sub_27A8、overlay-02 sub_2847、overlay-02 sub_2942、overlay-02 sub_2A2、overlay-02 sub_2CB5、overlay-02 sub_2D5E、overlay-02 sub_2DA9、overlay-02 sub_2E16、overlay-02 sub_2F0、overlay-02 sub_2F02、overlay-02 sub_30DD、overlay-02 sub_321F、overlay-02 sub_3284、overlay-02 sub_32D8、overlay-02 sub_3CA、overlay-02 sub_466、overlay-02 sub_841、overlay-02 sub_92C、overlay-02 sub_972、overlay-02 sub_9EA、overlay-02 sub_AD6、overlay-02 sub_BBB、overlay-02 sub_C15、overlay-02 sub_DA4、overlay-02 sub_E13、overlay-02 sub_E71、overlay-02 sub_E8、overlay-02 sub_EBD、overlay-02 sub_F0F |
| `overlay-07 entry#3` | `01FCh` | overlay-02 sub_3772、overlay-02 sub_BBB |
| `overlay-07 entry#4` | `0380h` | overlay-02 sub_3772、overlay-02 sub_BBB |
| `overlay-07 entry#5` | `0456h` | overlay-02 sub_466 |
| `overlay-07 entry#6` | `04BDh` | overlay-02 sub_1956、overlay-02 sub_2086、overlay-02 sub_3CA |
| `overlay-07 entry#7` | `0552h` | overlay-02 sub_841 |
| `overlay-07 entry#8` | `0591h` | overlay-02 sub_2086、overlay-02 sub_3CA、overlay-02 sub_801 |
| `overlay-07 entry#9` | `06E8h` | overlay-02 sub_1082、overlay-02 sub_1271、overlay-02 sub_1416、overlay-02 sub_19B、overlay-02 sub_1A9B、overlay-02 sub_2086、overlay-02 sub_244、overlay-02 sub_27A8、overlay-02 sub_2A2、overlay-02 sub_2E16、overlay-02 sub_2F02、overlay-02 sub_92C、overlay-02 sub_972、overlay-02 sub_DBF、overlay-02 sub_E13、overlay-02 sub_E71、overlay-02 sub_E8、overlay-02 sub_EBD |
| `overlay-07 entry#15` | `0D70h` | overlay-02 sub_1082、overlay-02 sub_1271、overlay-02 sub_13B6、overlay-02 sub_16B8、overlay-02 sub_177A、overlay-02 sub_19B、overlay-02 sub_2086、overlay-02 sub_244、overlay-02 sub_27A8、overlay-02 sub_2A2、overlay-02 sub_2E16、overlay-02 sub_92C、overlay-02 sub_DBF、overlay-02 sub_E13、overlay-02 sub_E71 |
| `overlay-07 entry#16` | `0F3Ah` | overlay-02 sub_E13 |
| `overlay-07 entry#17` | `108Dh` | overlay-02 sub_2A2、overlay-02 sub_972 |
| `overlay-07 entry#20` | `17EAh` | overlay-02 sub_1082、overlay-02 sub_2086、overlay-02 sub_27A8、overlay-04 sub_DD6、overlay-06 sub_6F6 |
| `overlay-07 entry#22` | `197Bh` | overlay-02 sub_11E |
| `overlay-07 entry#23` | `1A51h` | overlay-02 sub_11E、overlay-02 sub_DBF |
| `overlay-07 entry#24` | `1AB9h` | overlay-02 sub_107、overlay-02 sub_1A9B |
| `overlay-07 entry#25` | `1BE0h` | overlay-02 sub_2F02 |
| `overlay-07 entry#26` | `1B3Fh` | overlay-02 sub_2F02 |
| `overlay-07 entry#27` | `1E16h` | overlay-02 sub_1F46 |
| `overlay-07 entry#28` | `1F45h` | overlay-02 sub_1F46 |
| `overlay-07 entry#29` | `1FF5h` | overlay-02 sub_B3B |
| `overlay-07 entry#30` | `2137h` | overlay-02 sub_2086 |
| `overlay-07 entry#31` | `2252h` | overlay-02 sub_2942 |
| `overlay-07 entry#32` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0 |
| `overlay-08 entry#1` | `00F3h` | overlay-02 sub_19A3 |
| `overlay-08 entry#2` | `0000h` | overlay-02 sub_1956 |
| `overlay-09 entry#1` | `004Dh` | overlay-08 sub_26B、overlay-08 sub_3D5 |
| `overlay-09 entry#2` | `0000h` | overlay-02 sub_1956、overlay-08 sub_0 |
| `overlay-10 entry#1` | `1C3Eh` | overlay-08 sub_F3 |
| `overlay-10 entry#2` | `0000h` | overlay-02 sub_1956、overlay-08 sub_0 |
| `overlay-12 entry#2` | `0000h` | overlay-11 sub_0 |
| `overlay-13 entry#1` | `0000h` | overlay-08 sub_F3 |
| `overlay-13 entry#5` | `074Fh` | overlay-08 sub_CDA、overlay-09 sub_CB7 |
| `overlay-13 entry#6` | `095Ah` | overlay-08 sub_CDA、overlay-09 sub_CB7 |
| `overlay-13 entry#7` | `0D1Ch` | overlay-08 sub_CDA、overlay-09 sub_9C7 |
| `overlay-13 entry#8` | `0DD9h` | overlay-08 sub_3D5、overlay-09 sub_1681、overlay-09 sub_1926 |
| `overlay-13 entry#10` | `0F46h` | overlay-08 sub_EE9、overlay-09 sub_DB1 |
| `overlay-13 entry#11` | `1144h` | overlay-09 sub_DB1 |
| `overlay-13 entry#12` | `1227h` | overlay-08 sub_3D5、overlay-09 sub_24D |
| `overlay-13 entry#13` | `1433h` | overlay-09 sub_24D |
| `overlay-13 entry#14` | `194Ah` | overlay-08 sub_EE9、overlay-09 sub_DB1 |
| `overlay-13 entry#15` | `19D8h` | overlay-08 sub_EE9、overlay-09 sub_DB1 |
| `overlay-13 entry#17` | `1DF6h` | overlay-09 sub_3B1 |
| `overlay-13 entry#19` | `275Ah` | overlay-08 sub_3D5、overlay-09 sub_605 |
| `overlay-13 entry#20` | `2F3Ch` | overlay-08 sub_EE9 |
| `overlay-13 entry#21` | `3B37h` | overlay-08 sub_3D5 |
| `overlay-13 entry#23` | `29A2h` | overlay-09 sub_9C7、overlay-09 sub_DB1 |
| `overlay-13 entry#25` | `2E22h` | overlay-08 sub_997、overlay-10 sub_1C3E |
| `overlay-13 entry#27` | `3E56h` | overlay-09 sub_4D、overlay-09 sub_C7B |
| `overlay-13 entry#33` | `48DCh` | overlay-08 sub_3D5、overlay-09 sub_1273 |
| `overlay-13 entry#34` | `1D9Dh` | overlay-02 sub_1956、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0 |
| `overlay-14 entry#1` | `0BAFh` | overlay-02 sub_3772 |
| `overlay-14 entry#2` | `090Ah` | overlay-02 sub_3772 |
| `overlay-14 entry#3` | `0000h` | overlay-02 sub_0、overlay-11 sub_0 |
| `overlay-15 entry#1` | `23FCh` | overlay-02 sub_3073 |
| `overlay-15 entry#2` | `04CBh` | overlay-14 sub_90A |
| `overlay-15 entry#3` | `0000h` | overlay-02 sub_3073、overlay-02 sub_30DD、overlay-02 sub_3772 |
| `overlay-16 entry#1` | `0453h` | overlay-17 sub_3680 |
| `overlay-16 entry#4` | `0A7Eh` | overlay-17 sub_3680、overlay-17 sub_3EE3 |
| `overlay-16 entry#5` | `0C7Eh` | overlay-17 sub_260C |
| `overlay-16 entry#6` | `0DEAh` | overlay-17 sub_1DD、overlay-17 sub_1ECD |
| `overlay-16 entry#7` | `228Eh` | overlay-17 sub_3680 |
| `overlay-16 entry#8` | `3254h` | overlay-07 sub_456 |
| `overlay-16 entry#9` | `351Ch` | overlay-02 sub_2DA9 |
| `overlay-16 entry#10` | `3748h` | overlay-17 sub_1DD |
| `overlay-16 entry#11` | `3EDDh` | overlay-00 sub_17、overlay-02 sub_30DD、overlay-15 sub_23FC、overlay-17 sub_1DD |
| `overlay-16 entry#12` | `35A8h` | overlay-17 sub_3680 |
| `overlay-16 entry#14` | `0000h` | overlay-02 sub_0、overlay-07 sub_0、overlay-10 sub_0、overlay-14 sub_0、overlay-15 sub_0、overlay-17 sub_0 |
| `overlay-17 entry#1` | `01DDh` | overlay-02 sub_30DD |
| `overlay-17 entry#3` | `3A56h` | overlay-02 sub_2F0、overlay-02 sub_3251、overlay-02 sub_3772、overlay-05 sub_1370、overlay-05 sub_53C、overlay-15 sub_199C |
| `overlay-17 entry#4` | `3EE3h` | overlay-15 sub_1D0A |
| `overlay-17 entry#5` | `0000h` | overlay-02 sub_30DD、overlay-10 sub_0 |
| `overlay-17 entry#6` | `4D2Ah` | overlay-16 sub_1F21、overlay-16 sub_228E |
| `overlay-17 entry#7` | `45E9h` | overlay-16 sub_1F21、overlay-16 sub_228E |
| `overlay-18 entry#1` | `10FFh` | overlay-02 sub_30DD |
| `overlay-19 entry#1` | `0083h` | overlay-17 sub_11F7、overlay-17 sub_1467、overlay-17 sub_14FF、overlay-17 sub_1ECD、overlay-17 sub_2868 |
| `overlay-19 entry#2` | `0698h` | overlay-17 sub_1ECD、overlay-17 sub_2868 |
| `overlay-19 entry#3` | `09B3h` | overlay-17 sub_1ECD、overlay-17 sub_2724 |
| `overlay-19 entry#4` | `0597h` | overlay-17 sub_1ECD |
| `overlay-19 entry#5` | `0B75h` | overlay-04 sub_DD6、overlay-06 sub_6F6、overlay-08 sub_3D5、overlay-14 sub_90A、overlay-15 sub_23FC、overlay-17 sub_1DD |
| `overlay-19 entry#6` | `1588h` | overlay-08 sub_3D5 |
| `overlay-19 entry#7` | `1EE9h` | overlay-09 sub_1681、overlay-09 sub_1926、overlay-12 sub_5B6 |
| `overlay-19 entry#8` | `248Dh` | overlay-09 sub_4AA |
| `overlay-19 entry#9` | `3258h` | overlay-06 sub_355 |
| `overlay-19 entry#10` | `32DCh` | overlay-04 sub_DD6、overlay-06 sub_6F6、overlay-14 sub_90A、overlay-15 sub_1522、overlay-15 sub_1834、overlay-15 sub_1D0A、overlay-15 sub_23FC、overlay-17 sub_1DD |
| `overlay-19 entry#11` | `33ABh` | overlay-04 sub_F3、overlay-06 sub_474、overlay-16 sub_1F21、overlay-17 sub_4D2A |
| `overlay-19 entry#12` | `3474h` | overlay-13 sub_275A、overlay-15 sub_4CB、overlay-15 sub_942、overlay-15 sub_B9B、overlay-17 sub_4D2A |
| `overlay-19 entry#18` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-13 sub_1D9D、overlay-14 sub_0、overlay-15 sub_0、overlay-17 sub_0 |
| `overlay-20 entry#2` | `03B9h` | overlay-02 sub_2CB5、overlay-08 sub_997、overlay-14 sub_83C、overlay-14 sub_90A |
| `overlay-20 entry#3` | `0C9Ch` | overlay-15 sub_2309、overlay-15 sub_7B2 |
| `overlay-20 entry#4` | `0000h` | overlay-02 sub_0、overlay-08 sub_0、overlay-14 sub_0、overlay-15 sub_0、overlay-19 sub_0 |
| `overlay-21 entry#4` | `0073h` | overlay-19 sub_27D6 |
| `overlay-21 entry#5` | `0527h` | overlay-04 sub_DD6、overlay-06 sub_6F6 |
| `overlay-21 entry#7` | `064Fh` | overlay-04 sub_DD6、overlay-06 sub_6F6 |
| `overlay-21 entry#8` | `0CF0h` | overlay-04 sub_DD6、overlay-05 sub_ED7、overlay-06 sub_6F6 |
| `overlay-21 entry#9` | `021Ch` | overlay-16 sub_1F21、overlay-16 sub_228E |
| `overlay-21 entry#10` | `0B91h` | overlay-19 sub_2C57、overlay-19 sub_2FA9 |
| `overlay-21 entry#11` | `02ADh` | overlay-19 sub_2C57、overlay-19 sub_2FA9 |
| `overlay-21 entry#12` | `046Bh` | overlay-19 sub_2C57 |
| `overlay-21 entry#13` | `0A03h` | overlay-16 sub_1F21、overlay-19 sub_2FA9 |
| `overlay-21 entry#14` | `0F5Bh` | overlay-04 sub_DD6、overlay-05 sub_106C、overlay-06 sub_6F6 |
| `overlay-21 entry#15` | `0149h` | overlay-04 sub_F3、overlay-06 sub_474、overlay-19 sub_2A42 |
| `overlay-21 entry#16` | `019Eh` | overlay-04 sub_F3、overlay-06 sub_474、overlay-19 sub_2A42 |
| `overlay-21 entry#17` | `00C7h` | overlay-04 sub_F3、overlay-06 sub_474、overlay-19 sub_2A42 |
| `overlay-21 entry#18` | `0FEDh` | overlay-02 sub_1B53 |
| `overlay-21 entry#19` | `198Ch` | overlay-04 sub_DD6、overlay-06 sub_6F6 |
| `overlay-21 entry#20` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0、overlay-19 sub_0 |
| `overlay-22 entry#1` | `0219h` | overlay-19 sub_3474 |
| `overlay-22 entry#2` | `09FCh` | overlay-19 sub_3474 |
| `overlay-22 entry#3` | `0D67h` | overlay-09 sub_3B1、overlay-13 sub_200F |
| `overlay-22 entry#5` | `1263h` | overlay-08 sub_3D5、overlay-09 sub_4D、overlay-13 sub_275A、overlay-13 sub_42FD、overlay-15 sub_4CB、overlay-19 sub_248D |
| `overlay-22 entry#6` | `612Eh` | overlay-09 sub_4AA、overlay-15 sub_1CB、overlay-15 sub_34、overlay-15 sub_B9B、overlay-19 sub_1554、overlay-19 sub_248D、overlay-19 sub_EC5、overlay-20 sub_9DF |
| `overlay-22 entry#7` | `617Bh` | overlay-19 sub_248D、overlay-20 sub_9DF |
| `overlay-22 entry#8` | `6209h` | overlay-20 sub_91D、overlay-20 sub_9DF |
| `overlay-22 entry#9` | `3599h` | overlay-04 sub_97B |
| `overlay-22 entry#11` | `0000h` | overlay-04 sub_0、overlay-05 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-11 sub_0、overlay-13 sub_1D9D、overlay-15 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-20 sub_0 |
| `overlay-23 entry#1` | `0016h` | overlay-12 sub_1A65、overlay-12 sub_203C、overlay-12 sub_2D7D、overlay-12 sub_3A0、overlay-12 sub_54A、overlay-12 sub_8CD、overlay-13 sub_42FD、overlay-22 sub_45B5 |
| `overlay-23 entry#2` | `00C9h` | overlay-12 sub_2A2C、overlay-12 sub_9A7、overlay-12 sub_A0E、overlay-13 sub_403F、overlay-13 sub_4811、overlay-19 sub_1B00、overlay-22 sub_24CD、overlay-22 sub_262F、overlay-22 sub_3599、overlay-22 sub_4289、overlay-22 sub_4311、overlay-22 sub_4791、overlay-22 sub_4DA0、overlay-22 sub_5590 |
| `overlay-23 entry#3` | `010Eh` | overlay-04 sub_280、overlay-04 sub_324、overlay-04 sub_43D、overlay-04 sub_63E、overlay-04 sub_892、overlay-05 sub_53C、overlay-12 sub_212C、overlay-12 sub_2657、overlay-12 sub_2A2C、overlay-12 sub_2D7D、overlay-12 sub_54A、overlay-12 sub_769、overlay-12 sub_A0E、overlay-12 sub_E8、overlay-13 sub_45CF、overlay-13 sub_4703、overlay-19 sub_1B00、overlay-19 sub_37EC、overlay-20 sub_20、overlay-22 sub_20E1、overlay-22 sub_2FD1、overlay-22 sub_317E、overlay-22 sub_4194、overlay-22 sub_44D3、overlay-22 sub_48E4、overlay-22 sub_57E3 |
| `overlay-23 entry#4` | `03FEh` | overlay-08 sub_26B、overlay-08 sub_997、overlay-09 sub_1388、overlay-09 sub_DB1、overlay-10 sub_1C3E、overlay-12 sub_299C、overlay-13 sub_0、overlay-13 sub_1144、overlay-13 sub_124、overlay-13 sub_192、overlay-13 sub_31D、overlay-13 sub_DD9、overlay-22 sub_3804、overlay-22 sub_4289、overlay-22 sub_45B5、overlay-22 sub_F06 |
| `overlay-23 entry#5` | `0E45h` | overlay-08 sub_997、overlay-08 sub_CDA、overlay-09 sub_CB7、overlay-22 sub_2799、overlay-22 sub_4F3E |
| `overlay-23 entry#6` | `11D7h` | overlay-02 sub_2942 |
| `overlay-23 entry#7` | `123Fh` | overlay-22 sub_F06 |
| `overlay-23 entry#8` | `12EBh` | overlay-02 sub_2942、overlay-09 sub_2B1、overlay-09 sub_735、overlay-12 sub_19EE、overlay-12 sub_1A65、overlay-12 sub_2657、overlay-12 sub_2782、overlay-12 sub_289C、overlay-12 sub_2D7D、overlay-12 sub_A0E、overlay-13 sub_42FD、overlay-22 sub_1ABF、overlay-22 sub_20E1、overlay-22 sub_3947、overlay-22 sub_45B5、overlay-22 sub_4672、overlay-22 sub_4822、overlay-22 sub_4A8F、overlay-22 sub_4DA0、overlay-22 sub_5974、overlay-22 sub_5AA8、overlay-22 sub_5C86、overlay-22 sub_5E11、overlay-22 sub_6022、overlay-22 sub_F06 |
| `overlay-23 entry#9` | `137Fh` | overlay-02 sub_16D2、overlay-02 sub_1B53、overlay-02 sub_2942、overlay-04 sub_43D、overlay-07 sub_1F45、overlay-08 sub_1AB、overlay-09 sub_4AA、overlay-09 sub_4D、overlay-09 sub_605、overlay-09 sub_9C7、overlay-09 sub_DB1、overlay-10 sub_9FE、overlay-10 sub_B0A、overlay-10 sub_CAD、overlay-10 sub_CC、overlay-12 sub_15AB、overlay-12 sub_16EB、overlay-12 sub_1771、overlay-12 sub_1C1C、overlay-12 sub_212C、overlay-12 sub_21D5、overlay-12 sub_22A3、overlay-12 sub_2334、overlay-12 sub_23B2、overlay-12 sub_2657、overlay-12 sub_282F、overlay-12 sub_769、overlay-12 sub_A0E、overlay-12 sub_F39、overlay-13 sub_0、overlay-13 sub_1227、overlay-13 sub_2220、overlay-13 sub_3E56、overlay-13 sub_403F、overlay-13 sub_D1C、overlay-14 sub_302、overlay-14 sub_5E8、overlay-15 sub_1EFD、overlay-15 sub_1FB9、overlay-17 sub_11F7、overlay-17 sub_162A、overlay-17 sub_1ECD、overlay-17 sub_4A12、overlay-19 sub_248D、overlay-20 sub_C9C、overlay-21 sub_FAF、overlay-21 sub_FED、overlay-22 sub_1263、overlay-22 sub_1D55、overlay-22 sub_2180、overlay-22 sub_22B4、overlay-22 sub_26F6、overlay-22 sub_2BF9、overlay-22 sub_30C0、overlay-22 sub_317E、overlay-22 sub_370E、overlay-22 sub_3C98、overlay-22 sub_3F23、overlay-22 sub_3F6F、overlay-22 sub_3FF4、overlay-22 sub_407A、overlay-22 sub_43A7、overlay-22 sub_4822、overlay-22 sub_4BEC、overlay-22 sub_578B、overlay-22 sub_57E3、overlay-22 sub_5974、overlay-22 sub_5E11、overlay-22 sub_E11 |
| `overlay-23 entry#10` | `13CAh` | overlay-12 sub_196D、overlay-12 sub_19A2、overlay-12 sub_255A、overlay-12 sub_2657、overlay-12 sub_27FA、overlay-13 sub_192、overlay-13 sub_42FD、overlay-22 sub_1D93、overlay-22 sub_21C7、overlay-22 sub_2260、overlay-22 sub_370E、overlay-22 sub_4103、overlay-22 sub_413F、overlay-22 sub_43EA、overlay-22 sub_4493、overlay-22 sub_45B5、overlay-22 sub_4EB4、overlay-22 sub_54EB、overlay-22 sub_56C4、overlay-22 sub_5F53 |
| `overlay-23 entry#11` | `13EEh` | overlay-12 sub_106F、overlay-12 sub_10F9、overlay-12 sub_12BA、overlay-12 sub_12FD、overlay-12 sub_16EB、overlay-12 sub_212C、overlay-12 sub_21D5、overlay-12 sub_2212、overlay-12 sub_2782、overlay-12 sub_289C、overlay-12 sub_2A2C、overlay-12 sub_2C2E、overlay-12 sub_3C、overlay-13 sub_403F、overlay-13 sub_4811、overlay-16 sub_228E、overlay-17 sub_782、overlay-19 sub_36D8、overlay-19 sub_37EC、overlay-22 sub_1FA0、overlay-22 sub_24CD、overlay-22 sub_2799、overlay-22 sub_2BF9、overlay-22 sub_3F6F、overlay-22 sub_4F3E、overlay-22 sub_6022 |
| `overlay-23 entry#12` | `149Dh` | overlay-13 sub_D1C |
| `overlay-23 entry#13` | `1569h` | overlay-13 sub_15A4、overlay-22 sub_1263、overlay-22 sub_149E、overlay-22 sub_57E3、overlay-22 sub_5F53 |
| `overlay-23 entry#14` | `15A1h` | overlay-12 sub_299C、overlay-13 sub_31D |
| `overlay-23 entry#15` | `1603h` | overlay-09 sub_1388 |
| `overlay-23 entry#16` | `1643h` | overlay-22 sub_2F5E、overlay-22 sub_2FD1、overlay-22 sub_3599、overlay-22 sub_3804、overlay-22 sub_3EE2 |
| `overlay-23 entry#19` | `1720h` | overlay-22 sub_1FA0、overlay-22 sub_2BF9、overlay-22 sub_3F6F |
| `overlay-23 entry#20` | `1FD6h` | overlay-12 sub_10F9、overlay-12 sub_196D、overlay-12 sub_19A2、overlay-12 sub_2657、overlay-12 sub_27FA、overlay-12 sub_2CD9、overlay-12 sub_3E5、overlay-13 sub_42FD、overlay-19 sub_1B00、overlay-22 sub_3947、overlay-22 sub_45B5、overlay-22 sub_56C4、overlay-22 sub_5974、overlay-22 sub_5AA8、overlay-22 sub_5C86、overlay-22 sub_5E11、overlay-22 sub_F06 |
| `overlay-23 entry#21` | `2307h` | overlay-12 sub_A0E、overlay-22 sub_1ABF、overlay-22 sub_2DB6、overlay-22 sub_4432、overlay-22 sub_4672、overlay-22 sub_4822、overlay-22 sub_4A8F、overlay-22 sub_4BEC、overlay-22 sub_4DA0、overlay-22 sub_F06 |
| `overlay-23 entry#22` | `23FBh` | overlay-04 sub_43D、overlay-12 sub_13CF、overlay-15 sub_2255、overlay-19 sub_36D8、overlay-20 sub_862、overlay-22 sub_1D55、overlay-22 sub_3F23、overlay-22 sub_407A、overlay-22 sub_43A7、overlay-22 sub_578B |
| `overlay-23 entry#23` | `24EDh` | overlay-12 sub_212C、overlay-12 sub_2264、overlay-22 sub_2DB6 |
| `overlay-23 entry#24` | `18D8h` | overlay-19 sub_1B00、overlay-22 sub_1FA0、overlay-22 sub_20E1、overlay-22 sub_2180、overlay-22 sub_2BF9、overlay-22 sub_3599、overlay-22 sub_3F6F、overlay-22 sub_44D3 |
| `overlay-23 entry#25` | `0000h` | overlay-02 sub_0、overlay-05 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-12 sub_0、overlay-13 sub_1D9D、overlay-14 sub_0、overlay-16 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-20 sub_0、overlay-21 sub_0、overlay-22 sub_0 |
| `overlay-24 entry#1` | `0467h` | overlay-02 sub_1B53、overlay-05 sub_39、overlay-05 sub_CE6、overlay-06 sub_50、overlay-06 sub_6F6、overlay-19 sub_1554、overlay-19 sub_16A9、overlay-19 sub_19EA、overlay-19 sub_1EE9、overlay-19 sub_248D、overlay-19 sub_27D6、overlay-19 sub_2A42、overlay-19 sub_83 |
| `overlay-24 entry#2` | `0795h` | overlay-02 sub_2D15、overlay-02 sub_2DA9、overlay-02 sub_2F0、overlay-02 sub_3251、overlay-02 sub_3691、overlay-02 sub_3772、overlay-02 sub_D4F、overlay-04 sub_1009、overlay-04 sub_B12、overlay-04 sub_DD6、overlay-06 sub_6F6、overlay-07 sub_2252、overlay-12 sub_10F9、overlay-12 sub_3E5、overlay-14 sub_90A、overlay-15 sub_1522、overlay-15 sub_1834、overlay-15 sub_199C、overlay-15 sub_1D0A、overlay-15 sub_2309、overlay-15 sub_23FC、overlay-17 sub_1DD、overlay-17 sub_260C、overlay-20 sub_862、overlay-23 sub_16 |
| `overlay-24 entry#3` | `09C1h` | overlay-17 sub_2724、overlay-19 sub_698 |
| `overlay-24 entry#4` | `0939h` | overlay-19 sub_698 |
| `overlay-24 entry#5` | `0A34h` | overlay-08 sub_26B、overlay-08 sub_3D5、overlay-09 sub_1681、overlay-09 sub_1926、overlay-12 sub_818、overlay-13 sub_1227、overlay-13 sub_19D8、overlay-13 sub_275A、overlay-13 sub_3040、overlay-13 sub_33AC、overlay-13 sub_95A |
| `overlay-24 entry#6` | `0BE3h` | overlay-08 sub_CDA、overlay-09 sub_1201、overlay-13 sub_200F、overlay-13 sub_666、overlay-13 sub_74F、overlay-13 sub_95A |
| `overlay-24 entry#7` | `0C28h` | overlay-02 sub_2DA9、overlay-02 sub_32D8、overlay-05 sub_1736、overlay-06 sub_355、overlay-08 sub_26B、overlay-09 sub_1681、overlay-09 sub_1926、overlay-10 sub_10DD、overlay-12 sub_5B6、overlay-12 sub_818、overlay-13 sub_19D8、overlay-16 sub_228E、overlay-17 sub_2868、overlay-19 sub_1A00、overlay-19 sub_213C、overlay-19 sub_3258、overlay-19 sub_698、overlay-22 sub_F06 |
| `overlay-24 entry#8` | `1141h` | overlay-19 sub_83 |
| `overlay-24 entry#9` | `1107h` | overlay-02 sub_9EA、overlay-04 sub_F3、overlay-13 sub_33AC、overlay-19 sub_27D6、overlay-19 sub_597、overlay-19 sub_698、overlay-19 sub_83 |
| `overlay-24 entry#10` | `10CBh` | overlay-08 sub_11DF、overlay-08 sub_B06、overlay-09 sub_9C7、overlay-13 sub_3040、overlay-13 sub_31D、overlay-15 sub_11DE、overlay-15 sub_1B4D、overlay-15 sub_5CA、overlay-16 sub_3F96、overlay-16 sub_643、overlay-16 sub_DEA、overlay-17 sub_1ECD、overlay-17 sub_4D2A、overlay-19 sub_1494、overlay-19 sub_698、overlay-19 sub_83、overlay-19 sub_9B3、overlay-20 sub_549、overlay-22 sub_5B4、overlay-23 sub_1FD6 |
| `overlay-24 entry#11` | `120Ah` | overlay-13 sub_0 |
| `overlay-24 entry#14` | `13FDh` | overlay-12 sub_1771 |
| `overlay-24 entry#15` | `148Fh` | overlay-19 sub_3258、overlay-21 sub_1B |
| `overlay-24 entry#16` | `1583h` | overlay-13 sub_31D、overlay-22 sub_1263、overlay-22 sub_149E、overlay-22 sub_15CC、overlay-23 sub_1FD6 |
| `overlay-24 entry#17` | `15FFh` | overlay-02 sub_32D8、overlay-07 sub_1F45、overlay-12 sub_5B6、overlay-13 sub_19D8、overlay-19 sub_1A00、overlay-19 sub_213C、overlay-19 sub_22A3、overlay-19 sub_248D、overlay-19 sub_27D6、overlay-22 sub_617B |
| `overlay-24 entry#18` | `16E6h` | overlay-12 sub_5B6、overlay-16 sub_228E、overlay-19 sub_213C |
| `overlay-24 entry#19` | `177Ah` | overlay-04 sub_F3、overlay-06 sub_355、overlay-06 sub_474、overlay-08 sub_3D5、overlay-08 sub_997、overlay-08 sub_CDA、overlay-08 sub_EE9、overlay-09 sub_1273、overlay-12 sub_1CF6、overlay-12 sub_C61、overlay-13 sub_1227、overlay-13 sub_2220、overlay-13 sub_275A、overlay-13 sub_48DC、overlay-13 sub_D1C、overlay-15 sub_B9B、overlay-17 sub_260C、overlay-17 sub_3680、overlay-17 sub_4D2A、overlay-19 sub_1554、overlay-19 sub_1EE9、overlay-19 sub_213C、overlay-19 sub_21F3、overlay-19 sub_27D6、overlay-19 sub_2A42、overlay-19 sub_36D8、overlay-19 sub_37EC、overlay-19 sub_EC5、overlay-21 sub_198C、overlay-21 sub_21C、overlay-21 sub_46B、overlay-21 sub_A8A、overlay-22 sub_1263 |
| `overlay-24 entry#20` | `17D6h` | overlay-04 sub_47、overlay-09 sub_1388、overlay-09 sub_4D、overlay-12 sub_106F、overlay-12 sub_14A3、overlay-12 sub_15E6、overlay-12 sub_1729、overlay-12 sub_18AE、overlay-12 sub_1A65、overlay-12 sub_2657、overlay-12 sub_2782、overlay-12 sub_289C、overlay-12 sub_2CD9、overlay-12 sub_2D7D、overlay-12 sub_4E4、overlay-12 sub_5B6、overlay-12 sub_769、overlay-12 sub_818、overlay-12 sub_BE9、overlay-12 sub_F39、overlay-13 sub_1227、overlay-13 sub_275A、overlay-13 sub_31D、overlay-13 sub_403F、overlay-13 sub_42FD、overlay-13 sub_4811、overlay-13 sub_F46、overlay-15 sub_1834、overlay-15 sub_199C、overlay-15 sub_4CB、overlay-15 sub_5CA、overlay-15 sub_942、overlay-15 sub_B9B、overlay-19 sub_248D、overlay-19 sub_37EC、overlay-22 sub_1FA0、overlay-22 sub_20E1、overlay-22 sub_2799、overlay-22 sub_3D27、overlay-22 sub_3F6F、overlay-22 sub_4194、overlay-22 sub_4311、overlay-22 sub_44D3、overlay-22 sub_45B5、overlay-22 sub_48E4、overlay-22 sub_4A8F、overlay-22 sub_4F3E、overlay-22 sub_57E3、overlay-22 sub_5974、overlay-22 sub_5AA8、overlay-22 sub_5C86、overlay-22 sub_5E11、overlay-22 sub_5F53、overlay-22 sub_6022、overlay-22 sub_6209、overlay-23 sub_149D、overlay-23 sub_16、overlay-23 sub_1643、overlay-23 sub_1FD6、overlay-23 sub_2307、overlay-23 sub_24ED、overlay-23 sub_E45 |
| `overlay-24 entry#21` | `18BEh` | overlay-04 sub_F3、overlay-09 sub_1201、overlay-09 sub_4D、overlay-12 sub_A0E、overlay-12 sub_E8、overlay-13 sub_1227、overlay-13 sub_31D、overlay-15 sub_1834、overlay-15 sub_1B4D、overlay-15 sub_23FC、overlay-19 sub_248D、overlay-19 sub_37EC、overlay-20 sub_862、overlay-20 sub_C9C、overlay-22 sub_1263、overlay-22 sub_2799、overlay-22 sub_4F3E、overlay-22 sub_6209、overlay-23 sub_16、overlay-23 sub_1FD6、overlay-23 sub_2307 |
| `overlay-24 entry#22` | `18F3h` | overlay-13 sub_31D、overlay-17 sub_4D2A、overlay-19 sub_1554、overlay-19 sub_16A9、overlay-19 sub_3474、overlay-19 sub_83、overlay-19 sub_EC5、overlay-22 sub_6209 |
| `overlay-24 entry#23` | `1975h` | overlay-13 sub_2BFB、overlay-13 sub_3B37 |
| `overlay-24 entry#24` | `1AAAh` | overlay-12 sub_1A65、overlay-12 sub_2657、overlay-12 sub_289C、overlay-13 sub_2BFB、overlay-13 sub_4162、overlay-22 sub_1263、overlay-22 sub_1ABF、overlay-22 sub_3947、overlay-22 sub_3A03、overlay-22 sub_57E3、overlay-22 sub_5974、overlay-22 sub_5AA8、overlay-22 sub_5C86、overlay-22 sub_5E11、overlay-22 sub_5F53、overlay-22 sub_6022 |
| `overlay-24 entry#25` | `1AF7h` | overlay-12 sub_1A65、overlay-12 sub_2657、overlay-12 sub_289C、overlay-13 sub_2BFB、overlay-13 sub_394B、overlay-13 sub_4162、overlay-22 sub_149E、overlay-22 sub_1ABF、overlay-22 sub_3A03、overlay-22 sub_57E3、overlay-22 sub_5974、overlay-22 sub_5AA8、overlay-22 sub_5C86、overlay-22 sub_5E11、overlay-22 sub_5F53、overlay-22 sub_6022 |
| `overlay-24 entry#26` | `21DDh` | overlay-12 sub_1545、overlay-12 sub_A0E、overlay-12 sub_E8、overlay-13 sub_1227、overlay-22 sub_2F5E、overlay-22 sub_317E、overlay-22 sub_3599、overlay-22 sub_407A、overlay-22 sub_578B、overlay-23 sub_1FD6、overlay-23 sub_2307 |
| `overlay-24 entry#27` | `2396h` | overlay-02 sub_1416、overlay-02 sub_3284、overlay-04 sub_280、overlay-04 sub_324、overlay-04 sub_892、overlay-04 sub_97B、overlay-07 sub_2137、overlay-09 sub_735、overlay-12 sub_106F、overlay-12 sub_10F9、overlay-12 sub_1687、overlay-12 sub_1A65、overlay-12 sub_2212、overlay-12 sub_22A3、overlay-12 sub_54A、overlay-12 sub_6F2、overlay-13 sub_403F、overlay-13 sub_95A、overlay-19 sub_35F1、overlay-19 sub_37EC、overlay-19 sub_698、overlay-22 sub_1263、overlay-22 sub_20E1、overlay-22 sub_22B4、overlay-22 sub_24CD、overlay-22 sub_4194、overlay-22 sub_4311、overlay-22 sub_4791、overlay-22 sub_48E4、overlay-22 sub_4DA0、overlay-22 sub_5590、overlay-22 sub_8A1、overlay-23 sub_157C、overlay-23 sub_15A1、overlay-23 sub_1643、overlay-23 sub_18D8、overlay-23 sub_2307、overlay-23 sub_269、overlay-23 sub_E45 |
| `overlay-24 entry#28` | `2421h` | overlay-07 sub_2252、overlay-13 sub_31D、overlay-23 sub_1FD6 |
| `overlay-24 entry#29` | `2543h` | overlay-12 sub_13CF、overlay-22 sub_1D55、overlay-22 sub_3F23、overlay-22 sub_43A7 |
| `overlay-24 entry#30` | `25B0h` | overlay-09 sub_2B1、overlay-09 sub_4AA、overlay-09 sub_605、overlay-13 sub_2EAC、overlay-13 sub_2F3C、overlay-22 sub_1D2A、overlay-22 sub_3CED、overlay-22 sub_5AA8 |
| `overlay-24 entry#31` | `25D4h` | overlay-08 sub_997、overlay-08 sub_F3、overlay-10 sub_1800、overlay-12 sub_2CB、overlay-13 sub_1227、overlay-13 sub_2F3C、overlay-23 sub_24ED |
| `overlay-24 entry#32` | `2628h` | overlay-09 sub_1681、overlay-09 sub_1926、overlay-09 sub_3B1、overlay-09 sub_DB1、overlay-13 sub_3040、overlay-13 sub_33AC、overlay-13 sub_3E56、overlay-13 sub_666、overlay-13 sub_95A、overlay-13 sub_D1C、overlay-13 sub_F46、overlay-22 sub_1C39 |
| `overlay-24 entry#33` | `274Ch` | overlay-09 sub_DB1、overlay-12 sub_2657、overlay-12 sub_2CD9、overlay-12 sub_F39、overlay-13 sub_1CE2、overlay-13 sub_3040、overlay-13 sub_31F0、overlay-13 sub_42FD、overlay-13 sub_F46、overlay-22 sub_5974、overlay-22 sub_5E11 |
| `overlay-24 entry#34` | `2828h` | overlay-08 sub_3D5、overlay-08 sub_CDA、overlay-09 sub_1201、overlay-09 sub_4D、overlay-09 sub_9C7、overlay-09 sub_CB7、overlay-09 sub_DB1、overlay-12 sub_2657、overlay-12 sub_75、overlay-12 sub_E8、overlay-13 sub_1227、overlay-13 sub_19D8、overlay-13 sub_275A、overlay-13 sub_31F0、overlay-13 sub_45CF、overlay-13 sub_4703、overlay-13 sub_48DC、overlay-13 sub_D1C、overlay-19 sub_248D、overlay-22 sub_57E3、overlay-22 sub_5AA8、overlay-22 sub_5C86、overlay-22 sub_5E11、overlay-22 sub_5F53、overlay-23 sub_1540 |
| `overlay-24 entry#35` | `2877h` | overlay-09 sub_1201 |
| `overlay-24 entry#36` | `28B3h` | overlay-12 sub_2334、overlay-13 sub_2220、overlay-22 sub_1ABF、overlay-22 sub_1E7C、overlay-22 sub_1FA0、overlay-22 sub_21C7、overlay-22 sub_2260、overlay-22 sub_26F6、overlay-22 sub_2799、overlay-22 sub_2DB6、overlay-22 sub_317E、overlay-22 sub_352D、overlay-22 sub_370E、overlay-22 sub_3804、overlay-22 sub_3C98、overlay-22 sub_4311、overlay-22 sub_4F3E、overlay-22 sub_54EB、overlay-22 sub_D67、overlay-22 sub_E11、overlay-22 sub_F06 |
| `overlay-24 entry#37` | `2A6Dh` | overlay-02 sub_1A4F、overlay-02 sub_3073、overlay-02 sub_30DD、overlay-02 sub_321F、overlay-02 sub_3772、overlay-04 sub_B12、overlay-04 sub_DD6、overlay-05 sub_106C、overlay-05 sub_DB1、overlay-05 sub_ED7、overlay-06 sub_6F6、overlay-15 sub_11DE、overlay-15 sub_1D0A、overlay-15 sub_23FC、overlay-15 sub_4CB、overlay-15 sub_942、overlay-15 sub_B9B、overlay-18 sub_10FF、overlay-19 sub_213C、overlay-19 sub_2C57、overlay-19 sub_36D8、overlay-19 sub_37EC、overlay-19 sub_B75、overlay-22 sub_10D2 |
| `overlay-24 entry#38` | `2BAAh` | overlay-02 sub_2D15、overlay-02 sub_2F02、overlay-02 sub_D4F、overlay-14 sub_90A、overlay-14 sub_BAF、overlay-15 sub_2309、overlay-15 sub_23FC |
| `overlay-24 entry#39` | `2E05h` | overlay-08 sub_3D5、overlay-10 sub_1C3E、overlay-13 sub_275A、overlay-19 sub_1B00、overlay-19 sub_248D |
| `overlay-24 entry#40` | `2E6Ah` | overlay-02 sub_2D5E、overlay-19 sub_213C、overlay-19 sub_2C57、overlay-19 sub_36D8、overlay-19 sub_37EC、overlay-22 sub_10D2 |
| `overlay-24 entry#41` | `3018h` | overlay-08 sub_EE9、overlay-08 sub_FE8、overlay-09 sub_1201、overlay-09 sub_DB1、overlay-13 sub_1CE2、overlay-13 sub_3040、overlay-13 sub_31F0、overlay-13 sub_33AC、overlay-13 sub_DD9 |
| `overlay-24 entry#42` | `305Bh` | overlay-08 sub_EE9、overlay-09 sub_DB1、overlay-13 sub_19D8、overlay-13 sub_3040、overlay-13 sub_31F0、overlay-13 sub_33AC |
| `overlay-24 entry#43` | `30A0h` | overlay-09 sub_DB1、overlay-12 sub_299C、overlay-12 sub_FAA、overlay-13 sub_3040、overlay-13 sub_31F0、overlay-13 sub_33AC、overlay-13 sub_DD9 |
| `overlay-24 entry#44` | `3162h` | overlay-22 sub_3D27 |
| `overlay-24 entry#45` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-12 sub_0、overlay-13 sub_1D9D、overlay-14 sub_0、overlay-15 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-20 sub_0、overlay-21 sub_0、overlay-22 sub_0、overlay-23 sub_0 |
| `overlay-24 entry#46` | `32A9h` | overlay-16 sub_228E |
| `overlay-24 entry#47` | `3387h` | overlay-08 sub_104A、overlay-08 sub_997、overlay-09 sub_DB1 |
| `overlay-25 entry#1` | `03B1h` | overlay-16 sub_228E、overlay-17 sub_4D2A、overlay-22 sub_3D27 |
| `overlay-25 entry#2` | `0671h` | overlay-17 sub_1ECD、overlay-17 sub_2868 |
| `overlay-25 entry#3` | `07E3h` | overlay-17 sub_782 |
| `overlay-25 entry#4` | `0AD8h` | overlay-17 sub_782、overlay-17 sub_E93、overlay-19 sub_1B00 |
| `overlay-25 entry#5` | `0EE1h` | overlay-17 sub_1DD |
| `overlay-25 entry#6` | `139Ah` | overlay-17 sub_1DD |
| `overlay-25 entry#7` | `1292h` | overlay-17 sub_1DD、overlay-19 sub_248D |
| `overlay-25 entry#8` | `12EEh` | overlay-19 sub_248D |
| `overlay-25 entry#9` | `134Ah` | overlay-17 sub_1ECD、overlay-19 sub_83 |
| `overlay-25 entry#10` | `13BCh` | overlay-02 sub_1271、overlay-07 sub_7F1、overlay-12 sub_2C4F、overlay-12 sub_8CD、overlay-13 sub_1227、overlay-13 sub_192、overlay-13 sub_28C3、overlay-14 sub_288、overlay-19 sub_1B00、overlay-19 sub_35F1、overlay-19 sub_3656、overlay-19 sub_36D8、overlay-22 sub_20、overlay-22 sub_30C0、overlay-24 sub_28B3、overlay-24 sub_C28 |
| `overlay-25 entry#11` | `0000h` | overlay-02 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-12 sub_0、overlay-13 sub_1D9D、overlay-14 sub_0、overlay-16 sub_0、overlay-19 sub_0、overlay-22 sub_0、overlay-23 sub_0、overlay-24 sub_0 |
| `overlay-26 entry#1` | `0011h` | overlay-17 sub_3680、overlay-17 sub_782 |
| `overlay-26 entry#3` | `03D4h` | overlay-04 sub_DD6、overlay-05 sub_15A9、overlay-05 sub_A7E、overlay-06 sub_6F6、overlay-08 sub_11DF、overlay-08 sub_76C、overlay-08 sub_B06、overlay-13 sub_3040、overlay-13 sub_33AC、overlay-14 sub_90A、overlay-14 sub_BAF、overlay-15 sub_1522、overlay-15 sub_1834、overlay-15 sub_1B4D、overlay-15 sub_1D0A、overlay-15 sub_23FC、overlay-16 sub_3748、overlay-16 sub_3EDD、overlay-16 sub_DEA、overlay-17 sub_1DD、overlay-17 sub_2868、overlay-17 sub_3680、overlay-17 sub_3EE3、overlay-19 sub_B75、overlay-20 sub_712、overlay-22 sub_4BEC、overlay-24 sub_2E6A |
| `overlay-26 entry#4` | `07E5h` | overlay-02 sub_2086、overlay-07 sub_380、overlay-08 sub_3D5、overlay-08 sub_76C、overlay-09 sub_4D、overlay-10 sub_1C3E、overlay-11 sub_6D8、overlay-13 sub_1227、overlay-15 sub_23FC、overlay-16 sub_3748、overlay-16 sub_3F96、overlay-17 sub_1DD、overlay-17 sub_3680、overlay-17 sub_3EE3、overlay-18 sub_10FF、overlay-20 sub_C9C |
| `overlay-26 entry#5` | `0DF9h` | overlay-04 sub_B12、overlay-05 sub_CE6、overlay-06 sub_50、overlay-07 sub_18EE、overlay-15 sub_11DE、overlay-17 sub_11F7、overlay-17 sub_3680、overlay-17 sub_782、overlay-19 sub_1554、overlay-19 sub_16A9、overlay-19 sub_2C57、overlay-19 sub_2FA9、overlay-21 sub_CF0、overlay-22 sub_219、overlay-25 sub_EE1 |
| `overlay-26 entry#6` | `112Ah` | overlay-02 sub_30DD、overlay-04 sub_F3、overlay-08 sub_997、overlay-08 sub_CDA、overlay-13 sub_2F3C、overlay-15 sub_199C、overlay-15 sub_23FC、overlay-15 sub_942、overlay-15 sub_B9B、overlay-16 sub_643、overlay-16 sub_DEA、overlay-17 sub_1ECD、overlay-17 sub_260C、overlay-17 sub_3EE3、overlay-17 sub_4D2A、overlay-19 sub_1A00、overlay-19 sub_27D6、overlay-19 sub_2A42、overlay-19 sub_37EC、overlay-19 sub_EC5、overlay-20 sub_C9C、overlay-22 sub_1263、overlay-22 sub_159A |
| `overlay-26 entry#7` | `1199h` | overlay-04 sub_B12、overlay-17 sub_782 |
| `overlay-26 entry#8` | `1241h` | overlay-04 sub_B12、overlay-15 sub_11DE、overlay-17 sub_11F7、overlay-17 sub_3680、overlay-17 sub_782、overlay-19 sub_2C57、overlay-19 sub_2FA9、overlay-21 sub_CF0、overlay-22 sub_219、overlay-25 sub_EE1 |
| `overlay-26 entry#9` | `0000h` | overlay-02 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-11 sub_0、overlay-13 sub_1D9D、overlay-14 sub_0、overlay-15 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-20 sub_0、overlay-21 sub_0、overlay-22 sub_0、overlay-24 sub_0 |
| `overlay-27 entry#3` | `0007h` | overlay-26 sub_3D4 |
| `overlay-27 entry#4` | `0032h` | overlay-26 sub_3D4 |
| `overlay-27 entry#5` | `0063h` | overlay-26 sub_3D4 |
| `overlay-27 entry#6` | `0000h` | overlay-02 sub_17E7、overlay-02 sub_184B、overlay-02 sub_19EF |
| `overlay-28 entry#1` | `00CCh` | overlay-02 sub_2C8F、overlay-02 sub_2F02、overlay-02 sub_3073、overlay-02 sub_3691、overlay-02 sub_3772、overlay-02 sub_841、overlay-07 sub_591、overlay-24 sub_2A6D |
| `overlay-28 entry#2` | `0000h` | overlay-02 sub_0、overlay-07 sub_0、overlay-14 sub_0、overlay-19 sub_0、overlay-24 sub_0 |
| `overlay-29 entry#1` | `000Ch` | overlay-02 sub_2D15、overlay-02 sub_2F02、overlay-02 sub_841、overlay-07 sub_591、overlay-18 sub_B5F、overlay-24 sub_2A6D、overlay-26 sub_3D4 |
| `overlay-29 entry#4` | `010Bh` | overlay-02 sub_841、overlay-07 sub_591、overlay-15 sub_23FC、overlay-18 sub_B5F、overlay-24 sub_2A6D |
| `overlay-29 entry#5` | `055Bh` | overlay-02 sub_3691、overlay-10 sub_1C3E、overlay-14 sub_BAF、overlay-18 sub_B5F |
| `overlay-29 entry#6` | `06E1h` | overlay-07 sub_552、overlay-18 sub_10FF、overlay-24 sub_2A6D |
| `overlay-29 entry#7` | `05DDh` | overlay-07 sub_552、overlay-18 sub_10FF、overlay-24 sub_2A6D |
| `overlay-29 entry#8` | `0754h` | overlay-07 sub_591 |
| `overlay-29 entry#9` | `0813h` | overlay-02 sub_19EF、overlay-02 sub_841、overlay-02 sub_C15、overlay-16 sub_3748、overlay-18 sub_10FF |
| `overlay-29 entry#10` | `0881h` | overlay-02 sub_841、overlay-18 sub_10FF、overlay-28 sub_CC |
| `overlay-29 entry#11` | `0000h` | overlay-02 sub_0、overlay-07 sub_0、overlay-10 sub_0、overlay-11 sub_0、overlay-14 sub_0、overlay-17 sub_0、overlay-19 sub_0、overlay-24 sub_0、overlay-26 sub_0、overlay-28 sub_0 |
| `overlay-30 entry#2` | `016Bh` | overlay-28 sub_CC |
| `overlay-30 entry#4` | `06BDh` | overlay-02 sub_2F02、overlay-07 sub_1B3F、overlay-07 sub_4BD、overlay-10 sub_2F6、overlay-14 sub_83C、overlay-14 sub_90A |
| `overlay-30 entry#5` | `0587h` | overlay-10 sub_2F6、overlay-14 sub_302、overlay-14 sub_78E、overlay-14 sub_BAF |
| `overlay-30 entry#6` | `07C6h` | overlay-02 sub_2F02、overlay-02 sub_3691、overlay-07 sub_1B3F、overlay-10 sub_8C9、overlay-14 sub_83C、overlay-28 sub_CC |
| `overlay-30 entry#7` | `0841h` | overlay-14 sub_90A、overlay-28 sub_CC |
| `overlay-30 entry#8` | `104Ch` | overlay-02 sub_C15、overlay-07 sub_B6D、overlay-16 sub_3748 |
| `overlay-30 entry#9` | `133Ah` | overlay-02 sub_C15、overlay-16 sub_3748 |
| `overlay-30 entry#11` | `0000h` | overlay-02 sub_182E、overlay-02 sub_184B、overlay-02 sub_18B3、overlay-02 sub_19EF、overlay-09 sub_0、overlay-10 sub_0 |
| `overlay-31 entry#1` | `0007h` | overlay-22 sub_57E3、overlay-24 sub_1AF7 |
| `overlay-31 entry#2` | `019Dh` | overlay-22 sub_15FB、overlay-22 sub_3A03、overlay-24 sub_1AF7 |
| `overlay-31 entry#3` | `0246h` | overlay-22 sub_168B、overlay-22 sub_175B、overlay-22 sub_1930、overlay-22 sub_3A03、overlay-24 sub_1AF7 |
| `overlay-31 entry#4` | `054Ah` | overlay-13 sub_95A |
| `overlay-31 entry#5` | `03EAh` | overlay-09 sub_DB1、overlay-13 sub_33AC、overlay-13 sub_41CC |
| `overlay-31 entry#6` | `08E3h` | overlay-09 sub_2B1、overlay-12 sub_1809、overlay-12 sub_2B3E、overlay-13 sub_2220、overlay-13 sub_38AA、overlay-22 sub_370E、overlay-22 sub_48E4、overlay-23 sub_269、overlay-24 sub_2628、overlay-24 sub_274C |
| `overlay-31 entry#7` | `0000h` | overlay-02 sub_1956、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-13 sub_1D9D |
| `overlay-32 entry#2` | `0011h` | overlay-31 sub_8E3 |
| `overlay-32 entry#4` | `00F4h` | overlay-17 sub_3EE3、overlay-19 sub_B75、overlay-24 sub_2E05 |
| `overlay-32 entry#5` | `0134h` | overlay-08 sub_4D、overlay-17 sub_3EE3、overlay-19 sub_B75 |
| `overlay-32 entry#7` | `0320h` | overlay-08 sub_26B、overlay-13 sub_2220、overlay-13 sub_33AC、overlay-13 sub_394B、overlay-13 sub_3B37 |
| `overlay-32 entry#8` | `03EBh` | overlay-13 sub_74F、overlay-23 sub_1540 |
| `overlay-32 entry#9` | `0509h` | overlay-13 sub_1DF6、overlay-13 sub_33AC、overlay-22 sub_16BD、overlay-22 sub_1912、overlay-22 sub_2799、overlay-22 sub_317E、overlay-22 sub_3947、overlay-22 sub_3A03、overlay-22 sub_4F3E |
| `overlay-32 entry#10` | `0578h` | overlay-08 sub_B06、overlay-13 sub_74F、overlay-22 sub_48E4 |
| `overlay-32 entry#11` | `0732h` | overlay-13 sub_74F、overlay-24 sub_1AF7 |
| `overlay-32 entry#12` | `0763h` | overlay-08 sub_26B、overlay-09 sub_CB7、overlay-13 sub_19D8、overlay-22 sub_149E、overlay-24 sub_21DD |
| `overlay-32 entry#13` | `09DFh` | overlay-08 sub_997、overlay-08 sub_C6C、overlay-09 sub_DB1、overlay-13 sub_31F0、overlay-13 sub_33AC、overlay-13 sub_3B37、overlay-13 sub_48DC、overlay-13 sub_666、overlay-13 sub_74F、overlay-22 sub_2799、overlay-22 sub_370E、overlay-22 sub_48E4、overlay-22 sub_4F3E、overlay-24 sub_1AF7、overlay-24 sub_21DD、overlay-24 sub_2E05 |
| `overlay-32 entry#14` | `0B3Fh` | overlay-08 sub_CDA、overlay-09 sub_CB7、overlay-13 sub_19D8、overlay-22 sub_149E |
| `overlay-32 entry#15` | `0C36h` | overlay-08 sub_26B、overlay-08 sub_B06、overlay-09 sub_DB1、overlay-10 sub_1C3E、overlay-12 sub_1809、overlay-12 sub_1A65、overlay-12 sub_2657、overlay-12 sub_289C、overlay-12 sub_2B3E、overlay-13 sub_1DF6、overlay-13 sub_200F、overlay-13 sub_2220、overlay-13 sub_29A2、overlay-13 sub_2BFB、overlay-13 sub_31F0、overlay-13 sub_33AC、overlay-13 sub_38AA、overlay-13 sub_3B37、overlay-13 sub_4162、overlay-13 sub_41CC、overlay-13 sub_666、overlay-13 sub_95A、overlay-22 sub_1263、overlay-22 sub_1ABF、overlay-22 sub_2DB6、overlay-22 sub_3A03、overlay-22 sub_48E4、overlay-22 sub_4A8F、overlay-22 sub_54EB、overlay-22 sub_56C4、overlay-22 sub_57E3、overlay-22 sub_5974、overlay-22 sub_5AA8、overlay-22 sub_5C86、overlay-22 sub_5E11、overlay-22 sub_5F2F、overlay-22 sub_6022、overlay-23 sub_24ED、overlay-23 sub_269、overlay-23 sub_E45、overlay-24 sub_21DD、overlay-24 sub_2628、overlay-24 sub_274C |
| `overlay-32 entry#16` | `0C5Eh` | overlay-08 sub_26B、overlay-08 sub_B06、overlay-09 sub_DB1、overlay-10 sub_1C3E、overlay-12 sub_1809、overlay-12 sub_1A65、overlay-12 sub_2657、overlay-12 sub_289C、overlay-12 sub_2B3E、overlay-13 sub_1DF6、overlay-13 sub_200F、overlay-13 sub_2220、overlay-13 sub_29A2、overlay-13 sub_2BFB、overlay-13 sub_31F0、overlay-13 sub_33AC、overlay-13 sub_38AA、overlay-13 sub_3B37、overlay-13 sub_4162、overlay-13 sub_41CC、overlay-13 sub_666、overlay-13 sub_95A、overlay-22 sub_1263、overlay-22 sub_1ABF、overlay-22 sub_2DB6、overlay-22 sub_3A03、overlay-22 sub_48E4、overlay-22 sub_4A8F、overlay-22 sub_54EB、overlay-22 sub_56C4、overlay-22 sub_57E3、overlay-22 sub_5974、overlay-22 sub_5AA8、overlay-22 sub_5C86、overlay-22 sub_5E11、overlay-22 sub_5F53、overlay-22 sub_6022、overlay-23 sub_24ED、overlay-23 sub_269、overlay-23 sub_E45、overlay-24 sub_21DD、overlay-24 sub_2628、overlay-24 sub_274C |
| `overlay-32 entry#17` | `0C86h` | overlay-12 sub_1809、overlay-12 sub_2B3E、overlay-13 sub_38AA、overlay-23 sub_269、overlay-24 sub_2628、overlay-24 sub_274C |
| `overlay-32 entry#18` | `0CAEh` | overlay-08 sub_B06、overlay-13 sub_403F、overlay-13 sub_4811、overlay-13 sub_48DC、overlay-13 sub_74F、overlay-13 sub_95A、overlay-13 sub_F46、overlay-22 sub_48E4、overlay-23 sub_149D、overlay-23 sub_269、overlay-24 sub_21DD |
| `overlay-32 entry#19` | `0D06h` | overlay-08 sub_CDA、overlay-09 sub_735、overlay-10 sub_122C、overlay-23 sub_E45 |
| `overlay-32 entry#20` | `0E75h` | overlay-12 sub_299C、overlay-13 sub_1227、overlay-13 sub_31D、overlay-23 sub_16、overlay-23 sub_1FD6 |
| `overlay-32 entry#21` | `115Bh` | overlay-22 sub_2DB6、overlay-22 sub_48E4、overlay-23 sub_24ED |
| `overlay-32 entry#22` | `1355h` | overlay-08 sub_26B、overlay-08 sub_3D5、overlay-13 sub_1227、overlay-13 sub_275A、overlay-13 sub_3040、overlay-23 sub_149D、overlay-23 sub_24ED |
| `overlay-32 entry#23` | `0000h` | overlay-02 sub_1956、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-13 sub_1D9D、overlay-17 sub_0 |
| `overlay-33 entry#2` | `002Eh` | overlay-03 sub_11A、overlay-10 sub_102F |
| `overlay-33 entry#3` | `00F9h` | overlay-03 sub_11A、overlay-32 sub_174、overlay-32 sub_578、overlay-32 sub_81D |
| `overlay-33 entry#4` | `015Ah` | overlay-16 sub_A7E、overlay-17 sub_3A56、overlay-17 sub_3EE3 |
| `overlay-33 entry#5` | `01D4h` | overlay-02 sub_466、overlay-07 sub_1BE0、overlay-16 sub_3748、overlay-16 sub_A7E |
| `overlay-33 entry#6` | `0508h` | overlay-17 sub_3B99、overlay-32 sub_174、overlay-32 sub_320、overlay-32 sub_9DF、overlay-32 sub_B3F |
| `overlay-33 entry#7` | `0000h` | overlay-02 sub_0、overlay-05 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-10 sub_0、overlay-11 sub_0、overlay-13 sub_1D9D、overlay-14 sub_0、overlay-16 sub_0、overlay-17 sub_0、overlay-24 sub_0、overlay-28 sub_0、overlay-32 sub_0 |
| `overlay-34 entry#2` | `0000h` | overlay-02 sub_0、overlay-03 sub_0、overlay-04 sub_0、overlay-05 sub_0、overlay-06 sub_0、overlay-07 sub_0、overlay-08 sub_0、overlay-09 sub_0、overlay-10 sub_0、overlay-11 sub_0、overlay-12 sub_0、overlay-13 sub_1D9D、overlay-14 sub_0、overlay-15 sub_0、overlay-16 sub_0、overlay-19 sub_0、overlay-20 sub_0、overlay-21 sub_0、overlay-22 sub_0、overlay-23 sub_0、overlay-24 sub_0、overlay-26 sub_0、overlay-28 sub_0、overlay-29 sub_0、overlay-30 sub_0、overlay-32 sub_0、overlay-33 sub_0 |
| `overlay-34 entry#3` | `0240h` | overlay-02 sub_3621 |
| `overlay-35 entry#2` | `0047h` | overlay-30 sub_104C |
| `overlay-35 entry#3` | `0173h` | overlay-30 sub_11、overlay-30 sub_448 |
| `overlay-35 entry#4` | `0000h` | overlay-11 sub_0、overlay-30 sub_0 |
