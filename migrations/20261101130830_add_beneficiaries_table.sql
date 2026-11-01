-- +goose Up
CREATE TABLE beneficiaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL,
    name VARCHAR NOT NULL,
    iban VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT beneficiaries_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT beneficiaries_iban_not_blank CHECK (btrim(iban) <> '')
);

-- +goose Down
DROP TABLE beneficiaries;
