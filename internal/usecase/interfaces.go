// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"job_tasks/calendar/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	// User -.
	User interface {
		CreateUser(context.Context, *entity.UserRequest) (*entity.UserResponse, error)
		UpdateUser(context.Context, *entity.UpdateUserRequest) (*entity.UserResponse, error)
		GetUser(context.Context, *entity.GetRequest) (*entity.UserResponse, error)
		ListUsers(context.Context, *entity.GetListRequest) (*entity.Users, error)
		DeleteUser(context.Context, *entity.GetRequest) error
		CheckUniqueness(context.Context, *entity.GetRequest) (bool, error)
	}

	// UserRepo -.
	UserRepo interface {
		CreateUser(context.Context, *entity.UserRequest) (*entity.UserResponse, error)
		UpdateUser(context.Context, *entity.UpdateUserRequest) (*entity.UserResponse, error)
		GetUser(context.Context, *entity.GetRequest) (*entity.UserResponse, error)
		ListUsers(context.Context, *entity.GetListRequest) (*entity.Users, error)
		DeleteUser(context.Context, *entity.GetRequest) error
		CheckUniqueness(context.Context, *entity.GetRequest) (bool, error)
	}

	// Event -.
	Event interface {
		CreateEvent(context.Context, *entity.EventRequest) (*entity.EventResponse, error)
		UpdateEvent(context.Context, *entity.UpdateEventRequest) (*entity.EventResponse, error)
		GetEvent(context.Context, *entity.GetRequest) (*entity.EventResponse, error)
		ListEvents(context.Context, *entity.GetListRequest) (*entity.Events, error)
		DeleteEvent(context.Context, *entity.GetRequest) error
		AddFileToEvent(context.Context, *entity.AddFileToEventRequest) error
		GetAllFilesByEventId(context.Context, *entity.IdRequest) (*entity.Files, error)
		GetExpiredEventsByUserId(context.Context, *entity.GetRequest) (*entity.Events, error)
	}

	// EventRepo -.
	EventRepo interface {
		CreateEvent(context.Context, *entity.EventRequest) (*entity.EventResponse, error)
		UpdateEvent(context.Context, *entity.UpdateEventRequest) (*entity.EventResponse, error)
		GetEvent(context.Context, *entity.GetRequest) (*entity.EventResponse, error)
		ListEvents(context.Context, *entity.GetListRequest) (*entity.Events, error)
		DeleteEvent(context.Context, *entity.GetRequest) error
		AddFileToEvent(context.Context, *entity.AddFileToEventRequest) error
		GetAllFilesByEventId(context.Context, *entity.IdRequest) (*entity.Files, error)
		GetExpiredEventsByUserId(context.Context, *entity.GetRequest) (*entity.Events, error)
	}
)
