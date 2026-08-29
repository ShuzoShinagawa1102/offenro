package retailshoes

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/discovery"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/extension"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/fulfillment"
	coremodel "github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/offer"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/search"
	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/retailshoes/generated/model"
)

const Domain coremodel.Domain = "retail.shoes"

const dimensionBrand = "brand"

// Condition only adapts the generated OpenAPI model to the Core interface.
// Domain fields remain defined exclusively in schemas.yaml and generated/model.
type Condition struct {
	Request modelapi.RetailShoesSearchRequest
}

func (c *Condition) Domain() coremodel.Domain {
	return Domain
}

func (c *Condition) Validate() error {
	if strings.TrimSpace(c.Request.Criteria.Brand) == "" {
		return errors.New("brand is required")
	}
	if strings.TrimSpace(c.Request.Criteria.Size) == "" {
		return errors.New("size is required")
	}
	if c.Request.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
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
	Value modelapi.RetailShoesOffer
}

func (o *Offer) Domain() coremodel.Domain {
	return Domain
}

func asCondition(condition search.SearchCondition) (*Condition, error) {
	shoesCondition, ok := condition.(*Condition)
	if !ok {
		return nil, fmt.Errorf("invalid condition type for domain %s", Domain)
	}
	return shoesCondition, nil
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
