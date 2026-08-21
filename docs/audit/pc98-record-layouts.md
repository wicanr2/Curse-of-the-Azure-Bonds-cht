# PC-98 除錯符號裡的 record 版面

由 `cmd/borland-symbols -records` 產生，不要手改。連法與注意事項見 spec 1164。

## CHARITEMFILREC（14 欄，63 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 41 | `NAME` | STR40 |
| `+029h` | 1 | `TITLE` | （id 40） |
| `+02Ah` | 4 | `NEXT` | CHARITEMRECPTR |
| `+02Eh` | 1 | `ITEMPTR` | （id 8） |
| `+02Fh` | 3 | `NAMENUM` | （id 28） |
| `+032h` | 1 | `PLUS` | （id 4） |
| `+033h` | 1 | `PLUSSAVE` | （id 4） |
| `+034h` | 1 | `READY` | （id 40） |
| `+035h` | 1 | `IDENTIFIED` | （id 8） |
| `+036h` | 1 | `CURSED` | （id 40） |
| `+037h` | 2 | `ENCUMBERANCE` | （id 9） |
| `+039h` | 1 | `NUMITEMS` | （id 8） |
| `+03Ah` | 2 | `VALUE` | （id 9） |
| `+03Ch` | 3 | `SPECIAL` | （id 28） |

## CHARITEMREC（14 欄，103 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 81 | `NAME` | STR80 |
| `+051h` | 1 | `TITLE` | （id 40） |
| `+052h` | 4 | `NEXT` | CHARITEMRECPTR |
| `+056h` | 1 | `ITEMPTR` | （id 8） |
| `+057h` | 3 | `NAMENUM` | （id 28） |
| `+05Ah` | 1 | `PLUS` | （id 4） |
| `+05Bh` | 1 | `PLUSSAVE` | （id 4） |
| `+05Ch` | 1 | `READY` | （id 40） |
| `+05Dh` | 1 | `IDENTIFIED` | （id 8） |
| `+05Eh` | 1 | `CURSED` | （id 40） |
| `+05Fh` | 2 | `ENCUMBERANCE` | （id 9） |
| `+061h` | 1 | `NUMITEMS` | （id 8） |
| `+062h` | 2 | `VALUE` | （id 9） |
| `+064h` | 3 | `SPECIAL` | （id 28） |

## CHARREC（79 欄，423 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 16 | `NAME` | （id 3） |
| `+010h` | 12 | `STATS` | （id 28） |
| `+01Ch` | 1 | `PERCENTILESTRENGTH` | （id 8） |
| `+01Dh` | 1 | `ORGPERCENTILESTR` | （id 8） |
| `+01Eh` | 84 | `SPELLSMEMORIZED` | （id 28） |
| `+072h` | 1 | `MINREST` | （id 4） |
| `+073h` | 1 | `BASETHAC0` | （id 8） |
| `+074h` | 1 | `RACE` | CHARRACE |
| `+075h` | 1 | `CLASS` | CHARCLASS |
| `+076h` | 2 | `AGE` | （id 9） |
| `+078h` | 1 | `MAXHP` | （id 8） |
| `+079h` | 100 | `SPELLSKNOWN` | （id 28） |
| `+0DDh` | 1 | `ZEROLEVELBLOWS` | （id 8） |
| `+0DEh` | 1 | `SIZE` | （id 8） |
| `+0DFh` | 5 | `SAVINGTHROW` | （id 28） |
| `+0E4h` | 1 | `BASEMOVE` | （id 8） |
| `+0E5h` | 1 | `HIGHESTLEVEL` | LEVELTYPE |
| `+0E6h` | 1 | `HIGHESTPREVLEVEL` | LEVELTYPE |
| `+0E7h` | 1 | `LOSTLEVELS` | （id 8） |
| `+0E8h` | 1 | `LOSTHP` | （id 8） |
| `+0E9h` | 1 | `UNDEADLEVEL` | UNDEADTYPE |
| `+0EAh` | 8 | `THIEFABILITY` | （id 28） |
| `+0F2h` | 4 | `SPECIALS` | EFFECTPTR |
| `+0F6h` | 1 | `RAISED` | （id 8） |
| `+0F7h` | 1 | `MASTERCONTROL` | （id 8） |
| `+0F8h` | 1 | `MODIFIED` | （id 40） |
| `+0F9h` | 1 | `OLDCLASS` | CHARCLASS |
| `+0FAh` | 1 | `OLDLEVEL` | LEVELTYPE |
| `+0FBh` | 14 | `WEALTH` | （id 28） |
| `+109h` | 8 | `CURRENTLEVEL` | （id 28） |
| `+111h` | 8 | `PREVIOUSLEVEL` | （id 28） |
| `+119h` | 1 | `GENDER` | CHARGENDER |
| `+11Ah` | 1 | `RACETYPE` | （id 8） |
| `+11Bh` | 1 | `ALIGNMENT` | CHARALIGNMENT |
| `+11Ch` | 2 | `BASEATTBLOWS` | （id 28） |
| `+11Eh` | 2 | `BASEATTDICE` | （id 28） |
| `+120h` | 2 | `BASEATTSIDES` | （id 28） |
| `+122h` | 2 | `BASEATTADDS` | （id 28） |
| `+124h` | 1 | `BASEAC` | （id 8） |
| `+125h` | 1 | `USESSTRENGTHFLAG` | （id 40） |
| `+126h` | 1 | `RANDOMID` | （id 8） |
| `+127h` | 4 | `EXPERIENCE` | （id 6） |
| `+12Bh` | 1 | `CLASSCODE` | （id 8） |
| `+12Ch` | 1 | `ROLLEDHP` | （id 8） |
| `+12Dh` | 15 | `MAXSPELLSPERLEVEL` | （id 28） |
| `+13Ch` | 2 | `BASEEXP` | （id 9） |
| `+13Eh` | 1 | `EXPPERHP` | （id 8） |
| `+13Fh` | 1 | `HEAD` | （id 8） |
| `+140h` | 1 | `BODY` | （id 8） |
| `+141h` | 1 | `ICONHEAD` | （id 8） |
| `+142h` | 1 | `ICONBODY` | （id 8） |
| `+143h` | 1 | `ICONINDEX` | （id 8） |
| `+144h` | 1 | `ICONHEIGHT` | （id 8） |
| `+145h` | 7 | `COLORLIST` | CHARCOLORTABLE |
| `+14Ch` | 1 | `MONSTERTYPE` | （id 8） |
| `+14Dh` | 1 | `NUMITEMS` | （id 8） |
| `+14Eh` | 4 | `CHARACTERITEM` | CHARITEMRECPTR |
| `+152h` | 52 | `SLOT` | （id 28） |
| `+186h` | 1 | `HANDS` | （id 8） |
| `+187h` | 1 | `PLUSSAVES` | （id 4） |
| `+188h` | 2 | `ENCUMBERANCE` | （id 9） |
| `+18Ah` | 4 | `NEXT` | CHARRECPTR |
| `+18Eh` | 4 | `COMBAT` | COMBATVARPTR |
| `+192h` | 1 | `PDLNREMOVECURSE` | （id 8） |
| `+193h` | 1 | `DUM1` | （id 8） |
| `+194h` | 1 | `DUM2` | （id 8） |
| `+195h` | 1 | `DUM3` | （id 8） |
| `+196h` | 1 | `STATUS` | CHARSTATUS |
| `+197h` | 1 | `STATUSOK` | （id 40） |
| `+198h` | 1 | `ALLEGIANCE` | COMBATSIDES |
| `+199h` | 1 | `COMPUTERCONTROLLED` | （id 40） |
| `+19Ah` | 1 | `THAC0` | （id 8） |
| `+19Bh` | 2 | `AC` | （id 28） |
| `+19Dh` | 2 | `ATTBLOWS` | （id 28） |
| `+19Fh` | 2 | `ATTDICE` | （id 28） |
| `+1A1h` | 2 | `ATTSIDES` | （id 28） |
| `+1A3h` | 2 | `ATTADDS` | （id 28） |
| `+1A5h` | 1 | `CURRENTHP` | （id 8） |
| `+1A6h` | 1 | `MOVE` | （id 8） |

## CLOUDREC（8 欄，30 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 4 | `CASTER` | CHARRECPTR |
| `+004h` | 4 | `NEXT` | CLOUDPTR |
| `+008h` | 9 | `OLDTERR` | （id 28） |
| `+011h` | 9 | `DOSQUARE` | （id 28） |
| `+01Ah` | 1 | `ORIGINX` | （id 4） |
| `+01Bh` | 1 | `ORIGINY` | （id 4） |
| `+01Ch` | 1 | `WHICHCLOUD` | （id 8） |
| `+01Dh` | 1 | `DISPELLED` | （id 40） |

## COMBATVARREC（19 欄，22 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 1 | `CASTING` | （id 8） |
| `+001h` | 1 | `CANCAST` | （id 40） |
| `+002h` | 1 | `CANUSE` | （id 40） |
| `+003h` | 1 | `INIT` | （id 4） |
| `+004h` | 1 | `ATTTYPE` | （id 8） |
| `+005h` | 1 | `ZEROLEVBLOWS` | （id 8） |
| `+006h` | 1 | `MOVELEFT` | （id 8） |
| `+007h` | 1 | `GUARDING` | （id 40） |
| `+008h` | 1 | `ATTACKEDYET` | （id 40） |
| `+009h` | 1 | `FACING` | （id 8） |
| `+00Ah` | 4 | `TARGET` | CHARRECPTR |
| `+00Eh` | 1 | `DYING` | （id 8） |
| `+00Fh` | 1 | `NUMATTACKERS` | （id 8） |
| `+010h` | 1 | `TURNED` | （id 40） |
| `+011h` | 1 | `TRIEDTOTURN` | （id 40） |
| `+012h` | 1 | `ATTACKSPACING` | （id 8） |
| `+013h` | 1 | `CHARTYPE` | （id 41） |
| `+014h` | 1 | `ROUTING` | （id 40） |
| `+015h` | 1 | `PATHCHECKTYPE` | （id 8） |

## DATETIME（6 欄，12 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 2 | `YEAR` | （id 9） |
| `+002h` | 2 | `MONTH` | （id 9） |
| `+004h` | 2 | `DAY` | （id 9） |
| `+006h` | 2 | `HOUR` | （id 9） |
| `+008h` | 2 | `MIN` | （id 9） |
| `+00Ah` | 2 | `SEC` | （id 9） |

## DMYPTR

展不開：record "DMYPTR" first member index 0 is outside the member table

## DOT（11 欄，23 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 2 | `X` | （id 5） |
| `+002h` | 2 | `Y` | （id 5） |
| `+004h` | 2 | `TX` | （id 5） |
| `+006h` | 2 | `TY` | （id 5） |
| `+008h` | 2 | `RX` | （id 5） |
| `+00Ah` | 2 | `RY` | （id 5） |
| `+00Ch` | 2 | `VX` | （id 5） |
| `+00Eh` | 2 | `VY` | （id 5） |
| `+010h` | 1 | `STAGE` | （id 8） |
| `+011h` | 1 | `OLD` | （id 8） |
| `+012h` | 5 | `TIME` | （id 28） |

## EFFECTREC（5 欄，9 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 1 | `EFFECTNUM` | （id 8） |
| `+001h` | 2 | `DURATION` | （id 9） |
| `+003h` | 1 | `SPECIAL` | （id 8） |
| `+004h` | 1 | `SPECIALOFF` | （id 40） |
| `+005h` | 4 | `NEXT` | EFFECTPTR |

## FILEREC（6 欄，128 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 2 | `HANDLE` | （id 9） |
| `+002h` | 2 | `MODE` | （id 9） |
| `+004h` | 2 | `RECSIZE` | （id 9） |
| `+006h` | 26 | `PRIVATE` | （id 28） |
| `+020h` | 16 | `USERDATA` | （id 28） |
| `+030h` | 80 | `NAME` | （id 28） |

## FRAME（2 欄，8 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 4 | `DELAY` | （id 6） |
| `+004h` | 4 | `GRAPHDATA` | GRAFRECPTR |

## FREEREC（2 欄，8 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 4 | `ORGPTR` | （id 22） |
| `+004h` | 4 | `ENDPTR` | （id 22） |

## GOSUB_REC（2 欄，6 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 2 | `MEMADD` | （id 9） |
| `+002h` | 4 | `NEXT_ADD` | GOSUB_PTR |

## GRAFREC（9 欄，65524 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 2 | `H` | （id 9） |
| `+002h` | 2 | `W` | （id 9） |
| `+004h` | 2 | `XORG` | （id 5） |
| `+006h` | 2 | `YORG` | （id 5） |
| `+008h` | 1 | `COUNT` | （id 8） |
| `+009h` | 8 | `CMAP` | CMAPTYPE |
| `+011h` | 2 | `SIZE` | （id 9） |
| `+013h` | 4 | `MASK` | GRAFDATAPTR |
| `+017h` | 65501 | `DATA` | GRAFDATA |

## HILLCHAR（71 欄，188 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 4 | `RANDNUMB` | （id 6） |
| `+004h` | 16 | `CHARNAME` | （id 28） |
| `+014h` | 1 | `STRENGTH` | （id 8） |
| `+015h` | 1 | `STRPERCENT` | （id 8） |
| `+016h` | 1 | `INTELLIGENCE` | （id 8） |
| `+017h` | 1 | `WISDOM` | （id 8） |
| `+018h` | 1 | `DEXTERITY` | （id 8） |
| `+019h` | 1 | `CONSTITUTION` | （id 8） |
| `+01Ah` | 1 | `CHARISMA` | （id 8） |
| `+01Bh` | 1 | `LEVEL` | （id 8） |
| `+01Ch` | 1 | `ALIGNMENT` | （id 8） |
| `+01Dh` | 1 | `CHIME` | （id 8） |
| `+01Eh` | 2 | `AGE` | （id 9） |
| `+020h` | 1 | `HP` | （id 8） |
| `+021h` | 1 | `MAXHP` | （id 8） |
| `+022h` | 1 | `BASEHP` | （id 8） |
| `+023h` | 1 | `RODOFBLASTING` | （id 8） |
| `+024h` | 1 | `SSICLASS` | （id 8） |
| `+025h` | 1 | `SPECIALITEM` | （id 8） |
| `+026h` | 1 | `HARPER` | （id 8） |
| `+027h` | 1 | `UNUSED2` | （id 8） |
| `+028h` | 4 | `GOLD` | （id 6） |
| `+02Ch` | 1 | `SEX` | （id 8） |
| `+02Dh` | 1 | `RACE` | （id 8） |
| `+02Eh` | 4 | `EXPERIENCE` | （id 6） |
| `+032h` | 1 | `SHADOWS` | （id 8） |
| `+033h` | 1 | `PICKPOCKETS` | （id 8） |
| `+034h` | 1 | `CLIMB` | （id 8） |
| `+035h` | 1 | `CLASS` | （id 8） |
| `+036h` | 1 | `XPOS` | （id 8） |
| `+037h` | 1 | `YPOS` | （id 8） |
| `+038h` | 1 | `FACING` | （id 8） |
| `+039h` | 1 | `STAGE` | （id 8） |
| `+03Ah` | 1 | `RANDSEED` | （id 8） |
| `+03Bh` | 1 | `UNUSED3` | （id 8） |
| `+03Ch` | 2 | `TIMER` | （id 9） |
| `+03Eh` | 2 | `DAY` | （id 9） |
| `+040h` | 4 | `SECONDS` | （id 6） |
| `+044h` | 1 | `TIME` | （id 8） |
| `+045h` | 1 | `PICKS` | （id 8） |
| `+046h` | 60 | `PICKDATA` | （id 28） |
| `+082h` | 1 | `LOCKSTRIED` | （id 8） |
| `+083h` | 1 | `LOCKSPICKED` | （id 8） |
| `+084h` | 1 | `MAZESTRIED` | （id 8） |
| `+085h` | 1 | `MAZESCOMPLETED` | （id 8） |
| `+086h` | 1 | `SCROLL` | （id 8） |
| `+087h` | 1 | `POTION` | （id 8） |
| `+088h` | 1 | `WHERE` | （id 8） |
| `+089h` | 18 | `DOORS` | （id 28） |
| `+09Bh` | 1 | `QUAD` | （id 8） |
| `+09Ch` | 1 | `EQUEST` | （id 8） |
| `+09Dh` | 1 | `STEP` | （id 8） |
| `+09Eh` | 1 | `MAXSTEP` | （id 8） |
| `+09Fh` | 1 | `ARENA` | （id 8） |
| `+0A0h` | 2 | `SCORE` | （id 9） |
| `+0A2h` | 2 | `HISCORE` | （id 9） |
| `+0A4h` | 4 | `INBANK` | （id 6） |
| `+0A8h` | 1 | `EQUESTTRIED` | （id 8） |
| `+0A9h` | 1 | `EQUESTFINISHED` | （id 8） |
| `+0AAh` | 1 | `HORSE` | （id 8） |
| `+0ABh` | 1 | `SLEEP` | （id 8） |
| `+0ACh` | 7 | `PUBINFO` | （id 28） |
| `+0B3h` | 1 | `ARLEVEL` | （id 8） |
| `+0B4h` | 1 | `PREVCLUE` | （id 8） |
| `+0B5h` | 1 | `QUESTNUM` | （id 8） |
| `+0B6h` | 1 | `GUILDVISITS` | （id 8） |
| `+0B7h` | 1 | `CLEVEL` | （id 8） |
| `+0B8h` | 1 | `MLEVEL` | （id 8） |
| `+0B9h` | 1 | `FLEVEL` | （id 8） |
| `+0BAh` | 1 | `TLEVEL` | （id 8） |
| `+0BBh` | 1 | `PREVARLEVEL` | （id 8） |

## ITEMREC（16 欄，16 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 1 | `ITEMTYPE` | （id 8） |
| `+001h` | 1 | `HANDS` | （id 8） |
| `+002h` | 1 | `DICELARGE` | （id 8） |
| `+003h` | 1 | `SIDESLARGE` | （id 8） |
| `+004h` | 1 | `ADDSLARGE` | （id 4） |
| `+005h` | 1 | `NUMMISSLESHOTS` | （id 8） |
| `+006h` | 1 | `ARMOR` | （id 8） |
| `+007h` | 1 | `DAMAGETYPE` | （id 8） |
| `+008h` | 1 | `LOADPOINTS` | （id 8） |
| `+009h` | 1 | `DICENORMAL` | （id 8） |
| `+00Ah` | 1 | `SIDESNORMAL` | （id 8） |
| `+00Bh` | 1 | `ADDSNORMAL` | （id 4） |
| `+00Ch` | 1 | `MAXRANGE` | （id 8） |
| `+00Dh` | 1 | `CLASSCODE` | （id 8） |
| `+00Eh` | 1 | `MISSLETYPE` | （id 8） |
| `+00Fh` | 1 | `TEMP` | （id 8） |

## LISTITEM（3 欄，86 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 81 | `NAME` | STR80 |
| `+051h` | 1 | `TITLE` | （id 40） |
| `+052h` | 4 | `NEXT` | LISTITEMPTR |

## LOADSTATUS（2 欄，2 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 1 | `SPRITELOADED` | （id 40） |
| `+001h` | 1 | `PICLOADED` | （id 40） |

## MAP3D（4 欄，1024 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 256 | `NORTHEAST` | （id 28） |
| `+100h` | 256 | `SOUTHWEST` | （id 28） |
| `+200h` | 256 | `SPECIAL` | （id 28） |
| `+300h` | 256 | `BLOCK` | （id 28） |

## POOLCHAR（71 欄，285 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 16 | `NAME` | （id 3） |
| `+010h` | 6 | `STATS` | （id 28） |
| `+016h` | 1 | `PERCENTILESTRENGTH` | （id 8） |
| `+017h` | 21 | `SPELLSMEMORIZED` | （id 28） |
| `+02Ch` | 1 | `MINREST` | （id 4） |
| `+02Dh` | 1 | `BASETHAC0` | （id 8） |
| `+02Eh` | 1 | `RACE` | CHARRACE |
| `+02Fh` | 1 | `CLASS` | CHARCLASS |
| `+030h` | 2 | `AGE` | （id 9） |
| `+032h` | 1 | `MAXHP` | （id 8） |
| `+033h` | 56 | `SPELLSKNOWN` | （id 28） |
| `+06Bh` | 1 | `ZEROLEVELBLOWS` | （id 8） |
| `+06Ch` | 1 | `SIZE` | （id 8） |
| `+06Dh` | 5 | `SAVINGTHROW` | （id 28） |
| `+072h` | 1 | `BASEMOVE` | （id 8） |
| `+073h` | 1 | `HIGHESTLEVEL` | LEVELTYPE |
| `+074h` | 1 | `LOSTLEVELS` | （id 8） |
| `+075h` | 1 | `LOSTHP` | （id 8） |
| `+076h` | 1 | `UNDEADLEVEL` | UNDEADTYPE |
| `+077h` | 8 | `THIEFABILITY` | （id 28） |
| `+07Fh` | 4 | `SPECIALS` | EFFECTPTR |
| `+083h` | 1 | `RAISED` | （id 8） |
| `+084h` | 1 | `MASTERCONTROL` | （id 8） |
| `+085h` | 1 | `MODIFIED` | （id 40） |
| `+086h` | 1 | `OLDCLASS` | CHARCLASS |
| `+087h` | 1 | `OLDLEVEL` | LEVELTYPE |
| `+088h` | 14 | `WEALTH` | （id 28） |
| `+096h` | 8 | `CURRENTLEVEL` | （id 28） |
| `+09Eh` | 1 | `GENDER` | CHARGENDER |
| `+09Fh` | 1 | `RACETYPE` | （id 8） |
| `+0A0h` | 1 | `ALIGNMENT` | CHARALIGNMENT |
| `+0A1h` | 2 | `BASEATTBLOWS` | （id 28） |
| `+0A3h` | 2 | `BASEATTDICE` | （id 28） |
| `+0A5h` | 2 | `BASEATTSIDES` | （id 28） |
| `+0A7h` | 2 | `BASEATTADDS` | （id 28） |
| `+0A9h` | 1 | `BASEAC` | （id 8） |
| `+0AAh` | 1 | `USESSTRENGTHFLAG` | （id 40） |
| `+0ABh` | 1 | `RANDOMID` | （id 8） |
| `+0ACh` | 4 | `EXPERIENCE` | （id 6） |
| `+0B0h` | 1 | `CLASSCODE` | （id 8） |
| `+0B1h` | 1 | `ROLLEDHP` | （id 8） |
| `+0B2h` | 6 | `MAXSPELLSPERLEVEL` | （id 28） |
| `+0B8h` | 2 | `BASEEXP` | （id 9） |
| `+0BAh` | 1 | `EXPPERHP` | （id 8） |
| `+0BBh` | 1 | `HEAD` | （id 8） |
| `+0BCh` | 1 | `BODY` | （id 8） |
| `+0BDh` | 1 | `ICONHEAD` | （id 8） |
| `+0BEh` | 1 | `ICONBODY` | （id 8） |
| `+0BFh` | 1 | `ICONINDEX` | （id 8） |
| `+0C0h` | 1 | `ICONHEIGHT` | （id 8） |
| `+0C1h` | 6 | `COLORLIST` | PCHARCOLORTABLE |
| `+0C7h` | 1 | `NUMITEMS` | （id 8） |
| `+0C8h` | 4 | `CHARACTERITEM` | CHARITEMRECPTR |
| `+0CCh` | 52 | `SLOT` | （id 28） |
| `+100h` | 1 | `HANDS` | （id 8） |
| `+101h` | 1 | `PLUSSAVES` | （id 4） |
| `+102h` | 2 | `ENCUMBERANCE` | （id 9） |
| `+104h` | 4 | `NEXT` | CHARRECPTR |
| `+108h` | 4 | `COMBAT` | COMBATVARPTR |
| `+10Ch` | 1 | `STATUS` | CHARSTATUS |
| `+10Dh` | 1 | `STATUSOK` | （id 40） |
| `+10Eh` | 1 | `ALLEGIANCE` | COMBATSIDES |
| `+10Fh` | 1 | `COMPUTERCONTROLLED` | （id 40） |
| `+110h` | 1 | `THAC0` | （id 8） |
| `+111h` | 2 | `AC` | （id 28） |
| `+113h` | 2 | `ATTBLOWS` | （id 28） |
| `+115h` | 2 | `ATTDICE` | （id 28） |
| `+117h` | 2 | `ATTSIDES` | （id 28） |
| `+119h` | 2 | `ATTADDS` | （id 28） |
| `+11Bh` | 1 | `CURRENTHP` | （id 8） |
| `+11Ch` | 1 | `MOVE` | （id 8） |

## REGISTERS（10 欄，20 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 2 | `AX` | （id 9） |
| `+002h` | 2 | `BX` | （id 9） |
| `+004h` | 2 | `CX` | （id 9） |
| `+006h` | 2 | `DX` | （id 9） |
| `+008h` | 2 | `BP` | （id 9） |
| `+00Ah` | 2 | `SI` | （id 9） |
| `+00Ch` | 2 | `DI` | （id 9） |
| `+00Eh` | 2 | `DS` | （id 9） |
| `+010h` | 2 | `ES` | （id 9） |
| `+012h` | 2 | `FLAGS` | （id 9） |

## SEARCHREC（5 欄，43 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 21 | `FILL` | （id 28） |
| `+015h` | 1 | `ATTR` | （id 8） |
| `+016h` | 4 | `TIME` | （id 6） |
| `+01Ah` | 4 | `SIZE` | （id 6） |
| `+01Eh` | 13 | `NAME` | （id 3） |

## SEQUENCE（4 欄，67 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 1 | `FRAMECOUNT` | （id 8） |
| `+001h` | 1 | `ORGCOUNT` | （id 8） |
| `+002h` | 1 | `CURRENTFRAME` | （id 8） |
| `+003h` | 64 | `FRAMES` | （id 28） |

## SIGHTREC（3 欄，3 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 1 | `ID` | （id 8） |
| `+001h` | 1 | `RANGE` | （id 8） |
| `+002h` | 1 | `DIRECTION` | （id 8） |

## SPELLREC（16 欄，16 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 1 | `CLASS` | SPELLCLASSTYPE |
| `+001h` | 1 | `LEVEL` | LEVELTYPE |
| `+002h` | 1 | `RANGEFIXED` | （id 8） |
| `+003h` | 1 | `RANGEPERLEVEL` | （id 8） |
| `+004h` | 1 | `DURATIONFIXED` | （id 8） |
| `+005h` | 1 | `DURATIONPERLEVEL` | （id 8） |
| `+006h` | 1 | `AOECOMBAT` | （id 8） |
| `+007h` | 1 | `AOENONCOMBAT` | （id 8） |
| `+008h` | 1 | `SAVERESULT` | SAVERESULTS |
| `+009h` | 1 | `SAVEVS` | SAVINGTHROWTYPE |
| `+00Ah` | 1 | `EFFECTNUM` | （id 8） |
| `+00Bh` | 1 | `WHERECAST` | （id 41） |
| `+00Ch` | 1 | `CASTINGTIME` | （id 8） |
| `+00Dh` | 1 | `PRIORITY` | （id 8） |
| `+00Eh` | 1 | `CASTON` | （id 41） |
| `+00Fh` | 1 | `MINRANGE` | （id 8） |

## TACTICALMAP（8 欄，1257 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 1 | `H` | （id 8） |
| `+001h` | 1 | `V` | （id 8） |
| `+002h` | 1 | `VX` | （id 4） |
| `+003h` | 1 | `VY` | （id 4） |
| `+004h` | 1 | `CURSORON` | （id 40） |
| `+005h` | 1 | `CURSORSIZE` | （id 8） |
| `+006h` | 1 | `XRAY` | （id 40） |
| `+007h` | 1250 | `TD` | （id 28） |

## TARGETREC（4 欄，7 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 4 | `WHO` | CHARRECPTR |
| `+004h` | 1 | `X` | （id 4） |
| `+005h` | 1 | `Y` | （id 4） |
| `+006h` | 1 | `SPECIAL` | （id 8） |

## TDEFTYPE（4 欄，4 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 1 | `MV` | （id 8） |
| `+001h` | 1 | `HT` | （id 8） |
| `+002h` | 1 | `LOS` | （id 8） |
| `+003h` | 1 | `SYM` | （id 8） |

## TEXTREC（14 欄，256 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 2 | `HANDLE` | （id 9） |
| `+002h` | 2 | `MODE` | （id 9） |
| `+004h` | 2 | `BUFSIZE` | （id 9） |
| `+006h` | 2 | `PRIVATE` | （id 9） |
| `+008h` | 2 | `BUFPOS` | （id 9） |
| `+00Ah` | 2 | `BUFEND` | （id 9） |
| `+00Ch` | 4 | `BUFPTR` | （id 22） |
| `+010h` | 4 | `OPENFUNC` | （id 22） |
| `+014h` | 4 | `INOUTFUNC` | （id 22） |
| `+018h` | 4 | `FLUSHFUNC` | （id 22） |
| `+01Ch` | 4 | `CLOSEFUNC` | （id 22） |
| `+020h` | 16 | `USERDATA` | （id 28） |
| `+030h` | 80 | `NAME` | （id 28） |
| `+080h` | 128 | `BUFFER` | TEXTBUF |

## TREASUREREC（2 欄，32 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 28 | `MONEY` | POOLTABLE |
| `+01Ch` | 4 | `ITEMS` | CHARITEMRECPTR |

## VECTOR（13 欄，24 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 2 | `VX1` | （id 5） |
| `+002h` | 2 | `VY1` | （id 5） |
| `+004h` | 2 | `VX2` | （id 5） |
| `+006h` | 2 | `VY2` | （id 5） |
| `+008h` | 2 | `E` | （id 5） |
| `+00Ah` | 2 | `DX` | （id 5） |
| `+00Ch` | 2 | `DY` | （id 5） |
| `+00Eh` | 2 | `X` | （id 5） |
| `+010h` | 2 | `Y` | （id 5） |
| `+012h` | 2 | `XINC` | （id 5） |
| `+014h` | 2 | `YINC` | （id 5） |
| `+016h` | 1 | `RANGE` | （id 8） |
| `+017h` | 1 | `DIRECTION` | （id 8） |

## WALLSET（1 欄，2340 bytes）

| 位移 | 長度 | 欄位 | 型別 |
|---|---:|---|---|
| `+000h` | 2340 | `WSET` | （id 28） |

