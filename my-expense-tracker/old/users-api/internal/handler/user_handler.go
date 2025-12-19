package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/education/users-api/internal/model"
	"gitlab.com/education/users-api/internal/repository"
	"gitlab.com/education/users-api/internal/service"
)

// Хендлеры пользователя
type UserHandler struct {
	service service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

// Register маршрутов модуля пользователя
func (h *UserHandler) Register(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		users.GET("", h.ListUsers)
		users.POST("", h.CreateUser)
		users.GET("/:id", h.GetUser)
		users.PATCH("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
	}
}

// ListUsers godoc
// @Summary      Список пользователей
// @Description  Возвращает список пользователей
// @Tags         users
// @Produce      json
// @Success      200  {array}   model.UserResponse
// @Router       /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	ctx := c.Request.Context()
	users, err := h.service.ListUsers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := make([]model.UserResponse, 0, len(users))
	for i := range users {
		resp = append(resp, model.ToUserResponse(&users[i]))
	}
	c.JSON(http.StatusOK, resp)
}

// CreateUser godoc
// @Summary      Создать пользователя
// @Description  Создаёт нового пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      model.CreateUserRequest  true  "Данные пользователя"
// @Success      201   {object}  model.UserResponse
// @Router       /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.service.CreateUser(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, model.ToUserResponse(u))
}

// GetUser godoc
// @Summary      Получить пользователя
// @Description  Возвращает пользователя по ID
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "ID пользователя"  example(11111111-1111-1111-1111-111111111111)
// @Success      200  {object}  model.UserResponse
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	u, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.ToUserResponse(u))
}

// UpdateUser godoc
// @Summary      Обновить пользователя
// @Description  Частично обновляет данные пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "ID пользователя"  example(11111111-1111-1111-1111-111111111111)
// @Param        user  body      model.UpdateUserRequest  true  "Изменяемые поля"
// @Success      200   {object}  model.UserResponse
// @Failure      404   {object}  map[string]string
// @Router       /users/{id} [patch]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.service.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.ToUserResponse(u))
}

// DeleteUser godoc
// @Summary      Удалить пользователя
// @Description  Удаляет пользователя по ID
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "ID пользователя"  example(11111111-1111-1111-1111-111111111111)
// @Success      204  {object}  nil
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
