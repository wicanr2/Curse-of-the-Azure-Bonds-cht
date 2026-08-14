# 1082 — 結局場景：整段是無分支的直線敘事，兩平台的日／英台詞完整對照

- 證據等級：`exact`（DOS 側 649 條、PC-98 側 674 條，兩邊各自逐條讀完）
- 作法見 spec 783

## `dos overlay-18:010FFh`（`retf`）↔ `pc98 overlay-18:01213h`（`retf`）

兩側原本都是待解讀。這是**打倒 Tyranthraxus、青色枷解除、傳送到 Shadowdale
慶功宴**的**結局過場**。

⚠ 兩支的助憶碼序列相似度只有 **0.059**——**不是編譯差異，是整段重寫**。
DOS 有 649 條、PC-98 674 條，但 DOS **一個分支都沒有**
（指令只有 `push`／`mov`／`call`／`lea`／`xor`／`sub`／`les`／`shl`／`retf`／`pop`），
PC-98 多了開頭一段三行的條件。⇒ 兩邊要各自照抄，不能拿一邊套另一邊。

## 骨架

```pascal
DS:4FBBh := DS:4FBAh;   DS:4FBAh := 7;          { 存畫面模式，切成 7 }
…印一段…  印('Press any key to continue.')  等鍵…
…印一段…  等鍵…
bank0^[3FEh] := 1;                              { spec 1077 的觸發旗標 }
<near 0B5Fh>(…);                                { 動畫／音效 }
bank0^[3FEh] := 0;
<overlay-29 entry#7>(41h, 41h);
<overlay-29 entry#6>(1, 3, 3);
…印一段…  等鍵…
<overlay-29 entry#9>('z');                      { ★ 換 bigpic 'z' }
<overlay-29 entry#10>();
…印最後四段…
<near 09DBh>();
DS:4FBAh := DS:4FBBh;                            { 還原畫面模式 }
<overlay-24 entry#37>();
```

★ 印字一律是 `0542:04A4h`（PC-98 `0418:0E6Ah`），
八個參數 **`(1, 11h, 26h, 16h, 0Ah, 0, 首行旗標, 字串)`**——
只有每一段的**第一行**把旗標推 `1`，同段後續各行推 `0`。
★ 「按任意鍵」用 `0542:0946h`（PC-98 `0418:12EFh`），等鍵用
`overlay-26 entry#4`（PC-98 `entry#6`）。
★ `<near 0B5Fh>` 在 DOS 被叫兩次：`(1, 4Ah, 3, 3)` 與
`((0Ah − DS:4FA9h) × 2, 4Dh, 3, 3)`——`4Ah` ＝ `'J'`、`4Dh` ＝ `'M'`。
`DS:4FA9h` 是 spec 1076 從 `bank0^[1F8h]` 讀進來的那一格。

## ★★ PC-98 多出來的開頭

```pascal
DS:7F28h := DS:7F27h;  DS:7F27h := 7;  DS:7F16h := 9;
if DS:7F2Ch = 0 then begin
    DS:7F2Ch := 1;                               { ★ 0 → 1 }
    bank0^[1FEh] := bank0^[1FEh] xor 1;          { ★ 同一格的低位元翻面 }
    DS:0A2C6h := DS:0A2C7h;
end;
```

★ `DS:7F2Ch` 與 `bank0^[1FEh]` 的關係在 spec 1072 已定案
（`7F2Bh := bank0^[1FEh] div 2`、`7F2Ch := bank0^[1FEh] and 1`）
——這裡是**把那個奇偶位翻過來並同步兩邊**，形狀上是切換到另一半畫面。
⚠ 那段 `cmp DS:7F2Ch, 0 / jnz` 之後又立刻 `cmp DS:7F2Ch, 0 / jz`，
第二次判斷**結果一定相同**，等於固定走 `1` 那一支——原作的冗餘寫法。

## ★★★ 完整台詞對照（照播放順序）

| # | DOS（`CS:` 位址／長度） | PC-98（`CS:` 位址／bytes） |
|---|---|---|
| 1 | `0C6Ah` 46 `Tyranthraxus' spirit coalesces over the slain ` | `0D13h` 66 「ティランスラクサスの霊体がストーム・ジャイアントの死体に重なった。」 |
| 2 | `0C99h` 52 `storm giant. 'You have defeated me. Were it not for ` | `0D56h` 78 「よくぞ余を倒した。『ラサンダーの魔除け』さえなければ、汝らの身体を乗っ取り、」 |
| 3 | `0CCEh` 53 `the Amulet of Lythander, I could possess you and rob ` | `0DA5h` 34 「勝利をもぎとることもできたのだが。」 |
| 4 | `0D04h` 57 `you of your victory. Still I can escape through the pool.` | `0DC8h` 50 「しかし、まだ〈プール〉を通じて逃げることはできる」 |
| — | `0D3Eh` 26 `Press any key to continue.` | `0DFBh` 24 「何かキーを押してください」 |
| 5 | `0D59h` 48 `As you reach for the Pool of Radiance, he cries ` | `0E14h` 82 「きみたちが〈プール・オブ・レイディアンス〉に近づくと、ティランスラクサスは叫んだ。」 |
| 6 | `0D8Ah` 56 `out, 'Keep the Gauntlet of Moander away from there, you ` | `0E67h` 84 「『モーンダーの篭手』を近づけるな！　危険なエネルギーを解放しようとしているのだぞ！」 |
| 7 | `0DC3h` 52 `will unleash dangerous energies. Stay back!' As the ` | （併入上一段） |
| 8 | `0DF8h` 57 `gauntlet contacts the pool, it contracts and shatters it.` | `0EBCh` 66 「　篭手が〈プール〉に触れた瞬間、〈プール〉は収縮し、篭手は砕けた。」 |
| 9 | `0E32h` 49 `'I am trapped without escape, you have succeeded ` | `0F00h` 30 「「余は、逃れることができぬか。」 |
| 10 | `0E64h` 57 `where armies have not. Gloat while you may, Tyranthraxus ` | `0F1Fh` 52 「汝らは軍隊でも成しえなかったことを成し遂げたわけだ。」 |
| 11 | — | `0F54h` 34 「喜ぶがよい。笑みを浮かべるがよい。」 |
| 12 | `0E9Eh` 54 `is slain this day.' Before your eyes he crumbles into ` | `0F77h` 92 「この日、ティランスラクサスは滅した」　きみたちの前で、ティランスラクサスは縮み、姿を消した。」 |
| 13 | `0ED5h` 12 `nothingness.` | （併入上一段） |
| 14 | `0EE2h` 45 `You are certain he is destroyed because your ` | `0FD4h` 46 「ティランスラクサスが倒れたことは間違いがない。」 |
| 15 | `0F10h` 52 `final bond fades away. The Curse of the Azure Bonds ` | `1003h` 68 「というのも、きみたちの最後の〈紺青の呪縛〉がついに消え去ったからだ。」 |
| 16 | `0F45h` 50 `has finally been lifted from you! You are free at ` | `1048h` 30 「きみたちはついに、解放された！」 |
| 17 | `0F78h` 5 `last!` | （併入上一段） |
| 18 | `0F7Eh` 38 `The Knights of Myth Drannor rush in, '` | `1067h` 52 「ミス・ドラノールの騎士たちが部屋になだれこんできた。」 |
| 19 | `0FA5h` 52 `Congratulations, you have destroyed the Flamed One. ` | `109Ch` 56 「「おめでとう。きみたちはついに『炎のもの』を打ち倒したな」 |
| 20 | `0FDAh` 50 `With the power of Elminster, let us take you from ` | `10D5h` 70 「エルミンスターの力をもって、この忌まわしい場所から脱出させてあげよう。」 |
| 21 | `100Dh` 35 `this  foul place, to a fine feast.'` | `111Ch` 14 「そして、宴だ」」 |
| 22 | `1031h` 52 `You are teleported to Shadowdale, where festivities ` | `112Bh` 44 「きみたちはシャドウデイルへとテレポートした。」 |
| 23 | `1066h` 58 `have already begun. A huge cheer goes up at your arrival. ` | `1158h` 80 「そこでは、すでに祭宴が始まっていた。きみたちが着くと、大きな喚声が巻き起こった。」 |
| 24 | `10A1h` 53 `Gharri and Nacacia, arm in arm, yell congratulations ` | `11A9h` 80 「近くのさじきから、ジャーリーとナカシアが腕を組み、きみたちにおめでとうを叫んだ。」 |
| 25 | `10D7h` 39 `from the nearby stands. 'You have won!'` | `11FAh` 24 「きみたちは勝利したのだ！」 |

★ `CS:0EFFh` 是一個**長度 0 的空字串**，PC-98 拿它當空行。
★ DOS 是「一句話切成 45..57 bytes 的固定行」，**行尾的空白就是斷行處**；
PC-98 改成「一句一行」，所以行數不一樣（DOS 25 行、PC-98 22 行）。

## 中文化

⚠ **這是全遊戲的結局，語氣要一次到位。** 建議照 PC-98 的分句（一句一行），
因為 DOS 的斷行是英文的音節切點，直接照搬會在中文裡切在奇怪的位置。

| 建議中文 | 對應 |
|---|---|
| 「提蘭斯拉克蘇斯的靈體，覆上了風暴巨人的屍身。」 | 1 |
| 「你竟擊敗了我。若不是有萊桑德護符，我早已占據你的軀體，奪走你的勝利。」 | 2–3 |
| 「不過，我仍能從池中脫身。」 | 4 |
| 「請按任意鍵繼續」 | 提示 |
| 「當你伸手探向光輝之池，他大喊：」 | 5 |
| 「『別把莫安德手甲靠近那裡！你會釋放出危險的能量，退後！』」 | 6–7 |
| 「手甲一觸及池水，池子便收縮、碎裂。」 | 8 |
| 「『我逃不掉了。你們做到了千軍萬馬做不到的事。』」 | 9–10 |
| 「『盡情歡呼吧。今日，提蘭斯拉克蘇斯殞落。』你眼前，他化為烏有。」 | 11–13 |
| 「你確信他已被消滅——因為你最後一道枷印，就此淡去。」 | 14–15 |
| 「青色枷的詛咒，終於從你身上解除了。你自由了！」 | 16–17 |
| 「密斯卓諾的騎士們衝了進來：」 | 18 |
| 「『恭喜，你們打倒了炎之子。』」 | 19 |
| 「『就讓艾米斯特的力量，帶你離開這汙穢之地，去赴一場盛宴吧。』」 | 20–21 |
| 「你被傳送到影谷，慶典早已開始。你一到，歡聲如雷。」 | 22–23 |
| 「加里與娜卡希亞手挽著手，從看臺上高喊祝賀：『你們贏了！』」 | 24–25 |

⚠ **每一段的第一行要把印字的第 7 個參數推 `1`，其餘推 `0`**——
中文重新斷行時，這個旗標要跟著新的段落起點走，否則會少一次清畫面。
⚠ 專有名詞：Tyranthraxus／the Flamed One、Amulet of Lythander、
Gauntlet of Moander、Pool of Radiance、Myth Drannor、Elminster、Shadowdale、
Gharri、Nacacia——**要與全遊戲其他處的譯名表一致**。
⚠ DOS 那句 `'this  foul place'` 中間是**兩個空白**，是原文的排版瑕疵，不必照抄。

## 明確不宣稱

- 沒有宣稱印字那八個參數（`1, 11h, 26h, 16h, 0Ah, 0, …`）各自的意義
  （形狀上像是視窗矩形 ＋ 顏色，但本規格沒有讀 `0542:04A4h`）。
- 沒有宣稱 `near 0B5Fh`（DOS）／`near 0A6Fh`（PC-98）播的是什麼
  （參數含 `'J'`／`'M'` 兩個字元）。
- 沒有宣稱 `near 09DBh`（DOS）／`near 08EBh`＋`near 0B9Bh`（PC-98）的收尾動作。
- 沒有宣稱 `overlay-29 entry#6／#7／#10` 的介面。
- 沒有宣稱 `overlay-24 entry#37`（兩平台都在最後呼叫）做什麼。
- 沒有宣稱 PC-98 開頭那段翻轉 `bank0^[1FEh]` 低位元的用途。
- 沒有宣稱 `DS:7F16h := 9`（PC-98）與 `DS:8BF3h := 0Ah`（PC-98 中段）的意義。
