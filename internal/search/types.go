package search

import (
	"context"
	"errors"
	"time"
)

type Domain string

const (
	DomainTravelHotel Domain = "travel.hotel"
)

// SearchConditionは、各Domainの検索条件が満たす共通interface。
type SearchCondition interface {
	Domain() Domain
	Validate() error
}

// TravelHotelConditionはtravel.hotel専用の検索条件。
type TravelHotelCondition struct {
	Location string
	CheckIn  time.Time
	CheckOut time.Time
	Adults   int
	Rooms    int
	MaxPrice *int64
}

func (c TravelHotelCondition) Domain() Domain {
	return DomainTravelHotel
}

func (c TravelHotelCondition) Validate() error {
	if c.Location == "" {
		return errors.New("location is required")
	}

	if c.CheckIn.IsZero() {
		return errors.New("check-in is required")
	}

	if c.CheckOut.IsZero() {
		return errors.New("check-out is required")
	}

	if !c.CheckOut.After(c.CheckIn) {
		return errors.New("check-out must be after check-in")
	}

	if c.Adults <= 0 {
		return errors.New("adults must be greater than zero")
	}

	if c.Rooms <= 0 {
		return errors.New("rooms must be greater than zero")
	}

	if c.MaxPrice != nil && *c.MaxPrice < 0 {
		return errors.New("max price must not be negative")
	}

	return nil
}

// OfferはOffenro内部で扱う共通Offer。
type Offer struct {
	ID         string
	MerchantID string
	Domain     Domain
	Title      string
	Amount     int64
	Currency   string
}

// MerchantTargetは、Offer検索先となるMerchant Capability。
// TODO: 将来的にはDBの Merchant + MerchantCapability から生成する。
type MerchantTarget struct {
	MerchantID string
	Domain     Domain
	BaseURL    string
}

// MerchantRegistryは、指定Domainに対応するMerchantを探す役割。
type MerchantRegistry interface {
	FindByDomain(
		ctx context.Context,
		domain Domain,
	) ([]MerchantTarget, error)
}

// DomainSearcherは、Domain固有のMerchant APIを呼び出す役割。
type DomainSearcher interface {
	Search(
		ctx context.Context,
		merchant MerchantTarget,
		condition SearchCondition,
	) ([]Offer, error)
}
