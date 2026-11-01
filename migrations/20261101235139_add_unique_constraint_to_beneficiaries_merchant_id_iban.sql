-- +goose Up
ALTER TABLE beneficiaries
ADD CONSTRAINT beneficiaries_merchant_iban_unique
UNIQUE (merchant_id, iban);

-- +goose Down
ALTER TABLE beneficiaries
DROP CONSTRAINT beneficiaries_merchant_iban_unique;
