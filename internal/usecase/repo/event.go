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
	eventsTableName      = "events"
	filesTableName       = "files"
	eventsFilesTableName = "events_files"
)

// EventRepo -.
type EventRepo struct {
	*postgres.Postgres
}

// NewEventRepo -.
func NewEventRepo(pg *postgres.Postgres) *EventRepo {
	return &EventRepo{pg}
}

// CreateEvent -.
func (e *EventRepo) CreateEvent(ctx context.Context, event *entity.EventRequest) (*entity.EventResponse, error) {
	var (
		eventResponse entity.EventResponse
		updatedAt     sql2.NullTime
	)
	data := map[string]interface{}{
		"id":          uuid.NewString(),
		"user_id":     event.UserId,
		"title":       event.Title,
		"description": event.Description,
		"event_time":  event.EventTime,
	}
	sql, args, err := e.Builder.
		Insert(eventsTableName).
		SetMap(data).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("EventRepo CreateEvent - e.Builder: %w", err)
	}
	sql += " RETURNING id, user_id, title, description, event_time, created_at, updated_at"

	row := e.Pool.QueryRow(ctx, sql, args...)
	if err := row.Scan(&eventResponse.Id, &eventResponse.UserId, &eventResponse.Title, &eventResponse.Description, &eventResponse.EventTime, &eventResponse.CreatedAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("EventRepo - CreateEvent - row.Scan: %w", err)
	}
	if updatedAt.Valid {
		eventResponse.UpdatedAt = updatedAt.Time
	}
	return &eventResponse, nil
}

// GetEvent - .
func (e *EventRepo) GetEvent(ctx context.Context, req *entity.GetRequest) (*entity.EventResponse, error) {
	var (
		event     entity.EventResponse
		updatedAt sql2.NullTime
	)
	sql, args, err := e.Builder.
		Select("id, user_id, title, description, event_time, created_at, updated_at").
		From(eventsTableName).
		Where(squirrel.And{
			squirrel.Eq{req.Field: req.Value},
			squirrel.Eq{"deleted_at": nil},
		}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("EventRepo - GetEvent - e.Builder: %w", err)
	}

	row := e.Pool.QueryRow(ctx, sql, args...)

	if err := row.Scan(&event.Id, &event.UserId, &event.Title, &event.Description, &event.EventTime, &event.CreatedAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("EventRepo - GetEvent - row.Scan: %w", err)
	}
	if updatedAt.Valid {
		event.UpdatedAt = updatedAt.Time
	}
	return &event, nil
}

// UpdateEvent - .
func (e *EventRepo) UpdateEvent(ctx context.Context, event *entity.UpdateEventRequest) (*entity.EventResponse, error) {
	var (
		response  entity.EventResponse
		updatedAt sql2.NullTime
	)
	data := map[string]interface{}{
		"title":       event.Title,
		"description": event.Description,
		"event_time":  event.EventTime,
		"updated_at":  time.Now(),
	}
	sql, args, err := e.Builder.
		Update(eventsTableName).
		SetMap(data).
		Where(squirrel.And{
			squirrel.Eq{"id": event.Id},
			squirrel.Eq{"deleted_at": nil},
		},
		).ToSql()
	sql += " RETURNING id, user_id, title, description, event_time, created_at, updated_at"
	if err != nil {
		return nil, fmt.Errorf("EventRepo - UpdateEvent - e.Builder: %w", err)
	}

	row := e.Pool.QueryRow(ctx, sql, args...)
	if err := row.Scan(&response.Id, &response.UserId, &response.Title, &response.Description, &response.EventTime, &response.CreatedAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("eventRepo - UpdateEvent - row.Scan: %w", err)
	}
	if updatedAt.Valid {
		response.UpdatedAt = updatedAt.Time
	}
	return &response, nil
}

// ListEvents - .
func (e *EventRepo) ListEvents(ctx context.Context, req *entity.GetListRequest) (*entity.Events, error) {
	var (
		events entity.Events
	)
	builder := e.Builder.Select(
		"id", "user_id", "title", "description", "event_time", "created_at, updated_at").
		From(eventsTableName).Where(squirrel.Eq{"deleted_at": nil})
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

	sql, _, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ListEvents - u.eventsSelectQueryPrefix: %w", err)
	}
	rows, err := e.Pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("ListUsers  - u.Pool.Query: %w", err)
	}

	for rows.Next() {
		var (
			event     entity.EventResponse
			updatedAt sql2.NullTime
		)

		if err := rows.Scan(
			&event.Id,
			&event.UserId,
			&event.Title,
			&event.Description,
			&event.EventTime,
			&event.CreatedAt,
			&updatedAt); err != nil {
			return nil, fmt.Errorf("EventRepo - ListEvents - u.Row.Scan: %w", err)
		}

		if updatedAt.Valid {
			event.UpdatedAt = updatedAt.Time
		}

		events.Count++
		events.Events = append(events.Events, &event)
	}

	return &events, nil
}

// DeleteEvent - .
func (e *EventRepo) DeleteEvent(ctx context.Context, req *entity.GetRequest) error {
	sql, args, err := e.Builder.Update(eventsTableName).
		Set("deleted_at", time.Now()).
		Where(squirrel.And{
			squirrel.Eq{"deleted_at": nil},
			squirrel.Eq{req.Field: req.Value},
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("DeleteEvent - e.Builder: %w", err)
	}
	fmt.Println(sql)
	fmt.Println(req.Value)

	if _, err := e.Pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("DeleteEvent - u.Pool.Exec: %w", err)
	}

	return nil
}

// AddFileToEvent -.
func (e *EventRepo) AddFileToEvent(ctx context.Context, req *entity.AddFileToEventRequest) error {
	var (
		fileId string
	)
	sql, args, err := e.Builder.Insert(filesTableName).
		Columns("id, file_path").
		Values(uuid.NewString(), req.FilePath).
		ToSql()
	if err != nil {
		return fmt.Errorf("AddFileToEvent - e.Builder.Insert: %w", err)
	}
	sql += " RETURNING id"

	row := e.Pool.QueryRow(ctx, sql, args...)
	if err := row.Scan(&fileId); err != nil {
		return fmt.Errorf("AddFileToEvent - e.Pool.QueryRow: %w", err)
	}

	data := map[string]interface{}{
		"event_id": req.EventId,
		"file_id":  fileId,
	}

	sql, args, err = e.Builder.Insert(eventsFilesTableName).
		SetMap(data).ToSql()
	if err != nil {
		return fmt.Errorf("AddFileToEvent - e.Builder.Insert: %w", err)
	}
	_, err = e.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("AddFileToEvent - u.Pool.Exec: %w", err)
	}

	return nil
}

// GetAllFilesByEventId -.
func (e *EventRepo) GetAllFilesByEventId(ctx context.Context, req *entity.IdRequest) (*entity.Files, error) {
	var (
		files entity.Files
		file  entity.File
	)
	sql, args, err := e.Builder.Select("file_id").
		From(eventsFilesTableName).
		Where(
			squirrel.Eq{"event_id": req.EventId},
		).ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetAllFilesByEventId - e.Builder: %w", err)
	}

	rows, err := e.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("GetAllFilesByEventId - u.Pool.Query: %w", err)
	}

	for rows.Next() {
		var (
			FileId   string
			FilePath string
		)
		if err := rows.Scan(&FileId); err != nil {
			return nil, fmt.Errorf("GetAllFilesByEventId - e.Pool.Query: %w", err)
		}

		sql, args, err := e.Builder.Select("file_path").
			From(filesTableName).
			Where(squirrel.And{
				squirrel.Eq{"deleted_at": nil},
				squirrel.Eq{"id": FileId},
			}).ToSql()
		if err != nil {
			return nil, fmt.Errorf("GetAllFilesByEventId - e.Pool.Query: %w", err)
		}
		row := e.Pool.QueryRow(ctx, sql, args...)
		if err := row.Scan(&FilePath); err != nil {
			return nil, fmt.Errorf("GetAllFilesByEventId - e.Pool.Query: %w", err)
		}
		file.Id = FileId
		file.FilePath = FilePath
		files.Count++
		files.Files = append(files.Files, &file)
	}

	return &files, nil
}

// GetExpiredEventsByUserId -.
func (e *EventRepo) GetExpiredEventsByUserId(ctx context.Context, req *entity.GetRequest) (*entity.Events, error) {
	var (
		events entity.Events
	)
	sql, args, err := e.Builder.Select(
		"id", "user_id", "title", "description", "event_time", "created_at, updated_at").
		From(eventsTableName).
		Where(squirrel.And{
			squirrel.Eq{req.Field: req.Value},
			squirrel.Lt{"event_time": time.Now()},
			squirrel.Eq{"deleted_at": nil},
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetExpiredEventsByUserID - e.Builder: %w", err)
	}
	rows, err := e.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("GetExpiredEventsByUserID - u.Pool.Query: %w", err)
	}

	for rows.Next() {
		var (
			event     entity.EventResponse
			updatedAt sql2.NullTime
		)
		if err := rows.Scan(
			&event.Id,
			&event.UserId,
			&event.Title,
			&event.Description,
			&event.EventTime,
			&event.CreatedAt,
			&updatedAt); err != nil {
			return nil, fmt.Errorf("EventRepo - GetExpiredEventsByUserID - u.Row.Scan: %w", err)
		}
		if updatedAt.Valid {
			event.UpdatedAt = updatedAt.Time
		}
		events.Count++
		events.Events = append(events.Events, &event)
	}

	return &events, nil
}
