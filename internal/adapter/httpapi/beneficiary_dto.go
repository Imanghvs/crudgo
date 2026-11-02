package httpapi

import (
	"time"

	"github.com/google/uuid"
)

type CreateBeneficiaryRequest struct {
	MerchantID uuid.UUID `json:"merchant_id"`
	Name       string    `json:"name"`
	IBAN       string    `json:"iban"`
}

type BeneficiaryResponse struct {
	ID         uuid.UUID `json:"id"`
	MerchantID uuid.UUID `json:"merchant_id"`
	Name       string    `json:"name"`
	IBAN       string    `json:"iban"`
	CreatedAt  time.Time `json:"created_at"`
}
