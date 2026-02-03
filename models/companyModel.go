package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID                 uuid.UUID `db:"id"`
	Name               string    `db:"name"`
	Slug               string    `db:"slug"`
	Industry           string    `db:"industry"`
	Country            string    `db:"country"`
	Timezone           string    `db:"timezone"`
	Currency           string    `db:"currency"`
	RegistrationNumber string    `db:"registration_number"`
	TaxID              string    `db:"tax_id"`
	Address            string    `db:"address"`
	Phone              string    `db:"phone"`
	LogoURL            string    `db:"logo_url"`
	Status             string    `db:"status"`
	Settings           []byte    `db:"settings"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}
