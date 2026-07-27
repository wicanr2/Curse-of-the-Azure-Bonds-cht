# Gold Box audio knowledge

跨作品可重用的 audio layer 應分成三層：reference sound selector、raw WAV asset catalog、平台 playback adapter。遊戲 state 只發出 deterministic sound intent，不能在 rules／ECL VM 內直接建立 audio device。

CoAB reference `seg044.PlaySound` 使用 `Sound` selector：`2` missile、`3` magic hit、`5` death、`6` generic sample、`7` hit、`9` miss、`A` step、`B` generic sample、`D` start。已從 reference `Main/sounds` 保存 9 個 WAV，`internal/sound` 提供 selector mapping 與 Ebiten player；未被 resource table 證實的 selector 維持 no-op。

目前 adapter 已接入 title start、荒野移動、dungeon preview step，以及 State 發出的戰鬥命中／未命中／死亡／已實作法術 intent。後續 ECL routine 可沿用同一個 selector boundary；尚未因名稱或相鄰音效自行猜測完整音樂、MIDI、AdLib 或作品間共用的音效語意。
