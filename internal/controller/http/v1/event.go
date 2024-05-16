package v1

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/k0kubun/pp"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"job_tasks/calendar/internal/entity"
	"job_tasks/calendar/internal/usecase"
	"job_tasks/calendar/pkg/logger"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
)

type eventRoutes struct {
	t usecase.Event
	l logger.Interface
}

type File struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func newEventRoutes(handler *gin.RouterGroup, e usecase.Event, l logger.Interface) {
	r := &eventRoutes{e, l}

	h := handler.Group("/event")
	{
		h.POST("/create", r.CreateEvent)
		h.GET("/get", r.GetEvent)
		h.PUT("/update/:id", r.UpdateEvent)
		h.DELETE("/delete", r.DeleteEvent)
		h.GET("/list", r.ListEvents)
		h.POST("/file-upload", r.UploadFile)
		h.GET("/files", r.GetAllFilesByEventId)
		h.GET("/expired", r.GetExpiredEventsByUserId)
	}
}

// CreateEvent
// @Summary     Create Event
// @Description Create. POST request bilan body orqali beriladi.
// @ID          create-event
// @Tags  	    Event
// @Accept      json
// @Produce     json
// @Param       request body entity.EventRequest true "Create Event Request"
// @Success     200 {object} entity.EventResponse
// @Failure     500 {object} response
// @Router      /v1/event/create [post]
func (e *eventRoutes) CreateEvent(c *gin.Context) {
	var (
		body entity.EventRequest
	)

	if err := c.ShouldBindJSON(&body); err != nil {
		e.l.Error(err, "http - v1 - create-event - c.ShouldBindJSON(body)")
		errorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	event, err := e.t.CreateEvent(c.Request.Context(), &body)
	if err != nil {
		e.l.Error(err, "http - v1 - u.t.CreateEvent")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, event)
}

// GetEvent
// @Summary     Event
// @Description Fieldga qaysi fielddan qidirishni va valuega o'sha fieldning valuesi kiritiladi.
// @ID          get-event
// @Tags  	    Event
// @Accept      json
// @Produce     json
// @Param field query string true "Field request for Event"
// @Param value query string true "Value Request for Event"
// @Success     200 {object} entity.EventResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/event/get [get]
func (e *eventRoutes) GetEvent(c *gin.Context) {
	field := c.Query("field")
	value := c.Query("value")
	event, err := e.t.GetEvent(c.Request.Context(), &entity.GetRequest{
		Field: field,
		Value: value,
	})
	if err != nil {
		e.l.Error(err, "http - v1 - u.t.GetEvent")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, event)
}

// UpdateEvent
// @Summary     Event
// @Description Event ID ni pathga kiritiladi, bodydan esa, qaysi ma'lumotlar update bo'lishligi beriladi.
// @ID          update-event
// @Tags  	    Event
// @Accept      json
// @Produce     json
// @Param       id path string true "Event ID to update"
// @Param       request body entity.UpdateEventRequest true "Event details to update"
// @Success     200 {object} entity.EventResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/event/update/{id} [put]
func (e *eventRoutes) UpdateEvent(c *gin.Context) {
	var (
		body entity.EventRequest
	)
	if err := c.ShouldBindJSON(&body); err != nil {
		e.l.Error(err, "http - v1 - update-event - c.ShouldBindJSON(body)")
		errorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	id := c.Param("id")
	event, err := e.t.UpdateEvent(c.Request.Context(), &entity.UpdateEventRequest{
		Id:          id,
		Title:       body.Title,
		Description: body.Description,
		EventTime:   body.EventTime,
	})
	if err != nil {
		e.l.Error(err, "http - v1 - u.t.UpdateEvent")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, event)
}

// ListEvents
// @Summary     ListEvents
// @Description Page va Limit majburiy, field va value orqali eventlarni search qilish ham mumkin.
// @ID          list-events
// @Tags  	    Event
// @Accept      json
// @Produce     json
// @Param page query string true "Event Page request"
// @Param limit query string true "Event Limit request"
// @Param orderBy query string false "Event OrderBy request"
// @Param field query string false "Event Field request"
// @Param value query string false "Event Value request"
// @Success     200 {object} entity.Events
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/event/list [get]
func (e *eventRoutes) ListEvents(c *gin.Context) {
	page := c.Query("page")
	pageint, err := strconv.Atoi(page)
	limit := c.Query("limit")
	limitint, err := strconv.Atoi(limit)
	orderBy := c.Query("orderBy")
	field := c.Query("field")
	value := c.Query("value")

	events, err := e.t.ListEvents(c.Request.Context(), &entity.GetListRequest{
		Page:    int64(pageint),
		Limit:   int64(limitint),
		OrderBy: orderBy,
		Field:   field,
		Value:   value,
	})
	if err != nil {
		e.l.Error(err, "http - v1 - u.t.ListEvents")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, events)
}

// DeleteEvent
// @Summary     Events
// @Description Eventni soft delete qiladi. Fieldga id va valuega event idni berishingiz mumkin.
// @ID          delete-event
// @Tags  	    Event
// @Accept      json
// @Produce     json
// @Param field query string true "Event field"
// @Param value query string true "Event Value"
// @Success     200 {object} entity.MessageResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/event/delete [delete]
func (e *eventRoutes) DeleteEvent(c *gin.Context) {
	field := c.Query("field")
	value := c.Query("value")

	if err := e.t.DeleteEvent(c.Request.Context(), &entity.GetRequest{
		Field: field,
		Value: value,
	}); err != nil {
		e.l.Error(err, "http - v1 - u.t.DeleteEvent")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, &entity.MessageResponse{Message: "Event deleted successfully"})

}

// UploadFile
// @Summary     File Upload
// @Description Fayl yuklash va u faylni query orqali berilgan eventga biriktirib qo'yish mumkin.
// @Tags        Event
// @Accept      multipart/form-data
// @Produce     json
// @Param       file formData file true "File"
// @Param 		EventId query string false "Event Id request"
// @Success     200 {object} string
// @Failure     400 {object} string
// @Failure     500 {object} string
// @Router      /v1/event/file-upload [post]
func (e *eventRoutes) UploadFile(c *gin.Context) {
	endpoint := "0.0.0.0:9000"
	accessKeyID := "nodirbek"
	secretAccessKey := "nodirbek"
	bucketName := "files"
	eventId := c.Query("EventId")

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error connecting to MinIO server")
		log.Println("Error connecting to MinIO server:", err)
		return
	}

	var file File
	if err := c.ShouldBind(&file); err != nil {
		c.JSON(http.StatusBadRequest, "Error uploading file")
		log.Println("Error uploading file:", err)
		return
	}

	ext := filepath.Ext(file.File.Filename)
	id := uuid.New().String()
	objectName := id + ext

	fileReader, err := file.File.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error opening file")
		log.Println("Error opening file:", err)
		return
	}
	defer fileReader.Close()

	// Get the content type of the file
	contentType := file.File.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream" // default content type if unknown
	}

	_, err = minioClient.PutObject(context.Background(), bucketName, objectName, fileReader, file.File.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error uploading file to MinIO server")
		log.Println("Error uploading file to MinIO server:", err)
		return
	}

	minioURL := fmt.Sprintf("http://%s/%s/%s", "0.0.0.0:9000", bucketName, objectName)

	err = e.t.AddFileToEvent(c.Request.Context(), &entity.AddFileToEventRequest{
		EventId:  eventId,
		FilePath: minioURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error uploading file to MinIO server")
		log.Println("Error uploading file to MinIO server:", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":      minioURL,
		"event_id": eventId,
	})
}

// GetAllFilesByEventId
// @Summary     Event
// @Description Event ID ga bog'langan barcha filelarni olib keladi.
// @ID          get-all-files-by-event-id
// @Tags  	    Event
// @Accept      json
// @Produce     json
// @Param 		event-id query string true "Event id request"
// @Success     200 {object} entity.Files
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/event/files [get]
func (e *eventRoutes) GetAllFilesByEventId(c *gin.Context) {
	eventId := c.Query("event-id")

	files, err := e.t.GetAllFilesByEventId(c.Request.Context(), &entity.IdRequest{EventId: eventId})
	if err != nil {
		pp.Println(err)
		e.l.Error(err, "http - v1 - u.t.GetAllFilesByEventId\n")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, files)
}

// GetExpiredEventsByUserId
// @Summary     Event
// @Description Event vaqti xozirgi vaqtdan oldingilarni qaytaradi..
// @ID          get-files
// @Tags  	    Event
// @Accept      json
// @Produce     json
// @Param 		user-id query string true "User id request"
// @Success     200 {object} entity.Events
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/event/expired [get]
func (e *eventRoutes) GetExpiredEventsByUserId(c *gin.Context) {
	UserId := c.Query("user-id")

	events, err := e.t.GetExpiredEventsByUserId(c.Request.Context(), &entity.GetRequest{
		Field: "user_id",
		Value: UserId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, "database problems")
		log.Println("Error getting expired events by user id:", err)
		return
	}

	c.JSON(http.StatusOK, events)
}
