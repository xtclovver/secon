package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings" // Добавляем импорт strings
	"time"

	"github.com/gin-gonic/gin"

	"vacation-scheduler/internal/models"
	"vacation-scheduler/internal/services"
)

// Helper function to get an integer query parameter
func GetIntQueryParam(c *gin.Context, paramName string) *int {
	valStr := c.Query(paramName)
	if valStr == "" {
		return nil
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		// Можно логировать ошибку или просто игнорировать некорректный параметр
		log.Printf("Некорректное значение для параметра '%s': %v", paramName, err)
		return nil
	}
	return &val
}

// AppHandler объединяет обработчики для разных частей приложения
type AppHandler struct {
	vacationService services.VacationServiceInterface // Используем интерфейс
	userService     services.UserServiceInterface     // Добавляем сервис пользователей
}

// NewAppHandler создает новый экземпляр AppHandler
func NewAppHandler(vs services.VacationServiceInterface, us services.UserServiceInterface) *AppHandler {
	return &AppHandler{
		vacationService: vs,
		userService:     us, // Инициализируем userService
	}
}

// GetVacationLimit обработчик для получения лимита отпуска
func (h *AppHandler) GetVacationLimit(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат года"})
		return
	}

	// Получаем ID пользователя из контекста (установленного middleware аутентификации)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	// Получаем лимит отпуска из сервиса
	limit, err := h.vacationService.GetVacationLimit(userID.(int), year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения лимита отпуска: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, limit)
}

// SetVacationLimit обработчик для установки лимита отпуска администратором
func (h *AppHandler) SetVacationLimit(c *gin.Context) {
	// Структура для данных из тела запроса
	var input struct {
		UserID    int `json:"user_id" binding:"required"`
		Year      int `json:"year" binding:"required"`
		TotalDays int `json:"total_days" binding:"required"`
	}

	// Проверяем права администратора (предполагается, что middleware это делает)
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права администратора"})
		return
	}

	// Привязываем JSON из тела запроса
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные: " + err.Error()})
		return
	}

	// Вызываем сервис для установки лимита
	err := h.vacationService.SetVacationLimit(input.UserID, input.Year, input.TotalDays)
	if err != nil {
		// Обрабатываем возможные ошибки сервиса (например, отрицательное количество дней)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка установки лимита: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Лимит отпуска успешно установлен"})
}

// CreateVacationRequest обработчик для создания заявки на отпуск
func (h *AppHandler) CreateVacationRequest(c *gin.Context) {
	var request models.VacationRequest

	// Логируем тело запроса перед привязкой
	// bodyBytes, _ := io.ReadAll(c.Request.Body) // Читаем тело
	// c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Восстанавливаем тело для ShouldBindJSON
	// log.Printf("Получено тело запроса CreateVacationRequest: %s", string(bodyBytes))

	// Привязываем JSON с использованием encoding/json напрямую
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&request); err != nil {
		log.Printf("Ошибка декодирования JSON в CreateVacationRequest с encoding/json: %v", err) // Логируем ошибку декодирования
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка чтения данных: " + err.Error()})
		return
	}

	// Логируем распарсенный запрос
	log.Printf("Распарсенный запрос CreateVacationRequest (encoding/json): %+v", request)
	for i, p := range request.Periods {
		log.Printf("Период %d: Start=%v, End=%v, Days=%d", i+1, p.StartDate, p.EndDate, p.DaysCount)
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}
	request.UserID = userID.(int)

	// Валидация заявки
	if err := h.vacationService.ValidateVacationRequest(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Сохранение заявки
	if err := h.vacationService.SaveVacationRequest(&request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения заявки: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, request)
}

// SubmitVacationRequest обработчик для отправки заявки руководителю
func (h *AppHandler) SubmitVacationRequest(c *gin.Context) {
	requestIDStr := c.Param("id")
	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID заявки"})
		return
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	// Отправляем заявку руководителю
	if err := h.vacationService.SubmitVacationRequest(requestID, userID.(int)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отправки заявки: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка успешно отправлена руководителю"})
}

// GetVacationIntersections обработчик для получения пересечений отпусков
func (h *AppHandler) GetVacationIntersections(c *gin.Context) {
	departmentIDStr := c.Query("departmentId")
	departmentID, err := strconv.Atoi(departmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID подразделения"})
		return
	}

	yearStr := c.Query("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат года"})
		return
	}

	// Проверяем, является ли пользователь руководителем
	isManager, exists := c.Get("isManager")
	if !exists || !isManager.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права руководителя"})
		return
	}

	// Получаем пересечения отпусков
	intersections, err := h.vacationService.CheckIntersections(departmentID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при проверке пересечений: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, intersections)
}

// GetMyVacations обработчик для получения собственных заявок на отпуск
func (h *AppHandler) GetMyVacations(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	yearStr := c.Query("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат года"})
		return
	}

	// TODO: Получить statusFilter из query параметров, если нужно
	var statusFilter *int // Пока nil

	// Получение заявок пользователя из сервиса
	vacations, err := h.vacationService.GetUserVacations(userID.(int), year, statusFilter) // Добавлен statusFilter
	if err != nil {
		// Возвращаем общую ошибку сервера
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения заявок: " + err.Error()})
		return
	}

	// Возвращаем найденные заявки
	c.JSON(http.StatusOK, vacations)
}

// GetDepartmentVacations обработчик для получения отпусков сотрудников подразделения
func (h *AppHandler) GetDepartmentVacations(c *gin.Context) {
	isManager, exists := c.Get("isManager")
	if !exists || !isManager.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права руководителя"})
		return
	}

	departmentIDStr := c.Param("id")
	departmentID, err := strconv.Atoi(departmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID подразделения"})
		return
	}

	yearStr := c.Query("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат года"})
		return
	}

	// TODO: Получить statusFilter из query параметров, если нужно
	var statusFilter *int // Пока nil

	// Получение заявок подразделения из сервиса
	vacations, err := h.vacationService.GetDepartmentVacations(departmentID, year, statusFilter) // Добавлен statusFilter
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения заявок подразделения: " + err.Error()})
		return
	}

	// Возвращаем найденные заявки
	c.JSON(http.StatusOK, vacations)
}

// GetAllVacations обработчик для получения всех заявок (для админов/менеджеров)
func (h *AppHandler) GetAllVacations(c *gin.Context) {
	// Получаем ID запрашивающего пользователя из контекста
	requestingUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	// Получаем фильтры из query параметров
	yearFilter := GetIntQueryParam(c, "year")
	statusFilter := GetIntQueryParam(c, "status")
	userIDFilter := GetIntQueryParam(c, "userId")
	departmentIDFilter := GetIntQueryParam(c, "departmentId")

	// Если год не указан, можно использовать текущий или требовать параметр
	if yearFilter == nil {
		currentYear := time.Now().Year()
		yearFilter = &currentYear
		// Или вернуть ошибку:
		// c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр 'year' обязателен"})
		// return
	}

	// Вызываем сервис для получения всех заявок с учетом прав и фильтров
	vacations, err := h.vacationService.GetAllUserVacations(requestingUserID.(int), yearFilter, statusFilter, userIDFilter, departmentIDFilter)
	if err != nil {
		// Обрабатываем ошибки прав доступа или другие ошибки сервиса
		if err.Error() == "недостаточно прав для просмотра всех заявок" || err.Error() == "менеджер может просматривать заявки только своего отдела" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения списка заявок: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, vacations)
}

// CancelVacationRequest обработчик для отмены заявки (пользователем, менеджером, админом)
func (h *AppHandler) CancelVacationRequest(c *gin.Context) {
	requestIDStr := c.Param("id")
	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID заявки"})
		return
	}

	// Получаем ID пользователя из контекста
	cancellingUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	// Вызываем сервис для отмены заявки
	err = h.vacationService.CancelVacationRequest(requestID, cancellingUserID.(int))
	if err != nil {
		// Обрабатываем возможные ошибки (заявка не найдена, нет прав, нельзя отменить и т.д.)
		// Определяем код ошибки
		statusCode := http.StatusInternalServerError
		if err.Error() == "заявка не найдена" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "нет прав на отмену этой заявки" {
			statusCode = http.StatusForbidden
		} else if strings.HasPrefix(err.Error(), "нельзя отменить заявку в статусе") {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": "Ошибка отмены заявки: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка успешно отменена"})
}

// ApproveVacationRequest обработчик для утверждения заявки (менеджер/админ)
func (h *AppHandler) ApproveVacationRequest(c *gin.Context) {
	requestIDStr := c.Param("id")
	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID заявки"})
		return
	}

	// Получаем ID утверждающего пользователя из контекста
	approverID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	// Вызываем сервис для утверждения заявки
	err = h.vacationService.ApproveVacationRequest(requestID, approverID.(int))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := "Ошибка утверждения заявки: " + err.Error()
		if err.Error() == "заявка не найдена" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "недостаточно прав для утверждения этой заявки" {
			statusCode = http.StatusForbidden
		} else if strings.HasPrefix(err.Error(), "можно утвердить только заявку в статусе") || strings.HasPrefix(err.Error(), "недостаточно дней отпуска у сотрудника") {
			statusCode = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "ошибка при списании дней из лимита") {
			// Ошибка списания дней - критическая, но заявка уже утверждена
			statusCode = http.StatusConflict // Используем 409 Conflict для индикации частичного успеха с проблемой
			errMsg = "Заявка утверждена, но не удалось списать дни: " + err.Error()
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка успешно утверждена"})
}

// RejectVacationRequest обработчик для отклонения заявки (менеджер/админ)
func (h *AppHandler) RejectVacationRequest(c *gin.Context) {
	requestIDStr := c.Param("id")
	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID заявки"})
		return
	}

	// Получаем ID отклоняющего пользователя из контекста
	rejecterID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	// Получаем причину отклонения из тела запроса (опционально)
	var input struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil && err.Error() != "EOF" { // Игнорируем EOF, если тело пустое
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат причины отклонения: " + err.Error()})
		return
	}

	// Вызываем сервис для отклонения заявки
	err = h.vacationService.RejectVacationRequest(requestID, rejecterID.(int), input.Reason)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "заявка не найдена" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "недостаточно прав для отклонения этой заявки" {
			statusCode = http.StatusForbidden
		} else if strings.HasPrefix(err.Error(), "можно отклонить только заявку в статусе") {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": "Ошибка отклонения заявки: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка успешно отклонена"})
}

// GetAllUsersWithLimits обработчик для получения списка пользователей с лимитами (для админа)
func (h *AppHandler) GetAllUsersWithLimits(c *gin.Context) {
	// Проверяем права администратора
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права администратора"})
		return
	}

	// Получаем год из query параметра
	yearStr := c.Query("year")
	if yearStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр 'year' обязателен"})
		return
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат года"})
		return
	}

	// Вызываем сервис для получения пользователей с лимитами
	users, err := h.userService.GetAllUsersWithLimits(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователей: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// AuthHandler - структура для обработчиков аутентификации (добавлено для main.go)
type AuthHandler struct {
	// Здесь должны быть зависимости, например, сервис аутентификации
	authService *services.AuthService // Предполагаем, что такой сервис существует
}

// NewAuthHandler - конструктор для AuthHandler (добавлено для main.go)
func NewAuthHandler(as *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: as,
	}
}

// Login - обработчик для входа пользователя
func (h *AuthHandler) Login(c *gin.Context) {
	var credentials struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// Привязываем JSON из тела запроса к структуре credentials
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные: " + err.Error()})
		return
	}

	// Вызываем сервис для проверки логина и пароля
	token, user, err := h.authService.Login(credentials.Username, credentials.Password)
	if err != nil {
		// Если сервис вернул ошибку (неверные данные, ошибка БД и т.д.)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Отправляем токен и данные пользователя в ответе
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user, // Убедитесь, что пароль удален в сервисе перед возвратом
	})
}

// Register - обработчик для регистрации нового пользователя
func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Username        string `json:"username" binding:"required"`
		Password        string `json:"password" binding:"required"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
		FullName        string `json:"full_name" binding:"required"`
		Email           string `json:"email" binding:"required,email"`
		PositionID      *int   `json:"position_id"` // Должность опциональна при регистрации? Или required? Пока опционально.
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные: " + err.Error()})
		return
	}

	// Проверка совпадения паролей
	if input.Password != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пароли не совпадают"})
		return
	}

	// Вызов сервиса регистрации
	user, err := h.authService.Register(input.Username, input.Password, input.FullName, input.Email, input.PositionID)
	if err != nil {
		// Обработка ошибок сервиса (например, пользователь уже существует)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // Используем 409 Conflict для дубликата
		return
	}

	// Возвращаем созданного пользователя (без пароля)
	c.JSON(http.StatusCreated, user)
}

// GetPositions - обработчик для получения списка должностей
func (h *AppHandler) GetPositions(c *gin.Context) {
	positions, err := h.userService.GetAllPositionsGrouped()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения списка должностей: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, positions)
}

// UpdateUserProfile - обработчик для обновления профиля пользователя
func (h *AppHandler) UpdateUserProfile(c *gin.Context) {
	// Получаем ID целевого пользователя из URL
	targetUserIDStr := c.Param("id")
	targetUserID, err := strconv.Atoi(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	// Получаем данные запрашивающего пользователя из контекста (установленные middleware)
	requestingUserIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}
	requestingUserID := requestingUserIDVal.(int)

	isAdminVal, _ := c.Get("isAdmin")
	isManagerVal, _ := c.Get("isManager")
	isAdmin := isAdminVal.(bool)
	isManager := isManagerVal.(bool)

	// Создаем "фиктивного" пользователя для передачи в сервис (только с нужными полями для проверки прав)
	requestingUser := &models.User{
		ID:        requestingUserID,
		IsAdmin:   isAdmin,
		IsManager: isManager,
	}

	// Привязываем данные из тела запроса к DTO
	var updateData models.UserUpdateDTO
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные для обновления: " + err.Error()})
		return
	}

	// Вызываем сервис для обновления профиля
	err = h.userService.UpdateUserProfile(requestingUser, targetUserID, &updateData)
	if err != nil {
		// Определяем тип ошибки и возвращаем соответствующий статус
		statusCode := http.StatusInternalServerError
		errMsg := "Ошибка обновления профиля: " + err.Error()

		if strings.Contains(err.Error(), "недостаточно прав") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "не предоставлены") || strings.Contains(err.Error(), "нет допустимых полей") {
			statusCode = http.StatusBadRequest
		}
		// Можно добавить обработку других специфических ошибок сервиса/репозитория

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	// Возвращаем успешный ответ
	// Можно вернуть обновленные данные пользователя, если сервис их возвращает,
	// или просто сообщение об успехе.
	c.JSON(http.StatusOK, gin.H{"message": "Профиль пользователя успешно обновлен"})
}
