# 效果 handler 的修正表

由 `scripts/effect_modifier_table.py` 產生，判讀見 [spec 1123](../spec/1123-effect-modifier-table.md)。不要手改。

`CHECKFX(timing)` 只說「這個時機有哪些效果要介入」（[timing 表](checkfx-timing-table.md)）；本表是**數字**：每個 handler 把修正寫進哪個全域、加多少。

| 狀態 | 意思 |
|---|---|
| `decoded` | 整支都是對全域的加減／設定，沒有其他指令 |
| `partial` | 有加減，但還有沒解析的指令（條件、呼叫、動角色記錄）|
| `inert` | 只有序幕與 `retf`，什麼都不做 |
| `unread` | 有內容但沒有可辨識的加減 |

統計：decoded 8、inert 12、partial 31、unread 90

## 暫存全域

| 位址 | 名稱 | 佐證 |
|---|---|---|
| `DS:6F92h` | `saving_throw` | 豁免總和（`overlay-23 entry#8` 寫它，spec 1119） |
| `DS:6F94h` | `damage` | 傷害值（抗寒／抗火在傷害時機把它折半） |
| `DS:6F95h` | `damage_element` | 傷害屬性旗標：bit 0 火、bit 1 冷 |
| `DS:6F96h` | `movement` | 移動率（`overlay-13 entry#2` 寫它，spec 1122） |
| `DS:6F9Bh` | `attack_forced_miss` | 攻擊必失手旗標 |
| `DS:6F9Ch` | `scratch_6f9c` | （未定） |
| `DS:6F9Fh` | `modifier` | 共用修正暫存；語意由讀它的 timing 決定 |
| `DS:6FA2h` | `morale` | 士氣（`overlay-09:01388h` 寫它，spec 1122） |

## 逐效果碼

| 碼 | overlay:位移 | 狀態 | 修正 | 出現在哪些 timing |
|---:|---|---|---|---|
| `01h` | `overlay-12:009Eh` | decoded | `morale` ＋ 5、`modifier` ＋ 1 | `0Ah`、`11h` |
| `02h` | `overlay-12:00B0h` | decoded | `morale` −（夾底） 5、`modifier` − 1 | `0Ah`、`11h` |
| `03h` | `overlay-12:00E8h` | unread | — | `07h` |
| `04h` | `overlay-12:016Bh` | unread | — | `10h` |
| `05h` | `overlay-12:2E33h` | inert | — | — |
| `06h` | `overlay-12:0188h` | partial | `damage_element` ＝ 9 | `04h`、`0Ah` |
| `07h` | `overlay-12:01F1h` | partial | `player+19Ah` ＝ 60、`player+19Bh` ＝ 60 | `0Bh` |
| `08h` | `overlay-12:0238h` | unread | — | `0Bh`、`0Ch` |
| `09h` | `overlay-12:026Fh` | unread | — | `0Bh`、`0Ch` |
| `0Ah` | `overlay-12:02A6h` | partial | `damage` ÷ 2、`saving_throw` ＋ 3 | `06h`、`0Ch` |
| `0Bh` | `overlay-12:02CBh` | partial | `morale` ＝ 100 | `0Fh`、`11h`、`13h` |
| `0Ch` | `overlay-12:038Ch` | inert | — | — |
| `0Dh` | `overlay-12:03A0h` | unread | — | `0Fh` |
| `0Eh` | `overlay-12:03DCh` | inert | — | — |
| `0Fh` | `overlay-12:03E5h` | unread | — | — |
| `10h` | `overlay-12:0443h` | inert | — | — |
| `11h` | `overlay-12:044Ch` | partial | `saving_throw` ＋ 1 | `06h`、`0Bh`、`0Ch` |
| `12h` | `overlay-12:0479h` | unread | — | `0Ah` |
| `13h` | `overlay-12:2E33h` | inert | — | — |
| `14h` | `overlay-12:04ADh` | partial | `damage` ÷ 2、`saving_throw` ＋ 3 | `06h`、`0Ch` |
| `15h` | `overlay-12:04E4h` | partial | `combat_state+02h` ＝ 0、`combat_state+01h` ＝ 0 | `0Fh` |
| `16h` | `overlay-12:054Ah` | unread | — | — |
| `17h` | `overlay-12:05B6h` | unread | — | `13h` |
| `18h` | `overlay-12:2E33h` | inert | — | — |
| `19h` | `overlay-12:06F2h` | unread | — | `01h`、`10h` |
| `1Ah` | `overlay-12:0727h` | unread | — | `0Ah` |
| `1Bh` | `overlay-12:0075h` | unread | — | `07h` |
| `1Ch` | `overlay-12:0769h` | unread | — | `05h`、`06h` |
| `1Dh` | `overlay-12:07ECh` | unread | — | `04h` |
| `1Eh` | `overlay-12:0818h` | partial | `combat_state+02h` ＝ 0、`combat_state+01h` ＝ 0、`player+19Bh` ＝ 50 | `0Bh`、`0Fh` |
| `1Fh` | `overlay-12:0075h` | unread | — | `07h` |
| `20h` | `overlay-12:08CDh` | partial | `player+04h` ＝ 0、`player+198h` ＝ 1、`player+E9h` ＝ 0、`player+E4h` ＝ 12、`player+11Ah` ＝ 0 | — |
| `21h` | `overlay-12:0982h` | decoded | `modifier` − 4、`player+19Ah` − 4、`player+19Bh` − 4、`saving_throw` − 4 | `0Ah`、`0Bh`、`0Ch` |
| `22h` | `overlay-12:09A7h` | unread | — | — |
| `23h` | `overlay-12:0A0Eh` | partial | `combat_state+10h` ＝ 1、`player+198h` ＝ 1 | `15h` |
| `24h` | `overlay-12:0BAFh` | decoded | `modifier` − 4、`saving_throw` − 4 | `0Ah`、`0Ch` |
| `25h` | `overlay-12:0BC2h` | unread | — | `01h`、`10h` |
| `26h` | `overlay-12:038Ch` | inert | — | — |
| `27h` | `overlay-12:0BE9h` | unread | — | `12h` |
| `28h` | `overlay-12:0C61h` | unread | — | — |
| `29h` | `overlay-12:100Fh` | unread | — | `05h` |
| `2Ah` | `overlay-12:104Ch` | decoded | `movement` ÷ 2 | `12h` |
| `2Bh` | `overlay-12:106Fh` | unread | — | — |
| `2Ch` | `overlay-12:10F9h` | unread | — | — |
| `2Dh` | `overlay-12:0238h` | unread | — | `0Bh`、`0Ch` |
| `2Eh` | `overlay-12:026Fh` | unread | — | `0Bh`、`0Ch` |
| `2Fh` | `overlay-12:118Ah` | unread | — | `10h` |
| `30h` | `overlay-12:11D7h` | unread | — | `10h` |
| `31h` | `overlay-12:1200h` | partial | `modifier` − 1、`saving_throw` − 1 | `0Ah`、`0Ch` |
| `32h` | `overlay-12:124Bh` | partial | `saving_throw` ＋ 2 | `0Ch`、`14h` |
| `33h` | `overlay-12:0075h` | unread | — | `07h` |
| `34h` | `overlay-12:0075h` | unread | — | `07h` |
| `35h` | `overlay-12:0075h` | unread | — | `07h` |
| `36h` | `overlay-12:127Eh` | partial | `saving_throw` ＋ 2 | `0Ch`、`14h` |
| `37h` | `overlay-12:12B1h` | inert | — | — |
| `38h` | `overlay-12:12BAh` | unread | — | `08h`、`13h` |
| `39h` | `overlay-13:403Fh` | unread | — | `02h` |
| `3Ah` | `overlay-12:12DBh` | partial | `combat_state+06h` ＝ 0 | `12h` |
| `3Bh` | `overlay-12:12FDh` | unread | — | — |
| `3Ch` | `overlay-12:131Dh` | unread | — | `05h` |
| `3Dh` | `overlay-12:1374h` | partial | `damage` − 2、`saving_throw` ＋ 4 | `06h`、`0Ch` |
| `3Eh` | `overlay-12:13CFh` | unread | — | — |
| `3Fh` | `overlay-12:1415h` | unread | — | `06h`、`09h` |
| `40h` | `overlay-12:1569h` | unread | — | `03h` |
| `41h` | `overlay-12:157Fh` | unread | — | `03h` |
| `42h` | `overlay-12:1595h` | unread | — | `03h` |
| `43h` | `overlay-12:15ABh` | unread | — | `03h` |
| `44h` | `overlay-12:15E6h` | partial | `player+13h` ＝ 7、`player+15h` ＝ 7 | — |
| `45h` | `overlay-12:1687h` | unread | — | `01h` |
| `46h` | `overlay-12:16C2h` | unread | — | `03h` |
| `47h` | `overlay-12:16D8h` | decoded | `attack_forced_miss` ＝ 1、`modifier` − 4 | `01h`、`10h` |
| `48h` | `overlay-12:16EBh` | unread | — | `08h`、`13h` |
| `49h` | `overlay-12:1729h` | unread | — | `06h` |
| `4Ah` | `overlay-12:1765h` | unread | — | `17h` |
| `4Bh` | `overlay-12:1771h` | unread | — | `04h`、`0Ah`、`0Dh` |
| `4Ch` | `overlay-12:17C8h` | unread | — | `04h`、`0Ah` |
| `4Dh` | `overlay-12:1809h` | partial | `player+198h` ＝ 1、`player+F7h` ＝ 178、`combat_state+01h` ＝ 0、`player+197h` ＝ 0 | `0Fh` |
| `4Eh` | `overlay-12:192Dh` | unread | — | — |
| `4Fh` | `overlay-12:196Dh` | partial | `damage_element` ＝ 9 | `02h`、`03h` |
| `50h` | `overlay-12:19A2h` | partial | `damage_element` ＝ 16 | `02h` |
| `51h` | `overlay-12:19D7h` | decoded | `damage` ÷ 2 | `05h` |
| `52h` | `overlay-12:19EEh` | unread | — | `05h`、`06h`、`08h` |
| `53h` | `overlay-12:1A65h` | unread | — | `0Eh` |
| `54h` | `overlay-12:1C1Ch` | partial | `player+1A4h` ＋ 8 | `06h` |
| `55h` | `overlay-12:1C4Fh` | unread | — | `05h` |
| `57h` | `overlay-13:42FDh` | unread | — | `03h`、`0Eh` |
| `59h` | `overlay-12:1C90h` | unread | — | `08h`、`10h` |
| `5Bh` | `overlay-12:1CF6h` | unread | — | — |
| `5Ch` | `overlay-12:2E33h` | inert | — | — |
| `5Dh` | `overlay-12:1FC4h` | partial | `damage` ÷ 2 | `06h` |
| `5Eh` | `overlay-12:1FE4h` | unread | — | `05h` |
| `5Fh` | `overlay-12:203Ch` | partial | `player+04h` ＝ 0 | — |
| `60h` | `overlay-13:4811h` | unread | — | `02h` |
| `61h` | `overlay-12:2078h` | unread | — | `0Ch` |
| `62h` | `overlay-12:20FAh` | partial | `player+1A4h` ＋ 3 | `13h` |
| `63h` | `overlay-12:212Ch` | unread | — | `08h`、`0Dh` |
| `64h` | `overlay-12:21D5h` | unread | — | `0Dh` |
| `65h` | `overlay-12:2212h` | unread | — | `05h`、`06h` |
| `66h` | `overlay-12:2264h` | unread | — | — |
| `67h` | `overlay-12:22A3h` | unread | — | `04h` |
| `68h` | `overlay-12:231Eh` | unread | — | `05h` |
| `69h` | `overlay-12:2392h` | unread | — | `06h`、`09h` |
| `6Ah` | `overlay-12:23A2h` | unread | — | `06h`、`09h` |
| `6Bh` | `overlay-12:23B2h` | unread | — | `09h` |
| `6Ch` | `overlay-12:23D8h` | unread | — | `09h` |
| `6Dh` | `overlay-12:23EFh` | unread | — | `09h` |
| `6Eh` | `overlay-12:23FFh` | unread | — | `06h`、`09h` |
| `6Fh` | `overlay-12:2418h` | unread | — | `09h`、`0Ch` |
| `70h` | `overlay-12:243Bh` | unread | — | `06h`、`09h` |
| `71h` | `overlay-12:2454h` | partial | `damage` − 1 | `06h` |
| `72h` | `overlay-12:2499h` | partial | `damage` ÷ 2 | `06h` |
| `73h` | `overlay-12:24B9h` | unread | — | `05h` |
| `74h` | `overlay-12:251Ch` | unread | — | `05h` |
| `75h` | `overlay-12:255Ah` | unread | — | `05h` |
| `76h` | `overlay-12:259Ah` | partial | `damage` ÷ 2 | `06h` |
| `77h` | `overlay-12:25BAh` | unread | — | `05h` |
| `78h` | `overlay-12:2606h` | unread | — | `05h` |
| `79h` | `overlay-12:2657h` | unread | — | `0Eh` |
| `7Ah` | `overlay-12:2782h` | unread | — | `02h` |
| `7Bh` | `overlay-12:27FAh` | partial | `damage_element` ＝ 10 | `02h` |
| `7Ch` | `overlay-12:282Fh` | unread | — | `09h` |
| `7Dh` | `overlay-12:2855h` | unread | — | `09h`、`0Ch` |
| `7Eh` | `overlay-12:289Ch` | unread | — | `0Eh` |
| `7Fh` | `overlay-12:2E33h` | inert | — | — |
| `81h` | `overlay-12:297Ah` | unread | — | `06h`、`09h` |
| `82h` | `overlay-12:299Ch` | unread | — | `05h` |
| `85h` | `overlay-12:2AA8h` | unread | — | `06h` |
| `86h` | `overlay-12:2AD6h` | unread | — | `04h` |
| `87h` | `overlay-12:2B0Fh` | unread | — | `06h` |
| `88h` | `overlay-12:2B28h` | decoded | `combat_state+06h` ＝ 0 | `07h` |
| `89h` | `overlay-12:2B3Eh` | partial | `player+198h` ＝ 1、`player+F7h` ＝ 178 | — |
| `8Ah` | `overlay-12:2C2Eh` | unread | — | `16h` |
| `8Bh` | `overlay-13:45CFh` | partial | `player+19Ch` ＝ 2、`player+19Dh` ＝ 0、`player+19Eh` ＝ 2、`player+1A0h` ＝ 8 | `0Eh` |
| `8Ch` | `overlay-12:2E33h` | inert | — | — |
| `8Dh` | `overlay-12:2C4Fh` | unread | — | — |
| `8Eh` | `overlay-12:2C94h` | unread | — | — |
| `8Fh` | `overlay-12:2CD9h` | unread | — | `05h` |
| `90h` | `overlay-13:4703h` | partial | `player+19Ch` ＝ 1、`player+19Dh` ＝ 0、`player+19Eh` ＝ 2、`player+1A0h` ＝ 8 | — |
| `91h` | `overlay-12:2D7Dh` | unread | — | `02h` |
| `92h` | `overlay-12:2E33h` | inert | — | — |
| `93h` | `overlay-12:2A2Ch` | partial | `7582` ＝ 0 | — |
