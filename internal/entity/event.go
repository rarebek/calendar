package entity

import "time"

type EventRequest struct {
	UserId      string    `json:"user_id" example:"8ac01585-4559-49d1-8708-283e83da9b05"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventTime   time.Time `json:"event_time" example:"2024-05-16T12:00:00Z"`
}

type EventResponse struct {
	Id          string    `json:"id"`
	UserId      string    `json:"user_id" example:"8ac01585-4559-49d1-8708-283e83da9b05"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventTime   time.Time `json:"event_time" example:"2024-05-16T12:00:00Z"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateEventRequest struct {
	Id          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventTime   time.Time `json:"event_time" example:"2024-05-16T12:00:00Z"`
}

type Events struct {
	Events []*EventResponse `json:"events"`
	Count  int              `json:"count"`
}

type AddFileToEventRequest struct {
	EventId  string `json:"event_id"`
	FilePath string `json:"file_path"`
}

type EventFile struct {
	EventId string `json:"event_id"`
	FileId  string `json:"file_id"`
}

type File struct {
	Id       string `json:"id"`
	FilePath string `json:"file_path"`
}

type Files struct {
	Files []*File `json:"files"`
	Count int     `json:"count"`
}

type IdRequest struct {
	EventId string `json:"id"`
}
