package beneficiary

import (
	"time"

	"github.com/google/uuid"
)

type Beneficiary struct {
	ID         uuid.UUID
	MerchantID uuid.UUID
	Name       string
	IBAN       string
	CreatedAt  time.Time
}

type CreateInput struct {
	MerchantID uuid.UUID
	Name       string
	IBAN       string
}
