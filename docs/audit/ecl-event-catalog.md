# CoAB ECL 全事件靜態清冊摘要

> 本檔由 `cmd/ecl-event-catalog` 產生；請勿手動編輯。完整機器資料見
> [`ecl-event-catalog.json`](ecl-event-catalog.json)。

## 證據邊界

- 原始 archive：`curseoftheazurebonds.zip`
- SHA-256：`c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d`
- 位址空間：每個 block 解碼 payload offset；`code_address=0x8000+offset`。
- 推論等級：靜態 framing／直接 GOTO／GOSUB 可達性為 `exact`；effect kind 與直線序列是 audit 分類候選，不是原版 runtime 語意。
- 限制：TraceGraph follows statically visible GOTO/GOSUB and sequential fallthrough only.
- 限制：IF conditions, dynamic ON GOTO/ON GOSUB destinations, menus, CALL targets, and runtime state are not executed.
- 限制：ordered_effect_candidates are straight-line static candidates, not proof of original runtime order or side effects.

## 摘要

| 項目 | 數量 |
|---|---:|
| ECL DAX member | 6 |
| block | 25 |
| lifecycle entry | 125 |
| 不重複靜態可達 instruction | 1355 |
| 跨 effect-kind 直線候選 | 33 |

## Block 與 lifecycle entry

| Member | SHA-256 | Block | decoded bytes | lifecycle offsets | instructions | candidates |
|---|---|---:|---:|---|---:|---:|
| `ECL1.DAX` | `b365a7e2a4d6007137bfa7aa2d305569c86c14ea20d42c99e0668b381f276060` | `0x50` | 7516 | per_turn=0x006A<br>search_location=0x006B<br>pre_camp=0x1BF4<br>camp_interrupted=0x1C01<br>initial=0x0014 | 51 | 1 |
| `ECL1.DAX` | `b365a7e2a4d6007137bfa7aa2d305569c86c14ea20d42c99e0668b381f276060` | `0x51` | 7591 | per_turn=0x014A<br>search_location=0x014B<br>pre_camp=0x1C50<br>camp_interrupted=0x1C5D<br>initial=0x0014 | 36 | 1 |
| `ECL1.DAX` | `b365a7e2a4d6007137bfa7aa2d305569c86c14ea20d42c99e0668b381f276060` | `0x52` | 924 | per_turn=0x0397<br>search_location=0x0398<br>pre_camp=0x0395<br>camp_interrupted=0x0396<br>initial=0x0014 | 86 | 1 |
| `ECL2.DAX` | `ec2957d51c53d04a419f47453d345b25a5013f3cab483637acd8986739353338` | `0x01` | 7624 | per_turn=0x0137<br>search_location=0x0286<br>pre_camp=0x01EF<br>camp_interrupted=0x0225<br>initial=0x0014 | 44 | 0 |
| `ECL2.DAX` | `ec2957d51c53d04a419f47453d345b25a5013f3cab483637acd8986739353338` | `0x02` | 3550 | per_turn=0x0093<br>search_location=0x0133<br>pre_camp=0x00DB<br>camp_interrupted=0x0122<br>initial=0x0014 | 73 | 5 |
| `ECL2.DAX` | `ec2957d51c53d04a419f47453d345b25a5013f3cab483637acd8986739353338` | `0x03` | 7235 | per_turn=0x00F2<br>search_location=0x02CC<br>pre_camp=0x024C<br>camp_interrupted=0x02B0<br>initial=0x0014 | 40 | 0 |
| `ECL2.DAX` | `ec2957d51c53d04a419f47453d345b25a5013f3cab483637acd8986739353338` | `0x04` | 5972 | per_turn=0x0058<br>search_location=0x00E1<br>pre_camp=0x0072<br>camp_interrupted=0x00D0<br>initial=0x0014 | 41 | 0 |
| `ECL3.DAX` | `47435c33aba6b9cbf18d3de4ba9ad9fe77bfdb1a7fcb6432f32fd30cb00a4666` | `0x10` | 7665 | per_turn=0x0198<br>search_location=0x01BB<br>pre_camp=0x0275<br>camp_interrupted=0x02FA<br>initial=0x0014 | 84 | 4 |
| `ECL3.DAX` | `47435c33aba6b9cbf18d3de4ba9ad9fe77bfdb1a7fcb6432f32fd30cb00a4666` | `0x11` | 6432 | per_turn=0x02E1<br>search_location=0x0523<br>pre_camp=0x04F6<br>camp_interrupted=0x051E<br>initial=0x0014 | 44 | 2 |
| `ECL3.DAX` | `47435c33aba6b9cbf18d3de4ba9ad9fe77bfdb1a7fcb6432f32fd30cb00a4666` | `0x12` | 7360 | per_turn=0x00AB<br>search_location=0x00B7<br>pre_camp=0x0066<br>camp_interrupted=0x00A6<br>initial=0x0014 | 75 | 4 |
| `ECL3.DAX` | `47435c33aba6b9cbf18d3de4ba9ad9fe77bfdb1a7fcb6432f32fd30cb00a4666` | `0x15` | 3332 | per_turn=0x0113<br>search_location=0x02C8<br>pre_camp=0x00F5<br>camp_interrupted=0x010F<br>initial=0x0014 | 94 | 2 |
| `ECL4.DAX` | `60e60a5ee2564ef9d8ca49783a9a8c4ff055eb1c5cded923930c97386d5ea116` | `0x20` | 7409 | per_turn=0x01ED<br>search_location=0x029A<br>pre_camp=0x0265<br>camp_interrupted=0x028D<br>initial=0x0014 | 44 | 0 |
| `ECL4.DAX` | `60e60a5ee2564ef9d8ca49783a9a8c4ff055eb1c5cded923930c97386d5ea116` | `0x21` | 5456 | per_turn=0x0424<br>search_location=0x0479<br>pre_camp=0x04A7<br>camp_interrupted=0x04E3<br>initial=0x0014 | 41 | 1 |
| `ECL4.DAX` | `60e60a5ee2564ef9d8ca49783a9a8c4ff055eb1c5cded923930c97386d5ea116` | `0x22` | 5396 | per_turn=0x0465<br>search_location=0x050A<br>pre_camp=0x0482<br>camp_interrupted=0x04D4<br>initial=0x0014 | 25 | 0 |
| `ECL4.DAX` | `60e60a5ee2564ef9d8ca49783a9a8c4ff055eb1c5cded923930c97386d5ea116` | `0x23` | 2899 | per_turn=0x0030<br>search_location=0x0042<br>pre_camp=0x0041<br>camp_interrupted=0x0040<br>initial=0x0014 | 27 | 1 |
| `ECL4.DAX` | `60e60a5ee2564ef9d8ca49783a9a8c4ff055eb1c5cded923930c97386d5ea116` | `0x25` | 5600 | per_turn=0x015C<br>search_location=0x023F<br>pre_camp=0x022E<br>camp_interrupted=0x023B<br>initial=0x0014 | 47 | 3 |
| `ECL5.DAX` | `5b0d3420b0635e62aae9ec14b4e6637f592883a6b45ccae059eec2ce4c05f086` | `0x30` | 585 | per_turn=0x0242<br>search_location=0x0245<br>pre_camp=0x0243<br>camp_interrupted=0x0244<br>initial=0x0014 | 31 | 1 |
| `ECL5.DAX` | `5b0d3420b0635e62aae9ec14b4e6637f592883a6b45ccae059eec2ce4c05f086` | `0x31` | 3774 | per_turn=0x0397<br>search_location=0x04A5<br>pre_camp=0x00CD<br>camp_interrupted=0x00FE<br>initial=0x0014 | 46 | 0 |
| `ECL5.DAX` | `5b0d3420b0635e62aae9ec14b4e6637f592883a6b45ccae059eec2ce4c05f086` | `0x32` | 6048 | per_turn=0x011D<br>search_location=0x05B5<br>pre_camp=0x057F<br>camp_interrupted=0x05A8<br>initial=0x0014 | 51 | 1 |
| `ECL5.DAX` | `5b0d3420b0635e62aae9ec14b4e6637f592883a6b45ccae059eec2ce4c05f086` | `0x33` | 7392 | per_turn=0x07A3<br>search_location=0x0EB4<br>pre_camp=0x0E84<br>camp_interrupted=0x0EB0<br>initial=0x0014 | 52 | 1 |
| `ECL5.DAX` | `5b0d3420b0635e62aae9ec14b4e6637f592883a6b45ccae059eec2ce4c05f086` | `0x35` | 6800 | per_turn=0x020D<br>search_location=0x06B2<br>pre_camp=0x01EF<br>camp_interrupted=0x0209<br>initial=0x0014 | 83 | 1 |
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` | `0x40` | 7148 | per_turn=0x00AA<br>search_location=0x0231<br>pre_camp=0x0208<br>camp_interrupted=0x022D<br>initial=0x0014 | 46 | 0 |
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` | `0x42` | 6300 | per_turn=0x0081<br>search_location=0x01F2<br>pre_camp=0x01B8<br>camp_interrupted=0x01EE<br>initial=0x0014 | 55 | 1 |
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` | `0x43` | 5363 | per_turn=0x0104<br>search_location=0x01AD<br>pre_camp=0x017F<br>camp_interrupted=0x018C<br>initial=0x0014 | 43 | 1 |
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` | `0x45` | 2934 | per_turn=0x0102<br>search_location=0x02B1<br>pre_camp=0x00E4<br>camp_interrupted=0x00FE<br>initial=0x0014 | 96 | 2 |

## 跨類型副作用序列候選

以下每列只證明同一靜態直線區段可看見不同 effect kind。IF、動態選單、
CALL consumer、pause commit 與實際分支仍須逐列回到原始 bytes、IDA 與 runtime 閉合。

| 候選 ID | lifecycle | 靜態 effect sequence | 審查狀態 | 證據 |
|---|---|---|---|---|
| `ECL1.DAX/0x50/0x0014-0x006A` | initial | `0x001A LOAD FILES [resource_load]` → `0x0021 PICTURE [presentation]` | 未審查 | `static_straight_line_candidate` |
| `ECL1.DAX/0x51/0x0014-0x0050` | initial | `0x001A LOAD FILES [resource_load]` → `0x0021 PICTURE [presentation]` | 未審查 | `static_straight_line_candidate` |
| `ECL1.DAX/0x52/0x0014-0x0395` | initial | `0x0014 LOAD FILES [resource_load]` → `0x001B CLEAR BOX [presentation]` → `0x001C PICTURE [presentation]` → `0x001F ADD NPC [party_mutation]` → `0x0024 ADD NPC [party_mutation]` → `0x0029 ADD NPC [party_mutation]` → `0x002E PRINTCLEAR [text]` → `0x0053 PRINT [text]` → `0x007C PRINT [text]` → `0x00A7 PRINT [text]` → `0x00B0 DELAY [presentation]` → `0x00B1 DELAY [presentation]` → `0x00B2 DELAY [presentation]` → `0x00B3 DELAY [presentation]` → `0x00B4 PRINTCLEAR [text]` → `0x00DB PRINT [text]` → `0x0105 DELAY [presentation]` → `0x0106 DELAY [presentation]` → `0x0107 DELAY [presentation]` → `0x0108 DELAY [presentation]` → `0x010F PICTURE [presentation]` → `0x0112 PRINTCLEAR [text]` → `0x013A PRINT [text]` → `0x0166 PRINT [text]` → `0x0187 DELAY [presentation]` → `0x0188 DELAY [presentation]` → `0x0189 DELAY [presentation]` → `0x018A DELAY [presentation]` → `0x0191 CLEAR BOX [presentation]` → `0x0192 PICTURE [presentation]` → `0x0195 PRINTCLEAR [text]` → `0x01BC PRINT [text]` → `0x01E2 CALL [external_call]` → `0x01E6 CALL [external_call]` → `0x01EA CALL [external_call]` → `0x01EE CALL [external_call]` → `0x01F2 CALL [external_call]` → `0x01F6 CALL [external_call]` → `0x01FA CALL [external_call]` → `0x01FE CALL [external_call]` → `0x0202 CALL [external_call]` → `0x0206 CALL [external_call]` → `0x020A CALL [external_call]` → `0x0214 CLEARMONSTERS [combat_setup]` → `0x0227 SETUP MONSTER [combat_setup]` → `0x022E LOAD MONSTER [combat_setup]` → `0x0235 COMBAT [combat_boundary]` → `0x0236 DELAY [presentation]` → `0x0237 DELAY [presentation]` → `0x023E CLEAR BOX [presentation]` → `0x023F PICTURE [presentation]` → `0x0242 PRINTCLEAR [text]` → `0x026B PRINT [text]` → `0x0297 PRINT [text]` → `0x02A2 DELAY [presentation]` → `0x02A3 DELAY [presentation]` → `0x02A4 DELAY [presentation]` → `0x02A5 DELAY [presentation]` → `0x02A6 PICTURE [presentation]` → `0x02A9 PRINTCLEAR [text]` → `0x02D0 PRINT [text]` → `0x02FD PRINT [text]` → `0x0318 DELAY [presentation]` → `0x0319 DELAY [presentation]` → `0x031A DELAY [presentation]` → `0x031B DELAY [presentation]` → `0x031C PRINTCLEAR [text]` → `0x0343 PRINT [text]` → `0x036E PRINT [text]` → `0x038D DELAY [presentation]` → `0x038E DELAY [presentation]` → `0x038F DELAY [presentation]` → `0x0390 DELAY [presentation]` → `0x0391 PROGRAM [program_boundary]` | 未審查 | `static_straight_line_candidate` |
| `ECL2.DAX/0x02/0x021C-0x02CB` | per_turn | `0x022A PICTURE [presentation]` → `0x023A PRINTCLEAR [text]` → `0x0261 PRINT [text]` → `0x0276 PRINT [text]` → `0x029F PRINT [text]` | 未審查 | `static_straight_line_candidate` |
| `ECL2.DAX/0x02/0x02CB-0x0325` | per_turn | `0x02CB PRINT [text]` → `0x02F9 PRINT [text]` → `0x0307 PICTURE [presentation]` → `0x030A CALL [external_call]` → `0x0314 SETUP MONSTER [combat_setup]` | 未審查 | `static_straight_line_candidate` |
| `ECL2.DAX/0x02/0x03FE-0x0486` | per_turn | `0x03FE PRINTCLEAR [text]` → `0x042B PRINT [text]` → `0x045C CLEARMONSTERS [combat_setup]` → `0x045D LOAD MONSTER [combat_setup]` → `0x0464 LOAD MONSTER [combat_setup]` → `0x0477 LOAD CHARACTER [party_load]` | 未審查 | `static_straight_line_candidate` |
| `ECL2.DAX/0x02/0x04BC-0x053A` | per_turn | `0x04BC COMBAT [combat_boundary]` → `0x04BD PRINTCLEAR [text]` → `0x04E8 PRINT [text]` → `0x0500 PRINT [text]` → `0x052D PRINT [text]` | `covered/exact`；[1045-dungeon-main-loop-dos.md](../spec/1045-dungeon-main-loop-dos.md)、[1083-ecl-opcode-mnemonics.md](../spec/1083-ecl-opcode-mnemonics.md)、[1095-ecl-main-loop-and-combat-continuation.md](../spec/1095-ecl-main-loop-and-combat-continuation.md) | `static_straight_line_candidate` |
| `ECL2.DAX/0x02/0x0CCB-0x0CD4` | per_turn | `0x0CCB SPRITE OFF [presentation]` → `0x0CCC CALL [external_call]` → `0x0CD0 PRINTCLEAR [text]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x10/0x018C-0x0198` | camp_interrupted, per_turn | `0x018C PICTURE [presentation]` → `0x018F PRINTCLEAR [text]` → `0x0193 CALL [external_call]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x10/0x08D3-0x090F` | camp_interrupted | `0x08E2 PRINTCLEAR [text]` → `0x08FA CLEARMONSTERS [combat_setup]` → `0x08FB LOAD MONSTER [combat_setup]` → `0x0903 COMBAT [combat_boundary]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x10/0x0910-0x0967` | camp_interrupted | `0x0910 PRINTCLEAR [text]` → `0x092F PRINT RETURN [interaction_boundary]` → `0x0930 PRINT [text]` → `0x0954 PRINT [text]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x10/0x1560-0x15BC` | camp_interrupted | `0x157B CALL [external_call]` → `0x157F PRINTCLEAR [text]` → `0x159F PRINT [text]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x11/0x0597-0x05A3` | camp_interrupted | `0x0597 PICTURE [presentation]` → `0x059A PRINTCLEAR [text]` → `0x059E CALL [external_call]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x11/0x114E-0x1193` | camp_interrupted | `0x1154 SETUP MONSTER [combat_setup]` → `0x1161 PRINTCLEAR [text]` → `0x1185 CLEARMONSTERS [combat_setup]` → `0x1186 LOAD MONSTER [combat_setup]` → `0x118E COMBAT [combat_boundary]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x12/0x026E-0x027A` | camp_interrupted | `0x026E PICTURE [presentation]` → `0x0271 PRINTCLEAR [text]` → `0x0275 CALL [external_call]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x12/0x06BF-0x0704` | camp_interrupted | `0x06C5 SETUP MONSTER [combat_setup]` → `0x06D2 PRINTCLEAR [text]` → `0x06F6 CLEARMONSTERS [combat_setup]` → `0x06F7 LOAD MONSTER [combat_setup]` → `0x06FF COMBAT [combat_boundary]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x12/0x128E-0x12AF` | per_turn | `0x1294 PICTURE [presentation]` → `0x129D PRINTCLEAR [text]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x12/0x1AFB-0x1B0C` | per_turn | `0x1AFB PRINT RETURN [interaction_boundary]` → `0x1AFC PRINT [text]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x15/0x04CC-0x050A` | camp_interrupted | `0x04E6 SETUP MONSTER [combat_setup]` → `0x04F5 PRINTCLEAR [text]` | 未審查 | `static_straight_line_candidate` |
| `ECL3.DAX/0x15/0x050A-0x0578` | camp_interrupted | `0x0536 CLEARMONSTERS [combat_setup]` → `0x0537 LOAD MONSTER [combat_setup]` → `0x0548 LOAD MONSTER [combat_setup]` → `0x055A TREASURE [treasure_boundary]` → `0x056D COMBAT [combat_boundary]` | `covered/exact`；[255-ecl-treasure-signal.md](../spec/255-ecl-treasure-signal.md)、[257-treasure-random-and-pickup.md](../spec/257-treasure-random-and-pickup.md)、[258-treasure-combat-continuation.md](../spec/258-treasure-combat-continuation.md)、[558-pc98-ecl-treasure-combat-boundary.md](../spec/558-pc98-ecl-treasure-combat-boundary.md) | `static_straight_line_candidate` |
| `ECL4.DAX/0x21/0x05B5-0x0637` | camp_interrupted | `0x05BB SETUP MONSTER [combat_setup]` → `0x05C8 PRINTCLEAR [text]` → `0x0616 CLEARMONSTERS [combat_setup]` → `0x0617 LOAD MONSTER [combat_setup]` → `0x061F LOAD MONSTER [combat_setup]` → `0x0627 LOAD MONSTER [combat_setup]` → `0x062F COMBAT [combat_boundary]` | 未審查 | `static_straight_line_candidate` |
| `ECL4.DAX/0x23/0x0130-0x019D` | initial, per_turn | `0x0130 PRINTCLEAR [text]` → `0x015A PRINT RETURN [interaction_boundary]` → `0x015B PRINT [text]` → `0x0189 PRINT RETURN [interaction_boundary]` → `0x018A PRINT [text]` → `0x0199 HORIZONTAL MENU [interaction_boundary]` | 未審查 | `static_straight_line_candidate` |
| `ECL4.DAX/0x25/0x021F-0x023B` | per_turn | `0x021F DESTROY ITEMS [inventory]` → `0x0222 DESTROY ITEMS [inventory]` → `0x022B NEWECL [ecl_transition]` | 未審查 | `static_straight_line_candidate` |
| `ECL4.DAX/0x25/0x1271-0x12A7` | camp_interrupted | `0x1271 PRINT [text]` → `0x1288 CLEARMONSTERS [combat_setup]` → `0x1289 LOAD MONSTER [combat_setup]` → `0x1291 TREASURE [treasure_boundary]` → `0x12A2 COMBAT [combat_boundary]` | `covered/exact`；[255-ecl-treasure-signal.md](../spec/255-ecl-treasure-signal.md)、[257-treasure-random-and-pickup.md](../spec/257-treasure-random-and-pickup.md)、[258-treasure-combat-continuation.md](../spec/258-treasure-combat-continuation.md)、[558-pc98-ecl-treasure-combat-boundary.md](../spec/558-pc98-ecl-treasure-combat-boundary.md) | `static_straight_line_candidate` |
| `ECL4.DAX/0x25/0x1529-0x1534` | camp_interrupted | `0x1529 PICTURE [presentation]` → `0x152C PRINTCLEAR [text]` → `0x152F CALL [external_call]` | 未審查 | `static_straight_line_candidate` |
| `ECL5.DAX/0x30/0x0086-0x00B0` | initial | `0x0098 NEWECL [ecl_transition]` → `0x00A1 LOAD CHARACTER [party_load]` | 未審查 | `static_straight_line_candidate` |
| `ECL5.DAX/0x32/0x0774-0x07BA` | camp_interrupted | `0x077A SETUP MONSTER [combat_setup]` → `0x0787 PRINTCLEAR [text]` → `0x07A2 CLEARMONSTERS [combat_setup]` → `0x07A3 LOAD MONSTER [combat_setup]` → `0x07AA LOAD MONSTER [combat_setup]` → `0x07B1 LOAD MONSTER [combat_setup]` → `0x07B8 COMBAT [combat_boundary]` | 未審查 | `static_straight_line_candidate` |
| `ECL5.DAX/0x33/0x0045-0x00FD` | initial | `0x0051 LOAD FILES [resource_load]` → `0x0058 LOAD PIECES [resource_load]` → `0x0065 PICTURE [presentation]` → `0x0068 PRINTCLEAR [text]` → `0x008F PRINT [text]` → `0x00BB PRINT [text]` → `0x00E7 PRINT [text]` | 未審查 | `static_straight_line_candidate` |
| `ECL5.DAX/0x35/0x0982-0x0A06` | search_location | `0x0988 PICTURE [presentation]` → `0x0991 PRINTCLEAR [text]` → `0x09B8 PRINT [text]` → `0x09E3 PRINT [text]` | 未審查 | `static_straight_line_candidate` |
| `ECL6.DAX/0x42/0x003F-0x0081` | initial | `0x003F LOAD FILES [resource_load]` → `0x0046 LOAD PIECES [resource_load]` → `0x004D CALL [external_call]` → `0x0051 PRINTCLEAR [text]` → `0x0078 PRINT [text]` | 未審查 | `static_straight_line_candidate` |
| `ECL6.DAX/0x43/0x004D-0x008B` | initial | `0x0065 LOAD FILES [resource_load]` → `0x006C LOAD PIECES [resource_load]` → `0x007F CALL [external_call]` | 未審查 | `static_straight_line_candidate` |
| `ECL6.DAX/0x45/0x04C5-0x04F6` | camp_interrupted | `0x04D2 SETUP MONSTER [combat_setup]` → `0x04E1 PRINTCLEAR [text]` | 未審查 | `static_straight_line_candidate` |
| `ECL6.DAX/0x45/0x04F6-0x0575` | camp_interrupted | `0x0522 CLEARMONSTERS [combat_setup]` → `0x052A LOAD MONSTER [combat_setup]` → `0x0534 LOAD MONSTER [combat_setup]` → `0x0545 LOAD MONSTER [combat_setup]` → `0x0557 TREASURE [treasure_boundary]` → `0x056A COMBAT [combat_boundary]` | `covered/exact`；[255-ecl-treasure-signal.md](../spec/255-ecl-treasure-signal.md)、[257-treasure-random-and-pickup.md](../spec/257-treasure-random-and-pickup.md)、[258-treasure-combat-continuation.md](../spec/258-treasure-combat-continuation.md)、[558-pc98-ecl-treasure-combat-boundary.md](../spec/558-pc98-ecl-treasure-combat-boundary.md) | `static_straight_line_candidate` |

## 使用方式

1. 先以 JSON 的 member／block／offset 回讀原始 bytes 與完整 instruction operands。
2. 對候選建立 `re-closure-record-template.md`，補 writer→projection→consumer。
3. 以原版 runtime 或等價第一級證據確認實際分支與 commit phase。
4. 完成 ordered event contract 後，才將候選升格成 READY 規格與 engine transaction。
