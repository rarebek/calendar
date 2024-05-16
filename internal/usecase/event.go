package usecase

import (
	"context"
	"fmt"
	"job_tasks/calendar/internal/entity"
)

// EventUseCase -.
type EventUseCase struct {
	repo EventRepo
}

// NewEventUseCase -.
func NewEventUseCase(r EventRepo) *EventUseCase {
	return &EventUseCase{
		repo: r,
	}
}

// CreateEvent -.
func (euc *EventUseCase) CreateEvent(ctx context.Context, req *entity.EventRequest) (*entity.EventResponse, error) {
	response, err := euc.repo.CreateEvent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("EventUseCase - CreateEvent - euc.repo.CreateEvent: %w", err)
	}

	return response, nil
}

// UpdateEvent - .
func (euc *EventUseCase) UpdateEvent(ctx context.Context, req *entity.UpdateEventRequest) (*entity.EventResponse, error) {
	response, err := euc.repo.UpdateEvent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("EventUseCase - UpdateEvent - euc.repo.UpdateEvent: %w", err)
	}

	return response, nil
}

// GetEvent - .
func (euc *EventUseCase) GetEvent(ctx context.Context, req *entity.GetRequest) (*entity.EventResponse, error) {
	response, err := euc.repo.GetEvent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("EventUseCase - GetEvent - euc.repo.GetEvent: %w", err)
	}

	return response, nil
}

// ListEvents - .
func (euc *EventUseCase) ListEvents(ctx context.Context, req *entity.GetListRequest) (*entity.Events, error) {
	response, err := euc.repo.ListEvents(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("EventUseCase - ListEvents - uuc.repo.ListEvents: %w", err)
	}

	return response, nil
}

// DeleteEvent -.
func (euc *EventUseCase) DeleteEvent(ctx context.Context, req *entity.GetRequest) error {
	return euc.repo.DeleteEvent(ctx, req)
}

// AddFileToEvent -.
func (euc *EventUseCase) AddFileToEvent(ctx context.Context, req *entity.AddFileToEventRequest) error {
	return euc.repo.AddFileToEvent(ctx, req)
}

// GetAllFilesByEventId -.
func (euc *EventUseCase) GetAllFilesByEventId(ctx context.Context, req *entity.IdRequest) (*entity.Files, error) {
	files, err := euc.repo.GetAllFilesByEventId(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("EventUseCase - GetAllFilesByEventId - uuc.repo.GetAllFilesByEventId: %w", err)
	}

	return files, nil
}

// GetExpiredEventsByUserId -.
func (euc *EventUseCase) GetExpiredEventsByUserId(ctx context.Context, req *entity.GetRequest) (*entity.Events, error) {
	events, err := euc.repo.GetExpiredEventsByUserId(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("EventUseCase - GetExpiredEventsByUserId - uuc.repo.GetExpiredEventsByUserId: %w", err)
	}

	return events, nil
}
