package cart

import (
	"encoding/json"
	"fmt"
	"strings"
)

func validateNewItem(input AddItemInput) error {
	if strings.TrimSpace(input.OfferID) == "" {
		return fmt.Errorf("%w: offer_id is required", ErrInvalidInput)
	}
	if !validSnapshot(input.OfferSnapshot) {
		return fmt.Errorf("%w: offer_snapshot must be a JSON object", ErrInvalidInput)
	}
	if input.Amount < 0 {
		return fmt.Errorf("%w: amount must not be negative", ErrInvalidInput)
	}
	if !validCurrency(input.Currency) {
		return fmt.Errorf("%w: currency must be a three-letter code", ErrInvalidInput)
	}
	return nil
}

func validSnapshot(snapshot json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(snapshot))
	return len(trimmed) > 1 && trimmed[0] == '{' && json.Valid(snapshot)
}

func validCurrency(currency string) bool {
	normalized := normalizeCurrency(currency)
	if len(normalized) != 3 {
		return false
	}
	for _, character := range normalized {
		if character < 'A' || character > 'Z' {
			return false
		}
	}
	return true
}

func normalizeCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

func cloneJSON(value json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), value...)
}
