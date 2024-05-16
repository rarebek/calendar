package usecase

import (
	"context"
	"fmt"
	"job_tasks/calendar/internal/entity"
)

// UserUseCase -.
type UserUseCase struct {
	repo UserRepo
}

// NewUserUseCase -.
func NewUserUseCase(r UserRepo) *UserUseCase {
	return &UserUseCase{
		repo: r,
	}
}

// CreateUser -.
func (uuc *UserUseCase) CreateUser(ctx context.Context, req *entity.UserRequest) (*entity.UserResponse, error) {
	response, err := uuc.repo.CreateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - CreateEvent - uuc.repo.CreateEvent: %w", err)
	}

	return response, nil
}

// UpdateUser - .
func (uuc *UserUseCase) UpdateUser(ctx context.Context, req *entity.UpdateUserRequest) (*entity.UserResponse, error) {
	response, err := uuc.repo.UpdateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - UpdateUser - uuc.repo.UpdateUser: %w", err)
	}

	return response, nil
}

// GetUser - .
func (uuc *UserUseCase) GetUser(ctx context.Context, req *entity.GetRequest) (*entity.UserResponse, error) {
	response, err := uuc.repo.GetUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - GetUser - uuc.repo.GetUser: %w", err)
	}

	return response, nil
}

// ListUsers - .
func (uuc *UserUseCase) ListUsers(ctx context.Context, req *entity.GetListRequest) (*entity.Users, error) {
	response, err := uuc.repo.ListUsers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - ListUser - uuc.repo.ListUsers: %w", err)
	}

	return response, nil
}

// DeleteUser -.
func (uuc *UserUseCase) DeleteUser(ctx context.Context, req *entity.GetRequest) error {
	return uuc.repo.DeleteUser(ctx, req)
}

func (uuc *UserUseCase) CheckUniqueness(ctx context.Context, request *entity.GetRequest) (bool, error) {
	resp, err := uuc.repo.CheckUniqueness(ctx, request)
	if err != nil {
		return false, fmt.Errorf("UserUseCase - CheckUniqueness: %w", err)
	}

	return resp, nil
}
