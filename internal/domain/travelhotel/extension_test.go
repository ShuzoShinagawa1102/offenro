package travelhotel

import (
	"testing"
	"time"

	modelapi "github.com/ShuzoShinagawa1102/offenro/internal/domain/travelhotel/generated/model"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestConditionValidate(t *testing.T) {
	t.Parallel()

	checkIn := time.Date(2026, time.September, 10, 0, 0, 0, 0, time.UTC)
	checkOut := checkIn.AddDate(0, 0, 2)
	condition := &Condition{Request: modelapi.TravelHotelSearchRequest{
		Destination: modelapi.TravelHotelDestination{PrefectureCode: "14"},
		Stay: modelapi.TravelHotelStay{
			CheckIn:  openapi_types.Date{Time: checkIn},
			CheckOut: openapi_types.Date{Time: checkOut},
		},
		Guests: modelapi.TravelHotelGuests{Adults: 2, Rooms: 1},
	}}

	if err := condition.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConditionValidateRejectsInvalidStay(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.September, 10, 0, 0, 0, 0, time.UTC)
	condition := &Condition{Request: modelapi.TravelHotelSearchRequest{
		Destination: modelapi.TravelHotelDestination{PrefectureCode: "14"},
		Stay: modelapi.TravelHotelStay{
			CheckIn:  openapi_types.Date{Time: date},
			CheckOut: openapi_types.Date{Time: date},
		},
		Guests: modelapi.TravelHotelGuests{Adults: 2, Rooms: 1},
	}}

	if err := condition.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want invalid stay error")
	}
}
