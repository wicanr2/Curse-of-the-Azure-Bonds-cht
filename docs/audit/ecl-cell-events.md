# 每格事件對照表（哪一格演哪一場）

由 `cmd/ecl-cell-events` 產生，不要手改。

原作地城 block 的每格事件是 `AND C04F, 0x7F` ＋ `ON GOTO`：**地形碼就是索引**
（0 起算）。這張表把 `ON GOTO` 的目的地與 GEO 的地形碼 join 起來。

⚠ 有對照**不等於**走過去就會演：處理常式自己可能還有守衛（once-only 旗標、
`RANDOM`、前置劇情、`SEARCH`）。逐格實際驗過的清單見
`docs/plan/seg-22-chapter6-cell-sweep-report.md`。

⚠ 索引 0 是「沒有事件的地面」，格子欄一律留白（那是全圖大半）。

## ECL1／`0x50`

沒有以地形碼分派的每格事件；改用 `GETTABLE` ＋ `ON GOTO` 查表分派（尚未解讀）

## ECL1／`0x51`

沒有以地形碼分派的每格事件；改用 `GETTABLE` ＋ `ON GOTO` 查表分派（尚未解讀）

## ECL1／`0x52`

沒有以地形碼分派的每格事件

## ECL2／`0x01`

沒有以地形碼分派的每格事件；改用 `GETTABLE` ＋ `ON GOTO` 查表分派（尚未解讀）

## ECL2／`0x02`

地圖：`GEO2/0x01`；索引 ＝ 地形碼 `& 0x3F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 7 | `07` | `(11,7)`、`(5,10)` | 「AT THE END OF THE CORRIDOR, YOU SEE A HALFLING」 |
| 0 | `00` | — | — |
| 1 | `01` | `(3,3)`、`(9,3)` | 「BEFORE YOU STANDS A BURLY MAN SURROUNDED BY SEVERAL」 |
| 2 | `02` | `(1,3)`、`(2,3)`、`(13,5)`、`(14,5)`、`(13,6)`、`(14,6)` | 「YOU HAVE FOUND THE TREASURE ROOM.」 |
| 3 | `03` | `(3,12)`、`(10,15)` | 「YOU SEE GREEN SLIMY MARKS ON THE FLOOR, MORE DISTINCT」 |
| 4 | `04` | `(12,7)`、`(0,12)`、`(2,12)` | 「YOUR ENTRY IS GREETED BY HUNGRY SNARLS.  A FIRE KNIFE」 |
| 5 | `05` | `(15,7)`、`(6,12)` | 「YOU SEE A NUMBER OF CAGES THAT ONCE HELD MONKEYS.」 |
| 6 | `06` | `(11,2)`、`(11,3)`、`(6,13)` | 「HERE ON A TABLE IS AN OPEN GUEST BOOK. THE LAST」 |

## ECL2／`0x03`

沒有以地形碼分派的每格事件；改用 `GETTABLE` ＋ `ON GOTO` 查表分派（尚未解讀）

## ECL2／`0x04`

沒有以地形碼分派的每格事件

## ECL3／`0x10`

地圖：`GEO3/0x10`；索引 ＝ 地形碼 `& 0x3F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 8 | `08` | `(13,15)` | — |
| 11 | `0B` | `(11,7)`、`(11,8)` | 「SHAMBLING MOUNDS ATTEMPT TO DRAG A CLERICS BODY AWAY.」 |
| 16 | `10` | `(14,3)` | — |
| 18 | `12` | `(7,3)` | — |
| 21 | `15` | `(4,11)` | — |
| 26 | `1A` | `(1,2)`、`(1,3)` | 「TROOPS COME BURSTING OUT OF THE COMMANDER'S」 |
| 31 | `1F` | `(11,3)` | 「NOTICES A POSSIBLE SINKHOLE」 |
| 10 | `0A` | `(11,1)` | — |
| 28 | `1C` | — | — |
| 30 | `1E` | `(8,15)` | — |
| 38 | `26` | `(11,0)` | 「YOU SEE BEFORE THE PIT CREATED BY MOANDER IN HIS」 |
| 39 | `27` | `(5,0)`、`(6,0)`、`(5,1)`、`(6,1)`、`(6,2)`、`(5,3)`、`(6,3)`、`(4,4)`、`(5,4)`、`(4,5)`、`(5,5)` | — |
| 5 | `05` | `(11,15)` | 「EAST」 |
| 7 | `07` | `(15,15)` | 「WEST」 |
| 19 | `13` | `(0,13)`、`(0,14)` | — |
| 13 | `0D` | `(6,15)` | — |
| 15 | `0F` | `(10,13)`、`(10,14)`、`(12,14)` | — |
| 34 | `22` | `(8,14)` | — |
| 9 | `09` | `(11,2)` | 「NORTH」 |
| 12 | `0C` | `(5,12)` | 「YOU'VE COME UPON A DESTROYED CHECKPOINT.  THE MARK OF ZH…」 |
| 17 | `11` | `(14,6)`、`(15,7)` | — |
| 27 | `1B` | `(3,5)` | 「YOU ARE PICKED UP AND MARCHED TO」 |
| 33 | `21` | `(9,10)` | — |
| 36 | `24` | `(11,11)` | 「NOTICES THAT THE WALLS ARE VERY」 |
| 1 | `01` | `(1,6)` | 「EAST.」 |
| 23 | `17` | `(12,5)`、`(12,6)` | 「A FILTHY GROUP HAS BEEN PICKING THROUGH THE RUBBLE.  THE…」 |
| 25 | `19` | `(5,2)` | 「THIS IS THE BARRACKS.  THE ROOM IS ABOUT HALF」 |
| 0 | `00` | — | — |
| 4 | `04` | `(3,6)` | — |
| 6 | `06` | `(13,13)` | 「SOUTH」 |
| 22 | `16` | `(15,11)` | — |
| 24 | `18` | `(3,1)` | 「THIS IS THE RED GUARDS MESS HALL.」 |
| 32 | `20` | `(0,10)` | — |
| 37 | `25` | `(6,12)` | 「A RED PLUME GUARD GROWLS, 'NOBODY'S GOING TO MAKE US GO …」 |
| 2 | `02` | `(5,6)` | 「WEST」 |
| 14 | `0E` | `(1,8)` | — |
| 20 | `14` | `(7,9)`、`(6,10)` | — |
| 29 | `1D` | `(8,2)` | — |
| 35 | `23` | `(9,0)` | — |
| 3 | `03` | `(3,8)` | 「NORTH」 |

## ECL3／`0x11`

沒有以地形碼分派的每格事件

## ECL3／`0x12`

沒有以地形碼分派的每格事件

## ECL3／`0x15`

地圖：`GEO3/0x15`；索引 ＝ 地形碼 `& 0x3F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 0 | `00` | — | — |
| 1 | `01` | `(0,2)` | — |

## ECL4／`0x20`

地圖：`GEO4/0x20`；索引 ＝ 地形碼 `& 0x3F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 25 | `19` | `(12,5)` | 「'WELL, LOOKS LIKE SHE WAS RIGHT...'」 |
| 29 | `1D` | `(10,15)` | 「A DOG PIDDLES ON YOUR LEG.」 |
| 31 | `1F` | `(12,15)` | 「TROOPERS, PRIESTS, AND MAGES ALL AT ODDS.  AND」 |
| 8 | `08` | `(3,3)`、`(2,15)` | 「YOU SEE A PAIR OF CROSSED SWORDS OVER THE DOORWAY.」 |
| 22 | `16` | `(4,2)`、`(4,3)` | 「'THEY LOOK LIKE THE ONES THE HALFLING DESCRIBED...'」 |
| 33 | `21` | `(7,2)`、`(7,3)` | 「YOU SEE A LEGLESS SOLDIER SITTING IN A LITTLE CART.」 |
| 35 | `23` | `(11,5)` | 「FRESH BLOOD IS SPLASHED AGAINST THE WALL.」 |
| 38 | `26` | `(10,14)` | 「YOU CAN'T FIND A DECENT MARK ANYWHERE. EVERYBODY'S A MAG…」 |
| 13 | `0D` | — | — |
| 7 | `07` | `(1,3)`、`(4,15)`、`(5,15)` | 「WE GOT ROOMS YA CAN STAY IN.  YA GONNA STAY?」 |
| 9 | `09` | `(3,4)`、`(8,14)`、`(9,14)`、`(8,15)`、`(9,15)` | 「WEAPONS. YOU WANNA BUY?」 |
| 12 | `0C` | `(9,4)` | 「WHA'CHU WANT? YOU BUY?」 |
| 1 | `01` | `(6,4)`、`(1,8)`、`(2,8)`、`(3,8)` | 「YOU WISH TO PURCHASE MAGIC, YES?」 |
| 5 | `05` | `(1,7)`、`(2,7)`、`(3,7)`、`(7,15)` | 「PRIESTS OF BANE RUSH FROM THE TEMPLE! OTHERS WAIT」 |
| 11 | `0B` | `(9,3)` | 「EQUIPMENT SHOP」 |
| 17 | `11` | `(13,6)` | 「THE GORGE AND GROG SHOP」 |
| 23 | `17` | `(8,2)`、`(8,3)` | 「'HEY BUDDY!  WATCH WHERE YOU'RE... OH, UH EXCUSE ME, I'M…」 |
| 26 | `1A` | `(13,8)` | 「'LOOK, ON THEIR ARMS.  THAT'S THEM...'」 |
| 28 | `1C` | `(13,12)` | 「'FZOUL'S GOT MOST OF THE MONEY LOCKED UP IN THE ALTAR. I…」 |
| 37 | `25` | `(13,13)` | 「SOME ROTTEN FOOD DROPS ON TOP OF YOU FROM AN OPEN WINDOW…」 |
| 0 | `00` | — | — |
| 10 | `0A` | `(6,3)` | 「ZHENTIL MAGIC SHOP」 |
| 20 | `14` | `(10,8)` | 「YOU HAVE BROKEN INTO THE HOME OF A PRIVATE INDIVIDUAL.」 |
| 24 | `18` | `(11,3)` | 「'HEY, YOU WANT TO BUY SOME...NO, NO, I'M SORRY, UH,」 |
| 30 | `1E` | `(13,7)` | 「YOU SEE SOME OFF DUTY ZHENTARIM SOLDIERS WITH BULGING」 |
| 15 | `0F` | — | — |
| 32 | `20` | `(6,2)`、`(5,3)` | 「'THIS CITY'S READY TO BLOW ANYTIME.'」 |
| 36 | `24` | `(11,2)`、`(10,3)` | 「'I HEAR FZOUL'S GOT DIMSWART THE SAGE LOCKED UP.'」 |
| 3 | `03` | `(9,0)`、`(11,0)`、`(10,1)`、`(4,9)` | 「YOU ARE IN THE COURTHOUSE OF ZHENTIL.」 |
| 14 | `0E` | `(7,1)` | 「HEY YOU! GET OUTA THE MAGISTRATES OFFICE!」 |
| 18 | `12` | `(14,6)` | 「YOU HAVE BROKEN INTO THE HOME OF A PRIVATE INDIVIDUAL.」 |
| 21 | `15` | `(2,2)`、`(3,2)` | 「'AREN'T THOSE THE ONES SHE TOLD US ABOUT...?'」 |
| 39 | `27` | `(10,12)` | 「A DREAM-LIKE VOICE IN YOUR HEAD SAYS, 'GREAT」 |
| 6 | `06` | `(2,3)`、`(0,8)`、`(9,8)`、`(4,10)`、`(8,13)`、`(0,14)` | 「ZHENTIL INN -- NO RED PLUMES ALLOWED!」 |
| 19 | `13` | `(1,1)`、`(4,1)`、`(12,7)`、`(14,9)`、`(11,10)`、`(12,13)`、`(14,14)` | 「YOU HAVE BROKEN INTO THE HOME OF A PRIVATE INDIVIDUAL.」 |
| 27 | `1B` | `(13,10)` | 「A WOMAN AND HER YOUNG CHILD START TO PASS CLOSE TO YOU.」 |
| 34 | `22` | `(13,15)` | 「YOU SEE A GROUP OF PEOPLE DRAG A BODY AROUND THE CORNER.」 |
| 2 | `02` | `(2,0)`、`(3,0)`、`(1,11)`、`(2,11)`、`(3,11)` | 「THE GUARDS EYE YOU SUSPICIOUSLY.」 |
| 4 | `04` | `(10,11)`、`(7,14)` | 「A FEMALE HALFLING APPEARS FROM A HIDDEN NICHE.」 |
| 16 | `10` | — | — |

## ECL4／`0x21`

地圖：`GEO4/0x00`；索引 ＝ 地形碼 `& 0x3F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 2 | `02` | — | — |
| 3 | `03` | — | 「YOU ARE ATTACKED BY THE PRIESTS OF BANE.」 |
| 4 | `04` | — | 「OLIVE SUDDENLY APPEARS IN FRONT OF YOU.」 |
| 6 | `06` | — | 「A HOODED WOMAN SUDDENLY APPEARS AND SAYS,」 |
| 7 | `07` | — | — |
| 8 | `08` | — | 「YOU NOTICE A SMALL DOOR BELOW THE ALTAR.」 |
| 9 | `09` | — | 「THIS ROOM HAS MANY MIRRORS HANGING FROM THE CEILING IN」 |
| 0 | `00` | — | — |
| 1 | `01` | — | — |
| 5 | `05` | — | 「YOU SEE AN OLD MAN IN THE CELL.  HE INTRODUCES」 |

## ECL4／`0x22`

地圖：`GEO4/0x21`；索引 ＝ 地形碼 `& 0x3F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 10 | `0A` | `(15,2)` | — |
| 11 | `0B` | `(1,7)` | — |
| 12 | `0C` | `(0,6)` | — |
| 16 | `10` | `(2,11)` | 「YOU HAVE MET UP WITH DEXAM AND HIS MINIONS!」 |
| 20 | `14` | `(0,0)` | 「A GUT WRENCHING JERK SLAMS YOUR STOMACH TO YOUR」 |
| 9 | `09` | `(14,4)` | — |
| 17 | `11` | `(5,6)`、`(6,6)`、`(7,6)`、`(8,6)`、`(6,7)`、`(7,7)`、`(8,7)`、`(5,8)`、`(6,8)`、`(7,8)`、`(8,8)`、`(5,9)`、`(6,9)`、`(7,9)`、`(8,9)` | — |
| 13 | `0D` | `(10,14)` | — |
| 2 | `02` | `(5,2)` | — |
| 8 | `08` | `(9,5)` | — |
| 19 | `13` | `(11,15)` | — |
| 0 | `00` | — | — |
| 1 | `01` | `(12,0)` | — |
| 3 | `03` | `(11,9)` | — |
| 6 | `06` | `(4,12)` | — |
| 14 | `0E` | `(8,14)` | — |
| 15 | `0F` | `(1,10)` | 「YOU HAVE MET UP WITH DEXAM AND HIS MINIONS!」 |
| 18 | `12` | `(5,7)` | — |
| 21 | `15` | `(13,1)` | 「YOU SEE THE REMAINS OF AN ELF FIGHTER.」 |
| 4 | `04` | `(6,13)` | — |
| 5 | `05` | `(2,14)` | — |
| 7 | `07` | `(6,4)` | — |

## ECL4／`0x23`

沒有以地形碼分派的每格事件

## ECL4／`0x25`

地圖：`GEO4/0x25`；索引 ＝ 地形碼 `& 0x3F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 17 | `11` | `(15,0)` | 「BEFORE YOU IS A HIGH PRIEST OF BANE.」 |
| 21 | `15` | — | 「YOU ARE ATTACKED BY」 |
| 2 | `02` | `(0,0)`、`(1,0)`、`(2,0)`、`(0,1)`、`(1,1)`、`(2,1)` | 「THIS IS THE REMAINS OF A LARGE」 |
| 14 | `0E` | `(14,2)` | 「AN IMAGE FORMS IN FRONT OF YOU.」 |
| 16 | `10` | `(15,1)` | 「UP AND DOWN.」 |
| 18 | `12` | `(6,4)` | 「A DARK SHAPE SCUTTLES THROUGH THE ARCHWAY」 |
| 20 | `14` | `(5,3)` | 「MEDUSI AND THEIR BODYGUARDS ATTACK!」 |
| 26 | `1A` | `(14,8)` | 「YOU ARE ATTACKED BY」 |
| 13 | `0D` | `(14,0)` | 「UP AND DOWN.」 |
| 28 | `1C` | `(10,7)`、`(15,8)` | 「A BAND OF FIGHTERS BURSTS THROUGH THE」 |
| 31 | `1F` | `(7,7)` | 「A MEDUSA LOOKS UP FROM A GRISLY MEAL.」 |
| 1 | `01` | `(3,0)` | 「YOU ARE IN THE FOYER OF THE MANOR.」 |
| 19 | `13` | `(6,3)` | 「NOISE COMES FROM」 |
| 25 | `19` | `(13,8)` | 「A BAND OF FIGHTERS BURSTS THROUGH THE」 |
| 32 | `20` | `(3,7)` | 「A MEDUSA LOOKS UP FROM A GRISLY MEAL.」 |
| 36 | `24` | `(11,9)` | 「A MAN COWERS IN THE SHADOWS.  HE」 |
| 37 | `25` | `(6,12)` | 「A BEHOLDER FLOATS IN FRONT OF YOU.」 |
| 39 | `27` | `(5,11)` | 「A LARGE DROW ELF EXAMINES A SMALL CARD.」 |
| 41 | `29` | `(1,14)` | 「YOU ENTER A ROOM DOMINATED BY A LARGE」 |
| 8 | `08` | `(7,1)` | — |
| 11 | `0B` | `(11,0)` | 「YOU ARE ATTACKED BY」 |
| 22 | `16` | `(12,5)` | 「AN ARROW TRAP GOES OFF!」 |
| 35 | `23` | `(6,9)` | — |
| 4 | `04` | `(0,2)` | 「UP.」 |
| 5 | `05` | `(2,4)` | 「DOWN.」 |
| 7 | `07` | `(8,2)` | 「YOU ARE ATTACKED BY」 |
| 24 | `18` | `(12,7)` | 「A BAND OF FIGHTERS BURSTS THROUGH THE」 |
| 34 | `22` | `(5,9)` | — |
| 23 | `17` | `(12,6)` | 「YOU FEEL VERY STRANGE.」 |
| 29 | `1D` | `(9,7)` | 「YOU ARE ATTACKED BY」 |
| 30 | `1E` | `(10,5)` | 「THIS ROOM IS COVERED WITH DARK RED STAINS.」 |
| 33 | `21` | `(0,10)` | 「A LARGE ROUND SHAPE SLOWLY TURNS TOWARDS YOU.」 |
| 38 | `26` | `(5,13)` | 「LOCAL GUARDS COWER BACK AT YOUR APPROACH.」 |
| 0 | `00` | — | — |
| 3 | `03` | `(4,0)`、`(5,0)`、`(4,1)`、`(5,1)`、`(4,2)`、`(5,2)` | 「THIS IS THE LIBRARY.  MANY OLD BOOKS」 |
| 9 | `09` | `(8,0)`、`(9,0)`、`(8,1)`、`(9,1)` | — |
| 27 | `1B` | — | 「YOU ARE ATTACKED BY」 |
| 40 | `28` | `(3,11)` | 「SEVERAL BEHOLDERS ARE GATHERED AROUND A」 |
| 42 | `2A` | `(4,14)` | — |
| 6 | `06` | `(6,2)` | 「DOWN.」 |
| 10 | `0A` | `(10,1)` | 「YOU ARE ATTACKED BY」 |
| 12 | `0C` | `(12,0)` | 「UP.」 |
| 15 | `0F` | `(15,2)` | 「UP AND DOWN.」 |

## ECL5／`0x30`

沒有以地形碼分派的每格事件

## ECL5／`0x31`

地圖：`GEO5/0x32`；索引 ＝ 地形碼 `& 0x7F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 10 | `0A` | `(4,2)`、`(5,2)`、`(3,3)`、`(4,3)`、`(5,3)`、`(3,4)`、`(4,4)`、`(5,4)`、`(9,10)` | — |
| 11 | `0B` | `(0,7)`、`(1,7)`、`(2,7)`、`(0,8)`、`(1,8)`、`(2,8)`、`(0,9)`、`(1,9)`、`(2,9)`、`(12,11)` | — |
| 0 | `00` | — | — |
| 2 | `02` | `(13,5)`、`(5,7)`、`(5,12)` | — |
| 3 | `03` | `(0,2)`、`(0,6)`、`(13,6)`、`(13,7)`、`(13,8)`、`(0,10)` | — |
| 5 | `05` | `(15,8)` | — |
| 7 | `07` | `(14,2)` | — |
| 9 | `09` | `(0,0)`、`(1,0)`、`(5,0)`、`(0,1)`、`(5,1)`、`(7,2)`、`(1,3)`、`(0,4)`、`(1,4)`、`(0,5)`、`(5,5)`、`(4,6)`、`(5,6)`、`(4,9)`、`(5,9)`、`(5,10)`、`(0,11)`、`(4,11)`、`(5,11)`、`(0,12)`、`(1,12)` | — |
| 1 | `01` | `(2,0)`、`(3,0)`、`(14,5)` | — |
| 4 | `04` | `(4,0)`、`(1,1)`、`(4,5)`、`(14,7)`、`(5,8)`、`(4,10)`、`(1,11)` | 「YOU BURST IN ON SOME PEASANTS WHO SCUTTLE BACK AND」 |
| 6 | `06` | `(0,3)`、`(12,3)`、`(1,5)` | 「YOU ARE IN A SMALL GENERAL STORE. THE SHOPKEEPER」 |
| 8 | `08` | `(8,4)`、`(3,13)` | 「THIS BARN IS EMPTY -- SAVE FOR THE EFREET AND HIS」 |

## ECL5／`0x32`

地圖：`GEO5/0x32`；索引 ＝ 地形碼 `& 0x7F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 4 | `04` | `(4,0)`、`(1,1)`、`(4,5)`、`(14,7)`、`(5,8)`、`(4,10)`、`(1,11)` | 「FOUR FEMALE DARK ELVES STEP FROM THE SHADOWS.」 |
| 5 | `05` | `(15,8)` | 「A DARK ELFIN WOMAN STEPS FORWARD. HER HAIR IS DARK」 |
| 6 | `06` | `(0,3)`、`(12,3)`、`(1,5)` | 「SOME DARK ELVES ARE HERE, ATOP A MOUND OF FRESHLY」 |
| 10 | `0A` | `(4,2)`、`(5,2)`、`(3,3)`、`(4,3)`、`(5,3)`、`(3,4)`、`(4,4)`、`(5,4)`、`(9,10)` | 「THE DOOR IS GUARDED BY A SALAMANDER LED PATROL.」 |
| 11 | `0B` | `(0,7)`、`(1,7)`、`(2,7)`、`(0,8)`、`(1,8)`、`(2,8)`、`(0,9)`、`(1,9)`、`(2,9)`、`(12,11)` | 「MYSTIC SYMBOLS ADORNE THE WALLS. MAGES ARE HERE」 |
| 13 | `0D` | `(13,15)` | 「THIS WAY IS GUARDED BY EFREETI AND DARK ELVES.」 |
| 14 | `0E` | `(15,13)` | 「AT YOUR APPROACH, THE ELVES COLLAPSE THE TUNNEL.」 |
| 8 | `08` | `(8,4)`、`(3,13)` | 「THIS ROOM IS FILLED WITH CLOYING INCENSE SMOKE.」 |
| 15 | `0F` | `(6,15)` | 「SILK STEPS OUT FROM THE SHADOWS.' CONGRATULATIONS」 |
| 0 | `00` | — | — |
| 1 | `01` | `(2,0)`、`(3,0)`、`(14,5)` | 「YOU FIND AN ARROW POINTING WEST FAINTLY SCRATCHED」 |
| 7 | `07` | `(14,2)` | 「YOU HAVE DISTURBED A BARRACKS FULL OF DARK ELVES」 |
| 3 | `03` | `(0,2)`、`(0,6)`、`(13,6)`、`(13,7)`、`(13,8)`、`(0,10)` | 「YOU HEAR A PATROL APPROACH FROM BEHIND YOU.」 |
| 9 | `09` | `(0,0)`、`(1,0)`、`(5,0)`、`(0,1)`、`(5,1)`、`(7,2)`、`(1,3)`、`(0,4)`、`(1,4)`、`(0,5)`、`(5,5)`、`(4,6)`、`(5,6)`、`(4,9)`、`(5,9)`、`(5,10)`、`(0,11)`、`(4,11)`、`(5,11)`、`(0,12)`、`(1,12)` | — |
| 12 | `0C` | `(7,12)` | 「CURLED IN THE CENTER OF THIS ROOM IS THE HUGE SKELETAL」 |
| 16 | `10` | `(15,11)` | — |
| 2 | `02` | `(13,5)`、`(5,7)`、`(5,12)` | 「AN ARROW MADE OF SMALL STONES POINTS SOUTH, HERE.」 |

## ECL5／`0x33`

沒有以地形碼分派的每格事件；改用 `GETTABLE` ＋ `ON GOTO` 查表分派（尚未解讀）

## ECL5／`0x35`

地圖：`GEO5/0x35`；索引 ＝ 地形碼 `& 0x3F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 0 | `00` | — | — |
| 1 | `01` | `(0,2)` | — |
| 2 | `02` | `(0,13)` | — |

## ECL6／`0x40`

地圖：`GEO6/0x40`；索引 ＝ 地形碼 `& 0x7F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 9 | `09` | `(3,1)` | 「TWO SUITS OF ARMOR FLANK THIS STAIRWAY, RADIATING」 |
| 10 | `0A` | `(1,2)` | 「AS YOU APPROACH THE STAIRS A VOICE CRIES OUT,'」 |
| 4 | `04` | `(10,4)`、`(1,9)`、`(6,12)` | 「A THRI-KREEN IS EXCAVATING A GRAVE HERE. AT YOUR」 |
| 15 | `0F` | `(9,9)` | 「GUARDS HERE PREPARE FOR COMBAT.」 |
| 16 | `10` | `(7,9)`、`(8,9)` | 「THE THRI-KREEN HAVE BIVOUACKED HERE. THEY」 |
| 17 | `11` | `(9,2)` | 「WEBS FESTOON THIS MAUSOLEUM. THE WEBS ARE INHABITED.」 |
| 18 | `12` | `(10,1)` | 「YOU SEE A FUNNEL OF WEBS.」 |
| 20 | `14` | `(14,8)` | 「GLOWING SPIDERS SKITTER FORWARD AT YOUR APPROACH.」 |
| 12 | `0C` | `(4,8)` | 「A FIGURE APPEARS FROM THE SHADOWS. 'HAIL BONDED ONES!'」 |
| 0 | `00` | — | — |
| 1 | `01` | `(3,13)`、`(3,14)` | 「AN ELFISH SPIRIT APPEARS AND GREETS YOU.」 |
| 2 | `02` | `(6,14)` | 「A RED WEB STRETCHES ACROSS THE PASSAGE,」 |
| 3 | `03` | `(13,14)` | 「A SPIRIT APPEARS BEFORE YOU. 'I AM THE SPIRIT OF」 |
| 5 | `05` | `(13,6)` | 「HE MAKES A GESTURE OF FRIENDSHIP」 |
| 6 | `06` | `(13,4)` | — |
| 7 | `07` | `(14,3)` | 「NEAR THE ENTRANCE TO THIS BUILDING IS A CRUSHED」 |
| 8 | `08` | `(5,2)`、`(4,3)`、`(5,3)` | 「NEAR THE ENTRANCE TO THIS BUILDING IS A CRUSHED」 |
| 11 | `0B` | `(1,3)` | 「A SPIRIT APPEARS BEFORE YOU.'」 |
| 13 | `0D` | `(7,7)` | — |
| 14 | `0E` | `(10,7)`、`(9,8)` | 「A PARTY OF THRI-KREEN BAR YOUR ENTRANCE.」 |
| 19 | `13` | `(12,8)`、`(12,10)` | 「AS YOU ENTER, SPIDERS COME OUT OF THE SOLID WALLS.」 |
| 21 | `15` | `(14,10)` | 「SPIDERS HAVE GATHERED A PILE OF BONES HERE」 |
| 22 | `16` | `(1,1)`、`(2,1)`、`(4,1)`、`(2,2)`、`(3,2)`、`(4,2)` | — |

## ECL6／`0x42`

地圖：`GEO6/0x42`；索引 ＝ 地形碼 `& 0x7F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 9 | `09` | `(11,11)` | 「THE BRUSH IS DENSE AND FILLED WITH RUBBLE. A FEW」 |
| 10 | `0A` | `(1,3)`、`(1,4)` | 「AS YOU STEP INTO THIS OPULENT ROOM, YOU SEE SEVERAL」 |
| 11 | `0B` | `(9,2)`、`(11,2)` | 「TWO MARGOYLES ARE TORTURING A SMALL ANIMAL IN A」 |
| 12 | `0C` | `(5,6)` | 「A LONE MARGOYLE SKITTERS AWAY UNCOVERING A SEWER」 |
| 13 | `0D` | `(9,6)` | 「A RAKSHASA RESIDES HERE IN SPLENDOR」 |
| 0 | `00` | — | — |
| 1 | `01` | `(1,12)` | 「A RAKSHASA WITH MATTED FUR AND A DOUR EXPRESSION」 |
| 2 | `02` | `(3,14)` | 「SOME MARGOYLES AND HELL HOUNDS GUARD THE ENTRANCE」 |
| 3 | `03` | `(2,14)` | 「PILED WITHIN THIS BUILDING IS A LARGE ARRAY OF」 |
| 5 | `05` | `(3,6)` | 「THE HOUNDS LEFT LITTLE THAT IS RECOGNIZABLE AS HUMAN.」 |
| 6 | `06` | `(14,3)` | 「JUST AS THE DYING MAN DESCRIBED, YOU LOCATE A CACHE.」 |
| 7 | `07` | `(13,7)`、`(6,12)` | 「NAMELESS SLIDES OUT OF THE SHADOWS, 'THIS DIRECT」 |
| 8 | `08` | `(12,9)`、`(9,10)` | 「IN THE PLAZA AHEAD IS SOME DENSE BRUSH. A SMALL CHILD」 |
| 4 | `04` | `(3,5)`、`(2,6)`、`(3,7)` | 「AHEAD, YOU SEE A MAN RUNNING, NEAR EXHUASTION,」 |

## ECL6／`0x43`

地圖：`GEO6/0x43`；索引 ＝ 地形碼 `& 0x7F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 21 | `15` | `(9,2)` | 「THE ROOM CONTAINS OLD BIERS AND CASKETS.」 |
| 23 | `17` | `(10,7)` | 「STAIRS LEAD UP HERE. DO YOU WANT TO GO UP?」 |
| 26 | `1A` | `(6,1)` | 「'THE POWER OF YOUR BONDS HAS RETURNED. GROVEL AT」 |
| 11 | `0B` | `(11,12)`、`(12,12)` | 「THIS ROOM HAS BEEN CONVERTED TO AN OFFICE. THE」 |
| 0 | `00` | — | — |
| 1 | `01` | `(6,15)` | — |
| 3 | `03` | `(7,11)` | 「AS YOU ENTER, YOU HEAR A VOICE. 'FINALLY YOU HAVE」 |
| 5 | `05` | `(5,11)` | 「AS YOU ENTER, YOU HEAR A VOICE. 'FINALLY YOU HAVE」 |
| 6 | `06` | `(0,10)` | — |
| 9 | `09` | `(8,13)`、`(9,13)` | 「YOU HAVE WALKED INTO AN ELEGANT PRIVATE CHAPEL.」 |
| 14 | `0E` | `(14,11)` | 「TIERED BEDS LINE THE WALLS, FILLED WITH MEDITATING」 |
| 2 | `02` | `(7,13)` | 「AS YOU ENTER, YOU HEAR A VOICE. 'FINALLY YOU HAVE」 |
| 7 | `07` | `(1,9)` | 「A SULFEROUS SMELL ASSAULTS YOUR NOSTRILS. THE ROOM」 |
| 8 | `08` | `(4,9)` | 「THE ROOM SEEMS FILLED WITH BONES AND HIDEOUS」 |
| 12 | `0C` | `(13,14)` | 「YOU HAVE COME INTO THE KITCHEN. SLAVES DIVE UNDER」 |
| 13 | `0D` | `(15,15)` | 「THE SEWER HAS COLLAPSED. THIS IS NO LONGER A WAY OUT.」 |
| 17 | `11` | `(15,3)`、`(14,4)` | 「RATS SKITTER UNDER BAGS AS YOU OPEN THE DOOR INTO」 |
| 18 | `12` | `(14,1)`、`(15,2)` | 「THIS ROOM IS LINED WITH SHELVES OF MOULDERING BOOKS.」 |
| 19 | `13` | `(10,1)` | 「THE ROOM CONTAINS OLD BIERS AND CASKETS.」 |
| 4 | `04` | `(6,11)` | 「AS YOU ENTER, YOU HEAR A VOICE. 'FINALLY YOU HAVE」 |
| 10 | `0A` | `(9,12)`、`(10,12)` | 「YOU HAVE ENTERED AN ELEGANT BEDROOM WITH A」 |
| 15 | `0F` | `(15,7)`、`(14,8)` | 「THE ROOM IS FILLED WITH WORSHIPPING PRIESTS. THEY」 |
| 16 | `10` | `(14,5)`、`(15,6)` | 「A HIGH PRIEST IS HERE SUMMONING UP THE POWER OF」 |
| 22 | `16` | `(10,5)` | 「THE STENCH OF PRESERVING FLUIDS IS STRONG IN HERE.」 |
| 24 | `18` | `(2,5)` | 「STAIRS LEAD DOWN HERE. DO YOU WANT TO DESCEND?」 |
| 25 | `19` | `(2,3)` | — |
| 20 | `14` | `(8,0)` | 「THE ROOM CONTAINS OLD BIERS AND CASKETS.」 |

## ECL6／`0x45`

地圖：`GEO6/0x45`；索引 ＝ 地形碼 `& 0x3F`

| 索引 | 遮罩後 | 格子 | 那一場的第一句 |
|---:|---|---|---|
| 0 | `00` | — | — |
| 1 | `01` | `(0,0)` | — |

