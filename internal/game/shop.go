package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// ShopOffer is a script/data-layer shop stock entry. ITEMS descriptors do not
// contain prices, so a caller must provide the offer price explicitly.
type ShopOffer struct {
	Item  monster.ItemRecord
	Price uint16
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

// PoolPartyGold implements the manual's POOL command. Gold is moved out of
// each roster character and into one party pool; calling it again combines
// any newly acquired per-character gold into the same pool.
func (s *State) PoolPartyGold() error {
	if len(s.partyRoster) == 0 {
		return fmt.Errorf("cannot pool gold without a party")
	}
	for index := range s.partyRoster {
		s.moneyPool += uint32(s.partyRoster[index].Gold)
		s.partyRoster[index].Gold = 0
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

// BuyShopOffer spends the party pool on one injected offer and refreshes the
// corresponding combat fighter projection when that character is loaded.
func (s *State) BuyShopOffer(characterIndex, offerIndex int) error {
	if characterIndex < 0 || characterIndex >= len(s.partyRoster) {
		return fmt.Errorf("character index %d is out of range", characterIndex)
	}
	if offerIndex < 0 || offerIndex >= len(s.shopOffers) {
		return fmt.Errorf("shop offer index %d is out of range", offerIndex)
	}
	offer := s.shopOffers[offerIndex]
	if uint32(offer.Price) > s.moneyPool {
		return fmt.Errorf("money pool has %d gold; shop price is %d", s.moneyPool, offer.Price)
	}
	item := offer.Item
	item.Readied = false
	s.partyRoster[characterIndex].Equipment = append(s.partyRoster[characterIndex].Equipment, item)
	s.moneyPool -= uint32(offer.Price)
	if characterIndex < len(s.party) {
		fighter, err := s.fighterForCharacter(s.partyRoster[characterIndex])
		if err != nil {
			return err
		}
		s.party[characterIndex] = fighter
	}
	return nil
}
