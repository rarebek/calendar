package v1

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"html/template"
	"job_tasks/calendar/internal/entity"
	"job_tasks/calendar/internal/usecase"
	"job_tasks/calendar/pkg/logger"
	"job_tasks/calendar/pkg/password"
	"job_tasks/calendar/pkg/token"
	"math/rand"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type userRoutes struct {
	t      usecase.User
	l      logger.Interface
	redisc *redis.Client
}

func newUserRoutes(handler *gin.RouterGroup, u usecase.User, l logger.Interface, redisc *redis.Client) {
	r := &userRoutes{u, l, redisc}

	h := handler.Group("/user")
	{
		h.POST("/create", r.CreateUser)
		h.POST("/register", r.RegisterUser)
		h.GET("/login", r.LoginUser)
		h.GET("/verify", r.VerifyUser)
		h.GET("/get", r.GetUser)
		h.PUT("/update/:id", r.UpdateUser)
		h.DELETE("/delete", r.DeleteUser)
		h.GET("/list", r.ListUsers)
	}
}

// RegisterUser
// @Summary     Register User
// @Description Ro'yxatdan o'tish, kiritgan emailingizga OTP yuboradi, keyin VERIFY orqali tasdiqlaysiz.
// @ID          register-user
// @Tags  	    Auth
// @Accept      json
// @Produce     json
// @Param       request body entity.UserRequest true "Register User Request"
// @Success     200 {object} entity.MessageResponse
// @Failure     500 {object} response
// @Router      /v1/user/register [post]
func (u *userRoutes) RegisterUser(c *gin.Context) {
	var (
		body entity.UserRequest
	)

	if err := c.ShouldBindJSON(&body); err != nil {
		u.l.Error(err, "http - v1 - register-user - c.ShouldBindJSON(body)")
		errorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	body.Email = strings.TrimSpace(body.Email)
	body.Email = strings.ToLower(body.Email)

	if err := body.IsEmail(); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Incorrect email format. Please try again",
		})
		u.l.Error(err, "http - v1 - register-user - body.IsEmail()")
		return
	}

	if err := body.IsComplexPassword(); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Password must be at least 8 characters long and contain both upper and lower case letters",
		})
		u.l.Error(err, "http - v1 - register-user - body.IsComplexPassword()")
		return
	}

	isUniqueEmail, err := u.t.CheckUniqueness(c.Request.Context(), &entity.GetRequest{
		Field: "email",
		Value: body.Email,
	})
	if err != nil {
		u.l.Error(err, "http - v1 - u.t.RegisterUser")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	if !isUniqueEmail {
		u.l.Error(err, "http - v1 - u.t.RegisterUser")
		c.JSON(http.StatusConflict, &entity.MessageResponse{Message: "This email is already in use. Please choose another email"})
	}

	isUniqueUsername, err := u.t.CheckUniqueness(c.Request.Context(), &entity.GetRequest{
		Field: "username",
		Value: body.Username,
	})
	if err != nil {
		u.l.Error(err, "http - v1 - u.t.RegisterUser")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	if !isUniqueUsername {
		u.l.Error(err, "http - v1 - u.t.RegisterUser")
		c.JSON(http.StatusConflict, &entity.MessageResponse{Message: "This username is already in use. Please choose another username"})
	}

	byteData, err := json.Marshal(body)
	if err != nil {
		u.l.Error(err, "http - v1 - u.t.RegisterUser - json.Marshal(body)")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	code := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(900000) + 100000

	type PageData struct {
		OTP string
	}
	tpl := template.Must(template.ParseFiles("index.html"))
	data := PageData{
		OTP: strconv.Itoa(code),
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		u.l.Error(err, "http - v1 - u.t.RegisterUser - tpl.Execute(body)")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}
	htmlContent := buf.Bytes()

	auth := smtp.PlainAuth("", "nodirbekgolang@gmail.com", "byxgpemogydxsuom", "smtp.gmail.com")
	err = smtp.SendMail("smtp.gmail.com:587", auth, "nodirbekgolang@gmail.com", []string{body.Email}, []byte("To: "+body.Email+"\r\nSubject: Email verification\r\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"+string(htmlContent)))

	u.redisc.Set(c.Request.Context(), strconv.Itoa(code), byteData, time.Minute*3)

	c.JSON(http.StatusOK, &entity.MessageResponse{Message: "One time password sent to your email. Please verify."})
}

// VerifyUser
// @Summary     Verify User
// @Description Tasdiqlaganingizdan so'ng, Login qilishingiz mumkin.
// @ID          verify-user
// @Tags  	    Auth
// @Accept      json
// @Produce     json
// @Param 		email query string true "Email"
// @Param 		code query string true "Code"
// @Success     200 {object} entity.VerifyResponse
// @Failure     500 {object} response
// @Router      /v1/user/verify [get]
func (u *userRoutes) VerifyUser(c *gin.Context) {
	var (
		body entity.VerifyResponse
	)
	email := c.Query("email")
	code := c.Query("code")

	value, err := u.redisc.Get(c.Request.Context(), code).Result()
	if err != nil {
		u.l.Error(err, "http - v1 - u.VerifyUser - redisc.Get()")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	err = json.Unmarshal([]byte(value), &body)
	if err != nil {
		u.l.Error(err, "http - v1 - u.VerifyUser - json.Unmarshal(body)")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	if email != body.Email {
		c.JSON(http.StatusBadRequest, entity.MessageResponse{Message: "Incorrect email"})
		return
	}

	hashedPass, err := password.HashPassword(body.Password)
	if err != nil {
		u.l.Error(err, "http - v1 - u.VerifyUser - hash.Password()")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	access, refresh, err := token.GenerateTokens(body.Email, "user")
	if err != nil {
		u.l.Error(err, "http - v1 - u.VerifyUser - token.GenerateTokens()")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	body.Password = hashedPass
	body.AccessToken = access

	createdUser, err := u.t.CreateUser(c.Request.Context(), &entity.UserRequest{
		Email:        body.Email,
		Username:     body.Username,
		Password:     hashedPass,
		RefreshToken: refresh,
	})
	if err != nil {
		u.l.Error(err, "http - v1 - u.VerifyUser - t.CreateUser()")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	body.Id = createdUser.Id

	c.JSON(http.StatusOK, body)

}

// LoginUser
// @Summary     Login User
// @Description Login User
// @ID          login-user
// @Tags  	    Auth
// @Accept      json
// @Produce     json
// @Param 		email query string true "Email"
// @Param 		password query string true "Password"
// @Success     200 {object} entity.VerifyResponse
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Router      /v1/user/login [get]
func (u *userRoutes) LoginUser(c *gin.Context) {
	email := c.Query("email")
	passwordd := c.Query("password")
	if email == "" || passwordd == "" {
		errorResponse(c, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := u.t.GetUser(c.Request.Context(), &entity.GetRequest{
		Field: "email",
		Value: email,
	})
	if err != nil {
		u.l.Error(err, "http - v1 - u.t.GetUser")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	if user == nil || !password.CheckPasswordHash(passwordd, user.Password) {
		u.l.Error(err, "http - v1 - u.t.GetUser - password.CheckPasswordHash()")
		errorResponse(c, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	accessToken, refreshToken, err := token.GenerateTokens(email, "user")
	if err != nil {
		u.l.Error(err, "http - v1 - u.LoginUser - token.GenerateTokens()")
		errorResponse(c, http.StatusInternalServerError, "token generation failed")
		return
	}
	_, err = u.t.UpdateUser(c.Request.Context(), &entity.UpdateUserRequest{
		Id:           user.Id,
		RefreshToken: refreshToken,
	})
	if err != nil {
		u.l.Error(err, "http - v1 - u.LoginUser - t.UpdateUser()")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, &entity.VerifyResponse{
		Id:          user.Id,
		Email:       user.Email,
		Username:    user.Username,
		Password:    user.Password,
		AccessToken: accessToken,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}

// CreateUser
// @Summary     Create User
// @Description Yangi user create qilish.
// @ID          create-user
// @Tags  	    User
// @Accept      json
// @Produce     json
// @Param       request body entity.UserRequest true "Create User Request"
// @Success     200 {object} entity.UserResponse
// @Failure     500 {object} response
// @Router      /v1/user/create [post]
func (u *userRoutes) CreateUser(c *gin.Context) {
	var (
		body entity.UserRequest
	)

	if err := c.ShouldBindJSON(&body); err != nil {
		u.l.Error(err, "http - v1 - create-user - c.ShouldBindJSON(body)")
		errorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	hashPassword, err := password.HashPassword(body.Password)
	if err != nil {
		u.l.Error(err, "http - v1 - create-user - hash.Password()")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	body.Password = hashPassword

	user, err := u.t.CreateUser(c.Request.Context(), &body)
	if err != nil {
		u.l.Error(err, "http - v1 - u.t.CreateEvent")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUser
// @Summary     User
// @Description Get user fieldga qaysi fielddan olish va valuega osha filedning qiymatini kiritasiz.
// @ID          get-user
// @Tags  	    User
// @Accept      json
// @Produce     json
// @Param field query string true "Field request for User"
// @Param value query string true "Value Request for User"
// @Success     200 {object} entity.UserResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/user/get [get]
func (u *userRoutes) GetUser(c *gin.Context) {
	field := c.Query("field")
	value := c.Query("value")
	user, err := u.t.GetUser(c.Request.Context(), &entity.GetRequest{
		Field: field,
		Value: value,
	})
	if err != nil {
		u.l.Error(err, "http - v1 - u.t.GetUser")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser
// @Summary     User
// @Description Qaysi Userni update qilish, pathdagi id bilan va bodyga update bo'lishi kerak bo'lgan userning ma'lumotlari.
// @ID          update-user
// @Tags  	    User
// @Accept      json
// @Produce     json
// @Param       id path string true "User ID to update"
// @Param       request body entity.UpdateUserRequest true "User details to update"
// @Success     200 {object} entity.UserResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/user/update/{id} [put]
func (u *userRoutes) UpdateUser(c *gin.Context) {
	var (
		body entity.UserRequest
	)
	if err := c.ShouldBindJSON(&body); err != nil {
		u.l.Error(err, "http - v1 - update-user - c.ShouldBindJSON(body)")
		errorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	id := c.Param("id")
	user, err := u.t.UpdateUser(c.Request.Context(), &entity.UpdateUserRequest{
		Id:           id,
		Email:        body.Email,
		Username:     body.Username,
		Password:     body.Password,
		RefreshToken: body.RefreshToken,
	})
	if err != nil {
		u.l.Error(err, "http - v1 - u.t.UpdateUser")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers
// @Summary     ListUsers
// @Description ListUser page limit kiritish majburiy. Field va Value orqali Search qilishingiz va OrderBy orqali tartiblashingiz mumkin.
// @ID          list-users
// @Tags  	    User
// @Accept      json
// @Produce     json
// @Param page query string true "User Page request"
// @Param limit query string true "User Limit request"
// @Param orderBy query string false "User OrderBy request"
// @Param field query string false "User Field request"
// @Param value query string false "User Value request"
// @Success     200 {object} entity.Users
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/user/list [get]
func (u *userRoutes) ListUsers(c *gin.Context) {
	page := c.Query("page")
	pageint, err := strconv.Atoi(page)
	limit := c.Query("limit")
	limitint, err := strconv.Atoi(limit)
	orderBy := c.Query("orderBy")
	field := c.Query("field")
	value := c.Query("value")

	users, err := u.t.ListUsers(c.Request.Context(), &entity.GetListRequest{
		Page:    int64(pageint),
		Limit:   int64(limitint),
		OrderBy: orderBy,
		Field:   field,
		Value:   value,
	})
	if err != nil {
		u.l.Error(err, "http - v1 - u.t.ListUsers")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, users)
}

// DeleteUser
// @Summary     User
// @Description userni field va value orqali o'chirish.
// @ID          delete-user
// @Tags  	    User
// @Accept      json
// @Produce     json
// @Param field query string true "User field"
// @Param value query string true "User Value"
// @Success     200 {object} entity.MessageResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /v1/user/delete [delete]
func (u *userRoutes) DeleteUser(c *gin.Context) {
	field := c.Query("field")
	value := c.Query("value")

	if err := u.t.DeleteUser(c.Request.Context(), &entity.GetRequest{
		Field: field,
		Value: value,
	}); err != nil {
		u.l.Error(err, "http - v1 - u.t.DeleteUser")
		errorResponse(c, http.StatusInternalServerError, "database problems")
		return
	}

	c.JSON(http.StatusOK, &entity.MessageResponse{Message: "User deleted successfully"})

}
