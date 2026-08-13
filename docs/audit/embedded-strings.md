# overlay 內嵌字串（掃描結果）

由 `scripts/scan_pascal_strings.py` 產生。判準是「Pascal 短字串形狀」
＋「有 `mov di, offset` 引用」兩條同時成立，所以**這是下界不是全集**：
用其他方式取址的字串不會出現在這裡，不得由缺席推論不存在。

## dos（32 個模組，792 條）

### overlay-00

| offset | 內容 |
|---|---|
| `0000h` | Time to save your game |

### overlay-01

| offset | 內容 |
|---|---|
| `0051h` | based on the tsr novel "azure bonds" |
| `007Ah` | kate novak |
| `0089h` | jeff grubb |
| `0094h` | scenario created by: |
| `00A9h` | tsr, inc. |
| `00B7h` | george mac donald |
| `00C9h` | game created by: |
| `00DAh` | ssi special projects |
| `00EFh` | project leader: |
| `00FFh` | programming: |
| `010Ch` | scot bayless |
| `0119h` | russ brown |
| `0124h` | michael mancuso |
| `0134h` | development: |
| `0141h` | david shelley |
| `014Fh` | oran kangas |
| `015Bh` | graphic arts: |
| `0169h` | tom wahl |
| `0172h` | fred butts |
| `017Dh` | susan manley |
| `018Ah` | mark johnson |
| `0197h` | cyrus lum |
| `01A1h` | playtesting: |
| `01AEh` | jim jennings |
| `01BBh` | james kucera |
| `01C8h` | rick white |
| `01D3h` | robert daly |
| `062Eh` | title |

### overlay-02

| offset | 內容 |
|---|---|
| `0461h` | CPIC |
| `1037h` | PRESS BUTTON OR RETURN TO CONTINUE. |
| `105Bh` | PRESS <ENTER>/<RETURN> TO CONTINUE |
| `1B30h` | ITEM |
| `1B35h` | .dax |
| `1B3Ah` | Unable to find item file |
| `2029h` | ~COMBAT ~WAIT ~FLEE ~PARLAY |
| `2045h` | ~COMBAT ~WAIT ~FLEE ~ADVANCE |
| `2062h` | Both sides wait. |
| `2073h` | The monsters flee. |
| `2785h` | ~HAUGHTY ~SLY ~NICE ~MEEK ~ABUSIVE |
| `2903h` | The entire party is killed! |
| `291Fh` | press <enter>/<return> to continue |
| `30BAh` | You've won. Save before quitting?  |

### overlay-03

| offset | 內容 |
|---|---|
| `000Ch` | tiles |
| `0012h` | Align the espruar and dethek runes |
| `0035h` | shown below, on translation wheel |
| `0057h` | like this: |
| `0062h` | -..-..-.. |
| `006Ch` | - - - - - |
| `0076h` | ......... |
| `0080h` | Type the character in box number  |
| `00A2h` | under the  |
| `00ADh` | path. |
| `00B5h` | type character and press return:  |
| `00D7h` | Sorry, that's incorrect. |
| `00F0h` | An unseen force hurls you into the abyss! |

### overlay-04

| offset | 內容 |
|---|---|
| `0034h` | cast cure anyway:  |
| `00AAh` |  will only cost  |
| `00BBh` |  gold pieces. |
| `00C9h` | pay for cure  |
| `00D7h` | Not enough money. |
| `00E9h` | is cured. |
| `0263h` | is not blind. |
| `0271h` | Cure Blindness |
| `0306h` | is not Diseased. |
| `0317h` | Cure Disease |
| `03FDh` | Cure Light Wounds |
| `040Fh` | Cure Serious Wounds |
| `0423h` | Cure Critical Wounds |
| `0438h` | Heal |
| `0626h` | is not dead. |
| `0633h` | Raise Dead |
| `086Fh` | is not poisoned. |
| `0880h` | Neutralize Poison |
| `095Fh` | is not cursed. |
| `096Eh` | Remove Curse |
| `0A45h` | is not stoned. |
| `0A54h` | Stone to Flesh |
| `0AF0h` | , how can we help you? |
| `0B08h` | Heal Exit |
| `0D0Ch` | Heal View Take Pool Share Appraise Exit |
| `0D34h` | Heal View Pool Appraise Exit |
| `0D52h` | ~Yes ~No |
| `0D5Bh` | As you leave a priest says, "Excuse me but you have left some money here"  |
| `0DA6h` | Do you want to go back and retrieve your money? |

### overlay-05

| offset | 內容 |
|---|---|
| `09A4h` | The party has fled. |
| `09B8h` | You have lost the fight. |
| `09D1h` | You have won the duel. |
| `09E8h` | The party has won. |
| `09FBh` | The party has found Treasure! |
| `0A19h` | The duelist receives  |
| `0A2Fh` | Each character receives  |
| `0A48h` | experience points. |
| `0A5Bh` | press <enter>/<return> to continue |
| `0CD9h` | Items:  |
| `0CE1h` | Take |
| `0EBFh` | Take:  |
| `0EC6h` | Money Items Exit |
| `0FCDh` | View Pool Exit |
| `0FDCh` |  Exit |
| `0FE2h` |  Detect Exit |
| `0FEFh` | View Take Pool Share |
| `1004h` | View Take Pool |
| `1013h` | ~Yes ~No |
| `101Ch` | There is still treasure left.   |
| `103Ch` | Do you want to go back and claim your treasure? |
| `1466h` |  takes and hides  |
| `147Ch` |  share. |
| `1489h` | press <enter>/<return> to continue |
| `16E6h` | The monsters rejoice for the party has been destroyed |
| `171Ch` | Press any key to continue |

### overlay-06

| offset | 內容 |
|---|---|
| `0025h` |                                |
| `0044h` | Items:  |
| `034Ah` | OverLoaded |
| `0462h` | Not enough Money. |
| `062Ch` | Buy View Take Pool Share Appraise Exit |
| `0653h` | Buy View Pool Appraise Exit |
| `066Fh` | ~Yes ~No |
| `0678h` | As you Leave the Shopkeeper says, "Excuse me but you have Left Some Money here."   |
| `06CBh` | Do you want to go back and get your Money? |

### overlay-07

| offset | 內容 |
|---|---|
| `0361h` | Loading...Please Wait |
| `037Bh` | .dax |
| `0587h` | SPRIT |
| `1BD6h` | CPIC |
| `1BDBh` | ROLF |
| `2207h` |  dies.  |
| `220Fh` |  is hit FOR  |
| `221Ch` |  points of Damage. |
| `222Fh` | press <enter>/<return> to continue |

### overlay-08

| offset | 內容 |
|---|---|
| `03B0h` | Magic On |
| `03B9h` | Magic Off |
| `03C3h` | That doesn't work |
| `06FFh` | Move  |
| `0705h` | View Aim  |
| `070Fh` | Use  |
| `0714h` | Cast  |
| `071Ah` | Turn  |
| `0720h` | Quick Done |
| `096Fh` | Your Teammate is Dying |
| `0986h` | Continue Battle: |
| `0AD4h` | Move/Attack, Move Left =  |
| `0AF1h` | Flee: |
| `0AF7h` | can't go there |
| `0ED4h` | Not with that weapon |
| `0FA1h` | Guard  |
| `0FA8h` | Delay Quit  |
| `0FB4h` | Bandage  |
| `0FBDh` | Speed Exit |
| `11B8h` | GameSpeed ( |
| `11CAh` | Slower  |
| `11D2h` | Faster  |
| `11DAh` | Exit |

### overlay-09

| offset | 內容 |
|---|---|
| `003Eh` | flees in panic |
| `09ABh` | Move/Attack, Move Left =  |
| `1260h` | Magic On |
| `1269h` | Magic Off |
| `136Bh` | is forced to flee |
| `137Dh` | Surrenders |

### overlay-10

| offset | 內容 |
|---|---|
| `1017h` | DungCom |
| `101Fh` | WildCom |
| `1027h` | RandCom |
| `1C2Bh` | A battle begins... |

### overlay-11

| offset | 內容 |
|---|---|
| `002Fh` | Loading...Please Wait |
| `0045h` | COMSPR |
| `0050h` | ITEMS |

### overlay-12

| offset | 內容 |
|---|---|
| `00D0h` | is fighting with snakes |
| `0395h` | Suffocates |
| `04D8h` | is silenced |
| `0539h` | dies from poison |
| `05A8h` | Gains an item |
| `075Bh` | lost an image |
| `080Ch` | is coughing |
| `08C3h` | collapses |
| `09E0h` | runs away |
| `09EAh` | is confused |
| `09F6h` | goes berserk |
| `0A03h` | is enraged |
| `0BE4h` | ages |
| `0C46h` | The air clears a little... |
| `0F2Fh` | Avoids it |
| `1063h` | is weakened |
| `143Eh` | is Poisoned |
| `144Ah` | is killed |
| `14ECh` | is Paralyzed |
| `15CFh` | is stupid |
| `15D9h` | lost a spell |
| `171Bh` | is unaffected |
| `17FCh` | goes berzerk |
| `1A45h` | gazes... |
| `1A4Eh` | reflects it! |
| `1A5Bh` | is Stoned |
| `1CDBh` | The air clears a little... |
| `2031h` | Falls dead |
| `264Ch` | Spits Acid |
| `2775h` | is paralyzed |
| `2886h` | gazes... |
| `288Fh` | is paralyzed |
| `2CCDh` | gets zapped |
| `2D5Ch` | is dispelled |
| `2D69h` | resists dispel evil |

### overlay-13

| offset | 內容 |
|---|---|
| `0256h` | -Backstabs- |
| `0262h` | slays helpless |
| `0271h` | Attacks |
| `0279h` | (from behind)  |
| `0288h` | with one cruel blow |
| `029Ch` | Hitting for  |
| `02A9h` |  point  |
| `02B1h` |  points  |
| `02BAh` | of damage |
| `02C4h` | and Misses |
| `02CFh` | lost a spell |
| `02DCh` | goes down |
| `02E6h` | and is Dying |
| `0313h` | is killed |
| `0D01h` | Got Away |
| `0D0Ah` | Escape is blocked |
| `0F3Fh` | sweeps |
| `11EDh` | turns undead... |
| `11FDh` | is turned |
| `1207h` | Is destroyed |
| `1214h` | Nothing Happens... |
| `220Ah` | Already been targeted |
| `273Bh` | Camp Only Spell |
| `274Bh` | Begins Casting |
| `2F2Eh` | Attack Ally:  |
| `3009h` | Range =  |
| `3015h` | Target  |
| `301Dh` | Next Prev Manual  |
| `302Fh` | Center Exit |
| `303Bh` | Aim: |
| `3379h` | Range =  |
| `3385h` | Center Exit |
| `3391h` | Target  |
| `3399h` | (Use Cursor keys)  |
| `4036h` | engulfs  |
| `4286h` | fires a disintegrate ray |
| `429Fh` | is disintergrated |
| `42B1h` | fires a stone to flesh ray |
| `42CCh` | is Stoned |
| `42D6h` | fires a death ray |
| `42E8h` | is killed |
| `42F2h` | wounds you |
| `480Bh` | hugs  |
| `48C8h` | The Gods intervene! |

### overlay-14

| offset | 內容 |
|---|---|
| `0901h` | Not Here |
| `0B8Eh` | Bash |
| `0B93h` |  Pick |
| `0B99h` |  Knock |
| `0BA0h` |  Exit |
| `0BA6h` | Locked.  |

### overlay-15

| offset | 內容 |
|---|---|
| `0327h` | cannot cast spells in this area |
| `0347h` | is in no condition to  |
| `035Eh` | cast any spells |
| `036Eh` | memorize spells |
| `037Eh` | scribe any scrolls |
| `04B3h` | has no spells memorized |
| `0583h` | can memorize: |
| `0591h` |     Cleric Spells: |
| `05A4h` |      Druid Spells: |
| `05B7h` | Magic-User Spells: |
| `08F7h` | Memorize These Spells?  |
| `090Fh` | cannot memorize any spells |
| `092Ah` | Memorize these spells?  |
| `0AF9h` | Scribe These Spells?  |
| `0B0Fh` | has no copyable scrolls |
| `0B27h` | You already know that spell |
| `0B43h` | You are already scibing that spell |
| `0B66h` | You can not scribe that spell. |
| `0B85h` | Scribe these spells?  |
| `0EEEh` | Funky-- |
| `0EF6h` | Dispel Evil |
| `0F02h` | Faerie Fire |
| `0F0Eh` | Fumbling |
| `0F17h` | Helpless |
| `0F20h` | Confused |
| `0F29h` | Cause Disease |
| `0F37h` | Hot Fire Shield |
| `0F47h` | Cold Fire Shield |
| `0F58h` | Poisoned |
| `0F61h` | Regenerating |
| `0F6Eh` | Fire Resistance |
| `0F7Eh` | Minor Globe of Invulnerability |
| `0F9Dh` | enfeebled |
| `0FA7h` | invisible to animals |
| `0FBCh` | Invisible |
| `0FC6h` | Camouflaged |
| `0FD2h` | protected from dragon breath |
| `0FEFh` | berserk |
| `0FF7h` | Displaced |
| `1001h` |  <No Spell Effects> |
| `17F4h` | Party Order:  |
| `1822h` | has been selected |
| `192Eh` | quit TO DOS:  |
| `193Ch` | will be gone |
| `1949h` | Drop from party?  |
| `195Bh` | bids you farewell |
| `196Dh` | is dumped in a ditch |
| `1982h` | Breathes A sigh of relief |
| `1AE6h` | Game Speed =  |
| `1AF4h` |  (0=fastest 9=slowest) |
| `1B0Bh` |  Faster |
| `1B13h` |  Slower |
| `1B1Bh` |  Exit |
| `1B21h` | Game Speed: |
| `1CC9h` | Alter:  |
| `1CD1h` | Pics on   |
| `1CDBh` | Animation on   |
| `1CEAh` | Animation off   |
| `1CFAh` | Pics off   |
| `1D05h` | Exit |
| `23ADh` | The party makes camp... |
| `23E5h` | Camp: |
| `23EBh` | Quit TO DOS  |

### overlay-16

| offset | 內容 |
|---|---|
| `0077h` | .SAV |
| `007Ch` | from saved game  |
| `043Bh` | *.guy |
| `0441h` | *.cha |
| `0447h` | *.sav |
| `044Dh` | *.hil |
| `05CAh` | CURSE |
| `05D0h` | Put save disk in  |
| `05E4h` | Is your save disk in drive  |
| `0604h` | Unexpected error during save:  |
| `0A72h` | CHEAD |
| `0A78h` | CBODY |
| `0C70h` | .guy |
| `0C75h` | .swg |
| `0D1Bh` | Can't save.  No room on this disk. |
| `0D3Eh` | Lose character?  |
| `0D4Fh` | Ok  Try another disk |
| `0D64h` | .guy |
| `0D6Bh` | .sav |
| `0D90h` | Put save disk in  |
| `0DA4h` | Unexpected error during save:  |
| `0DC3h` | Overwrite  |
| `0DD1h` | New file name:  |
| `0DE1h` | .swg |
| `222Ah` | Put save disk in  |
| `223Eh` | Loading...Please Wait |
| `2254h` | .guy |
| `225Bh` | .cha |
| `2260h` | .swg |
| `2269h` | .spc |
| `3230h` | .dax |
| `3235h` | Unable to load monster |
| `3517h` | CPIC |
| `36D4h` | savgam |
| `36DBh` | .dat |
| `36E2h` | Load Which Game:  |
| `3714h` | Put save disk in  |
| `3726h` | Loading...Please Wait |
| `373Eh` | .sav |
| `3743h` | CPIC |
| `3DF9h` | Save Which Game:  |
| `3E0Bh` | A B C D E F G H I J |
| `3E3Fh` | savgam |
| `3E46h` | .dat |
| `3E4Bh` | Can't save.  No room on this disk. |
| `3E8Eh` | Put save disk in  |
| `3EA2h` | Unexpected error during save:  |
| `3EC1h` | Saving...Please Wait |
| `3ED6h` | CHRDAT |

### overlay-17

| offset | 內容 |
|---|---|
| `0115h` | Choose a function  |
| `0168h` | You've already defeated Tyranthraxus. |
| `018Eh` | Quit to DOS  |
| `019Bh` | Game not saved.  Quit anyway?  |
| `01BAh` | Free training on |
| `01CBh` | Free training off |
| `06FCh` | Pick Race |
| `070Ah` | Select |
| `0711h` | Pick Gender |
| `071Dh` | Pick Class |
| `0728h` | Pick Alignment |
| `0759h` | Reroll stats?  |
| `0768h` | Character name:  |
| `0779h` | Save  |
| `25A7h` | Drop  |
| `25ADh` |  forever?  |
| `25B8h` | Are you sure?  |
| `25C7h` | You dump  |
| `25D1h` |  out back. |
| `25DCh` |  bids you farewell. |
| `25F0h` |  breathes a sigh of relief. |
| `2841h` |  can't be modified. |
| `2855h` | Modify:  |
| `285Eh` | Keep Exit |
| `35E6h` | Add from where?  |
| `35F7h` | Curse Pool Hillsfar Exit |
| `3610h` | Add a character:  |
| `3622h` | Add  |
| `362Ah` | paladins do not join with evil scum |
| `364Eh` | too many rangers in party |
| `3668h` |  will tolerate no evil! |
| `3E75h` | ready   action |
| `3E88h` | Small |
| `3E8Eh` | Large |
| `3E95h` | Sorry, not in CGA |
| `3EA7h` | Hair |
| `3EACh` | Face |
| `3ED1h` | Is this icon ok?  |
| `4C6Ah` | we only train conscious people |
| `4C89h` | Training costs 1000 gp. |
| `4CA1h` | We don't train that class here |
| `4CC0h` | Not Enough Experience |
| `4CD6h` |  will become: |
| `4CE4h` |     a level  |
| `4CF3h` | and a level  |
| `4D00h` | Do you wish to train?  |
| `4D17h` | Congratulations... |

### overlay-18

| offset | 內容 |
|---|---|
| `0C6Ah` | Tyranthraxus' spirit coalesces over the slain  |
| `0C99h` | storm giant. 'You have defeated me. Were it not for  |
| `0CCEh` | the Amulet of Lythander, I could possess you and rob  |
| `0D04h` | you of your victory. Still I can escape through the pool. |
| `0D3Eh` | Press any key to continue. |
| `0D59h` | As you reach for the Pool of Radiance, he cries  |
| `0D8Ah` | out, 'Keep the Gauntlet of Moander away from there, you  |
| `0DC3h` | will unleash dangerous energies. Stay back!' As the  |
| `0DF8h` | gauntlet contacts the pool, it contracts and shatters it. |
| `0E32h` | 'I am trapped without escape, you have succeeded  |
| `0E64h` | where armies have not. Gloat while you may, Tyranthraxus  |
| `0E9Eh` | is slain this day.' Before your eyes he crumbles into  |
| `0ED5h` | nothingness. |
| `0EE2h` | You are certain he is destroyed because your  |
| `0F10h` | final bond fades away. The Curse of the Azure Bonds  |
| `0F45h` | has finally been lifted from you! You are free at  |
| `0F78h` | last! |
| `0F7Eh` | The Knights of Myth Drannor rush in, ' |
| `0FA5h` | Congratulations, you have destroyed the Flamed One.  |
| `0FDAh` | With the power of Elminster, let us take you from  |
| `100Dh` | this  foul place, to a fine feast.' |
| `1031h` | You are teleported to Shadowdale, where festivities  |
| `1066h` | have already begun. A huge cheer goes up at your arrival.  |
| `10A1h` | Gharri and Nacacia, arm in arm, yell congratulations  |
| `10D7h` | from the nearby stands. 'You have won!' |

### overlay-19

| offset | 內容 |
|---|---|
| `0039h` | (NPC) |
| `003Fh` | Age  |
| `0044h` | STR  |
| `0049h` | INT  |
| `004Eh` | WIS  |
| `0053h` | DEX  |
| `0058h` | CON  |
| `005Dh` | CHA  |
| `0062h` | Level |
| `006Ah` | Exp  |
| `006Fh` | Weapon |
| `0076h` | Armor |
| `007Ch` | Status |
| `064Eh` | AC     |
| `0655h` | HP     |
| `065Ch` | THAC0    |
| `066Bh` | Damage         |
| `067Ah` | Encumbrance       |
| `068Ch` | Movement    |
| `0B28h` | Items  |
| `0B2Fh` | Spells  |
| `0B37h` | Trade  |
| `0B3Eh` | Drop  |
| `0B44h` | Heal  |
| `0B4Ah` | Cure  |
| `0B50h` | Exit |
| `0E75h` | Must be unreadied |
| `0E87h` |  was going to scribe from that scroll |
| `0EADh` | is it Okay to lose it?  |
| `0FBCh` | itemptr:       |
| `0FCBh` | namenum(1):    |
| `0FDAh` | namenum(2):    |
| `0FE9h` | namenum(3):    |
| `0FF8h` | plus:          |
| `1007h` | plussave:      |
| `1016h` | ready:         |
| `1025h` | identified:    |
| `1034h` | cursed:        |
| `1043h` | value:         |
| `1052h` | special(1):    |
| `1061h` | special(2):    |
| `1070h` | special(3):    |
| `107Fh` | dice large:     |
| `108Fh` | sides large:    |
| `109Fh` | press a key |
| `1513h` | Ready |
| `1519h` |  Use |
| `151Eh` |  Trade |
| `1525h` |  Drop |
| `152Bh` |  Halve |
| `1532h` |  Join |
| `1538h` |  Sell |
| `1542h` | Items |
| `1548h` | Ready Item |
| `1553h` | Must be Readied |
| `1563h` | Your  |
| `1569h` | will be gone forever |
| `157Eh` | Drop It?  |
| `1EADh` | It's Cursed |
| `1EB9h` | Wrong Class |
| `1EC5h` | already using  |
| `1ED4h` | Your hands are full! |
| `2120h` | Trade with Whom? |
| `2131h` | Overloaded |
| `21E2h` | Can't halve that |
| `2474h` | uses an item |
| `2481h` | Item: |
| `2487h` | oops! |
| `2773h` | I'll give you  |
| `2782h` |  gold pieces for your  |
| `2799h` | Is It a Deal?  |
| `27A8h` | Sold! |
| `27AEh` | Overloaded.  Money will be put in pool. |
| `29B8h` | For 200 gold pieces I'll identify your  |
| `29E0h` | Is It a Deal?  |
| `29EFh` | Not Enough Money |
| `2A00h` | I can't tell anything new about your  |
| `2A26h` | It looks like some sort of  |
| `2C13h` | Trade to? |
| `2C1Fh` | Select type of coin  |
| `2C34h` |  Select |
| `2C3Ch` | How much  |
| `2C46h` | will you trade?  |
| `2F72h` | Select type of coin  |
| `2F87h` |  Select |
| `2F8Fh` | How much  |
| `2F99h` | will you drop?  |
| `3421h` | in Memory |
| `342Bh` | in Grimoire |
| `3437h` | on Scroll |
| `3441h` | on Scrolls |
| `344Ch` | to Choose |
| `3456h` | to Memorize |
| `3462h` | to Scribe |
| `346Ch` | Spells  |
| `36AFh` | Heal whom?  |
| `36BBh` |  feels better |
| `36C9h` |  is unaffected |
| `37B8h` | Cure whom?  |
| `37C4h` | is not diseased |
| `37D4h` | cure anyway:  |
| `37E2h` |  is cured |

### overlay-20

| offset | 內容 |
|---|---|
| `05C7h` | Rest Time: |
| `06CBh` | Rest Days Hours Mins Add Subtract Exit |
| `0848h` | The Whole Party Is Healed |
| `090Fh` | has memorized |
| `09D3h` | has scribed |
| `0C68h` | Stop Resting?  |
| `0C77h` | Your repose is suddenly interrupted! |

### overlay-21

| offset | 內容 |
|---|---|
| `01F4h` | Overloaded.  Money will be put in Pool. |
| `0460h` | Overloaded |
| `0A7Fh` | Overloaded |
| `0B58h` | Gems  |
| `0B5Eh` | Gold  |
| `0B64h` | Platinum  |
| `0B6Eh` | electrum  |
| `0B78h` | silver  |
| `0B80h` | Copper  |
| `0B88h` | Jewelry  |
| `0CBAh` | Select type of coin  |
| `0CCFh` | Select |
| `0CD6h` | How much  |
| `0CE0h` | will you take?  |
| `18BAh` | No Gems or Jewelry |
| `18CDh` |  Gem |
| `18D2h` |  Gems |
| `18D8h` |  piece of Jewelry |
| `18EAh` |  pieces of Jewelry |
| `18FDh` | You have a fine collection of: |
| `191Ch` |   Gems |
| `1923h` |   Jewelry |
| `192Dh` |  Exit |
| `1933h` | Appraise :  |
| `193Fh` | The Gem is Valued at  |
| `1955h` |  gp. |
| `195Ah` | Sell |
| `195Fh` | Sell Keep |
| `1969h` | You can :  |
| `1974h` | The Jewel is Valued at  |

### overlay-22

| offset | 內容 |
|---|---|
| `018Fh` | Cast |
| `0194h` | Memorize |
| `019Dh` | Scribe |
| `01A4h` | Learn |
| `01CAh` | Choose Spell:  |
| `10BFh` | Cast Spell on whom |
| `11ECh` | can't be cast here... |
| `1202h` | Lose it?  |
| `120Ch` | That Item |
| `1216h` | is a combat-only item... |
| `122Fh` | Use it?  |
| `1238h` | miscasts |
| `1241h` | casts |
| `1247h` | Abort Spell?  |
| `1255h` | Spell Aborted |
| `1CEEh` | is Blessed |
| `1D20h` | is Cursed |
| `1DC9h` | is affected |
| `1E02h` | is protected |
| `1E3Ch` | is cold-resistant |
| `1EB0h` | is unaffected |
| `1EBEh` | is charmed |
| `1F86h` | is stronger |
| `1F92h` | is unaffected |
| `20D0h` | has been reduced |
| `2174h` | is friendly |
| `2226h` | is shielded |
| `22A7h` | falls asleep |
| `23EAh` | is held |
| `2449h` | is fire resistant |
| `2488h` | is silenced |
| `24C1h` | is affected |
| `2575h` | is charmed |
| `2675h` | is invisible |
| `26AFh` | Knock-Knock |
| `26E8h` | is duplicated |
| `2748h` | is weakened |
| `2781h` | Creates a noxious cloud |
| `2DAAh` | is animated |
| `2F56h` | can see |
| `2F9Bh` | is blind |
| `3087h` | is diseased |
| `3172h` | is affected |
| `3522h` | is praying |
| `3576h` | is un-cursed |
| `3583h` | has an item un-cursed |
| `3696h` | has been cursed! |
| `36D4h` | is blinking |
| `3913h` | is Hasted |
| `3CE3h` | is Slowed |
| `3D1Bh` | is restored |
| `3ED8h` | is Speedy |
| `3F63h` | is stronger |
| `4036h` | is paralyzed |
| `4070h` | is Healed |
| `40C8h` | is invisible |
| `4178h` | is unpoisoned |
| `4186h` | is unaffected |
| `42FFh` | smashes them flat |
| `4426h` | is affected |
| `44C9h` | is raised |
| `459Eh` | is slain |
| `45A7h` | is unaffected |
| `4665h` | is entangled |
| `471Dh` | is highlighted |
| `474Ch` | is invisible |
| `4786h` | is charmed |
| `4816h` | is confused |
| `48DAh` | teleports |
| `4A72h` | runs in terror |
| `4A81h` | is unaffected |
| `4BB3h` | flame type:  |
| `4BC0h` | Hot Cold |
| `4BC9h` | is protected |
| `4BD7h` | Abort spell?  |
| `4BE5h` | Yes No |
| `4D8Ch` | is clumsy |
| `4D96h` | is slowed |
| `4EEAh` | is protected |
| `4F24h` | Creates a poisonous cloud |
| `5781h` | is Healed |
| `57D9h` | Breathes! |
| `5953h` | Spits Acid |
| `595Eh` | Spits Acid and Misses |
| `5A9Ah` | breathes acid |
| `5C78h` | breathes fire |
| `5E03h` | Breathes Fire |
| `5F1Eh` | throws lightning |
| `600Ch` | gazes... |
| `6015h` | is paralyzed |
| `61F4h` | Casts a Spell |
| `6202h` | Spell: |

### overlay-23

| offset | 內容 |
|---|---|
| `0E03h` | starts to cough |
| `0E13h` | chokes and gags from nausea |
| `0E2Fh` | is Poisoned |
| `0E3Bh` | is killed |
| `163Ah` | is Cured |
| `1F19h` | takes  |
| `1F20h` |  points of damage  |
| `1F33h` | takes 1 point of damage  |
| `1F4Ch` | from Fire |
| `1F56h` | from Cold |
| `1F60h` | from Electricity |
| `1F71h` | from Acid |
| `1F7Bh` | from Magic |
| `1F86h` | lost a spell |
| `1F93h` | Goes Down |
| `1F9Dh` | , and is Dying |
| `1FCCh` | is killed |
| `22F9h` | is Unaffected |
| `24CCh` | stands up and grins |
| `24E0h` | gets back up |

### overlay-24

| offset | 內容 |
|---|---|
| `0452h` |  Yes   |
| `0459h` |  No    |
| `0789h` | Name |
| `078Eh` | AC  HP |
| `0A1Ch` | Hitpoints |
| `0A29h` | (Helpless) |
| `15C4h` | Nil Item pointer... |
| `15D8h` | Tried to Lose item & couldn't find it! |
| `251Fh` | is fully healed |
| `252Fh` | is partially healed |
| `286Eh` | Guarding |
| `2B99h` |  camping |
| `2BA2h` |  search |
| `2E3Bh` |  Exit |
| `2E63h` | Select |
| `337Bh` | is bandaged |

### overlay-25

| offset | 內容 |
|---|---|
| `0EA1h` | Pick New Class |
| `0EB0h` |  doesn't qualify. |
| `0EC3h` | Select |
| `0ECAh` |  is now a 1st level  |

### overlay-26

| offset | 內容 |
|---|---|
| `0DE7h` |  Next |
| `0DEDh` |  Prev |
| `0DF3h` |  Exit |
| `1103h` | Yes No |

### overlay-29

| offset | 內容 |
|---|---|
| `00D8h` | Loading...Please Wait |
| `00F2h` | FINAL |
| `00F8h` | .dax |
| `00FDh` | PIC not found |
| `05C4h` | HEAD |
| `05C9h` | head not found |
| `05D8h` | BODY |
| `0735h` | Illegal range in Show3DSprite. |
| `080Ch` | bigpic |

### overlay-30

| offset | 內容 |
|---|---|
| `0FFFh` | WALLDEF |
| `1007h` | .dax |
| `100Ch` | Unable to load  |
| `101Ch` |  from WALLDEF |
| `1314h` | .dax |
| `1319h` | Unable to load geo in Load3DMap. |

### overlay-33

| offset | 內容 |
|---|---|
| `000Ch` | Start range error in Load24x24Set |
| `01BCh` | CHEAD |
| `01C2h` | CBODY |
| `01C8h` | COMSPR |
| `01CFh` | ICON |

### overlay-34

| offset | 內容 |
|---|---|
| `0007h` | EXIT |
| `000Ch` | GOTO |
| `0011h` | GOSUB |
| `0017h` | COMPARE |
| `0023h` | SUBTRAT |
| `002Bh` | DIVIDE |
| `0032h` | MULTIPLY |
| `003Bh` | RANDOM |
| `0042h` | SAVE |
| `0047h` | LOAD CHARACTER |
| `0056h` | LOAD MONSTER |
| `0063h` | SETUP MONSTER |
| `0071h` | APPROACH |
| `007Ah` | PICTURE |
| `0082h` | INPUT NUMBER |
| `008Fh` | INPUT STRING |
| `009Ch` | PRINT |
| `00A2h` | PRINTCLEAR |
| `00ADh` | RETURN |
| `00B4h` | COMPARE AND |
| `00C0h` | VERTICAL MENU |
| `00CEh` | IF =  |
| `00D4h` | IF <> |
| `00DAh` | IF < |
| `00DFh` | IF > |
| `00E4h` | IF <= |
| `00EAh` | IF >= |
| `00F0h` | CLEARMONSTERS |
| `00FEh` | PARTYSTRENGTH |
| `010Ch` | CHECKPARTY |
| `0117h` | NEWECL |
| `011Eh` | LOAD FILES |
| `0129h` | LOAD PIECES |
| `0135h` | PARTY SURPRISE |
| `0144h` | SURPRISE |
| `014Dh` | COMBAT |
| `0154h` | ON GOTO |
| `015Ch` | ON GOSUB |
| `0165h` | TREASURE |
| `0172h` | ENCOUNTER MENU |
| `0181h` | GETTABLE |
| `018Ah` | HORIZONTAL MENU |
| `019Ah` | PARLAY |
| `01A1h` | CALL |
| `01A6h` | DAMAGE |
| `01B4h` | SPRITE OFF |
| `01BFh` | FIND ITEM |
| `01C9h` | PRINT RETURN |
| `01D6h` | ECL CLOCK |
| `01E0h` | SAVE TABLE |
| `01EBh` | ADD NPC |
| `01F3h` | PROGRAM |
| `01FFh` | DELAY |
| `0205h` | SPELL |
| `020Bh` | PROTECTION |
| `0216h` | CLEAR BOX |
| `0220h` | DUMP |
| `0225h` | FIND SPECIAL |
| `0232h` | DESTROY ITEMS |

### overlay-35

| offset | 內容 |
|---|---|
| `0027h` | 8x8d |
| `002Ch` | Unable to load  |
| `003Ch` |  from 8x8D |
| `0150h` | Bad symbol number in Put8x8Symbol. |

## pc98（31 個模組，843 條）

### overlay-00

| offset | 內容 |
|---|---|
| `0000h` | ゲームをセーブしてください |

### overlay-01

| offset | 內容 |
|---|---|
| `0051h` | BASED ON THE TSR NOVEL "AZURE BONDS" |
| `007Ah` | Kate Novak |
| `0089h` | Jeff Grubb |
| `0094h` | SCENARIO CREATED BY: |
| `00A9h` | TSR,INC. |
| `00B6h` | Jeff Grubb,George Mac Donald |
| `00D3h` | GAME CREATED BY: |
| `00E4h` | SSI SPECIAL PROJECTS |
| `00F9h` | PROJECT LEADER: |
| `0109h` | George Mac Donald |
| `011Bh` | PROGRAMMING: |
| `0128h` | Scot Bayless |
| `0135h` | Russ Brown |
| `0140h` | Michael Mancuso |
| `0150h` | DEVELOPMENT: |
| `015Dh` | David Shelley |
| `016Bh` | Oran Kangas |
| `0177h` | GRAPHIC ARTS: |
| `0185h` | Tom Wahl |
| `018Eh` | Fred Butts |
| `0199h` | Susan Manley |
| `01A6h` | Mark Johnson |
| `01B3h` | Cyrus Lum |
| `01BDh` | PLAYTESTING: |
| `01CAh` | Jim Jennings |
| `01D7h` | James Kucera |
| `01E4h` | Rick White |
| `01EFh` | Robert Daly |
| `01FBh` | GAME CONVERTED BY PONYCANYON |
| `0218h` |  ＳＰＥＣＩＡＬ　ＰＲＯＪＥＣＴＳ　ＶＥＲＳＩＯＮ１．２ |
| `0250h` |  PONYCANYON INC. |
| `0261h` |  Kunihiko Kagawa |
| `0272h` |  Yoshiaki Matsumoto |
| `0286h` |  Group SNE |
| `0291h` |  Hitoshi Yasuda |
| `02A1h` |  Miyuki Kiyomatsu |
| `02B3h` |  S.R.S. |
| `02BBh` |  Seishi Yokota |
| `02CAh` |  MUSIC COMPOSED BY |
| `02DDh` |  Takeshi Yasuda |
| `02EDh` | Marionette Inc. |
| `02FDh` | Yoshiaki Sakaguchi |
| `0310h` | Masato Kobayashi |
| `0930h` | title |

### overlay-02

| offset | 內容 |
|---|---|
| `00A2h` | ┴t.ﾄ> |
| `0488h` | 読み込み中です |
| `0497h` | CPIC |
| `10A3h` | リターン・キーを押してください |
| `1BBDh` | ITEM |
| `1BC2h` | .dax |
| `1BC7h` | アイテム・ファイルが見つかりません |
| `2208h` | お互いに様子を見ている |
| `221Fh` | 相手は逃げた |
| `2AB7h` | パーティーは全滅した！ |
| `335Fh` | データの整理をします リターン・キーを押してください |
| `39EDh` | データの整理をします リターン・キーを押してください |

### overlay-03

| offset | 內容 |
|---|---|
| `001Eh` | tiles |
| `0024h` | 次のエスプルーアー文字とデテーク文字とを合わせてください。 |
| `005Fh` | －・・－・・－・・ |
| `0072h` | －　－　－　－　－ |
| `0085h` | ‥‥‥‥‥‥‥‥‥ |
| `0098h` | で示された経路の、 |
| `00ABh` | 番のマスには何とありますか？ |
| `00CAh` | １文字入力してリターン・キーを押してください |
| `00F7h` | 間違ってます |
| `0104h` | 目に見えぬ力が、きみを〈奈落〉に放りこんだ！ |

### overlay-04

| offset | 內容 |
|---|---|
| `0034h` | それでもかけてもらいますか？ |
| `00B4h` | は、 |
| `00B9h` | ｇｐでかけてあげましょう。 |
| `00D4h` | かけてもらいますか？ |
| `00E9h` | お金が足りません |
| `00FAh` | は治った |
| `0273h` | は盲目ではありませんよ。 |
| `028Ch` | キュア・ブラインドネス |
| `0329h` | は病いに冒されてはいませんよ。 |
| `0348h` | キュア・ディジーズ |
| `0434h` | キュア・ライト・ウーンズ |
| `044Dh` | キュア・シリアス・ウーンズ |
| `0468h` | キュア・クリティカル・ウーンズ |
| `0487h` | ヒール |
| `0671h` | は死んではいませんよ |
| `0686h` | は生き返りませんよ |
| `0699h` | レイズ・デッド |
| `090Ah` | は毒を受けてはいませんよ |
| `0923h` | ニュートラライズ・ポイズン |
| `0A0Bh` | は呪われてはいませんよ |
| `0A22h` | リムーブ・カース |
| `0AFDh` | は石になってはいませんよ |
| `0B16h` | ストーン・トゥ・フレッシュ |
| `0BBEh` | さん。どのような治療をお望みですか？ |
| `0BE3h` |  治療  |
| `0BEAh` | やめる |
| `0BF2h` | 治療　やめる |
| `0E30h` |      治療  ｷｬﾗｸﾀｰ 金戻す 金集ﾒﾙ 金分ｹﾙ    見積る                      やめる     |
| `0E81h` |      治療  ｷｬﾗｸﾀｰ        金集ﾒﾙ           見積る                      やめる     |
| `0ED3h` |  はい 　いいえ |
| `0EE2h` | きみたちが行こうとすると、僧侶が呼び止めた。「お金をお忘れですよ」 |
| `0F25h` | 引き返してお金を取りますか？ |

### overlay-05

| offset | 內容 |
|---|---|
| `0990h` | パーティーは逃げた。 |
| `09A5h` | きみたちは戦いに敗れた。 |
| `09BEh` | きみは決闘に勝った。 |
| `09D3h` | パーティーは勝った。 |
| `09E8h` | パーティーは宝物を見つけた！ |
| `0A05h` | 決闘者は |
| `0A0Eh` | 各キャラクターは |
| `0A1Fh` | 点の経験点を得た。 |
| `0A32h` | リターン・キーを押してください           |
| `0CD4h` |  取る  |
| `0CDBh` | アイテム |
| `0CE4h` | 取る |
| `0EDFh` | 取る： |
| `0EE6h` |    お金   ｱｲﾃﾑ                やめる     |
| `1018h` | ｷｬﾗｸﾀｰ |
| `101Fh` | 金集ﾒﾙ |
| `1026h` | 抜ける |
| `102Dh` | ﾃﾞｨﾃｸﾄ |
| `1034h` |  取る  |
| `103Bh` | 金分ｹﾙ |
| `1042h` | まだ宝物が残っている。 |
| `1059h` | 戻って宝物を取りますか？ |
| `14BDh` | は自分の分け前を取り、隠した。 |
| `14DCh` | リターン・キーを押してください           |
| `172Dh` | モンスターはパーティーを全滅させ、喜んでいる。 |
| `175Ch` | 何かキーを押してください |

### overlay-06

| offset | 內容 |
|---|---|
| `0025h` |  買う  |
| `002Ch` | アイテム： |
| `0393h` | 持ちすぎです |
| `04ADh` | お金が足りません |
| `0676h` |      買う  ｷｬﾗｸﾀｰ 金戻す 金集ﾒﾙ 金分ｹﾙ    見積る                      抜ける     |
| `06C7h` |      買う  ｷｬﾗｸﾀｰ        金集ﾒﾙ           見積る                      抜ける     |
| `0718h` | きみたちが行こうとすると、店主が呼び止めた。「お金をお忘れですよ」 |
| `075Bh` | 引き返してお金を取りますか？ |

### overlay-07

| offset | 內容 |
|---|---|
| `0034h` | ALIAS |
| `003Ah` | エイリアス |
| `0045h` | DRAGONBAIT |
| `0050h` | ドラゴンベイト |
| `005Fh` | AKABAR BEL AKAS |
| `006Fh` | アーカバー・ベル・アーカッシュ |
| `0481h` | 読みこみ中です |
| `0494h` | .dax |
| `0681h` | SPRIT |
| `1B7Dh` | CPIC |
| `1B82h` | ロルフ |
| `2192h` | ALIAS |
| `2198h` | エイリアス |
| `21A3h` | DRAGONBAIT |
| `21AEh` | ドラゴンベイト |
| `21BDh` | AKABAR BEL AKAS |
| `21CDh` | アーカバー・ベル・アーカッシュ |
| `21F2h` | 死んだ。 |
| `21FBh` | ポイントのダメージを受けた |
| `2216h` | リターン・キーを押してください |

### overlay-08

| offset | 內容 |
|---|---|
| `0400h` | 魔法使用 |
| `0409h` | 魔法不使用 |
| `071Ch` |  移動  |
| `0723h` | ｷｬﾗｸﾀｰ |
| `072Ah` |  照準  |
| `0731h` |  使う  |
| `0738h` |  魔ｶｹﾙ |
| `073Fh` |  退散  |
| `0746h` |  自動  |
| `074Dh` | その他 |
| `0979h` | 瀕死のキャラクターがいる |
| `0992h` | 戦闘を続けますか？ |
| `0AE5h` | 移動／攻撃，残り移動力＝ |
| `0B02h` | 逃げますか？ |
| `0B0Fh` | そこへは行けない |
| `0F06h` | そこへは進めない |
| `0FCFh` |  防御  |
| `0FD6h` |  待機  |
| `0FDDh` |  完了  |
| `0FE4h` | 手当て |
| `0FEBh` |  速度  |
| `0FF2h` | 抜ける |
| `1019h` | 効果がなかった |
| `122Ch` | 速さ（ |
| `1236h` | 抜ける |
| `123Fh` |  遅く  |
| `1246h` |  速く  |

### overlay-09

| offset | 內容 |
|---|---|
| `003Eh` | は恐慌にとらわれ、逃げ出した。 |
| `09CDh` | 移動／攻撃，残り移動力＝ |
| `1282h` | 魔法使用 |
| `128Bh` | 魔法不使用 |
| `1385h` | は退散させられた。 |
| `1398h` | は降伏した。 |

### overlay-10

| offset | 內容 |
|---|---|
| `100Bh` | DungCom |
| `1013h` | WildCom |
| `101Bh` | RandCom |
| `1C19h` | 戦いが始まった |

### overlay-11

| offset | 內容 |
|---|---|
| `002Fh` | 読みこみ中です |
| `0042h` | ITEMS |
| `0501h` | COMSPR |

### overlay-12

| offset | 內容 |
|---|---|
| `00D0h` | は蛇と戦っている。 |
| `038Ah` | は窒息した。 |
| `04CFh` | は沈黙させられた。 |
| `0537h` | は毒のために死んだ。 |
| `05AAh` | はアイテムを得た。 |
| `0762h` | は幻影を失った。 |
| `0816h` | は咳きこんでいる。 |
| `08D4h` | は倒れた。 |
| `09F2h` | は逃げ去った。 |
| `0A01h` | は混乱した。 |
| `0A0Eh` | は狂暴化した。 |
| `0A1Dh` | は怒りに我を忘れた。 |
| `0BFDh` | は歳をとった。 |
| `0C69h` | 空気は多少きれいになった |
| `0F50h` | は避けた。 |
| `1084h` | は弱まった。 |
| `1460h` | は毒を受けた。 |
| `146Fh` | は死んだ。 |
| `1512h` | は麻痺した。 |
| `15F5h` | は愚かになった。 |
| `1606h` | は呪文を唱えきれなかった。 |
| `1756h` | は影響を受けなかった。 |
| `1840h` | は狂暴化した。 |
| `1A8Bh` | は睨んだ。 |
| `1A96h` | は視線をはねかえした！ |
| `1AADh` | は石になった。 |
| `1D45h` | 空気は多少きれいになった |
| `2099h` | は倒れ、死んだ。 |
| `26AEh` | は酸を吹いた。 |
| `27E3h` | は麻痺した。 |
| `28F4h` | は睨んだ。 |
| `28FFh` | は麻痺した。 |
| `2D45h` | は報復を受けた。 |
| `2DD9h` | は異界に送り返された。 |
| `2DF0h` | は「ディスペル・イービル」に耐えた。 |

### overlay-13

| offset | 內容 |
|---|---|
| `0256h` | はバックスタッブを行ない、　 |
| `0273h` | は無力な |
| `027Ch` | は攻撃し、 |
| `028Ah` | の背後 |
| `0291h` | 無慈悲な一撃を加えた。 |
| `02A8h` | 命中して |
| `02B1h` | ポイントのダメージを与えた。 |
| `02CEh` | はダメージを受けなかった。 |
| `02E9h` | にかわされた。 |
| `02F8h` | は呪文を無駄にした。 |
| `030Dh` | は倒れた。 |
| `0318h` | 瀕死の状態だ。 |
| `0347h` | そして、死んだ。 |
| `0D61h` | は逃げきった。 |
| `0D70h` | 逃げられなかった |
| `0FA4h` | なぎ払った。 |
| `1258h` | アンデッドの退散を試みた。 |
| `1273h` | は退散させられた。 |
| `1286h` | は破壊された。 |
| `1295h` | 何も起こらなかった |
| `2244h` | すでに目標として選んである |
| `2773h` | キャンプでのみ使う呪文だ |
| `278Ch` | は呪文を唱え始めた。 |
| `2DA0h` | 味方を攻撃するのですか？ |
| `2E86h` | 距離＝ |
| `2E90h` |  決定  |
| `2E97h` |  次順  |
| `2E9Eh` |  前順  |
| `2EA5h` |  手動  |
| `2EACh` |  中央  |
| `2EB3h` | やめる |
| `3261h` | 距離＝ |
| `326Bh` |  中央  |
| `3272h` | やめる |
| `3279h` |  決定  |
| `3280h` | （テン・キーを使ってください） |
| `3F62h` | を包みこんだ。 |
| `41E4h` | は原子分解光線を発射した。 |
| `41FFh` | は粉々になった。 |
| `4210h` | は石化光線を発射した。 |
| `4227h` | は石になった。 |
| `4236h` | は殺人光線を発射した。 |
| `424Dh` | は死んだ。 |
| `4258h` | はきみを傷つけた。 |
| `477Ch` | を絞めつけた。 |
| `4867h` | 神々の介入だ！ |

### overlay-14

| offset | 內容 |
|---|---|
| `08D3h` | ここではできない |
| `0B9Fh` | 蹴破る |
| `0BA6h` | 鍵開け |
| `0BADh` | ノック |
| `0BB4h` | やめる |
| `0BBBh` | 鍵がかかっている |

### overlay-15

| offset | 內容 |
|---|---|
| `0327h` | ここでは、魔法が使えない |
| `0340h` | 状態ではない |
| `034Dh` | は呪文を唱えられる |
| `0360h` | は魔法を記憶できる |
| `0373h` | は魔法を書き写せる |
| `04A3h` | は呪文をまったく覚えていない。 |
| `057Ah` | は次の数だけ呪文を記憶できる。 |
| `0599h` | 　　　僧侶魔法： |
| `05AAh` | ドルイド僧魔法： |
| `05BBh` | 　魔法使い魔法： |
| `08F9h` | これらの呪文を覚えますか？ |
| `0914h` | は呪文の記憶はできない。 |
| `0AE9h` | これらの呪文を書き写しますか？ |
| `0B08h` | は書き写せる巻物を持っていない。 |
| `0B29h` | その魔法はすでに知っている。 |
| `0B46h` | その魔法はすでに書き写そうとしている。 |
| `0B6Dh` | その魔法は書き写せない。 |
| `0ED9h` | 怯えている |
| `0EE4h` | 悪の退散 |
| `0EEDh` | フェアリー・ファイア |
| `0F02h` | 遅鈍 |
| `0F07h` | 無力 |
| `0F0Ch` | 混乱 |
| `0F11h` | 病い |
| `0F16h` | 熱いファイア・シールド |
| `0F2Dh` | 冷たいファイア・シールド |
| `0F49h` | 再生中 |
| `0F50h` | 炎からの防護 |
| `0F5Dh` | 低呪文封じ |
| `0F68h` | 暗愚 |
| `0F6Dh` | 動物に対する透明 |
| `0F7Eh` | 透明 |
| `0F83h` | 偽装 |
| `0F88h` | 竜の息吹よりの防護 |
| `0F9Bh` | 狂暴化 |
| `0FA2h` | 位置のごまかし |
| `0FB1h` | 〈魔法の影響はない〉 |
| `17C2h` | 並び順： |
| `17EBh` | を選んだ。 |
| `191Eh` | ゲームを終わりますか？ |
| `1935h` | は完全にいなくなる。 |
| `194Ah` | かまわず離脱させますか？ |
| `1963h` | は別れを告げた。 |
| `1974h` | はドブに捨てられた。 |
| `1989h` | は安堵のため息をもらした。 |
| `1B38h` | 速さ＝ |
| `1B3Fh` | 　（０＝速い　９＝遅い） |
| `1B58h` |  速く  |
| `1B5Fh` |  遅く  |
| `1B66h` | 抜ける |
| `1B6Dh` | 速度： |
| `1D0Ah` |  順序  |
| `1D11h` |  離脱  |
| `1D18h` |  速度  |
| `1D1Fh` |  ｱｲｺﾝ  |
| `1D26h` |  ｱﾆﾒ無 |
| `1D2Dh` |  ｱﾆﾒ有 |
| `1D34h` | 抜ける |
| `1D3Bh` | は変更できない。 |
| `2414h` | パーティーはキャンプを張った。 |

### overlay-16

| offset | 內容 |
|---|---|
| `014Fh` | .SAV |
| `05FCh` | *.guy |
| `0602h` | *.cha |
| `0608h` | *.sav |
| `060Eh` | *.hil |
| `07D6h` | *.hil |
| `0839h` |    はい                       いいえ     |
| `0862h` | セーブディスクを作成しますか？  |
| `0882h` | セーブディスクを作成中です。 |
| `089Fh` | セーブディスクの作成に失敗しました。リターン・キーを押してください。 |
| `0AA9h` | ディスクに異常があります。 |
| `0AC4h` | ドライブ２にセーブ・ディスクを入れてください |
| `0AF1h` | セーブ中にエラーが発生しました |
| `0B30h` | ロード中にエラーが発生しました |
| `1055h` | CHEAD |
| `105Bh` | CBODY |
| `1253h` | .guy |
| `1258h` | .swg |
| `1382h` | ディスクを交換してください。 |
| `139Fh` | ディスクに余裕がありませんので、セーブできません |
| `13D0h` | キャラクターをあきらめますか？ |
| `13EFh` |    ｱｷﾗﾒﾙ                      別ﾃｨｽｸ     |
| `1418h` | .guy |
| `141Fh` | .sav |
| `1424h` | 同名のキャラクターがいます。 |
| `1441h` | 上書きしますか？ |
| `1452h` | 新しい名前は？ |
| `1461h` | ライトプロテクトをはずしてください。 |
| `1486h` | ディスクに異常があります。 |
| `14A1h` | セーブ中にエラーが発生しました |
| `14C0h` | .swg |
| `29EDh` | ドライブ２にセーブ・ディスクを入れてください |
| `2A1Ah` | 読みこみ中です |
| `2A29h` | .GUY |
| `2A2Eh` | .CHA |
| `2A33h` | .guy |
| `2A3Ah` | .cha |
| `2A3Fh` | .swg |
| `2A48h` | .spc |
| `3A6Bh` | MONCHA |
| `3A72h` | .dax |
| `3A77h` | モンスターをロードできませんでした |
| `3A9Ah` | MONSPC |
| `3AA1h` | MONITM |
| `3DFAh` | CPIC |
| `4159h` | SaveList.EST |
| `4166h` | ドライブ２にセーブ・ディスクを入れてください |
| `427Dh` | savgam |
| `4284h` | .dat |
| `4290h` | どのデータを読みこみますか？ |
| `42CEh` | パーティーは |
| `42DBh` | ます。 |
| `42E2h` | にいます。 |
| `42EDh` | よろしいですか？ |
| `4602h` |       ①     ②     ③     ④     ⑤        ⑥     ⑦     ⑧     ⑨     ⑩       |
| `4653h` | どこにセーブしますか？ |
| `468Dh` | 番は既に登録されています。 |
| `46A8h` | よろしいですか？ |
| `46B9h` | ゲームを終わりますか |
| `4972h` | savgam |
| `4979h` | .dat |
| `497Eh` | ドライブ２にセーブ・ディスクを入れてください |
| `49ABh` | 読みこみ中です |
| `49BCh` | .sav |
| `49C1h` | CPIC |
| `4F42h` | savgam |
| `4F49h` | .dat |
| `4F4Eh` | ディスクに余裕がありませんので、セーブできません |
| `4F9Fh` | ライトプロテクトをはずしてください。 |
| `4FC4h` | セーブ中にエラーが発生しました－ |
| `4FE5h` | SaveList.EST |
| `4FF2h` | 書きこみ中です |
| `5001h` | CHRDAT |

### overlay-17

| offset | 內容 |
|---|---|
| `01A1h` | 誰を消去しますか？ |
| `01B4h` | 誰を調整しますか？ |
| `01C7h` | 誰を訓練しますか？ |
| `01DAh` | 誰を変更しますか？ |
| `01EDh` | このキャラクターはクラス変更できません。 |
| `0216h` | 誰の詳細を見ますか？ |
| `022Bh` | 誰を離脱させますか？ |
| `0240h` | きみたちは勝利した。 |
| `0255h` | ゲームを終わる前にセーブしておきますか？ |
| `027Eh` | ゲームを終わりますか？ |
| `0295h` | データがセーブされていません。 |
| `02B4h` | それでも終わりますか？ |
| `09F5h` | 種族を選んでください |
| `0A0Dh` |  決定  |
| `0A14h` | 決定 |
| `0A19h` | 性別を選んでください |
| `0A2Eh` | クラスを選んでください |
| `0A45h` | Select |
| `0A4Ch` | アラインメントを選んでください |
| `0A8Dh` | 能力値を決めなおしますか？ |
| `0AA8h` | キャラクターの名前は？ |
| `0ABFh` | このキャラクタ名は受け付けられません。 |
| `0AE6h` | をセーブしますか？ |
| `2813h` | ALIAS |
| `2819h` | エイリアス |
| `2824h` | DRAGONBAIT |
| `282Fh` | ドラゴンベイト |
| `283Eh` | AKABAR BEL AKAS |
| `284Eh` | アーカバー・ベル・アーカッシュ |
| `286Dh` | は永遠にいなくなりますよ？ |
| `288Ah` | かまわないのですね？ |
| `289Fh` | きみたちは |
| `28AAh` | を捨てた |
| `28B3h` | はさようならを言った |
| `28C8h` | は安堵のため息をもらした |
| `2ADEh` |    はい                       いいえ     |
| `2BAAh` | 名前を変更しますか？ |
| `2BBFh` |                 |
| `2BCFh` | 新しい名前は？ |
| `2BDEh` | このキャラクタ名は受け付けられません。 |
| `2C05h` | これでよろしいですか？ |
| `2E52h` | の数値は変更できない |
| `2E67h` | 数値調整： |
| `2E72h` |    決定　                     やめる     |
| `3AF3h` | どこから加えますか？ |
| `3B08h` |   カース プール ﾋﾙｽﾞﾌｧ        やめる     |
| `3B31h` | 加える |
| `3B38h` | キャラクター参加： |
| `3B4Bh` | 加える　やめる |
| `3B5Dh` | 聖騎士は邪悪なクズとは同行しない |
| `3B7Eh` | レンジャーの数が多すぎます |
| `3B99h` | は邪悪を許さない |
| `3BAAh` |                 |
| `452Dh` | 古いアイコン |
| `453Ah` |  準備状態　　　　　行動状態 |
| `4556h` |  新しいアイコン |
| `4566h` |   小   |
| `456Dh` |   大   |
| `45B5h` | このアイコンでいいですか？ |
| `5332h` | 意識のあるかたでないと訓練はできません |
| `5359h` | 訓練料は１０００ｇｐです |
| `5372h` | ここではそのクラスの訓練はできません |
| `5397h` | あなたは十分に成長しています。 |
| `53B6h` | もうこれ以上訓練はできません。 |
| `53D5h` | 経験が不足しています |
| `53EDh` | レベルの |
| `53F8h` | になれます |
| `5406h` | 訓練しますか？ |
| `5415h` | おめでとう |

### overlay-18

| offset | 內容 |
|---|---|
| `0B8Ch` | .DAX |
| `0B91h` | CED9.DAX |
| `0D13h` | ティランスラクサスの霊体がストーム・ジャイアントの死体に重なった。 |
| `0D56h` | 「よくぞ余を倒した。『ラサンダーの魔除け』さえなければ、汝らの身体を乗っ取り、 |
| `0DA5h` | 勝利をもぎとることもできたのだが。 |
| `0DC8h` | しかし、まだ〈プール〉を通じて逃げることはできる」 |
| `0DFBh` | 何かキーを押してください |
| `0E14h` | きみたちが〈プール・オブ・レイディアンス〉に近づくと、ティランスラクサスは叫んだ。 |
| `0E67h` | 「『モーンダーの篭手』を近づけるな！　危険なエネルギーを解放しようとしているのだぞ！ |
| `0EBCh` | 　篭手が〈プール〉に触れた瞬間、〈プール〉は収縮し、篭手は砕けた。 |
| `0F00h` | 「余は、逃れることができぬか。 |
| `0F1Fh` | 汝らは軍隊でも成しえなかったことを成し遂げたわけだ。 |
| `0F54h` | 喜ぶがよい。笑みを浮かべるがよい。 |
| `0F77h` | この日、ティランスラクサスは滅した」　きみたちの前で、ティランスラクサスは縮み、姿を消した。 |
| `0FD4h` | ティランスラクサスが倒れたことは間違いがない。 |
| `1003h` | というのも、きみたちの最後の〈紺青の呪縛〉がついに消え去ったからだ。 |
| `1048h` | きみたちはついに、解放された！ |
| `1067h` | ミス・ドラノールの騎士たちが部屋になだれこんできた。 |
| `109Ch` | 「おめでとう。きみたちはついに『炎のもの』を打ち倒したな |
| `10D5h` | エルミンスターの力をもって、この忌まわしい場所から脱出させてあげよう。 |
| `111Ch` | そして、宴だ」 |
| `112Bh` | きみたちはシャドウデイルへとテレポートした。 |
| `1158h` | そこでは、すでに祭宴が始まっていた。きみたちが着くと、大きな喚声が巻き起こった。 |
| `11A9h` | 近くのさじきから、ジャーリーとナカシアが腕を組み、きみたちにおめでとうを叫んだ。 |
| `11FAh` | きみたちは勝利したのだ！ |

### overlay-19

| offset | 內容 |
|---|---|
| `0039h` | （ＮＰＣ） |
| `0044h` | 年齢  |
| `004Ah` | 体力度 |
| `0051h` | 知識度 |
| `0058h` | 賢明度 |
| `005Fh` | 敏捷度 |
| `0066h` | 耐久度 |
| `006Dh` | 魅力度 |
| `0074h` | 経験レベル |
| `0081h` | 経験点  |
| `0089h` | 武器 |
| `008Eh` |  鎧  |
| `0093h` | 状態 |
| `066Eh` | ＡＣ |
| `0673h` | ＨＰ |
| `0678h` | ＴＨＡＣ０ |
| `0689h` | ダメージ |
| `0692h` | 荷重 |
| `0697h` | 移動力 |
| `09B1h` | ００ |
| `0B31h` |  ｱｲﾃﾑ  |
| `0B38h` |  魔法  |
| `0B3Fh` |  渡す  |
| `0B46h` | 捨てる |
| `0B4Dh` |  回復  |
| `0B54h` |  治療  |
| `0B5Bh` | 抜ける |
| `0E23h` | 装備をはずしてください |
| `0E3Ah` | はその巻物から呪文を書き写そうとしています。 |
| `0E67h` | この巻物からの書き写しを中止しますか？ |
| `0F85h` | itemptr:       |
| `0F94h` | namenum(1):    |
| `0FA3h` | namenum(2):    |
| `0FB2h` | namenum(3):    |
| `0FC1h` | plus:          |
| `0FD0h` | plussave:      |
| `0FDFh` | ready:         |
| `0FEEh` | identified:    |
| `0FFDh` | cursed:        |
| `100Ch` | value:         |
| `101Bh` | special(1):    |
| `102Ah` | special(2):    |
| `1039h` | special(3):    |
| `1048h` | dice large:     |
| `1058h` | sides large:    |
| `1068h` | press a key |
| `14DCh` |  準備  |
| `14E3h` |  使う  |
| `14EAh` |  渡す  |
| `14F1h` | 捨てる |
| `14F8h` | 分ける |
| `14FFh` |  ﾏﾄﾒﾙ  |
| `1506h` |  売る  |
| `150Dh` |  鑑定  |
| `1514h` | アイテム |
| `151Dh` | 準備する　アイテム |
| `1530h` | 装備をしてください |
| `1543h` | あなたの |
| `154Ch` | は永遠になくなります。 |
| `1563h` | それでも捨てますか？ |
| `1E65h` | 呪われています |
| `1E74h` | クラスが一致しません |
| `1E89h` | すでに |
| `1E90h` | を使用中です |
| `1E9Dh` | 手がいっぱいです |
| `20E4h` | 誰に渡しますか？ |
| `20F5h` | 持ちすぎです |
| `21A8h` | それは分けることはできません |
| `2446h` | はアイテムを使った。 |
| `245Bh` | アイテム： |
| `2466h` | うわっ！ |
| `2784h` | ｇｐで引き取ろう。 |
| `2797h` | それでいいかね？ |
| `27A8h` | 毎度ありがとう |
| `27B7h` | 持ちすぎだ。お金は地面に置いておくよ |
| `29BEh` | の鑑定料は２００ｇｐだ。 |
| `29D7h` | それでいいかね？ |
| `29E8h` | お金が足りません |
| `29F9h` | その |
| `29FEh` | については、特に目新しい事実はないですな |
| `2A27h` | のようです |
| `2C14h` | 誰にわたしますか？ |
| `2C2Ah` |  決定  |
| `2C31h` | 貨幣の種類を選んでください |
| `2C4Ch` | 決定 |
| `2C51h` | どれだけ |
| `2C5Ah` | わたしますか？ |
| `2FA8h` |  決定  |
| `2FAFh` | 貨幣の種類を選んでください |
| `2FCAh` | 決定 |
| `2FCFh` | どれだけ |
| `2FD8h` | 捨てますか？ |
| `347Fh` | 記憶内 |
| `3486h` | 魔法書上 |
| `348Fh` | 巻物上 |
| `3496h` | 選んでください |
| `34A5h` | 記憶する魔法を選んでください |
| `34C2h` | 書きこむ魔法を選んでください |
| `34DFh` | 呪文 |
| `3716h` | 誰を回復させますか？ |
| `372Bh` | ALIAS |
| `3731h` | エイリアス |
| `373Ch` | DRAGONBAIT |
| `3747h` | ドラゴンベイト |
| `3756h` | AKABAR BEL AKAS |
| `3766h` | アーカバー・ベル・アーカッシュ |
| `3785h` | は気分がよくなった |
| `3798h` | には影響がなかった |
| `3920h` | 誰を治療しますか？ |
| `3933h` | は病気にかかっていない |
| `394Ah` | それでも治療しますか？ |
| `3961h` | ALIAS |
| `3967h` | エイリアス |
| `3972h` | DRAGONBAIT |
| `397Dh` | ドラゴンベイト |
| `398Ch` | AKABAR BEL AKAS |
| `399Ch` | アーカバー・ベル・アーカッシュ |
| `39BBh` | は治療を受けた |

### overlay-20

| offset | 內容 |
|---|---|
| `05C7h` | 休息時間 |
| `06C9h` |       日     時     分    休む            増やす 減らす               やめる     |
| `0883h` | パーティーは回復した |
| `0945h` | を覚えた |
| `0A04h` | を書き写した |
| `0C9Ah` | 休息を中断しますか？ |
| `0CAFh` | 休んでいるどころではなくなった！ |

### overlay-21

| offset | 內容 |
|---|---|
| `01F3h` | 持ちすぎです。お金は地面に置きますよ |
| `0478h` | 持ちすぎです |
| `0A99h` | 持ちすぎです |
| `0B74h` | 宝石飾り |
| `0B7Dh` | 宝石 |
| `0B82h` | ｇｐ |
| `0B87h` | ｐｐ |
| `0B8Ch` | ｅｐ |
| `0B91h` | ｓｐ |
| `0B96h` | ｃｐ |
| `0DA5h` |  決定  |
| `0DACh` | 貨幣の種類を選んでください |
| `0DC7h` | 決定 |
| `0DCCh` | どれだけ |
| `0DD5h` | 持っていきますか？ |
| `1948h` | 宝石も宝石飾りも持っていませんよ |
| `1969h` | 宝石 |
| `196Eh` | 宝石飾り |
| `1977h` | あなたのお持ちになっているのは |
| `1996h` |  宝石  |
| `199Dh` | 宝石飾 |
| `19A4h` | 抜ける |
| `19ABh` | 見積る： |
| `19B4h` | 宝石の値段は |
| `19C1h` | ｇｐです。 |
| `19CCh` |    売る                                  |
| `19F5h` |    売る　                      売ﾗﾅｲ     |
| `1A1Eh` | どうしますか？ |
| `1A2Dh` | 宝石飾りの値段は |

### overlay-22

| offset | 內容 |
|---|---|
| `01AFh` | 呪文を選んでください |
| `01C4h` | かける |
| `01CBh` | 記憶ｽﾙ |
| `01D2h` |  書ｷｺﾑ |
| `01D9h` |  学ぶ  |
| `01E0h` |        |
| `111Bh` | 誰にかけますか？ |
| `1394h` | はここではかけられない。 |
| `13ADh` | 無駄に使いますか？ |
| `13C0h` | そのアイテムは |
| `13CFh` | 戦闘専用だ。 |
| `13DCh` | それでも使いますか？ |
| `13F1h` | は呪文を唱えそこねた |
| `1406h` | の呪文をかけた。 |
| `1417h` | 呪文を中断しますか？ |
| `142Ch` | 呪文は中断された |
| `1F42h` | は祝福を受けた。 |
| `1F7Ah` | は呪いを受けた。 |
| `202Ah` | は魔法にかかった。 |
| `206Ah` | は守られた。 |
| `20A4h` | は冷気に対する防護を得た。 |
| `2121h` | は影響を受けなかった。 |
| `2138h` | は魅了された。 |
| `2204h` | は強くなった。 |
| `2213h` | は影響を受けなかった。 |
| `234Dh` | は小さくなった。 |
| `23F1h` | は友好的になった。 |
| `24AAh` | は魔法の楯に守られた。 |
| `2536h` | は眠りに落ちた。 |
| `267Dh` | は金縛りにあった。 |
| `26E1h` | は火炎に対する防護を得た。 |
| `2729h` | は沈黙した。 |
| `2763h` | は魔法にかかった。 |
| `281Eh` | は魅了された。 |
| `2922h` | は透明になった。 |
| `2960h` | こんこん |
| `2996h` | は分身した。 |
| `29F5h` | は弱くなった。 |
| `2A31h` | はひどい臭いのガスを作り出した。 |
| `3063h` | は動き始めた。 |
| `3212h` | の視力は戻った。 |
| `3260h` | は視力を奪われた。 |
| `3356h` | は病いに冒された。 |
| `3448h` | は魔法にかかった。 |
| `37F5h` | は祈っている。 |
| `384Dh` | の呪いはとけた。 |
| `385Eh` | のアイテムの呪いは消えた。 |
| `3976h` | は呪われた！ |
| `39B0h` | は点滅している。 |
| `3BBCh` | は加速された。 |
| `3F99h` | は減速された。 |
| `3FD6h` | はレベルを回復した。 |
| `419Ch` | はすばやくなった。 |
| `4230h` | は強くなった。 |
| `4306h` | は麻痺した。 |
| `4340h` | は回復した。 |
| `439Bh` | は透明になった。 |
| `444Fh` | の毒は拭われた。 |
| `4460h` | は影響を受けなかった。 |
| `45E2h` | は叩きつけた。 |
| `4706h` | は魔法にかかった。 |
| `47B0h` | は蘇生した。 |
| `4888h` | は死んだ。 |
| `4893h` | は影響を受けなかった。 |
| `495Ah` | は絡みつかれた。 |
| `4A16h` | に光がまとわりついた。 |
| `4A4Dh` | は透明になった。 |
| `4A8Bh` | は魅了された。 |
| `4B1Fh` | は混乱した。 |
| `4BE4h` | はテレポートした。 |
| `4D85h` | は恐怖にかられ、逃げ出した。 |
| `4DA2h` | は影響を受けなかった。 |
| `4EDDh` | 炎の種別は？ |
| `4EEAh` |     熱                          冷       |
| `4F13h` | は守られた。 |
| `4F21h` | 呪文を中断しますか？ |
| `4F36h` |    はい                       いいえ     |
| `5125h` | はのろまになった。 |
| `5138h` | は減速された。 |
| `5291h` | は守られた。 |
| `52CBh` | は毒雲を作り出した。 |
| `5B1Eh` | は回復した。 |
| `5B79h` | は致命の息吹を吹いた！ |
| `5D08h` | は酸を吹いた。 |
| `5E23h` | は酸を吹いた。 |
| `600Ah` | は炎を吹いた。 |
| `619Eh` | は炎を吹いた。 |
| `62C2h` | は電光を投げつけた。 |
| `63BCh` | は睨んだ。 |
| `63C7h` | は麻痺した。 |
| `65AEh` | は呪文を唱えた。 |
| `65BFh` | 呪文 |
| `65C4h` | は、 |

### overlay-23

| offset | 內容 |
|---|---|
| `0DE8h` | は咳きこみ始めた。 |
| `0DFBh` | は窒息し、喉をかきむしった。 |
| `0E18h` | は毒を受けた。 |
| `0E27h` | は死んだ。 |
| `1623h` | は呪われた。 |
| `1F12h` | ポイントのダメージを受けた。 |
| `1F2Fh` | は火によって |
| `1F3Ch` | は冷気によって |
| `1F4Bh` | は電撃によって |
| `1F5Ah` | は酸によって |
| `1F67h` | は魔法によって |
| `1F76h` | は呪文を唱えられなくなった。 |
| `1F93h` | は倒れた。 |
| `1F9Eh` | そして、瀕死の状態だ。 |
| `1FD5h` | は死んだ。 |
| `1FE0h` | は、ダメージを受けなかった。 |
| `2310h` | には影響がなかった。 |
| `24EAh` | は立ち上がり、ニヤリと笑った。 |
| `2509h` | は起き上がった。 |

### overlay-24

| offset | 內容 |
|---|---|
| `0452h` |  装備中   |
| `045Ch` |  　　　   |
| `08ABh` | 名前 |
| `08B0h` | ＡＣ　　ＨＰ |
| `08BDh` | ALIAS |
| `08C3h` | エイリアス |
| `08CEh` | DRAGONBAIT |
| `08D9h` | ドラゴンベイト |
| `08E8h` | AKABAR BEL AKAS |
| `08F8h` | アーカバー・ベル・アーカッシュ |
| `0C3Ah` | ヒットポイント |
| `0C49h` | ＡＣ |
| `0C4Eh` | 無力 |
| `177Ah` | アイテム・ポインターがありません |
| `2756h` | は完全に回復した |
| `2767h` | は少し回復した |
| `2AA1h` | 防御している |
| `2E76h` | キャンプ中 |
| `2E81h` | 捜索モード |
| `311Dh` | 抜ける |
| `3124h` |  決定  |
| `35C5h` | は手当てを受けた。 |

### overlay-25

| offset | 內容 |
|---|---|
| `0E89h` | 新しいクラスを選んでください |
| `0EA6h` | ALIAS |
| `0EACh` | エイリアス |
| `0EB7h` | DRAGONBAIT |
| `0EC2h` | ドラゴンベイト |
| `0ED1h` | AKABAR BEL AKAS |
| `0EE1h` | アーカバー・ベル・アーカッシュ |
| `0F00h` | は不適格です |
| `0F0Dh` |  決定  |
| `0F15h` | 決定 |
| `0F1Ah` | は１経験レベルの |
| `0F2Bh` | です |

### overlay-26

| offset | 內容 |
|---|---|
| `0851h` |        |
| `0898h` |      |
| `08A0h` |         |
| `0E9Eh` | 　次　 |
| `0EC5h` | 　前　 |
| `0ECCh` | 抜ける |
| `1254h` |    はい                       いいえ     |

### overlay-29

| offset | 內容 |
|---|---|
| `00D8h` | 読みこみ中です |
| `00EBh` | FINAL |
| `00F1h` | .dax |
| `00F6h` | PIC not found  |
| `0105h` | SPRIT |
| `04C7h` | HEAD |
| `04CCh` | head not found |
| `04DBh` | BODY |
| `064Dh` | Illegal range in Show3DSprite. |
| `0724h` | bigpic |

### overlay-30

| offset | 內容 |
|---|---|
| `0F44h` | WALLDEF |
| `0F4Ch` | .dax |
| `0F51h` | Unable to load  |
| `0F61h` |  from WALLDEF |
| `122Dh` | .dax |
| `1232h` | Unable to load geo in Load3DMap. |

### overlay-33

| offset | 內容 |
|---|---|
| `000Ch` | Start range error in Load24x24Set |
| `002Eh` | DungCom |
| `0036h` | WildCom |
| `003Eh` | RandCom |
| `0046h` | tiles |
| `035Ch` | CHEAD |
| `0362h` | CBODY |
| `0368h` | COMSPR |

### overlay-35

| offset | 內容 |
|---|---|
| `0027h` | 8x8d |
| `002Ch` | .dax |
| `0031h` | Unable to load  |
| `0041h` |  from 8x8D |
| `0307h` | Bad symbol number in Put8x8Symbol. |
