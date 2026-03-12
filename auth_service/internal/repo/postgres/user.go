package postgres

import (
	"auth_service/internal/domain"
	"context"
	"errors"

	uuid "github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, email, username, passwordHash string) (*domain.User, error) {
	const q = `INSERT INTO users (id, email, username, password_hash) VALUES ($1, $2, $3, $4)
		RETURNING id, email, username, password_hash, created_at, updated_at`
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	var u domain.User
	err = r.pool.QueryRow(ctx, q, id.String(), email, username, passwordHash).Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) ByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `SELECT id, email, username, password_hash, created_at, updated_at FROM users WHERE id = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `SELECT id, email, username, password_hash, created_at, updated_at FROM users WHERE email = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) ByUsername(ctx context.Context, username string) (*domain.User, error) {
	const q = `SELECT id, email, username, password_hash, created_at, updated_at FROM users WHERE username = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, username).Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
