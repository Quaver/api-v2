package db

type OrderItemCategory int8

const (
	OrderItemCategoryDonator OrderItemCategory = iota
	OrderItemCategoryBadge
)

type OrderItem struct {
	Id                 int               `gorm:"column:id; PRIMARY_KEY" json:"id"`
	Category           OrderItemCategory `gorm:"column:category" json:"category"`
	Name               string            `gorm:"column:name" json:"name"`
	PriceSteam         int               `gorm:"column:price_steam" json:"price_steam"`
	PriceStripe        int               `gorm:"column:price_stripe" json:"price_stripe"`
	MaxQuantityAllowed int               `gorm:"column:max_qty_allowed" json:"max_qty_allowed"`
	DonatorBundleItem  bool              `gorm:"column:donator_bundle_item" json:"donator_bundle_item"`
	InStock            bool              `gorm:"column:in_stock" json:"in_stock"`
	CanGift            bool              `gorm:"column:can_gift" json:"can_gift"`
	Visible            bool              `gorm:"column:visible" json:"visible"`
	BadgeId            *int              `gorm:"column:badge_id" json:"badge_id,omitempty"`
}

func (*OrderItem) TableName() string {
	return "order_items"
}

// GetOrderItemById Retrieves an order item from the database by id
func GetOrderItemById(id int) (*OrderItem, error) {
	var item *OrderItem

	result := SQL.
		Where("id = ?", id).
		First(&item)

	if result.Error != nil {
		return nil, result.Error
	}

	return item, nil
}
