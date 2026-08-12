package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// モック用のMerchantサーバ

type SearchResponse struct {
	Offers []Offer `json:"offers"`
}

type Offer struct {
	OfferID  string `json:"offer_id"`
	Title    string `json:"title"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /shoes/search", func(w http.ResponseWriter, r *http.Request) {
		response := SearchResponse{
			Offers: []Offer{
				{
					OfferID:  "merchant-offer-001",
					Title:    "Sample Sneaker",
					Amount:   9800,
					Currency: "JPY",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Println("merchant mock started on :8081")

	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal(err)
	}
}
