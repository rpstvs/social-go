package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrDuplicateUsername = errors.New("username already exists")
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  Password `json:"-"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	IsActive  bool     `json:"is_active"`
}

type Password struct {
	Text *string
	Hash []byte
}

func (p *Password) Set(text string) error {

	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)

	if err != nil {
		return err
	}
	p.Text = &text
	p.Hash = hash
	return nil
}

func (p *Password) Validate(password string) error {
	return bcrypt.CompareHashAndPassword(p.Hash, []byte(password))

}

type UsersStore struct {
	db *sql.DB
}

func (s *UsersStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
	INSERT INTO users (username, password, email)
	VALUES($1,$2,$3)
	RETURNING id, created_at`

	err := tx.QueryRowContext(ctx, query, user.Username, user.Password, user.Email).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err

		}

	}
	return nil
}

func (s *UsersStore) GetById(ctx context.Context, id int64) (*User, error) {
	var user User

	query := `
	SELECT id, username, email, password, created_at
	from users
	WHERE id = $1 AND is_active = true;`

	err := s.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.Password.hash, &user.CreatedAt)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			{
				return nil, ErrNotFound
			}
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (s *UsersStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	query := `
	SELECT id, username, email, password, created_at
	from users
	WHERE email = $1 AND is_active = true;`

	err := s.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			{
				return nil, ErrNotFound
			}
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (u *UsersStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {

		if err := u.Create(ctx, tx, user); err != nil {
			return err
		}

		if err := u.createUserInvitation(ctx, tx, token, exp, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (u *UsersStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, invitationExp time.Duration, userId int64) error {
	query := `
	INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1,$2,$3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)

	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userId, time.Now().Add(invitationExp))

	if err != nil {
		return err
	}

	return nil
}

func (u *UsersStore) Activate(ctx context.Context, token string) error {

	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		user, err := u.getUserFromInvitation(ctx, tx, token)

		if err != nil {
			return err
		}

		user.IsActive = true

		if err := u.update(ctx, tx, user); err != nil {
			return err
		}

		if err := u.deleteUserInvitation(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (u *UsersStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
	SELECT u.id, u.username, u.email, u.created_at, u.is_active
	FROM users u
	JOIN user_invitations ui ON u.id = ui.user_id
	WHERE ui.token =$1 AND ui.expiry > $2
	`

	hashToken := sha256.Sum256([]byte(token))

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(user.ID, user.Username, user.Email, user.CreatedAt, user.isActive)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (u *UsersStore) update(ctx context.Context, tx *sql.Tx, user *User) error {

	query := `
		UPDATE users
		SET username = $1, email = $2, is_active = $3 
		WHERE user_id = $4
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)

	if err != nil {
		return err
	}

	return nil
}

func (u *UsersStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userId int64) error {

	query := `
		DELETE from user_invitations
		WHERE user_id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, userId)

	if err != nil {
		return err
	}

	return nil
}
