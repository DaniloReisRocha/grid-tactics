package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gopkg.in/guregu/null.v3"
)

type Account struct {
	AccoID    uuid.UUID `db:"acco_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	DeletedAt null.Time `db:"deleted_at,omitempty"`
}

type AuthenticationEmail struct {
	AuemID    int64     `db:"auem_id"`
	AccoID    uuid.UUID `db:"acco_id"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
}
