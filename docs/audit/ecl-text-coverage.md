# ECL 文字覆蓋（原作文字段落 → game pack）

由 `cmd/ecl-text-coverage` 產生，不要手改。

- 分母是**控制流走訪**得到的頁，不是位址順序切出來的段：從五個 lifecycle entry 出發，跟循序、`GOTO`／`GOSUB`／`RETURN`、`IF` 的兩條路（spec 1106）、以及 `25h ON GOTO`／`26h ON GOSUB` 的每一個目的地。`15h`／`2Bh` 選單與 `25h`／`26h` 是變長指令，長度一律用 `ecl.RecordEnd`——它們的 `Instruction.Next` 指向自己的第一個運算元，信它會把資料當程式解而且不報錯。
- 分母與比對**走兩趟**：`walkPages` 不帶文字（狀態只有位址、子程式摘要與一個位元），整份 corpus 都走得完，所以「有哪些頁」是完整的；`walkRuns` 帶文字用來比對，碰到狀態上限只會**少判**，不會讓頁從分母裡消失。兩者的差額由 `TestPageWalkCoversEveryPageTheRunWalkFinds` 印出來（目前 0 頁）。
- 比對以 **run** 為單位：runtime 把一次執行累積的文字**一次**交給 `MatchText`，所以一條規則可以橫跨好幾頁（開場捲軸就是七頁一條）。run 在會 `return result` 的指令處結束（選單、戰鬥、寶物、輸入、換 block、離開）。
- 一條 run 命中，它經過的每一頁就都算接上了——規則命中的是那一整份文字，不是其中某一頁。反過來，同一份文字可能印在好幾個位址上，那些位址一律算進同一條 run。
- `variable-insert` 是**唯一**還無法靜態驗證的一類：頁裡印的是執行期的值（城名、酒館傳聞編號、隊員名），靜態文字裡沒有那幾個字，規則要靠它才會命中。逐城的 `*.edge`、`*.tavern-tale-*`、`world-route.*`、`world.night-note.*` 都屬於這一類，要靠實機路徑驗。
- `subroutine` 是共用子程式的片段（`WHAT DO YOU DO?`、`UP`／`DOWN`）。判準是「落在被 `GOSUB` 呼叫的範圍內」**且**「從來沒有和別的頁同屬一份 run」——實機它一定被併進呼叫端那一頁。**不要**替它們寫規則：只有一兩個字的 `all_contains` 會攔截到別的文字。
- `gosub_inserts`／`dynamic_branch` 只是**註記**，不是狀態：那兩種插入已經在走訪時展開，展開後沒命中就照樣算 `unmatched`。兩個計數與 `matched`／`unmatched` 有重疊，加起來不等於 `groups`。
- ⚠ `ECL1.DAX/0x50`／`0x51` 的**比對**走訪會碰到狀態上限（4,000,000）而提早停：世界地圖那一段 `IF` 分岔多又長。碰到上限會在 stderr 印一行，**沒有靜默截斷**；分母不受影響（見上一條），代價只是那個 block 可能有頁比對不到而算成待辦。實測目前兩趟走訪找到的頁完全相同，代價為 0。
- 『已接上』只表示有一條 text_rule 的 all_contains 全部命中，不代表譯文正確或事件副作用已還原。

## 摘要

| 處置 | 數量 | 意思 |
|---|---:|---|
| 控制流可達的頁 | 1022 | 分母 |
| `matched` | 999 | 有規則命中它所在的 run |
| **`unmatched`** | **0** | **還沒寫規則——這才是待辦** |
| `variable-insert` | 16 | 頁裡印的是執行期的值（城名、傳聞編號、隊員名），靜態驗不到 |
| `subroutine` | 7 | 共用子程式的片段，實機一定被併進呼叫端那一頁 |

下面兩欄是**註記**不是狀態：它們與上表重疊，不要加總。

| 註記 | 數量 | 意思 |
|---|---:|---|
| `gosub_inserts` | 32 | 這一頁的文字被 `GOSUB` 進去的子程式插過一段（走訪時已展開） |
| `dynamic_branch` | 228 | 這一頁所在的 run 有 `ON GOTO`／`ON GOSUB`（走訪時已展開每個目的地） |

## 未接上的段落，依 block

| Block | 未接上 |
|---|---:|

## 逐段

| Block | offset | 處置 | 已接上的規則 | 原作文字 |
|---|---|---|---|---|
| `ECL1.DAX/0x50` | `0x00C8` | `variable-insert` | — | YOU ARE AT THE EDGE OF . WILL YOU ENTER OR CONTINUE YOUR JOURNEY? |
| `ECL1.DAX/0x50` | `0x01C4` | `matched` | `standing-stone.grey-man` | YOU ARE AT THE STANDING STONES. A GREY ROBED MAN SITS AGAINST A STONE, HIS FACE COVERED IN SHADOW. |
| `ECL1.DAX/0x50` | `0x0226` | `matched` | `myth-drannor.tyranthraxus-reveal` | HE SPEAKS, ' YOU PRESENTLY SERVE MASTER . AS HE RISES, THE ROBE FALLS AWAY, REVEALING TYRANTHRAXUS. 'YOU HAVE DONE WELL. MEET ME AT MYTH DRANNOR.' HE FADES AWAY… ⚠ 另有 `GOSUB 0x0268` 印出的一段 |
| `ECL1.DAX/0x50` | `0x030B` | `matched` | `standing-stone.seek-red` | HE CONTINUES, ' SEEK GREEN TO THE NORTHWEST.' |
| `ECL1.DAX/0x50` | `0x031E` | `matched` | `standing-stone.seek-red` | THE ROBE COLLAPSES WHEN YOU ATTACK. FROM ABOVE COMES A VOICE, ' SEEK GREEN TO THE NORTHWEST.' |
| `ECL1.DAX/0x50` | `0x056C` | `matched` | `tilverton.entry-barred` | GUARDS BAR YOUR WAY. |
| `ECL1.DAX/0x50` | `0x0586` | `matched` | `world.town.sewer-entrance` | YOU FIND AN ENTRANCE TO THE SEWERS. |
| `ECL1.DAX/0x50` | `0x05EC` | `matched` | `world.tavern.drank-the-night-away` | YOU ARE IN . WHAT PLACE WILL YOU VISIT? |
| `ECL1.DAX/0x50` | `0x073B` | `matched` | `ashabenford.barn-for-zhentil` | 'DON'T THINK MUCH OF ZHENTIL KEEPERS. THE BARN'S THE BEST YOU'LL GET.' |
| `ECL1.DAX/0x50` | `0x07A2` | `matched` | `essembra.branching-oak` | 'WELCOME TO THE ASHABENFORD ARMS.' |
| `ECL1.DAX/0x50` | `0x0892` | `matched` | `journal-trigger.shadowdale-warning-18` | A HOODED, GREY ROBED MAN SITS IN A DARK CORNER. HE MOTIONS YOU OVER, SPEAKS, AND YOU RECORD IT IN JOURNAL ENTRY 18. THEN, HE VANISHES. ⚠ 另有 `GOSUB 0x08D2` 印出的一段 |
| `ECL1.DAX/0x50` | `0x0963` | `matched` | `ashabenford.tavern-turns-friendly` | THE PATRONS BECOME SURLY. THE BARTENDER SUGGESTS YOU LEAVE. A FARMER STANDS UP AND TELLS THEM OF YOUR DEFEAT OF THE ETTINS. THEY QUIET DOWN AND BECOME FRIENDLY. |
| `ECL1.DAX/0x50` | `0x0B34` | `matched` | `world.town.abuzz-about-akabar` | THE TOWN IS ABUZZ ABOUT AKABAR, WHO FREED HAP FROM HORRID MONSTERS. |
| `ECL1.DAX/0x50` | `0x0BAE` | `matched` | `essembra.outdoor-bar` | YOU ARE IN A N OUTDOOR BAR, OVERLOOKING THE WOODS. WHAT WILL YOU DO? |
| `ECL1.DAX/0x50` | `0x0BBB` | `variable-insert` | — | YOU OVERHEAR TAVERN TALE . |
| `ECL1.DAX/0x50` | `0x0BDE` | `matched` | `world.tavern.what-will-you-drink` | WHAT WILL YOU DRINK? |
| `ECL1.DAX/0x50` | `0x0C52` | `matched` | `world.tavern.drank-the-night-away` | AFTER DRINKING THE NIGHT AWAY, YOU AWAKE IN THE INN. |
| `ECL1.DAX/0x50` | `0x0EB1` | `variable-insert` | — | HOW WILL YOU GET TO ? |
| `ECL1.DAX/0x50` | `0x1161` | `matched` | `shadow-gap.new-inn` | YOU REACH SHADOW GAP. ATOP IT, YOU NOTE A RECENTLY BUILT INN -- A PERFECT PLACE TO REST. |
| `ECL1.DAX/0x50` | `0x11AE` | `matched` | `world.wilderness.night-siege` | HOWEVER, ALL IS NOT WELL. INDISTINCT SHAPES CIRCLE THE BUILDING. A BATTLE IS INEVITABLE IF YOU ARE GOING TO GET SOME SLEEP. |
| `ECL1.DAX/0x50` | `0x1229` | `matched` | `world.wilderness.displacer-collars` | THE DISPLACER BEASTS ARE WEARING COLLARS. NEARBY IS A DARK ELF WHO HAS BEEN SLAIN BY THE BEASTS. |
| `ECL1.DAX/0x50` | `0x127F` | `matched` | `world.inn-dark-elf-letter` | THE INNKEEPER TELLS YOU THAT THE ELF ARRIVED SOME DAYS AGO. IN HIS ROOM IS A LETTER, AND YOU RECORD IT IN JOURNAL ENTRY 58. ⚠ 另有 `GOSUB 0x12C5` 印出的一段 |
| `ECL1.DAX/0x50` | `0x12E8` | `matched` | `ashabenford.tilvers-gap` | THE MOUNTAINS RISE INTO AN IMPASSABLE WALL, BROKEN ONLY BY TILVER'S GAP. FLYING SHAPES SPIRAL DOWN FROM THE SNOWY PEAKS. |
| `ECL1.DAX/0x50` | `0x1375` | `matched` | `hap.black-dragons` | SAILING ACROSS THE SKY ARE GREAT BLACK SHAPES. SUDDENLY, THEY SWEEP DOWN, REVEALED AS FEARSOME BLACK DRAGONS. |
| `ECL1.DAX/0x50` | `0x13FF` | `matched` | `post-wizard.dracolich` | OUT OF A COPSE OF TREES COMES A SKELETAL SHAPE. THE HAND OF FEAR GRIPS YOUR HEART AS YOU RECOGNIZE THE VOICE. 'YOU HAVE DEPRIVED ME OF MY TUTOR. SSSTILL, I CAN … |
| `ECL1.DAX/0x50` | `0x14C7` | `matched` | `world.wilderness.voonlar-stragglers` | AS YOU TRAVEL SOUTH, YOU ARE PASSED BY THE STRAGGLERS OF A DEFEATED ARMY, BEARING THE BANNER OF VOONLAR. |
| `ECL1.DAX/0x50` | `0x153D` | `matched` | `world.wilderness.giant-footprints` | AMIDST THE FIELDS OF DAGGERDALE, YOU FIND HUGE FOOTPRINTS LEADING TOWARD A SMALL FARM. IN THE RAYS OF THE SETTING SUN, YOU SEE THE SILHOUETTES OF GIANTS. WHAT D… |
| `ECL1.DAX/0x50` | `0x15FA` | `matched` | `world.wilderness.farmer-thanks` | THE FARMER'S FAMILY COMES UP AND THANKS YOU HEARTILY. YOU ARE INVITED IN AND ENJOY THE FIRST HOME COOKED MEAL IN MANY DAYS. |
| `ECL1.DAX/0x50` | `0x1689` | `matched` | `world.wilderness.lizardmen-riverbank` | LIZARDMEN RUSH FROM A SWAMPY SECTION OF THE RIVERBANK. |
| `ECL1.DAX/0x50` | `0x16DA` | `matched` | `world.myth-drannor-scouts-camp` | PARALLELING YOU THROUGH THE FOREST IS A GROUP WEARING THE LIVERY OF MYTH DRANNOR. IN THE EVENING, THEY INVITE YOU TO THEIR CAMP. THEY TALK, AND YOU RECORD IT IN… |
| `ECL1.DAX/0x50` | `0x17A6` | `matched` | `world.wilderness.morning-farewell` | IN THE MORNING, THEY WISH YOU LUCK AND SEE YOU OFF. |
| `ECL1.DAX/0x50` | `0x17E3` | `matched` | `world.wilderness.nacacia-rides-by` | A YOUNG WOMAN WITH A PURPLE SASH RACES BY ON HORSEBACK. MOMENTS LATER A MAN WITH A HAMMER RIDES BY. AS HE FADES IN THE DISTANCE, YOU CAN HEAR, ' NACACIA ... WAI… |
| `ECL1.DAX/0x50` | `0x189B` | `matched` | `shadow-gap.fire-knives-patrol` | YOU ARE AMBUSHED BY FIRE KNIVES DISGUISED AS A PATROL. |
| `ECL1.DAX/0x50` | `0x1904` | `matched` | `world.wilderness.centaur-hospitality` | CENTAURS APPROACH AND OFFER YOU HOSPITALITY. WHAT DO YOU DO? |
| `ECL1.DAX/0x50` | `0x1956` | `matched` | `world.centaur-village` | THEY TAKE YOU TO THEIR VILLAGE AND EXCHANGE TALES WITH YOU, AND YOU RECORD IT IN JOURNAL ENTRY 45. THEN, YOU PART COMPANY. ⚠ 另有 `GOSUB 0x1989` 印出的一段 |
| `ECL1.DAX/0x50` | `0x19AD` | `matched` | `world.wilderness.regrets-accepted` | THEY ACCEPT YOUR REGRETS AND LEAVE. |
| `ECL1.DAX/0x50` | `0x1A30` | `variable-insert` | — | ONE MORNING, THE PARTY SPOTS A NOTE PINNED TO 'S CHEST. YOU READ IT, AND YOU RECORD IT IN JOURNAL ENTRY . ⚠ 另有 `GOSUB 0x1A6E` 印出的一段 |
| `ECL1.DAX/0x50` | `0x1AAA` | `matched` | `world.bridge-riddle.intro` | YOUR WAY IS BLOCKED BY AN IMPASSABLE CHASM. A NARROW BRIDGE IS GUARDED BY AN OLD MAN. HE CACKLES, ' YOU MUST ANSWER ME BEFORE THE OTHER SIDE YE SEE.' |
| `ECL1.DAX/0x50` | `0x1B27` | `matched` | `world.bridge-riddle.quest` | WHAT IS YOUR QUEST? |
| `ECL1.DAX/0x50` | `0x1B3F` | `matched` | `world.bridge-riddle.fruit` | WHAT IS YOUR FAVORITE FRUIT? |
| `ECL1.DAX/0x50` | `0x1B5D` | `matched` | `world.bridge-riddle.pass` | WHAT DOES THIS MEAN? |
| `ECL1.DAX/0x50` | `0x1B73` | `matched` | `world.bridge-riddle.pass` | YOU MAY PASS. |
| `ECL1.DAX/0x51` | `0x0072` | `matched` | `zhentil.depart-invasion` | AS YOU DEPART, A HUGE FORCE OF ZHENTIL KEEP TROOPS INVADES THE CITY FROM THE NORTH. |
| `ECL1.DAX/0x51` | `0x00C5` | `matched` | `zhentil.depart-civil-war` | AS YOU LEAVE, YOU SEE THAT CIVIL WAR IS ERUPTING. MANY FACTIONS ARE USING THE HAVOC YOU'VE CAUSED TO EVEN OLD SCORES. THE CITY WILL BE UNSAFE FOR A WHILE. |
| `ECL1.DAX/0x51` | `0x018B` | `variable-insert` | — | YOU ARE AT THE EDGE OF . WILL YOU ENTER OR CONTINUE YOUR JOURNEY? |
| `ECL1.DAX/0x51` | `0x02D5` | `matched` | `world.city-gate.combat-too-heavy` | THE COMBAT IS TOO HEAVY TO ENTER. |
| `ECL1.DAX/0x51` | `0x0305` | `matched` | `pit.enter` | YOU ENTER THE PIT. |
| `ECL1.DAX/0x51` | `0x032F` | `matched` | `world.city-gate.city-sealed` | THE GUARDS INFORM YOU THAT THE CITY IS SEALED, PENDING A REORGANIZATION OF THE GOVERNMENT. |
| `ECL1.DAX/0x51` | `0x03A6` | `matched` | `world.wilderness.green-robed-followers` | YOU ARE IN . WHAT PLACE WILL YOU VISIT? |
| `ECL1.DAX/0x51` | `0x0597` | `matched` | `yulash.zhentrim-scum` | 'BEGONE ZHENTRIM SCUM. YOU'LL NOT SOIL MY ESTABLISHMENT.' |
| `ECL1.DAX/0x51` | `0x05E2` | `matched` | `hillsfar.bridled-storm-welcome` | 'WELCOME ADVENTURERS TO THE BRIDLED STORM.' NOTING YOUR SIGILS, HE CONTINUES, 'FOR ZHENTRIM GENTRY I WILL PREPARE OUR BEST ROOMS.' |
| `ECL1.DAX/0x51` | `0x081B` | `matched` | `hillsfar.red-plumes-spill-drinks` | SOME RED PLUMES COME OVER AND SPILL YOUR DRINKS. THEY ORDER YOU TO CLEAN UP THE MESS. DO YOU? |
| `ECL1.DAX/0x51` | `0x08B8` | `matched` | `hillsfar.dockside-bar` | YOU ARE IN A BURNT OUT HULK OF A ONCE FINE INN. WHAT WILL YOU DO? |
| `ECL1.DAX/0x51` | `0x08C6` | `variable-insert` | — | YOU OVERHEAR TAVERN TALE . |
| `ECL1.DAX/0x51` | `0x08E9` | `matched` | `world.tavern.what-will-you-drink` | WHAT WILL YOU DRINK? |
| `ECL1.DAX/0x51` | `0x095D` | `matched` | `world.tavern.drank-the-night-away` | AFTER DRINKING THE NIGHT AWAY, YOU AWAKE IN THE INN. |
| `ECL1.DAX/0x51` | `0x098E` | `matched` | `world.tavern.wake-in-alley` | YOU AWAKE WITH A SPLITTING HEADACHE AND A LIGHT PURSE IN AN ALLEY BEHIND THE BAR. |
| `ECL1.DAX/0x51` | `0x0B84` | `variable-insert` | — | HOW WILL YOU GET TO ? |
| `ECL1.DAX/0x51` | `0x0E08` | `matched` | `world.wilderness.voonlar-vanguard` | HALFWAY TO VOONLAR, YOU SPOT THE VANGUARD OF AN ARMY HEADING SOUTH ALONG THE ROAD, BEARING THE BANNER OF VOONLAR. WHAT DO YOU DO? |
| `ECL1.DAX/0x51` | `0x0EA0` | `matched` | `world.wilderness.marching-on-shadowdale` | YOU ARE TAKEN TO THE OFFICERS WHO STATE THAT THEY ARE MARCHING ON SHADOWDALE. NOTING THE SIGILS ON YOUR ARMS, THEY WILL LET YOU PASS. WHAT DO YOU DO? |
| `ECL1.DAX/0x51` | `0x0F30` | `matched` | `world.wilderness.penalty-for-spying` | THE FIRST OFFICER YOU MEET STATES FLATLY, 'THE PENALTY FOR SPYING IS DEATH. SLAY THEM.' |
| `ECL1.DAX/0x51` | `0x0F7C` | `matched` | `world.wilderness.green-robed-followers` | THE ARMY CONTINUES ON PAST YOU. |
| `ECL1.DAX/0x51` | `0x0FB1` | `matched` | `world.wilderness.army-routs-north` | THE REST OF THE ARMY, CONFUSED BY THE SUDDEN DEMISE OF ITS LEAD FORCES, ROUTS BACK TO THE NORTH. |
| `ECL1.DAX/0x51` | `0x1024` | `matched` | `world.wilderness.used-path-north` | TRAVELLING THROUGH THE WILDS, YOU COME UPON A HEAVILY USED PATH TO THE NORTH. DO YOU INVESTIGATE IT? |
| `ECL1.DAX/0x51` | `0x1072` | `matched` | `world.wilderness.craggy-peak-patrol` | THE TRAIL LEADS UP INTO A CRAGGY WILDERNESS, FINALLY ENDING BELOW A GIGANTIC PEAK. A SQUAD OF ZHENTIL TROOPERS COMES OUT TO MEET YOU. |
| `ECL1.DAX/0x51` | `0x10E4` | `matched` | `zhentil.mercenary-briefing` | ONE COMES UP AND SHAKES YOUR HAND. 'YOU MUST BE THE MERCENARY GROUP.' HE GOES ON TO EXPLAIN THE SITUATION, AND YOU RECORD IT IN JOURNAL ENTRY 36. WHAT DO YOU DO… ⚠ 另有 `GOSUB 0x113D` 印出的一段 |
| `ECL1.DAX/0x51` | `0x1178` | `matched` | `teshwave.left-in-charge` | THE TROOPERS GO TO ALERT TESHWAVE AND LEAVE YOU IN CHARGE OF THE BUGBEARS AND WORGS. YOUR FORCES ARE IMPATIENT AND WISH TO BEGIN THE MARCH. |
| `ECL1.DAX/0x51` | `0x125E` | `matched` | `teshwave.bugbears-rescue` | BUGBEARS AND WORGS COME TO THE RESCUE OF THE TROOPERS. |
| `ECL1.DAX/0x51` | `0x12F2` | `matched` | `teshwave.reach-dagger-falls` | YOU ESCAPE FROM THE MONSTERS AND REACH DAGGER FALLS BEFORE THEM. THE CITY MOBILIZES AND ROUTS THE DISORGANIZED CREATURES. THE ZHENTRIM RESCUE FORCE IS POLITELY … |
| `ECL1.DAX/0x51` | `0x138E` | `matched` | `teshwave.keys-to-the-city` | AS A REWARD YOU ARE GIVEN THE KEYS TO THE CITY, AND A GREAT FEAST AND PARADE ARE HELD IN YOUR HONOR. |
| `ECL1.DAX/0x51` | `0x13ED` | `matched` | `teshwave.monsters-infight` | ON THE LONG ROAD TO DAGGER FALLS, THE MONSTERS BEGIN TO FIGHT AMONGST THEMSELVES. ONE NIGHT, THEIR LEADERS DECIDE THEY SHOULD COMMAND ALONE. |
| `ECL1.DAX/0x51` | `0x1468` | `matched` | `teshwave.sweep-down-melee` | YOU SWEEP DOWN UPON THE ZHENTRIM FORCES AS THEY PREPARE TO MARCH ON DAGGER FALLS, AND A VAST MELEE ENSUES. |
| `ECL1.DAX/0x51` | `0x14DC` | `matched` | `teshwave.both-routed` | THE FORCES TURN OUT TO BE CLOSELY MATCHED. BOTH THE TROOPERS AND THE MONSTERS ARE ROUTED. TESHWAVE ITSELF ONLY SUFFERED MINOR DAMAGE. THE MONSTERS FLEE INTO THE… |
| `ECL1.DAX/0x51` | `0x156E` | `matched` | `teshwave.army-melts-away` | BEREFT OF EFFECTIVE LEADERSHIP, THE ARMY OF MONSTERS MELTS INTO THE WILDS. |
| `ECL1.DAX/0x51` | `0x15E2` | `matched` | `hillsfar.boat-forced-ashore` | AS YOUR BOAT TRAVELS UPSTREAM DOWNSTREAM , ANOTHER CRAFT COMES ALONG SIDE AND FORCES YOU ASHORE. BOARD AND ATTACK. ⚠ 另有 `GOSUB 0x15F4` 印出的一段 |
| `ECL1.DAX/0x51` | `0x164D` | `matched` | `hillsfar.captain-impressed` | THE CAPTAIN IS IMPRESSED WITH YOUR SKILL AND INVITES YOU TO SAIL WITH HIM ANY TIME. |
| `ECL1.DAX/0x51` | `0x169A` | `matched` | `hillsfar.pirates-flee` | A PIRATE CRAFT APPROACHES UNTIL THE PARTY IS SPOTTED. THEN, THEY PILE ON THE CANVAS AND FLEE. |
| `ECL1.DAX/0x51` | `0x17AC` | `matched` | `hillsfar.abandoned-outpost` | ALONG THE COAST, YOU PASS AN ABANDONED ZHENTRIM OUTPOST AND THE REMAINS OF A BUCCANEER'S BASE. |
| `ECL1.DAX/0x51` | `0x180C` | `matched` | `zhentil.patrol_pass` | YOU ARE CONFRONTED BY A PATROL FROM ZHENTIL KEEP. THE OFFICER STEPS FORWARD. 'YOU ARE LISTED IN OUR REPORTS AS CONDEMNED TO DEATH, SENTENCE TO BE CARRIED OUT IM… |
| `ECL1.DAX/0x51` | `0x1924` | `matched` | `world.wilderness.green-robed-followers` | A GREEN ROBED GROUP FOLLOWS YOU. WHAT DO YOU DO? |
| `ECL1.DAX/0x51` | `0x196B` | `matched` | `world.wilderness.robes-thrown-off` | THEY THROW OFF THEIR ROBES AND ATTACK. |
| `ECL1.DAX/0x51` | `0x19A5` | `matched` | `world.wilderness.chosen-ones-to-yulash` | THEY SAY THEY ARE FOLLOWING THE 'CHOSEN ONES' TO YULASH. |
| `ECL1.DAX/0x51` | `0x19DD` | `matched` | `world.wilderness.yulash-forest-slash` | NEAR YULASH, YOU DISCOVER A GREAT SLASH THROUGH THE FOREST TO THE SOUTH. DO YOU INVESTIGATE IT? |
| `ECL1.DAX/0x51` | `0x1A26` | `matched` | `world.wilderness.filth-lair` | AFTER SEVERAL MILES, IT ENDS IN A HUGE PILE OF DECAYING FILTH WHICH MONSTERS NOW USE AS A LAIR. DO YOU INVESTIGATE IT? |
| `ECL1.DAX/0x51` | `0x1AB7` | `matched` | `yulash.red-plume-patrol` | YOU ARE APPROACHED BY A RED PLUME PATROL. ONE SNARLS, 'YOUR TATOO BETRAYS YOU AS A ZHENTRIM SPY.' |
| `ECL1.DAX/0x51` | `0x1B7A` | `matched` | `hillsfar.fire-knives-ambush` | YOU ARE AMBUSHED BY FIRE KNIVES DISGUISED AS FIGHTERS. |
| `ECL1.DAX/0x52` | `0x002E` | `matched` | `opening.curse-summary` | ON YOUR WAY TO THE TOWN OF TILVERTON YOU ARE AMBUSHED, CAPTURED, AND KNOCKED UNCONSCIOUS. WHEN YOU AWAKE YOUR PARTY HAS BEEN CURSED WITH FIVE AZURE SYMBOLS. |
| `ECL1.DAX/0x52` | `0x00B4` | `matched` | `opening.curse-summary` | THE SYMBOLS ENSNARE YOUR WILL LIKE METAL BONDS. AND WHEN THE BONDS GLOW YOU MUST DO AS THEY COMMAND. |
| `ECL1.DAX/0x52` | `0x0112` | `matched` | `opening.curse-summary` | YOUR ONLY HOPE IS TO SEARCH THE FORGOTTEN REALMS FOR THE MEMBERS OF THE ALLIANCE WHO CREATED THE BONDS AND REGAIN CONTROL OF YOUR OWN DESTINY. |
| `ECL1.DAX/0x52` | `0x0195` | `matched` | `opening.curse-summary` | NOWHERE IN THE REALMS IS COMPLETELY SAFE. EVEN THE MOST PEACEFUL SCENE CAN HIDE A DEADLY FOE. |
| `ECL1.DAX/0x52` | `0x0242` | `matched` | `opening.demo-closing` | WHILE THE RISK OF ADVENTURE IS GREAT, THE REWARDS ARE GREATER. FAME AND FORTUNE COME TO THOSE WHO BRAVE THE WILDS. |
| `ECL1.DAX/0x52` | `0x02A9` | `matched` | `opening.demo-closing` | BUT THE ULTIMATE PRIZE IS NOT GOLD OR POWER, IT IS CONTROL OF YOUR OWN DESTINY. AND TO WIN CONTROL YOU MUST DEFEAT THE ULTIMATE ENEMY. |
| `ECL1.DAX/0x52` | `0x031C` | `matched` | `opening.demo-closing` | YOUR ENEMIES ARE PREPARED. THE FORGOTTEN REALMS AWAIT. GO FORWARD AND DEFEAT THE NEW ALLIANCE, TO BE FREE OF THE CURSE OF THE AZURE BONDS. |
| `ECL2.DAX/0x01` | `0x0050` | `matched` | `opening.new-game-awakening` | YOU AWAKEN IN A SMALL ROOM. LOOKING AROUND, YOU NOTICE THAT ALL YOUR GEAR IS GONE, AS IS YOUR MEMORY OF RECENT EVENTS. |
| `ECL2.DAX/0x01` | `0x00B7` | `matched` | `opening.new-game-marks` | ADDING TO YOUR DISQUIET, YOU NOTICE THAT YOUR SWORD ARM HAS BEEN SOMEHOW IMPRINTED WITH STRANGE PATTERNS. THE REST OF YOUR PARTY ARE IDENTICALLY MARKED. |
| `ECL2.DAX/0x01` | `0x0193` | `matched` | `tilverton.door-locked` | THE DOOR IS LOCKED. YOUR ATTEMPT TO ENTER HAS ATTRACTED A GROUP OF ROYAL GUARDS. |
| `ECL2.DAX/0x01` | `0x024E` | `matched` | `tilverton.patrol-attacks` | A PATROL ARRIVES. ROYAL GUARDS TELL YOU TO MOVE ALONG. |
| `ECL2.DAX/0x01` | `0x0263` | `matched` | `tilverton.royal-guards-move-along` | ROYAL GUARDS TELL YOU TO MOVE ALONG. |
| `ECL2.DAX/0x01` | `0x03AB` | `matched` | `tilverton.shop.good-day-purchase` | 'GOOD DAY TO YOU, GENTLE PERSONS. DO YOU WISH TO MAKE A PURCHASE?' |
| `ECL2.DAX/0x01` | `0x0411` | `matched` | `tilverton.shop.return-soon` | 'THANK YOU. RETURN SOON.' YOU MOVE AWAY. |
| `ECL2.DAX/0x01` | `0x0443` | `matched` | `tilverton.smithy.cormyr-steel` | 'WE HAVE A SELECTION OF THE FINEST CORMYR STEEL. INTERESTED? |
| `ECL2.DAX/0x01` | `0x04A5` | `matched` | `tilverton.smithy.strike-true-farewell` | 'MAY YOU ALWAYS STRIKE TRUE.' YOU MOVE AWAY. |
| `ECL2.DAX/0x01` | `0x04C2` | `matched` | `tilverton.shop.good-day-then` | 'GOOD DAY THEN.' YOU MOVE AWAY. |
| `ECL2.DAX/0x01` | `0x04E8` | `matched` | `tilverton.inn.innkeeper-welcome` | 'WELCOME TO THE FAIR CITY OF TILVERTON,' BEAMS THE INNKEEPER. THEN SHE NOTICES YOUR COLLECTIVE SCOWLS. |
| `ECL2.DAX/0x01` | `0x0540` | `matched` | `journal-trigger.tilverton-inn-31` | 'PLEASE CALM DOWN WHILE I EXPLAIN.' YOU LISTEN, AND YOU RECORD IT IN JOURNAL ENTRY 31. 'PERHAPS THE SAGE WILL HELP. YOU CAN GET WEAPONS FROM THE SHOP ACROSS THE… ⚠ 另有 `GOSUB 0x056A` 印出的一段 |
| `ECL2.DAX/0x01` | `0x05C5` | `matched` | `tilverton.inn.nightmare-man` | ON THE BED IN THIS ROOM IS A DISHEVELED MAN, TOSSING IN THE THROES OF SOME GREAT NIGHTMARE. HE SCREAMS, 'FLAMING GIANT ... BLOOD RED MAGE ... THE GLINTING KNIFE… |
| `ECL2.DAX/0x01` | `0x0669` | `matched` | `tilverton.inn.man-left-alone` | HIS VOICE DROPS TO INCOHERENT MUMBLING. THE INNKEEPER RUSHES UP. 'PLEASE LEAVE THE MAN ALONE. HE WAS FOUND IN THIS STATE JUST BEFORE YOU ARRIVED. HE WAS LYING N… |
| `ECL2.DAX/0x01` | `0x073D` | `matched` | `tilverton.tavern.bartender-black-eye` | YOU NOTE THE BARTENDER HAS A BLACK EYE AND HIS ARM IN A SLING. |
| `ECL2.DAX/0x01` | `0x079E` | `matched` | `tilverton.tavern.bartender-excuse` | HE LOOKS AT YOUR BLUE SIGILS AND BLANCHES, THEN STAMMERS, 'J-JUST TRIPPED OVER THE BAR WHILE CLEANING THE PLACE. THESE THINGS HAPPEN.' HE GIVES A FORCED LAUGH. |
| `ECL2.DAX/0x01` | `0x0823` | `matched` | `tilverton.tavern.whats-your-pleasure` | 'WHAT'S YOUR PLEASURE?' |
| `ECL2.DAX/0x01` | `0x08BD` | `matched` | `tilverton.tavern.peppery-drink` | HE HANDS YOU A GLASS OF CLEAR LIQUID WITH A FEW SMALL SEEDS FLOATING IN IT. IT HAS A PEPPERY SMELL. |
| `ECL2.DAX/0x01` | `0x0930` | `matched` | `tilverton.tavern.spray-the-patron` | THE FIERINESS OF THE DRINK IS TOO MUCH FOR YOU, AND YOU SPRAY IT ALL OVER A FELLOW PATRON. HE, IN TURN, SMILES. |
| `ECL2.DAX/0x01` | `0x0995` | `matched` | `tilverton.tavern.steaming-glass` | HE BRINGS OUT A STEAMING GLASS THAT SMELLS STRONGLY OF ALCOHOL AND AN UNIDENTIFIABLE SCENT. |
| `ECL2.DAX/0x01` | `0x09F1` | `matched` | `tilverton.tavern.unable-to-move` | FOR A FEW MOMENTS AFTER YOU DRINK IT, YOU FIND YOURSELF UNABLE TO MOVE, THOUGH THE EFFECT WAS NOT UNPLEASANT. |
| `ECL2.DAX/0x01` | `0x0A56` | `matched` | `tilverton.tavern.brown-liquid` | HE HANDS YOU A MUG OF OPAQUE BROWN LIQUID. |
| `ECL2.DAX/0x01` | `0x0A9B` | `matched` | `tilverton.tavern.brawl` | YOU GET INTO A BRAWL. |
| `ECL2.DAX/0x01` | `0x0ACA` | `matched` | `tilverton.tavern.tossed-outside` | THE SURVIVORS ARE TOSSED OUTSIDE. |
| `ECL2.DAX/0x01` | `0x0B04` | `matched` | `tilverton.tavern.care-for-a-tip` | 'CARE FOR A TIP?' |
| `ECL2.DAX/0x01` | `0x0B1D` | `variable-insert` | — | HE TELLS TAVERN TALE # |
| `ECL2.DAX/0x01` | `0x0B5B` | `matched` | `tilverton.tavern.special-customer` | 'A SPECIAL CUSTOMER'S ARRIVED. YOU HAVE TO SLIP OUTSIDE FOR A MOMENT.' DO YOU GO? |
| `ECL2.DAX/0x01` | `0x0BAB` | `matched` | `tilverton.tavern.time-to-move-on` | 'DRINK UP GENTLEMEN. IT'S TIME FOR YOU TO MOVE ON.' DO YOU GO? |
| `ECL2.DAX/0x01` | `0x0BF3` | `matched` | `tilverton.tavern.purple-sash-slips-in` | AS YOU BEGIN TO WALK OUT THE DOOR, YOU SEE A YOUNG WOMAN WITH A PURPLE SASH SLIP IN THE SIDE DOOR. A FEW OF THE OTHER PATRONS HANG BACK, AS IF TO MEET HER. |
| `ECL2.DAX/0x01` | `0x0C82` | `matched` | `tilverton.tavern.commotion-outside` | AS YOU CONSIDER YOUR NEXT MOVE, YOU HEAR A COMMOTION AROUND THE SIDE OF THE BUILDING. DO YOU GO TO INVESTIGATE? |
| `ECL2.DAX/0x01` | `0x0D31` | `matched` | `journal-trigger.tavern-knife-17` | THERE IS NOTHING HERE NOW, EXCEPT FOR AN ORNATE KNIFE AND YOU RECORD IT IN JOURNAL ENTRY 17. ⚠ 另有 `GOSUB 0x0D5F` 印出的一段 |
| `ECL2.DAX/0x01` | `0x0D81` | `matched` | `tilverton.tavern.party-passes-out` | AFTER MUCH ROISTERING, THE PARTY PASSES OUT. |
| `ECL2.DAX/0x01` | `0x0DB9` | `matched` | `tilverton.tavern.wake-at-inn` | YOU AWAKEN IN YOUR BEDS AT THE INN. THE INNKEEPER POKES HIS HEAD IN. 'YOU'RE LUCKY THE BARTENDER'S A GOOD MAN. HE DROPPED YOU OFF HERE TO SLEEP IT OFF.' |
| `ECL2.DAX/0x01` | `0x0E49` | `matched` | `tilverton.filani.side-door-only` | 'SORRY, BUT THIS IS THE SIDE DOOR. ENTER THROUGH THE FRONT DOOR.' YOU MOVE AWAY. |
| `ECL2.DAX/0x01` | `0x0E9C` | `matched` | `tilverton.filani.introduces` | 'I AM THE SAGE FILANI. YOU ARE HERE ABOUT THE SIGILS, CORRECT?' |
| `ECL2.DAX/0x01` | `0x0EDB` | `matched` | `tilverton.filani.half-your-funds` | 'THIS IS AN INTERESTING CASE. I'LL DO IT FOR HALF YOUR FUNDS. HOW MUCH DO YOU HAVE?' |
| `ECL2.DAX/0x01` | `0x0F5A` | `matched` | `journal-trigger.filani-38` | SHE TALKS AND YOU RECORD IT IN JOURNAL ENTRY 38. ⚠ 另有 `GOSUB 0x0F64` 印出的一段 |
| `ECL2.DAX/0x01` | `0x0F7C` | `matched` | `tilverton.filani.sages-are-not-fools` | 'DO NOT THINK SAGES ARE FOOLS.' SHE SENDS YOU OUT. YOU MOVE AWAY. |
| `ECL2.DAX/0x01` | `0x0FA9` | `matched` | `tilverton.filani.nothing-to-discuss` | 'THEN WE HAVE NOTHING TO DISCUSS.' YOU MOVE AWAY. |
| `ECL2.DAX/0x01` | `0x0FE4` | `matched` | `tilverton.training-hall.want-to-train` | 'DO YOU WANT TO TRAIN?' |
| `ECL2.DAX/0x01` | `0x1020` | `matched` | `tilverton.training-hall.great-progress` | 'YOU'RE SHOWING GREAT PROGRESS. RETURN AGAIN WHEN YOU ARE READY.' YOU EXIT THE HALL. |
| `ECL2.DAX/0x01` | `0x1091` | `matched` | `tilverton.temple.acolyte-greeting` | AN ACCOLYTE GREETS YOU. 'IF YOU WANT HEALING, PRAY AT THE ALTAR.  THE HIGH PRIEST'S RESIDENCE IS IN THE SOUTH.' HE BOWS AND WALKS AWAY. |
| `ECL2.DAX/0x01` | `0x111F` | `matched` | `tilverton.high-priest-intro` | 'I AM THE HIGH PRIEST. YOU LOOK TROUBLED, MY CHILDREN. DO YOU WISH TO TELL ME YOUR STORY?' |
| `ECL2.DAX/0x01` | `0x117D` | `matched` | `journal-trigger.high-priest-19` | HE LISTENS SYMPATHETICALLY AND CASTS A REMOVE CURSE SPELL AND YOU RECORD IT IN JOURNAL ENTRY 19. ⚠ 另有 `GOSUB 0x11AF` 印出的一段 |
| `ECL2.DAX/0x01` | `0x11C1` | `matched` | `tilverton.temple.go-with-gond` | 'GO WITH GOND.' YOU MOVE AWAY. |
| `ECL2.DAX/0x01` | `0x11FD` | `matched` | `tilverton.temple.service-begins` | AS YOU WANDER THE TEMPLE, PEOPLE BEGIN TO STREAM IN. A PRIEST STANDS BEFORE THE ALTAR AND BEGINS A SERVICE. YOU TURN TO LISTEN. |
| `ECL2.DAX/0x01` | `0x1274` | `matched` | `tilverton.temple.sermon-theme` | THE THRUST OF THE SERMON IS ON EVERYTHING'S PLACE IN THE WORLD, ESPECIALLY TILVERTON AND THE CORMYR DEFENSE FORCE. |
| `ECL2.DAX/0x01` | `0x12D8` | `matched` | `tilverton.temple.occupation-force` | NEAR YOU, SOMEONE COMMENTS, 'DEFENSE, HAH, SHOULD BE OCCUPATION FORCE.' THE SERVICE WINDS DOWN AND PEOPLE FILE OUT. |
| `ECL2.DAX/0x01` | `0x1345` | `matched` | `tilverton.alley.psst-commere` | 'PSST.  COMMERE.' COMES A WHISPER FROM DOWN THE ALLEY. |
| `ECL2.DAX/0x01` | `0x139A` | `matched` | `tilverton.carriage-make-way` | YOU ENCOUNTER A GROUP OF ROYAL GUARDS. THEY TELL YOU THAT THIS WAY IS CLOSED BECAUSE THE ROYAL CARRIAGE IS COMING SOON, AND THEY SEND YOU BACK. YOU MOVE AWAY. |
| `ECL2.DAX/0x01` | `0x142E` | `matched` | `world.move-away` | YOU MOVE AWAY. |
| `ECL2.DAX/0x01` | `0x145E` | `matched` | `tilverton.guards-spot-party` | 'THERE THEY ARE!' |
| `ECL2.DAX/0x01` | `0x14A1` | `matched` | `tilverton.reinforcements-flee` | REINFORCEMENTS ARE COMING. YOU MOVE AWAY. |
| `ECL2.DAX/0x01` | `0x14E7` | `matched` | `tilverton.carriage-bond-compulsion` | YOU HEAR THE KING'S VOICE COMING FROM THE CARRIAGE.  SUDDENLY THE SIGILS ON YOUR ARM GLOW BRIGHTLY. YOU FIND YOURSELF UNABLE TO RESIST A COMPULSION TO ATTACK TH… |
| `ECL2.DAX/0x01` | `0x157A` | `matched` | `tilverton.carriage-false-king` | AS THE CARRIAGE RETREATS, A YOUNG MAN LEANS OUT AND CROAKS, 'DON'T KILL ME. I'M NOT REALLY THE KING.' THEN AS HE SPOTS YOUR BLUE MARKINGS, HE FAINTS BACK, CRYIN… |
| `ECL2.DAX/0x01` | `0x1626` | `matched` | `tilverton.carriage-alarm` | A LOUD BELL STARTS RINGING BEHIND YOU. THE GUARDS RUSH TOWARD YOU WITH SWORDS DRAWN. |
| `ECL2.DAX/0x01` | `0x1695` | `matched` | `tilverton.carriage-abduction` | YOU SEE, BEYOND THE CHARGING GUARDS, TWO RED ROBED MEN JUMP THE CARRIAGE. THEY HAUL OUT A THIN MAN AND DRAG HIM INTO AN ALLEYWAY. |
| `ECL2.DAX/0x01` | `0x1717` | `matched` | `tilverton.carriage-surrender` | ONE CALLS OUT, 'DO YOU SURRENDER?' |
| `ECL2.DAX/0x01` | `0x1761` | `matched` | `tilverton.thief-offers-escape` | A THIEF APPEARS AND OFFERS TO LEAD YOU TO SAFETY. DO YOU ACCEPT? |
| `ECL2.DAX/0x01` | `0x17A4` | `matched` | `tilverton.thief-slips-away` | THE THIEF SLIPS AWAY SILENTLY. |
| `ECL2.DAX/0x01` | `0x17D8` | `matched` | `tilverton.carriage-guild-arrival` | THE MAN LEADS YOU THROUGH HIDDEN PASSAGES, EMERGING IN A DARK UNDERGROUND AREA, THE THIEVES' GUILD. |
| `ECL2.DAX/0x01` | `0x183C` | `matched` | `tilverton.narrow-street` | THE STREET GROWS NARROW AND DIM HERE, WITH TRASH PILED IN THE CORNERS. |
| `ECL2.DAX/0x01` | `0x1881` | `matched` | `tilverton.carriage-jailed` | YOU ARE THROWN IN JAIL. |
| `ECL2.DAX/0x01` | `0x18A9` | `matched` | `tilverton.carriage-thief-rescue` | AFTER A FEW HOURS, ONE WALL SLIDES OPEN AND A THIEF APPEARS. HE HANDS YOU YOUR EQUIPMENT AND SIGNALS YOU TO FOLLOW HIM. |
| `ECL2.DAX/0x01` | `0x199E` | `matched` | `tilverton.patrol-surrender-demand` | A PATROL SPOTS YOU. ONE ASKS YOU TO SURRENDER. DO YOU? |
| `ECL2.DAX/0x01` | `0x19F1` | `matched` | `tilverton.rumor.sewer-scream` | AS YOU ARE PASSING THROUGH THE CROWDS, YOU HEAR, ' I'M CERTAIN IT WAS A DRAGON THAT PASSED OVER LAST NIGHT. GOND HELP US IF THERE'S ANOTHER FLIGHT OF THE DRAGON… |
| `ECL2.DAX/0x01` | `0x1BCD` | `matched` | `tilverton.alley.man-calls` | A MAN YELLS FROM A NEARBY ALLEYWAY, 'OVER HERE, BEFORE THE GUARDS FINISH YOU OFF! |
| `ECL2.DAX/0x01` | `0x1C14` | `matched` | `tilverton.alley.safer` | THE DARK ALLEYWAYS LOOK SAFER THAN THE MAIN STREETS. |
| `ECL2.DAX/0x01` | `0x1C42` | `matched` | `tilverton.alley.not-searched` | THE GUARDS DO NOT APPEAR TO BE SEARCHING THE ALLEYS. |
| `ECL2.DAX/0x01` | `0x1C70` | `matched` | `tilverton.alley.escape-route` | A MAN WAVES AT YOU FROM AN ALLEYWAY. IT LOOKS LIKE AN AVENUE OF ESCAPE. |
| `ECL2.DAX/0x01` | `0x1D58` | `variable-insert` | — | YOU SEE A SIGN OVERHEAD |
| `ECL2.DAX/0x02` | `0x0195` | `matched` | `tilverton.guildmaster-greeting` | BEFORE YOU STANDS A BURLY MAN SURROUNDED BY SEVERAL ALERT BODYGUARDS. HE SEEMS SOMEWHAT NERVOUS. 'YOU LOOK A LITTLE ROCKY, CARE TO REST?' |
| `ECL2.DAX/0x02` | `0x023A` | `matched` | `tilverton.guildmaster-briefing` | 'OKAY, LET'S CONTINUE. THE FIRE KNIVES HAVE THE KING'S DAUGHTER IN THE HIDEOUT, HOPING TO LURE HIM INTO A TRAP. I CANNOT DIRECTLY INTERVENE, BUT I CAN OFFER INF… ⚠ 另有 `GOSUB 0x02C7` 印出的一段 |
| `ECL2.DAX/0x02` | `0x02CB` | `matched` | `tilverton.guild-breach` | SUDDENLY, THE SIDE DOOR EXPLODES INWARD WITH A DEAFENING CRASH. |
| `ECL2.DAX/0x02` | `0x0325` | `matched` | `tilverton.guild-fire-knife-command` | 'TRAITOROUS SCUM!' HISSES A FIRE KNIFE. 'SEIZE THEM ALL,' HE COMMANDS HIS SIZABLE BAND OF THIEVES. |
| `ECL2.DAX/0x02` | `0x0379` | `matched` | `tilverton.guild-poisoned-dagger` | AS HIS MEN SPREADOUT, THE GUILDMASTER HURLS A POISONED DAGGER, WHICH CATCHES THE FIRE KNIFE IN THE THROAT. HIS BODY SLUMPS TO THE FLOOR, TWITCHING VIOLENTLY. |
| `ECL2.DAX/0x02` | `0x03FE` | `matched` | `tilverton.guild-battle-joined` | AS YOU PREPARE TO MEET THE ONSLAUGHT, YOU SEE AN ARROW HIT THE GUILDMASTER IN THE CHEST. THEN, THE BATTLE IS JOINED. |
| `ECL2.DAX/0x02` | `0x04BD` | `matched` | `journal-trigger.guildmaster-map-4` | THE GUILDMASTER GASPS, 'ON BALANCE, I'D RATHER BE IN YULASH,' AND THEN HE DIES. YOU FIND INFORMATION ON HIS BODY AND LOG IT AS JOURNAL ENTRY 4. |
| `ECL2.DAX/0x02` | `0x0552` | `matched` | `tilverton.hideout.treasure-room` | YOU HAVE FOUND THE TREASURE ROOM. |
| `ECL2.DAX/0x02` | `0x0585` | `matched` | `tilverton.hideout.safe-door` | THE DOOR IS STOUT ENOUGH TO HOLD AGAINST AN ATTACK. YOU COULD REST HERE SAFELY. |
| `ECL2.DAX/0x02` | `0x05E8` | `matched` | `tilverton.guild-sewer-traces` | YOU SEE GREEN SLIMY MARKS ON THE FLOOR, MORE DISTINCT NEAR THE DOOR. |
| `ECL2.DAX/0x02` | `0x064A` | `matched` | `tilverton.guild-kennel-intro` | YOUR ENTRY IS GREETED BY HUNGRY SNARLS.  A FIRE KNIFE RELEASES THE PACK. |
| `ECL2.DAX/0x02` | `0x06A6` | `matched` | `tilverton.guild-kennel-aftermath` | THIS ROOM IS LITTERED WITH GNAWED BONES AND LEASHES. |
| `ECL2.DAX/0x02` | `0x06F1` | `matched` | `tilverton.guild-monkey-cages` | YOU SEE A NUMBER OF CAGES THAT ONCE HELD MONKEYS. |
| `ECL2.DAX/0x02` | `0x078D` | `matched` | `tilverton.fire-knives-spot-you` | A PARTY OF FIRE KNIVES SPOTS YOU. |
| `ECL2.DAX/0x02` | `0x07AE` | `matched` | `tilverton.hideout.they-charge` | THEY CHARGE. |
| `ECL2.DAX/0x02` | `0x07EA` | `matched` | `tilverton.guild-assassins-attack` | ASSASSINS LEAP ON YOU. |
| `ECL2.DAX/0x02` | `0x081C` | `matched` | `tilverton.hideout.howling-ahead` | YOU HEAR HOWLING FROM AHEAD. |
| `ECL2.DAX/0x02` | `0x0839` | `matched` | `tilverton.hideout.dogs-released` | THE FIRE KNIVES RELEASE THEIR DOGS. |
| `ECL2.DAX/0x02` | `0x0881` | `matched` | `tilverton.hideout.stone-shatters` | A STONE SHATTERS AGAINST THE WALL NEAR YOUR HEAD. |
| `ECL2.DAX/0x02` | `0x08AE` | `matched` | `tilverton.hideout.monkeys-attack` | MONKEYS AND FIRE KNIVES ATTACK. |
| `ECL2.DAX/0x02` | `0x08F4` | `matched` | `tilverton.hideout.menagerie` | YOU ARE ATTACKED BY A MENAGERIE. |
| `ECL2.DAX/0x02` | `0x0963` | `matched` | `tilverton.running-thieves` | YOU COME UPON SOME RUNNING THIEVES. WHAT DO YOU DO? |
| `ECL2.DAX/0x02` | `0x09BC` | `matched` | `tilverton.hideout.thieves-leap` | THE THIEVES LEAP INTO THE FRAY. |
| `ECL2.DAX/0x02` | `0x09E1` | `matched` | `tilverton.running-thieves-warning` | THEY YELL, AS THEY RUN PAST, 'THE FIRE KNIVES ARE PUSHING UP FROM THE SOUTH. THEY'RE BOILING OUT OF THE SEWERS.' |
| `ECL2.DAX/0x02` | `0x0A47` | `matched` | `tilverton.guild-metal-and-animals` | THE CLANG OF METAL ON METAL AND THE GROWLS OF ANIMALS LOCKED IN MORTAL COMBAT ECHO THROUGH THE HALLS. |
| `ECL2.DAX/0x02` | `0x0AA2` | `matched` | `tilverton.guild-bodies-after-battle` | BODIES LIE TWISTED ONE ABOUT ANOTHER, LOCKED IN COMBAT UNTIL DEATH. |
| `ECL2.DAX/0x02` | `0x0AE3` | `matched` | `tilverton.hideout.wounded-dogs` | WOUNDED DOGS BACK AWAY FROM YOU, GROWLING. THEIR MASTERS LIE DEAD ON THE FLOOR. |
| `ECL2.DAX/0x02` | `0x0B2D` | `matched` | `tilverton.hideout.green-robed-woman` | LYING IN A POOL OF BLOOD IS A YOUNG WOMAN IN A GREEN ROBE. NEAR HER IS A BROKEN STAFF, SURMOUNTED BY A HAND WITH A MOUTH FOR A PALM, AND A PIECE OF PAPER. |
| `ECL2.DAX/0x02` | `0x0BAF` | `matched` | `tilverton.hideout.watch-the-chosen-ones` | THE PAPER READS, 'KEEP WATCH ON THE CHOSEN ONES.' |
| `ECL2.DAX/0x02` | `0x0BEE` | `matched` | `tilverton.guild-guest-book` | HERE ON A TABLE IS AN OPEN GUEST BOOK. THE LAST ENTRY READS, 'O.RUSKETTLE BARD OF THE REALMS - TOUCH THE HARP AND LOSE YOUR HAND.' |
| `ECL2.DAX/0x02` | `0x0C75` | `matched` | `tilverton.guild-halfling` | AT THE END OF THE CORRIDOR, YOU SEE A HALFLING WITH A HARP DODGE INTO A DOORWAY AND DISAPPEAR. |
| `ECL2.DAX/0x02` | `0x0CD6` | `matched` | `tilverton.hideout.growl-as-you-escape` | THEY GROWL LOUDLY AS YOU ESCAPE. |
| `ECL2.DAX/0x03` | `0x004D` | `matched` | `tilverton.sewers-entry` | YOU ARE ENTERING THE FOUL SMELLING, SLIME COVERED SEWERS OF TILVERTON. BECAUSE OF THE SLIPPERY FOOTING AND LOW CEILING, IT IS APPARENT THAT FIGHTING WILL BE AWK… |
| `ECL2.DAX/0x03` | `0x01C1` | `matched` | `tilverton.sewers.sealed-exit` | THE WAY IS BLOCKED. A PLACARD PROCLAIMS, 'SEALED BY ORDER OF HIS MAJESTY KING AZOUN IV.' DO YOU WISH EXIT TO THE WILDERNESS? |
| `ECL2.DAX/0x03` | `0x034A` | `matched` | `tilverton.sewers-checkpoint` | YOU ARE AT A CHECKPOINT. THE FIRE KNIVES DEMAND YOUR IMMEDIATE SURRENDER. DO YOU SURRENDER? |
| `ECL2.DAX/0x03` | `0x03FC` | `matched` | `tilverton.sewers-hide-bodies` | YOU QUICKLY HIDE THEIR BODIES. |
| `ECL2.DAX/0x03` | `0x0429` | `matched` | `tilverton.sewers-knight-appears` | HERE LIES THE SLAUGHTERED REMAINS OF A FIRE KNIFE CHECKPOINT. AS YOU CAUTIOUSLY LOOK IT OVER, A MAN STEPS OUT OF THE SHADOWS. HE HOLDS A SWORD AND WEARS THE LIV… |
| `ECL2.DAX/0x03` | `0x04E0` | `matched` | `tilverton.sewers-knight-allegiance` | 'YOU BEAR BLUE TATTOO MARKINGS OF THE FIRE KNIVES, YET I HAVE HEARD RUMORS THAT THE ACCURSED FLAMED ONE WAS USING SUCH THINGS TO CONTROL PEOPLE. TO WHOM DO YOU … |
| `ECL2.DAX/0x03` | `0x05A6` | `matched` | `tilverton.sewers.honorable-man` | I AM IN CHARGE OF HOLDING THIS CORRIDOR AGAINST YOUR KIND. YOU MAY FLEE OR SURRENDER. I'M AN HONORABLE MAN.' WHAT DO YOU DO? |
| `ECL2.DAX/0x03` | `0x0637` | `matched` | `tilverton.sewers.gharri-is-late` | HE ACCEPTS YOUR SURRENDER, AND YOU WAIT. AFTER AN HOUR OF AMIABLE CONVERSATION, HE MUTTERS, 'GHARRI'S LATE -- KNEW HIS RESCUE PLAN WAS STUPID.' THEN HE SAYS TO … |
| `ECL2.DAX/0x03` | `0x06F4` | `matched` | `tilverton.sewers-knight-princess-friend` | HE LAUGHS. 'THAT PRINCESS IS A POPULAR GIRL! WELL, CONTINUE SOUTH AND DON'T KILL THE CLERIC WITH A HAMMER. HE'S TRYING TO RESCUE HER AS WELL.' HE LETS YOU PASS. |
| `ECL2.DAX/0x03` | `0x078D` | `matched` | `tilverton.sewers.come-to-your-death` | 'THEN COME TO YOUR DEATH.' |
| `ECL2.DAX/0x03` | `0x07BA` | `matched` | `tilverton.sewers.dont-kill-the-cleric` | 'FOR A SMALL CITY'S SEWER, THIS PLACE TEAMS WITH ACTIVITY. IF YOU'RE HEADING FOR THE HIDEOUT, DON'T KILL THE CLERIC WITH THE HAMMER. HE'S WITH ME.' HE LETS YOU … |
| `ECL2.DAX/0x03` | `0x085B` | `matched` | `tilverton.sewers.piercing-alarm` | THEY COLLECT MONEY AND EQUIPMENT AND HEAD OFF TO ENJOY THE SPOILS. THEY LAUGH. 'THE MASTER IS AT THE SOUTH END OF THE SEWERS. TELL HIM WE SENT YOU ON.' |
| `ECL2.DAX/0x03` | `0x08FA` | `matched` | `tilverton.sewers.sound-cut-off` | YOU HEAR A SOUND, SUDDENLY CUT OFF, TO THE SOUTH AND WEST. |
| `ECL2.DAX/0x03` | `0x0A0E` | `variable-insert` | — | YOU SPOT THE REMAINS OF A CHECKPOINT TO THE |
| `ECL2.DAX/0x03` | `0x0ADC` | `matched` | `fire-knife.secret-door-found` | YOU FOUND A SECRET DOOR TO THE |
| `ECL2.DAX/0x03` | `0x0B1C` | `matched` | `tilverton.sewers.hidden-training-hall` | YOU ENTER THE HIDDEN CHAMBERS. |
| `ECL2.DAX/0x03` | `0x0B6C` | `matched` | `tilverton.sewers.guild-training-offer` | WELCOME TO THE SECRET TRAINING HALL OF THE GUILD. WOULD YOU CARE TO TRAIN? |
| `ECL2.DAX/0x03` | `0x0BCB` | `matched` | `tilverton.sewers.purple-cloth-and-otyugh` | JUST SEACH AROUND THIS AREA TO FIND US AGAIN. |
| `ECL2.DAX/0x03` | `0x0C0E` | `matched` | `tilverton.sewers.purple-cloth` | YOU SEE A SCRAP OF PURPLE CLOTH CLINGING TO THE BOTTOM OF THE SOUTH DOOR HINGE. |
| `ECL2.DAX/0x03` | `0x0C66` | `matched` | `tilverton.sewers.hand-with-mouth` | BURNT INTO THE WALL HERE IS THE SYMBOL OF A HAND WITH A MOUTH FOR A PALM. THE FAINT STENCH OF DECAYED FLESH SEEMS TO HANG HERE. |
| `ECL2.DAX/0x03` | `0x0CF3` | `matched` | `tilverton.sewers.pack-of-otyugh` | A TERRIBLE STENCH ASSAULTS YOUR SENSES AS YOU ENTER THIS ROOM. FROM OUT OF A PARTICULARLY NASTY PILE COMES A PACK OF OTYUGH. |
| `ECL2.DAX/0x03` | `0x0DB4` | `matched` | `tilverton.sewers.otyugh-pyramids` | PILES OF EXCREMENT HAVE BEEN SHAPED INTO PYRAMIDS HERE. MANY OTYUGH ARE SMOOTHING THE SIDES AND MAKING ARTISTIC EMBELLISHMENTS. A GLINT OF METAL COMES FROM ONE … |
| `ECL2.DAX/0x03` | `0x0E95` | `matched` | `tilverton.sewers.telepathic-bargain` | A TELEPATHIC BUZZ FILLS YOUR MIND. 'IF YOU WISH THE SHINY THING, THEN WE MUST HAVE TREASURE IN RETURN. TO THE SOUTH ARE OTHERS OF OUR ILK, WHO HAVE TWO FINE SME… |
| `ECL2.DAX/0x03` | `0x0F63` | `matched` | `tilverton.sewers.otyugh-toll` | 'WHERE ARE THE TREASURES YOU HAVE PROMISED US? YOU MAY NOT PASS THROUGH THIS ROOM WITHOUT THE FOOD.' |
| `ECL2.DAX/0x03` | `0x0FDA` | `matched` | `tilverton.sewers.otyugh-payment-accepted` | THE OTYUGH BOUNCE ABOUT IN APPARENT JOY OF YOUR GIFTS. THEY PLUCK THE OBJECT FROM THE FETID PYRAMID AND TOSS IT TO YOU, THEN RELIEVE YOU OF YOUR UNPLEASANT BAGG… |
| `ECL2.DAX/0x03` | `0x1089` | `matched` | `tilverton.sewers.zhentrim-jewelry` | THE OBJECT IS A GLITTERING PIECE OF JEWELRY, ORNATELY SCULPTED IN THE SYMBOL OF THE ZHENTRIM. THE OBJECT LOOKS TO HAVE BEEN WORN UNTIL RECENTLY. YOU ARE GIVEN A… |
| `ECL2.DAX/0x03` | `0x1148` | `matched` | `tilverton.sewers.otyugh-friends` | 'PASS THROUGH FREELY, YOU ARE FRIENDS.' |
| `ECL2.DAX/0x03` | `0x118C` | `matched` | `tilverton.sewers.two-piles-attack` | THE ROOM IS FILLED WITH FILTH, THOUGH MOST OF THE SMELL COMES FROM TWO PILES NEAR THE CENTER. THE OTYUGH ATTACK IMMEDIATELY. |
| `ECL2.DAX/0x03` | `0x1225` | `matched` | `tilverton.sewers.carry-the-piles` | IT IS OBVIOUS WHICH PILES THE OTHER OTYUGH WANTED. THOUGH UNPLEASANT, YOU BELIEVE YOU COULD CARRY THEM BACK TO THE OTHER OTYUGHS. DO YOU WANT TO? |
| `ECL2.DAX/0x03` | `0x12CA` | `matched` | `tilverton.sewers.trolls-and-crocodiles` | THE ROOM IS SWAMPY, AND YOU SINK DOWN TO YOUR KNEES. SOME TROLLS ARE SITTING ON A SMALL TUSSOCK, TOSSING HUNKS OF MEAT TO THE WALLOWING CROCODILES. THEY TURN TO… |
| `ECL2.DAX/0x03` | `0x13B2` | `matched` | `tilverton.sewers.bonegrinder-ambush` | AS YOU OPEN THE DOOR, HANDS REACH DOWN FROM ABOVE. THEN COMES A DEEP BASS VOICE. 'WAIT, YOU'RE NOT BONEGRINDER -- BUT YOU'LL PROBABLY TASTE BETTER.' |
| `ECL2.DAX/0x03` | `0x1457` | `matched` | `tilverton.sewers.safe-place-to-rest` | WITH THE MONSTERS DEFEATED, THIS LOOKS LIKE A SAFE PLACE TO REST. |
| `ECL2.DAX/0x03` | `0x149C` | `matched` | `tilverton.sewers.something-on-the-ceiling` | YOU SPOT SOMETHING FLAPPING ON THE CEILING. TO TELL WHAT IT IS, SOMEONE WILL NEED TO CLIMB UP. DOES ANYONE WANT TO?  ONLY A THIEF COULD SUCCEED. |
| `ECL2.DAX/0x03` | `0x1535` | `matched` | `tilverton.sewers.too-unhealthy` | THAT ONE IS TOO UNHEALTHY. |
| `ECL2.DAX/0x03` | `0x1566` | `matched` | `tilverton.sewers.walls-too-slimy` | THE WALLS PROVE TOO SLIMY FOR TO SUCCEED. HE FALLS. |
| `ECL2.DAX/0x03` | `0x15A4` | `matched` | `tilverton.sewers.walls-too-slimy` | DOES ANYONE WANT TO TRY AGAIN? |
| `ECL2.DAX/0x03` | `0x15CB` | `matched` | `tilverton.sewers.swatch-of-cloth` | THE THIEF RETRIEVES A SWATCH OF CLOTH FROM A SEALED TRAPDOOR. THE DOOR WAS TOO WELL JAMMED TO OPEN. THE SCENT OF A TAVERN WAFTS DOWN THROUGH THE DOOR. YOU RECOG… |
| `ECL2.DAX/0x03` | `0x1722` | `matched` | `tilverton.sewers.guild-battle-echoes` | YOU STILL HEAR THE OCCASIONAL SOUNDS OF BATTLE ECHOING FROM THE GUILD HALL. |
| `ECL2.DAX/0x03` | `0x1765` | `matched` | `tilverton.sewers.paper-scrap` | A PIECE OF PAPER POKES ABOVE THE MUCK HERE. YOU RECORD IT IN JOURNAL ENTRY 41. |
| `ECL2.DAX/0x03` | `0x17AA` | `matched` | `tilverton.sewers.tattered-green-robe` | A TATTERED GREEN ROBE LIES TRAMPLED IN THE MUCK. |
| `ECL2.DAX/0x03` | `0x17D5` | `matched` | `tilverton.sewers.shuffling-feet` | YOU HEAR THE SHUFFLING OF LARGE FEET, BUT CAN'T LOCATE THE DIRECTION BECAUSE OF THE ECHOES. |
| `ECL2.DAX/0x03` | `0x1824` | `matched` | `tilverton.sewers.two-red-robed-assassins` | STUFFED IN A CREVICE HERE ARE THE SLAIN BODIES OF TWO RED ROBED ASSASINS. IT WOULD TAKE SOMETHING POWERFUL TO HAVE LODGED THEM SO TIGHTLY. |
| `ECL2.DAX/0x03` | `0x189A` | `matched` | `tilverton.sewers.defective-sigils` | THE REMAINS OF A BODY IS HERE. AN ARM IS MARKED WITH DEFECTIVE VERSIONS OF THE SIGILS ON YOUR ARMS. |
| `ECL2.DAX/0x03` | `0x18EF` | `matched` | `tilverton.sewers.giant-rats-flee` | RATS, THE SIZE OF LARGE DOGS, RUSH AWAY AT YOUR APPROACH. |
| `ECL2.DAX/0x03` | `0x1924` | `matched` | `tilverton.sewers.troll-pieces` | PIECES OF TROLLS LIE STREWN ABOUT HERE. WHAT DO YOU DO? |
| `ECL2.DAX/0x03` | `0x1988` | `matched` | `tilverton.sewers.pieces-coalesce` | THE PIECES EVENTUALLY COALESCE INTO THREE TROLLS. |
| `ECL2.DAX/0x03` | `0x19C6` | `matched` | `tilverton.sewers.sear-the-pieces` | YOU QUICKLY SEAR THE PIECES INTO INACTIVITY. |
| `ECL2.DAX/0x03` | `0x1A09` | `matched` | `tilverton.sewers.assassins-find-camp` | A BAND OF ASSASINS HAS LOCATED YOUR CAMP. |
| `ECL2.DAX/0x03` | `0x1A7C` | `matched` | `tilverton.sewers.hungry-trolls` | YOU MEET HUNGRY TROLLS. |
| `ECL2.DAX/0x03` | `0x1AA6` | `matched` | `tilverton.sewers.otyugh-mistake-food` | SOME OTYUGHS MISTAKE YOU FOR FOOD. |
| `ECL2.DAX/0x03` | `0x1AD8` | `matched` | `tilverton.sewers.trolls-walking-crocodiles` | SOME TROLLS ARE OUT WALKING THEIR CROCODILES. |
| `ECL2.DAX/0x03` | `0x1B19` | `matched` | `tilverton.sewers.large-shapes-corridor` | VERY LARGE SHAPES APPROACH DOWN THE CORIDOR. |
| `ECL2.DAX/0x03` | `0x1C20` | `matched` | `tilverton.sewers.piercing-alarm` | YOU HEAR A PIERCING ALARM. |
| `ECL2.DAX/0x04` | `0x003A` | `matched` | `fire-knife.hideout-entry` | YOU ARE ENTERING THE HIDEOUT. |
| `ECL2.DAX/0x04` | `0x0172` | `matched` | `fire-knife.checkpoint-remains` | YOU SEE THE REMAINS OF A FIRE KNIFE CHECKPOINT. ONLY FIRE KNIVES ARE AMONG THE BODIES. |
| `ECL2.DAX/0x04` | `0x01D2` | `matched` | `fire-knife.checkpoint-abandoned` | THERE ARE SIGNS THAT THIS IS NORMALLY A CHECKPOINT, BUT IT HAS BEEN HURRIEDLY ABANDONED. |
| `ECL2.DAX/0x04` | `0x0235` | `matched` | `tilverton.sewers-checkpoint` | YOU ARE AT A CHECKPOINT. THE FIRE KNIVES DEMAND YOUR IMMEDIATE SURRENDER. DO YOU SURRENDER? |
| `ECL2.DAX/0x04` | `0x02E3` | `matched` | `fire-knife.taken-to-leader` | YOU ARE BEING TAKEN TO THEIR LEADER. |
| `ECL2.DAX/0x04` | `0x030B` | `matched` | `journal-trigger.fire-knives-leader-11` | YOU MEET THE LEADER OF THE FIRE KNIVES. WITH A SINISTER SNEER, HE SPEAKS, AND YOU RECORD IT IN JOURNAL ENTRY 11. ⚠ 另有 `GOSUB 0x0348` 印出的一段 |
| `ECL2.DAX/0x04` | `0x039A` | `matched` | `journal-trigger.fire-knives-victory-54` | NOW THAT THE FIRE KNIVES HAVE BEEN DEFEATED, THE PRINCESS THREATENS THE LEADER, AND YOU RECORD IT IN JOURNAL ENTRY 54. ⚠ 另有 `GOSUB 0x03DC` 印出的一段 |
| `ECL2.DAX/0x04` | `0x03F6` | `matched` | `journal-trigger.fire-knives-royal-arrival-53` | JUST AS YOU SET ABOUT FREEING GIOGI, THE ROOM SHAKES, AND YOU RECORD IT IN JOURNAL ENTRY 53. ⚠ 另有 `GOSUB 0x0425` 印出的一段 |
| `ECL2.DAX/0x04` | `0x043C` | `matched` | `bond-dream.first-night` | ON THE FIRST NIGHT OUTSIDE THE CITY, YOU ARE ALL OVERCOME BY A STRANGE LETHARGY. EVEN THE WATCH DRIFTS OFF. THEN, WITHOUT WARNING, YOU ARE GRIPPED IN A VIVID DR… |
| `ECL2.DAX/0x04` | `0x04CB` | `matched` | `bond-dream.masters-taunt` | FOUR FACES LEER DOWN AT YOU, CONTEMPTUOUS OF YOUR SUCCESS. A DARK AND FOREBODING VOICE INTONES, ' THE FIRST AND WEAKEST OF YOUR MASTERS HAS FALLEN. NOW YOUR FEE… |
| `ECL2.DAX/0x04` | `0x056F` | `matched` | `bond-dream.masters-prophecy` | 'YOU SHALL SERVE EACH OF US IN TURN THOUGH YOUR SPIRIT REBELS, THE WIZARD IN RED, THE WOMAN IN GREEN AND THE LORD OF THE BLACK. FINALLY, YOUR SOULS QUENCHED, YO… |
| `ECL2.DAX/0x04` | `0x0613` | `matched` | `bond-dream.ends` | THE FACES LAUGH WITH EVIL JOY. THE DREAM FADES, AND YOU AWAKE IN A COLD SWEAT. |
| `ECL2.DAX/0x04` | `0x0697` | `variable-insert` | — | YOU SPOT A CHECKPOINT TO THE |
| `ECL2.DAX/0x04` | `0x073F` | `matched` | `fire-knife.secret-door-found` | YOU FOUND A SECRET DOOR TO THE |
| `ECL2.DAX/0x04` | `0x0788` | `matched` | `fire-knife.armory-metal-box` | YOU FOUND THE ARMORY. YOU FIND A LARGE METAL BOX.  DO YOU OPEN IT? |
| `ECL2.DAX/0x04` | `0x0804` | `matched` | `fire-knife.hospital-room` | THE ROOM HAS BEEN CONVERTED TO A HOSPITAL. WOUNDED FIRE KNIVES LIE MOANING IN THE BEDS. THOSE THAT CAN, RUSH OUT THROUGH THE WEST WALL. |
| `ECL2.DAX/0x04` | `0x0878` | `matched` | `fire-knife.wounded-mutterings` | YOU REALIZE THAT THE REST ARE HARMLESS. PASSING THROUGH, YOU HEAR VARIOUS MUTTERINGS, AND YOU RECORD IT IN JOURNAL ENTRY 27. ⚠ 另有 `GOSUB 0x08BE` 印出的一段 |
| `ECL2.DAX/0x04` | `0x08DD` | `matched` | `fire-knife.storage-room` | THE ROOM SEEMS TO BE USED AS A STORAGE AREA. |
| `ECL2.DAX/0x04` | `0x0915` | `matched` | `fire-knife.torture-tables` | AS YOU POKE AROUND, YOU DISCOVER A SET OF TABLES FITTED WITH STRAPS AND SOME INTRICATE TOOLS. THE TIP OF ONE HAS A BLUISH HUE THAT MATCHES YOUR SIGILS. |
| `ECL2.DAX/0x04` | `0x09B8` | `matched` | `fire-knife.torture-chamber` | THIS DARK AND SMOKY ROOM IS ADORNED WITH ALL FORMS OF TORTURE IMPLEMENTS. ON A RACK IN THE MIDDLE OF THE ROOM, IS A WELL MUSCLED, MIDDLE AGED MAN. AROUND HIM AR… |
| `ECL2.DAX/0x04` | `0x0A5C` | `matched` | `fire-knife.assassins-engrossed` | THE ASSASSINS ARE ENGROSSED IN THEIR TASK AND SEEM UNAWARE OF YOU. WHAT DO YOU DO? |
| `ECL2.DAX/0x04` | `0x0AE2` | `matched` | `fire-knife.rescue-nacacia` | YOU RELEASE THE MAN FROM THE RACK, AND HE COLLAPSES TO THE FLOOR. HE HAS SUFFERED VERY BADLY AT THE HANDS OF THE FIRE KNIVES. HE WHISPERS, 'DON'T WORRY ABOUT ME… |
| `ECL2.DAX/0x04` | `0x0B8C` | `matched` | `fire-knife.refuses-assistance` | HE REFUSES ALL OFFERS OF ASSISTANCE, INSISTING THAT YOU SAVE YOUR ENERGY FOR THE RESCUE. ALL HE'LL ACCEPT IS HIS HAMMER THAT WAS LEANING AGAINST THE RACK. HE IN… |
| `ECL2.DAX/0x04` | `0x0C3E` | `matched` | `fire-knife.blade-barrier` | YOU STOP AT THE ENTRANCE TO THIS ROOM. IN FRONT OF YOU IS A CLOUD OF BLADES WHIRLING ABOUT ONE ANOTHER. A METALLIC WHINE MAKES IT ALMOST IMPOSSIBLE TO HEAR. WHA… |
| `ECL2.DAX/0x04` | `0x0CFC` | `matched` | `fire-knife.blade-barrier-damage` | THE BLADES TEAR INTO YOU. |
| `ECL2.DAX/0x04` | `0x0D21` | `matched` | `fire-knife.blade-barrier-fades` | AFTER A FEW MOMENTS THE BLADES SLOW DOWN AND FADE AWAY. THE WHINE DROPS TO A WHISPER AND ENDS. THE ROOM LOOKS AS IF OTHERS HAD BEEN CAUGHT BY THE BLADES. |
| `ECL2.DAX/0x04` | `0x0DB6` | `matched` | `fire-knife.frozen-room` | ABOUT THE ROOM ARE A NUMBER OF PEOPLE FROZEN IN POSITIONS OF BATTLE. A COUPLE HAVE TUMBLED OVER AND LIE IN AWKWARD PILES. A COUPLE OF THEM ARE BEGINNING TO MOVE… |
| `ECL2.DAX/0x04` | `0x0E74` | `matched` | `journal-trigger.frozen-room-26` | YOU DISARM THE FIRE KNIVES AS THEY RETURN TO MOBILITY. THEY SEEM BEWILDERED, AND YOU GATHER SOME USEFUL INFORMATION, AND YOU RECORD IT IN JOURNAL ENTRY 26. ⚠ 另有 `GOSUB 0x0ED5` 印出的一段 |
| `ECL2.DAX/0x04` | `0x0EE3` | `matched` | `fire-knife.frozen-kill` | YOU SLAUGHTER THEM BEFORE THEY RECOVER FROM BEING HELD. |
| `ECL2.DAX/0x04` | `0x0F22` | `matched` | `fire-knife.office` | THIS IS AN ORNATE ROOM, APPARENTLY THE OFFICE OF SOMEONE HIGH UP IN THE FIRE KNIVES. |
| `ECL2.DAX/0x04` | `0x0F9C` | `matched` | `journal-trigger.fire-knives-office-9` | WITHIN THE DRAWERS OF A ROSEWOOD DESK YOU FIND SOME INTERESTING PAPERS, AND YOU RECORD IT IN JOURNAL ENTRY 9. OTHER ITEMS ALSO COME TO YOUR ATTENTION. ⚠ 另有 `GOSUB 0x0FD8` 印出的一段 |
| `ECL2.DAX/0x04` | `0x1028` | `matched` | `fire-knife.smoky-hall` | AS YOU ENTER THIS HALLWAY, YOU DETECT A STRANGE SMOKY SCENT. |
| `ECL2.DAX/0x04` | `0x106D` | `matched` | `fire-knife.ordered-bedroom` | THIS IS AN EXTREMELY WELL ORDERED BEDROOM, EVERYTHING SEEMS EXACTLY IN PLACE. A SEARCH OF THE ROOM TURNS UP NOTHING OF PARTICULAR VALUE. AS YOU LEAVE, UNSEEN SE… |
| `ECL2.DAX/0x04` | `0x112A` | `matched` | `fire-knife.burned-library` | THIS ROOM WAS ONCE A LIBRARY, BUT THE SHELVES AND THEIR CONTENTS ARE NOW ASH. SOME PARTS ARE STILL SMOKING. IN THE CENTER OF THE ROOM IS A CHARRED BODY. CLUTCHE… |
| `ECL2.DAX/0x04` | `0x11CD` | `matched` | `journal-trigger.fire-knives-paper-29` | THE HAND KEPT THE PAPER FROM BEING DESTROYED. YOU TAKE IT, AND YOU RECORD IT IN JOURNAL ENTRY 29. ⚠ 另有 `GOSUB 0x11FF` 印出的一段 |
| `ECL2.DAX/0x04` | `0x121B` | `matched` | `fire-knife.burned-lab` | THIS WAS ONCE A LAB, BUT THE SAME INTENSE FLAME HAS SWEPT THROUGH HERE AS WELL. NOTHING ESCAPED DESTRUCTION. |
| `ECL2.DAX/0x04` | `0x1287` | `matched` | `fire-knife.shrouded-bodies` | WITHIN THE ROOM ARE TWO ROWS OF SHROUDED BODIES. AT THE HEAD OF EACH ROW IS A SIGN. ONE READS 'TO BE RAISED', THE OTHER 'TO BE BURIED'. |
| `ECL2.DAX/0x04` | `0x1360` | `matched` | `fire-knife.hideout-quiet` | THE HIDEOUT IS STRANGELY QUIET FOR SUCH A NORMALLY ACTIVE PLACE. |
| `ECL2.DAX/0x04` | `0x139B` | `matched` | `fire-knife.kybor-incinerated` | FROM SOMEWHERE NEARBY, YOU HEAR A PANICKED VOICE, 'I WENT TO GET KYBOR, TO INTERROGATE THE PRISONER. HIS WHOLE PLACE WAS INCINERATED. THIS PLACE MUST BE CURSED.… |
| `ECL2.DAX/0x04` | `0x1422` | `matched` | `fire-knife.azoun-effigy` | AGAINST THE WALL HERE, YOU SEE A STUFFED FIGURE OF KING AZOUN. IT SHOWS THE MARKS OF INNUMERABLE KNIFE TOSSES. |
| `ECL2.DAX/0x04` | `0x1483` | `matched` | `fire-knife.kill-vangerdahast` | ECHOING THROUGH THE SILENT CORRIDORS COMES AN ANGRY VOICE. 'KYBOR'S DEAD?! WELL, YOU WERE SUPPOSED TO GUARD HIM, SO YOU GET THE HONOR OF KILLING VANGERDAHAST IN… |
| `ECL2.DAX/0x04` | `0x1512` | `matched` | `fire-knife.program-them-for-voice` | THE EERIE SILENCE CARRIES ANOTHER VOICE, 'THAT FOP COULDN'T FOOL ANYONE ON SIGHT. NEXT TIME, WE SHOULD PROGRAM THEM FOR SOMETHING OTHER THAN VOICE.' |
| `ECL2.DAX/0x04` | `0x15AF` | `matched` | `fire-knife.patrol-charges` | YOU ARE SPOTTED BY A FIRE KNIFE PATROL, WHO CHARGE IMMEDIATELY. |
| `ECL2.DAX/0x04` | `0x16DE` | `matched` | `fire-knife.shrill-alarm` | YOU HEAR A SHRILL ALARM. |
| `ECL3.DAX/0x10` | `0x0082` | `matched` | `yulash.entry` | SMOKE RISES FROM BEHIND THE RUINED WALLS OF YULASH. THE SOUND OF BATTLE RINGS OUT FROM INSIDE HOW DO YOU ENTER? |
| `ECL3.DAX/0x10` | `0x032D` | `matched` | `yulash.no-sleeping` | HEY, NO SLEEPING HERE! |
| `ECL3.DAX/0x10` | `0x037C` | `matched` | `yulash.zhentarim-spies` | TROOPS COME BURSTING OUT OF THE COMMANDER'S OFFICE. SOMEONE YELLS, 'STOP THEM! THEY'RE SPIES FOR ZHENTIL KEEP! WHAT DO YOU DO? |
| `ECL3.DAX/0x10` | `0x0470` | `matched` | `yulash.why-let-them-go` | 'WHY'D YOU LET THEM GO?  COWARDLY LOT, AREN'T YOU?  COME WITH ME.' |
| `ECL3.DAX/0x10` | `0x05AE` | `matched` | `lava-tube.dream-warning` | A DREAM-LIKE VOICE IN YOUR HEAD SAYS,'GREAT DANGER LIES BEFORE YOU. BE FULLY PREPARED!' |
| `ECL3.DAX/0x10` | `0x062A` | `matched` | `yulash.checkpoint-deserted` | THE CHECKPOINT IS DESERTED. |
| `ECL3.DAX/0x10` | `0x06BE` | `matched` | `yulash.guard-looks-up` | WHO'S THAT? |
| `ECL3.DAX/0x10` | `0x06DD` | `matched` | `yulash.guard-looks-up` | A RED PLUME GUARD LOOKS UP FROM THE CHECKPOINT. WHAT DO YOU DO? |
| `ECL3.DAX/0x10` | `0x0730` | `matched` | `yulash.passed-the-checkpoint` | YOU HAVE PASSED THE CHECKPOINT. |
| `ECL3.DAX/0x10` | `0x0753` | `matched` | `yulash.destroyed-checkpoint` | YOU'VE COME UPON A DESTROYED CHECKPOINT.  THE MARK OF ZHENTIL KEEP IS SMEARED IN BLOOD ON A WALL. |
| `ECL3.DAX/0x10` | `0x07BA` | `matched` | `yulash.see-commander` | 'YOU MUST COME WITH US TO SEE THE COMMANDER.' WHAT DO YOU DO? |
| `ECL3.DAX/0x10` | `0x0818` | `matched` | `yulash.guards-wave-through` | THE GUARDS WAVE YOU THROUGH. |
| `ECL3.DAX/0x10` | `0x086C` | `matched` | `yulash.red-plume-salute` | A BAND OF RED PLUME GUARDS SALUTE YOU AND PASS ON BY. |
| `ECL3.DAX/0x10` | `0x08E2` | `matched` | `yulash.red-plume-attack` | THE RED PLUME GUARDS ATTACK! |
| `ECL3.DAX/0x10` | `0x0910` | `matched` | `yulash.red-plume-avenge-commander` | RED PLUME GUARDS RUSH AT YOU YELLING, 'THAT'S THEM!  THEY'RE THE SCUM WHO KILLED THE COMMANDER!' |
| `ECL3.DAX/0x10` | `0x0971` | `matched` | `yulash.trying-to-sneak-past` | SO, TRYING TO SNEAK PAST US, HUH? |
| `ECL3.DAX/0x10` | `0x0995` | `matched` | `yulash.checkpoint-halt` | HALT! |
| `ECL3.DAX/0x10` | `0x09AF` | `matched` | `yulash.checkpoint-halt` | A GUARD WARILY COMES OUT OF A CHECKPOINT. OTHER GUARDS GATHER BEHIND HIM. WHAT DO YOU DO? |
| `ECL3.DAX/0x10` | `0x0AA3` | `matched` | `yulash.led-to-commander` | YOU HAVE BEEN LED IN TO SEE THE RED PLUME COMMANDER. |
| `ECL3.DAX/0x10` | `0x0AD1` | `matched` | `yulash.commander-business` | THE COMMANDER DEMANDS TO KNOW YOUR BUSINESS IN YULASH. HOW DO YOU RESPOND? |
| `ECL3.DAX/0x10` | `0x0BEB` | `matched` | `yulash.swanmay-approved` | I SEE YOU HAVE ONE OF THE SWANMAYS WITH YOU. THAT IS GOOD. |
| `ECL3.DAX/0x10` | `0x0C35` | `matched` | `journal-trigger.yulash-commander-22` | YOU HAVE PLEASED THE COMMANDER. YOU RECORD HIS REMARKS AS JOURNAL ENTRY 22. |
| `ECL3.DAX/0x10` | `0x0C90` | `matched` | `yulash.commander-side-door` | THE COMMANDER SHOWS YOU OUT THE SIDE DOOR. |
| `ECL3.DAX/0x10` | `0x0CBB` | `matched` | `yulash.obviously-looters` | YOU ARE OBVIOUSLY LOOTERS.  MY FORCES SHALL DEAL WITH YOU ACCORDINGLY. |
| `ECL3.DAX/0x10` | `0x0CFA` | `matched` | `yulash.red-plumes-attack` | THE RED PLUMES ATTACK. |
| `ECL3.DAX/0x10` | `0x0D3A` | `matched` | `yulash.commander-office-map` | YOU RIFLE THE COMMANDER'S OFFICE AND FIND A MAP. YOU RECORD THE FIND AT JOURNAL ENTRY 52. |
| `ECL3.DAX/0x10` | `0x0E6B` | `matched` | `yulash.shambling-mounds` | SHAMBLING MOUNDS RISE UP FROM THE DEBRIS AROUND YOU. WHAT DO YOU DO? |
| `ECL3.DAX/0x10` | `0x0EDF` | `matched` | `yulash.dead-cleric-bone-wand` | THE DEAD CLERIC CLUTCHES A BONE WAND. |
| `ECL3.DAX/0x10` | `0x0FAB` | `matched` | `yulash.zhentil-marauders` | A BAND OF ZHENTIL KEEP MARAUDERS JUMP YOU. |
| `ECL3.DAX/0x10` | `0x118B` | `matched` | `yulash.looters-are-fire-knives` | THE LOOTERS PULL VERY WICKED LOOKING KNIVES FROM BENEATH THEIR RAGS.  YOU NOTICE THE SYMBOL OF THE FIRE KNIVES EMBLAZONED ON THEIR LEATHER ARMOR. |
| `ECL3.DAX/0x10` | `0x1222` | `matched` | `yulash.people-let-go` | THE PEOPLE CRINGE IN FEAR, CRYING OUT FOR HELP. YOU REALIZE THAT THESE ARE JUST PEOPLE TRYING TO EKE OUT SOME SORT OF EXISTANCE AND LET THEM GO. |
| `ECL3.DAX/0x10` | `0x12F0` | `matched` | `yulash.looters-tell` | THE LOOTERS TELL WHAT THEY KNOW ABOUT THE CITY. YOU RECORD THIS AS JOURNAL ENTRY 34. |
| `ECL3.DAX/0x10` | `0x133A` | `matched` | `yulash.looters-leave` | THE LOOTERS ATTEMPT TO LEAVE. WHAT DO YOU DO? |
| `ECL3.DAX/0x10` | `0x13B3` | `matched` | `yulash.purse-lighter` | AFTER THEY LEAVE, NOTICES THAT HIS PURSE IS CONSIDERABLY LIGHTER. |
| `ECL3.DAX/0x10` | `0x1421` | `matched` | `yulash.red-guards-mess-hall` | THIS IS THE RED GUARDS MESS HALL. |
| `ECL3.DAX/0x10` | `0x1441` | `matched` | `yulash.mess-hall-fee` | IT COSTS 1 PLATINUM PIECE TO REMAIN. DO YOU PAY? |
| `ECL3.DAX/0x10` | `0x147C` | `matched` | `yulash.mess-hall-leave` | THEN YOU MUST LEAVE. |
| `ECL3.DAX/0x10` | `0x14A2` | `variable-insert` | — | AS YOU CONSUME THE LOCAL EXCUSE FOR FOOD AND DRINK, YOU OVERHEAR TAVERN TALE |
| `ECL3.DAX/0x10` | `0x14F9` | `matched` | `yulash.barracks-rest` | THIS IS THE BARRACKS.  THE ROOM IS ABOUT HALF FULL OF RESTING MEN. THEY POINT OUT A SPOT WHERE YOUR PARTY CAN REST. |
| `ECL3.DAX/0x10` | `0x157F` | `matched` | `yulash.captured-cell` | YOU WAKE UP IN A RATHER DREARY ROOM. THE RATS EYE YOU EXPECTANTLY. |
| `ECL3.DAX/0x10` | `0x15E1` | `matched` | `yulash.waiting-room` | YOU ARE PICKED UP AND MARCHED TO SEE THE COMMANDER. |
| `ECL3.DAX/0x10` | `0x1684` | `matched` | `yulash.sinkhole-avoided` | NOTICES A POSSIBLE SINKHOLE IN YOUR PATH AND LEADS YOU AROUND IT. |
| `ECL3.DAX/0x10` | `0x16C4` | `matched` | `yulash.ground-collapses` | THE GROUND SUDDENLY COLLAPSES BENEATH YOU! YOU FALL INTO A PIT FORMED FROM THE DEBRIS OF THE RUINED CITY. |
| `ECL3.DAX/0x10` | `0x17AC` | `matched` | `yulash.weak-walls-avoided` | NOTICES THAT THE WALLS ARE VERY WEAK IN THIS AREA AND LEADS YOU PAST THEM. |
| `ECL3.DAX/0x10` | `0x17F3` | `matched` | `yulash.wall-crumbles` | A WALL SUDDENLY CRUMBLES AND FALLS ON TOP OF YOU! |
| `ECL3.DAX/0x10` | `0x19AE` | `matched` | `yulash.waiting-room` | THIS IS THE COMMANDER'S WAITING ROOM. REMAIN HERE UNTIL YOU ARE CALLED. |
| `ECL3.DAX/0x10` | `0x1AFE` | `matched` | `yulash.riders-burst-out` | JUST BEFORE YOU ENTER A MAN MOUNTED ON A LARGE HORSE BURSTS OUT OF YULASH AND RUNS OVER |
| `ECL3.DAX/0x10` | `0x1B56` | `matched` | `yulash.riders-burst-out` | AS THE HORSE GALLOPS BY, YOU NOTICE A WOMAN DRESSED IN PURPLE CLINGING TO THE LARGE MAN'S BACK. AS THEY SPEED AWAY YOU HEAR HER CALL OUT, 'SORRY!' |
| `ECL3.DAX/0x10` | `0x1C53` | `matched` | `zhentil.fzoul-bond-compels` | THE BOND OF FZOUL ON YOUR ARM SUDDENLY GIVES OFF A BLUE GLOW. YOU FEEL AN UNAVOIDABLE COMPULSION TO DRAW YOUR WEAPONS AND ATTACK. |
| `ECL3.DAX/0x10` | `0x1CCC` | `matched` | `yulash.pit-entrance` | YOU SEE BEFORE THE PIT CREATED BY MOANDER IN HIS LAST INCARNATION.  STEP FORWARD TO ENTER THE DARK DEMESNE. |
| `ECL3.DAX/0x10` | `0x1D2C` | `subroutine` | — | WHAT DO YOU DO? |
| `ECL3.DAX/0x10` | `0x1D3C` | `matched` | `yulash.checkpoint-to-the-east` | YOU SEE A CHECKPOINT MANNED BY RED PLUME GUARDS TO THE EAST. WHAT DO YOU DO? |
| `ECL3.DAX/0x11` | `0x0074` | `matched` | `pit.opening-dead-cultists` | YOU SEE THREE CULTISTS LYING DEAD ON THE FLOOR. JUST AHEAD OF YOU, ANOTHER CLERIC GASPS FOR BREATH. |
| `ECL3.DAX/0x11` | `0x00CA` | `matched` | `pit.opening-chosen` | THE WOUNDED CLERIC'S EYES WIDEN IN FANATIC TRIUMPH. HE HOWLS, 'THE CHOSEN ONES!' |
| `ECL3.DAX/0x11` | `0x0115` | `matched` | `pit.trapped` | THE CLERIC SLAMS HIS FIST AGAINST A PROTRUDING ROCK. THE CEILING BEHIND YOU COLLAPSES. YOU ARE TRAPPED IN THE PIT OF MOANDER. |
| `ECL3.DAX/0x11` | `0x0184` | `matched` | `pit.cleric-dies` | THE CLERIC GIVES YOU ONE LAST TRIUMPHANT GLARE, COUGHS BLOOD AND DIES AT YOUR FEET. |
| `ECL3.DAX/0x11` | `0x01CD` | `matched` | `pit.ambience` | YOU HEAR THE SOUNDS OF BATTLE IN THE DISTANCE. THERE IS A VAGUE SMELL OF BAKED BREAD. |
| `ECL3.DAX/0x11` | `0x0289` | `matched` | `pit.backdoor-entry` | YOU HAVE ENTERED THROUGH THE BACKDOOR OF THE PIT. AN OMINOUS SILENCE PERVADES THE ATMOSPHERE. |
| `ECL3.DAX/0x11` | `0x03B7` | `matched` | `pit.alias-leaves` | WE MUST LEAVE YOU NOW. ALIAS THANKS YOU FOR YOUR HELP.  SHE WISHES YOU LUCK AND LEAVES. |
| `ECL3.DAX/0x11` | `0x0425` | `matched` | `pit.dragonbait-follows` | THE SMELL OF HONEYSUCKLE FILLS THE AIR. DRAGONBAIT QUIETLY LOOKS AT EACH CHARACTER IN TURN, BOWS AND FOLLOWS ALIAS. |
| `ECL3.DAX/0x11` | `0x0492` | `matched` | `pit.dragonbait-carries-alias` | DRAGONBAIT PICKS UP THE BODY OF ALIAS, TURNS TO YOU IN SORROW, AND STARTS RUNNING TOWARDS HILLSFAR. |
| `ECL3.DAX/0x11` | `0x0636` | `matched` | `pit.dead-cultist-blood` | THE BODY OF A DEAD CULTIST LIES IN A POOL OF FRESH BLOOD. THE SOUNDS OF BATTLE SEEM TO COME FROM THE DOOR TO THE SOUTH. |
| `ECL3.DAX/0x11` | `0x06A1` | `matched` | `pit.baked-bread-again` | AGAIN, THAT VAGUE SMELL OF BAKED BREAD. |
| `ECL3.DAX/0x11` | `0x06E6` | `matched` | `pit.bleeding-cleric-grins` | A BLEEDING CLERIC CRAWLS OUT OF THE DOOR TO THE NORTH. AS HE TURNS AND SEES YOU, HIS FACE DEFORMS INTO A RICTUS GRIN. |
| `ECL3.DAX/0x11` | `0x074F` | `matched` | `pit.cleric-dies-praising-moander` | THE CLERIC SAYS, 'MOANDER BE PRAISED.  THE SACRIFICE CAN CONTINUE.' HE DIES--HIS DEAD FEATURES CONTINUE TO GRIN IN HORRID TRIUMPH. |
| `ECL3.DAX/0x11` | `0x07CE` | `matched` | `pit.ringing-weapons-north` | YOU HEAR THE RINGING OF WEAPONS TO THE NORTH. |
| `ECL3.DAX/0x11` | `0x0817` | `matched` | `pit.green-ichor` | GREEN ICHOR COVERS THE FLOOR AND WALLS. A SHAMBLING MOUND LIES IN PIECES AT YOUR FEET. YOU HEAR THE SQUEAL OF VEGEPYGMIES AND SHOUTS OF MEN TO THE SOUTH. |
| `ECL3.DAX/0x11` | `0x089B` | `matched` | `pit.bread-smell-stronger-only` | THE BAKED BREAD SMELL IS STRONGER. |
| `ECL3.DAX/0x11` | `0x08DC` | `matched` | `pit.pile-of-dead` | A PILE OF DEAD CLERICS, SHAMBLING MOUNDS AND VEGEPYGMIES CAN BE SEEN THROUGH THE DOOR TO THE WEST. |
| `ECL3.DAX/0x11` | `0x0930` | `matched` | `pit.cleric-crashes-through-door` | AS YOU WATCH, A CLERIC COMES CRASHING THROUGH THE DOOR AND DIES AT YOUR FEET. THE BAKED BREAD SMELL IS OVERPOWERED BY THE SCENT OF TAR. |
| `ECL3.DAX/0x11` | `0x09A7` | `matched` | `pit.deathly-quiet` | A SUDDEN, DEATHLY QUIET DESCENDS OVER THE ENVIRONS. |
| `ECL3.DAX/0x11` | `0x09EF` | `matched` | `pit.alias-dragonbait-meet` | YOU SEE A FEMALE FIGHTER AND A STRANGE-LOOKING LIZARD MAN. THE SHARP SCENTS OF VIOLETS, BRIMSTONE AND HONEYSUCKLE FOLLOW RAPIDLY UPON EACH OTHER. |
| `ECL3.DAX/0x11` | `0x0A6F` | `matched` | `pit.alias-bonded-reaction` | THE FEMALE FIGHTER GASPS, 'THEY'RE BONDED!' WHAT DO YOU DO? |
| `ECL3.DAX/0x11` | `0x0AE0` | `matched` | `pit.fighter-were-on-your-side` | THE FIGHTER SAYS, 'WHAT THE HELL ARE YOU DOING? WE'RE ON YOUR SIDE! LET'S GET AWAY FROM THESE IDIOTS.' |
| `ECL3.DAX/0x11` | `0x0B3B` | `matched` | `pit.slip-past-and-disappear` | THEY QUICKLY SLIP PAST YOU AND DISAPPEAR. |
| `ECL3.DAX/0x11` | `0x0B84` | `matched` | `pit.alias-dragonbait-introduction` | THE FIGHTER INTRODUCES HERSELF AS ALIAS AND HER COMPANION AS DRAGONBAIT. SHE MENTIONS THAT SHE HAD TATTOOS SIMILAR TO YOURS. SHE ASKS YOU TO TELL YOUR STORY. |
| `ECL3.DAX/0x11` | `0x0C73` | `matched` | `journal-trigger.alias-story-3` | SHE TELLS HER STORY.  YOU NOTE THIS AS JOURNAL ENTRY 3. |
| `ECL3.DAX/0x11` | `0x0CA7` | `matched` | `pit.alias-dragonbait-join` | DO YOU WANT THEM TO JOIN YOU? |
| `ECL3.DAX/0x11` | `0x0CCF` | `matched` | `pit.they-leave-roses` | THEY LEAVE.  YOU SMELL ROSES. |
| `ECL3.DAX/0x11` | `0x0CF0` | `matched` | `pit.alias-dragonbait-joined` | ALIAS AND DRAGONBAIT JOIN YOUR PARTY. |
| `ECL3.DAX/0x11` | `0x0D1F` | `matched` | `pit.alias-dragonbait-joined` | ALIAS SAYS SARDONICALLY, 'THERE'S ALSO THE MATTER OF THE TREASURE THAT MOGION, THE HIGH PRIESTESS, KEEPS BEHIND HER ALTAR.' |
| `ECL3.DAX/0x11` | `0x0D8D` | `matched` | `pit.stairs-down` | YOU SEE STAIRS LEADING DOWN TO THE SOUTH. DO YOU WISH TO GO DOWN? |
| `ECL3.DAX/0x11` | `0x0E65` | `matched` | `pit.vegepygmies-attack` | YOU ARE ATTACKED BY VEGEPYGMIES! |
| `ECL3.DAX/0x11` | `0x0EC7` | `matched` | `pit.vegepygmies-retreat` | YOU COME UPON SOME VEGEPYGMIES. THEY START TO BACK AWAY, POINTING TO THE DOOR BEHIND YOU. WHAT DO YOU DO? |
| `ECL3.DAX/0x11` | `0x0FB5` | `matched` | `pit.shambling-mounds-attack` | YOU ARE ATTACKED BY SHAMBLING MOUNDS! |
| `ECL3.DAX/0x11` | `0x1004` | `matched` | `pit.shambling-mounds-push` | SOME SHAMBLING MOUNDS ARE ATTEMPTING TO PUSH YOU BACK INTO THE CORRIDOR. WHAT DO YOU DO? |
| `ECL3.DAX/0x11` | `0x10BA` | `matched` | `pit.giant-slugs` | GIANT SLUGS ARE CROSSING YOUR PATH. WHAT DO YOU DO? |
| `ECL3.DAX/0x11` | `0x10FA` | `matched` | `pit.giant-slugs-attack` | GIANT SLUGS ATTACK! |
| `ECL3.DAX/0x11` | `0x1161` | `matched` | `pit.fanatics-attack` | MOANDER FANATICS ATTACK! |
| `ECL3.DAX/0x11` | `0x1272` | `matched` | `pit.exit-last-stand` | YOU ARE ATTACKED BY A LARGE FORCE OF CULTISTS IN A LAST-DITCH EFFORT TO STOP YOU. |
| `ECL3.DAX/0x11` | `0x12B9` | `matched` | `pit.kahee-lee` | KAHEE-LEEEEE!!!!! |
| `ECL3.DAX/0x11` | `0x135A` | `matched` | `pit.lemon-tang` | A SLIGHT TANG OF LEMON IS IN THE AIR. ALIAS SAYS, 'DRAGONBAIT SMELLS FRESH AIR.  THERE MUST BE A WAY OUT AROUND HERE.' |
| `ECL3.DAX/0x11` | `0x13DB` | `matched` | `pit.strong-lemon` | A STRONG LEMON SCENT PERMEATS THE AIR. ALIAS SAYS, 'THIS IS WHERE DRAGONBAIT SMELLED THE FRESH AIR.' |
| `ECL3.DAX/0x11` | `0x1481` | `matched` | `pit.brimstone-smell` | YOU SMELL BRIMSTONE. DRAGONBAIT LOOKS DISTURBED ABOUT SOMETHING. |
| `ECL3.DAX/0x11` | `0x14C4` | `matched` | `pit.alias-crawling-with-monsters` | ALIAS NODS. SHE SAYS, 'I DON'T LIKE THIS EITHER. THIS PLACE WAS CRAWLING WITH MONSTERS BEFORE YOU GUYS SHOWED UP.' |
| `ECL3.DAX/0x11` | `0x159B` | `matched` | `pit.altar-on-lower-level` | ALIAS SAYS, 'I THINK THE ALTAR MUST BE ON THE LOWER LEVEL.' |
| `ECL3.DAX/0x11` | `0x1617` | `matched` | `pit.alias-rotting-heap` | ALIAS MUTTERS, 'I CAN'T BELIEVE THEY'RE BRINGING BACK THAT ROTTING HEAP OF GARBAGE.  WHO COULD WORSHIP A GOD LIKE THAT?' |
| `ECL3.DAX/0x11` | `0x1683` | `matched` | `pit.monsters-flee` | THE MONSTERS FLEE! |
| `ECL3.DAX/0x11` | `0x16F0` | `matched` | `pit.mogion-image-summons` | AN IMAGE APPEARS BEFORE YOU. SHE SAYS, 'OH NO, MY PETS.  YOU MUST NOT LEAVE NOW.  YOU ARE THE GUESTS OF HONOR AT OUR SACRED RITES.  COME TO ME IN THE LOWER LEVE… |
| `ECL3.DAX/0x11` | `0x177C` | `matched` | `pit.pushed-back` | A STRONG FORCE PUSHES YOU BACK INTO THE ROOM. |
| `ECL3.DAX/0x12` | `0x012A` | `matched` | `pit.dead-zhentrim` | YOU SEE THE MANGLED REMAINS OF A DEAD ZHENTRIM FIGHTER. WHAT DO YOU DO? |
| `ECL3.DAX/0x12` | `0x0183` | `matched` | `pit.zhentrim-scroll` | GRASPED IN THE FIGHTER'S FIST IS AN OFFICIAL LOOKING SCROLL WITH THE SEAL OF ZHENTIL. YOU RECORD IT'S CONTENTS AS JOURNAL ENTRY 46. |
| `ECL3.DAX/0x12` | `0x027A` | `matched` | `pit.stairs-up` | YOU SEE STAIRS GOING UP IN THE NORTH WALL. DO YOU WISH TO GO UP? |
| `ECL3.DAX/0x12` | `0x0350` | `matched` | `pit.vegepygmies-attack` | YOU ARE ATTACKED BY VEGEPYGMIES! |
| `ECL3.DAX/0x12` | `0x04E3` | `matched` | `pit.shambling-mounds-attack` | YOU ARE ATTACKED BY SHAMBLING MOUNDS! |
| `ECL3.DAX/0x12` | `0x05E2` | `matched` | `pit.mounds-and-slugs-attack` | SHAMBLING MOUNDS AND SLUGS ATTACK! |
| `ECL3.DAX/0x12` | `0x0637` | `matched` | `pit.giant-slugs-appear` | GIANT SLUGS APPEAR. WHAT DO YOU DO? |
| `ECL3.DAX/0x12` | `0x066B` | `matched` | `pit.giant-slugs-attack` | GIANT SLUGS ATTACK! |
| `ECL3.DAX/0x12` | `0x06D2` | `matched` | `pit.fanatics-attack` | MOANDER FANATICS ATTACK! |
| `ECL3.DAX/0x12` | `0x083D` | `matched` | `pit.alias-grabs-hand` | ALIAS GRABS 'S HAND. WHERE DID YOU GET THIS TATOO? WHAT DO YOU DO? |
| `ECL3.DAX/0x12` | `0x0896` | `matched` | `pit.dark-elf-took-the-name` | SO, THE DARK ELF HAS TAKEN THE NAME OF THE ORGANIZATION I SUPPOSEDLY BELONGED TO. THAT'S SOMETHING I'LL HAVE TO INVESTIGATE. |
| `ECL3.DAX/0x12` | `0x0922` | `matched` | `pit.exchange-glances` | ALIAS AND DRAGONBAIT EXCHANGE GLANCES. THE SCENT OF BAKED BREAD BRIEFLY CROSSES YOUR NOSTRILS. |
| `ECL3.DAX/0x12` | `0x0974` | `matched` | `pit.you-lie` | YOU LIE! YOU MUST TELL ME THE TRUTH! WHAT DO YOU DO? |
| `ECL3.DAX/0x12` | `0x09B6` | `matched` | `pit.cant-stay-here` | WE CAN'T STAY HERE. GOODBYE. |
| `ECL3.DAX/0x12` | `0x0A49` | `matched` | `pit.peasant-fed-to-slug` | YOU SEE A LARGE GROUP OF CULTISTS FEEDING A PEASANT TO A GIANT SLUG. ALIAS BLANCHES, THE SMELL OF BAKED BREAD PERMEATES THE AIR. |
| `ECL3.DAX/0x12` | `0x0ADF` | `matched` | `pit.alias-charges` | BEFORE YOU CAN RUN, ALIAS AND DRAGONBAIT CHARGE. |
| `ECL3.DAX/0x12` | `0x0B63` | `matched` | `pit.altar-entrance` | ALIAS SAYS, 'THIS IS IT, THE ENTRANCE TO THE ALTAR OF MOGION.' YOU SMELL VIOLETS. |
| `ECL3.DAX/0x12` | `0x0BE1` | `matched` | `pit.attacked-by-cultists` | YOU ARE ATTACKED BY CULTISTS OF MOANDER. |
| `ECL3.DAX/0x12` | `0x0C7E` | `matched` | `pit.alias-dragonbait-meet` | YOU SEE A FEMALE FIGHTER AND A STRANGE-LOOKING LIZARD MAN. THE SHARP SCENTS OF VIOLETS, BRIMSTONE AND HONEYSUCKLE FOLLOW RAPIDLY UPON EACH OTHER. |
| `ECL3.DAX/0x12` | `0x0CFE` | `matched` | `pit.alias-bonded-reaction` | THE FEMALE FIGHTER GASPS, 'THEY'RE BONDED!' WHAT DO YOU DO? |
| `ECL3.DAX/0x12` | `0x0D65` | `matched` | `pit.fighter-were-on-your-side` | THE FIGHTER SAYS, 'WHAT THE HELL ARE YOU DOING? WE'RE ON YOUR SIDE! LET'S GET AWAY FROM THESE IDIOTS.' |
| `ECL3.DAX/0x12` | `0x0DD9` | `matched` | `pit.alias-introduces` | THE FIGHTER INTRODUCES HERSELF AS ALIAS AND HER COMPANION AS DRAGONBAIT. SHE MENTIONS THAT SHE HAD TATOOS SIMILAR TO YOURS. |
| `ECL3.DAX/0x12` | `0x0E43` | `matched` | `pit.tell-your-story` | SHE ASKS YOU TO TELL YOUR STORY. |
| `ECL3.DAX/0x12` | `0x0EC6` | `matched` | `journal-trigger.alias-story-3` | SHE TELLS HER STORY.  YOU NOTE THIS AS JOURNAL ENTRY 3. |
| `ECL3.DAX/0x12` | `0x0EFA` | `matched` | `pit.alias-dragonbait-join` | DO YOU WANT THEM TO JOIN YOU? |
| `ECL3.DAX/0x12` | `0x0F23` | `matched` | `pit.alias-bursts-out` | ALIAS AND DRAGONBAIT BURST OUT OF THE DOOR TO YOUR EAST. |
| `ECL3.DAX/0x12` | `0x0F54` | `matched` | `pit.alias-offers-help` | ALIAS SAYS, 'LISTEN, MOGION IS ABOUT TO RAISE MOANDER AGAIN.  LET US HELP YOU!' WHAT DO YOU DO? |
| `ECL3.DAX/0x12` | `0x0FB8` | `matched` | `pit.its-your-dead-body` | OK, IT'S YOUR DEAD BODY. ALIAS AND DRAGONBAIT LEAVE. |
| `ECL3.DAX/0x12` | `0x0FF1` | `matched` | `pit.alias-joins-treasure` | ALIAS AND DRAGONBAIT JOIN YOUR PARTY. |
| `ECL3.DAX/0x12` | `0x1025` | `matched` | `pit.alias-joins-journal-3` | AS ALIAS TELLS YOU HER STORY, YOU NOTE IT DOWN AS JOURNAL ENTRY 3. |
| `ECL3.DAX/0x12` | `0x1061` | `matched` | `pit.alias-joins-treasure` | ALIAS SAYS, 'THERE'S ALSO THE MATTER OF THE TREASURE THAT IS KEPT BEHIND THE ALTAR.' |
| `ECL3.DAX/0x12` | `0x10FA` | `matched` | `pit.awake-in-pain` | YOU AWAKE IN A FOG OF PAIN. |
| `ECL3.DAX/0x12` | `0x1121` | `matched` | `pit.mogion-heal-victims` | 'I NEED TO HAVE MY VICTIMS HEALTHY.  SEND THEM TO A ROOM UNTIL THEY HAVE REGAINED THEIR STRENGTH.' |
| `ECL3.DAX/0x12` | `0x11A1` | `matched` | `pit.mogion-altar` | YOU SEE A PRIESTESS TURN AND SMILE WICKEDLY. SHE STANDS BEFORE AN ALTAR. YOU ARE SURROUNDED BY CULTISTS CHANTING IN A LOW DRONE. |
| `ECL3.DAX/0x12` | `0x11E0` | `matched` | `pit.surrounded-by-chanting` | YOU ARE SURROUNDED BY CULTISTS CHANTING IN A LOW DRONE. |
| `ECL3.DAX/0x12` | `0x1235` | `matched` | `pit.alias-identifies-mogion` | ALIAS MUTTERS, 'THAT'S THE PRIESTESS OF MOANDER.' SHE TURNS HER HEAD AND SPITS ON THE GROUND. |
| `ECL3.DAX/0x12` | `0x129D` | `matched` | `pit.mogion-self-identifies` | 'I AM MOGION.' |
| `ECL3.DAX/0x12` | `0x12C2` | `matched` | `pit.mogion-greeting` | MOGION SAYS, 'I AM SO GLAD YOU ARRIVED.  IT IS SO HARD TO DO ANYTHING CONSTRUCTIVE WITHOUT THE PROPER TOOLS. DON'T YOU AGREE?' WHAT DO YOU DO? |
| `ECL3.DAX/0x12` | `0x134D` | `matched` | `pit.bond-paralysis` | BEFORE YOU CAN ACT, A BLUE FLASH COMES FROM THE SIGILS ON YOUR ARMS AND SURROUNDS YOU.  YOU FIND THAT YOU CANNOT MOVE. |
| `ECL3.DAX/0x12` | `0x13C9` | `matched` | `pit.alias-dragonbait-tendrils` | TENDRILS COME UP FROM THE FLOOR AND WRAP THEMSELVES AROUND ALIAS AND DRAGONBAIT. |
| `ECL3.DAX/0x12` | `0x1433` | `matched` | `pit.mogion-ritual` | MOGION TURNS TO THE ALTAR.  THE VOLUME OF CHANTING RISES TO A CRESENDO. |
| `ECL3.DAX/0x12` | `0x1473` | `matched` | `pit.dimensional-window` | THE BLUE LIGHT THAT SURROUNDS YOU STARTS TO STREAM TOWARDS MOGION.  ENERGY DRAWN FROM YOUR BONDS FORM A DIMENSIONAL WINDOW ABOVE THE ALTAR. |
| `ECL3.DAX/0x12` | `0x14EC` | `matched` | `pit.moander-returns` | MOGION SHRIEKS, 'MOANDER RETURNS!' YOU SEE A DIGUSTING MASS OF SLIME, MOLD AND REFUSE START TO OOZE FROM THE DIMENSIONAL RIFT. |
| `ECL3.DAX/0x12` | `0x1559` | `matched` | `pit.bond-fades` | AS THE ENERGY IN THE DIMENSIONAL RIFT INCREASES, YOU FEEL THE SIGIL OF MOANDER BEGIN TO BURN. AS THE OPENING WIDENS, YOU NOTICE THE BOND OF MOANDER BEGIN TO FAD… |
| `ECL3.DAX/0x12` | `0x15EE` | `matched` | `pit.bond-broken` | THE SIGIL DISAPPEARS!  THE PARALYSIS THAT GRIPPED YOU IS NOW GONE! |
| `ECL3.DAX/0x12` | `0x1652` | `matched` | `pit.alias-attack-mogion` | ALIAS AND DRAGONBAIT HAVE HACKED THEIR WAY FREE. SHE HISSES, 'ATTACK THEM NOW, UNLESS YOU WISH TO FIGHT A GOD!' WHAT DO YOU DO? |
| `ECL3.DAX/0x12` | `0x1682` | `matched` | `pit.alias-hacked-free` | ALIAS HAS HACKED HER WAY FREE. SHE HISSES, 'ATTACK THEM NOW, UNLESS YOU WISH TO FIGHT A GOD!' WHAT DO YOU DO? |
| `ECL3.DAX/0x12` | `0x16D5` | `matched` | `pit.dragonbait-frees-himself` | DRAGONBAIT FREES HIMSELF AND WAITS EXPECTANTLY FOR YOUR DECISION. WHAT DO YOU DO? |
| `ECL3.DAX/0x12` | `0x1769` | `matched` | `pit.monsters-catch-you` | THE MONSTERS CATCH YOU. |
| `ECL3.DAX/0x12` | `0x178F` | `matched` | `pit.rift-closes` | THE DIMENSIONAL RIFT SNAPS SHUT. |
| `ECL3.DAX/0x12` | `0x17B5` | `matched` | `pit.remnants-scream` | THE THREE PSUEDOPODS OF MOANDER THAT HAD MADE IT ACROSS SUDDENLY SPROUT HUNDREDS OF MOUTHS, ALL SCREAMING, 'YOU HAVE KILLED ME!' |
| `ECL3.DAX/0x12` | `0x1823` | `matched` | `pit.remnants-attack` | THE OOZING MOUNDS TURN AND ATTACK YOU! |
| `ECL3.DAX/0x12` | `0x1857` | `matched` | `pit.remnants-attack` | THE MONSTERS CATCH YOU. |
| `ECL3.DAX/0x12` | `0x187B` | `matched` | `pit.gauntlet` | YOU FIND THE GAUNTLET OF MOANDER IN THE SLIMY REMAINS. |
| `ECL3.DAX/0x12` | `0x18CD` | `matched` | `pit.priest-flees` | A PRIEST BURSTS INTO THE ROOM, LOOKS AROUND IN HORROR, AND RUNS BACK INTO THE CORRIDORS YELLING, ' THEY HAVE KILLED THE GOD!' |
| `ECL3.DAX/0x12` | `0x195E` | `matched` | `pit.whole-place-after-us` | ALIAS GRUMBLES, 'OH HELL, NOW THE WHOLE PLACE WILL BE AFTER US.' SHE SIGHS AND LOOKS ABOUT IN DISGUST. AFTER A MOMENT, SHE BRIGHTENS AND SAYS, 'THE TREASURE IS … |
| `ECL3.DAX/0x12` | `0x1A20` | `matched` | `pit.altar-treasure` | YOU HAVE FOUND A CACHE OF JEWELS AND GEMS! |
| `ECL3.DAX/0x12` | `0x1A5A` | `matched` | `journal-trigger.pit-temple-map-20` | YOU HAVE ALSO FOUND A MAP OF THE TEMPLE.  YOU RECORD IT AS JOURNAL ENTRY 20. |
| `ECL3.DAX/0x12` | `0x1AFC` | `subroutine` | — | WHAT DO YOU DO? |
| `ECL3.DAX/0x15` | `0x007E` | `matched` | `yulash.war-blasted-section` | YOU FIND A WAR BLASTED SECTION OF THE CITY. |
| `ECL3.DAX/0x15` | `0x00C9` | `matched` | `yulash.unconquered-part` | YOU HAVE FOUND AN UNCONQUERED PART OF THE CITY. |
| `ECL3.DAX/0x15` | `0x01D6` | `matched` | `yulash.false-door` | YOU HAVE FOUND A FALSE DOOR. |
| `ECL3.DAX/0x15` | `0x0212` | `matched` | `yulash.about-to-leave` | YOU ARE ABOUT TO LEAVE. DO YOU WANT TO? |
| `ECL3.DAX/0x15` | `0x0390` | `matched` | `yulash.path-out` | YOU LOCATE A PATH OUT. DO YOU WANT TO EXIT? |
| `ECL3.DAX/0x15` | `0x04F5` | `matched` | `dark-elf-caves.monsters-spotted` | YOU SPOT MONSTERS. |
| `ECL3.DAX/0x15` | `0x05AA` | `matched` | `yulash.graffiti-mogion` | YOU FIND GRAFFITI CRUDELY SCRAWLED, 'MOGION IS NOT MOANDER'S CHOSEN ONE. WE SHALL CALL DOWN HIS GLORY. |
| `ECL3.DAX/0x15` | `0x0603` | `matched` | `yulash.green-robed-bodies` | YOU COME UPON A LARGE PILE OF GREEN ROBED BODIES. STRANGELY, THE FLESH SEEMS TO CHANGE TO PLANT LIFE BEFORE YOUR EYES. EVEN WORSE, THE BODIES ARE NOT DEAD. YOU … |
| `ECL3.DAX/0x15` | `0x06AC` | `matched` | `yulash.plant-life-frenzy` | SUDDENLY, THE PLANT LIFE WHIPS ABOUT IN A FRENZY AND FORMS INTO HIDEOUS SHAPES THAT REACH FOR YOU. |
| `ECL3.DAX/0x15` | `0x0720` | `matched` | `yulash.cultists-call-a-piece` | A LARGE NUMBER OF CULTISTS ARE GATHERED TOGETHER. ONE INTONES, 'OUR BROTHERS AND SISTERS HAVE SHOWN US THE WAY. WE MAY SHED OUR FLESH AND BRING FORTH THE BEAUTI… |
| `ECL3.DAX/0x15` | `0x07CF` | `matched` | `yulash.cultists-form-one-thing` | THE CULTISTS GATHER TOGETHER AND SHRIVEL INTO A HEAP. ONCE AGAIN, FLESH BEGINS TO CHANGE AND FLOW. THIS TIME, ONLY ONE PLANT THING IS FORMING, BUT IT IS VERY LA… |
| `ECL3.DAX/0x15` | `0x08AA` | `matched` | `yulash.red-plume-phlan-plot` | YOU HEAR A PARTY OF RED PLUMES PLANNING TO LOOSE A PACK OF MOANDER CULTISTS ON PHLAN AND THEN BLAME IT ON THE ZHENTRIM. IN THIS WAY, THEY CAN HAVE PHLAN ALLY AG… |
| `ECL3.DAX/0x15` | `0x096F` | `matched` | `yulash.zhentrim-contingent` | YOU COME UPON A CONTINGENT OF ZHENTRIM TROOPS. THEY ARE MUTTERING ABOUT RED PLUMES AND THE DIFFICULTY IN FINDING THEM. WHAT DO YOU DO? |
| `ECL3.DAX/0x15` | `0x0A0B` | `matched` | `yulash.red-plumes-watching-cultists` | YOU SEE SOME RED PLUMES KEEPING AN EYE ON SOME GREEN ROBED CULTISTS. WHAT DO YOU DO? |
| `ECL4.DAX/0x20` | `0x00D0` | `matched` | `zhentil.guards_question` | 'STOP!  GET OVER HERE!  WE HAVE SOME QUESTIONS FOR YOU!' THE GUARDS TAKE YOU ASIDE FOR QUESTIONING. YOU RECORD THIS AS JOURNAL ENTRY 32. |
| `ECL4.DAX/0x20` | `0x0148` | `matched` | `zhentil.guards_warning` | AS YOU LEAVE, YOU OVERHEAR A GUARD SAY, 'JUST AS THE LITTLE THIEF SAID.' THE COMMANDER REPLIES, 'DON'T WORRY.  THEY'LL GET THEIRS INSIDE.' |
| `ECL4.DAX/0x20` | `0x01C2` | `matched` | `zhentil.inner_city` | YOU JOIN THE TEEMING MASSES OF THE INNER CITY. |
| `ECL4.DAX/0x20` | `0x0223` | `matched` | `zhentil.enter-prompt` | DO YOU WISH TO ENTER? |
| `ECL4.DAX/0x20` | `0x032D` | `matched` | `lava-tube.dream-warning` | A DREAM-LIKE VOICE IN YOUR HEAD SAYS, 'GREAT DANGER LIES BEFORE YOU.  BE FULLY PREPARED!' |
| `ECL4.DAX/0x20` | `0x037E` | `matched` | `zhentil.sign.inn` | ZHENTIL INN -- NO RED PLUMES ALLOWED! |
| `ECL4.DAX/0x20` | `0x03A1` | `matched` | `zhentil.sign.crossed-swords` | YOU SEE A PAIR OF CROSSED SWORDS OVER THE DOORWAY. |
| `ECL4.DAX/0x20` | `0x03CE` | `matched` | `zhentil.sign.magic-shop` | ZHENTIL MAGIC SHOP |
| `ECL4.DAX/0x20` | `0x03E4` | `matched` | `zhentil.sign.equipment-shop` | EQUIPMENT SHOP |
| `ECL4.DAX/0x20` | `0x03F6` | `matched` | `zhentil.sign.gorge-and-grog` | THE GORGE AND GROG SHOP ARENA COMBAT BETTING |
| `ECL4.DAX/0x20` | `0x0423` | `matched` | `zhentil.street.troops-take-bodies` | JUST BEFORE YOU PASS OUT FROM PAIN, YOU SEE TROOPS RUSH UP AND FORCE YOUR FOES TO GIVE UP YOUR SEMI-CONSCIOUS BODIES. |
| `ECL4.DAX/0x20` | `0x0590` | `matched` | `zhentil.street.let-you-go` | WE'LL LET YOU GO THIS TIME. |
| `ECL4.DAX/0x20` | `0x05BD` | `matched` | `zhentil.street.take-them-to-magistrate` | ZHENTIL KEEP TROOPERS AND MAGES RUSH YOU YELLING, 'THERE THEY ARE!   ARE THEY THE ONES?' THE HELL WITH THIS.  LET'S TAKE THEM TO THE MAGISTRATE. |
| `ECL4.DAX/0x20` | `0x0747` | `matched` | `zhentil.street.alright-get-out` | ALRIGHT, GET OUT OF HERE. |
| `ECL4.DAX/0x20` | `0x0772` | `matched` | `zhentil.street.fzoul-will-deal` | THOSE ARE THE ONES! LORD FZOUL WILL DEAL WITH THEM! |
| `ECL4.DAX/0x20` | `0x08C6` | `matched` | `zhentil.street.blessings-of-bane` | MAY THE BLESSINGS OF BANE GO WITH YOU. |
| `ECL4.DAX/0x20` | `0x08FB` | `matched` | `dexam.attack` | GET THEM! THEY ATTACKED OUR BROTHERS!. |
| `ECL4.DAX/0x20` | `0x0986` | `matched` | `zhentil.street.halfling-scuttles` | YOU SEE A SMALL HALFLING SCUTTLING ABOUT IN THE DISTANCE. |
| `ECL4.DAX/0x20` | `0x09CD` | `matched` | `zhentil.street.she-ducks-into-crowd` | SHE DUCKS INTO THE CROWD AND DISAPPEARS. |
| `ECL4.DAX/0x20` | `0x0A09` | `matched` | `zhentil.street.dog-piddles` | A DOG PIDDLES ON YOUR LEG. |
| `ECL4.DAX/0x20` | `0x0A55` | `matched` | `zhentil.street.loose-money-pouch` | YOU SEE SOME OFF DUTY ZHENTARIM SOLDIERS WITH BULGING PURSES WALKING DOWN THE STREET. A PICKPOCKET IN YOUR PARTY SEES THAT ONE OF THEM HAS HIS MONEY POUCH VERY … |
| `ECL4.DAX/0x20` | `0x0B3B` | `matched` | `zhentil.street.not-healthy-enough` | THAT CHARACTER IS NOT HEALTHY ENOUGH. |
| `ECL4.DAX/0x20` | `0x0B66` | `matched` | `zhentil.street.drunken-pickpocket` | ACTS LIKE A DRUNK COMING OUT OF THE TAVERN. THE THIEF STUMBLES INTO THE ZHENTRIM AND STARTS APOLOGIZING IN THE MOST PITEOUS, DRUNKEN MANNER. |
| `ECL4.DAX/0x20` | `0x0BF9` | `matched` | `zhentil.street.soldier-cuffs` | THE SOLDIER CUFFS ON THE HEAD, LAUGHS AND GOES ON HIS WAY. GRINS SLYLY IN TRIUMPH. |
| `ECL4.DAX/0x20` | `0x0C61` | `matched` | `zhentil.purse-note` | YOU FIND A NOTE TUCKED INTO A SMALL POCKET OF THE PURSE.  YOU ENTER THIS AS JOURNAL ENTRY 21. |
| `ECL4.DAX/0x20` | `0x0CB6` | `matched` | `zhentil.street.dagger-in-his-side` | AS FEELS THE DAGGER ENTER HIS SIDE, HE REALIZES THAT HE WILL PROBABLY NEED MORE PRACTICE. |
| `ECL4.DAX/0x20` | `0x0D13` | `matched` | `zhentil.street.dagger-in-his-side` | THE REST OF THE GUARDS DRAW THEIR WEAPONS AND ATTACK. YOU DRAW YOUR WEAPONS AND WADE INTO THE FRAY. |
| `ECL4.DAX/0x20` | `0x0D84` | `matched` | `zhentil.street.guards-crash-out` | SOME ZHENTIL KEEP GUARDS COME CRASHING OUT OF THE TAVERN. THEY GLARE AT YOU SUSPICIOUSLY AS YOU PASS BY. |
| `ECL4.DAX/0x20` | `0x0DBA` | `matched` | `zhentil.street.glare-suspiciously` | THEY GLARE AT YOU SUSPICIOUSLY AS YOU PASS BY. |
| `ECL4.DAX/0x20` | `0x0EA1` | `matched` | `zhentil.street.legless-soldier` | YOU SEE A LEGLESS SOLDIER SITTING IN A LITTLE CART. WHEN HE SPOTS YOU, HE POINTS AND STARTS TO YELL, 'YULASH! YOU'RE THE ONES WHO...YOU...YOU... ' |
| `ECL4.DAX/0x20` | `0x0F1D` | `matched` | `zhentil.fritz-accusation` | 'YOU KILLED FRITZ!' HE STARTS GRUNTING INCOHERANTLY AND PUSHING HIMSELF DOWN THE STREET. WHAT DO YOU DO? |
| `ECL4.DAX/0x20` | `0x0F93` | `matched` | `zhentil.fritz-killed` | YOU KILL THE SOLDIER IN COLD BLOOD.  THE CROWD GASPS AND STARTS EDGING AWAY.  PEOPLE ON THE EDGE OF THE CROWD START RUNNING. |
| `ECL4.DAX/0x20` | `0x0FFE` | `matched` | `zhentil.fritz-let-go` | HE CONTINUES DOWN THE STREET. THE CROWDS HAVE THINNED OUT A BIT. |
| `ECL4.DAX/0x20` | `0x1018` | `matched` | `zhentil.street.crowds-thinned` | THE CROWDS HAVE THINNED OUT A BIT. |
| `ECL4.DAX/0x20` | `0x1050` | `matched` | `zhentil.street.body-around-corner` | YOU SEE A GROUP OF PEOPLE DRAG A BODY AROUND THE CORNER. |
| `ECL4.DAX/0x20` | `0x1095` | `matched` | `zhentil.street.fresh-blood` | FRESH BLOOD IS SPLASHED AGAINST THE WALL. |
| `ECL4.DAX/0x20` | `0x1113` | `matched` | `zhentil.street.rotten-food` | SOME ROTTEN FOOD DROPS ON TOP OF YOU FROM AN OPEN WINDOW. |
| `ECL4.DAX/0x20` | `0x11D6` | `matched` | `zhentil.street.city-gone-silent` | THE CITY HAS GONE SILENT.  STREETS ARE DESERTED. AN OMINOUS DREAD SEEPS INTO YOUR BONES. |
| `ECL4.DAX/0x20` | `0x1227` | `matched` | `zhentil.rumor.no-decent-mark` | YOU OVERHEAR, YOU CAN'T FIND A DECENT MARK ANYWHERE. EVERYBODY'S A MAGE OR A PRIEST WITH MAGICAL PROTECTION.  NOW PHLAN, THERE'S A CITY FOR A SLIPPERY HAND. |
| `ECL4.DAX/0x20` | `0x1268` | `matched` | `zhentil.street.ones-she-told-us-about` | 'AREN'T THOSE THE ONES SHE TOLD US ABOUT...?' |
| `ECL4.DAX/0x20` | `0x13D8` | `matched` | `zhentil.street.people-run-indoors` | PEOPLE ARE RUNNING INDOORS AND SHUTTING WINDOWS. |
| `ECL4.DAX/0x20` | `0x1453` | `matched` | `zhentil.street.woman-and-child` | A WOMAN AND HER YOUNG CHILD START TO PASS CLOSE TO YOU. SHE GRABS THE CHILD IN PANIC AND RUSHES OFF. |
| `ECL4.DAX/0x20` | `0x1511` | `matched` | `zhentil.shop.magic-purchase` | YOU WISH TO PURCHASE MAGIC, YES? |
| `ECL4.DAX/0x20` | `0x155F` | `matched` | `zhentil.shop.get-out` | THEN GET OUT! |
| `ECL4.DAX/0x20` | `0x1571` | `matched` | `zhentil.shop.guards-eye-you` | THE GUARDS EYE YOU SUSPICIOUSLY. |
| `ECL4.DAX/0x20` | `0x159D` | `matched` | `zhentil.shop.weapons` | WEAPONS. YOU WANNA BUY? |
| `ECL4.DAX/0x20` | `0x15F4` | `matched` | `zhentil.shop.whachu-want` | WHA'CHU WANT? YOU BUY? |
| `ECL4.DAX/0x20` | `0x1652` | `matched` | `zhentil.inn.rooms-available` | WE GOT ROOMS YA CAN STAY IN.  YA GONNA STAY? |
| `ECL4.DAX/0x20` | `0x167F` | `matched` | `zhentil.inn.rooms-upstairs` | YER ROOMS UPSTAIRS. |
| `ECL4.DAX/0x20` | `0x16C1` | `matched` | `zhentil.court.no-trial` | YOU ARE IN THE COURTHOUSE OF ZHENTIL. NO TRIAL IS CURRENTLY IS SESSION. |
| `ECL4.DAX/0x20` | `0x171E` | `matched` | `zhentil.olive_appears` | A FEMALE HALFLING APPEARS FROM A HIDDEN NICHE. YOU ENTER HER STATEMENT AS JOURNAL ENTRY 50. |
| `ECL4.DAX/0x20` | `0x176C` | `matched` | `zhentil.olive_follow` | DO YOU FOLLOW HER? |
| `ECL4.DAX/0x20` | `0x179B` | `matched` | `zhentil.halfling.the-hard-way` | YOU'LL JUST HAVE TO DO THIS THE HARD WAY THEN! |
| `ECL4.DAX/0x20` | `0x17C9` | `matched` | `zhentil.halfling.coming-this-time` | BACK AGAIN. COMING THIS TIME? |
| `ECL4.DAX/0x20` | `0x17F3` | `matched` | `zhentil.halfling.ducks-through` | THE HALFLING DUCKS THROUGH A SMALL OPENING AND DISAPPEARS. |
| `ECL4.DAX/0x20` | `0x1854` | `matched` | `zhentil.temple.priests-rush-out` | PRIESTS OF BANE RUSH FROM THE TEMPLE! OTHERS WAIT JUST INSIDE THE TEMPLE DOOR. |
| `ECL4.DAX/0x20` | `0x18F6` | `matched` | `zhentil.temple.come-now` | YOU MUST COME NOW! |
| `ECL4.DAX/0x20` | `0x1972` | `matched` | `zhentil.temple.surrender-demand` | A MUCH LARGER GROUP OF REINFORCEMENTS IS RUSHING OUT OF THE TEMPLE YELLING SURRENDER! DO YOU DO SO? |
| `ECL4.DAX/0x20` | `0x19F6` | `matched` | `zhentil.temple.dark-temple` | YOU STAND BEFORE THE DARK TEMPLE OF BANE. |
| `ECL4.DAX/0x20` | `0x1A6A` | `matched` | `zhentil.street.private-home` | YOU HAVE BROKEN INTO THE HOME OF A PRIVATE INDIVIDUAL. THERE IS NOTHING OF INTEREST HERE. THE MAGISTRATE HAS BEEN NOTIFIED OF A PUBLIC DISTURBANCE. |
| `ECL4.DAX/0x20` | `0x1B3A` | `matched` | `zhentil.street.magistrates-office` | HEY YOU! GET OUTA THE MAGISTRATES OFFICE! WHAT DO YOU DO? |
| `ECL4.DAX/0x20` | `0x1B7D` | `matched` | `zhentil.court.guards-attack` | THE GUARDS ATTACK! |
| `ECL4.DAX/0x20` | `0x1BA8` | `matched` | `zhentil.court.first-wave-defeated` | YOU HAVE DEFEATED THE FIRST WAVE OF GUARDS. OTHERS ARE POURING INTO THE ROOM! WHAT DO YOU DO? |
| `ECL4.DAX/0x20` | `0x1C1B` | `matched` | `zhentil.fzoul-bond-compels` | THE BOND OF FZOUL ON YOUR ARM SUDDENLY GIVES OFF A BLUE GLOW. YOU FEEL AN UNAVOIDABLE COMPULSION TO DRAW YOUR WEAPONS AND ATTACK. |
| `ECL4.DAX/0x20` | `0x1CDE` | `subroutine` | — | WHAT DO YOU DO? |
| `ECL4.DAX/0x21` | `0x0080` | `matched` | `zhentil.dimswart_appears` | YOU ARE DRAGGED THROUGH THE TEMPLE. |
| `ECL4.DAX/0x21` | `0x01A7` | `matched` | `zhentil.dimswart_appears` | YOU WAKE UP IN A DARK, DREARY CELL. |
| `ECL4.DAX/0x21` | `0x01D5` | `matched` | `zhentil.temple.sealed-by-priest` | A BLINDING FLASH ENVELOPES THE MAIN DOORS. |
| `ECL4.DAX/0x21` | `0x0207` | `matched` | `zhentil.temple.sealed-by-priest` | YOU SEE A PRIEST IN AN OPEN DOORWAY WITH A CRUMBLING SCROLL IN HIS HANDS. HE YELLS, 'I HAVE SEALED THE TEMPLE.  THE HERETICS ARE TRAPPED!'  HE DUCKS BEHIND A DO… |
| `ECL4.DAX/0x21` | `0x02C9` | `matched` | `zhentil.dark_shrine_entry` | YOU HAVE ENTERED THE DARK SHRINE THROUGH A HOLE IN THE WALL CREATED BY A MAGICAL DEVICE CARRIED BY THE HALFLING.  AS THE WALLS CLOSE BEHIND YOU, YOU SEE THE DEV… |
| `ECL4.DAX/0x21` | `0x0364` | `matched` | `zhentil.olive_explains` | AS OLIVE LEADS YOU THROUGH THE DUNGEON SHE EXPLAINS ABOUT DIMSWART THE SAGE.  YOU NOTE THIS DOWN AS JOURNAL ENTRY 51. |
| `ECL4.DAX/0x21` | `0x03E6` | `matched` | `zhentil.dimswart_door` | DIMSWART IS JUST BEYOND THIS DOOR. |
| `ECL4.DAX/0x21` | `0x0407` | `matched` | `zhentil.olive_leaves` | OLIVE SMILES AND DISAPPEARS. |
| `ECL4.DAX/0x21` | `0x0508` | `matched` | `zhentil.dimswart_appears` | YOU SEE AN OLD MAN IN THE CELL.  HE INTRODUCES HIMSELF AND YOU RECORD HIS REMARKS AS JOURNAL ENTRY 12. |
| `ECL4.DAX/0x21` | `0x0560` | `matched` | `zhentil.dimswart_join` | WILL YOU TAKE DIMSWART ALONG? |
| `ECL4.DAX/0x21` | `0x05C8` | `matched` | `zhentil.temple.attacked-by-priests` | YOU ARE ATTACKED BY THE PRIESTS OF BANE. |
| `ECL4.DAX/0x21` | `0x065C` | `matched` | `zhentil.olive_cell_hint` | OLIVE SUDDENLY APPEARS IN FRONT OF YOU. SHE SAYS, 'SO, YOU'RE FINALLY HERE. THE MAN I TOLD YOU ABOUT IS IN THE CELL TO THE SOUTH.' |
| `ECL4.DAX/0x21` | `0x06CD` | `matched` | `zhentil.olive_repeats_dimswart` | AS SHE EXPLAINS ABOUT DIMSWART, YOU NOTE HER REMARKS AS JOURNAL ENTRY 51 |
| `ECL4.DAX/0x21` | `0x070E` | `matched` | `zhentil.olive_leaves` | OLIVE SMILES AND DISAPPEARS. |
| `ECL4.DAX/0x21` | `0x07A2` | `matched` | `zhentil.temple.secret-door-east` | YOU FIND A SECRET DOOR TO THE WEST. |
| `ECL4.DAX/0x21` | `0x0804` | `matched` | `zhentil.hooded_offer` | A HOODED WOMAN SUDDENLY APPEARS AND SAYS, FOLLOW ME.  I CAN GET YOU OUT OF HERE! YOU ARE TRAPPED HERE! MY MASTER CAN HELP YOU. |
| `ECL4.DAX/0x21` | `0x08AA` | `matched` | `zhentil.temple.hooded-figure-leaves` | UNLIKE YOU, I CANNOT REMAIN HERE. THE HOODED FIGURE DISAPPEARS! ⚠ 另有 `GOSUB 0x08E0` 印出的一段 |
| `ECL4.DAX/0x21` | `0x08F7` | `matched` | `zhentil.temple.traced-the-magic` | A LARGE GROUP BURSTS IN ON YOU. WE HAVE TRACED THE MAGIC TO HERE! |
| `ECL4.DAX/0x21` | `0x0937` | `matched` | `zhentil.fzoul_interrupts` | A STRANGE SHIMMER SURROUNDS YOU. THE ROOM AROUND YOU BEGINS TO FADE. JUST BEFORE IT DOES, A BAND OF CLERICS BURST INTO THE ROOM LED BY FZOUL CHEMBRYL. HE YELLS,… |
| `ECL4.DAX/0x21` | `0x09E5` | `matched` | `zhentil.fzoul_retreats` | AS THE ROOM FADES AWAY, HE TURNS AND RUSHES OUT OF THE ROOM. |
| `ECL4.DAX/0x21` | `0x0B86` | `matched` | `zhentil.temple.sorry-to-disturb` | SORRY TO DISTURB YOU.  BANES BLESSING ON YOU. |
| `ECL4.DAX/0x21` | `0x0BAC` | `matched` | `zhentil.temple.sorry-to-disturb` | THE PRIESTS LEAVE. |
| `ECL4.DAX/0x21` | `0x0BC5` | `matched` | `zhentil.temple.one-moves-away` | AS THE PRIESTS TALK, ONE STARTS TO MOVE AWAY, POSSIBLY TO WARN OTHERS. DO YOU ATTACK? |
| `ECL4.DAX/0x21` | `0x0C18` | `matched` | `zhentil.temple.priests-stalling` | THE PRIESTS KEEP TALKING.  YOU THINK THEY'RE STALLING. WHAT DO YOU DO? |
| `ECL4.DAX/0x21` | `0x0C7D` | `matched` | `zhentil.temple.priests-attack` | REINFORCEMENTS ARRIVE, AND THE PRIESTS ATTACK! |
| `ECL4.DAX/0x21` | `0x0CF0` | `matched` | `zhentil.temple.spotted-by-patrol` | YOU ARE SPOTTED BY A PATROL OF PRIESTS! |
| `ECL4.DAX/0x21` | `0x0D15` | `matched` | `zhentil.temple.priests-attack` | THE PRIESTS ATTACK! |
| `ECL4.DAX/0x21` | `0x0D99` | `matched` | `zhentil.temple.door-below-altar` | YOU NOTICE A SMALL DOOR BELOW THE ALTAR. WHAT DO YOU DO? |
| `ECL4.DAX/0x21` | `0x0DF9` | `matched` | `zhentil.temple.gas-cloud` | A CLOUD OF GAS ENVELOPES THE PARTY! |
| `ECL4.DAX/0x21` | `0x0E3E` | `matched` | `zhentil.temple.no-detect-trap` | NO DETECT TRAP SPELL IN PARTY. |
| `ECL4.DAX/0x21` | `0x0E6D` | `matched` | `zhentil.temple.detects-a-trap` | DETECTS A TRAP! WHAT DO YOU DO? |
| `ECL4.DAX/0x21` | `0x0EFE` | `matched` | `zhentil.temple.you-succeeded` | YOU SUCCEEDED! |
| `ECL4.DAX/0x21` | `0x0F32` | `matched` | `zhentil.temple.mirror-room` | THIS ROOM HAS MANY MIRRORS HANGING FROM THE CEILING IN NO APPARENT ORDER. AS YOU WALK, DIFFERENT MIRRORS ACTIVATE AS YOU PASS. VAGUE SHAPES SEEM TO SHIMMER INSI… |
| `ECL4.DAX/0x21` | `0x0FD9` | `matched` | `zhentil.temple.scrying-room` | DIMSWART LOOKS ABOUT IN AWE. 'THIS IS A SCRYING ROOM.  FZOUL MUST BE SPYING ON ALL THE DIFFERENT CREATURES IN ZHENTIL THAT COULD POTENTIALLY BE USED AGAINST HIM… |
| `ECL4.DAX/0x21` | `0x1472` | `matched` | `zhentil.temple.olive-in-crevice` | YOU SEE OLIVE RUSKETTLE SQUEEZING THROUGH A SMALL CREVICE.  SHE SEEMS TO BE SURROUNDED BY ROCK.  BY THE MOVEMENT OF HER LIPS, IT IS EVIDENT THAT THE AIR IS TURN… |
| `ECL4.DAX/0x21` | `0x147C` | `matched` | `zhentil.fzoul-bond-compels` | THE BOND OF FZOUL ON YOUR ARM SUDDENLY GIVES OFF A BLUE GLOW. YOU FEEL AN UNAVOIDABLE COMPULSION TO DRAW YOUR WEAPONS AND ATTACK. |
| `ECL4.DAX/0x21` | `0x152E` | `matched` | `zhentil.hooded_follow` | DO YOU FOLLOW? |
| `ECL4.DAX/0x22` | `0x0071` | `matched` | `dexam.dimswart-whispers` | YOU WAKE UP IN A CONFUSION OF VOICES. THE OLD MAN INTRODUCES HIMSELF AS DIMSWART THE SAGE, AND YOU RECORD HIS HURRIED WHISPERS AS JOURNAL ENTRY 12. |
| `ECL4.DAX/0x22` | `0x00ED` | `matched` | `dexam.arrival` | YOU ARE LED ALONG A DARK CORRIDOR. |
| `ECL4.DAX/0x22` | `0x0153` | `matched` | `dexam.arrival` | YOU SEE BEFORE YOU, DEXAM THE BEHOLDER! THE HOODED WOMAN WALKS TO A MILLING GROUP OF ARMORED MINOTAURS. THEY IMMEDIATELY SNAP TO ATTENTION. |
| `ECL4.DAX/0x22` | `0x01CA` | `matched` | `dexam.journal_30` | DEXAM SPEAKS.  YOU RECORD HIS SPEECH AS JOURNAL ENTRY 30. |
| `ECL4.DAX/0x22` | `0x01FF` | `matched` | `dexam.amulet_choice` | DIMSWART THE SAGE WHISPERS, 'LOOK ON THE ALTAR. THAT'S THE AMULET OF LATHANDER.' WHAT DO YOU DO? |
| `ECL4.DAX/0x22` | `0x0287` | `matched` | `dexam.fzoul_journal_7` | BEFORE YOU CAN ACT, AN IMPRESSIVE MAN, BACKED BY TROOPS, PRIESTS AND MAGICIANS, RUSHES INTO THE ROOM. HIS STATEMENT IS IN JOURNAL ENTRY 7. |
| `ECL4.DAX/0x22` | `0x02FD` | `matched` | `dexam.kills_fzoul` | DEXAM ROARS, 'THEN DIE, HERETIC!' AND BLASTS FZOUL INTO A HEAP OF ASH.  YOU FEEL THE BONDS ON YOUR ARM START TO WRITHE. |
| `ECL4.DAX/0x22` | `0x036E` | `matched` | `dexam.fzoul_bond_fades` | THE BOND OF FZOUL FADES. |
| `ECL4.DAX/0x22` | `0x038A` | `matched` | `dexam.kill_order` | NOW THAT THE BOND IS GONE, THEY ARE OF NO USE TO ME. KILL THEM. |
| `ECL4.DAX/0x22` | `0x03C4` | `matched` | `dexam.amulet_rises` | DEXAM TURNS TO THE ALTAR AND STARTS TO LEAVE. YOU NOTICE THE AMULET RISING TO MEET HIM. |
| `ECL4.DAX/0x22` | `0x0411` | `matched` | `dexam.altar_melee` | AS DEXAM LEAVES, HIS FORCES AND THE REMAINING TROOPS OF FZOUL START TO MELEE IN THE ROOM AROUND YOU. |
| `ECL4.DAX/0x22` | `0x0561` | `matched` | `dexam.teleport-jerk` | A GUT WRENCHING JERK SLAMS YOUR STOMACH TO YOUR SPINE.  YOU FEEL AS IF YOU WERE PROPELLED FROM A CANNON. |
| `ECL4.DAX/0x22` | `0x05C1` | `matched` | `dexam.walls-blur` | YOU SEE THE WALLS AROUND BLUR TO A GRAY SMEAR. |
| `ECL4.DAX/0x22` | `0x05EB` | `matched` | `dexam.slammed-against-wall` | YOU ARE SUDDENLY SLAMMED AGAINST A WALL. |
| `ECL4.DAX/0x22` | `0x0631` | `matched` | `dexam.dead-elf.remains` | YOU SEE THE REMAINS OF AN ELF FIGHTER. WHAT DO YOU DO? |
| `ECL4.DAX/0x22` | `0x069B` | `matched` | `dexam.dead-elf.pouch` | BURIED BENEATH THE MOLDERING CLOTHING,  YOU FIND A LEATHER POUCH. |
| `ECL4.DAX/0x22` | `0x074B` | `matched` | `dexam.dead-elf.pouch-moves` | THE POUCH MOVES.  YOU HEAR THE CRACKLING OF SOME PAPER-LIKE SUBSTANCE AND THE CLINK OF METAL. |
| `ECL4.DAX/0x22` | `0x07B3` | `matched` | `dexam.spell-unavailable` | YOU DON'T HAVE THAT SPELL AVAILABLE. |
| `ECL4.DAX/0x22` | `0x07DE` | `matched` | `dexam.discovers-a-trap` | DISCOVERS A TRAP! WHAT DO YOU DO? |
| `ECL4.DAX/0x22` | `0x0813` | `matched` | `dexam.dead-elf.map` | YOU DISCOVER A MAP.  ON IT, YOU SEE DEXAMS ALTAR INDICATED AND A PATH THAT SEEMS TO LEAD OUTSIDE. YOU PLACE IT IN YOUR JOURNAL AS ENTRY 59. |
| `ECL4.DAX/0x22` | `0x0905` | `matched` | `dexam.stop-theres-a-trap` | SUDDENLY YELLS, 'STOP! THERE'S A TRAP!' WHAT DO YOU DO? |
| `ECL4.DAX/0x22` | `0x096C` | `matched` | `dexam.removes-gas-trap` | REMOVES THE GAS TRAP. |
| `ECL4.DAX/0x22` | `0x098B` | `matched` | `dexam.dead-elf.gas-trap` | A GAS TRAP GOES OFF! |
| `ECL4.DAX/0x22` | `0x09B3` | `matched` | `dexam.dead-elf.drop-pouch` | YOU DROP THE POUCH AND LEAVE. |
| `ECL4.DAX/0x22` | `0x0A15` | `matched` | `dexam.minotaurs-and-priests` | YOU ARE ATTACKED BY A BAND OF MINOTAURS AND PRIESTS. |
| `ECL4.DAX/0x22` | `0x0A84` | `matched` | `dexam.ogres-attack` | OGRES ATTACK! |
| `ECL4.DAX/0x22` | `0x0ABF` | `matched` | `world.encounter.otyughs` | YOU ARE ATTACKED BY OTYUGHS! |
| `ECL4.DAX/0x22` | `0x0AFE` | `matched` | `dexam.gryphons-attack` | GRYPHONS ATTACK! |
| `ECL4.DAX/0x22` | `0x0B3B` | `matched` | `world.encounter.manticores` | YOU ARE ATTACKED BY MANTICORES! |
| `ECL4.DAX/0x22` | `0x0BA1` | `matched` | `dexam.final_reveal` | YOU HAVE MET UP WITH DEXAM AND HIS MINIONS! |
| `ECL4.DAX/0x22` | `0x0BC8` | `matched` | `dexam.final_reveal` | THE HOODED WOMAN REVEALS HERSELF TO BE A MEDUSA! |
| `ECL4.DAX/0x22` | `0x0C05` | `matched` | `dexam.attack` | THEY ATTACK! |
| `ECL4.DAX/0x22` | `0x0C4E` | `matched` | `dexam.amulet_retrieved` | YOU RETRIEVE THE AMULET OF LATHANDER FROM THE MANGLED REMAINS OF DEXAM. |
| `ECL4.DAX/0x22` | `0x0CBB` | `matched` | `dexam.zhentil_attack` | YOU ARE ATTACKED BY FORCES OF ZHENTIL KEEP! |
| `ECL4.DAX/0x22` | `0x0E92` | `matched` | `dexam.clash-of-weapons` | YOU HEAR THE CLASH OF WEAPONS IN THE DISTANCE. |
| `ECL4.DAX/0x22` | `0x0ED3` | `matched` | `dexam.water-trickles` | WATER TRICKLES DOWN THE SIDES OF THE CAVE. |
| `ECL4.DAX/0x22` | `0x0F11` | `matched` | `dexam.letters-as` | YOU SEE THE LETTERS -- A.S.-- SCRATCHED INTO THE WALL. |
| `ECL4.DAX/0x22` | `0x0F5B` | `matched` | `dexam.ground-trembles` | YOU FEEL THE GROUND BEGIN TO TREMBLE AND SHAKE. IT BUILDS TO A LOUD RUMBLE AND SLOWLY DIES. SMALL ROCKS CLATTER FROM THE CEILING. |
| `ECL4.DAX/0x22` | `0x0FE2` | `matched` | `dexam.muffled-grunts` | YOU HEAR THE MUFFLED GRUNTS AND GROWLS OF SEVERAL LARGE ANIMALS. |
| `ECL4.DAX/0x22` | `0x1034` | `matched` | `dexam.torn-parchment` | YOU FIND A PARTIALLY DESTROYED PIECE OF PARCHMENT. ON IT YOU READ: ...IN THE NORTHWEST...THE WAY OUT... ARCHWAY AND WEST. |
| `ECL4.DAX/0x22` | `0x10BA` | `matched` | `dexam.beholder-unaware` | YOU NOTICE A BEHOLDER FLOATING IN THE DARK SHADOWS AHEAD OF YOU.  ITS MAIN EYE IS TURNED AWAY, AND IT DOESN'T SEEM TO NOTICE YOUR PRESENCE. WHAT DO YOU DO? |
| `ECL4.DAX/0x22` | `0x115C` | `matched` | `dexam.actually-a-gas-spore` | AS YOU CAUTIOUSLY PASS BY, YOU DISCOVER THAT WHAT YOU THOUGHT WAS A BEHOLDER, WAS ACTUALLY A GAS SPORE. YOU BREATH A SIGH OF RELIEF THAT YOU DIDN'T ATTACK THE N… |
| `ECL4.DAX/0x22` | `0x1202` | `matched` | `dexam.thats-not-a-beholder` | YOU EASILY SNEAK UP ON THE FLOATING MONSTER. AS YOUR SWORD DESCENDS IN SAVAGE FURY, YOU HEAR DIMSWART YELL, 'WAIT! THAT'S NOT A BEHOLDER IT'S...!!!' |
| `ECL4.DAX/0x22` | `0x127F` | `matched` | `dexam.gas-spore-explodes-on-you` | TOO LATE, YOU WATCH IN HORROR AS YOUR SWORD PENETRATES THE GAS SPORE WHICH EXPLODES AND COVERS THE PARTY IN A HEAVY, CHOKING GAS. |
| `ECL4.DAX/0x22` | `0x1314` | `matched` | `dexam.departure.olive` | OLIVE RUSKETTLE APPEARS JUST OUTSIDE THE DOOR. SHE SAYS, 'SO, YOU MADE IT.  I'M SURPRISED.' |
| `ECL4.DAX/0x22` | `0x1362` | `matched` | `dexam.departure.dimswart` | 'DIMSWART, YOU COME WITH ME. A MAN FROM SHADOWDALE HAS SOME QUESTIONS FOR YOU.' OLIVE AND DIMSWART LEAVE. |
| `ECL4.DAX/0x22` | `0x13C0` | `matched` | `dexam.departure.gharri` | YOU SEE A RIDER IN THE DISTANCE.  HE STOPS SUDDENLY AND A MOMENT LATER YOU HEAR WHAT SOUNDS LIKE 'GHARRIIII' COME WAFTING THROUGH THE AIR. |
| `ECL4.DAX/0x22` | `0x143A` | `matched` | `dexam.departure.riders` | THE RIDER CHANGES DIRECTION AND GALLOPS OVER TO A WOMAN DRESSED IN PURPLE.  THEY EMBRACE, THE WOMAN CLIMBS ONTO THE BACK OF THE HORSE AND THEY RIDE AWAY. |
| `ECL4.DAX/0x23` | `0x0080` | `matched` | `zhentil.court.dragged-in` | IN YOUR FOGGY HAZE, YOU ARE DRAGGED INTO THE COURTROOM. YOU SEE THE MAGISTRATE LOOK UP AND SMILE WITH ANTICIPATION. |
| `ECL4.DAX/0x23` | `0x00E1` | `matched` | `zhentil.court.cell-until-trial` | THE MAGISTRATE ORDERS YOU INTO A CELL TO REST UP FOR THE TRIAL. |
| `ECL4.DAX/0x23` | `0x0130` | `matched` | `zhentil.court.session` | OYEZ, OYEZ.  THE COURT OF ZHENTIL IS NOW IN SESSION. YOU ARE ON TRIAL FOR MALICIOUS CONDUCT AGAINST THE STATE. HOW DO YOU PLEA? |
| `ECL4.DAX/0x23` | `0x01BB` | `matched` | `zhentil.court.fine` | YOUR FINE IS 95 PERCENT OF YOUR MONEY AND 20 PERCENT OF ALL OTHER ITEMS. |
| `ECL4.DAX/0x23` | `0x0202` | `matched` | `zhentil.court.send-to-arena` | SEND THEM TO THE ARENA WHERE THEIR WORTHINESS TO REMAIN IN ZHENTIL WILL BE TESTED. |
| `ECL4.DAX/0x23` | `0x024B` | `matched` | `zhentil.arena.dragged-in` | YOU ARE DRAGGED INTO THE ARENA. |
| `ECL4.DAX/0x23` | `0x0280` | `matched` | `zhentil.arena.dragged-in` | YOU HAVE BEEN LED INTO A LARGE OPEN SPACE. THE DIRT IS SCUFFED AND STAINED A SUSPICIOUS REDDISH BROWN. |
| `ECL4.DAX/0x23` | `0x02DC` | `matched` | `zhentil.arena.fanfare` | THE WALLS AROUND ARE PIERCED WITH WINDOWS AND COVERED WITH PEOPLE WATCHING IN ANTICIPATION. YOU SUDDENLY HEAR A GREAT, BLARING FANFARE. |
| `ECL4.DAX/0x23` | `0x0396` | `matched` | `zhentil.arena.gladiators-attack` | YOU ARE BEING ATTACKED BY GLADIATORS! |
| `ECL4.DAX/0x23` | `0x03CE` | `matched` | `zhentil.court.bane-smiles` | YOU ARE ESCORTED BACK OUTSIDE. |
| `ECL4.DAX/0x23` | `0x0423` | `matched` | `zhentil.court.bane-smiles` | THE MAGISTRATE SAYS FROM HIS DOOR, 'BANE SMILES ON YOU. YOU MAY GO IN PEACE.' |
| `ECL4.DAX/0x23` | `0x0468` | `matched` | `zhentil.court.dont-screw-up` | 'JUST DON'T SCREW UP AGAIN!' |
| `ECL4.DAX/0x23` | `0x048C` | `matched` | `zhentil.arena.gryphons` | GRYPHONS! |
| `ECL4.DAX/0x23` | `0x0566` | `matched` | `zhentil.tavern.drink-prompt` | 'WHAT CAN I GET YOU TO DRINK?' |
| `ECL4.DAX/0x23` | `0x05BC` | `matched` | `zhentil.tavern.moo-juice` | THE PROPRIETOR CHUCKLES AND CRYS IN A LOUD VOICE, 'GET THESE 'ADVENTURERS' SOME MOO-JUICE!' |
| `ECL4.DAX/0x23` | `0x0628` | `matched` | `zhentil.tavern.gamble-prompt` | WOULD YOU LIKE TO GAMBLE ON THE ARENA COMBATS? |
| `ECL4.DAX/0x23` | `0x0672` | `matched` | `zhentil.arena-rules` | THE RULES ARE EXPLAINED TO YOU, AND YOU NOTE THEM DOWN AS JOURNAL ENTRY 23. |
| `ECL4.DAX/0x23` | `0x06C4` | `matched` | `zhentil.tavern.deadbeats` | 'THEN GET OUT OF MY ESTABLISHMENT! DEADBEATS.' |
| `ECL4.DAX/0x23` | `0x06EE` | `matched` | `zhentil.tavern.stay-prompt` | DO YOU WANT TO STAY IN THE TAVERN? |
| `ECL4.DAX/0x23` | `0x0739` | `matched` | `zhentil.arena.condemned-vs-monsters` | THE CONDEMNED PRISONERS VS. THE MONSTERS OF ZHENTIL KEEP! WHO DO YOU PICK? |
| `ECL4.DAX/0x23` | `0x079D` | `matched` | `zhentil.arena.accused-vs-monsters` | THE ACCUSED VS. THE MONSTERS OF ZHENTIL KEEP! WHO DO YOU PICK? |
| `ECL4.DAX/0x23` | `0x07F0` | `matched` | `zhentil.arena.decreed-guilty` | THE FIGHT IS ON! |
| `ECL4.DAX/0x23` | `0x0827` | `matched` | `zhentil.arena.prisoners-lost` | THE PRISONERS LOST. SO WHAT ELSE IS NEW? I'M VERY SORRY, PERHAPS NEXT TIME? |
| `ECL4.DAX/0x23` | `0x0857` | `matched` | `zhentil.arena.decreed-guilty` | BANE HAS DECREED THE PRISONERS GUILTY I'M VERY SORRY, PERHAPS NEXT TIME? |
| `ECL4.DAX/0x23` | `0x0885` | `matched` | `zhentil.arena.accused-innocent` | BANE IS MERCIFUL.  THE ACCUSED ARE INNOCENT. I'M VERY SORRY, PERHAPS NEXT TIME? |
| `ECL4.DAX/0x23` | `0x08AD` | `matched` | `zhentil.arena.prisoners-won` | UNBELIEVABLE, THOSE RAGGED PRISONERS ACTUALLY WON! I'M VERY SORRY, PERHAPS NEXT TIME? |
| `ECL4.DAX/0x23` | `0x08DA` | `matched` | `zhentil.arena.killed-each-other` | THEY HAVE KILLED EACH OTHER! I'M VERY SORRY, PERHAPS NEXT TIME? |
| `ECL4.DAX/0x23` | `0x095B` | `variable-insert` | — | CONGRATULATIONS, YOU HAVE WON PLATINUM. |
| `ECL4.DAX/0x23` | `0x099C` | `variable-insert` | — | YOU OVERHEAR TAVERN TALE # |
| `ECL4.DAX/0x23` | `0x09FC` | `matched` | `zhentil.tavern.one-platinum` | THAT'LL BE 1 PLATINUM PIECE. |
| `ECL4.DAX/0x23` | `0x0A43` | `matched` | `zhentil.tavern.cant-afford` | YOU CAN'T AFFORD IT. 'ANYONE ELSE HAVE ANY MONEY?' |
| `ECL4.DAX/0x25` | `0x005D` | `matched` | `world.magic-shop` | YOU DISCOVER A SMALL MAGIC SHOP. WHAT TO YOU DO? |
| `ECL4.DAX/0x25` | `0x00CE` | `matched` | `dexam.tower-of-oxam-rumor` | 'THE MULMASTER BEHOLDER CORPS IS RUMORED TO BE HOLED UP IN THE TOWER OF OXAM.' |
| `ECL4.DAX/0x25` | `0x01DF` | `matched` | `world.drow-items-fade` | AFTER A FEW HOURS OUT IN THE SUN, YOUR DROW ITEMS FADE TO USELESSNESS. |
| `ECL4.DAX/0x25` | `0x02CF` | `matched` | `dexam.manor-foyer` | YOU ARE IN THE FOYER OF THE MANOR. |
| `ECL4.DAX/0x25` | `0x030B` | `matched` | `dexam.sitting-room` | THIS IS THE REMAINS OF A LARGE SITTING ROOM. |
| `ECL4.DAX/0x25` | `0x0385` | `matched` | `dexam.library` | THIS IS THE LIBRARY.  MANY OLD BOOKS LINE THE SHELVES. |
| `ECL4.DAX/0x25` | `0x03DA` | `matched` | `dexam.books-crumble` | EVERY BOOK YOU TOUCH CRUMBLES INTO ASHES. THERE IS NOTHING HERE. |
| `ECL4.DAX/0x25` | `0x05A9` | `matched` | `dexam.image-welcomes` | AN IMAGE FORMS IN FRONT OF YOU. |
| `ECL4.DAX/0x25` | `0x05D3` | `matched` | `dexam.image-welcomes` | 'WELCOME, MY FRIENDS.  I AWAIT YOUR ARRIVAL. IN THE MEANWHILE, HAVE SOME FUN WITH MY PETS.' |
| `ECL4.DAX/0x25` | `0x0785` | `matched` | `dexam.high-priest-of-bane` | BEFORE YOU IS A HIGH PRIEST OF BANE. |
| `ECL4.DAX/0x25` | `0x0820` | `matched` | `dexam.dark-shape-archway-only` | A DARK SHAPE SCUTTLES THROUGH THE ARCHWAY AHEAD AND TO YOUR LEFT. |
| `ECL4.DAX/0x25` | `0x0876` | `matched` | `dexam.noise-from-west-wall` | NOISE COMES FROM THE WALL TO THE WEST. |
| `ECL4.DAX/0x25` | `0x08BF` | `matched` | `dexam.medusi-attack` | MEDUSI AND THEIR BODYGUARDS ATTACK! |
| `ECL4.DAX/0x25` | `0x091A` | `matched` | `dexam.arrow-trap` | AN ARROW TRAP GOES OFF! |
| `ECL4.DAX/0x25` | `0x093E` | `matched` | `dexam.feel-very-strange` | YOU FEEL VERY STRANGE. |
| `ECL4.DAX/0x25` | `0x09A6` | `matched` | `world.encounter.fighter-band` | A BAND OF FIGHTERS BURSTS THROUGH THE DOOR BEHIND YOU. |
| `ECL4.DAX/0x25` | `0x09FA` | `matched` | `dexam.red-stained-room` | THIS ROOM IS COVERED WITH DARK RED STAINS. CHAINS ARE EMBEDDED IN THE WALLS, AND PILES OF BONES ARE SCATTERED ABOUT. |
| `ECL4.DAX/0x25` | `0x0A9C` | `matched` | `dexam.medusa-grisly-meal` | A MEDUSA LOOKS UP FROM A GRISLY MEAL. |
| `ECL4.DAX/0x25` | `0x0B02` | `matched` | `dexam.round-shape-turns` | A LARGE ROUND SHAPE SLOWLY TURNS TOWARDS YOU. WHAT TO YOU DO? |
| `ECL4.DAX/0x25` | `0x0B4A` | `matched` | `dexam.gas-spore-explodes` | THE GAS SPORE EXPLODES AT THE TOUCH OF YOUR WEAPONS. |
| `ECL4.DAX/0x25` | `0x0C01` | `matched` | `dexam.man-cowers-beholder-corps` | A MAN COWERS IN THE SHADOWS.  HE GASPS, 'THE BEHOLDER CORPS IS... GAAK!' |
| `ECL4.DAX/0x25` | `0x0C49` | `matched` | `dexam.round-shape-disappears` | A LARGE ROUND SHAPE DISAPPEARS INTO THE DARKNESS. |
| `ECL4.DAX/0x25` | `0x0C9B` | `matched` | `dexam.corps-is-merciful` | A BEHOLDER FLOATS IN FRONT OF YOU. 'YOU ARE NOT OF THE CREATURES INVITED TO THIS CONFERENCE.  HOWEVER, THE CORPS IS MERCIFUL. GO NOW, IF YOU VALUE YOUR LIVES.' |
| `ECL4.DAX/0x25` | `0x0D85` | `matched` | `dexam.guards-cower` | LOCAL GUARDS COWER BACK AT YOUR APPROACH. WHAT TO YOU DO? |
| `ECL4.DAX/0x25` | `0x0DC9` | `matched` | `dexam.furry-faces` | THE GUARDS GET VERY STRANGE, WICKED SMILES ON THEIR SUDDENLY FURRY FACES. |
| `ECL4.DAX/0x25` | `0x0E32` | `matched` | `dexam.invitation-card` | YOU FIND A SMALL, ELEGANT CARD WHICH SAYS... THE HIGH IMPERCEPTOR OF MULMASTER COMMANDS YOUR PRESENSE AT THE TOWER OF OXAM.  THE BEHOLDER CORPS WILL BE YOUR HOS… |
| `ECL4.DAX/0x25` | `0x0EDF` | `matched` | `dexam.drow-examines-card` | A LARGE DROW ELF EXAMINES A SMALL CARD. HE LOOKS UP AT YOU AND DRAWS HIS WEAPON. WHAT TO YOU DO? |
| `ECL4.DAX/0x25` | `0x0F8A` | `matched` | `dexam.beholders-floating-object` | SEVERAL BEHOLDERS ARE GATHERED AROUND A SMALL, FLOATING OBJECT. AS YOU ENTER, ONE OF THEM LOOKS UP AND NOTICES YOU. WHAT TO YOU DO? |
| `ECL4.DAX/0x25` | `0x1034` | `matched` | `dexam.find-a-list` | YOU FIND A LIST. |
| `ECL4.DAX/0x25` | `0x1047` | `matched` | `dexam.attendance-list` | RAKSHASA   YES EFREETI    NO VAMPIRE    NO DROW       YES LICH       NO |
| `ECL4.DAX/0x25` | `0x10B1` | `matched` | `dexam.conference-table` | YOU ENTER A ROOM DOMINATED BY A LARGE CONFERENCE TABLE. ON EACH SIDE IS AN ASSEMBLAGE OF RAKSHASA, DROW, PRIESTS AND BEHOLDERS. THEY ARE DISCUSSING THE FATE OF … |
| `ECL4.DAX/0x25` | `0x1534` | `matched` | `world.encounter.fighter-band` | YOU ARE ATTACKED BY FIGHTERS AND CLERICS! |
| `ECL4.DAX/0x25` | `0x1585` | `matched` | `dexam.circular-stairway` | YOU SEE A CIRCULAR STAIRWAY GOING DOWN. DO YOU WISH TO GO DOWN? |
| `ECL4.DAX/0x25` | `0x15B7` | `subroutine` | — | WHAT TO YOU DO? |
| `ECL5.DAX/0x30` | `0x0034` | `matched` | `area5.dark-elf-gear-decays` | AFTER A SHORT TIME IN THE SUNLIGHT, YOUR DARK ELF WEAPONS AND ARMOR DECAY TO USELESSNESS. |
| `ECL5.DAX/0x30` | `0x0125` | `matched` | `area5.depart-akabar` | AKABAR SPEAKS, 'YOUR HELP WAS INVALUABLE TO ME. GOOD LUCK ON YOUR QUEST. I FEAR I HAVE BUSINESS TO ATTEND TO.' HE LEAVES. |
| `ECL5.DAX/0x30` | `0x018A` | `matched` | `area5.depart-akabar-reluctant` | AKABAR FROWNS AT YOU AND SAYS, 'I AM SURE YOU MUST HAVE YOUR REASONS TO LEAVE, BUT I MUST FREE THIS TOWN FROM THE WIZARD'S TYRANY. GOOD DAY.' HE STAMPS OFF. |
| `ECL5.DAX/0x31` | `0x005F` | `matched` | `hap.abandoned-village` | THIS RUN DOWN VILLAGE IS STRANGELY QUIET. THE WIND WHISTLES DOWN THE EMPTY STREET, PAST SHUTTERED WINDOWS. NO ONE IS ABOUT. |
| `ECL5.DAX/0x31` | `0x01B0` | `matched` | `hap.akabar-presses-forward` | AKABAR PRESSES FORWARD. 'YOU FILTHY CREATURES THINK YOU CAN CRUSH THIS TOWN. WE'LL DESTROY YOU FIRST!' THE ELVES' EXPRESSIONS HARDEN. |
| `ECL5.DAX/0x31` | `0x0226` | `matched` | `hap.patrol-kicks-dirt` | 'OFF THE STREET PEASANT SCUM. YOUR PATHETIC KIND MAKE US SICK.' THE PATROL KICKS DIRT ON YOUR SHOES AND MOVES ON. |
| `ECL5.DAX/0x31` | `0x028C` | `matched` | `hap.dark-elf-attack` | 'YOU UNGRATEFUL SLIME. YOU'VE TRIED OUR PATIENCE ONCE TOO OFTEN. BE HAPPY FOR A QUICK DEATH.' |
| `ECL5.DAX/0x31` | `0x034C` | `matched` | `hap.move-on-to-the-inn` | 'PLEASE MOVE ON TO THE INN. YOU PUT US IN GRAVE DANGER.' THE PEASANTS THEN FLEE. |
| `ECL5.DAX/0x31` | `0x03F7` | `matched` | `hap.leave` | YOU ARE HEADING BACK TO THE WILDERNESS. DO YOU WANT TO CONTINUE? |
| `ECL5.DAX/0x31` | `0x0442` | `matched` | `hap.map-route` | DO YOU WANT TO FOLLOW THE MAP TO THE CAVES OR GO INTO THE WILDERNESS? |
| `ECL5.DAX/0x31` | `0x0554` | `matched` | `hap.hiding-peasants` | YOU BURST IN ON SOME PEASANTS WHO SCUTTLE BACK AND CRY, 'LEAVE BEFORE THE HORDE FINDS YOU WITH US.' WHAT DO YOU DO? |
| `ECL5.DAX/0x31` | `0x05DD` | `matched` | `hap.peasants-flee` | THE CRINGING PEASANTS FLEE OUT INTO THE STREET. |
| `ECL5.DAX/0x31` | `0x061B` | `matched` | `hap.stay-as-long-as-you-like` | 'THANK YOU AGAIN FOR ALL YOU'VE DONE. STAY AS LONG AS YOU'D LIKE.' |
| `ECL5.DAX/0x31` | `0x0683` | `matched` | `hap.general-store` | YOU ARE IN A SMALL GENERAL STORE. THE SHOPKEEPER RUSHES UP AT YOUR ENTRANCE.'THANK YOU FOR SAVING OUR TOWN. ANYTHING I OWN IS YOURS -- FOR A SIGNIFICANT DISCOUN… |
| `ECL5.DAX/0x31` | `0x0775` | `matched` | `hap.efreet-barn` | THIS BARN IS EMPTY -- SAVE FOR THE EFREET AND HIS DARK ELFIN COHORTS. |
| `ECL5.DAX/0x31` | `0x07BD` | `matched` | `hap.efreet-threat` | THE EFREET VOICE BOOMS OUT,'SO, THE PATHETIC WORMS SHOW SOME SPINE. WE WILL KILL YOU, THEN BURN DOWN THIS WRETCHED HEAP OF HOVELS. YOU HAVE BROUGHT DOOM ON YOUR… |
| `ECL5.DAX/0x31` | `0x08B7` | `matched` | `hap.efreet-map` | ON THE BODY OF THE EFREET IS A MAP INDICATING THE TOWN AND A CAVE. |
| `ECL5.DAX/0x31` | `0x0902` | `matched` | `hap.liberated-crowd` | A SHORT TIME AFTER THE SOUNDS OF BATTLE FADE, A FEW TIMID HEADS POKE INTO THE BARN. THEN SWIFTLY A HUGE CROWD GATHERS. SOON THE VILLAGE IS RINGING WITH LOUD CHE… |
| `ECL5.DAX/0x31` | `0x0999` | `matched` | `hap.elder-thanks` | AN ELDER OF THE VILLAGE COMES FORWARD. 'WE WILL BE FOREVER IN YOUR DEBT. YOU WILL ALWAYS BE WELCOME IN HAPTOOTH.' |
| `ECL5.DAX/0x31` | `0x0A07` | `matched` | `hap.elder-wizard-tower` | THE ELDER LOWERS HIS VOICE. 'I DO NOT WISH TO SEEM UNGRATEFUL, BUT THESE ELVES ARE CONTROLLED FROM THE WIZARD'S TOWER NEARBY. I FEAR WE WILL ONLY BE SAFE IF YOU… |
| `ECL5.DAX/0x31` | `0x0AAE` | `matched` | `hap.akabar-secret-routes` | AKABAR MENTIONS THAT HE HAS HEARD OF SECRET TRADE ROUTES THAT LEAD PAST THE TOWER. HE WILL BE HAPPY TO GUIDE THE PARTY THERE. |
| `ECL5.DAX/0x31` | `0x0B59` | `matched` | `hap.akabar-join` | 'SO YOU HAVE FINALLY COME. I AM AKABAR BEL AKASH. BETWEEN US, WE CAN CRUSH THIS DARK WAVE.' WILL YOU LET HIM JOIN YOUR PARTY? |
| `ECL5.DAX/0x31` | `0x0BDE` | `matched` | `hap.akabar-alone` | 'I UNDERSTAND. I WILL DO MY BEST ON MY OWN.' |
| `ECL5.DAX/0x31` | `0x0C24` | `matched` | `hap.inn-before-liberation` | A SURLY INNKEEPER COMES UP.'CLOSE THE DOOR! THE HORDE IS ABOUT. YOU CAN STAY IF YOU WANT, JUST KEEP LOW.' DO YOU STAY? |
| `ECL5.DAX/0x31` | `0x0CA2` | `matched` | `hap.thank-you-for-bravery` | 'THANK YOU FOR YOUR BRAVERY. STAY AS LONG AS YOU WANT. WE'LL GET YOU THE BEST WE HAVE.' |
| `ECL5.DAX/0x31` | `0x0D2F` | `matched` | `hap.temple-of-sune` | A PRIEST APPROACHES, 'WE WILL HIDE YOU HERE, IN THE TEMPLE OF SUNE, BUT WE CAN DO NO MORE WHILE THE HORDE IS ABOUT. WE ARE TOO WEAK IN THIS TOWN.' |
| `ECL5.DAX/0x31` | `0x0DBE` | `matched` | `hap.come-in-and-rest` | 'PLEASE COME IN AND REST YOURSELVES. WE SHALL HEAL YOU IF YOU WISH.' |
| `ECL5.DAX/0x32` | `0x0072` | `matched` | `lava-tube.entry` | YOU HAVE ENTERED AN ANCIENT LAVA TUBE. ASH COVERS THE FLOOR. |
| `ECL5.DAX/0x32` | `0x00BC` | `matched` | `lava-tube.ambush` | FROM HIDDEN ALCOVES COMES A WAVE OF HEAT, FOLLOWED BY SALAMANDERS AND DARK ELVES. |
| `ECL5.DAX/0x32` | `0x0165` | `matched` | `lava-tube.salamander-pools` | THE ROOM IS FILLED WITH ACTIVE GEYSERS AND LAVA PITS. SALAMANDERS ARE SPORTING IN THE POOLS. |
| `ECL5.DAX/0x32` | `0x0204` | `matched` | `lava-tube.intense-heat` | INTENSE HEAT WASHES OVER YOU. |
| `ECL5.DAX/0x32` | `0x029E` | `matched` | `lava-tube.sly-parlay` | 'WE HAVE NO LOVE FOR DARK ELVES. TAKE ANY TREASURE YOU WISH.' |
| `ECL5.DAX/0x32` | `0x02E1` | `matched` | `lava-tube.nice-parlay` | 'YOU COLD THINGS SHOULD LEAVE BEFORE CRIMDRAC FINDS YOU.' |
| `ECL5.DAX/0x32` | `0x031B` | `matched` | `lava-tube.fireproof-casks` | AMONGST THE POOLS OF LAVA, ARE SIX FIREPROOF CASKS. DOES ANYONE WANT TO GO AND OPEN ONE? |
| `ECL5.DAX/0x32` | `0x0384` | `matched` | `lava-tube.cask-unhealthy` | THAT ONE IS NOT HEALTHY ENOUGH. |
| `ECL5.DAX/0x32` | `0x03C3` | `matched` | `lava-tube.cask-heat-retreat` | THE HEAT IS TOO INTENSE. YOU HAVE TO RETREAT. DOES ANYONE WANT TO TRY AGAIN? |
| `ECL5.DAX/0x32` | `0x043A` | `matched` | `lava-tube.cask-try-another` | DO YOU WANT TO TRY FOR ANOTHER? |
| `ECL5.DAX/0x32` | `0x0459` | `matched` | `lava-tube.casks-emptied` | ALL THE CASKS HAVE BEEN EMPTIED. |
| `ECL5.DAX/0x32` | `0x047C` | `matched` | `lava-tube.exit-to-wilderness` | THIS WILL TAKE YOU TO THE WILDERNESS. DO YOU WANT TO CONTINUE? |
| `ECL5.DAX/0x32` | `0x04BA` | `matched` | `lava-tube.return-or-beyond` | DO YOU WANT TO RETURN TO HAPTOOTH VILLAGE OR HEAD BEYOND? |
| `ECL5.DAX/0x32` | `0x0513` | `matched` | `wizard-tower.courtyard.entering` | YOU ARE HEADING UP INTO THE WIZARD'S TOWER. |
| `ECL5.DAX/0x32` | `0x055F` | `matched` | `lava-tube.blocked-rubble` | THE WAY IS BLOCKED WITH RUBBLE. |
| `ECL5.DAX/0x32` | `0x0684` | `matched` | `lava-tube.too-weak-toll` | 'YOU ARE TOO WEAK. WE WILL TAKE SOME OF YOUR TREASURE, AND THEN YOU CAN HEAD ON TO CRIMDRAC.' DO YOU SURRENDER? |
| `ECL5.DAX/0x32` | `0x06DD` | `matched` | `lava-tube.they-take-money` | THEY TAKE MONEY AND ITEMS. |
| `ECL5.DAX/0x32` | `0x0787` | `matched` | `lava-tube.patrol-attacks` | A PATROL SPOTS YOU AND ATTACKS. |
| `ECL5.DAX/0x32` | `0x07DA` | `matched` | `lava-tube.elves-see-mark` | A PARTY OF ELVES SPOTS YOU, BUT SEEING SILK'S MARK, THEY MOVE AWAY. |
| `ECL5.DAX/0x32` | `0x0872` | `matched` | `lava-tube.arrow-in-ash` | YOU FIND AN ARROW POINTING WEST FAINTLY SCRATCHED IN THE ASH. |
| `ECL5.DAX/0x32` | `0x08C1` | `matched` | `lava-tube.stone-arrow-south` | AN ARROW MADE OF SMALL STONES POINTS SOUTH, HERE. |
| `ECL5.DAX/0x32` | `0x0908` | `matched` | `lava-tube.elves-see-mark` | YOU HEAR A PATROL APPROACH FROM BEHIND YOU. |
| `ECL5.DAX/0x32` | `0x095E` | `matched` | `lava-tube.four-dark-elves` | FOUR FEMALE DARK ELVES STEP FROM THE SHADOWS. 'WE WERE EXPECTING YOU. WE WILL ESCORT YOU TO OUR COMMANDER, WHO HAS A PROPOSITION YOU MAY BE INTERESTED IN.' |
| `ECL5.DAX/0x32` | `0x09E4` | `matched` | `lava-tube.accompany-them` | WILL YOU ACCOMPANY THEM? |
| `ECL5.DAX/0x32` | `0x0A40` | `matched` | `lava-tube.next-cavern` | AS YOU WISH. WE WILL BE IN THE NEXT CAVERN IF YOU CHANGE YOUR MIND. |
| `ECL5.DAX/0x32` | `0x0A87` | `matched` | `lava-tube.paranoid-fools` | 'PARANOID FOOLS!' THEY CRY AS THEY SCATTER AND ESCAPE INTO THE SHADOWS. |
| `ECL5.DAX/0x32` | `0x0AFB` | `matched` | `lava-tube.silk-steps-forward` | A DARK ELFIN WOMAN STEPS FORWARD. HER HAIR IS DARK WITH A SINGLE STREAK OF WHITE. ON THE BACK OF HER HAND IS A SYMBOL THAT RESEMBLES A STYLIZED SWAN. |
| `ECL5.DAX/0x32` | `0x0B89` | `matched` | `lava-tube.silk-introduction` | 'YOU MAY CALL ME SILK. I HAVE BEEN WAITING FOR A GROUP SUCH AS YOU. , STEP FORWARD AND RECEIVE OUR MARK.' |
| `ECL5.DAX/0x32` | `0x0BEE` | `matched` | `lava-tube.silk-mark-offer` | SHE CONTINUES, AND YOU RECORD IT IN JOURNAL ENTRY 44. WILL STEP FORWARD? |
| `ECL5.DAX/0x32` | `0x0C4B` | `matched` | `lava-tube.silk-marks-hand` | STEPS FORWARD. SILK PASSES HER HAND OVER THE BACK OF 'S HAND, AND A TATOO OF A STYLIZED SWAN APPEARS. |
| `ECL5.DAX/0x32` | `0x0CAC` | `matched` | `lava-tube.silk-symbol-luck` | 'I WISH YOU GOOD LUCK. THE SYMBOL MAY MAKE SOME PATROLS AVOID YOU, BUT IT WON'T BE COMPLETELY SAFE. SOME OF THE GUARDS ARE RELATED TO ME, AND I STILL HAVE INFLU… |
| `ECL5.DAX/0x32` | `0x0D5F` | `matched` | `lava-tube.silk-refused` | 'THAT IS TOO BAD. PERHAPS YOU WILL CHANGE YOUR MIND LATER. I CAN BE OF NO MORE HELP.' SHE LEAVES. |
| `ECL5.DAX/0x32` | `0x0DB7` | `matched` | `lava-tube.silk-needs-a-woman` | IT IS TOO BAD YOU HAVE NO WOMEN. ONLY THEY CAN WEAR OUR MARK. I AM AFRAID I CAN BE OF NO MORE HELP UNTIL ONE JOINS YOU.' SHE LEAVES. |
| `ECL5.DAX/0x32` | `0x0E9A` | `matched` | `lava-tube.violated-our-precinct` | SOME DARK ELVES ARE HERE, ATOP A MOUND OF FRESHLY TURNED EARTH. 'YOU HAVE VIOLATED OUR PRECINCT. YOUR LIFE IS FORFEIT.' STRANGE SHAPES RISE UP BEHIND THEM. |
| `ECL5.DAX/0x32` | `0x0F53` | `matched` | `lava-tube.barracks-disturbed` | YOU HAVE DISTURBED A BARRACKS FULL OF DARK ELVES WHO RISE UP IN ANGER. |
| `ECL5.DAX/0x32` | `0x0FD9` | `matched` | `lava-tube.incense-room` | THIS ROOM IS FILLED WITH CLOYING INCENSE SMOKE. A LARGE NUMBER OF DARK ELFIN CLERICS ARE LYING ON PILLOWS AND SMOKING FROM HOUKAS. |
| `ECL5.DAX/0x32` | `0x1053` | `matched` | `lava-tube.move-on-sister` | ONE OF THE CLERICS SPEAKS, 'MOVE ON SISTER. WE SHOULD HAVE NO QUARREL.' DO YOU RETREAT? |
| `ECL5.DAX/0x32` | `0x10E0` | `matched` | `lava-tube.guarded-door` | THE DOOR IS GUARDED BY A SALAMANDER LED PATROL. |
| `ECL5.DAX/0x32` | `0x112D` | `matched` | `lava-tube.dream-warning` | A DREAM-LIKE VOICE IN YOUR HEAD SAYS, 'GREAT DANGER LIES BEFORE YOU. BE FULLY PREPARED!' |
| `ECL5.DAX/0x32` | `0x11A1` | `matched` | `lava-tube.mage-meditation-room` | MYSTIC SYMBOLS ADORNE THE WALLS. MAGES ARE HERE MEDITATING AND CHANTING. ONE SPEAKS. ' |
| `ECL5.DAX/0x32` | `0x11F9` | `matched` | `lava-tube.dangerous-ground-sister` | YOU TREAD ON DANGEROUS GROUND SISTER. RETREAT NOW LEST WE BE FORCED TO RAISE OUR HANDS AGAINST YOU.' DO YOU RETREAT? |
| `ECL5.DAX/0x32` | `0x1290` | `matched` | `lava-tube.crimdrac-introduces` | CURLED IN THE CENTER OF THIS ROOM IS THE HUGE SKELETAL FORM OF A DRACOLICH. 'I AM CRIMDRAC, MORTALSSS. YOU HAVE REACHED THE HEART OF MY DOMAIN. IT PLEASSSESSS M… |
| `ECL5.DAX/0x32` | `0x1331` | `matched` | `lava-tube.crimdrac-foolishness` | 'FOOLISSSHNESSS WASSS ALWAYSSS IN A MORTAL'SSS HEART. PERHAPSSS YOU WILL RECONSSSIDER LATER.' |
| `ECL5.DAX/0x32` | `0x13A5` | `matched` | `lava-tube.crimdrac-short-walk` | CRIMDRAC ROUNDS YOU UP AND TAKES YOU ON A SHORT WALK. |
| `ECL5.DAX/0x32` | `0x13E3` | `matched` | `lava-tube.crimdrac-short-walk` | HE GIVES YOU A PUSH UP A RAMP. 'GIVE MY GREETINGSSS TO DRACANDROSSS.' |
| `ECL5.DAX/0x32` | `0x1449` | `matched` | `lava-tube.guarded-by-efreeti` | THIS WAY IS GUARDED BY EFREETI AND DARK ELVES. |
| `ECL5.DAX/0x32` | `0x1497` | `matched` | `lava-tube.tunnel-collapsed` | AT YOUR APPROACH, THE ELVES COLLAPSE THE TUNNEL. YOU HEAR, 'NONE SHALL REACH THE DIVINE CITY.' |
| `ECL5.DAX/0x32` | `0x1518` | `matched` | `lava-tube.silk-buys-heart` | SILK STEPS OUT FROM THE SHADOWS.' CONGRATULATIONS ON YOUR SUCCESSES. I WILL PAY YOU WELL FOR THE DRAGON'S BLACK HEART AND BLACK EGGS , IF YOU ARE INTERESTED. AR… |
| `ECL5.DAX/0x32` | `0x15EA` | `matched` | `lava-tube.carry-our-mark` | 'THEN TAKE CARE, AND CARRY OUR MARK PROUDLY. THE SAFEST WAY OUT SHOULD BE THROUGH THE SECRET DOOR OF THE TOWER.' SHE SLIPS BACK INTO THE SHADOWS. |
| `ECL5.DAX/0x33` | `0x0068` | `matched` | `wizard-tower.courtyard.description` | YOU HAVE COME OUT INTO THE COURTYARD OF A FIVE STORY TOWER. THE STONEWORK HAS BEEN MYSTICLY PROTECTED , SO IS FLAWLESS AND BEAUTIFUL. SURROUNDING THE TOWER ARE … |
| `ECL5.DAX/0x33` | `0x0120` | `matched` | `wizard-tower.dracandros.arrival` | AN IMPRESSIVE ROBED FIGURE APPROACHES YOU. ' I AM DRACANDROS. I AM GLAD YOU HAVE FINALLY ARRIVED. TIME IS SHORT AND YOU MUST PLAY YOUR PART.' |
| `ECL5.DAX/0x33` | `0x01BC` | `matched` | `wizard-tower.dracandros.freezes-party` | 'FREEZE WHERE YOU STAND! I'VE NO TIME FOR THIS NOW!' THE BONDS PARALYZE YOU. A DARK ELF RESTRAINS AKABAR WITH A SPELL. ⚠ 另有 `GOSUB 0x01FB` 印出的一段 |
| `ECL5.DAX/0x33` | `0x0241` | `matched` | `wizard-tower.dragon-roof` | YOU ARE SUDDENLY ON THE ROOF OF THE TOWER AMIDST A HUGE HOST OF BLACK DRAGONS. |
| `ECL5.DAX/0x33` | `0x028A` | `matched` | `wizard-tower.dragon-steps-out` | ONE OF THE DRAGONS DISENGAGES HIMSELF FROM THE PACK. |
| `ECL5.DAX/0x33` | `0x02CA` | `matched` | `wizard-tower.dracandros.attack-order` | 'ATTACK THE DRAGON AS ELMINSTER TOLD YOU TO!' |
| `ECL5.DAX/0x33` | `0x02F6` | `matched` | `wizard-tower.dragon-illusion` | UNDER THE FORCE OF THE BONDS, YOU RUSH FORWARD AND ATTACK THE DRAGON, BUT WITH A BLOW, IT DISAPPEARS IN A PUFF OF SMOKE. THE DRAGON WAS ONLY AN ILLUSION! |
| `ECL5.DAX/0x33` | `0x0387` | `matched` | `wizard-tower.dracandros.journal-15` | 'FREEZE, BASE SLAYERS OF DRAGONKIND!' CRIES DRACANDROS. YOU ARE IMMOBILIZED. HE THEN TURNS TO THE ASSEMBLED DRAGONS AND GIVES A SPEECH YOU RECORD AS JOURNAL ENT… |
| `ECL5.DAX/0x33` | `0x041E` | `matched` | `wizard-tower.dracandros.bond-fades` | DRACANDROS' MUMBLED PHRASE CAUSES YOUR BONDS TO TO FADE. |
| `ECL5.DAX/0x33` | `0x04C0` | `matched` | `wizard-tower.dragons-condemn` | 'YOU ARE RIGHT DRACANDROS. THEY CONDEMN THEMSELVES.' |
| `ECL5.DAX/0x33` | `0x050F` | `matched` | `wizard-tower.dragons-depart` | 'THIS IS A MATTER BETWEEN MEN. WE LEAVE YOU TO YOUR SQUABBLES.' THE DRAGONS TAKE FLIGHT AND LEAVE. |
| `ECL5.DAX/0x33` | `0x0573` | `matched` | `wizard-tower.dracandros.calls-troops` | 'TROOPS DEFEND ME!' A PATROL RUSHES FORWARD WHILE DRACANDROS FLEES DOWN THE STAIRS. |
| `ECL5.DAX/0x33` | `0x0609` | `matched` | `wizard-tower.dragons-convinced` | 'YOU HAVE CONVINCED US THAT THERE IS NO PLOT AGAINST DRAGONKIND. WE LEAVE YOU NOW TO SETTLE YOUR DISPUTE WITH DRACANDROS.' |
| `ECL5.DAX/0x33` | `0x0676` | `matched` | `wizard-tower.take-dragon-heart` | DURING THE FIGHT, DRACANDROS ESCAPED DOWNSTAIRS. THE DRAGON BODIES LIE STREWN ABOUT. DO YOU TAKE ONE OF THEIR HEARTS? |
| `ECL5.DAX/0x33` | `0x06EC` | `matched` | `wizard-tower.dragon-heart-acid` | AS YOU CUT INTO THE DRAGON TO REMOVE ITS VITALS, YOU ARE DRENCHED IN A SPRAY OF ACID, BUT MANAGE TO EXTRACT THE HEART. |
| `ECL5.DAX/0x33` | `0x0768` | `matched` | `wizard-tower.safe-roof` | IT LOOKS LIKE YOU CAN HOLD THE ROOF WELL ENOUGH TO REST SAFELY. |
| `ECL5.DAX/0x33` | `0x07EA` | `matched` | `wizard-tower.roof-exit` | THIS IS THE WAY DOWN TO THE CAVES. HOWEVER, YOU ALSO NOTE A SECRET PASSAGE THAT WILL TAKE YOU DIRECTLY TO THE WILDERNESS. WHICH DO YOU TAKE? |
| `ECL5.DAX/0x33` | `0x08A6` | `matched` | `wizard-tower.wilderness-exit` | DO YOU WANT TO STOP BY HAPTOOTH VILLAGE OR DEPART THE AREA? |
| `ECL5.DAX/0x33` | `0x091F` | `matched` | `wizard-tower.stairs.down` | THE STAIRS LEAD DOWN HERE. DO YOU WANT TO TAKE THEM? ⚠ 另有 `GOSUB 0x093E` 印出的一段 |
| `ECL5.DAX/0x33` | `0x09B7` | `subroutine` | — | UP |
| `ECL5.DAX/0x33` | `0x09BD` | `subroutine` | — | DOWN |
| `ECL5.DAX/0x33` | `0x09EA` | `matched` | `wizard-tower.shield-of-pain` | GUARDING THE STAIRS IS A BURLY DARK ELF IN PLATE AND SHIELD. HE RAISES HIS SHIELD AT YOUR APPROACH, REVEALING A SYMBOL OF PAIN. AS EVERYONE WRITHES IN AGONY, HE… |
| `ECL5.DAX/0x33` | `0x0AB4` | `matched` | `wizard-tower.identical-dark-elf` | AT THE BASE OF THIS SET OF STAIRS IS A DARK ELF NEARLY IDENTICAL TO THE ONE YOU FACED AT THE LAST STAIRS. |
| `ECL5.DAX/0x33` | `0x0B0D` | `matched` | `wizard-tower.pass-for-items` | 'YOU MAY PASS IF I CAN SELECT SOME OF YOUR ITEMS AND MONEY.' WILL YOU PAY? |
| `ECL5.DAX/0x33` | `0x0B5C` | `matched` | `wizard-tower.takes-items-walks-off` | HE TAKES SOME ITEMS, AND THEN WALKS OFF. |
| `ECL5.DAX/0x33` | `0x0B82` | `matched` | `wizard-tower.puff-of-smoke` | REFUSING TO BACK DOWN, YOU RUSH THE DARK ELF, AND HE DISAPPEARS IN A PUFF OF SMOKE -- ANOTHER ILLUSION. |
| `ECL5.DAX/0x33` | `0x0BF1` | `matched` | `wizard-tower.folded-paper` | ON THE FLOOR IS A FOLDED PIECE OF PAPER WITH THE TITLE,'AVOIDING TOWER TRAPS.' DO YOU PICK IT UP AND READ IT? |
| `ECL5.DAX/0x33` | `0x0C72` | `matched` | `wizard-tower.not-healthy-enough` | THAT ONE ISN'T HEALTHY ENOUGH. |
| `ECL5.DAX/0x33` | `0x0CE4` | `matched` | `wizard-tower.explosive-rune-note` | THE NOTE SAYS, '1. DO NOT READ EXPLOSIVE RUNES.' THE PAPER THEN EXPLODES. |
| `ECL5.DAX/0x33` | `0x0D76` | `matched` | `wizard-tower.avoid-explosive-rune` | YOU AVOID READING AN EXPLOSIVE RUNE AND DISCOVER A WAY PAST THE TRAP AT THE NEXT STAIRWAY. |
| `ECL5.DAX/0x33` | `0x0DE9` | `matched` | `wizard-tower.stairs-go-flat` | THE BOTTOM SET OF STAIRS GOES FLAT, PROPELLING THE PARTY INTO A SET OF UPTHRUST SPIKES. |
| `ECL5.DAX/0x33` | `0x0E40` | `matched` | `wizard-tower.hidden-button` | YOU FIND THE HIDDEN BUTTON MENTIONED IN THE NOTE AND BYPASS THE SPIKE TRAP. |
| `ECL5.DAX/0x33` | `0x0EEC` | `matched` | `wizard-tower.dracandros.final-challenge` | DRACANDROS IS DRAGGING A HEAVY BAG ACROSS THE COURTYARD. DROPPING THE BAG, HE SAYS, 'YOU HAVE DOGGED MY STEPS FOR TOO LONG. NOW YOU SHALL BE DESTROYED.' |
| `ECL5.DAX/0x33` | `0x0F8E` | `matched` | `wizard-tower.helm-of-dragons` | AMONG HIS TREASURES YOU FIND THE HELM OF DRAGONS. |
| `ECL5.DAX/0x33` | `0x0FE6` | `matched` | `wizard-tower.laboratory.pool` | YOU HAVE FOUND A MYSTICAL LABORATORY. IN THE CENTER IS A MURKY, SHALLOW POOL WHICH SMELLS OF ROTTING VEGETATION. DO YOU GO FORWARD? |
| `ECL5.DAX/0x33` | `0x105C` | `matched` | `wizard-tower.laboratory.pool-depth` | THE POOL LOOKS TO BE SEVERAL FEET DEEP THOUGH YOU CANNOT SEE TO THE BOTTOM. DO YOU REACH INTO THE POOL? |
| `ECL5.DAX/0x33` | `0x10B9` | `matched` | `wizard-tower.laboratory.ellipsoid` | ROOTING AROUND IN THE MUCK, YOU FIND A LEATHERY ELLIPSOID ABOUT THE SIZE OF A HEAD. IT IS WARM AND PULSES FAINTLY. DO YOU TAKE IT? |
| `ECL5.DAX/0x33` | `0x1148` | `matched` | `wizard-tower.sphere-trial.sign` | A SIGN OVER THIS EERIE BLACK DOOR SAYS, 'TRIAL OF THE SPHERE. ONE CHALLENGER ONLY.' DOES ANYONE WISH TO ENTER? |
| `ECL5.DAX/0x33` | `0x11D6` | `matched` | `wizard-tower.sphere-trial.cannot-enter` | THAT ONE CANNOT ENTER. |
| `ECL5.DAX/0x33` | `0x121B` | `matched` | `wizard-tower.sphere-trial.chamber` | YOU PASS INTO A DIM ROOM OCCUPIED BY A YOUNG MAGE DRESSED IN RED AND HOLDING A SILVERY ROD. BETWEEN THE TWO OF YOU IS A FLOATING DEAD BLACK SPHERE. |
| `ECL5.DAX/0x33` | `0x1298` | `matched` | `wizard-tower.sphere-trial.rules` | THE WIZARD INTONES, 'THIS IS A SPHERE OF ANNIHILATION. ITS TOUCH MEANS UTTER DESTRUCTION. ONLY A MAGE CAN CONTROL ITS FLIGHT. YOU MAY SURRENDER AT ANY TIME.' |
| `ECL5.DAX/0x33` | `0x147B` | `variable-insert` | — | THE SPHERE MOVES TOWARD THE OPPOSING MAGE. . IT IS NOW ONLY FEET FROM . ⚠ 另有 `GOSUB 0x1490` 印出的一段 |
| `ECL5.DAX/0x33` | `0x150A` | `matched` | `wizard-tower.dark-elf-party` | THE SPHERE REACHES , WHO IS SUCKED INTO THE SPHERE AND GONE FOREVER. |
| `ECL5.DAX/0x33` | `0x1565` | `matched` | `wizard-tower.sphere-kills-wizard` | AS THE SPHERE REACHES THE RED WIZARD, HE RAISES HIS ROD IN PANIC. THE TWO MEET IN A TREMENDOUS EXPLOSION THAT ROCKS THE BUILDING, LEAVING NOTHING OF THE SPHERE … |
| `ECL5.DAX/0x33` | `0x15F4` | `matched` | `wizard-tower.chests-valuables` | THE CHESTS LINING THE WALLS CONTAIN MANY VALUABLES. |
| `ECL5.DAX/0x33` | `0x1642` | `variable-insert` | — | IT IS NOW ONLY FEET FROM THE RED WIZARD. |
| `ECL5.DAX/0x33` | `0x1697` | `matched` | `wizard-tower.spoils-for-the-victor` | THE SPHERE HALTS BETWEEN YOU. THE RED WIZARD THEN HAS YOU DROP SOME ITEMS AND MONEY. HE SMILES AND SAYS, 'SPOILS FOR THE VICTOR.' SUDDENLY YOU ARE OUTSIDE. |
| `ECL5.DAX/0x33` | `0x1782` | `matched` | `wizard-tower.dark-elf-owlbear-guard` | THIS AREA IS GUARDED BY DARK ELVES WITH OWL BEARS. |
| `ECL5.DAX/0x33` | `0x1800` | `matched` | `wizard-tower.dragon-pen` | YOU ARE ASSAULTED BY A REPTILIAN STENCH. YOU SEE MASSIVE CHAINS ATTACHED TO THE WALLS WITH HUGE, BUT EMPTY COLLARS. ACROSS THE ROOM YOU HEAR A DEEP GROWL. |
| `ECL5.DAX/0x33` | `0x18A1` | `matched` | `wizard-tower.bedroom` | YOU HAVE ENTERED A VERY ELEGANT BEDROOM, RED SILK TAPESTRIES ADORN THE WALLS. THE ROOM IS DISARRANGED AS IF SOMEONE HAD QUICKLY PACKED. DO YOU WANT TO TAKE THE … |
| `ECL5.DAX/0x33` | `0x1968` | `matched` | `wizard-tower.library` | THE WALLS OF THIS ROOM ARE LINED WITH BOOKS ON MANY SUBJECTS. NONE RADIATE MAGIC. |
| `ECL5.DAX/0x33` | `0x19C7` | `matched` | `wizard-tower.scroll-room` | THE WALLS ARE COVERED WITH RACKS FOR SCROLL CASES. THE SCROLLS LOOK TO HAVE BEEN RECENTLY REMOVED. |
| `ECL5.DAX/0x33` | `0x1A8C` | `matched` | `wizard-tower.sucked-into-sphere` | AN EFREET LEADS A BAND OF DARK ELVES AGAINST YOU. |
| `ECL5.DAX/0x33` | `0x1ADA` | `matched` | `wizard-tower.salamanders` | A SULFEROUS SMELL ASSAULTS YOU AS YOU MEET SALAMANDERS. |
| `ECL5.DAX/0x33` | `0x1B2B` | `matched` | `wizard-tower.dark-elf-party` | A PARTY OF DARK ELVES APPROACHES. |
| `ECL5.DAX/0x33` | `0x1B72` | `matched` | `wizard-tower.armored-contingent` | AN ARMORED CONTINGENT OF DARK ELVES CHARGES. |
| `ECL5.DAX/0x35` | `0x0072` | `matched` | `dark-elf-caves.torn-up-party` | YOU COME UPON A TORN UP PARTY COMING OUT. ONE SAYS, 'IT'S PRETTY FIERCE DOWN THERE. THE DEEPER YOU GO THE NASTIER THE CREATURES. LUCKILY, IF YOU SEARCH YOU CAN … |
| `ECL5.DAX/0x35` | `0x0133` | `matched` | `dark-elf-caves.dank-cavern` | YOU HAVE ENTERED A DANK CAVERN. |
| `ECL5.DAX/0x35` | `0x017F` | `matched` | `dark-elf-caves.dark-pit` | YOU HAVE ENTERED A DARK PIT. |
| `ECL5.DAX/0x35` | `0x01C8` | `matched` | `dark-elf-caves.twisting-cavern` | YOU HAVE ENTERED A TWISTING CAVERN. |
| `ECL5.DAX/0x35` | `0x04B2` | `matched` | `yulash.false-door` | YOU HAVE FOUND A FALSE DOOR. |
| `ECL5.DAX/0x35` | `0x04EE` | `matched` | `yulash.about-to-leave` | YOU ARE ABOUT TO LEAVE. DO YOU WANT TO? |
| `ECL5.DAX/0x35` | `0x0541` | `matched` | `dark-elf-caves.items-fade` | YOUR DARK ELF ITEMS FADE AWAY IN THE SUNLIGHT. |
| `ECL5.DAX/0x35` | `0x078B` | `matched` | `dark-elf-caves.fresh-air-exit` | YOU FEEL A BREATH OF FRESH AIR. DO YOU WANT TO EXIT? |
| `ECL5.DAX/0x35` | `0x08FF` | `matched` | `dark-elf-caves.monsters-spotted` | YOU SPOT MONSTERS. |
| `ECL5.DAX/0x35` | `0x0991` | `matched` | `dark-elf-caves.tarsus-offer` | 'GREETINGS TRAVELLERS, AND WELCOME TO MY HUMBLE ABODE. I AM TARSUS AND, FOR A TITHE, I CAN TRANSPORT YOU BACK TO THE SURFACE. INTERESTED?' |
| `ECL5.DAX/0x35` | `0x0A12` | `matched` | `dark-elf-caves.tarsus-accepted` | 'YOU WILL NOT REGRET THIS DESCISION.' |
| `ECL5.DAX/0x35` | `0x0A3B` | `matched` | `dark-elf-caves.tarsus-declined` | 'I TRAVEL RARELY, SO YOU SHOULD FIND ME HERE IF YOU CHANGE YOUR MIND.' |
| `ECL5.DAX/0x35` | `0x0AC5` | `matched` | `dark-elf-caves.wounded-man` | YOU FIND A WOUNDED MAN. HE WHISPERS, 'THE DROW HAVE KIDNAPPED THE MAGISTRATE'S DAUGHTER. THEY ARE TAKING HER BELOW. SAVE HER.' HE EXPIRES. |
| `ECL5.DAX/0x35` | `0x0B4C` | `matched` | `dark-elf-caves.rear-guard` | THE DROW TURN TOWARD YOU, 'SO MORE SURFACE SCUM INVADE OUR HALLS. AS REAR GUARD, WE MUST DEFEAT THEM.' |
| `ECL5.DAX/0x35` | `0x0BA4` | `matched` | `dexam.attack` | THEY ATTACK. |
| `ECL5.DAX/0x35` | `0x0BE1` | `matched` | `dexam.attack` | YOU SUDDENLY COME UPON THE MAIN FORCE OF DROW, WHO ARE DRAGGING ALONG A YOUNG WOMAN. THEY ATTACK IMMEDIATELY. |
| `ECL5.DAX/0x35` | `0x0C7B` | `matched` | `dark-elf-caves.drow-left-with-girl` | THE DROW HAVE LEFT WITH THE GIRL. |
| `ECL5.DAX/0x35` | `0x0C9E` | `matched` | `dark-elf-caves.girl-fled` | DURING THE FIGHT, THE GIRL FLED IN PANIC. SHE IS NOWHERE TO BE SEEN. |
| `ECL5.DAX/0x35` | `0x0CE4` | `matched` | `dark-elf-caves.maiden-and-dragons` | A LOVELY MAIDEN IS BEING MENACED BY FIERCE BLACK DRAGONS. DO YOU WANT TO SAVE HER? |
| `ECL5.DAX/0x35` | `0x0D4B` | `matched` | `dark-elf-caves.girl-asks-shadowdale` | THE GIRL ESCAPED, BUT THE DRAGONS RUSH AFTER HER. |
| `ECL5.DAX/0x35` | `0x0D84` | `matched` | `dark-elf-caves.girl-asks-shadowdale` | THE GIRL SMILES DEMURELY, 'THANK YOU KIND WARRIORS. YOU HAVE SAVED ME. RETURN ME TO SHADOWDALE, AND MY FATHER WILL PAY YOU HANDSOMELY.' WILL YOU TAKE HER BACK? |
| `ECL5.DAX/0x35` | `0x0E20` | `matched` | `dark-elf-caves.surface-thanks` | YOU RETRACE YOUR PATH AND REACH THE SURFACE. PEOPLE GATHER AROUND AND THANK YOU. ONE OLDER MAN PUSHES FORWARD. 'YOU ARE VERY BRAVE WARRIORS. HERE TAKE ALL I CAN… |
| `ECL5.DAX/0x35` | `0x0EC8` | `matched` | `dark-elf-caves.try-on-my-own` | ' I WILL TRY ON MY OWN. STILL THANK YOU FOR YOUR HELP. |
| `ECL5.DAX/0x35` | `0x0EFF` | `matched` | `dark-elf-caves.woman-devoured-only` | THE WOMAN IS DEVOURED, THEN, THE DRAGONS TURN TO YOU. YOU FEEL THAT THE GODS WILL NOT FAVOR YOU IN THIS BATTLE. |
| `ECL5.DAX/0x35` | `0x0FA2` | `matched` | `dark-elf-caves.pursuit-note` | YOU COME UPON A NOTE WHICH READS, 'HE ESCAPED INTO THE CAVES AND STILL HAS THE ITEM. WE ARE IN PURSUIT.' |
| `ECL5.DAX/0x35` | `0x0FFF` | `matched` | `dark-elf-caves.ashen-remains` | THE ASHEN REMAINS OF ANOTHER PARTY OF ADVENTURERS LIE HERE. A FEW ITEMS HAVE SURVIVED. |
| `ECL5.DAX/0x35` | `0x1072` | `matched` | `dark-elf-caves.young-man-calls-protector` | YOU SPOT A YOUNG MAN CARRYING A LARGE GEM. WHEN HE NOTES YOUR PRESENCE, HE TURNS. 'YOU ARE BECOMING TIRESOME. I CALL THE PROTECTOR!' A SHIMMER APPEARS BETWEEN Y… |
| `ECL5.DAX/0x35` | `0x1107` | `matched` | `dark-elf-caves.protector-arrives` | A HIDEOUS SKELETAL FORM APPEARS BETWEEN YOU. FROM BEHIND YOU HEAR A SCREAM, 'THE PROTECTOR ARRIVES. SLAY THEM, SLAY, SLAY, SLAY!' THE MONSTERS STRIKES. |
| `ECL5.DAX/0x35` | `0x1197` | `matched` | `dark-elf-caves.body-fades` | THE BODY FADES AWAY. THE MAN HAS RUN OFF. |
| `ECL5.DAX/0x35` | `0x11D1` | `matched` | `dark-elf-caves.gem-dashed` | THE MAN TURNS, 'YOU SHALL NEVER HAVE IT! SINCE YOU INSIST ON PURSUIT, THEN DIE!' HE DASHES THE GEM TO THE GROUND. A HUGE FIRE ROARS UP CONSUMING HIM. SHAPES FOR… |
| `ECL5.DAX/0x35` | `0x1273` | `matched` | `dark-elf-caves.creatures-hostile` | THE CREATURES SEEM HOSTILE. |
| `ECL5.DAX/0x35` | `0x12A7` | `matched` | `dark-elf-caves.gem-remains-valuable` | YOU FIND THAT THE REMAINS OF THE GEM ARE STILL VALUABLE. |
| `ECL5.DAX/0x35` | `0x1312` | `matched` | `dark-elf-caves.owl-bear-drawings` | YOU COME UPON SOME CRUDE DRAWINGS SHOWING OWL BEARS BEING CREATED BY THE HAND OF A DROW. THE ARTWORK LOOKS FAIRLY RECENT. |
| `ECL5.DAX/0x35` | `0x137C` | `matched` | `dark-elf-caves.bloody-altar` | YOU FIND A MAKESHIFT ALTAR SPATTERED WITH FRESH BLOOD. BEARLIKE FOOTPRINTS CIRCLE THE ALTAR. |
| `ECL5.DAX/0x35` | `0x13DF` | `matched` | `dark-elf-caves.owl-bear-circle` | YOU HEAR A LOW, GROWLED CHANT. AHEAD OF YOU IS A CIRCLE OF OWL BEARS PRANCING AROUND A SINGLE DROW. AT YOUR APPROACH, THEY SHUFFLE TO A HALT AND STARE AT YOU. T… |
| `ECL5.DAX/0x35` | `0x1491` | `matched` | `dark-elf-caves.owl-bears-charge` | THE OWL BEARS CHARGE. |
| `ECL6.DAX/0x40` | `0x0077` | `matched` | `myth-drannor.helm-north` | THE HELM OF DRAGONS REPORTS TYRANTHRAXUS IS TO THE NORTH. |
| `ECL6.DAX/0x40` | `0x00E1` | `matched` | `myth-drannor.more-ruins` | YOU ARE HEADING TOWARD MORE RUINS. WILL YOU TRAVEL VIA A PATH OR THROUGH THE WOODS? |
| `ECL6.DAX/0x40` | `0x0187` | `matched` | `myth-drannor.outer.heading-wilderness` | YOU ARE HEADING TOWARD THE WILDERNESS. DO YOU CONTINUE? |
| `ECL6.DAX/0x40` | `0x01EC` | `matched` | `myth-drannor.outer.rubble-blocks-entrance` | RUBBLE HAS BLOCKED THE ENTRANCE. |
| `ECL6.DAX/0x40` | `0x03E6` | `matched` | `myth-drannor.outer.thri-kreen-master` | A THRI-KREEN STEPS FORWARD ASKING, 'WHO IS YOUR MASTER?' WHAT DO YOU SAY? |
| `ECL6.DAX/0x40` | `0x043F` | `matched` | `myth-drannor.outer.no-quarrel-with-him` | 'WE HAVE NO QUARREL WITH HIM.' THEY LEAVE. |
| `ECL6.DAX/0x40` | `0x046A` | `matched` | `myth-drannor.outer.that-is-unknown` | THAT IS UNKNOWN. DIE! |
| `ECL6.DAX/0x40` | `0x04F7` | `matched` | `myth-drannor.outer.group-of-knights` | YOU SEE A GROUP OF KNIGHTS. |
| `ECL6.DAX/0x40` | `0x0573` | `matched` | `myth-drannor.outer.harper-diversion` | 'YOUR CAUSE IS JUST. WHEN YOU ARE IN THE TEMPLE, WE WILL CAUSE A DIVERSION TO DRAW OFF SOME TROOPS. GOOD LUCK!' |
| `ECL6.DAX/0x40` | `0x05DF` | `matched` | `myth-drannor.outer.harper-diversion` | 'WE HAVE HAD REPORTS OF YOU FROM TILVERTON, WHERE YOU HELPED A BROTHER. |
| `ECL6.DAX/0x40` | `0x0620` | `matched` | `myth-drannor.outer.you-are-offensive` | 'YOU ARE OFFENSIVE. WE LEAVE YOU TO YOUR FATE.' THEY LEAVE. |
| `ECL6.DAX/0x40` | `0x066C` | `matched` | `myth-drannor.elf-spirit.greeting` | AN ELFISH SPIRIT APPEARS AND GREETS YOU. WHAT DO YOU DO ? ⚠ 另有 `GOSUB 0x068D` 印出的一段 |
| `ECL6.DAX/0x40` | `0x06BC` | `matched` | `myth-drannor.spirit-disappears` | THE SPIRIT DISAPPEARS. |
| `ECL6.DAX/0x40` | `0x06D8` | `matched` | `myth-drannor.elf-spirit.flee` | 'SO, YOU ARE SHEEP! THEN YOU SHALL FEED ME.' THE SPIRIT FADES. |
| `ECL6.DAX/0x40` | `0x072B` | `matched` | `myth-drannor.elf-spirit.journal-25` | THE SPIRIT TALKS OF THE GLEN AND YOU RECORD IT IN JOURNAL ENTRY 25. THEN, THE SPIRIT FADES. ⚠ 另有 `GOSUB 0x0744` 印出的一段 |
| `ECL6.DAX/0x40` | `0x0770` | `matched` | `myth-drannor.red-web` | A RED WEB STRETCHES ACROSS THE PASSAGE, GLOWING DULLY. WHAT DO YOU DO ? ⚠ 另有 `GOSUB 0x079F` 印出的一段 |
| `ECL6.DAX/0x40` | `0x07E3` | `matched` | `myth-drannor.red-web.stuck` | YOU FIND YOURSELF STUCK FAST. |
| `ECL6.DAX/0x40` | `0x0821` | `matched` | `myth-drannor.red-web.rakshasa` | THE SPIRIT APPEARS LAUGHING. IT FADES, REVEALING A RAKSHASA. |
| `ECL6.DAX/0x40` | `0x0895` | `matched` | `myth-drannor.red-web.word` | WHAT WORD DO YOU SAY? |
| `ECL6.DAX/0x40` | `0x08AE` | `matched` | `myth-drannor.red-web.brighter` | THE WEB GLOWS MORE BRIGHTLY. |
| `ECL6.DAX/0x40` | `0x08CE` | `matched` | `myth-drannor.red-web.hack` | AS YOU STRIKE, THE GLOW FADES FROM THE WEBS, REVEALING SEVERAL WIRE SNARES AS WELL. IN THE DISTANCE YOU HEAR CURSING, THEN RUNNING FEET. ⚠ 另有 `GOSUB 0x08DD` 印出的一段 |
| `ECL6.DAX/0x40` | `0x090F` | `matched` | `myth-drannor.red-web.spiders` | SOME SPIDERS INVESTIGATE THE NOISE. |
| `ECL6.DAX/0x40` | `0x0944` | `matched` | `myth-drannor.red-web.free` | THE GLOW FADES FROM THE WEBS, REVEALING SEVERAL WIRE SNARES AS WELL. YOU EVENTUALLY FREE YOURSELF. |
| `ECL6.DAX/0x40` | `0x09A2` | `matched` | `myth-drannor.daemir.forgiveness-offer` | A SPIRIT APPEARS BEFORE YOU. 'I AM THE SPIRIT OF PRINCESS DAEMIR. YOU HAVE DESPOILED THE GLEN -- PERHAPS FROM IGNORANCE. IF YOU KNEEL NOW, YOU MAY ACCEPT MY FOR… |
| `ECL6.DAX/0x40` | `0x0A4A` | `matched` | `myth-drannor.daemir.forgiven` | 'YOU ARE FORGIVEN . I AM AMAZED THAT THE SWANMAYS ARE STILL ACTIVE IN THE REALMS. I RECALL MY TENURE WITH THEM. YOU ARE LUCKY TO HAVE IN YOUR PARTY .' SHE FADES… |
| `ECL6.DAX/0x40` | `0x0ABD` | `matched` | `myth-drannor.daemir.blessing` | 'GO FORTH WITH MY BLESSING . I AM AMAZED THAT THE SWANMAYS ARE STILL ACTIVE IN THE REALMS. I RECALL MY TENURE WITH THEM. YOU ARE LUCKY TO HAVE IN YOUR PARTY .' … |
| `ECL6.DAX/0x40` | `0x0AE7` | `matched` | `myth-drannor.daemir.curse` | 'YOU PRESUME TOO MUCH! AS LONG AS YOU ARE IN MYTH DRANNOR, YOUR WEAPONS WILL TWIST IN YOUR HANDS .' SHE FADES AWAY. |
| `ECL6.DAX/0x40` | `0x0C17` | `matched` | `myth-drannor.grave.thri-kreen` | A THRI-KREEN IS EXCAVATING A GRAVE HERE. AT YOUR APPROACH IT TURNS AND ATTACKS. |
| `ECL6.DAX/0x40` | `0x0C8F` | `matched` | `myth-drannor.grave.skeleton` | YOU SEE A PARTIALLY EXCAVATED ELF SKELETON. JEWELRY GLITTERS ON ITS WRIST. WHAT WILL YOU DO? |
| `ECL6.DAX/0x40` | `0x0DC8` | `matched` | `myth-drannor.red-plume.journal` | THE RED PLUME TELLS YOU HIS TALE AND YOU RECORD IT IN JOURNAL ENTRY 33. WHAT DO YOU DO ? ⚠ 另有 `GOSUB 0x0DE4`／`0x0DEE` 印出的一段 |
| `ECL6.DAX/0x40` | `0x0E2C` | `matched` | `myth-drannor.red-plume.warning` | FOLLOW ME. THANK YOU FOR YOUR MAGNANAMOUS GESTURE. |
| `ECL6.DAX/0x40` | `0x0E94` | `matched` | `myth-drannor.red-plume.no-payment` | WITHIN AN UNEARTHED GRAVE ARE GLOWING SPIDERS WHO SWARM OUT TO PROTECT THEIR EGGS. THE RED PLUME STEPS AWAY AND FIRES -- AT YOU. HIS SHAPE CHANGES. |
| `ECL6.DAX/0x40` | `0x0F5C` | `matched` | `myth-drannor.red-plume.warning` | A WHISPERY VOICE CRIES OUT,'BEWARE, A TRAP.' THE RED PLUME SNARLS AND CUTS OFF THE VOICE WITH A GESTURE. |
| `ECL6.DAX/0x40` | `0x0FD0` | `matched` | `myth-drannor.red-plume.refusal` | 'THEN I SHALL TRY MYSELF.' HE GOES THROUGH THE GATE. SUDDENLY, A SCREAM PIERCES THE AIR. DO YOU GO TO INVESTIGATE? |
| `ECL6.DAX/0x40` | `0x1059` | `matched` | `myth-drannor.court.entry` | NEAR THE ENTRANCE TO THIS BUILDING IS A CRUSHED THRI-KREEN. IN THE DOORWAY IS A GHOSTLY SHAPE. |
| `ECL6.DAX/0x40` | `0x10AD` | `matched` | `myth-drannor.court.enter` | DO YOU WANT TO ENTER THE BUILDING? |
| `ECL6.DAX/0x40` | `0x10DE` | `matched` | `myth-drannor.outer.spirit-raises-arms` | AS YOU MOVE FORWARD, THE SPIRIT RAISES ITS ARMS. ROCKS AND TOMBSTONE WHIRL AROUND YOU. |
| `ECL6.DAX/0x40` | `0x1134` | `matched` | `myth-drannor.outer.fades-away` | AS IF EXHAUSTED, IT FADES AWAY. |
| `ECL6.DAX/0x40` | `0x115C` | `matched` | `myth-drannor.court.welcome` | THE SPIRIT SPEAKS, 'WELCOME WARRIORS. ENTER AND MEET OUR QUEEN.' |
| `ECL6.DAX/0x40` | `0x11B8` | `matched` | `myth-drannor.court.armor` | TWO SUITS OF ARMOR FLANK THIS STAIRWAY, RADIATING FAINT MAGIC. ONE SPEAKS,' YOUR KIND MAY NOT SEE THE QUEEN.' |
| `ECL6.DAX/0x40` | `0x1276` | `matched` | `myth-drannor.outer.spears-crossed` | THE SUITS CROSS THEIR SPEARS ACROSS THE ENTRANCE. DO YOU CONTINUE? |
| `ECL6.DAX/0x40` | `0x12B0` | `matched` | `myth-drannor.court.armor-bows` | THE ARMOR SEEMS TO BOW AS YOU PASS. |
| `ECL6.DAX/0x40` | `0x12F1` | `matched` | `myth-drannor.court.armor-crumbles` | IN RESPONSE, THE ARMOR CRUMBLES INTO RUSTY FLAKES. |
| `ECL6.DAX/0x40` | `0x1355` | `matched` | `myth-drannor.court.hostile` | AS YOU APPROACH THE STAIRS A VOICE CRIES OUT,' DESPOILERS SHALL FIGHT DESPOILERS.' |
| `ECL6.DAX/0x40` | `0x13F9` | `matched` | `myth-drannor.court.reward` | A SPIRIT APPEARS BEFORE YOU.' THIS GLEN HAS BEEN CRUSHED BY YOUR KIND. IF YOU WILL LEAVE, YOU MAY HAVE THE REMNANTS OF OUR TREASURE. WILL YOU?' |
| `ECL6.DAX/0x40` | `0x14A5` | `matched` | `myth-drannor.court.refusal` | 'THEN LIE WITH US!' THE TOWER BEGINS TO COLLAPSE AROUND YOU. YOU RUSH OUTSIDE. |
| `ECL6.DAX/0x40` | `0x14F5` | `matched` | `myth-drannor.court.collapse` | THE TOWER FALLS INTO A PILE OF RUBBLE. |
| `ECL6.DAX/0x40` | `0x158A` | `matched` | `myth-drannor.court.farewell` | 'ALAS, MY TIME IS SHORT HERE. MY BEST WISHES TO YOU.' SHE DISAPPEARS. |
| `ECL6.DAX/0x40` | `0x163C` | `matched` | `myth-drannor.clan-figure.journal` | THE FIGURE BREAKS INTO RAPID SPEECH AND YOU RECORD IT IN JOURNAL ENTRY 56. 'HURRY ON, HURRY, HURRY!' HE LEAVES. ⚠ 另有 `GOSUB 0x165A` 印出的一段 |
| `ECL6.DAX/0x40` | `0x1696` | `matched` | `myth-drannor.thri-kreen.entrance` | A PARTY OF THRI-KREEN BAR YOUR ENTRANCE. |
| `ECL6.DAX/0x40` | `0x16D6` | `matched` | `myth-drannor.thri-kreen.guards` | GUARDS HERE PREPARE FOR COMBAT. |
| `ECL6.DAX/0x40` | `0x171B` | `matched` | `myth-drannor.thri-kreen.bivouac` | THE THRI-KREEN HAVE BIVOUACKED HERE. THEY PREPARE TO MAKE A STAND. |
| `ECL6.DAX/0x40` | `0x177F` | `matched` | `myth-drannor.thri-kreen.reinforcements` | OTHER THRI-KREEN RESPOND TO THE NOISE. |
| `ECL6.DAX/0x40` | `0x17CB` | `matched` | `myth-drannor.thri-kreen.stragglers` | A FEW MORE STRAGGLE IN. |
| `ECL6.DAX/0x40` | `0x1801` | `matched` | `myth-drannor.thri-kreen.valuables` | YOU GATHER UP SOME VALUABLES. |
| `ECL6.DAX/0x40` | `0x183F` | `matched` | `myth-drannor.spider-mausoleum` | WEBS FESTOON THIS MAUSOLEUM. THE WEBS ARE INHABITED. |
| `ECL6.DAX/0x40` | `0x188C` | `matched` | `myth-drannor.spider-funnel` | YOU SEE A FUNNEL OF WEBS. ⚠ 另有 `GOSUB 0x18A2` 印出的一段 |
| `ECL6.DAX/0x40` | `0x18B7` | `matched` | `myth-drannor.spider-warning` | A WHISPERY VOICE CALLS OUT,'THE SPIDERS WILL GUARD THEIR NEST FIERCELY.' DO YOU CONTINUE? |
| `ECL6.DAX/0x40` | `0x1908` | `matched` | `myth-drannor.spider-eggs` | SPIDERS LEAP FORTH. YOU CAN SEE SOME EGGS BEHIND THEM. |
| `ECL6.DAX/0x40` | `0x196D` | `matched` | `myth-drannor.phase-spider-wall` | AS YOU ENTER, SPIDERS COME OUT OF THE SOLID WALLS. |
| `ECL6.DAX/0x40` | `0x19C0` | `matched` | `myth-drannor.phase-spider-glowing` | GLOWING SPIDERS SKITTER FORWARD AT YOUR APPROACH. |
| `ECL6.DAX/0x40` | `0x1A12` | `matched` | `myth-drannor.phase-spider-bones` | SPIDERS HAVE GATHERED A PILE OF BONES HERE AND DEFEND IT. |
| `ECL6.DAX/0x40` | `0x1BA3` | `matched` | `myth-drannor.bones.prompt` | WHAT DO YOU DO WITH THE BONES? |
| `ECL6.DAX/0x40` | `0x1BCC` | `subroutine` | — | DO YOU CONTINUE? |
| `ECL6.DAX/0x42` | `0x0051` | `matched` | `myth-drannor.helm-north` | THE HELM OF DRAGONS REPORTS TYRANTHRAXUS TO THE NORTH. |
| `ECL6.DAX/0x42` | `0x00BD` | `matched` | `myth-drannor.outer.ruined-temple` | YOU ARE HEADING TOWARD A LARGE RUINED TEMPLE. DO YOU CONTINUE? |
| `ECL6.DAX/0x42` | `0x010F` | `matched` | `myth-drannor.outer.sewer-warning` | A DREAM-LIKE VOICE IN YOUR HEAD SAYS, 'GREAT DANGER LIES BEFORE YOU. BE FULLY PREPARED!' DO YOU STILL CONTINUE? |
| `ECL6.DAX/0x42` | `0x0177` | `matched` | `myth-drannor.outer.graveyard` | YOU ARE HEADING INTO THE GRAVEYARD. DO YOU CONTINUE? |
| `ECL6.DAX/0x42` | `0x03C0` | `matched` | `myth-drannor.outer.yell-a-word` | YOU HAVE TIME TO YELL BUT A SINGLE WORD BEFORE THE BEASTS ARE UPON YOU. (TYPE A SINGLE WORD) |
| `ECL6.DAX/0x42` | `0x0427` | `matched` | `myth-drannor.outer.monsters-move-off` | THE MONSTERS STOP THEIR ATTACK, GROWL AND MOVE OFF. |
| `ECL6.DAX/0x42` | `0x0513` | `matched` | `myth-drannor.outer.do-not-find-you-again` | 'DO NOT LET US FIND YOU AGAIN.' THEY LEAVE. |
| `ECL6.DAX/0x42` | `0x053F` | `matched` | `myth-drannor.outer.dislike-your-response` | THE MONSTERS DON'T LIKE YOUR RESPONSE. |
| `ECL6.DAX/0x42` | `0x05CE` | `matched` | `myth-drannor.outer.tirsheya.tale` | THE RAKSHASA REPLIES, ' I AM TIRSHEYA.' HE GOES ON TO TELL HIS TALE, AND YOU RECORD IT IN JOURNAL ENTRY 5. WILL YOU GO WITH HIM? |
| `ECL6.DAX/0x42` | `0x05FE` | `matched` | `myth-drannor.outer.tirsheya.greeting` | A RAKSHASA WITH MATTED FUR AND A DOUR EXPRESSION COMES AROUND THE CORNER. HE MAKES A GESTURE OF PEACE. WHAT DO YOU DO? |
| `ECL6.DAX/0x42` | `0x06D1` | `matched` | `myth-drannor.outer.tirsheya.guards` | 'THEN COME WITH ME.' |
| `ECL6.DAX/0x42` | `0x06FB` | `matched` | `myth-drannor.outer.tirsheya.guards` | THE ENTRANCE IS GUARDED. DO YOU ATTACK? |
| `ECL6.DAX/0x42` | `0x072F` | `matched` | `myth-drannor.outer.tirsheya-sprints-away` | AS YOU PULL YOUR WEAPONS, TIRSHEYA SHRUGS HIS SHOULDERS, AND THEN SPRINTS AWAY. |
| `ECL6.DAX/0x42` | `0x0779` | `matched` | `myth-drannor.outer.tirsheya-dejected` | LOOKING DEJECTED, TIRSHEYA TURNS AND LEAVES. |
| `ECL6.DAX/0x42` | `0x07A5` | `matched` | `myth-drannor.outer.guards-notice` | BEFORE YOU CAN PULL BACK, THE GUARDS NOTICE YOU. DO YOU RUN? |
| `ECL6.DAX/0x42` | `0x0801` | `matched` | `myth-drannor.outer.tirsheya.beyrha-arrives` | THE SOUNDS OF BATTLE HAVE BROUGHT ANOTHER RAKSHASA AND RETINUE. HE CALLS OUT, 'SO TIRSHEYA, STOOPING TO ROBBING THE CLAN?' |
| `ECL6.DAX/0x42` | `0x086B` | `matched` | `myth-drannor.outer.tirsheya.ultimatum` | 'KNOW HUMANS, THAT IF YOU BRING ME HIS HEAD, YOU WILL BE FORGIVEN. OTHERWISE YOU SHALL BE TONIGHT'S MAIN DISH. THIS IS THE PROMISE OF BEYRHA.' WHO DO YOU ATTACK… |
| `ECL6.DAX/0x42` | `0x0916` | `matched` | `myth-drannor.outer.betrayal-cost` | TIRSHEYA SNARLS,'YOUR BETRAYAL WILL NOT BE WITHOUT COST.' HE ATTACKS. |
| `ECL6.DAX/0x42` | `0x096C` | `matched` | `myth-drannor.outer.beyrha-you-may-leave` | BEYRHA SPEAKS, 'YOU HAVE DONE WELL. YOU MAY LEAVE.' DO YOU LEAVE? |
| `ECL6.DAX/0x42` | `0x09D3` | `matched` | `myth-drannor.outer.tirsheya.attack-beyrha` | 'YOU ARE FOOLS, BUT WE WILL THANK YOU AT DINNER.' |
| `ECL6.DAX/0x42` | `0x0AE3` | `matched` | `myth-drannor.outer.proves-his-cheating` | TIRSHEYA RUNS INTO THE STOREHOUSE. HE RUNS OUT A MOMENT LATER WITH SOME PAPERS. 'THIS PROVES HIS CHEATING. TAKE WHAT YOU WISH FROM THE STOREHOUSE. GOOD DAY.' |
| `ECL6.DAX/0x42` | `0x0B67` | `matched` | `myth-drannor.outer.moves-away-quickly` | HE MOVES QUICKLY AWAY BEFORE YOU CAN STOP HIM. |
| `ECL6.DAX/0x42` | `0x0B95` | `matched` | `myth-drannor.outer.guards-cant-decide` | AS YOU FLEE, TIRSHEYA RUNS IN A DIFFERENT DIRECTION. THE GUARDS CAN'T DECIDE WHO TO CHASE, SO YOU GET AWAY. |
| `ECL6.DAX/0x42` | `0x0C03` | `matched` | `myth-drannor.outer.storehouse.guards` | SOME MARGOYLES AND HELL HOUNDS GUARD THE ENTRANCE TO THIS BUILDING. THEY PREPARE FOR COMBAT . DO YOU FLEE? |
| `ECL6.DAX/0x42` | `0x0C80` | `matched` | `myth-drannor.outer.storehouse.supplies` | PILED WITHIN THIS BUILDING IS A LARGE ARRAY OF FOODSTUFFS, CLOTHING AND TRINKETS. EVENTUALLY YOU FIND A FEW VALUABLES. |
| `ECL6.DAX/0x42` | `0x0D15` | `matched` | `myth-drannor.outer.fugitive.intro` | AHEAD, YOU SEE A MAN RUNNING, NEAR EXHUASTION, CHASED  BY A PACK OF HELL HOUNDS. THE MAN STUMBLES AND COLLAPSES. DO YOU GO TO THE RESCUE? |
| `ECL6.DAX/0x42` | `0x0DB7` | `matched` | `myth-drannor.outer.fugitive.clue` | THE MAN THANKS YOU AND WHISPERS, 'I MANAGED TO HIDE IT BEFORE I WAS CAPTURED... RUINED BUILDING... NORTHEAST. YOUR REWARD.' HIS BREATHING STOPS. |
| `ECL6.DAX/0x42` | `0x0E41` | `matched` | `myth-drannor.outer.fugitive.killed` | THE HOUNDS LEAP ON HIM AND TEAR HIM TO SHREDS. THEY SOON COMPLETE THEIR WORK. DO YOU ATTACK? |
| `ECL6.DAX/0x42` | `0x0E9F` | `matched` | `myth-drannor.outer.fugitive.hounds-leave` | THE HOUNDS LEAVE THE REMAINS. |
| `ECL6.DAX/0x42` | `0x0ECB` | `matched` | `myth-drannor.outer.fugitive.remains` | THE HOUNDS LEFT LITTLE THAT IS RECOGNIZABLE AS HUMAN. YOU FIND NOTHING OF VALUE. |
| `ECL6.DAX/0x42` | `0x0F26` | `matched` | `myth-drannor.outer.fugitive.cache` | JUST AS THE DYING MAN DESCRIBED, YOU LOCATE A CACHE. |
| `ECL6.DAX/0x42` | `0x0F85` | `matched` | `myth-drannor.outer.nameless-warning` | NAMELESS SLIDES OUT OF THE SHADOWS, 'THIS DIRECT APPROACH IS DANGEROUS, BUT THE TEMPLE IS TO THE NORTH. NOW GET ON WITH IT!' HE LEAVES. |
| `ECL6.DAX/0x42` | `0x100B` | `matched` | `myth-drannor.outer.brush-decoy` | IN THE PLAZA AHEAD IS SOME DENSE BRUSH. A SMALL CHILD OR HOBBIT RACES OVER AND DIVES INTO IT. A HELL HOUND TRIES TO LEAP AFTER HIM. DO YOU GO TO THE RESCUE? |
| `ECL6.DAX/0x42` | `0x109A` | `matched` | `myth-drannor.outer.brush-rescue` | YOU ARRIVE BEFORE THE HOUND CAN FORCE ITS WAY IN AND CUT IT DOWN. |
| `ECL6.DAX/0x42` | `0x10DC` | `matched` | `myth-drannor.outer.brush-ambush` | A RAKSHASA STANDS UP FROM THE BUSH, 'WHY THANK YOU. SUCH KINDNESS SHOULD BE REWARDED.' HE GESTURES, AND MARGOYLES ON NEARBY ROOFS PELT YOU WITH ROCKS. |
| `ECL6.DAX/0x42` | `0x1165` | `matched` | `myth-drannor.outer.brush-attack` | THE MONSTERS THEN LEAP DOWN TO ATTACK. |
| `ECL6.DAX/0x42` | `0x119F` | `matched` | `myth-drannor.outer.brush-victim` | AFTER A MOMENT, THE HOUND EMERGES FROM THE BUSH WITH SOMETHING BLOODY IN ITS MOUTH AND TROTS OFF. |
| `ECL6.DAX/0x42` | `0x11F3` | `matched` | `myth-drannor.outer.brush-bloodstains` | THE BRUSH IS DENSE AND FILLED WITH RUBBLE. A FEW BLOODSTAINS MARK THE LEAVES. |
| `ECL6.DAX/0x42` | `0x1249` | `matched` | `myth-drannor.outer.gambling-room` | AS YOU STEP INTO THIS OPULENT ROOM, YOU SEE SEVERAL MALE RAKSHASAS GAMBLING WITH DICE. FEMALE RAKSHASAS LOUNGE AROUND ON PILLOWS. MARGOYLES ARE WATCHING THE GAM… |
| `ECL6.DAX/0x42` | `0x12DA` | `matched` | `myth-drannor.outer.gambling-rise` | THEY NOTICE YOU ONLY NOW AND SLOWLY RISE TO THEIR FEET . DO YOU FLEE? |
| `ECL6.DAX/0x42` | `0x132C` | `matched` | `myth-drannor.outer.gambling-treasure` | YOU GATHER THE VALUABLES. |
| `ECL6.DAX/0x42` | `0x1360` | `matched` | `myth-drannor.outer.margoyle-trap` | TWO MARGOYLES ARE TORTURING A SMALL ANIMAL IN A DOORWAY JUST AHEAD. THEY ARE UNAWARE OF YOU. DO YOU WANT TO ATTACK? |
| `ECL6.DAX/0x42` | `0x13DF` | `matched` | `myth-drannor.outer.margoyle-collapse` | YOU SWIFTLY KILL THEM. BEHIND YOU COMES LAUGHTER, THEN, THE DOORWAY COLLAPSES ONTO YOU. |
| `ECL6.DAX/0x42` | `0x143D` | `matched` | `myth-drannor.outer.margoyle-rakshasa` | A RAKSHASA COMES UP WHILE YOU ARE BURIED. DO YOU PLAY DEAD? |
| `ECL6.DAX/0x42` | `0x1479` | `matched` | `myth-drannor.outer.margoyle-surprise` | YOU SURPRISE HIM AS HE REACHES FOR YOU. |
| `ECL6.DAX/0x42` | `0x14AA` | `matched` | `myth-drannor.outer.margoyle-retreat` | SEEING THAT THE COLLAPSE FAILED TO FINISH YOU OFF, HE TURNS ON HIS HEEL AND DISAPPEARS. |
| `ECL6.DAX/0x42` | `0x1512` | `matched` | `myth-drannor.outer.sewer-margoyle` | A LONE MARGOYLE SKITTERS AWAY UNCOVERING A SEWER GRATE. IT SEEMS TO BE TRYING TO FLEE. DO YOU LET IT? |
| `ECL6.DAX/0x42` | `0x156D` | `matched` | `myth-drannor.outer.sewer-escape` | IT RUNS DOWN THROUGH THE SEWER AND DISAPPEARS. |
| `ECL6.DAX/0x42` | `0x159E` | `matched` | `myth-drannor.outer.sewer-kill` | YOU SLAUGHTER IT WHERE IT STANDS. |
| `ECL6.DAX/0x42` | `0x15C1` | `matched` | `myth-drannor.outer.sewer-grate` | THERE IS A SEWER GRATE HERE. DO YOU WANT TO ENTER? |
| `ECL6.DAX/0x42` | `0x165C` | `matched` | `myth-drannor.outer.rakshasa-threat` | THE RAKSHASA STANDS. SEVERAL MORE APPEAR. 'TO HAVE DISTURBED ONE OF MY RANK INVITES DEATH. PREPARE FOR YOURS.' |
| `ECL6.DAX/0x42` | `0x16E9` | `matched` | `myth-drannor.outer.rakshasa-parlay` | THE RAKSHASA SMILES. 'YOU HAVE A BOLDNESS THAT IS REFRESHING.' HE GOES ON TO TELL YOU HIS STORY AND YOU RECORD IT IN JOURNAL ENTRY 57. ⚠ 另有 `GOSUB 0x1737` 印出的一段 |
| `ECL6.DAX/0x43` | `0x00A1` | `matched` | `myth-drannor.kitchen-arrival` | THE SEWER ENDS IN A DARKENED KITCHEN. A RUMBLE AND A CLOUD OF DUST FORCE YOU FROM THE TUNNEL. |
| `ECL6.DAX/0x43` | `0x0131` | `matched` | `myth-drannor.inner.door-resists` | THIS DOOR RESISTS ALL ATTEMPTS TO BE OPENED. |
| `ECL6.DAX/0x43` | `0x018C` | `matched` | `myth-drannor.inner.bands-deny-rest` | YOUR BANDS WILL NOT LET YOU REST. |
| `ECL6.DAX/0x43` | `0x0370` | `matched` | `myth-drannor.inner.monsters-confused` | THE MONSTERS LOOK CONFUSED AND WANDER AWAY. |
| `ECL6.DAX/0x43` | `0x03D8` | `matched` | `myth-drannor.inner.minions-attack` | MINIONS OF TYRANTHRAXUS RUSH TO ATTACK YOU. |
| `ECL6.DAX/0x43` | `0x04A6` | `matched` | `myth-drannor.inner.helm-northeast` | THE HELM OF DRAGONS SHOWS TYRANTHRAXUS ON THE SECOND FLOOR, NORTHEAST CORNER. |
| `ECL6.DAX/0x43` | `0x04EC` | `matched` | `myth-drannor.inner.bonds-returning` | THE BONDS ARE WIGGLING BENEATH THE SKIN. THEIR POWER IS SLOWLY RETURNING. |
| `ECL6.DAX/0x43` | `0x053D` | `matched` | `myth-drannor.inner.ritual.arrival` | AS YOU ENTER, YOU HEAR A VOICE. 'FINALLY YOU HAVE COME. STEP FORWARD.' UNABLE TO CONTROL YOURSELF, YOU STEP FORWARD INTO THE CENTER OF THE ROOM. |
| `ECL6.DAX/0x43` | `0x05C6` | `matched` | `myth-drannor.inner.ritual.control` | 'THROUGH THE BONDS, YOU ARE MINE TO CONTROL. FOLLOW ME! |
| `ECL6.DAX/0x43` | `0x060D` | `matched` | `myth-drannor.inner.ritual.journal` | TYRANTHRAXUS MAKES A SPEECH, AND YOU RECORD IT IN JOURNAL ENTRY 48. |
| `ECL6.DAX/0x43` | `0x064A` | `matched` | `myth-drannor.inner.ritual.hand-over` | 'HAND YOUR TOYS OVER TO MY PRIEST.' THE PARTY HANDS OVER THE THREE ARTIFACTS, THOUGH THEY RESIST WITH ALL THEIR MIGHT. |
| `ECL6.DAX/0x43` | `0x06C0` | `matched` | `myth-drannor.inner.ritual.dispose-order` | 'NOW PRIEST, DISPOSE OF THESE UNPLEASANT ITEMS THROUGH THE POOL.' |
| `ECL6.DAX/0x43` | `0x06FF` | `matched` | `myth-drannor.inner.ritual.pool` | THE PRIEST MOVES WITH GREAT FLOURISH AND GRACEFULLY ARCS EACH ARTIFACT INTO THE POOL, WHICH SWIRLS FOR A MOMENT, AND THE ITEMS ARE GONE. |
| `ECL6.DAX/0x43` | `0x0776` | `matched` | `myth-drannor.inner.ritual.parchment` | 'I EVEN HAVE A PARCHMENT WITH THE PHRASE TO RELEASE YOUR BONDS, IN CASE ANYTHING GOES WRONG.' HE HANDS THE PARCHMENT TO THE PRIEST. 'DISPOSE OF THIS THROUGH THE… |
| `ECL6.DAX/0x43` | `0x0801` | `matched` | `myth-drannor.inner.ritual.final-spell` | 'YOU BONDED ONES COME WITH ME. IT IS TIME TO COMPLETE THE FINAL SPELL.' |
| `ECL6.DAX/0x43` | `0x0850` | `matched` | `myth-drannor.inner.ritual.nameless-reveal` | THE PRIEST THROWS BACK HIS HOOD, REVEALING HIMSELF AS NAMELESS. HE YELLS,'IT WON'T BE THAT EASY!' HE TOSSES THE ARTIFACTS BACK TO YOU. |
| `ECL6.DAX/0x43` | `0x08C6` | `matched` | `myth-drannor.inner.ritual.nameless-falls` | 'A PITIFUL ATTEMPT, YOU HAVE GAINED NOTHING BUT A SWIFT DEATH!' HE STRIKES NAMELESS DOWN WITH A SINGLE BLOW. WITH HIS LAST BREATH, NAMELESS MOUTHS THE MEANINGLE… |
| `ECL6.DAX/0x43` | `0x0963` | `matched` | `myth-drannor.inner.ritual.bonds-fade` | THE PARTY FEELS THE BOND'S CONTROL FADE, BUT TYRANTHRAXUS' SIGIL DOES NOT GO AWAY. TYRANTHRAXUS SMILES. 'ONLY I CAN REMOVE THOSE BONDS. SOON YOU WILL BE MINE AG… |
| `ECL6.DAX/0x43` | `0x09F1` | `matched` | `myth-drannor.inner.ritual.recover` | AS THE PARTY RETRIEVES THE ARTIFACTS, TYRANTHRAXUS LOOKS AFRAID. 'KILL THEM MY PETS!' HE THEN RUSHES OFF. |
| `ECL6.DAX/0x43` | `0x0A77` | `matched` | `myth-drannor.inner.temple-alarm` | A TREMENDOUS NOISE IS HEARD FROM OUTSIDE. SOME OF TYRANTHRAXUS' FORCE GOES TO PROTECT THE TEMPLE. |
| `ECL6.DAX/0x43` | `0x0AF5` | `matched` | `myth-drannor.inner.kennel` | A SULFEROUS SMELL ASSAULTS YOUR NOSTRILS. THE ROOM HAS BEEN CONVERTED TO A KENNEL. THE INHABITANTS RISE AT YOUR APPROACH. |
| `ECL6.DAX/0x43` | `0x0B7A` | `matched` | `myth-drannor.inner.statuary` | THE ROOM SEEMS FILLED WITH BONES AND HIDEOUS STATUARY -- UNTIL THE STATUES BEGIN TO MOVE. |
| `ECL6.DAX/0x43` | `0x0BF2` | `matched` | `myth-drannor.inner.chapel` | YOU HAVE WALKED INTO AN ELEGANT PRIVATE CHAPEL. A WELL-DRESSED, ELDERLY MAN IS SEATED HERE. |
| `ECL6.DAX/0x43` | `0x0C42` | `matched` | `myth-drannor.inner.chapel-priest` | 'SO YOU ARE TYRANTHRAXUS' GRAND TOOLS. I COULD ASK YOU TO SURRENDER, BUT HE HAS BEEN TOO POPULAR WITH BANE OF LATE. IT WOULD HURT HIM MORE IF YOU DIED. SO SORRY… |
| `ECL6.DAX/0x43` | `0x0D07` | `matched` | `myth-drannor.inner.bedroom` | YOU HAVE ENTERED AN ELEGANT BEDROOM WITH A FOUR-POSTER BED, TAPESTRIES, GOLD FIXTURES. DO YOU WANT TO LOOT THE ROOM? |
| `ECL6.DAX/0x43` | `0x0D92` | `matched` | `myth-drannor.inner.office` | THIS ROOM HAS BEEN CONVERTED TO AN OFFICE. THE WALLS ARE COVERED WITH ICONS OF BANE. |
| `ECL6.DAX/0x43` | `0x0DF0` | `matched` | `myth-drannor.inner.kitchen` | YOU HAVE COME INTO THE KITCHEN. SLAVES DIVE UNDER TABLES AT YOUR ARRIVAL. THEY ARE NO THREAT TO YOU. |
| `ECL6.DAX/0x43` | `0x0E4F` | `matched` | `myth-drannor.inner.sewer-collapsed` | THE SEWER HAS COLLAPSED. THIS IS NO LONGER A WAY OUT. |
| `ECL6.DAX/0x43` | `0x0E92` | `matched` | `myth-drannor.inner.tiered-beds` | TIERED BEDS LINE THE WALLS, FILLED WITH MEDITATING PRIESTS. THEY AREN'T HAPPY ABOUT BEING DISTURBED. |
| `ECL6.DAX/0x43` | `0x0F0B` | `matched` | `myth-drannor.inner.worshipping-priests` | THE ROOM IS FILLED WITH WORSHIPPING PRIESTS. THEY NOTE YOUR ENTRY WITH ANGER. |
| `ECL6.DAX/0x43` | `0x0F6E` | `matched` | `myth-drannor.inner.magic-circle` | A HIGH PRIEST IS HERE SUMMONING UP THE POWER OF BANE. BLUE LIGHTNING ARCS AT YOU AS YOU ENTER THE MAGIC CIRCLE. |
| `ECL6.DAX/0x43` | `0x0FDB` | `matched` | `myth-drannor.inner.magic-circle-disrupted` | ON THE POSITIVE SIDE, YOU HAVE DISRUPTED THE CEREMONY. |
| `ECL6.DAX/0x43` | `0x1032` | `matched` | `myth-drannor.inner.food-storeroom` | RATS SKITTER UNDER BAGS AS YOU OPEN THE DOOR INTO THIS FOOD STOREROOM. |
| `ECL6.DAX/0x43` | `0x107A` | `matched` | `myth-drannor.inner.library` | THIS ROOM IS LINED WITH SHELVES OF MOULDERING BOOKS. HIDDEN AWAY ARE A FEW VALUABLES. ⚠ 另有 `GOSUB 0x10AB` 印出的一段 |
| `ECL6.DAX/0x43` | `0x10B2` | `matched` | `myth-drannor.inner.hidden-valuables` | HIDDEN AWAY ARE A FEW VALUABLES. |
| `ECL6.DAX/0x43` | `0x110E` | `matched` | `myth-drannor.inner.biers` | THE ROOM CONTAINS OLD BIERS AND CASKETS. |
| `ECL6.DAX/0x43` | `0x1142` | `matched` | `myth-drannor.inner.preservation-room` | THE STENCH OF PRESERVING FLUIDS IS STRONG IN HERE. TO THE NORTH ARE SEVERAL TABLES WITH WRAPPED FORMS. THEY SEEM TO SHIFT AND RISE AT YOUR ENTRANCE. |
| `ECL6.DAX/0x43` | `0x11BF` | `matched` | `myth-drannor.inner.preservation-illusion` | BUT SOON YOU REALIZE THAT THE LIGHT IS PLAYING TRICKS ON YOUR EYES. |
| `ECL6.DAX/0x43` | `0x11FD` | `matched` | `myth-drannor.inner.stairs-up` | STAIRS LEAD UP HERE. DO YOU WANT TO GO UP? |
| `ECL6.DAX/0x43` | `0x123D` | `matched` | `myth-drannor.inner.stairs-down` | STAIRS LEAD DOWN HERE. DO YOU WANT TO DESCEND? |
| `ECL6.DAX/0x43` | `0x1287` | `matched` | `myth-drannor.inner.final-compulsion` | 'THE POWER OF YOUR BONDS HAS RETURNED. GROVEL AT MY FEET THIS INSTANT!' YOUR BODY BEGINS TO BOW DOWN. |
| `ECL6.DAX/0x43` | `0x12DD` | `matched` | `myth-drannor.inner.final-defiance` | WITH A GREAT FORCE OF WILL, YOU OVERCOME THE COMPULSION. 'SO BE IT! HERE IS YOUR DESTINY -- GROUND TO DUST BENEATH MY FEET!' |
| `ECL6.DAX/0x43` | `0x1348` | `matched` | `myth-drannor.inner.final-amulet` | HE SCOWLS. 'THAT AMULET WILL LET YOU SCRATCH ME, BUT IT IS FAR FROM ENOUGH TO DEFEAT ME. FOR MY GREATER GLORY, LET YOUR LIVES BE FORFEIT!' |
| `ECL6.DAX/0x45` | `0x0084` | `matched` | `myth-drannor.ruins.found-some-ruins` | YOU FOUND SOME RUINS. |
| `ECL6.DAX/0x45` | `0x00BE` | `matched` | `myth-drannor.ruins.eerie-block` | YOU HAVE FOUND AN EERIE BLOCK OF RUINS. |
| `ECL6.DAX/0x45` | `0x01C5` | `matched` | `yulash.false-door` | YOU HAVE FOUND A FALSE DOOR. |
| `ECL6.DAX/0x45` | `0x0201` | `matched` | `yulash.about-to-leave` | YOU ARE ABOUT TO LEAVE. DO YOU WANT TO? |
| `ECL6.DAX/0x45` | `0x0379` | `matched` | `yulash.path-out` | YOU LOCATE A PATH OUT. DO YOU WANT TO EXIT? |
| `ECL6.DAX/0x45` | `0x04E1` | `matched` | `dark-elf-caves.monsters-spotted` | YOU SPOT MONSTERS. |
| `ECL6.DAX/0x45` | `0x05A7` | `matched` | `myth-drannor.ruins.hidden-stone-note` | YOU FIND A NOTE THAT SAYS, 'SEEK THE HIDDEN STONE, BENEATH IT LIES THE RIVER PIRATES' TREASURE.' |
| `ECL6.DAX/0x45` | `0x060D` | `matched` | `myth-drannor.ruins.priest-and-pirates` | YOU SPOT A PRIEST ARGUING WITH SOME PIRATES ABOUT A MAP. THEY SEEM UNAWARE OF YOU. WHAT DO YOU DO? |
| `ECL6.DAX/0x45` | `0x06B0` | `matched` | `myth-drannor.ruins.cryptic-map` | YOU FIND A CRYPTIC MAP MARKED WITH AN X. |
| `ECL6.DAX/0x45` | `0x06D6` | `matched` | `myth-drannor.ruins.marked-area-treasure` | YOU FIND THE AREA MARKED ON THE MAP. A STONE BENEATH YOUR FEET SEEMS LOOSE. SHIFTING IT REVEALS A TREASURE. |
| `ECL6.DAX/0x45` | `0x0773` | `matched` | `myth-drannor.ruins.hidden-pinnace` | YOU FIND A SMALL PINNACE HIDDEN HERE. WET FOOTPRINTS LEAD OFF. IT APPEARS THAT SOMETHING HEAVY WAS BEING DRAGGED. |
| `ECL6.DAX/0x45` | `0x07EA` | `matched` | `myth-drannor.ruins.pirates-leap` | PIRATES LEAP UPON YOU. |
| `ECL6.DAX/0x45` | `0x082B` | `matched` | `myth-drannor.ruins.pirates-digging` | SOME PIRATES ARE BUSILY DIGGING HERE. SUDDENLY, ONE SHOUTS, 'HOLD IT MATES. WE BEEN SPOTTED. KEEL HAUL EM.' THE PIRATES RUSH FORWARD. |
| `ECL6.DAX/0x45` | `0x08B4` | `matched` | `myth-drannor.ruins.find-their-treasure` | YOU FIND THEIR TREASURE. |
