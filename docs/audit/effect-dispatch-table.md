# 效果分派表（DOS `DS:6FA6h` ／ PC-98 `DS:A040h`）

本檔由 `scripts/effect_dispatch_table.py` 產生，判讀見 `docs/spec/1005-effect-dispatch-table.md`。

分派端 `overlay-23:00C9h`（`CALLEFFECT`，spec 576）以 **效果碼 × 4** 索引，效果碼從 **1** 起算。

| 效果碼 | DOS 處理常式 | PC-98 處理常式 | 台帳 | 判讀 |
|---|---|---|---|---|
| 1 | `overlay-12:0009Eh`（entry#7） | `overlay-12:0009Eh`（entry#7） | 已解讀 | 573 |
| 2 | `overlay-12:000B0h`（entry#8） | `overlay-12:000B0h`（entry#8） | 已解讀 | 573 |
| 3 | `overlay-12:000E8h`（entry#9） | `overlay-12:000E3h`（entry#9） | 已解讀 | 676 |
| 4 | `overlay-12:0016Bh`（entry#10） | `overlay-12:00166h`（entry#10） | 已解讀 | 573 |
| 5 | `overlay-12:02E33h`（entry#135） | `overlay-12:02ECBh`（entry#135） | 已解讀 | 569 |
| 6 | `overlay-12:00188h`（entry#11） | `overlay-12:00183h`（entry#11） | 已解讀 | 676 |
| 7 | `overlay-12:001F1h`（entry#12） | `overlay-12:001E6h`（entry#12） | 已解讀 | 638 |
| 8 | `overlay-12:00238h`（entry#13） | `overlay-12:0022Dh`（entry#13） | 已解讀 | 633 |
| 9 | `overlay-12:0026Fh`（entry#14） | `overlay-12:00264h`（entry#14） | 已解讀 | 633 |
| 10 | `overlay-12:002A6h`（entry#15） | `overlay-12:0029Bh`（entry#15） | 已解讀 | 573 |
| 11 | `overlay-12:002CBh`（entry#16） | `overlay-12:002C0h`（entry#16） | 已解讀 | 745 |
| 12 | `overlay-12:0038Ch`（entry#17） | `overlay-12:00381h`（entry#17） | 已解讀 | 569 |
| 13 | `overlay-12:003A0h`（entry#18） | `overlay-12:00397h`（entry#18） | 已解讀 | 633 |
| 14 | `overlay-12:003DCh`（entry#19） | `overlay-12:003D3h`（entry#19） | 已解讀 | 569 |
| 15 | `overlay-12:003E5h`（entry#20） | `overlay-12:003DCh`（entry#20） | 已解讀 | 676 |
| 16 | `overlay-12:00443h`（entry#21） | `overlay-12:0043Ah`（entry#21） | 已解讀 | 569 |
| 17 | `overlay-12:0044Ch`（entry#22） | `overlay-12:00443h`（entry#22） | 已解讀 | 573 |
| 18 | `overlay-12:00479h`（entry#23） | `overlay-12:00470h`（entry#23） | 已解讀 | 633 |
| 19 | `overlay-12:02E33h`（entry#135） | `overlay-12:02ECBh`（entry#135） | 已解讀 | 569 |
| 20 | `overlay-12:004ADh`（entry#24） | `overlay-12:004A4h`（entry#24） | 已解讀 | 573 |
| 21 | `overlay-12:004E4h`（entry#25） | `overlay-12:004E2h`（entry#25） | 已解讀 | 638 |
| 22 | `overlay-12:0054Ah`（entry#26） | `overlay-12:0054Ch`（entry#26） | 已解讀 | 677 |
| 23 | `overlay-12:005B6h`（entry#27） | `overlay-12:005BDh`（entry#27） | 已解讀 | 747 |
| 24 | `overlay-12:02E33h`（entry#135） | `overlay-12:02ECBh`（entry#135） | 已解讀 | 569 |
| 25 | `overlay-12:006F2h`（entry#28） | `overlay-12:006F9h`（entry#28） | 已解讀 | 633 |
| 26 | `overlay-12:00727h`（entry#29） | `overlay-12:0072Eh`（entry#29） | 已解讀 | 633 |
| 27 | `overlay-12:00075h`（entry#5） | `overlay-12:00075h`（entry#5） | 已解讀 | 573 |
| 28 | `overlay-12:00769h`（entry#31） | `overlay-12:00773h`（entry#31） | 已解讀 | 677 |
| 29 | `overlay-12:007ECh`（entry#32） | `overlay-12:007F6h`（entry#32） | 已解讀 | 573 |
| 30 | `overlay-12:00818h`（entry#33） | `overlay-12:00829h`（entry#33） | 已解讀 | 744 |
| 31 | `overlay-12:00075h`（entry#5） | `overlay-12:00075h`（entry#5） | 已解讀 | 573 |
| 32 | `overlay-12:008CDh`（entry#34） | `overlay-12:008DFh`（entry#34） | 已解讀 | 744 |
| 33 | `overlay-12:00982h`（entry#35） | `overlay-12:00994h`（entry#35） | 已解讀 | 573 |
| 34 | `overlay-12:009A7h`（entry#36） | `overlay-12:009B9h`（entry#36） | 已解讀 | 633 |
| 35 | `overlay-12:00A0Eh`（entry#37） | `overlay-12:00A32h`（entry#37） | 已解讀 | 831 |
| 36 | `overlay-12:00BAFh`（entry#38） | `overlay-12:00BC8h`（entry#38） | 已解讀 | 573 |
| 37 | `overlay-12:00BC2h`（entry#39） | `overlay-12:00BDBh`（entry#39） | 已解讀 | 573 |
| 38 | `overlay-12:0038Ch`（entry#17） | `overlay-12:00381h`（entry#17） | 已解讀 | 569 |
| 39 | `overlay-12:00BE9h`（entry#40） | `overlay-12:00C0Ch`（entry#40） | 已解讀 | 678 |
| 40 | `overlay-12:00C61h`（entry#41） | `overlay-12:00C82h`（entry#41） | 已解讀 | 748 |
| 41 | `overlay-12:0100Fh`（entry#44） | `overlay-12:01030h`（entry#44） | 已解讀 | 1005 |
| 42 | `overlay-12:0104Ch`（entry#45） | `overlay-12:0106Dh`（entry#45） | 已解讀 | 573 |
| 43 | `overlay-12:0106Fh`（entry#46） | `overlay-12:01091h`（entry#46） | 已解讀 | 679 |
| 44 | `overlay-12:010F9h`（entry#47） | `overlay-12:0111Bh`（entry#47） | 已解讀 | 679 |
| 45 | `overlay-12:00238h`（entry#13） | `overlay-12:0022Dh`（entry#13） | 已解讀 | 633 |
| 46 | `overlay-12:0026Fh`（entry#14） | `overlay-12:00264h`（entry#14） | 已解讀 | 633 |
| 47 | `overlay-12:0118Ah`（entry#48） | `overlay-12:011ACh`（entry#48） | 已解讀 | 638 |
| 48 | `overlay-12:011D7h`（entry#49） | `overlay-12:011F9h`（entry#49） | 已解讀 | 573 |
| 49 | `overlay-12:01200h`（entry#50） | `overlay-12:01222h`（entry#50） | 已解讀 | 638 |
| 50 | `overlay-12:0124Bh`（entry#51） | `overlay-12:0126Dh`（entry#51） | 已解讀 | 638 |
| 51 | `overlay-12:00075h`（entry#5） | `overlay-12:00075h`（entry#5） | 已解讀 | 573 |
| 52 | `overlay-12:00075h`（entry#5） | `overlay-12:00075h`（entry#5） | 已解讀 | 573 |
| 53 | `overlay-12:00075h`（entry#5） | `overlay-12:00075h`（entry#5） | 已解讀 | 573 |
| 54 | `overlay-12:0127Eh`（entry#52） | `overlay-12:012A0h`（entry#52） | 已解讀 | 638 |
| 55 | `overlay-12:012B1h`（entry#53） | `overlay-12:012D3h`（entry#53） | 已解讀 | 569 |
| 56 | `overlay-12:012BAh`（entry#54） | `overlay-12:012DCh`（entry#54） | 已解讀 | 569 |
| 57 | `overlay-13:0403Fh`（entry#28） | `overlay-13:03F71h`（entry#28） | 已解讀 | 826 |
| 58 | `overlay-12:012DBh`（entry#55） | `overlay-12:012FDh`（entry#55） | 已解讀 | 573 |
| 59 | `overlay-12:012FDh`（entry#56） | `overlay-12:0131Fh`（entry#56） | 已解讀 | 573 |
| 60 | `overlay-12:0131Dh`（entry#57） | `overlay-12:0133Fh`（entry#57） | 已解讀 | 639 |
| 61 | `overlay-12:01374h`（entry#58） | `overlay-12:01396h`（entry#58） | 已解讀 | 739 |
| 62 | `overlay-12:013CFh`（entry#59） | `overlay-12:013F1h`（entry#59） | 已解讀 | 639 |
| 63 | `overlay-12:01415h`（entry#60） | `overlay-12:01437h`（entry#60） | 已解讀 | 1005 |
| 64 | `overlay-12:01569h`（entry#63） | `overlay-12:0158Fh`（entry#63） | 已解讀 | 1005 |
| 65 | `overlay-12:0157Fh`（entry#64） | `overlay-12:015A5h`（entry#64） | 已解讀 | 1005 |
| 66 | `overlay-12:01595h`（entry#65） | `overlay-12:015BBh`（entry#65） | 已解讀 | 1005 |
| 67 | `overlay-12:015ABh`（entry#66） | `overlay-12:015D1h`（entry#66） | 已解讀 | 1005 |
| 68 | `overlay-12:015E6h`（entry#67） | `overlay-12:01621h`（entry#67） | 已解讀 | 740 |
| 69 | `overlay-12:01687h`（entry#68） | `overlay-12:016C2h`（entry#68） | 已解讀 | 639 |
| 70 | `overlay-12:016C2h`（entry#69） | `overlay-12:016FDh`（entry#69） | 已解讀 | 569 |
| 71 | `overlay-12:016D8h`（entry#70） | `overlay-12:01713h`（entry#70） | 已解讀 | 573 |
| 72 | `overlay-12:016EBh`（entry#71） | `overlay-12:01726h`（entry#71） | 已解讀 | 573 |
| 73 | `overlay-12:01729h`（entry#72） | `overlay-12:0176Dh`（entry#72） | 已解讀 | 639 |
| 74 | `overlay-12:01765h`（entry#73） | `overlay-12:017A9h`（entry#73） | 已解讀 | 569 |
| 75 | `overlay-12:01771h`（entry#74） | `overlay-12:017B5h`（entry#74） | 已解讀 | 639 |
| 76 | `overlay-12:017C8h`（entry#75） | `overlay-12:0180Ch`（entry#75） | 已解讀 | 639 |
| 77 | `overlay-12:01809h`（entry#76） | `overlay-12:0184Fh`（entry#76） | 已解讀 | 743 |
| 78 | `overlay-12:0192Dh`（entry#77） | `overlay-12:01973h`（entry#77） | 已解讀 | 1005 |
| 79 | `overlay-12:0196Dh`（entry#78） | `overlay-12:019B3h`（entry#78） | 已解讀 | 640 |
| 80 | `overlay-12:019A2h`（entry#79） | `overlay-12:019E8h`（entry#79） | 已解讀 | 640 |
| 81 | `overlay-12:019D7h`（entry#80） | `overlay-12:01A1Dh`（entry#80） | 已解讀 | 573 |
| 82 | `overlay-12:019EEh`（entry#81） | `overlay-12:01A34h`（entry#81） | 已解讀 | 640 |
| 83 | `overlay-12:01A65h`（entry#82） | `overlay-12:01ABCh`（entry#82） | 已解讀 | 832 |
| 84 | `overlay-12:01C1Ch`（entry#83） | `overlay-12:01C86h`（entry#83） | 已解讀 | 641 |
| 85 | `overlay-12:01C4Fh`（entry#84） | `overlay-12:01CB9h`（entry#84） | 已解讀 | 641 |
| 86 | `overlay-22:05974h`（entry#110）（INITSPELLS，spec 1202） | `overlay-22:05D17h`（entry#111）（INITSPELLS，spec 1202） | 已解讀 | 723 |
| 87 | `overlay-13:042FDh`（entry#29） | `overlay-13:0426Bh`（entry#29） | 已解讀 | 842 |
| 88 | `overlay-22:057E3h`（entry#109）（INITSPELLS，spec 1202） | `overlay-22:05B90h`（entry#110）（INITSPELLS，spec 1202） | 已解讀 | 847 |
| 89 | `overlay-12:01C90h`（entry#85） | `overlay-12:01CFAh`（entry#85） | 已解讀 | 641 |
| 90 | `overlay-22:05AA8h`（entry#111）（INITSPELLS，spec 1202） | `overlay-22:05E32h`（entry#112）（INITSPELLS，spec 1202） | 已解讀 | 725 |
| 91 | `overlay-12:01CF6h`（entry#86） | `overlay-12:01D5Eh`（entry#86） | 已解讀 | 748 |
| 92 | `overlay-12:02E33h`（entry#135） | `overlay-12:02ECBh`（entry#135） | 已解讀 | 569 |
| 93 | `overlay-12:01FC4h`（entry#87） | `overlay-12:0202Ch`（entry#87） | 已解讀 | 573 |
| 94 | `overlay-12:01FE4h`（entry#88） | `overlay-12:0204Ch`（entry#88） | 已解讀 | 641 |
| 95 | `overlay-12:0203Ch`（entry#89） | `overlay-12:020AAh`（entry#89） | 已解讀 | 641 |
| 96 | `overlay-13:04811h`（entry#32） | `overlay-13:0478Bh`（entry#32） | 已解讀 | 770 |
| 97 | `overlay-12:02078h`（entry#90） | `overlay-12:020E6h`（entry#90） | 已解讀 | 741 |
| 98 | `overlay-12:020FAh`（entry#91） | `overlay-12:0215Ch`（entry#91） | 已解讀 | 641 |
| 99 | `overlay-12:0212Ch`（entry#92） | `overlay-12:0218Eh`（entry#92） | 已解讀 | 745 |
| 100 | `overlay-12:021D5h`（entry#93） | `overlay-12:02237h`（entry#93） | 已解讀 | 642 |
| 101 | `overlay-12:02212h`（entry#94） | `overlay-12:02274h`（entry#94） | 已解讀 | 642 |
| 102 | `overlay-12:02264h`（entry#95） | `overlay-12:022C6h`（entry#95） | 已解讀 | 642 |
| 103 | `overlay-12:022A3h`（entry#96） | `overlay-12:02305h`（entry#96） | 已解讀 | 741 |
| 104 | `overlay-12:0231Eh`（entry#97） | `overlay-12:02380h`（entry#97） | 已解讀 | 569 |
| 105 | `overlay-12:02392h`（entry#99） | `overlay-12:023F4h`（entry#99） | 已解讀 | 569 |
| 106 | `overlay-12:023A2h`（entry#100） | `overlay-12:02404h`（entry#100） | 已解讀 | 569 |
| 107 | `overlay-12:023B2h`（entry#101） | `overlay-12:02414h`（entry#101） | 已解讀 | 573 |
| 108 | `overlay-12:023D8h`（entry#102） | `overlay-12:0243Ah`（entry#102） | 已解讀 | 573 |
| 109 | `overlay-12:023EFh`（entry#103） | `overlay-12:02451h`（entry#103） | 已解讀 | 569 |
| 110 | `overlay-12:023FFh`（entry#104） | `overlay-12:02461h`（entry#104） | 已解讀 | 573 |
| 111 | `overlay-12:02418h`（entry#105） | `overlay-12:0247Ah`（entry#105） | 已解讀 | 573 |
| 112 | `overlay-12:0243Bh`（entry#106） | `overlay-12:0249Dh`（entry#106） | 已解讀 | 573 |
| 113 | `overlay-12:02454h`（entry#107） | `overlay-12:024B6h`（entry#107） | 已解讀 | 642 |
| 114 | `overlay-12:02499h`（entry#108） | `overlay-12:024FBh`（entry#108） | 已解讀 | 573 |
| 115 | `overlay-12:024B9h`（entry#109） | `overlay-12:0251Bh`（entry#109） | 已解讀 | 739 |
| 116 | `overlay-12:0251Ch`（entry#110） | `overlay-12:0257Eh`（entry#110） | 已解讀 | 642 |
| 117 | `overlay-12:0255Ah`（entry#111） | `overlay-12:025BCh`（entry#111） | 已解讀 | 737 |
| 118 | `overlay-12:0259Ah`（entry#112） | `overlay-12:025FCh`（entry#112） | 已解讀 | 573 |
| 119 | `overlay-12:025BAh`（entry#113） | `overlay-12:0261Ch`（entry#113） | 已解讀 | 737 |
| 120 | `overlay-12:02606h`（entry#114） | `overlay-12:02668h`（entry#114） | 已解讀 | 737 |
| 121 | `overlay-12:02657h`（entry#115） | `overlay-12:026BDh`（entry#115） | 已解讀 | 819 |
| 122 | `overlay-12:02782h`（entry#116） | `overlay-12:027F0h`（entry#116） | 已解讀 | 741 |
| 123 | `overlay-12:027FAh`（entry#117） | `overlay-12:02868h`（entry#117） | 已解讀 | 738 |
| 124 | `overlay-12:0282Fh`（entry#118） | `overlay-12:0289Dh`（entry#118） | 已解讀 | 573 |
| 125 | `overlay-12:02855h`（entry#119） | `overlay-12:028C3h`（entry#119） | 已解讀 | 585 |
| 126 | `overlay-12:0289Ch`（entry#120） ⚠ INITSPELLS 又寫成 `overlay-22:06022h`（entry#115），生效者未判定（spec 1202） | `overlay-12:0290Ch`（entry#120） | 已解讀 | 747 |
| 127 | `overlay-12:02E33h`（entry#135） | `overlay-12:02ECBh`（entry#135） | 已解讀 | 569 |
| 128 | `overlay-22:05C86h`（entry#112）（INITSPELLS，spec 1202） | `overlay-22:06019h`（entry#113）（INITSPELLS，spec 1202） | 已解讀 | 735 |
| 129 | `overlay-12:0297Ah`（entry#121） | `overlay-12:029F2h`（entry#121） | 已解讀 | 573 |
| 130 | `overlay-12:0299Ch`（entry#122） | `overlay-12:02A14h`（entry#122） | 已解讀 | 742 |
| 131 | `overlay-22:05E11h`（entry#113）（INITSPELLS，spec 1202） | `overlay-22:061ADh`（entry#114）（INITSPELLS，spec 1202） | 已解讀 | 720 |
| 132 | `overlay-22:05F2Fh`（entry#114）（INITSPELLS，spec 1202） | `overlay-22:062D7h`（entry#115）（INITSPELLS，spec 1202） | 已解讀 | 981 |
| 133 | `overlay-12:02AA8h`（entry#124） | `overlay-12:02B20h`（entry#124） | 已解讀 | 573 |
| 134 | `overlay-12:02AD6h`（entry#125） | `overlay-12:02B4Eh`（entry#125） | 已解讀 | 738 |
| 135 | `overlay-12:02B0Fh`（entry#126） | `overlay-12:02B87h`（entry#126） | 已解讀 | 573 |
| 136 | `overlay-12:02B28h`（entry#127） | `overlay-12:02BA0h`（entry#127） | 已解讀 | 573 |
| 137 | `overlay-12:02B3Eh`（entry#128） | `overlay-12:02BB6h`（entry#128） | 已解讀 | 746 |
| 138 | `overlay-12:02C2Eh`（entry#129） | `overlay-12:02CA6h`（entry#129） | 已解讀 | 569 |
| 139 | `overlay-13:045CFh`（entry#30） | `overlay-13:0453Dh`（entry#30） | 已解讀 | 851 |
| 140 | `overlay-12:02E33h`（entry#135） | `overlay-12:02ECBh`（entry#135） | 已解讀 | 569 |
| 141 | `overlay-12:02C4Fh`（entry#130） | `overlay-12:02CC7h`（entry#130） | 已解讀 | 737 |
| 142 | `overlay-12:02C94h`（entry#131） | `overlay-12:02D0Ch`（entry#131） | 已解讀 | 738 |
| 143 | `overlay-12:02CD9h`（entry#132） | `overlay-12:02D56h`（entry#132） | 已解讀 | 742 |
| 144 | `overlay-13:04703h`（entry#31） | `overlay-13:04671h`（entry#31） | 已解讀 | 805 |
| 145 | `overlay-12:02D7Dh`（entry#133） | `overlay-12:02E15h`（entry#133） | 已解讀 | 746 |
| 146 | `overlay-12:02E33h`（entry#135） | `overlay-12:02ECBh`（entry#135） | 已解讀 | 569 |
| 147 | `overlay-12:02A2Ch`（entry#123） | `overlay-12:02AA4h`（entry#123） | 已解讀 | 743 |
