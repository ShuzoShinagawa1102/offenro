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

type SearchCondition interface {
	Domain() Domain
	Validate() error
}

type TravelHotelDestination struct {
	PrefectureCode string
}

type TravelHotelStay struct {
	CheckIn  time.Time
	CheckOut time.Time
}

type TravelHotelGuests struct {
	Adults int
	Rooms  int
}

type TravelHotelFilters struct {
	MaxPrice *int64
}

type TravelHotelCondition struct {
	Destination TravelHotelDestination
	Stay        TravelHotelStay
	Guests      TravelHotelGuests
	Filters     TravelHotelFilters
}

func (c TravelHotelCondition) Domain() Domain {
	return DomainTravelHotel
}

func (c TravelHotelCondition) Validate() error {
	if c.Destination.PrefectureCode == "" {
		return errors.New("prefecture code is required")
	}

	if c.Stay.CheckIn.IsZero() {
		return errors.New("check-in is required")
	}

	if c.Stay.CheckOut.IsZero() {
		return errors.New("check-out is required")
	}

	if !c.Stay.CheckOut.After(c.Stay.CheckIn) {
		return errors.New("check-out must be after check-in")
	}

	if c.Guests.Adults <= 0 {
		return errors.New("adults must be greater than zero")
	}

	if c.Guests.Rooms <= 0 {
		return errors.New("rooms must be greater than zero")
	}

	if c.Filters.MaxPrice != nil &&
		*c.Filters.MaxPrice < 0 {
		return errors.New("max price must not be negative")
	}

	return nil
}

type Offer interface {
	GetDomain() Domain
}

type OfferBase struct {
	ID         string
	MerchantID string
	Domain     Domain
	Amount     int64
	Currency   string
}

type TravelHotel struct {
	HotelID        string
	Name           string
	PrefectureCode string
	PrefectureName string
	City           string
}

type TravelHotelOffer struct {
	Base  OfferBase
	Hotel TravelHotel
	Stay  TravelHotelStay
}

func (o TravelHotelOffer) GetDomain() Domain {
	return o.Base.Domain
}

type MerchantTarget struct {
	MerchantID string
	Domain     Domain
	BaseURL    string
}

type MerchantRegistry interface {
	FindByDomain(
		ctx context.Context,
		domain Domain,
	) ([]MerchantTarget, error)
}

type MerchantDiscovery interface {
	FindMerchants(
		ctx context.Context,
		condition SearchCondition,
		candidates []MerchantTarget,
		limit int,
	) ([]MerchantTarget, error)
}

type DomainSearcher interface {
	Search(
		ctx context.Context,
		merchant MerchantTarget,
		condition SearchCondition,
	) ([]Offer, error)
}
