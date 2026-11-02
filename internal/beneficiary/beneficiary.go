package beneficiary

import (
	"errors"
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

var ErrAlreadyExists = errors.New("beneficiary already exists")

var ErrNotFound = errors.New("beneficiary not found")
