# overlay 內嵌字串（掃描結果）

由 `scripts/scan_pascal_strings.py` 產生。採信一條字串的依據有兩種：
「有 `mov di, offset` 引用」（標 `ref`），或「連續 6 條首尾相接的字串表」（標 `chain`）。
兩者都是保守判準，**這是下界不是全集**：用其他方式取址、又不在連續表
裡的字串不會出現在這裡，不得由缺席推論不存在。

## dos（32 個模組，792 條）

### overlay-00

| offset | 依據 | 內容 |
|---|---|---|
| `0000h` | ref | Time to save your game |

### overlay-01

| offset | 依據 | 內容 |
|---|---|---|
| `0051h` | ref | based on the tsr novel "azure bonds" |
| `007Ah` | ref | kate novak |
| `0089h` | ref | jeff grubb |
| `0094h` | ref | scenario created by: |
| `00A9h` | ref | tsr, inc. |
| `00B7h` | ref | george mac donald |
| `00C9h` | ref | game created by: |
| `00DAh` | ref | ssi special projects |
| `00EFh` | ref | project leader: |
| `00FFh` | ref | programming: |
| `010Ch` | ref | scot bayless |
| `0119h` | ref | russ brown |
| `0124h` | ref | michael mancuso |
| `0134h` | ref | development: |
| `0141h` | ref | david shelley |
| `014Fh` | ref | oran kangas |
| `015Bh` | ref | graphic arts: |
| `0169h` | ref | tom wahl |
| `0172h` | ref | fred butts |
| `017Dh` | ref | susan manley |
| `018Ah` | ref | mark johnson |
| `0197h` | ref | cyrus lum |
| `01A1h` | ref | playtesting: |
| `01AEh` | ref | jim jennings |
| `01BBh` | ref | james kucera |
| `01C8h` | ref | rick white |
| `01D3h` | ref | robert daly |
| `062Eh` | ref | title |

### overlay-02

| offset | 依據 | 內容 |
|---|---|---|
| `0461h` | ref | CPIC |
| `1037h` | ref | PRESS BUTTON OR RETURN TO CONTINUE. |
| `105Bh` | ref | PRESS <ENTER>/<RETURN> TO CONTINUE |
| `1B30h` | ref | ITEM |
| `1B35h` | ref | .dax |
| `1B3Ah` | ref | Unable to find item file |
| `2029h` | ref | ~COMBAT ~WAIT ~FLEE ~PARLAY |
| `2045h` | ref | ~COMBAT ~WAIT ~FLEE ~ADVANCE |
| `2062h` | ref | Both sides wait. |
| `2073h` | ref | The monsters flee. |
| `2785h` | ref | ~HAUGHTY ~SLY ~NICE ~MEEK ~ABUSIVE |
| `2903h` | ref | The entire party is killed! |
| `291Fh` | ref | press <enter>/<return> to continue |
| `30BAh` | ref | You've won. Save before quitting?  |

### overlay-03

| offset | 依據 | 內容 |
|---|---|---|
| `000Ch` | ref | tiles |
| `0012h` | ref | Align the espruar and dethek runes |
| `0035h` | ref | shown below, on translation wheel |
| `0057h` | ref | like this: |
| `0062h` | ref | -..-..-.. |
| `006Ch` | ref | - - - - - |
| `0076h` | ref | ......... |
| `0080h` | ref | Type the character in box number  |
| `00A2h` | ref | under the  |
| `00ADh` | ref | path. |
| `00B5h` | ref | type character and press return:  |
| `00D7h` | ref | Sorry, that's incorrect. |
| `00F0h` | ref | An unseen force hurls you into the abyss! |

### overlay-04

| offset | 依據 | 內容 |
|---|---|---|
| `0034h` | ref | cast cure anyway:  |
| `00AAh` | ref |  will only cost  |
| `00BBh` | ref |  gold pieces. |
| `00C9h` | ref | pay for cure  |
| `00D7h` | ref | Not enough money. |
| `00E9h` | ref | is cured. |
| `0263h` | ref | is not blind. |
| `0271h` | ref | Cure Blindness |
| `0306h` | ref | is not Diseased. |
| `0317h` | ref | Cure Disease |
| `03FDh` | ref | Cure Light Wounds |
| `040Fh` | ref | Cure Serious Wounds |
| `0423h` | ref | Cure Critical Wounds |
| `0438h` | ref | Heal |
| `0626h` | ref | is not dead. |
| `0633h` | ref | Raise Dead |
| `086Fh` | ref | is not poisoned. |
| `0880h` | ref | Neutralize Poison |
| `095Fh` | ref | is not cursed. |
| `096Eh` | ref | Remove Curse |
| `0A45h` | ref | is not stoned. |
| `0A54h` | ref | Stone to Flesh |
| `0AF0h` | ref | , how can we help you? |
| `0B08h` | ref | Heal Exit |
| `0D0Ch` | ref | Heal View Take Pool Share Appraise Exit |
| `0D34h` | ref | Heal View Pool Appraise Exit |
| `0D52h` | ref | ~Yes ~No |
| `0D5Bh` | ref | As you leave a priest says, "Excuse me but you have left some money here"  |
| `0DA6h` | ref | Do you want to go back and retrieve your money? |

### overlay-05

| offset | 依據 | 內容 |
|---|---|---|
| `09A4h` | ref | The party has fled. |
| `09B8h` | ref | You have lost the fight. |
| `09D1h` | ref | You have won the duel. |
| `09E8h` | ref | The party has won. |
| `09FBh` | ref | The party has found Treasure! |
| `0A19h` | ref | The duelist receives  |
| `0A2Fh` | ref | Each character receives  |
| `0A48h` | ref | experience points. |
| `0A5Bh` | ref | press <enter>/<return> to continue |
| `0CD9h` | ref | Items:  |
| `0CE1h` | ref | Take |
| `0EBFh` | ref | Take:  |
| `0EC6h` | ref | Money Items Exit |
| `0FCDh` | ref | View Pool Exit |
| `0FDCh` | ref |  Exit |
| `0FE2h` | ref |  Detect Exit |
| `0FEFh` | ref | View Take Pool Share |
| `1004h` | ref | View Take Pool |
| `1013h` | ref | ~Yes ~No |
| `101Ch` | ref | There is still treasure left.   |
| `103Ch` | ref | Do you want to go back and claim your treasure? |
| `1466h` | ref |  takes and hides  |
| `147Ch` | ref |  share. |
| `1489h` | ref | press <enter>/<return> to continue |
| `16E6h` | ref | The monsters rejoice for the party has been destroyed |
| `171Ch` | ref | Press any key to continue |

### overlay-06

| offset | 依據 | 內容 |
|---|---|---|
| `0025h` | ref |                                |
| `0044h` | ref | Items:  |
| `034Ah` | ref | OverLoaded |
| `0462h` | ref | Not enough Money. |
| `062Ch` | ref | Buy View Take Pool Share Appraise Exit |
| `0653h` | ref | Buy View Pool Appraise Exit |
| `066Fh` | ref | ~Yes ~No |
| `0678h` | ref | As you Leave the Shopkeeper says, "Excuse me but you have Left Some Money here."   |
| `06CBh` | ref | Do you want to go back and get your Money? |

### overlay-07

| offset | 依據 | 內容 |
|---|---|---|
| `0361h` | ref | Loading...Please Wait |
| `037Bh` | ref | .dax |
| `0587h` | ref | SPRIT |
| `1BD6h` | ref | CPIC |
| `1BDBh` | ref | ROLF |
| `2207h` | ref |  dies.  |
| `220Fh` | ref |  is hit FOR  |
| `221Ch` | ref |  points of Damage. |
| `222Fh` | ref | press <enter>/<return> to continue |

### overlay-08

| offset | 依據 | 內容 |
|---|---|---|
| `03B0h` | ref | Magic On |
| `03B9h` | ref | Magic Off |
| `03C3h` | ref | That doesn't work |
| `06FFh` | ref | Move  |
| `0705h` | ref | View Aim  |
| `070Fh` | ref | Use  |
| `0714h` | ref | Cast  |
| `071Ah` | ref | Turn  |
| `0720h` | ref | Quick Done |
| `096Fh` | ref | Your Teammate is Dying |
| `0986h` | ref | Continue Battle: |
| `0AD4h` | ref | Move/Attack, Move Left =  |
| `0AF1h` | ref | Flee: |
| `0AF7h` | ref | can't go there |
| `0ED4h` | ref | Not with that weapon |
| `0FA1h` | ref | Guard  |
| `0FA8h` | ref | Delay Quit  |
| `0FB4h` | ref | Bandage  |
| `0FBDh` | ref | Speed Exit |
| `11B8h` | ref | GameSpeed ( |
| `11CAh` | ref | Slower  |
| `11D2h` | ref | Faster  |
| `11DAh` | ref | Exit |

### overlay-09

| offset | 依據 | 內容 |
|---|---|---|
| `003Eh` | ref | flees in panic |
| `09ABh` | ref | Move/Attack, Move Left =  |
| `1260h` | ref | Magic On |
| `1269h` | ref | Magic Off |
| `136Bh` | ref | is forced to flee |
| `137Dh` | ref | Surrenders |

### overlay-10

| offset | 依據 | 內容 |
|---|---|---|
| `1017h` | ref | DungCom |
| `101Fh` | ref | WildCom |
| `1027h` | ref | RandCom |
| `1C2Bh` | ref | A battle begins... |

### overlay-11

| offset | 依據 | 內容 |
|---|---|---|
| `002Fh` | ref | Loading...Please Wait |
| `0045h` | ref | COMSPR |
| `0050h` | ref | ITEMS |

### overlay-12

| offset | 依據 | 內容 |
|---|---|---|
| `00D0h` | ref | is fighting with snakes |
| `0395h` | ref | Suffocates |
| `04D8h` | ref | is silenced |
| `0539h` | ref | dies from poison |
| `05A8h` | ref | Gains an item |
| `075Bh` | ref | lost an image |
| `080Ch` | ref | is coughing |
| `08C3h` | ref | collapses |
| `09E0h` | ref | runs away |
| `09EAh` | ref | is confused |
| `09F6h` | ref | goes berserk |
| `0A03h` | ref | is enraged |
| `0BE4h` | ref | ages |
| `0C46h` | ref | The air clears a little... |
| `0F2Fh` | ref | Avoids it |
| `1063h` | ref | is weakened |
| `143Eh` | ref | is Poisoned |
| `144Ah` | ref | is killed |
| `14ECh` | ref | is Paralyzed |
| `15CFh` | ref | is stupid |
| `15D9h` | ref | lost a spell |
| `171Bh` | ref | is unaffected |
| `17FCh` | ref | goes berzerk |
| `1A45h` | ref | gazes... |
| `1A4Eh` | ref | reflects it! |
| `1A5Bh` | ref | is Stoned |
| `1CDBh` | ref | The air clears a little... |
| `2031h` | ref | Falls dead |
| `264Ch` | ref | Spits Acid |
| `2775h` | ref | is paralyzed |
| `2886h` | ref | gazes... |
| `288Fh` | ref | is paralyzed |
| `2CCDh` | ref | gets zapped |
| `2D5Ch` | ref | is dispelled |
| `2D69h` | ref | resists dispel evil |

### overlay-13

| offset | 依據 | 內容 |
|---|---|---|
| `0256h` | ref | -Backstabs- |
| `0262h` | ref | slays helpless |
| `0271h` | ref | Attacks |
| `0279h` | ref | (from behind)  |
| `0288h` | ref | with one cruel blow |
| `029Ch` | ref | Hitting for  |
| `02A9h` | ref |  point  |
| `02B1h` | ref |  points  |
| `02BAh` | ref | of damage |
| `02C4h` | ref | and Misses |
| `02CFh` | ref | lost a spell |
| `02DCh` | ref | goes down |
| `02E6h` | ref | and is Dying |
| `0313h` | ref | is killed |
| `0D01h` | ref | Got Away |
| `0D0Ah` | ref | Escape is blocked |
| `0F3Fh` | ref | sweeps |
| `11EDh` | ref | turns undead... |
| `11FDh` | ref | is turned |
| `1207h` | ref | Is destroyed |
| `1214h` | ref | Nothing Happens... |
| `220Ah` | ref | Already been targeted |
| `273Bh` | ref | Camp Only Spell |
| `274Bh` | ref | Begins Casting |
| `2F2Eh` | ref | Attack Ally:  |
| `3009h` | ref | Range =  |
| `3015h` | ref | Target  |
| `301Dh` | ref | Next Prev Manual  |
| `302Fh` | ref | Center Exit |
| `303Bh` | ref | Aim: |
| `3379h` | ref | Range =  |
| `3385h` | ref | Center Exit |
| `3391h` | ref | Target  |
| `3399h` | ref | (Use Cursor keys)  |
| `4036h` | ref | engulfs  |
| `4286h` | ref | fires a disintegrate ray |
| `429Fh` | ref | is disintergrated |
| `42B1h` | ref | fires a stone to flesh ray |
| `42CCh` | ref | is Stoned |
| `42D6h` | ref | fires a death ray |
| `42E8h` | ref | is killed |
| `42F2h` | ref | wounds you |
| `480Bh` | ref | hugs  |
| `48C8h` | ref | The Gods intervene! |

### overlay-14

| offset | 依據 | 內容 |
|---|---|---|
| `0901h` | ref | Not Here |
| `0B8Eh` | ref | Bash |
| `0B93h` | ref |  Pick |
| `0B99h` | ref |  Knock |
| `0BA0h` | ref |  Exit |
| `0BA6h` | ref | Locked.  |

### overlay-15

| offset | 依據 | 內容 |
|---|---|---|
| `0327h` | ref | cannot cast spells in this area |
| `0347h` | ref | is in no condition to  |
| `035Eh` | ref | cast any spells |
| `036Eh` | ref | memorize spells |
| `037Eh` | ref | scribe any scrolls |
| `04B3h` | ref | has no spells memorized |
| `0583h` | ref | can memorize: |
| `0591h` | ref |     Cleric Spells: |
| `05A4h` | ref |      Druid Spells: |
| `05B7h` | ref | Magic-User Spells: |
| `08F7h` | ref | Memorize These Spells?  |
| `090Fh` | ref | cannot memorize any spells |
| `092Ah` | ref | Memorize these spells?  |
| `0AF9h` | ref | Scribe These Spells?  |
| `0B0Fh` | ref | has no copyable scrolls |
| `0B27h` | ref | You already know that spell |
| `0B43h` | ref | You are already scibing that spell |
| `0B66h` | ref | You can not scribe that spell. |
| `0B85h` | ref | Scribe these spells?  |
| `0EEEh` | ref | Funky-- |
| `0EF6h` | ref | Dispel Evil |
| `0F02h` | ref | Faerie Fire |
| `0F0Eh` | ref | Fumbling |
| `0F17h` | ref | Helpless |
| `0F20h` | ref | Confused |
| `0F29h` | ref | Cause Disease |
| `0F37h` | ref | Hot Fire Shield |
| `0F47h` | ref | Cold Fire Shield |
| `0F58h` | ref | Poisoned |
| `0F61h` | ref | Regenerating |
| `0F6Eh` | ref | Fire Resistance |
| `0F7Eh` | ref | Minor Globe of Invulnerability |
| `0F9Dh` | ref | enfeebled |
| `0FA7h` | ref | invisible to animals |
| `0FBCh` | ref | Invisible |
| `0FC6h` | ref | Camouflaged |
| `0FD2h` | ref | protected from dragon breath |
| `0FEFh` | ref | berserk |
| `0FF7h` | ref | Displaced |
| `1001h` | ref |  <No Spell Effects> |
| `17F4h` | ref | Party Order:  |
| `1822h` | ref | has been selected |
| `192Eh` | ref | quit TO DOS:  |
| `193Ch` | ref | will be gone |
| `1949h` | ref | Drop from party?  |
| `195Bh` | ref | bids you farewell |
| `196Dh` | ref | is dumped in a ditch |
| `1982h` | ref | Breathes A sigh of relief |
| `1AE6h` | ref | Game Speed =  |
| `1AF4h` | ref |  (0=fastest 9=slowest) |
| `1B0Bh` | ref |  Faster |
| `1B13h` | ref |  Slower |
| `1B1Bh` | ref |  Exit |
| `1B21h` | ref | Game Speed: |
| `1CC9h` | ref | Alter:  |
| `1CD1h` | ref | Pics on   |
| `1CDBh` | ref | Animation on   |
| `1CEAh` | ref | Animation off   |
| `1CFAh` | ref | Pics off   |
| `1D05h` | ref | Exit |
| `23ADh` | ref | The party makes camp... |
| `23E5h` | ref | Camp: |
| `23EBh` | ref | Quit TO DOS  |

### overlay-16

| offset | 依據 | 內容 |
|---|---|---|
| `0077h` | ref | .SAV |
| `007Ch` | ref | from saved game  |
| `043Bh` | ref | *.guy |
| `0441h` | ref | *.cha |
| `0447h` | ref | *.sav |
| `044Dh` | ref | *.hil |
| `05CAh` | ref | CURSE |
| `05D0h` | ref | Put save disk in  |
| `05E4h` | ref | Is your save disk in drive  |
| `0604h` | ref | Unexpected error during save:  |
| `0A72h` | ref | CHEAD |
| `0A78h` | ref | CBODY |
| `0C70h` | ref | .guy |
| `0C75h` | ref | .swg |
| `0D1Bh` | ref | Can't save.  No room on this disk. |
| `0D3Eh` | ref | Lose character?  |
| `0D4Fh` | ref | Ok  Try another disk |
| `0D64h` | ref | .guy |
| `0D6Bh` | ref | .sav |
| `0D90h` | ref | Put save disk in  |
| `0DA4h` | ref | Unexpected error during save:  |
| `0DC3h` | ref | Overwrite  |
| `0DD1h` | ref | New file name:  |
| `0DE1h` | ref | .swg |
| `222Ah` | ref | Put save disk in  |
| `223Eh` | ref | Loading...Please Wait |
| `2254h` | ref | .guy |
| `225Bh` | ref | .cha |
| `2260h` | ref | .swg |
| `2269h` | ref | .spc |
| `3230h` | ref | .dax |
| `3235h` | ref | Unable to load monster |
| `3517h` | ref | CPIC |
| `36D4h` | ref | savgam |
| `36DBh` | ref | .dat |
| `36E2h` | ref | Load Which Game:  |
| `3714h` | ref | Put save disk in  |
| `3726h` | ref | Loading...Please Wait |
| `373Eh` | ref | .sav |
| `3743h` | ref | CPIC |
| `3DF9h` | ref | Save Which Game:  |
| `3E0Bh` | ref | A B C D E F G H I J |
| `3E3Fh` | ref | savgam |
| `3E46h` | ref | .dat |
| `3E4Bh` | ref | Can't save.  No room on this disk. |
| `3E8Eh` | ref | Put save disk in  |
| `3EA2h` | ref | Unexpected error during save:  |
| `3EC1h` | ref | Saving...Please Wait |
| `3ED6h` | ref | CHRDAT |

### overlay-17

| offset | 依據 | 內容 |
|---|---|---|
| `0115h` | ref | Choose a function  |
| `0168h` | ref | You've already defeated Tyranthraxus. |
| `018Eh` | ref | Quit to DOS  |
| `019Bh` | ref | Game not saved.  Quit anyway?  |
| `01BAh` | ref | Free training on |
| `01CBh` | ref | Free training off |
| `06FCh` | ref | Pick Race |
| `070Ah` | ref | Select |
| `0711h` | ref | Pick Gender |
| `071Dh` | ref | Pick Class |
| `0728h` | ref | Pick Alignment |
| `0759h` | ref | Reroll stats?  |
| `0768h` | ref | Character name:  |
| `0779h` | ref | Save  |
| `25A7h` | ref | Drop  |
| `25ADh` | ref |  forever?  |
| `25B8h` | ref | Are you sure?  |
| `25C7h` | ref | You dump  |
| `25D1h` | ref |  out back. |
| `25DCh` | ref |  bids you farewell. |
| `25F0h` | ref |  breathes a sigh of relief. |
| `2841h` | ref |  can't be modified. |
| `2855h` | ref | Modify:  |
| `285Eh` | ref | Keep Exit |
| `35E6h` | ref | Add from where?  |
| `35F7h` | ref | Curse Pool Hillsfar Exit |
| `3610h` | ref | Add a character:  |
| `3622h` | ref | Add  |
| `362Ah` | ref | paladins do not join with evil scum |
| `364Eh` | ref | too many rangers in party |
| `3668h` | ref |  will tolerate no evil! |
| `3E75h` | ref | ready   action |
| `3E88h` | ref | Small |
| `3E8Eh` | ref | Large |
| `3E95h` | ref | Sorry, not in CGA |
| `3EA7h` | ref | Hair |
| `3EACh` | ref | Face |
| `3ED1h` | ref | Is this icon ok?  |
| `4C6Ah` | ref | we only train conscious people |
| `4C89h` | ref | Training costs 1000 gp. |
| `4CA1h` | ref | We don't train that class here |
| `4CC0h` | ref | Not Enough Experience |
| `4CD6h` | ref |  will become: |
| `4CE4h` | ref |     a level  |
| `4CF3h` | ref | and a level  |
| `4D00h` | ref | Do you wish to train?  |
| `4D17h` | ref | Congratulations... |

### overlay-18

| offset | 依據 | 內容 |
|---|---|---|
| `0C6Ah` | ref | Tyranthraxus' spirit coalesces over the slain  |
| `0C99h` | ref | storm giant. 'You have defeated me. Were it not for  |
| `0CCEh` | ref | the Amulet of Lythander, I could possess you and rob  |
| `0D04h` | ref | you of your victory. Still I can escape through the pool. |
| `0D3Eh` | ref | Press any key to continue. |
| `0D59h` | ref | As you reach for the Pool of Radiance, he cries  |
| `0D8Ah` | ref | out, 'Keep the Gauntlet of Moander away from there, you  |
| `0DC3h` | ref | will unleash dangerous energies. Stay back!' As the  |
| `0DF8h` | ref | gauntlet contacts the pool, it contracts and shatters it. |
| `0E32h` | ref | 'I am trapped without escape, you have succeeded  |
| `0E64h` | ref | where armies have not. Gloat while you may, Tyranthraxus  |
| `0E9Eh` | ref | is slain this day.' Before your eyes he crumbles into  |
| `0ED5h` | ref | nothingness. |
| `0EE2h` | ref | You are certain he is destroyed because your  |
| `0F10h` | ref | final bond fades away. The Curse of the Azure Bonds  |
| `0F45h` | ref | has finally been lifted from you! You are free at  |
| `0F78h` | ref | last! |
| `0F7Eh` | ref | The Knights of Myth Drannor rush in, ' |
| `0FA5h` | ref | Congratulations, you have destroyed the Flamed One.  |
| `0FDAh` | ref | With the power of Elminster, let us take you from  |
| `100Dh` | ref | this  foul place, to a fine feast.' |
| `1031h` | ref | You are teleported to Shadowdale, where festivities  |
| `1066h` | ref | have already begun. A huge cheer goes up at your arrival.  |
| `10A1h` | ref | Gharri and Nacacia, arm in arm, yell congratulations  |
| `10D7h` | ref | from the nearby stands. 'You have won!' |

### overlay-19

| offset | 依據 | 內容 |
|---|---|---|
| `0039h` | ref | (NPC) |
| `003Fh` | ref | Age  |
| `0044h` | ref | STR  |
| `0049h` | ref | INT  |
| `004Eh` | ref | WIS  |
| `0053h` | ref | DEX  |
| `0058h` | ref | CON  |
| `005Dh` | ref | CHA  |
| `0062h` | ref | Level |
| `006Ah` | ref | Exp  |
| `006Fh` | ref | Weapon |
| `0076h` | ref | Armor |
| `007Ch` | ref | Status |
| `064Eh` | ref | AC     |
| `0655h` | ref | HP     |
| `065Ch` | ref | THAC0    |
| `066Bh` | ref | Damage         |
| `067Ah` | ref | Encumbrance       |
| `068Ch` | ref | Movement    |
| `0B28h` | ref | Items  |
| `0B2Fh` | ref | Spells  |
| `0B37h` | ref | Trade  |
| `0B3Eh` | ref | Drop  |
| `0B44h` | ref | Heal  |
| `0B4Ah` | ref | Cure  |
| `0B50h` | ref | Exit |
| `0E75h` | ref | Must be unreadied |
| `0E87h` | ref |  was going to scribe from that scroll |
| `0EADh` | ref | is it Okay to lose it?  |
| `0FBCh` | ref | itemptr:       |
| `0FCBh` | ref | namenum(1):    |
| `0FDAh` | ref | namenum(2):    |
| `0FE9h` | ref | namenum(3):    |
| `0FF8h` | ref | plus:          |
| `1007h` | ref | plussave:      |
| `1016h` | ref | ready:         |
| `1025h` | ref | identified:    |
| `1034h` | ref | cursed:        |
| `1043h` | ref | value:         |
| `1052h` | ref | special(1):    |
| `1061h` | ref | special(2):    |
| `1070h` | ref | special(3):    |
| `107Fh` | ref | dice large:     |
| `108Fh` | ref | sides large:    |
| `109Fh` | ref | press a key |
| `1513h` | ref | Ready |
| `1519h` | ref |  Use |
| `151Eh` | ref |  Trade |
| `1525h` | ref |  Drop |
| `152Bh` | ref |  Halve |
| `1532h` | ref |  Join |
| `1538h` | ref |  Sell |
| `1542h` | ref | Items |
| `1548h` | ref | Ready Item |
| `1553h` | ref | Must be Readied |
| `1563h` | ref | Your  |
| `1569h` | ref | will be gone forever |
| `157Eh` | ref | Drop It?  |
| `1EADh` | ref | It's Cursed |
| `1EB9h` | ref | Wrong Class |
| `1EC5h` | ref | already using  |
| `1ED4h` | ref | Your hands are full! |
| `2120h` | ref | Trade with Whom? |
| `2131h` | ref | Overloaded |
| `21E2h` | ref | Can't halve that |
| `2474h` | ref | uses an item |
| `2481h` | ref | Item: |
| `2487h` | ref | oops! |
| `2773h` | ref | I'll give you  |
| `2782h` | ref |  gold pieces for your  |
| `2799h` | ref | Is It a Deal?  |
| `27A8h` | ref | Sold! |
| `27AEh` | ref | Overloaded.  Money will be put in pool. |
| `29B8h` | ref | For 200 gold pieces I'll identify your  |
| `29E0h` | ref | Is It a Deal?  |
| `29EFh` | ref | Not Enough Money |
| `2A00h` | ref | I can't tell anything new about your  |
| `2A26h` | ref | It looks like some sort of  |
| `2C13h` | ref | Trade to? |
| `2C1Fh` | ref | Select type of coin  |
| `2C34h` | ref |  Select |
| `2C3Ch` | ref | How much  |
| `2C46h` | ref | will you trade?  |
| `2F72h` | ref | Select type of coin  |
| `2F87h` | ref |  Select |
| `2F8Fh` | ref | How much  |
| `2F99h` | ref | will you drop?  |
| `3421h` | ref | in Memory |
| `342Bh` | ref | in Grimoire |
| `3437h` | ref | on Scroll |
| `3441h` | ref | on Scrolls |
| `344Ch` | ref | to Choose |
| `3456h` | ref | to Memorize |
| `3462h` | ref | to Scribe |
| `346Ch` | ref | Spells  |
| `36AFh` | ref | Heal whom?  |
| `36BBh` | ref |  feels better |
| `36C9h` | ref |  is unaffected |
| `37B8h` | ref | Cure whom?  |
| `37C4h` | ref | is not diseased |
| `37D4h` | ref | cure anyway:  |
| `37E2h` | ref |  is cured |

### overlay-20

| offset | 依據 | 內容 |
|---|---|---|
| `05C7h` | ref | Rest Time: |
| `06CBh` | ref | Rest Days Hours Mins Add Subtract Exit |
| `0848h` | ref | The Whole Party Is Healed |
| `090Fh` | ref | has memorized |
| `09D3h` | ref | has scribed |
| `0C68h` | ref | Stop Resting?  |
| `0C77h` | ref | Your repose is suddenly interrupted! |

### overlay-21

| offset | 依據 | 內容 |
|---|---|---|
| `01F4h` | ref | Overloaded.  Money will be put in Pool. |
| `0460h` | ref | Overloaded |
| `0A7Fh` | ref | Overloaded |
| `0B58h` | ref | Gems  |
| `0B5Eh` | ref | Gold  |
| `0B64h` | ref | Platinum  |
| `0B6Eh` | ref | electrum  |
| `0B78h` | ref | silver  |
| `0B80h` | ref | Copper  |
| `0B88h` | ref | Jewelry  |
| `0CBAh` | ref | Select type of coin  |
| `0CCFh` | ref | Select |
| `0CD6h` | ref | How much  |
| `0CE0h` | ref | will you take?  |
| `18BAh` | ref | No Gems or Jewelry |
| `18CDh` | ref |  Gem |
| `18D2h` | ref |  Gems |
| `18D8h` | ref |  piece of Jewelry |
| `18EAh` | ref |  pieces of Jewelry |
| `18FDh` | ref | You have a fine collection of: |
| `191Ch` | ref |   Gems |
| `1923h` | ref |   Jewelry |
| `192Dh` | ref |  Exit |
| `1933h` | ref | Appraise :  |
| `193Fh` | ref | The Gem is Valued at  |
| `1955h` | ref |  gp. |
| `195Ah` | ref | Sell |
| `195Fh` | ref | Sell Keep |
| `1969h` | ref | You can :  |
| `1974h` | ref | The Jewel is Valued at  |

### overlay-22

| offset | 依據 | 內容 |
|---|---|---|
| `018Fh` | ref | Cast |
| `0194h` | ref | Memorize |
| `019Dh` | ref | Scribe |
| `01A4h` | ref | Learn |
| `01CAh` | ref | Choose Spell:  |
| `10BFh` | ref | Cast Spell on whom |
| `11ECh` | ref | can't be cast here... |
| `1202h` | ref | Lose it?  |
| `120Ch` | ref | That Item |
| `1216h` | ref | is a combat-only item... |
| `122Fh` | ref | Use it?  |
| `1238h` | ref | miscasts |
| `1241h` | ref | casts |
| `1247h` | ref | Abort Spell?  |
| `1255h` | ref | Spell Aborted |
| `1CEEh` | ref | is Blessed |
| `1D20h` | ref | is Cursed |
| `1DC9h` | ref | is affected |
| `1E02h` | ref | is protected |
| `1E3Ch` | ref | is cold-resistant |
| `1EB0h` | ref | is unaffected |
| `1EBEh` | ref | is charmed |
| `1F86h` | ref | is stronger |
| `1F92h` | ref | is unaffected |
| `20D0h` | ref | has been reduced |
| `2174h` | ref | is friendly |
| `2226h` | ref | is shielded |
| `22A7h` | ref | falls asleep |
| `23EAh` | ref | is held |
| `2449h` | ref | is fire resistant |
| `2488h` | ref | is silenced |
| `24C1h` | ref | is affected |
| `2575h` | ref | is charmed |
| `2675h` | ref | is invisible |
| `26AFh` | ref | Knock-Knock |
| `26E8h` | ref | is duplicated |
| `2748h` | ref | is weakened |
| `2781h` | ref | Creates a noxious cloud |
| `2DAAh` | ref | is animated |
| `2F56h` | ref | can see |
| `2F9Bh` | ref | is blind |
| `3087h` | ref | is diseased |
| `3172h` | ref | is affected |
| `3522h` | ref | is praying |
| `3576h` | ref | is un-cursed |
| `3583h` | ref | has an item un-cursed |
| `3696h` | ref | has been cursed! |
| `36D4h` | ref | is blinking |
| `3913h` | ref | is Hasted |
| `3CE3h` | ref | is Slowed |
| `3D1Bh` | ref | is restored |
| `3ED8h` | ref | is Speedy |
| `3F63h` | ref | is stronger |
| `4036h` | ref | is paralyzed |
| `4070h` | ref | is Healed |
| `40C8h` | ref | is invisible |
| `4178h` | ref | is unpoisoned |
| `4186h` | ref | is unaffected |
| `42FFh` | ref | smashes them flat |
| `4426h` | ref | is affected |
| `44C9h` | ref | is raised |
| `459Eh` | ref | is slain |
| `45A7h` | ref | is unaffected |
| `4665h` | ref | is entangled |
| `471Dh` | ref | is highlighted |
| `474Ch` | ref | is invisible |
| `4786h` | ref | is charmed |
| `4816h` | ref | is confused |
| `48DAh` | ref | teleports |
| `4A72h` | ref | runs in terror |
| `4A81h` | ref | is unaffected |
| `4BB3h` | ref | flame type:  |
| `4BC0h` | ref | Hot Cold |
| `4BC9h` | ref | is protected |
| `4BD7h` | ref | Abort spell?  |
| `4BE5h` | ref | Yes No |
| `4D8Ch` | ref | is clumsy |
| `4D96h` | ref | is slowed |
| `4EEAh` | ref | is protected |
| `4F24h` | ref | Creates a poisonous cloud |
| `5781h` | ref | is Healed |
| `57D9h` | ref | Breathes! |
| `5953h` | ref | Spits Acid |
| `595Eh` | ref | Spits Acid and Misses |
| `5A9Ah` | ref | breathes acid |
| `5C78h` | ref | breathes fire |
| `5E03h` | ref | Breathes Fire |
| `5F1Eh` | ref | throws lightning |
| `600Ch` | ref | gazes... |
| `6015h` | ref | is paralyzed |
| `61F4h` | ref | Casts a Spell |
| `6202h` | ref | Spell: |

### overlay-23

| offset | 依據 | 內容 |
|---|---|---|
| `0E03h` | ref | starts to cough |
| `0E13h` | ref | chokes and gags from nausea |
| `0E2Fh` | ref | is Poisoned |
| `0E3Bh` | ref | is killed |
| `163Ah` | ref | is Cured |
| `1F19h` | ref | takes  |
| `1F20h` | ref |  points of damage  |
| `1F33h` | ref | takes 1 point of damage  |
| `1F4Ch` | ref | from Fire |
| `1F56h` | ref | from Cold |
| `1F60h` | ref | from Electricity |
| `1F71h` | ref | from Acid |
| `1F7Bh` | ref | from Magic |
| `1F86h` | ref | lost a spell |
| `1F93h` | ref | Goes Down |
| `1F9Dh` | ref | , and is Dying |
| `1FCCh` | ref | is killed |
| `22F9h` | ref | is Unaffected |
| `24CCh` | ref | stands up and grins |
| `24E0h` | ref | gets back up |

### overlay-24

| offset | 依據 | 內容 |
|---|---|---|
| `0452h` | ref |  Yes   |
| `0459h` | ref |  No    |
| `0789h` | ref | Name |
| `078Eh` | ref | AC  HP |
| `0A1Ch` | ref | Hitpoints |
| `0A29h` | ref | (Helpless) |
| `15C4h` | ref | Nil Item pointer... |
| `15D8h` | ref | Tried to Lose item & couldn't find it! |
| `251Fh` | ref | is fully healed |
| `252Fh` | ref | is partially healed |
| `286Eh` | ref | Guarding |
| `2B99h` | ref |  camping |
| `2BA2h` | ref |  search |
| `2E3Bh` | ref |  Exit |
| `2E63h` | ref | Select |
| `337Bh` | ref | is bandaged |

### overlay-25

| offset | 依據 | 內容 |
|---|---|---|
| `0EA1h` | ref | Pick New Class |
| `0EB0h` | ref |  doesn't qualify. |
| `0EC3h` | ref | Select |
| `0ECAh` | ref |  is now a 1st level  |

### overlay-26

| offset | 依據 | 內容 |
|---|---|---|
| `0DE7h` | ref |  Next |
| `0DEDh` | ref |  Prev |
| `0DF3h` | ref |  Exit |
| `1103h` | ref | Yes No |

### overlay-29

| offset | 依據 | 內容 |
|---|---|---|
| `00D8h` | ref | Loading...Please Wait |
| `00F2h` | ref | FINAL |
| `00F8h` | ref | .dax |
| `00FDh` | ref | PIC not found |
| `05C4h` | ref | HEAD |
| `05C9h` | ref | head not found |
| `05D8h` | ref | BODY |
| `0735h` | ref | Illegal range in Show3DSprite. |
| `080Ch` | ref | bigpic |

### overlay-30

| offset | 依據 | 內容 |
|---|---|---|
| `0FFFh` | ref | WALLDEF |
| `1007h` | ref | .dax |
| `100Ch` | ref | Unable to load  |
| `101Ch` | ref |  from WALLDEF |
| `1314h` | ref | .dax |
| `1319h` | ref | Unable to load geo in Load3DMap. |

### overlay-33

| offset | 依據 | 內容 |
|---|---|---|
| `000Ch` | ref | Start range error in Load24x24Set |
| `01BCh` | ref | CHEAD |
| `01C2h` | ref | CBODY |
| `01C8h` | ref | COMSPR |
| `01CFh` | ref | ICON |

### overlay-34

| offset | 依據 | 內容 |
|---|---|---|
| `0007h` | ref | EXIT |
| `000Ch` | ref | GOTO |
| `0011h` | ref | GOSUB |
| `0017h` | ref | COMPARE |
| `0023h` | ref | SUBTRAT |
| `002Bh` | ref | DIVIDE |
| `0032h` | ref | MULTIPLY |
| `003Bh` | ref | RANDOM |
| `0042h` | ref | SAVE |
| `0047h` | ref | LOAD CHARACTER |
| `0056h` | ref | LOAD MONSTER |
| `0063h` | ref | SETUP MONSTER |
| `0071h` | ref | APPROACH |
| `007Ah` | ref | PICTURE |
| `0082h` | ref | INPUT NUMBER |
| `008Fh` | ref | INPUT STRING |
| `009Ch` | ref | PRINT |
| `00A2h` | ref | PRINTCLEAR |
| `00ADh` | ref | RETURN |
| `00B4h` | ref | COMPARE AND |
| `00C0h` | ref | VERTICAL MENU |
| `00CEh` | ref | IF =  |
| `00D4h` | ref | IF <> |
| `00DAh` | ref | IF < |
| `00DFh` | ref | IF > |
| `00E4h` | ref | IF <= |
| `00EAh` | ref | IF >= |
| `00F0h` | ref | CLEARMONSTERS |
| `00FEh` | ref | PARTYSTRENGTH |
| `010Ch` | ref | CHECKPARTY |
| `0117h` | ref | NEWECL |
| `011Eh` | ref | LOAD FILES |
| `0129h` | ref | LOAD PIECES |
| `0135h` | ref | PARTY SURPRISE |
| `0144h` | ref | SURPRISE |
| `014Dh` | ref | COMBAT |
| `0154h` | ref | ON GOTO |
| `015Ch` | ref | ON GOSUB |
| `0165h` | ref | TREASURE |
| `0172h` | ref | ENCOUNTER MENU |
| `0181h` | ref | GETTABLE |
| `018Ah` | ref | HORIZONTAL MENU |
| `019Ah` | ref | PARLAY |
| `01A1h` | ref | CALL |
| `01A6h` | ref | DAMAGE |
| `01B4h` | ref | SPRITE OFF |
| `01BFh` | ref | FIND ITEM |
| `01C9h` | ref | PRINT RETURN |
| `01D6h` | ref | ECL CLOCK |
| `01E0h` | ref | SAVE TABLE |
| `01EBh` | ref | ADD NPC |
| `01F3h` | ref | PROGRAM |
| `01FFh` | ref | DELAY |
| `0205h` | ref | SPELL |
| `020Bh` | ref | PROTECTION |
| `0216h` | ref | CLEAR BOX |
| `0220h` | ref | DUMP |
| `0225h` | ref | FIND SPECIAL |
| `0232h` | ref | DESTROY ITEMS |

### overlay-35

| offset | 依據 | 內容 |
|---|---|---|
| `0027h` | ref | 8x8d |
| `002Ch` | ref | Unable to load  |
| `003Ch` | ref |  from 8x8D |
| `0150h` | ref | Bad symbol number in Put8x8Symbol. |

## pc98（31 個模組，843 條）

### overlay-00

| offset | 依據 | 內容 |
|---|---|---|
| `0000h` | ref | ゲームをセーブしてください |

### overlay-01

| offset | 依據 | 內容 |
|---|---|---|
| `0051h` | ref | BASED ON THE TSR NOVEL "AZURE BONDS" |
| `007Ah` | ref | Kate Novak |
| `0089h` | ref | Jeff Grubb |
| `0094h` | ref | SCENARIO CREATED BY: |
| `00A9h` | ref | TSR,INC. |
| `00B6h` | ref | Jeff Grubb,George Mac Donald |
| `00D3h` | ref | GAME CREATED BY: |
| `00E4h` | ref | SSI SPECIAL PROJECTS |
| `00F9h` | ref | PROJECT LEADER: |
| `0109h` | ref | George Mac Donald |
| `011Bh` | ref | PROGRAMMING: |
| `0128h` | ref | Scot Bayless |
| `0135h` | ref | Russ Brown |
| `0140h` | ref | Michael Mancuso |
| `0150h` | ref | DEVELOPMENT: |
| `015Dh` | ref | David Shelley |
| `016Bh` | ref | Oran Kangas |
| `0177h` | ref | GRAPHIC ARTS: |
| `0185h` | ref | Tom Wahl |
| `018Eh` | ref | Fred Butts |
| `0199h` | ref | Susan Manley |
| `01A6h` | ref | Mark Johnson |
| `01B3h` | ref | Cyrus Lum |
| `01BDh` | ref | PLAYTESTING: |
| `01CAh` | ref | Jim Jennings |
| `01D7h` | ref | James Kucera |
| `01E4h` | ref | Rick White |
| `01EFh` | ref | Robert Daly |
| `01FBh` | ref | GAME CONVERTED BY PONYCANYON |
| `0218h` | ref |  ＳＰＥＣＩＡＬ　ＰＲＯＪＥＣＴＳ　ＶＥＲＳＩＯＮ１．２ |
| `0250h` | ref |  PONYCANYON INC. |
| `0261h` | ref |  Kunihiko Kagawa |
| `0272h` | ref |  Yoshiaki Matsumoto |
| `0286h` | ref |  Group SNE |
| `0291h` | ref |  Hitoshi Yasuda |
| `02A1h` | ref |  Miyuki Kiyomatsu |
| `02B3h` | ref |  S.R.S. |
| `02BBh` | ref |  Seishi Yokota |
| `02CAh` | ref |  MUSIC COMPOSED BY |
| `02DDh` | ref |  Takeshi Yasuda |
| `02EDh` | ref | Marionette Inc. |
| `02FDh` | ref | Yoshiaki Sakaguchi |
| `0310h` | ref | Masato Kobayashi |
| `0930h` | ref | title |

### overlay-02

| offset | 依據 | 內容 |
|---|---|---|
| `00A2h` | ref | ┴t.ﾄ> |
| `0488h` | ref | 読み込み中です |
| `0497h` | ref | CPIC |
| `10A3h` | ref | リターン・キーを押してください |
| `1BBDh` | ref | ITEM |
| `1BC2h` | ref | .dax |
| `1BC7h` | ref | アイテム・ファイルが見つかりません |
| `2208h` | ref | お互いに様子を見ている |
| `221Fh` | ref | 相手は逃げた |
| `2AB7h` | ref | パーティーは全滅した！ |
| `335Fh` | ref | データの整理をします リターン・キーを押してください |
| `39EDh` | ref | データの整理をします リターン・キーを押してください |

### overlay-03

| offset | 依據 | 內容 |
|---|---|---|
| `001Eh` | ref | tiles |
| `0024h` | ref | 次のエスプルーアー文字とデテーク文字とを合わせてください。 |
| `005Fh` | ref | －・・－・・－・・ |
| `0072h` | ref | －　－　－　－　－ |
| `0085h` | ref | ‥‥‥‥‥‥‥‥‥ |
| `0098h` | ref | で示された経路の、 |
| `00ABh` | ref | 番のマスには何とありますか？ |
| `00CAh` | ref | １文字入力してリターン・キーを押してください |
| `00F7h` | ref | 間違ってます |
| `0104h` | ref | 目に見えぬ力が、きみを〈奈落〉に放りこんだ！ |

### overlay-04

| offset | 依據 | 內容 |
|---|---|---|
| `0034h` | ref | それでもかけてもらいますか？ |
| `00B4h` | ref | は、 |
| `00B9h` | ref | ｇｐでかけてあげましょう。 |
| `00D4h` | ref | かけてもらいますか？ |
| `00E9h` | ref | お金が足りません |
| `00FAh` | ref | は治った |
| `0273h` | ref | は盲目ではありませんよ。 |
| `028Ch` | ref | キュア・ブラインドネス |
| `0329h` | ref | は病いに冒されてはいませんよ。 |
| `0348h` | ref | キュア・ディジーズ |
| `0434h` | ref | キュア・ライト・ウーンズ |
| `044Dh` | ref | キュア・シリアス・ウーンズ |
| `0468h` | ref | キュア・クリティカル・ウーンズ |
| `0487h` | ref | ヒール |
| `0671h` | ref | は死んではいませんよ |
| `0686h` | ref | は生き返りませんよ |
| `0699h` | ref | レイズ・デッド |
| `090Ah` | ref | は毒を受けてはいませんよ |
| `0923h` | ref | ニュートラライズ・ポイズン |
| `0A0Bh` | ref | は呪われてはいませんよ |
| `0A22h` | ref | リムーブ・カース |
| `0AFDh` | ref | は石になってはいませんよ |
| `0B16h` | ref | ストーン・トゥ・フレッシュ |
| `0BBEh` | ref | さん。どのような治療をお望みですか？ |
| `0BE3h` | ref |  治療  |
| `0BEAh` | ref | やめる |
| `0BF2h` | ref | 治療　やめる |
| `0E30h` | ref |      治療  ｷｬﾗｸﾀｰ 金戻す 金集ﾒﾙ 金分ｹﾙ    見積る                      やめる     |
| `0E81h` | ref |      治療  ｷｬﾗｸﾀｰ        金集ﾒﾙ           見積る                      やめる     |
| `0ED3h` | ref |  はい 　いいえ |
| `0EE2h` | ref | きみたちが行こうとすると、僧侶が呼び止めた。「お金をお忘れですよ」 |
| `0F25h` | ref | 引き返してお金を取りますか？ |

### overlay-05

| offset | 依據 | 內容 |
|---|---|---|
| `0990h` | ref | パーティーは逃げた。 |
| `09A5h` | ref | きみたちは戦いに敗れた。 |
| `09BEh` | ref | きみは決闘に勝った。 |
| `09D3h` | ref | パーティーは勝った。 |
| `09E8h` | ref | パーティーは宝物を見つけた！ |
| `0A05h` | ref | 決闘者は |
| `0A0Eh` | ref | 各キャラクターは |
| `0A1Fh` | ref | 点の経験点を得た。 |
| `0A32h` | ref | リターン・キーを押してください           |
| `0CD4h` | ref |  取る  |
| `0CDBh` | ref | アイテム |
| `0CE4h` | ref | 取る |
| `0EDFh` | ref | 取る： |
| `0EE6h` | ref |    お金   ｱｲﾃﾑ                やめる     |
| `1018h` | ref | ｷｬﾗｸﾀｰ |
| `101Fh` | ref | 金集ﾒﾙ |
| `1026h` | ref | 抜ける |
| `102Dh` | ref | ﾃﾞｨﾃｸﾄ |
| `1034h` | ref |  取る  |
| `103Bh` | ref | 金分ｹﾙ |
| `1042h` | ref | まだ宝物が残っている。 |
| `1059h` | ref | 戻って宝物を取りますか？ |
| `14BDh` | ref | は自分の分け前を取り、隠した。 |
| `14DCh` | ref | リターン・キーを押してください           |
| `172Dh` | ref | モンスターはパーティーを全滅させ、喜んでいる。 |
| `175Ch` | ref | 何かキーを押してください |

### overlay-06

| offset | 依據 | 內容 |
|---|---|---|
| `0025h` | ref |  買う  |
| `002Ch` | ref | アイテム： |
| `0393h` | ref | 持ちすぎです |
| `04ADh` | ref | お金が足りません |
| `0676h` | ref |      買う  ｷｬﾗｸﾀｰ 金戻す 金集ﾒﾙ 金分ｹﾙ    見積る                      抜ける     |
| `06C7h` | ref |      買う  ｷｬﾗｸﾀｰ        金集ﾒﾙ           見積る                      抜ける     |
| `0718h` | ref | きみたちが行こうとすると、店主が呼び止めた。「お金をお忘れですよ」 |
| `075Bh` | ref | 引き返してお金を取りますか？ |

### overlay-07

| offset | 依據 | 內容 |
|---|---|---|
| `0034h` | ref | ALIAS |
| `003Ah` | ref | エイリアス |
| `0045h` | ref | DRAGONBAIT |
| `0050h` | ref | ドラゴンベイト |
| `005Fh` | ref | AKABAR BEL AKAS |
| `006Fh` | ref | アーカバー・ベル・アーカッシュ |
| `0481h` | ref | 読みこみ中です |
| `0494h` | ref | .dax |
| `0681h` | ref | SPRIT |
| `1B7Dh` | ref | CPIC |
| `1B82h` | ref | ロルフ |
| `2192h` | ref | ALIAS |
| `2198h` | ref | エイリアス |
| `21A3h` | ref | DRAGONBAIT |
| `21AEh` | ref | ドラゴンベイト |
| `21BDh` | ref | AKABAR BEL AKAS |
| `21CDh` | ref | アーカバー・ベル・アーカッシュ |
| `21F2h` | ref | 死んだ。 |
| `21FBh` | ref | ポイントのダメージを受けた |
| `2216h` | ref | リターン・キーを押してください |

### overlay-08

| offset | 依據 | 內容 |
|---|---|---|
| `0400h` | ref | 魔法使用 |
| `0409h` | ref | 魔法不使用 |
| `071Ch` | ref |  移動  |
| `0723h` | ref | ｷｬﾗｸﾀｰ |
| `072Ah` | ref |  照準  |
| `0731h` | ref |  使う  |
| `0738h` | ref |  魔ｶｹﾙ |
| `073Fh` | ref |  退散  |
| `0746h` | ref |  自動  |
| `074Dh` | ref | その他 |
| `0979h` | ref | 瀕死のキャラクターがいる |
| `0992h` | ref | 戦闘を続けますか？ |
| `0AE5h` | ref | 移動／攻撃，残り移動力＝ |
| `0B02h` | ref | 逃げますか？ |
| `0B0Fh` | ref | そこへは行けない |
| `0F06h` | ref | そこへは進めない |
| `0FCFh` | ref |  防御  |
| `0FD6h` | ref |  待機  |
| `0FDDh` | ref |  完了  |
| `0FE4h` | ref | 手当て |
| `0FEBh` | ref |  速度  |
| `0FF2h` | ref | 抜ける |
| `1019h` | ref | 効果がなかった |
| `122Ch` | ref | 速さ（ |
| `1236h` | ref | 抜ける |
| `123Fh` | ref |  遅く  |
| `1246h` | ref |  速く  |

### overlay-09

| offset | 依據 | 內容 |
|---|---|---|
| `003Eh` | ref | は恐慌にとらわれ、逃げ出した。 |
| `09CDh` | ref | 移動／攻撃，残り移動力＝ |
| `1282h` | ref | 魔法使用 |
| `128Bh` | ref | 魔法不使用 |
| `1385h` | ref | は退散させられた。 |
| `1398h` | ref | は降伏した。 |

### overlay-10

| offset | 依據 | 內容 |
|---|---|---|
| `100Bh` | ref | DungCom |
| `1013h` | ref | WildCom |
| `101Bh` | ref | RandCom |
| `1C19h` | ref | 戦いが始まった |

### overlay-11

| offset | 依據 | 內容 |
|---|---|---|
| `002Fh` | ref | 読みこみ中です |
| `0042h` | ref | ITEMS |
| `0501h` | ref | COMSPR |

### overlay-12

| offset | 依據 | 內容 |
|---|---|---|
| `00D0h` | ref | は蛇と戦っている。 |
| `038Ah` | ref | は窒息した。 |
| `04CFh` | ref | は沈黙させられた。 |
| `0537h` | ref | は毒のために死んだ。 |
| `05AAh` | ref | はアイテムを得た。 |
| `0762h` | ref | は幻影を失った。 |
| `0816h` | ref | は咳きこんでいる。 |
| `08D4h` | ref | は倒れた。 |
| `09F2h` | ref | は逃げ去った。 |
| `0A01h` | ref | は混乱した。 |
| `0A0Eh` | ref | は狂暴化した。 |
| `0A1Dh` | ref | は怒りに我を忘れた。 |
| `0BFDh` | ref | は歳をとった。 |
| `0C69h` | ref | 空気は多少きれいになった |
| `0F50h` | ref | は避けた。 |
| `1084h` | ref | は弱まった。 |
| `1460h` | ref | は毒を受けた。 |
| `146Fh` | ref | は死んだ。 |
| `1512h` | ref | は麻痺した。 |
| `15F5h` | ref | は愚かになった。 |
| `1606h` | ref | は呪文を唱えきれなかった。 |
| `1756h` | ref | は影響を受けなかった。 |
| `1840h` | ref | は狂暴化した。 |
| `1A8Bh` | ref | は睨んだ。 |
| `1A96h` | ref | は視線をはねかえした！ |
| `1AADh` | ref | は石になった。 |
| `1D45h` | ref | 空気は多少きれいになった |
| `2099h` | ref | は倒れ、死んだ。 |
| `26AEh` | ref | は酸を吹いた。 |
| `27E3h` | ref | は麻痺した。 |
| `28F4h` | ref | は睨んだ。 |
| `28FFh` | ref | は麻痺した。 |
| `2D45h` | ref | は報復を受けた。 |
| `2DD9h` | ref | は異界に送り返された。 |
| `2DF0h` | ref | は「ディスペル・イービル」に耐えた。 |

### overlay-13

| offset | 依據 | 內容 |
|---|---|---|
| `0256h` | ref | はバックスタッブを行ない、　 |
| `0273h` | ref | は無力な |
| `027Ch` | ref | は攻撃し、 |
| `028Ah` | ref | の背後 |
| `0291h` | ref | 無慈悲な一撃を加えた。 |
| `02A8h` | ref | 命中して |
| `02B1h` | ref | ポイントのダメージを与えた。 |
| `02CEh` | ref | はダメージを受けなかった。 |
| `02E9h` | ref | にかわされた。 |
| `02F8h` | ref | は呪文を無駄にした。 |
| `030Dh` | ref | は倒れた。 |
| `0318h` | ref | 瀕死の状態だ。 |
| `0347h` | ref | そして、死んだ。 |
| `0D61h` | ref | は逃げきった。 |
| `0D70h` | ref | 逃げられなかった |
| `0FA4h` | ref | なぎ払った。 |
| `1258h` | ref | アンデッドの退散を試みた。 |
| `1273h` | ref | は退散させられた。 |
| `1286h` | ref | は破壊された。 |
| `1295h` | ref | 何も起こらなかった |
| `2244h` | ref | すでに目標として選んである |
| `2773h` | ref | キャンプでのみ使う呪文だ |
| `278Ch` | ref | は呪文を唱え始めた。 |
| `2DA0h` | ref | 味方を攻撃するのですか？ |
| `2E86h` | ref | 距離＝ |
| `2E90h` | ref |  決定  |
| `2E97h` | ref |  次順  |
| `2E9Eh` | ref |  前順  |
| `2EA5h` | ref |  手動  |
| `2EACh` | ref |  中央  |
| `2EB3h` | ref | やめる |
| `3261h` | ref | 距離＝ |
| `326Bh` | ref |  中央  |
| `3272h` | ref | やめる |
| `3279h` | ref |  決定  |
| `3280h` | ref | （テン・キーを使ってください） |
| `3F62h` | ref | を包みこんだ。 |
| `41E4h` | ref | は原子分解光線を発射した。 |
| `41FFh` | ref | は粉々になった。 |
| `4210h` | ref | は石化光線を発射した。 |
| `4227h` | ref | は石になった。 |
| `4236h` | ref | は殺人光線を発射した。 |
| `424Dh` | ref | は死んだ。 |
| `4258h` | ref | はきみを傷つけた。 |
| `477Ch` | ref | を絞めつけた。 |
| `4867h` | ref | 神々の介入だ！ |

### overlay-14

| offset | 依據 | 內容 |
|---|---|---|
| `08D3h` | ref | ここではできない |
| `0B9Fh` | ref | 蹴破る |
| `0BA6h` | ref | 鍵開け |
| `0BADh` | ref | ノック |
| `0BB4h` | ref | やめる |
| `0BBBh` | ref | 鍵がかかっている |

### overlay-15

| offset | 依據 | 內容 |
|---|---|---|
| `0327h` | ref | ここでは、魔法が使えない |
| `0340h` | ref | 状態ではない |
| `034Dh` | ref | は呪文を唱えられる |
| `0360h` | ref | は魔法を記憶できる |
| `0373h` | ref | は魔法を書き写せる |
| `04A3h` | ref | は呪文をまったく覚えていない。 |
| `057Ah` | ref | は次の数だけ呪文を記憶できる。 |
| `0599h` | ref | 　　　僧侶魔法： |
| `05AAh` | ref | ドルイド僧魔法： |
| `05BBh` | ref | 　魔法使い魔法： |
| `08F9h` | ref | これらの呪文を覚えますか？ |
| `0914h` | ref | は呪文の記憶はできない。 |
| `0AE9h` | ref | これらの呪文を書き写しますか？ |
| `0B08h` | ref | は書き写せる巻物を持っていない。 |
| `0B29h` | ref | その魔法はすでに知っている。 |
| `0B46h` | ref | その魔法はすでに書き写そうとしている。 |
| `0B6Dh` | ref | その魔法は書き写せない。 |
| `0ED9h` | ref | 怯えている |
| `0EE4h` | ref | 悪の退散 |
| `0EEDh` | ref | フェアリー・ファイア |
| `0F02h` | ref | 遅鈍 |
| `0F07h` | ref | 無力 |
| `0F0Ch` | ref | 混乱 |
| `0F11h` | ref | 病い |
| `0F16h` | ref | 熱いファイア・シールド |
| `0F2Dh` | ref | 冷たいファイア・シールド |
| `0F49h` | ref | 再生中 |
| `0F50h` | ref | 炎からの防護 |
| `0F5Dh` | ref | 低呪文封じ |
| `0F68h` | ref | 暗愚 |
| `0F6Dh` | ref | 動物に対する透明 |
| `0F7Eh` | ref | 透明 |
| `0F83h` | ref | 偽装 |
| `0F88h` | ref | 竜の息吹よりの防護 |
| `0F9Bh` | ref | 狂暴化 |
| `0FA2h` | ref | 位置のごまかし |
| `0FB1h` | ref | 〈魔法の影響はない〉 |
| `17C2h` | ref | 並び順： |
| `17EBh` | ref | を選んだ。 |
| `191Eh` | ref | ゲームを終わりますか？ |
| `1935h` | ref | は完全にいなくなる。 |
| `194Ah` | ref | かまわず離脱させますか？ |
| `1963h` | ref | は別れを告げた。 |
| `1974h` | ref | はドブに捨てられた。 |
| `1989h` | ref | は安堵のため息をもらした。 |
| `1B38h` | ref | 速さ＝ |
| `1B3Fh` | ref | 　（０＝速い　９＝遅い） |
| `1B58h` | ref |  速く  |
| `1B5Fh` | ref |  遅く  |
| `1B66h` | ref | 抜ける |
| `1B6Dh` | ref | 速度： |
| `1D0Ah` | ref |  順序  |
| `1D11h` | ref |  離脱  |
| `1D18h` | ref |  速度  |
| `1D1Fh` | ref |  ｱｲｺﾝ  |
| `1D26h` | ref |  ｱﾆﾒ無 |
| `1D2Dh` | ref |  ｱﾆﾒ有 |
| `1D34h` | ref | 抜ける |
| `1D3Bh` | ref | は変更できない。 |
| `2414h` | ref | パーティーはキャンプを張った。 |

### overlay-16

| offset | 依據 | 內容 |
|---|---|---|
| `014Fh` | ref | .SAV |
| `05FCh` | ref | *.guy |
| `0602h` | ref | *.cha |
| `0608h` | ref | *.sav |
| `060Eh` | ref | *.hil |
| `07D6h` | ref | *.hil |
| `0839h` | ref |    はい                       いいえ     |
| `0862h` | ref | セーブディスクを作成しますか？  |
| `0882h` | ref | セーブディスクを作成中です。 |
| `089Fh` | ref | セーブディスクの作成に失敗しました。リターン・キーを押してください。 |
| `0AA9h` | ref | ディスクに異常があります。 |
| `0AC4h` | ref | ドライブ２にセーブ・ディスクを入れてください |
| `0AF1h` | ref | セーブ中にエラーが発生しました |
| `0B30h` | ref | ロード中にエラーが発生しました |
| `1055h` | ref | CHEAD |
| `105Bh` | ref | CBODY |
| `1253h` | ref | .guy |
| `1258h` | ref | .swg |
| `1382h` | ref | ディスクを交換してください。 |
| `139Fh` | ref | ディスクに余裕がありませんので、セーブできません |
| `13D0h` | ref | キャラクターをあきらめますか？ |
| `13EFh` | ref |    ｱｷﾗﾒﾙ                      別ﾃｨｽｸ     |
| `1418h` | ref | .guy |
| `141Fh` | ref | .sav |
| `1424h` | ref | 同名のキャラクターがいます。 |
| `1441h` | ref | 上書きしますか？ |
| `1452h` | ref | 新しい名前は？ |
| `1461h` | ref | ライトプロテクトをはずしてください。 |
| `1486h` | ref | ディスクに異常があります。 |
| `14A1h` | ref | セーブ中にエラーが発生しました |
| `14C0h` | ref | .swg |
| `29EDh` | ref | ドライブ２にセーブ・ディスクを入れてください |
| `2A1Ah` | ref | 読みこみ中です |
| `2A29h` | ref | .GUY |
| `2A2Eh` | ref | .CHA |
| `2A33h` | ref | .guy |
| `2A3Ah` | ref | .cha |
| `2A3Fh` | ref | .swg |
| `2A48h` | ref | .spc |
| `3A6Bh` | ref | MONCHA |
| `3A72h` | ref | .dax |
| `3A77h` | ref | モンスターをロードできませんでした |
| `3A9Ah` | ref | MONSPC |
| `3AA1h` | ref | MONITM |
| `3DFAh` | ref | CPIC |
| `4159h` | ref | SaveList.EST |
| `4166h` | ref | ドライブ２にセーブ・ディスクを入れてください |
| `427Dh` | ref | savgam |
| `4284h` | ref | .dat |
| `4290h` | ref | どのデータを読みこみますか？ |
| `42CEh` | ref | パーティーは |
| `42DBh` | ref | ます。 |
| `42E2h` | ref | にいます。 |
| `42EDh` | ref | よろしいですか？ |
| `4602h` | ref |       ①     ②     ③     ④     ⑤        ⑥     ⑦     ⑧     ⑨     ⑩       |
| `4653h` | ref | どこにセーブしますか？ |
| `468Dh` | ref | 番は既に登録されています。 |
| `46A8h` | ref | よろしいですか？ |
| `46B9h` | ref | ゲームを終わりますか |
| `4972h` | ref | savgam |
| `4979h` | ref | .dat |
| `497Eh` | ref | ドライブ２にセーブ・ディスクを入れてください |
| `49ABh` | ref | 読みこみ中です |
| `49BCh` | ref | .sav |
| `49C1h` | ref | CPIC |
| `4F42h` | ref | savgam |
| `4F49h` | ref | .dat |
| `4F4Eh` | ref | ディスクに余裕がありませんので、セーブできません |
| `4F9Fh` | ref | ライトプロテクトをはずしてください。 |
| `4FC4h` | ref | セーブ中にエラーが発生しました－ |
| `4FE5h` | ref | SaveList.EST |
| `4FF2h` | ref | 書きこみ中です |
| `5001h` | ref | CHRDAT |

### overlay-17

| offset | 依據 | 內容 |
|---|---|---|
| `01A1h` | ref | 誰を消去しますか？ |
| `01B4h` | ref | 誰を調整しますか？ |
| `01C7h` | ref | 誰を訓練しますか？ |
| `01DAh` | ref | 誰を変更しますか？ |
| `01EDh` | ref | このキャラクターはクラス変更できません。 |
| `0216h` | ref | 誰の詳細を見ますか？ |
| `022Bh` | ref | 誰を離脱させますか？ |
| `0240h` | ref | きみたちは勝利した。 |
| `0255h` | ref | ゲームを終わる前にセーブしておきますか？ |
| `027Eh` | ref | ゲームを終わりますか？ |
| `0295h` | ref | データがセーブされていません。 |
| `02B4h` | ref | それでも終わりますか？ |
| `09F5h` | ref | 種族を選んでください |
| `0A0Dh` | ref |  決定  |
| `0A14h` | ref | 決定 |
| `0A19h` | ref | 性別を選んでください |
| `0A2Eh` | ref | クラスを選んでください |
| `0A45h` | ref | Select |
| `0A4Ch` | ref | アラインメントを選んでください |
| `0A8Dh` | ref | 能力値を決めなおしますか？ |
| `0AA8h` | ref | キャラクターの名前は？ |
| `0ABFh` | ref | このキャラクタ名は受け付けられません。 |
| `0AE6h` | ref | をセーブしますか？ |
| `2813h` | ref | ALIAS |
| `2819h` | ref | エイリアス |
| `2824h` | ref | DRAGONBAIT |
| `282Fh` | ref | ドラゴンベイト |
| `283Eh` | ref | AKABAR BEL AKAS |
| `284Eh` | ref | アーカバー・ベル・アーカッシュ |
| `286Dh` | ref | は永遠にいなくなりますよ？ |
| `288Ah` | ref | かまわないのですね？ |
| `289Fh` | ref | きみたちは |
| `28AAh` | ref | を捨てた |
| `28B3h` | ref | はさようならを言った |
| `28C8h` | ref | は安堵のため息をもらした |
| `2ADEh` | ref |    はい                       いいえ     |
| `2BAAh` | ref | 名前を変更しますか？ |
| `2BBFh` | ref |                 |
| `2BCFh` | ref | 新しい名前は？ |
| `2BDEh` | ref | このキャラクタ名は受け付けられません。 |
| `2C05h` | ref | これでよろしいですか？ |
| `2E52h` | ref | の数値は変更できない |
| `2E67h` | ref | 数値調整： |
| `2E72h` | ref |    決定　                     やめる     |
| `3AF3h` | ref | どこから加えますか？ |
| `3B08h` | ref |   カース プール ﾋﾙｽﾞﾌｧ        やめる     |
| `3B31h` | ref | 加える |
| `3B38h` | ref | キャラクター参加： |
| `3B4Bh` | ref | 加える　やめる |
| `3B5Dh` | ref | 聖騎士は邪悪なクズとは同行しない |
| `3B7Eh` | ref | レンジャーの数が多すぎます |
| `3B99h` | ref | は邪悪を許さない |
| `3BAAh` | ref |                 |
| `452Dh` | ref | 古いアイコン |
| `453Ah` | ref |  準備状態　　　　　行動状態 |
| `4556h` | ref |  新しいアイコン |
| `4566h` | ref |   小   |
| `456Dh` | ref |   大   |
| `45B5h` | ref | このアイコンでいいですか？ |
| `5332h` | ref | 意識のあるかたでないと訓練はできません |
| `5359h` | ref | 訓練料は１０００ｇｐです |
| `5372h` | ref | ここではそのクラスの訓練はできません |
| `5397h` | ref | あなたは十分に成長しています。 |
| `53B6h` | ref | もうこれ以上訓練はできません。 |
| `53D5h` | ref | 経験が不足しています |
| `53EDh` | ref | レベルの |
| `53F8h` | ref | になれます |
| `5406h` | ref | 訓練しますか？ |
| `5415h` | ref | おめでとう |

### overlay-18

| offset | 依據 | 內容 |
|---|---|---|
| `0B8Ch` | ref | .DAX |
| `0B91h` | ref | CED9.DAX |
| `0D13h` | ref | ティランスラクサスの霊体がストーム・ジャイアントの死体に重なった。 |
| `0D56h` | ref | 「よくぞ余を倒した。『ラサンダーの魔除け』さえなければ、汝らの身体を乗っ取り、 |
| `0DA5h` | ref | 勝利をもぎとることもできたのだが。 |
| `0DC8h` | ref | しかし、まだ〈プール〉を通じて逃げることはできる」 |
| `0DFBh` | ref | 何かキーを押してください |
| `0E14h` | ref | きみたちが〈プール・オブ・レイディアンス〉に近づくと、ティランスラクサスは叫んだ。 |
| `0E67h` | ref | 「『モーンダーの篭手』を近づけるな！　危険なエネルギーを解放しようとしているのだぞ！ |
| `0EBCh` | ref | 　篭手が〈プール〉に触れた瞬間、〈プール〉は収縮し、篭手は砕けた。 |
| `0F00h` | ref | 「余は、逃れることができぬか。 |
| `0F1Fh` | ref | 汝らは軍隊でも成しえなかったことを成し遂げたわけだ。 |
| `0F54h` | ref | 喜ぶがよい。笑みを浮かべるがよい。 |
| `0F77h` | ref | この日、ティランスラクサスは滅した」　きみたちの前で、ティランスラクサスは縮み、姿を消した。 |
| `0FD4h` | ref | ティランスラクサスが倒れたことは間違いがない。 |
| `1003h` | ref | というのも、きみたちの最後の〈紺青の呪縛〉がついに消え去ったからだ。 |
| `1048h` | ref | きみたちはついに、解放された！ |
| `1067h` | ref | ミス・ドラノールの騎士たちが部屋になだれこんできた。 |
| `109Ch` | ref | 「おめでとう。きみたちはついに『炎のもの』を打ち倒したな |
| `10D5h` | ref | エルミンスターの力をもって、この忌まわしい場所から脱出させてあげよう。 |
| `111Ch` | ref | そして、宴だ」 |
| `112Bh` | ref | きみたちはシャドウデイルへとテレポートした。 |
| `1158h` | ref | そこでは、すでに祭宴が始まっていた。きみたちが着くと、大きな喚声が巻き起こった。 |
| `11A9h` | ref | 近くのさじきから、ジャーリーとナカシアが腕を組み、きみたちにおめでとうを叫んだ。 |
| `11FAh` | ref | きみたちは勝利したのだ！ |

### overlay-19

| offset | 依據 | 內容 |
|---|---|---|
| `0039h` | ref | （ＮＰＣ） |
| `0044h` | ref | 年齢  |
| `004Ah` | ref | 体力度 |
| `0051h` | ref | 知識度 |
| `0058h` | ref | 賢明度 |
| `005Fh` | ref | 敏捷度 |
| `0066h` | ref | 耐久度 |
| `006Dh` | ref | 魅力度 |
| `0074h` | ref | 経験レベル |
| `0081h` | ref | 経験点  |
| `0089h` | ref | 武器 |
| `008Eh` | ref |  鎧  |
| `0093h` | ref | 状態 |
| `066Eh` | ref | ＡＣ |
| `0673h` | ref | ＨＰ |
| `0678h` | ref | ＴＨＡＣ０ |
| `0689h` | ref | ダメージ |
| `0692h` | ref | 荷重 |
| `0697h` | ref | 移動力 |
| `09B1h` | ref | ００ |
| `0B31h` | ref |  ｱｲﾃﾑ  |
| `0B38h` | ref |  魔法  |
| `0B3Fh` | ref |  渡す  |
| `0B46h` | ref | 捨てる |
| `0B4Dh` | ref |  回復  |
| `0B54h` | ref |  治療  |
| `0B5Bh` | ref | 抜ける |
| `0E23h` | ref | 装備をはずしてください |
| `0E3Ah` | ref | はその巻物から呪文を書き写そうとしています。 |
| `0E67h` | ref | この巻物からの書き写しを中止しますか？ |
| `0F85h` | ref | itemptr:       |
| `0F94h` | ref | namenum(1):    |
| `0FA3h` | ref | namenum(2):    |
| `0FB2h` | ref | namenum(3):    |
| `0FC1h` | ref | plus:          |
| `0FD0h` | ref | plussave:      |
| `0FDFh` | ref | ready:         |
| `0FEEh` | ref | identified:    |
| `0FFDh` | ref | cursed:        |
| `100Ch` | ref | value:         |
| `101Bh` | ref | special(1):    |
| `102Ah` | ref | special(2):    |
| `1039h` | ref | special(3):    |
| `1048h` | ref | dice large:     |
| `1058h` | ref | sides large:    |
| `1068h` | ref | press a key |
| `14DCh` | ref |  準備  |
| `14E3h` | ref |  使う  |
| `14EAh` | ref |  渡す  |
| `14F1h` | ref | 捨てる |
| `14F8h` | ref | 分ける |
| `14FFh` | ref |  ﾏﾄﾒﾙ  |
| `1506h` | ref |  売る  |
| `150Dh` | ref |  鑑定  |
| `1514h` | ref | アイテム |
| `151Dh` | ref | 準備する　アイテム |
| `1530h` | ref | 装備をしてください |
| `1543h` | ref | あなたの |
| `154Ch` | ref | は永遠になくなります。 |
| `1563h` | ref | それでも捨てますか？ |
| `1E65h` | ref | 呪われています |
| `1E74h` | ref | クラスが一致しません |
| `1E89h` | ref | すでに |
| `1E90h` | ref | を使用中です |
| `1E9Dh` | ref | 手がいっぱいです |
| `20E4h` | ref | 誰に渡しますか？ |
| `20F5h` | ref | 持ちすぎです |
| `21A8h` | ref | それは分けることはできません |
| `2446h` | ref | はアイテムを使った。 |
| `245Bh` | ref | アイテム： |
| `2466h` | ref | うわっ！ |
| `2784h` | ref | ｇｐで引き取ろう。 |
| `2797h` | ref | それでいいかね？ |
| `27A8h` | ref | 毎度ありがとう |
| `27B7h` | ref | 持ちすぎだ。お金は地面に置いておくよ |
| `29BEh` | ref | の鑑定料は２００ｇｐだ。 |
| `29D7h` | ref | それでいいかね？ |
| `29E8h` | ref | お金が足りません |
| `29F9h` | ref | その |
| `29FEh` | ref | については、特に目新しい事実はないですな |
| `2A27h` | ref | のようです |
| `2C14h` | ref | 誰にわたしますか？ |
| `2C2Ah` | ref |  決定  |
| `2C31h` | ref | 貨幣の種類を選んでください |
| `2C4Ch` | ref | 決定 |
| `2C51h` | ref | どれだけ |
| `2C5Ah` | ref | わたしますか？ |
| `2FA8h` | ref |  決定  |
| `2FAFh` | ref | 貨幣の種類を選んでください |
| `2FCAh` | ref | 決定 |
| `2FCFh` | ref | どれだけ |
| `2FD8h` | ref | 捨てますか？ |
| `347Fh` | ref | 記憶内 |
| `3486h` | ref | 魔法書上 |
| `348Fh` | ref | 巻物上 |
| `3496h` | ref | 選んでください |
| `34A5h` | ref | 記憶する魔法を選んでください |
| `34C2h` | ref | 書きこむ魔法を選んでください |
| `34DFh` | ref | 呪文 |
| `3716h` | ref | 誰を回復させますか？ |
| `372Bh` | ref | ALIAS |
| `3731h` | ref | エイリアス |
| `373Ch` | ref | DRAGONBAIT |
| `3747h` | ref | ドラゴンベイト |
| `3756h` | ref | AKABAR BEL AKAS |
| `3766h` | ref | アーカバー・ベル・アーカッシュ |
| `3785h` | ref | は気分がよくなった |
| `3798h` | ref | には影響がなかった |
| `3920h` | ref | 誰を治療しますか？ |
| `3933h` | ref | は病気にかかっていない |
| `394Ah` | ref | それでも治療しますか？ |
| `3961h` | ref | ALIAS |
| `3967h` | ref | エイリアス |
| `3972h` | ref | DRAGONBAIT |
| `397Dh` | ref | ドラゴンベイト |
| `398Ch` | ref | AKABAR BEL AKAS |
| `399Ch` | ref | アーカバー・ベル・アーカッシュ |
| `39BBh` | ref | は治療を受けた |

### overlay-20

| offset | 依據 | 內容 |
|---|---|---|
| `05C7h` | ref | 休息時間 |
| `06C9h` | ref |       日     時     分    休む            増やす 減らす               やめる     |
| `0883h` | ref | パーティーは回復した |
| `0945h` | ref | を覚えた |
| `0A04h` | ref | を書き写した |
| `0C9Ah` | ref | 休息を中断しますか？ |
| `0CAFh` | ref | 休んでいるどころではなくなった！ |

### overlay-21

| offset | 依據 | 內容 |
|---|---|---|
| `01F3h` | ref | 持ちすぎです。お金は地面に置きますよ |
| `0478h` | ref | 持ちすぎです |
| `0A99h` | ref | 持ちすぎです |
| `0B74h` | ref | 宝石飾り |
| `0B7Dh` | ref | 宝石 |
| `0B82h` | ref | ｇｐ |
| `0B87h` | ref | ｐｐ |
| `0B8Ch` | ref | ｅｐ |
| `0B91h` | ref | ｓｐ |
| `0B96h` | ref | ｃｐ |
| `0DA5h` | ref |  決定  |
| `0DACh` | ref | 貨幣の種類を選んでください |
| `0DC7h` | ref | 決定 |
| `0DCCh` | ref | どれだけ |
| `0DD5h` | ref | 持っていきますか？ |
| `1948h` | ref | 宝石も宝石飾りも持っていませんよ |
| `1969h` | ref | 宝石 |
| `196Eh` | ref | 宝石飾り |
| `1977h` | ref | あなたのお持ちになっているのは |
| `1996h` | ref |  宝石  |
| `199Dh` | ref | 宝石飾 |
| `19A4h` | ref | 抜ける |
| `19ABh` | ref | 見積る： |
| `19B4h` | ref | 宝石の値段は |
| `19C1h` | ref | ｇｐです。 |
| `19CCh` | ref |    売る                                  |
| `19F5h` | ref |    売る　                      売ﾗﾅｲ     |
| `1A1Eh` | ref | どうしますか？ |
| `1A2Dh` | ref | 宝石飾りの値段は |

### overlay-22

| offset | 依據 | 內容 |
|---|---|---|
| `01AFh` | ref | 呪文を選んでください |
| `01C4h` | ref | かける |
| `01CBh` | ref | 記憶ｽﾙ |
| `01D2h` | ref |  書ｷｺﾑ |
| `01D9h` | ref |  学ぶ  |
| `01E0h` | ref |        |
| `111Bh` | ref | 誰にかけますか？ |
| `1394h` | ref | はここではかけられない。 |
| `13ADh` | ref | 無駄に使いますか？ |
| `13C0h` | ref | そのアイテムは |
| `13CFh` | ref | 戦闘専用だ。 |
| `13DCh` | ref | それでも使いますか？ |
| `13F1h` | ref | は呪文を唱えそこねた |
| `1406h` | ref | の呪文をかけた。 |
| `1417h` | ref | 呪文を中断しますか？ |
| `142Ch` | ref | 呪文は中断された |
| `1F42h` | ref | は祝福を受けた。 |
| `1F7Ah` | ref | は呪いを受けた。 |
| `202Ah` | ref | は魔法にかかった。 |
| `206Ah` | ref | は守られた。 |
| `20A4h` | ref | は冷気に対する防護を得た。 |
| `2121h` | ref | は影響を受けなかった。 |
| `2138h` | ref | は魅了された。 |
| `2204h` | ref | は強くなった。 |
| `2213h` | ref | は影響を受けなかった。 |
| `234Dh` | ref | は小さくなった。 |
| `23F1h` | ref | は友好的になった。 |
| `24AAh` | ref | は魔法の楯に守られた。 |
| `2536h` | ref | は眠りに落ちた。 |
| `267Dh` | ref | は金縛りにあった。 |
| `26E1h` | ref | は火炎に対する防護を得た。 |
| `2729h` | ref | は沈黙した。 |
| `2763h` | ref | は魔法にかかった。 |
| `281Eh` | ref | は魅了された。 |
| `2922h` | ref | は透明になった。 |
| `2960h` | ref | こんこん |
| `2996h` | ref | は分身した。 |
| `29F5h` | ref | は弱くなった。 |
| `2A31h` | ref | はひどい臭いのガスを作り出した。 |
| `3063h` | ref | は動き始めた。 |
| `3212h` | ref | の視力は戻った。 |
| `3260h` | ref | は視力を奪われた。 |
| `3356h` | ref | は病いに冒された。 |
| `3448h` | ref | は魔法にかかった。 |
| `37F5h` | ref | は祈っている。 |
| `384Dh` | ref | の呪いはとけた。 |
| `385Eh` | ref | のアイテムの呪いは消えた。 |
| `3976h` | ref | は呪われた！ |
| `39B0h` | ref | は点滅している。 |
| `3BBCh` | ref | は加速された。 |
| `3F99h` | ref | は減速された。 |
| `3FD6h` | ref | はレベルを回復した。 |
| `419Ch` | ref | はすばやくなった。 |
| `4230h` | ref | は強くなった。 |
| `4306h` | ref | は麻痺した。 |
| `4340h` | ref | は回復した。 |
| `439Bh` | ref | は透明になった。 |
| `444Fh` | ref | の毒は拭われた。 |
| `4460h` | ref | は影響を受けなかった。 |
| `45E2h` | ref | は叩きつけた。 |
| `4706h` | ref | は魔法にかかった。 |
| `47B0h` | ref | は蘇生した。 |
| `4888h` | ref | は死んだ。 |
| `4893h` | ref | は影響を受けなかった。 |
| `495Ah` | ref | は絡みつかれた。 |
| `4A16h` | ref | に光がまとわりついた。 |
| `4A4Dh` | ref | は透明になった。 |
| `4A8Bh` | ref | は魅了された。 |
| `4B1Fh` | ref | は混乱した。 |
| `4BE4h` | ref | はテレポートした。 |
| `4D85h` | ref | は恐怖にかられ、逃げ出した。 |
| `4DA2h` | ref | は影響を受けなかった。 |
| `4EDDh` | ref | 炎の種別は？ |
| `4EEAh` | ref |     熱                          冷       |
| `4F13h` | ref | は守られた。 |
| `4F21h` | ref | 呪文を中断しますか？ |
| `4F36h` | ref |    はい                       いいえ     |
| `5125h` | ref | はのろまになった。 |
| `5138h` | ref | は減速された。 |
| `5291h` | ref | は守られた。 |
| `52CBh` | ref | は毒雲を作り出した。 |
| `5B1Eh` | ref | は回復した。 |
| `5B79h` | ref | は致命の息吹を吹いた！ |
| `5D08h` | ref | は酸を吹いた。 |
| `5E23h` | ref | は酸を吹いた。 |
| `600Ah` | ref | は炎を吹いた。 |
| `619Eh` | ref | は炎を吹いた。 |
| `62C2h` | ref | は電光を投げつけた。 |
| `63BCh` | ref | は睨んだ。 |
| `63C7h` | ref | は麻痺した。 |
| `65AEh` | ref | は呪文を唱えた。 |
| `65BFh` | ref | 呪文 |
| `65C4h` | ref | は、 |

### overlay-23

| offset | 依據 | 內容 |
|---|---|---|
| `0DE8h` | ref | は咳きこみ始めた。 |
| `0DFBh` | ref | は窒息し、喉をかきむしった。 |
| `0E18h` | ref | は毒を受けた。 |
| `0E27h` | ref | は死んだ。 |
| `1623h` | ref | は呪われた。 |
| `1F12h` | ref | ポイントのダメージを受けた。 |
| `1F2Fh` | ref | は火によって |
| `1F3Ch` | ref | は冷気によって |
| `1F4Bh` | ref | は電撃によって |
| `1F5Ah` | ref | は酸によって |
| `1F67h` | ref | は魔法によって |
| `1F76h` | ref | は呪文を唱えられなくなった。 |
| `1F93h` | ref | は倒れた。 |
| `1F9Eh` | ref | そして、瀕死の状態だ。 |
| `1FD5h` | ref | は死んだ。 |
| `1FE0h` | ref | は、ダメージを受けなかった。 |
| `2310h` | ref | には影響がなかった。 |
| `24EAh` | ref | は立ち上がり、ニヤリと笑った。 |
| `2509h` | ref | は起き上がった。 |

### overlay-24

| offset | 依據 | 內容 |
|---|---|---|
| `0452h` | ref |  装備中   |
| `045Ch` | ref |  　　　   |
| `08ABh` | ref | 名前 |
| `08B0h` | ref | ＡＣ　　ＨＰ |
| `08BDh` | ref | ALIAS |
| `08C3h` | ref | エイリアス |
| `08CEh` | ref | DRAGONBAIT |
| `08D9h` | ref | ドラゴンベイト |
| `08E8h` | ref | AKABAR BEL AKAS |
| `08F8h` | ref | アーカバー・ベル・アーカッシュ |
| `0C3Ah` | ref | ヒットポイント |
| `0C49h` | ref | ＡＣ |
| `0C4Eh` | ref | 無力 |
| `177Ah` | ref | アイテム・ポインターがありません |
| `2756h` | ref | は完全に回復した |
| `2767h` | ref | は少し回復した |
| `2AA1h` | ref | 防御している |
| `2E76h` | ref | キャンプ中 |
| `2E81h` | ref | 捜索モード |
| `311Dh` | ref | 抜ける |
| `3124h` | ref |  決定  |
| `35C5h` | ref | は手当てを受けた。 |

### overlay-25

| offset | 依據 | 內容 |
|---|---|---|
| `0E89h` | ref | 新しいクラスを選んでください |
| `0EA6h` | ref | ALIAS |
| `0EACh` | ref | エイリアス |
| `0EB7h` | ref | DRAGONBAIT |
| `0EC2h` | ref | ドラゴンベイト |
| `0ED1h` | ref | AKABAR BEL AKAS |
| `0EE1h` | ref | アーカバー・ベル・アーカッシュ |
| `0F00h` | ref | は不適格です |
| `0F0Dh` | ref |  決定  |
| `0F15h` | ref | 決定 |
| `0F1Ah` | ref | は１経験レベルの |
| `0F2Bh` | ref | です |

### overlay-26

| offset | 依據 | 內容 |
|---|---|---|
| `0851h` | ref |        |
| `0898h` | ref |      |
| `08A0h` | ref |         |
| `0E9Eh` | ref | 　次　 |
| `0EC5h` | ref | 　前　 |
| `0ECCh` | ref | 抜ける |
| `1254h` | ref |    はい                       いいえ     |

### overlay-29

| offset | 依據 | 內容 |
|---|---|---|
| `00D8h` | ref | 読みこみ中です |
| `00EBh` | ref | FINAL |
| `00F1h` | ref | .dax |
| `00F6h` | ref | PIC not found  |
| `0105h` | ref | SPRIT |
| `04C7h` | ref | HEAD |
| `04CCh` | ref | head not found |
| `04DBh` | ref | BODY |
| `064Dh` | ref | Illegal range in Show3DSprite. |
| `0724h` | ref | bigpic |

### overlay-30

| offset | 依據 | 內容 |
|---|---|---|
| `0F44h` | ref | WALLDEF |
| `0F4Ch` | ref | .dax |
| `0F51h` | ref | Unable to load  |
| `0F61h` | ref |  from WALLDEF |
| `122Dh` | ref | .dax |
| `1232h` | ref | Unable to load geo in Load3DMap. |

### overlay-33

| offset | 依據 | 內容 |
|---|---|---|
| `000Ch` | ref | Start range error in Load24x24Set |
| `002Eh` | ref | DungCom |
| `0036h` | ref | WildCom |
| `003Eh` | ref | RandCom |
| `0046h` | ref | tiles |
| `035Ch` | ref | CHEAD |
| `0362h` | ref | CBODY |
| `0368h` | ref | COMSPR |

### overlay-35

| offset | 依據 | 內容 |
|---|---|---|
| `0027h` | ref | 8x8d |
| `002Ch` | ref | .dax |
| `0031h` | ref | Unable to load  |
| `0041h` | ref |  from 8x8D |
| `0307h` | ref | Bad symbol number in Put8x8Symbol. |
