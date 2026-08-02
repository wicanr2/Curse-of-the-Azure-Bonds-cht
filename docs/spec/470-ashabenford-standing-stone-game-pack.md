# 第四百七十輪：阿沙本福德與立石群事件資料化

狀態：`READY`

## 範圍

本輪把十一個作品文字邊界由 Go／舊 UI locale 移入 CoAB game-pack：

- 提爾隘口飛行身影、阿沙本福德 edge／places／河畔酒館。
- 酒館傳聞 28。
- 暗影隘口偽裝火刀伏擊。
- 立石群灰袍男子、四位主人與「往南尋找紅色之人」。
- 前往艾森布拉與哈普的 route prompt。

每條規則保存原始 ECL fragments 與 en／zh-TW 訊息。酒館傳聞 28 是 CoAB
作品內容，不再列入共用酒館 UI catalog；UI 只負責顯示事件文字。

## 正常玩家路徑

既有長回歸由提爾佛頓離場開始，依序驗證：

1. TRAIL 經提爾隘口，迎戰八隻鷹馬後抵達 Ashabenford edge。
2. 進城 PICTURE 80、城鎮 selector `06h`、places、河畔酒館與傳聞 28。
3. 酒館返回 places 不重播進城音樂，離城切回 wilderness selector `05h`。
4. 前往 Standing Stone 的暗影隘口六名火刀伏擊。
5. 灰袍男子的三個文字 boundary、`THANK HIM／ATTACK／LEAVE` 與 world value 4。
6. Journey On 的 Essembra route prompt、到城，再由 Essembra 驗證 Hap route
   prompt 與既有三黑龍事件銜接。

## 驗證與邊界

- 十一條規則在 en／zh-TW 均命中唯一 rule ID、非空訊息且不附帶手札頁。
- 玩家路徑期望由 stable ID 取得，不複製目前譯文。
- Go 漢字 literal exact baseline 由 603 降至 594；`localization_debt`
  108→99，frontend 135、runtime 360 不變。河畔酒館與傳聞舊分支只回傳 locale
  key，因此事件數與下降數不同。

本輪不代表阿沙本福德所有服務／傳聞、立石群所有主人狀態、鷹馬／火刀完整
能力、動態戰鬥演出、音效或世界旅行系統全部完成。
