# 內容相同、位址不同的字串常數

由 `scripts/duplicate_strings.py` 產生。編譯器沒有合併相同的字串
常數，所以同一句話在同一個 overlay 裡可能有好幾份。**中文化必須每一份
都改**——只改一份會讓遊戲在某些路徑顯示中文、某些路徑顯示英文，而觸發
條件通常和法術或分支綁在一起，很難重現。

來源掃描是下界（見 `scan_pascal_strings.py`），所以本表也是下界：
**沒列到不代表沒有重複**。

## 同一模組內重複（87 組）

| 平台 | 模組 | 份數 | 位址 | 內容 |
|---|---|---:|---|---|
| dos | overlay-16 | 5 | `05D0h` `0D90h` `222Ah` `3714h` `3E8Eh` | Put save disk in  |
| dos | overlay-22 | 5 | `1EB0h` `1F92h` `4186h` `45A7h` `4A81h` | is unaffected |
| pc98 | overlay-22 | 5 | `2121h` `2213h` `4460h` `4893h` `4DA2h` | は影響を受けなかった。 |
| dos | overlay-22 | 4 | `1DC9h` `24C1h` `3172h` `4426h` | is affected |
| pc98 | overlay-16 | 4 | `0AC4h` `29EDh` `4166h` `497Eh` | ドライブ２にセーブ・ディスクを入れてください |
| pc98 | overlay-22 | 4 | `202Ah` `2763h` `3448h` `4706h` | は魔法にかかった。 |
| dos | overlay-16 | 3 | `0C70h` `0D64h` `2254h` | .guy |
| dos | overlay-16 | 3 | `0C75h` `0DE1h` `2260h` | .swg |
| dos | overlay-16 | 3 | `0604h` `0DA4h` `3EA2h` | Unexpected error during save:  |
| dos | overlay-22 | 3 | `1EBEh` `2575h` `4786h` | is charmed |
| dos | overlay-22 | 3 | `2675h` `40C8h` `474Ch` | is invisible |
| dos | overlay-22 | 3 | `1E02h` `4BC9h` `4EEAh` | is protected |
| pc98 | overlay-12 | 3 | `1512h` `27E3h` `28FFh` | は麻痺した。 |
| pc98 | overlay-16 | 3 | `4284h` `4979h` `4F49h` | .dat |
| pc98 | overlay-16 | 3 | `1253h` `1418h` `2A33h` | .guy |
| pc98 | overlay-16 | 3 | `1258h` `14C0h` `2A3Fh` | .swg |
| pc98 | overlay-16 | 3 | `427Dh` `4972h` `4F42h` | savgam |
| pc98 | overlay-22 | 3 | `206Ah` `4F13h` `5291h` | は守られた。 |
| pc98 | overlay-22 | 3 | `2922h` `439Bh` `4A4Dh` | は透明になった。 |
| pc98 | overlay-22 | 3 | `2138h` `281Eh` `4A8Bh` | は魅了された。 |
| dos | overlay-05 | 2 | `0A5Bh` `1489h` | press <enter>/<return> to continue |
| dos | overlay-12 | 2 | `0C46h` `1CDBh` | The air clears a little... |
| dos | overlay-12 | 2 | `1A45h` `2886h` | gazes... |
| dos | overlay-12 | 2 | `2775h` `288Fh` | is paralyzed |
| dos | overlay-13 | 2 | `302Fh` `3385h` | Center Exit |
| dos | overlay-13 | 2 | `3009h` `3379h` | Range =  |
| dos | overlay-13 | 2 | `3015h` `3391h` | Target  |
| dos | overlay-13 | 2 | `0313h` `42E8h` | is killed |
| dos | overlay-16 | 2 | `36DBh` `3E46h` | .dat |
| dos | overlay-16 | 2 | `0D6Bh` `373Eh` | .sav |
| dos | overlay-16 | 2 | `3517h` `3743h` | CPIC |
| dos | overlay-16 | 2 | `0D1Bh` `3E4Bh` | Can't save.  No room on this disk. |
| dos | overlay-16 | 2 | `223Eh` `3726h` | Loading...Please Wait |
| dos | overlay-16 | 2 | `36D4h` `3E3Fh` | savgam |
| dos | overlay-19 | 2 | `2C34h` `2F87h` |  Select |
| dos | overlay-19 | 2 | `2C3Ch` `2F8Fh` | How much  |
| dos | overlay-19 | 2 | `2799h` `29E0h` | Is It a Deal?  |
| dos | overlay-19 | 2 | `2C1Fh` `2F72h` | Select type of coin  |
| dos | overlay-19 | 2 | `0B2Fh` `346Ch` | Spells  |
| dos | overlay-21 | 2 | `0460h` `0A7Fh` | Overloaded |
| dos | overlay-22 | 2 | `4070h` `5781h` | is Healed |
| dos | overlay-22 | 2 | `4036h` `6015h` | is paralyzed |
| dos | overlay-22 | 2 | `1F86h` `3F63h` | is stronger |
| dos | overlay-23 | 2 | `0E3Bh` `1FCCh` | is killed |
| dos | overlay-30 | 2 | `1007h` `1314h` | .dax |
| pc98 | overlay-02 | 2 | `335Fh` `39EDh` | データの整理をします リターン・キーを押してください |
| pc98 | overlay-05 | 2 | `0A32h` `14DCh` | リターン・キーを押してください           |
| pc98 | overlay-07 | 2 | `005Fh` `21BDh` | AKABAR BEL AKAS |
| pc98 | overlay-07 | 2 | `0034h` `2192h` | ALIAS |
| pc98 | overlay-07 | 2 | `0045h` `21A3h` | DRAGONBAIT |
| pc98 | overlay-07 | 2 | `006Fh` `21CDh` | アーカバー・ベル・アーカッシュ |
| pc98 | overlay-07 | 2 | `003Ah` `2198h` | エイリアス |
| pc98 | overlay-07 | 2 | `0050h` `21AEh` | ドラゴンベイト |
| pc98 | overlay-12 | 2 | `0A0Eh` `1840h` | は狂暴化した。 |
| pc98 | overlay-12 | 2 | `1A8Bh` `28F4h` | は睨んだ。 |
| pc98 | overlay-12 | 2 | `0C69h` `1D45h` | 空気は多少きれいになった |
| pc98 | overlay-16 | 2 | `060Eh` `07D6h` | *.hil |
| pc98 | overlay-16 | 2 | `141Fh` `49BCh` | .sav |
| pc98 | overlay-16 | 2 | `3DFAh` `49C1h` | CPIC |
| pc98 | overlay-16 | 2 | `4159h` `4FE5h` | SaveList.EST |
| pc98 | overlay-16 | 2 | `42EDh` `46A8h` | よろしいですか？ |
| pc98 | overlay-16 | 2 | `0AF1h` `14A1h` | セーブ中にエラーが発生しました |
| pc98 | overlay-16 | 2 | `139Fh` `4F4Eh` | ディスクに余裕がありませんので、セーブできません |
| pc98 | overlay-16 | 2 | `0AA9h` `1486h` | ディスクに異常があります。 |
| pc98 | overlay-16 | 2 | `1461h` `4F9Fh` | ライトプロテクトをはずしてください。 |
| pc98 | overlay-16 | 2 | `2A1Ah` `49ABh` | 読みこみ中です |
| pc98 | overlay-17 | 2 | `0ABFh` `2BDEh` | このキャラクタ名は受け付けられません。 |
| pc98 | overlay-19 | 2 | `3756h` `398Ch` | AKABAR BEL AKAS |
| pc98 | overlay-19 | 2 | `372Bh` `3961h` | ALIAS |
| pc98 | overlay-19 | 2 | `373Ch` `3972h` | DRAGONBAIT |
| pc98 | overlay-19 | 2 | `2797h` `29D7h` | それでいいかね？ |
| pc98 | overlay-19 | 2 | `2C51h` `2FCFh` | どれだけ |
| pc98 | overlay-19 | 2 | `3766h` `399Ch` | アーカバー・ベル・アーカッシュ |
| pc98 | overlay-19 | 2 | `3731h` `3967h` | エイリアス |
| pc98 | overlay-19 | 2 | `3747h` `397Dh` | ドラゴンベイト |
| pc98 | overlay-19 | 2 | `2C31h` `2FAFh` | 貨幣の種類を選んでください |
| pc98 | overlay-21 | 2 | `0B74h` `196Eh` | 宝石飾り |
| pc98 | overlay-21 | 2 | `0478h` `0A99h` | 持ちすぎです |
| pc98 | overlay-22 | 2 | `4340h` `5B1Eh` | は回復した。 |
| pc98 | overlay-22 | 2 | `2204h` `4230h` | は強くなった。 |
| pc98 | overlay-22 | 2 | `3F99h` `5138h` | は減速された。 |
| pc98 | overlay-22 | 2 | `600Ah` `619Eh` | は炎を吹いた。 |
| pc98 | overlay-22 | 2 | `5D08h` `5E23h` | は酸を吹いた。 |
| pc98 | overlay-22 | 2 | `4306h` `63C7h` | は麻痺した。 |
| pc98 | overlay-22 | 2 | `1417h` `4F21h` | 呪文を中断しますか？ |
| pc98 | overlay-23 | 2 | `0E27h` `1FD5h` | は死んだ。 |
| pc98 | overlay-30 | 2 | `0F4Ch` `122Dh` | .dax |

## 跨模組重複（76 句）

同一句話出現在多個 overlay。這類不會互相影響，但翻譯要一致，
否則同一句在不同場景會有兩種譯法。

| 平台 | 模組數 | 模組 | 內容 |
|---|---:|---|---|
| dos | 6 | overlay-05 overlay-14 overlay-15 overlay-21 overlay-24 overlay-26 |  Exit |
| pc98 | 6 | overlay-02 overlay-07 overlay-16 overlay-29 overlay-30 overlay-35 | .dax |
| dos | 5 | overlay-02 overlay-07 overlay-16 overlay-29 overlay-30 | .dax |
| pc98 | 5 | overlay-07 overlay-17 overlay-19 overlay-24 overlay-25 | AKABAR BEL AKAS |
| pc98 | 5 | overlay-07 overlay-17 overlay-19 overlay-24 overlay-25 | ALIAS |
| pc98 | 5 | overlay-07 overlay-17 overlay-19 overlay-24 overlay-25 | DRAGONBAIT |
| pc98 | 5 | overlay-07 overlay-17 overlay-19 overlay-24 overlay-25 | アーカバー・ベル・アーカッシュ |
| pc98 | 5 | overlay-07 overlay-17 overlay-19 overlay-24 overlay-25 | エイリアス |
| pc98 | 5 | overlay-07 overlay-17 overlay-19 overlay-24 overlay-25 | ドラゴンベイト |
| dos | 4 | overlay-07 overlay-11 overlay-16 overlay-29 | Loading...Please Wait |
| dos | 4 | overlay-17 overlay-21 overlay-24 overlay-25 | Select |
| pc98 | 4 | overlay-16 overlay-17 overlay-22 overlay-26 |    はい                       いいえ     |
| pc98 | 4 | overlay-12 overlay-13 overlay-22 overlay-23 | は死んだ。 |
| pc98 | 4 | overlay-07 overlay-11 overlay-16 overlay-29 | 読みこみ中です |
| dos | 3 | overlay-02 overlay-07 overlay-16 | CPIC |
| dos | 3 | overlay-08 overlay-15 overlay-19 | Exit |
| dos | 3 | overlay-12 overlay-13 overlay-23 | is killed |
| dos | 3 | overlay-12 overlay-13 overlay-23 | lost a spell |
| dos | 3 | overlay-02 overlay-05 overlay-07 | press <enter>/<return> to continue |
| dos | 3 | overlay-04 overlay-05 overlay-06 | ~Yes ~No |
| pc98 | 3 | overlay-02 overlay-07 overlay-16 | CPIC |
| pc98 | 3 | overlay-04 overlay-06 overlay-19 | お金が足りません |
| pc98 | 3 | overlay-12 overlay-13 overlay-23 | は倒れた。 |
| pc98 | 3 | overlay-06 overlay-19 overlay-21 | 持ちすぎです |
| dos | 2 | overlay-16 overlay-33 | CBODY |
| dos | 2 | overlay-16 overlay-33 | CHEAD |
| dos | 2 | overlay-11 overlay-33 | COMSPR |
| dos | 2 | overlay-17 overlay-19 | Drop  |
| dos | 2 | overlay-19 overlay-21 | How much  |
| dos | 2 | overlay-05 overlay-06 | Items:  |
| dos | 2 | overlay-08 overlay-09 | Magic Off |
| dos | 2 | overlay-08 overlay-09 | Magic On |
| dos | 2 | overlay-08 overlay-09 | Move/Attack, Move Left =  |
| dos | 2 | overlay-19 overlay-21 | Overloaded |
| dos | 2 | overlay-19 overlay-21 | Select type of coin  |
| dos | 2 | overlay-12 overlay-22 | Spits Acid |
| dos | 2 | overlay-30 overlay-35 | Unable to load  |
| dos | 2 | overlay-22 overlay-26 | Yes No |
| dos | 2 | overlay-12 overlay-22 | gazes... |
| dos | 2 | overlay-12 overlay-23 | is Poisoned |
| dos | 2 | overlay-12 overlay-13 | is Stoned |
| dos | 2 | overlay-12 overlay-22 | is confused |
| dos | 2 | overlay-12 overlay-22 | is paralyzed |
| dos | 2 | overlay-12 overlay-22 | is silenced |
| dos | 2 | overlay-12 overlay-22 | is unaffected |
| dos | 2 | overlay-12 overlay-22 | is weakened |
| pc98 | 2 | overlay-16 overlay-33 | CBODY |
| pc98 | 2 | overlay-16 overlay-33 | CHEAD |
| pc98 | 2 | overlay-11 overlay-33 | COMSPR |
| pc98 | 2 | overlay-10 overlay-33 | DungCom |
| pc98 | 2 | overlay-10 overlay-33 | RandCom |
| pc98 | 2 | overlay-07 overlay-29 | SPRIT |
| pc98 | 2 | overlay-30 overlay-35 | Unable to load  |
| pc98 | 2 | overlay-10 overlay-33 | WildCom |
| pc98 | 2 | overlay-03 overlay-33 | tiles |
| pc98 | 2 | overlay-19 overlay-21 | どれだけ |
| pc98 | 2 | overlay-12 overlay-22 | は影響を受けなかった。 |
| pc98 | 2 | overlay-12 overlay-23 | は毒を受けた。 |
| pc98 | 2 | overlay-12 overlay-22 | は混乱した。 |
| pc98 | 2 | overlay-12 overlay-22 | は睨んだ。 |
| pc98 | 2 | overlay-12 overlay-13 | は石になった。 |
| pc98 | 2 | overlay-09 overlay-13 | は退散させられた。 |
| pc98 | 2 | overlay-12 overlay-22 | は酸を吹いた。 |
| pc98 | 2 | overlay-12 overlay-22 | は麻痺した。 |
| pc98 | 2 | overlay-05 overlay-19 | アイテム |
| pc98 | 2 | overlay-06 overlay-19 | アイテム： |
| pc98 | 2 | overlay-15 overlay-17 | ゲームを終わりますか？ |
| pc98 | 2 | overlay-02 overlay-07 | リターン・キーを押してください |
| pc98 | 2 | overlay-05 overlay-18 | 何かキーを押してください |
| pc98 | 2 | overlay-04 overlay-06 | 引き返してお金を取りますか？ |
| pc98 | 2 | overlay-16 overlay-17 | 新しい名前は？ |
| pc98 | 2 | overlay-08 overlay-09 | 移動／攻撃，残り移動力＝ |
| pc98 | 2 | overlay-19 overlay-21 | 貨幣の種類を選んでください |
| pc98 | 2 | overlay-08 overlay-09 | 魔法不使用 |
| pc98 | 2 | overlay-08 overlay-09 | 魔法使用 |
| pc98 | 2 | overlay-05 overlay-08 | ｷｬﾗｸﾀｰ |
