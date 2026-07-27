package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// ShopOffer is a script/data-layer shop stock entry. ITEMS descriptors do not
// contain prices, so a caller must provide the offer price explicitly.
type ShopOffer struct {
	Item  monster.ItemRecord
	Price uint16
}

type TreasureKind uint8

const (
	TreasureGems TreasureKind = iota + 1
	TreasureJewelry
)

// AppraisalOffers contains explicit shopkeeper offers. The Ready flags keep
// an absent offer distinct from a legitimate zero-gold offer.
type AppraisalOffers struct {
	Gems         uint16
	GemsReady    bool
	Jewelry      uint16
	JewelryReady bool
}

func (s *State) SetAppraisalOffers(offers AppraisalOffers) {
	s.appraisalOffers = offers
}

// SetShopOffers replaces the current city shop stock with a defensive copy.
// It is intentionally separate from the UI so each city/ECL script can load
// its own stock without changing the Shop Menu state machine.
func (s *State) SetShopOffers(offers []ShopOffer) {
	s.shopOffers = append([]ShopOffer(nil), offers...)
}

func (s *State) ShopOffers() []ShopOffer {
	return append([]ShopOffer(nil), s.shopOffers...)
}

func (s *State) MoneyPool() uint32 { return s.moneyPool }

// PoolPartyGold implements the manual's POOL command. All five typed coin
// fields contribute their reference gold worth before being cleared.
func (s *State) PoolPartyGold() error {
	if len(s.partyRoster) == 0 {
		return fmt.Errorf("cannot pool gold without a party")
	}
	for index := range s.partyRoster {
		s.moneyPool += characterCoinGoldWorth(s.partyRoster[index])
		s.partyRoster[index].Copper = 0
		s.partyRoster[index].Silver = 0
		s.partyRoster[index].Electrum = 0
		s.partyRoster[index].Gold = 0
		s.partyRoster[index].Platinum = 0
	}
	return nil
}

// TakeGold implements a bounded TAKE operation for one character. The UI can
// later supply the selected character and requested amount.
func (s *State) TakeGold(characterIndex int, amount uint32) error {
	if characterIndex < 0 || characterIndex >= len(s.partyRoster) {
		return fmt.Errorf("character index %d is out of range", characterIndex)
	}
	if amount > s.moneyPool {
		return fmt.Errorf("money pool has %d gold; requested %d", s.moneyPool, amount)
	}
	if uint32(s.partyRoster[characterIndex].Gold)+amount > 0xFFFF {
		return fmt.Errorf("character gold would exceed 65535")
	}
	s.moneyPool -= amount
	s.partyRoster[characterIndex].Gold += uint16(amount)
	return nil
}

// ShareGold distributes the pool as evenly as possible in marching order.
// Any remainder is assigned one gold at a time from the first character.
func (s *State) ShareGold() error {
	if len(s.partyRoster) == 0 {
		return fmt.Errorf("cannot share gold without a party")
	}
	share := s.moneyPool / uint32(len(s.partyRoster))
	remainder := s.moneyPool % uint32(len(s.partyRoster))
	for index := range s.partyRoster {
		amount := share
		if uint32(index) < remainder {
			amount++
		}
		if uint32(s.partyRoster[index].Gold)+amount > 0xFFFF {
			return fmt.Errorf("character %d gold would exceed 65535", index)
		}
	}
	for index := range s.partyRoster {
		amount := share
		if uint32(index) < remainder {
			amount++
		}
		s.partyRoster[index].Gold += uint16(amount)
	}
	s.moneyPool = 0
	return nil
}

// BuyShopOffer follows CityShop's payment order: the selected character's
// five coin fields are used first by gold worth, then the pooled GP fallback.
func (s *State) BuyShopOffer(characterIndex, offerIndex int) error {
	if characterIndex < 0 || characterIndex >= len(s.partyRoster) {
		return fmt.Errorf("character index %d is out of range", characterIndex)
	}
	if offerIndex < 0 || offerIndex >= len(s.shopOffers) {
		return fmt.Errorf("shop offer index %d is out of range", offerIndex)
	}
	offer := s.shopOffers[offerIndex]
	character := &s.partyRoster[characterIndex]
	if characterCoinGoldWorth(*character) >= uint32(offer.Price) {
		subtractCharacterGoldWorth(character, uint32(offer.Price))
	} else if uint32(offer.Price) <= s.moneyPool {
		s.moneyPool -= uint32(offer.Price)
	} else {
		return fmt.Errorf("character and money pool cannot pay %d gold", offer.Price)
	}
	item := offer.Item
	item.Readied = false
	character.Equipment = append(character.Equipment, item)
	if characterIndex < len(s.party) {
		fighter, err := s.fighterForCharacter(s.partyRoster[characterIndex])
		if err != nil {
			return err
		}
		s.party[characterIndex] = fighter
	}
	return nil
}

func characterCoinGoldWorth(character party.Character) uint32 {
	copper := uint64(character.Copper) +
		uint64(character.Silver)*10 +
		uint64(character.Electrum)*100 +
		uint64(character.Gold)*200 +
		uint64(character.Platinum)*1000
	return uint32(copper / 200)
}

func subtractCharacterGoldWorth(character *party.Character, gold uint32) {
	values := [5]int{
		int(character.Copper), int(character.Silver), int(character.Electrum),
		int(character.Gold), int(character.Platinum),
	}
	perCopper := [5]int{1, 10, 100, 200, 1000}
	coppers := int(gold) * 200
	for coin := 0; coppers > 0 && coin < len(values); coin++ {
		subCoins := coppers/perCopper[coin] + 1
		if values[coin] < subCoins {
			subCoins = values[coin]
		}
		coppers -= perCopper[coin] * subCoins
		values[coin] -= subCoins
	}
	if coppers < 0 {
		coppers = -coppers
		for coin := len(values) - 1; coppers > 0 && coin >= 0; coin-- {
			addCoins := coppers / perCopper[coin]
			coppers -= perCopper[coin] * addCoins
			values[coin] += addCoins
		}
	}
	character.Copper = uint16(values[0])
	character.Silver = uint16(values[1])
	character.Electrum = uint16(values[2])
	character.Gold = uint16(values[3])
	character.Platinum = uint16(values[4])
}

// SellShopItem transfers an item's documented raw Value to the party pool
// and removes it from the selected character. Readied and cursed items stay
// protected by the same inventory rules used by the DOS adapter.
func (s *State) SellShopItem(characterIndex, itemIndex int) error {
	if characterIndex < 0 || characterIndex >= len(s.partyRoster) {
		return fmt.Errorf("character index %d is out of range", characterIndex)
	}
	character := &s.partyRoster[characterIndex]
	if itemIndex < 0 || itemIndex >= len(character.Equipment) {
		return fmt.Errorf("item index %d is out of range", itemIndex)
	}
	item := character.Equipment[itemIndex]
	if item.Readied {
		return fmt.Errorf("cannot sell a readied item")
	}
	if item.Cursed {
		return fmt.Errorf("cannot sell a cursed item")
	}
	if item.Value <= 0 {
		return fmt.Errorf("item has no positive documented sale value")
	}
	s.moneyPool += uint32(item.Value)
	character.Equipment = append(character.Equipment[:itemIndex], character.Equipment[itemIndex+1:]...)
	if characterIndex < len(s.party) {
		fighter, err := s.fighterForCharacter(*character)
		if err != nil {
			return err
		}
		s.party[characterIndex] = fighter
	}
	return nil
}

// IdentifyShopItem charges the manual-confirmed 200 GP fee. The raw item is
// returned for the UI, while magical name/effect decoding remains explicit
// future data work rather than being fabricated from the type byte.
func (s *State) IdentifyShopItem(characterIndex, itemIndex int) (monster.ItemRecord, error) {
	if characterIndex < 0 || characterIndex >= len(s.partyRoster) {
		return monster.ItemRecord{}, fmt.Errorf("character index %d is out of range", characterIndex)
	}
	character := &s.partyRoster[characterIndex]
	if itemIndex < 0 || itemIndex >= len(character.Equipment) {
		return monster.ItemRecord{}, fmt.Errorf("item index %d is out of range", itemIndex)
	}
	item := character.Equipment[itemIndex]
	if err := character.PayIdentifyFee(); err != nil {
		return monster.ItemRecord{}, err
	}
	return item, nil
}

// AppraiseTreasure accepts an injected shopkeeper offer and transfers the
// selected character's entire gem/jewelry holding into the party pool.
func (s *State) AppraiseTreasure(characterIndex int, kind TreasureKind) (uint16, error) {
	if characterIndex < 0 || characterIndex >= len(s.partyRoster) {
		return 0, fmt.Errorf("character index %d is out of range", characterIndex)
	}
	character := &s.partyRoster[characterIndex]
	var amount, offer uint16
	var ready bool
	switch kind {
	case TreasureGems:
		amount, offer, ready = character.Gems, s.appraisalOffers.Gems, s.appraisalOffers.GemsReady
	case TreasureJewelry:
		amount, offer, ready = character.Jewelry, s.appraisalOffers.Jewelry, s.appraisalOffers.JewelryReady
	default:
		return 0, fmt.Errorf("unknown appraisal treasure kind %d", kind)
	}
	if amount == 0 {
		return 0, fmt.Errorf("character has no treasure kind %d", kind)
	}
	if !ready {
		return 0, fmt.Errorf("appraisal offer for treasure kind %d is not loaded", kind)
	}
	s.moneyPool += uint32(offer)
	if kind == TreasureGems {
		character.Gems = 0
	} else {
		character.Jewelry = 0
	}
	return offer, nil
}
