package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohdrashid9678/xlink/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type PostgresUserRepository struct{ pool *pgxpool.Pool }

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func scanUser(row pgx.Row) (*models.User, error) {
	user := new(models.User)
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	created, err := scanUser(r.pool.QueryRow(ctx, `INSERT INTO users (id, email, password_hash, name)
        VALUES ($1, $2, $3, $4)
        RETURNING id, email, password_hash, name, created_at, updated_at`,
		user.ID, user.Email, user.PasswordHash, user.Name))
	return created, mapDatabaseError(err)
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := scanUser(r.pool.QueryRow(ctx, `SELECT id, email, password_hash, name, created_at, updated_at
        FROM users WHERE email = $1`, email))
	return user, mapDatabaseError(err)
}
