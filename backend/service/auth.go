package service

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"

	"golang.org/x/crypto/bcrypt"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	ginjwt "github.com/appleboy/gin-jwt/v2"
	sq "github.com/elgris/sqrl"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// AuthenticatorService Service
type AuthenticatorService struct {
	DB *sqlx.DB
}

// Login anthenticator
func (u *AuthenticatorService) Run(form models.LoginForm) (a *models.Account, err error) {

	usr, err := getUserFromEmail(u.DB, form.Email)
	if err != nil {
		return nil, errors.Wrap(err, "user from email")
	}

	err = bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(form.Password))
	if err != nil {
		return nil, ginjwt.ErrFailedAuthentication
	}

	return &models.Account{
		AccoID:    usr.AccoID,
		Name:      usr.Name,
		CreatedAt: usr.CreatedAt,
		DeletedAt: usr.DeletedAt,
	}, nil
}

type authenticationUser struct {
	AccoID    uuid.UUID `db:"acco_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	DeletedAt null.Time `db:"deleted_at,omitempty"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
}

// getUserFromEmail return user from key
func getUserFromEmail(db *sqlx.DB, email string) (*authenticationUser, error) {

	query := psql.Select(
		"acc.acco_id",
		`acc."name"`,
		"acc.created_at",
		"acc.deleted_at",
		"ae.email",
		"ae.password").
		From("account acc").
		Join("authentication_email ae using(acco_id)").
		Where(sq.Eq{"acc.deleted_at": nil}).
		Where("lower(trim(ae.email)) = lower(trim(?))", email)

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	resp := &authenticationUser{}
	err = db.Get(resp, qSQL, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no user found with the provided email")
		}
		return nil, err
	}
	return resp, err
}
