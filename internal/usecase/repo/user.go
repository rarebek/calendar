package repo

import (
	"context"
	sql2 "database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"job_tasks/calendar/internal/entity"
	"job_tasks/calendar/pkg/postgres"
	"time"
)

const (
	usersTableName = "users"
)

// UserRepo -.
type UserRepo struct {
	*postgres.Postgres
}

// NewUserRepo -.
func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (u *UserRepo) usersSelectQueryPrefix() squirrel.SelectBuilder {
	return u.Builder.
		Select(
			"id", "username", "password",
		).From(usersTableName).
		Where(squirrel.Eq{"deleted_at": nil})
}

// CreateUser -.
func (u *UserRepo) CreateUser(ctx context.Context, user *entity.UserRequest) (*entity.UserResponse, error) {
	var (
		userResponse entity.UserResponse
		updatedAt    sql2.NullTime
	)
	data := map[string]interface{}{
		"id":       uuid.NewString(),
		"email":    user.Email,
		"username": user.Username,
		"password": user.Password,
	}
	sql, args, err := u.Builder.
		Insert(usersTableName).
		SetMap(data).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("UserRepo - CreateEvent - r.Builder: %w", err)
	}
	sql += " RETURNING id, email, username, password, created_at, updated_at"

	row := u.Pool.QueryRow(ctx, sql, args...)
	if err := row.Scan(&userResponse.Id, &userResponse.Email, &userResponse.Username, &userResponse.Password, &userResponse.CreatedAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("UserRepo - CreateEvent - row.Scan: %w", err)
	}
	if updatedAt.Valid {
		userResponse.UpdatedAt = updatedAt.Time
	}
	return &userResponse, nil
}

// GetUser - .
func (u *UserRepo) GetUser(ctx context.Context, req *entity.GetRequest) (*entity.UserResponse, error) {
	var (
		user      entity.UserResponse
		updatedAt sql2.NullTime
	)
	sql, args, err := u.Builder.
		Select("id, email, username, password, created_at, updated_at").
		From(usersTableName).
		Where(squirrel.And{
			squirrel.Eq{req.Field: req.Value},
			squirrel.Eq{"deleted_at": nil},
		}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetUser - r.Builder: %w", err)
	}

	row := u.Pool.QueryRow(ctx, sql, args...)

	if err := row.Scan(&user.Id, &user.Email, &user.Username, &user.Password, &user.CreatedAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("UserRepo - GetUser - row.Scan: %w", err)
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}
	return &user, nil
}

// UpdateUser - .
func (u *UserRepo) UpdateUser(ctx context.Context, user *entity.UpdateUserRequest) (*entity.UserResponse, error) {
	var (
		response  entity.UserResponse
		updatedAt sql2.NullTime
	)
	fmt.Println(user)
	data := map[string]interface{}{
		"email":      user.Email,
		"username":   user.Username,
		"password":   user.Password,
		"updated_at": time.Now(),
	}
	sql, args, err := u.Builder.
		Update(usersTableName).
		SetMap(data).
		Where(squirrel.And{
			squirrel.Eq{"id": user.Id},
			squirrel.Eq{"deleted_at": nil},
		},
		).ToSql()
	sql += " RETURNING id,email, username, password, created_at, updated_at"
	if err != nil {
		return nil, fmt.Errorf("UserRepo - UpdateUser - r.Builder: %w", err)
	}

	row := u.Pool.QueryRow(ctx, sql, args...)
	if err := row.Scan(&response.Id, &response.Email, &response.Username, &response.Password, &response.CreatedAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("UserRepo - UpdateUser - row.Scan: %w", err)
	}
	if updatedAt.Valid {
		response.UpdatedAt = updatedAt.Time
	}
	return &response, nil
}

// ListUsers - .
func (u *UserRepo) ListUsers(ctx context.Context, req *entity.GetListRequest) (*entity.Users, error) {
	var (
		users entity.Users
	)
	builder := u.Builder.Select(
		"id", "email", "username", "password", "created_at", "updated_at").
		From(usersTableName).Where(squirrel.Eq{"deleted_at": nil})
	if req.Field != "" && req.Value != "" {
		builder = builder.Where(squirrel.And{
			squirrel.ILike{req.Field: req.Value + "%"},
		})
	}
	if req.OrderBy != "" {
		builder = builder.OrderBy(req.OrderBy)
	}

	offset := (req.Page - 1) * req.Limit

	builder = builder.Limit(uint64(req.Limit)).Offset(uint64(offset))

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ListUsers - u.usersSelectQueryPrefix: %w", err)
	}
	fmt.Println(sql)

	rows, err := u.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListUsers  - u.Pool.Query: %w", err)
	}

	for rows.Next() {
		var (
			user      entity.UserResponse
			updatedAt sql2.NullTime
		)

		if err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.Username,
			&user.Password,
			&user.CreatedAt,
			&updatedAt); err != nil {
			return nil, fmt.Errorf("UserRepo - ListUsers - u.Row.Scan: %w", err)
		}

		if updatedAt.Valid {
			user.UpdatedAt = updatedAt.Time
		}

		users.Count++
		users.Users = append(users.Users, &user)
	}

	return &users, nil
}

// DeleteUser - .
func (u *UserRepo) DeleteUser(ctx context.Context, req *entity.GetRequest) error {
	sql, args, err := u.Builder.Update(usersTableName).
		Set("deleted_at", time.Now()).
		Where(squirrel.And{
			squirrel.Eq{"deleted_at": nil},
			squirrel.Eq{req.Field: req.Value},
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("DeleteUser - u.Builder: %w", err)
	}
	fmt.Println(sql)
	fmt.Println(req.Value)

	if _, err := u.Pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("DeleteUser - u.Pool.Exec: %w", err)
	}

	return nil
}

// CheckUniqueness -.
func (u *UserRepo) CheckUniqueness(ctx context.Context, req *entity.GetRequest) (bool, error) {
	var count int
	query, args, err := u.Builder.
		Select("COUNT(*)").
		From(usersTableName).
		Where(squirrel.And{
			squirrel.Eq{req.Field: req.Value},
			squirrel.Eq{"deleted_at": nil},
		}).ToSql()
	if err != nil {
		return false, fmt.Errorf("UserRepo - CheckUniqueness - u.Builder: %w", err)
	}

	err = u.Pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("UserRepo - CheckUniqueness - u.Pool.QueryRow: %w", err)
	}

	return count == 0, nil
}
