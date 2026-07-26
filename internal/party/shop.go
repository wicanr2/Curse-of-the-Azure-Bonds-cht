package party

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

const (
	// ShopIdentifyFee is the fixed fee documented by the CoAB manual. The
	// actual identification result remains an item/rules-layer concern.
	ShopIdentifyFee uint16 = 200
	maxGold                = ^uint16(0)
)

// BuyItem applies one shop offer to a character. The price is supplied by a
// decoded shop offer because the standalone ITEMS descriptor has no price
// field. The purchased item starts un-readied and is appended as one item
// instance; Count=0 retains the DOS single-item encoding.
func (c *Character) BuyItem(item monster.ItemRecord, price uint16) error {
	if c == nil {
		return fmt.Errorf("cannot buy item for nil character")
	}
	if uint32(c.Gold) < uint32(price) {
		return fmt.Errorf("character has %d gold; shop price is %d", c.Gold, price)
	}
	item.Readied = false
	c.Gold -= price
	c.Equipment = append(c.Equipment, item)
	return nil
}

// SellItem removes one non-readied item and credits the shop's explicit offer.
// The gold overflow check happens before removal so a rejected transaction
// cannot partially mutate inventory.
func (c *Character) SellItem(index int, offer uint16) (monster.ItemRecord, error) {
	if c == nil {
		return monster.ItemRecord{}, fmt.Errorf("cannot sell item for nil character")
	}
	if uint32(c.Gold)+uint32(offer) > uint32(maxGold) {
		return monster.ItemRecord{}, fmt.Errorf("sale would exceed %d gold", maxGold)
	}
	item, err := c.RemoveItem(index)
	if err != nil {
		return monster.ItemRecord{}, err
	}
	c.Gold += offer
	return item, nil
}

// PayIdentifyFee charges the documented shop ID fee. It intentionally does
// not fabricate an item name; the identification result must come from the
// original item/effect catalog.
func (c *Character) PayIdentifyFee() error {
	if c == nil {
		return fmt.Errorf("cannot identify item for nil character")
	}
	if c.Gold < ShopIdentifyFee {
		return fmt.Errorf("character has %d gold; identification costs %d", c.Gold, ShopIdentifyFee)
	}
	c.Gold -= ShopIdentifyFee
	return nil
}
