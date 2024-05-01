package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/mike-kimani/whitepointinventory/internal/database"
)


func (apiCfg * apiConfig) handlerCreatePayment(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct{
		Cash int32 `json:"cash_paid"`
		PricePerChicken int32 `json:"price_per_chicken_paid"`
		FarmerID uuid.UUID `json:"farmer_id"`
	}
	params := parameters{}

	cash_balance :=  sql.NullInt32{}

	chicken_balance := sql.NullInt32{}

	decode := json.NewDecoder(r.Body)

	err := decode.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't decode json data into params: %v", err))
		return
	}

	payment, err := apiCfg.DB.CreatePayment(r.Context(), database.CreatePaymentParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CashPaid: params.Cash,
		PricePerChickenPaid: params.PricePerChicken,
		UserID: user.ID,
		FarmerID: params.FarmerID,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't create purchase: %v", err))
		return
	}

	cash_balance.Int32 = params.Cash
	cash_balance.Valid = true
	// TODO: Handle operations that result in floats
	chicken_balance.Int32 = params.Cash/params.PricePerChicken
	chicken_balance.Valid = true

	err = apiCfg.DB.DecreaseCashOwed(r.Context(), database.DecreaseCashOwedParams{
		CashBalance: cash_balance,
		ID: params.FarmerID,
	})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't decrease cash owed to the farmer: %v", err))
	}

	err = apiCfg.DB.DecreaseChickenOwed(r.Context(), database.DecreaseChickenOwedParams{
		ChickenBalance: chicken_balance,
		ID: params.FarmerID,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't decrease chicken owed to the farmer: %v", err))
	}

	respondWithJSON(w, 201, payment)
}

func (apiCfg *apiConfig) handlerGetPayments(w http.ResponseWriter, r *http.Request) {
	payments, err := apiCfg.DB.GetPayments(r.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't fetch payments: %v", err))
	}

	respondWithJSON(w, 200, payments)
}

func (apiCfg *apiConfig) handlerGetPaymentByID(w http.ResponseWriter, r * http.Request) {
	type parameters struct{
		ID uuid.UUID `json:"payment_id"`
	}
	params := parameters{}

	decode := json.NewDecoder(r.Body)
	err := decode.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't decode json: %v", err))
		return
	}

	payment, err := apiCfg.DB.GetPaymentByID(r.Context(), params.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get payment: %v", err))
		return
	}

	respondWithJSON(w, 200, payment)
}

func (apiCfg *apiConfig) handlerDeletePayment(w http.ResponseWriter, r *http.Request, user database.User) {
	paymentIDStr := chi.URLParam(r, "payment_id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't parse payment UUID: %v", err))
	}

	cash_balance := sql.NullInt32{}
	chicken_owed := sql.NullInt32{}

	payment, err := apiCfg.DB.GetPaymentByID(r.Context(), paymentID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get payment: %v", err))
		return
	}
	cash_balance.Int32 = payment.CashPaid
	cash_balance.Valid = true
	// TODO: Handle operations that result in floats
	chicken_owed.Int32 = payment.CashPaid/payment.PricePerChickenPaid
	chicken_owed.Valid = true

	err = apiCfg.DB.IncreaseCashOwed(r.Context(), database.IncreaseCashOwedParams{
		CashBalance: cash_balance,
		ID: payment.FarmerID,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't increase cash owned: %v", err))
		return
	}

	err = apiCfg.DB.IncreaseChickenOwed(r.Context(), database.IncreaseChickenOwedParams{
		ChickenBalance: chicken_owed,
		ID: payment.FarmerID,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't increase chicken owed: %v", err))
		return
	}
	
	err =  apiCfg.DB.DeletePayments(r.Context(), payment.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't delete payment: %v", err))
		return
	}

	respondWithJSON(w, 200, fmt.Sprintf("Deletion successfully done by %v", user.Name))

}