# 英日字串對照（由函式配對 ＋ 引用序位推得）

每一列的依據有兩層：兩支函式的助憶碼序列完全相同（結構同構），
以及該字串是兩邊 body 裡**第幾個** `mov di, offset` 的目標（同一條指令）。
不是按模組內的出現順序對——英文有單複數分歧、日文沒有，條數不同，
按序號對會整段錯位。

沒配上的字串不會出現在這裡，**這是下界不是全集**。


## 日文版不是逐句直譯

`overlay-08` 有一個確認過的例子：DOS `0ED4h` 的 `Not with that weapon`
（武器不對）在 PC-98 是 `0F06h`「そこへは進めない」（過不去）。兩者是同一支
函式（助憶碼序列 67 條指令完全相同）裡的同一條 `mov di, offset`，所以配對
沒錯——是**日文版換了說法**。同一個模組裡另有 `can't go there` ↔
「そこへは行けない」，可見日文版把兩種情境併成相近的兩句。

**中文化以 DOS 英文為準**：英文是原作，日文是再創作。這張表的日文欄是輔助
理解語境用的，不是翻譯來源。

| 模組 | DOS | PC-98 | 英文 | 日文 |
|---|---|---|---|---|
| overlay-00 | `0000h` | `0000h` | Time to save your game | ゲームをセーブしてください |
| overlay-04 | `0263h` | `0273h` | is not blind. | は盲目ではありませんよ。 |
| overlay-04 | `0271h` | `028Ch` | Cure Blindness | キュア・ブラインドネス |
| overlay-04 | `0306h` | `0329h` | is not Diseased. | は病いに冒されてはいませんよ。 |
| overlay-04 | `0317h` | `0348h` | Cure Disease | キュア・ディジーズ |
| overlay-04 | `086Fh` | `090Ah` | is not poisoned. | は毒を受けてはいませんよ |
| overlay-04 | `0880h` | `0923h` | Neutralize Poison | ニュートラライズ・ポイズン |
| overlay-04 | `095Fh` | `0A0Bh` | is not cursed. | は呪われてはいませんよ |
| overlay-04 | `096Eh` | `0A22h` | Remove Curse | リムーブ・カース |
| overlay-06 | `034Ah` | `0393h` | OverLoaded | 持ちすぎです |
| overlay-06 | `0462h` | `04ADh` | Not enough Money. | お金が足りません |
| overlay-08 | `0ED4h` | `0F06h` | Not with that weapon | そこへは進めない |
| overlay-10 | `1017h` | `100Bh` | DungCom | DungCom |
| overlay-10 | `101Fh` | `1013h` | WildCom | WildCom |
| overlay-10 | `1027h` | `101Bh` | RandCom | RandCom |
| overlay-12 | `00D0h` | `00D0h` | is fighting with snakes | は蛇と戦っている。 |
| overlay-12 | `0395h` | `038Ah` | Suffocates | は窒息した。 |
| overlay-12 | `04D8h` | `04CFh` | is silenced | は沈黙させられた。 |
| overlay-12 | `0539h` | `0537h` | dies from poison | は毒のために死んだ。 |
| overlay-12 | `05A8h` | `05AAh` | Gains an item | はアイテムを得た。 |
| overlay-12 | `075Bh` | `0762h` | lost an image | は幻影を失った。 |
| overlay-12 | `080Ch` | `0816h` | is coughing | は咳きこんでいる。 |
| overlay-12 | `08C3h` | `08D4h` | collapses | は倒れた。 |
| overlay-12 | `0BE4h` | `0BFDh` | ages | は歳をとった。 |
| overlay-12 | `0F2Fh` | `0F50h` | Avoids it | は避けた。 |
| overlay-12 | `1063h` | `1084h` | is weakened | は弱まった。 |
| overlay-12 | `171Bh` | `1756h` | is unaffected | は影響を受けなかった。 |
| overlay-12 | `2031h` | `2099h` | Falls dead | は倒れ、死んだ。 |
| overlay-12 | `2775h` | `27E3h` | is paralyzed | は麻痺した。 |
| overlay-12 | `2CCDh` | `2D45h` | gets zapped | は報復を受けた。 |
| overlay-12 | `2D5Ch` | `2DD9h` | is dispelled | は異界に送り返された。 |
| overlay-12 | `2D69h` | `2DF0h` | resists dispel evil | は「ディスペル・イービル」に耐えた。 |
| overlay-13 | `0D01h` | `0D61h` | Got Away | は逃げきった。 |
| overlay-13 | `0D0Ah` | `0D70h` | Escape is blocked | 逃げられなかった |
| overlay-13 | `2F2Eh` | `2DA0h` | Attack Ally:  | 味方を攻撃するのですか？ |
| overlay-13 | `4286h` | `41E4h` | fires a disintegrate ray | は原子分解光線を発射した。 |
| overlay-13 | `429Fh` | `41FFh` | is disintergrated | は粉々になった。 |
| overlay-13 | `42B1h` | `4210h` | fires a stone to flesh ray | は石化光線を発射した。 |
| overlay-13 | `42CCh` | `4227h` | is Stoned | は石になった。 |
| overlay-13 | `42D6h` | `4236h` | fires a death ray | は殺人光線を発射した。 |
| overlay-13 | `42E8h` | `424Dh` | is killed | は死んだ。 |
| overlay-13 | `42F2h` | `4258h` | wounds you | はきみを傷つけた。 |
| overlay-15 | `04B3h` | `04A3h` | has no spells memorized | は呪文をまったく覚えていない。 |
| overlay-16 | `0A72h` | `1055h` | CHEAD | CHEAD |
| overlay-16 | `0A78h` | `105Bh` | CBODY | CBODY |
| overlay-19 | `0E75h` | `0E23h` | Must be unreadied | 装備をはずしてください |
| overlay-19 | `0E87h` | `0E3Ah` |  was going to scribe from that scroll | はその巻物から呪文を書き写そうとしています。 |
| overlay-19 | `0EADh` | `0E67h` | is it Okay to lose it?  | この巻物からの書き写しを中止しますか？ |
| overlay-19 | `2120h` | `20E4h` | Trade with Whom? | 誰に渡しますか？ |
| overlay-19 | `2131h` | `20F5h` | Overloaded | 持ちすぎです |
| overlay-19 | `21E2h` | `21A8h` | Can't halve that | それは分けることはできません |
| overlay-20 | `05C7h` | `05C7h` | Rest Time: | 休息時間 |
| overlay-20 | `0848h` | `0883h` | The Whole Party Is Healed | パーティーは回復した |
| overlay-20 | `090Fh` | `0945h` | has memorized | を覚えた |
| overlay-20 | `09D3h` | `0A04h` | has scribed | を書き写した |
| overlay-20 | `0C68h` | `0C9Ah` | Stop Resting?  | 休息を中断しますか？ |
| overlay-20 | `0C77h` | `0CAFh` | Your repose is suddenly interrupted! | 休んでいるどころではなくなった！ |
| overlay-21 | `01F4h` | `01F3h` | Overloaded.  Money will be put in Pool. | 持ちすぎです。お金は地面に置きますよ |
| overlay-21 | `0460h` | `0478h` | Overloaded | 持ちすぎです |
| overlay-21 | `0A7Fh` | `0A99h` | Overloaded | 持ちすぎです |
| overlay-22 | `1CEEh` | `1F42h` | is Blessed | は祝福を受けた。 |
| overlay-22 | `1D20h` | `1F7Ah` | is Cursed | は呪いを受けた。 |
| overlay-22 | `2174h` | `23F1h` | is friendly | は友好的になった。 |
| overlay-22 | `22A7h` | `2536h` | falls asleep | は眠りに落ちた。 |
| overlay-22 | `2575h` | `281Eh` | is charmed | は魅了された。 |
| overlay-22 | `26E8h` | `2996h` | is duplicated | は分身した。 |
| overlay-22 | `2781h` | `2A31h` | Creates a noxious cloud | はひどい臭いのガスを作り出した。 |
| overlay-22 | `2DAAh` | `3063h` | is animated | は動き始めた。 |
| overlay-22 | `2F56h` | `3212h` | can see | の視力は戻った。 |
| overlay-22 | `3522h` | `37F5h` | is praying | は祈っている。 |
| overlay-22 | `3913h` | `3BBCh` | is Hasted | は加速された。 |
| overlay-22 | `3CE3h` | `3F99h` | is Slowed | は減速された。 |
| overlay-22 | `3D1Bh` | `3FD6h` | is restored | はレベルを回復した。 |
| overlay-22 | `3ED8h` | `419Ch` | is Speedy | はすばやくなった。 |
| overlay-22 | `3F63h` | `4230h` | is stronger | は強くなった。 |
| overlay-22 | `42FFh` | `45E2h` | smashes them flat | は叩きつけた。 |
| overlay-22 | `4426h` | `4706h` | is affected | は魔法にかかった。 |
| overlay-22 | `4665h` | `495Ah` | is entangled | は絡みつかれた。 |
| overlay-22 | `471Dh` | `4A16h` | is highlighted | に光がまとわりついた。 |
| overlay-22 | `4786h` | `4A8Bh` | is charmed | は魅了された。 |
| overlay-22 | `4816h` | `4B1Fh` | is confused | は混乱した。 |
| overlay-22 | `4A72h` | `4D85h` | runs in terror | は恐怖にかられ、逃げ出した。 |
| overlay-22 | `4A81h` | `4DA2h` | is unaffected | は影響を受けなかった。 |
| overlay-22 | `4D8Ch` | `5125h` | is clumsy | はのろまになった。 |
| overlay-22 | `4D96h` | `5138h` | is slowed | は減速された。 |
| overlay-23 | `0E03h` | `0DE8h` | starts to cough | は咳きこみ始めた。 |
| overlay-23 | `0E13h` | `0DFBh` | chokes and gags from nausea | は窒息し、喉をかきむしった。 |
| overlay-23 | `0E2Fh` | `0E18h` | is Poisoned | は毒を受けた。 |
| overlay-23 | `0E3Bh` | `0E27h` | is killed | は死んだ。 |
| overlay-23 | `163Ah` | `1623h` | is Cured | は呪われた。 |
| overlay-23 | `22F9h` | `2310h` | is Unaffected | には影響がなかった。 |
| overlay-23 | `24CCh` | `24EAh` | stands up and grins | は立ち上がり、ニヤリと笑った。 |
| overlay-23 | `24E0h` | `2509h` | gets back up | は起き上がった。 |
| overlay-24 | `0A1Ch` | `0C3Ah` | Hitpoints | ヒットポイント |
| overlay-24 | `0A29h` | `0C4Eh` | (Helpless) | 無力 |
| overlay-24 | `251Fh` | `2756h` | is fully healed | は完全に回復した |
| overlay-24 | `252Fh` | `2767h` | is partially healed | は少し回復した |
| overlay-24 | `286Eh` | `2AA1h` | Guarding | 防御している |
| overlay-24 | `2B99h` | `2E76h` |  camping | キャンプ中 |
| overlay-24 | `2BA2h` | `2E81h` |  search | 捜索モード |
| overlay-24 | `337Bh` | `35C5h` | is bandaged | は手当てを受けた。 |
| overlay-29 | `0735h` | `064Dh` | Illegal range in Show3DSprite. | Illegal range in Show3DSprite. |
