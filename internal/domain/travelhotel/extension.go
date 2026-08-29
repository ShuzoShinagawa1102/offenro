package travelhotel

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/discovery"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/extension"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/fulfillment"
	coremodel "github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/offer"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/search"
	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel/generated/model"
)

const Domain coremodel.Domain = "travel.hotel"

const dimensionPrefectureCode = "prefecture_code"

// Condition only adapts the generated OpenAPI model to the Core interface.
// Domain fields remain defined exclusively in schemas.yaml and generated/model.
type Condition struct {
	Request modelapi.TravelHotelSearchRequest
}

func (c *Condition) Domain() coremodel.Domain {
	return Domain
}

func (c *Condition) Validate() error {
	if c.Request.Destination.PrefectureCode == "" {
		return errors.New("prefecture code is required")
	}
	if c.Request.Stay.CheckIn.Time.IsZero() {
		return errors.New("check-in is required")
	}
	if c.Request.Stay.CheckOut.Time.IsZero() {
		return errors.New("check-out is required")
	}
	if !c.Request.Stay.CheckOut.Time.After(c.Request.Stay.CheckIn.Time) {
		return errors.New("check-out must be after check-in")
	}
	if c.Request.Guests.Adults <= 0 {
		return errors.New("adults must be greater than zero")
	}
	if c.Request.Guests.Rooms <= 0 {
		return errors.New("rooms must be greater than zero")
	}
	if c.Request.Filters != nil &&
		c.Request.Filters.MaxPrice != nil &&
		*c.Request.Filters.MaxPrice < 0 {
		return errors.New("max price must not be negative")
	}

	return nil
}

// Offer only adapts the generated OpenAPI model to the Core interface.
type Offer struct {
	Value modelapi.TravelHotelOffer
}

func (o *Offer) Domain() coremodel.Domain {
	return Domain
}

func asCondition(condition search.SearchCondition) (*Condition, error) {
	hotelCondition, ok := condition.(*Condition)
	if !ok {
		return nil, fmt.Errorf("invalid condition type for domain %s", Domain)
	}
	return hotelCondition, nil
}

type Extension struct {
	searcher     *Searcher
	discovery    *MerchantDiscovery
	indexBuilder *IndexBuilder
	fulfiller    *Fulfiller
}

var _ extension.Extension = (*Extension)(nil)

func New(
	repository discovery.IndexRepository,
	httpClient *http.Client,
	offerIssuer offer.Issuer,
) *Extension {
	return &Extension{
		searcher:     NewSearcher(httpClient, offerIssuer),
		discovery:    NewMerchantDiscovery(repository),
		indexBuilder: NewIndexBuilder(httpClient),
		fulfiller:    NewFulfiller(httpClient),
	}
}

func (e *Extension) Domain() coremodel.Domain {
	return Domain
}

func (e *Extension) Searcher() search.DomainSearcher {
	return e.searcher
}

func (e *Extension) MerchantDiscovery() search.MerchantDiscovery {
	return e.discovery
}

func (e *Extension) IndexBuilder() discovery.DomainIndexBuilder {
	return e.indexBuilder
}

func (e *Extension) OfferVerifier() offer.Verifier {
	return e.searcher
}

func (e *Extension) Fulfiller() fulfillment.DomainFulfiller {
	return e.fulfiller
}
