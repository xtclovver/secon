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
	vacationService services.VacationServiceInterface           // Используем интерфейс
	userService     services.UserServiceInterface               // Добавляем сервис пользователей
	unitService     services.OrganizationalUnitServiceInterface // Добавляем сервис орг. юнитов
}

// NewAppHandler создает новый экземпляр AppHandler
func NewAppHandler(vs services.VacationServiceInterface, us services.UserServiceInterface, ous services.OrganizationalUnitServiceInterface) *AppHandler { // Добавлен ous
	return &AppHandler{
		vacationService: vs,
		userService:     us,  // Инициализируем userService
		unitService:     ous, // Инициализируем unitService
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

	// Проверяем, существует ли пользователь, для которого устанавливается лимит
	targetUser, errUser := h.userService.FindByID(input.UserID)
	if errUser != nil {
		// Ошибка при поиске пользователя (например, ошибка БД)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки пользователя: " + errUser.Error()})
		return
	}
	if targetUser == nil {
		// Пользователь не найден
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь с указанным ID не найден"})
		return
	}

	// Вызываем сервис для установки лимита
	err := h.vacationService.SetVacationLimit(input.UserID, input.Year, input.TotalDays)
	if err != nil {
		// Обрабатываем возможные ошибки сервиса (например, отрицательное количество дней или ошибка репозитория)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка установки лимита: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Лимит отпуска успешно установлен"})
}

// CreateVacationRequest обработчик для создания заявки на отпуск
func (h *AppHandler) CreateVacationRequest(c *gin.Context) {
	var request models.VacationRequest

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
	unitIDStr := c.Query("unitId") // departmentId -> unitId
	unitID, err := strconv.Atoi(unitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID орг. юнита"}) // Обновлено сообщение
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
	intersections, err := h.vacationService.CheckIntersections(unitID, year) // departmentID -> unitID
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

// GetOrganizationalUnitVacations обработчик для получения отпусков сотрудников орг. юнита
func (h *AppHandler) GetOrganizationalUnitVacations(c *gin.Context) { // GetDepartmentVacations -> GetOrganizationalUnitVacations
	// Права доступа пока оставим для менеджера или админа (админ получит доступ через middleware)
	isManagerVal, _ := c.Get("isManager")
	isAdminVal, _ := c.Get("isAdmin")
	isManager := isManagerVal.(bool)
	isAdmin := isAdminVal.(bool)

	if !isAdmin && !isManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права руководителя или администратора"})
		return
	}

	unitIDStr := c.Param("id") // departmentIDStr -> unitIDStr
	unitID, err := strconv.Atoi(unitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID орг. юнита"}) // Обновлено сообщение
		return
	}

	yearStr := c.Query("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат года"})
		return
	}

	// Получаем statusFilter из query параметров
	statusFilter := GetIntQueryParam(c, "status")

	// Получение заявок орг. юнита из сервиса
	vacations, err := h.vacationService.GetOrganizationalUnitVacations(unitID, year, statusFilter) // GetDepartmentVacations -> GetOrganizationalUnitVacations, departmentID -> unitID
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения заявок орг. юнита: " + err.Error()}) // Обновлено сообщение
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
	unitIDFilter := GetIntQueryParam(c, "unitId") // departmentId -> unitId

	// Если год не указан, можно использовать текущий или требовать параметр
	if yearFilter == nil {
		currentYear := time.Now().Year()
		yearFilter = &currentYear
		// Или вернуть ошибку:
		// c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр 'year' обязателен"})
		// return
	}

	// Вызываем сервис для получения всех заявок с учетом прав и фильтров
	vacations, err := h.vacationService.GetAllUserVacations(requestingUserID.(int), yearFilter, statusFilter, userIDFilter, unitIDFilter) // departmentIDFilter -> unitIDFilter
	if err != nil {
		// Обрабатываем ошибки прав доступа или другие ошибки сервиса
		if err.Error() == "недостаточно прав для просмотра всех заявок" || err.Error() == "менеджер может просматривать заявки только своего организационного юнита" { // Обновлено сообщение об ошибке
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

	// Получаем необязательный параметр ?force=true из query
	forceStr := c.Query("force")
	force := forceStr == "true"

	// Вызываем сервис для утверждения заявки с флагом force, получаем конфликты и ошибку
	conflicts, err := h.vacationService.ApproveVacationRequest(requestID, approverID.(int), force)
	if err != nil {
		// Обработка ошибок, возникших при попытке утверждения (права, статус, ошибка БД и т.д.)
		statusCode := http.StatusInternalServerError
		errMsg := "Ошибка утверждения заявки: " + err.Error()
		errStr := err.Error()
		if strings.Contains(errStr, "не найден") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(errStr, "не имеет прав") {
			statusCode = http.StatusForbidden
		} else if strings.Contains(errStr, "можно утвердить только заявку в статусе") {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	// Проверяем результат:
	if len(conflicts) > 0 && !force {
		// Конфликты найдены, и force был false. Заявка НЕ утверждена.
		// Возвращаем статус 409 Conflict со списком конфликтов.
		log.Printf("[Handler ApproveVacationRequest] Request %d approval blocked due to conflicts (force=false).", requestID)
		c.JSON(http.StatusConflict, gin.H{
			"error":     "Обнаружены конфликты с отпусками других сотрудников на той же должности.",
			"conflicts": conflicts,
		})
		return
	}

	// Если мы здесь, значит:
	// 1. Конфликтов не было ИЗНАЧАЛЬНО.
	// 2. Конфликты были, но force был true, и заявка была УТВЕРЖДЕНА (если не было других ошибок).
	// В обоих случаях возвращаем HTTP 200 OK.
	log.Printf("[Handler ApproveVacationRequest] Request %d approved successfully (force=%t, conflicts returned: %d).", requestID, force, len(conflicts))
	response := gin.H{"message": "Заявка успешно утверждена"}
	if len(conflicts) > 0 && force {
		// Если force был true и были конфликты, возвращаем их как предупреждение.
		response["warnings"] = conflicts
	}
	c.JSON(http.StatusOK, response)
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
		Login    string `json:"login" binding:"required"` // username -> login
		Password string `json:"password" binding:"required"`
	}

	// Привязываем JSON из тела запроса к структуре credentials
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные: " + err.Error()})
		return
	}

	// Вызываем сервис для проверки логина и пароля
	token, user, err := h.authService.Login(credentials.Login, credentials.Password) // credentials.Username -> credentials.Login
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
	// Структура для входящих данных (PascalCase как ожидает фронтенд/валидатор)
	// Username заменено на Login, убрана валидация email для Login
	var input struct {
		Login                string `json:"Login" binding:"required"` // username -> Login (PascalCase)
		Password             string `json:"Password" binding:"required"`
		ConfirmPassword      string `json:"ConfirmPassword" binding:"required"`
		FullName             string `json:"FullName" binding:"required"`
		Email                string `json:"Email" binding:"required"` // Оставляем Email, но без валидации email
		PositionID           *int   `json:"PositionID"`               // Оставляем PositionID в PascalCase
		OrganizationalUnitID *int   `json:"OrganizationalUnitID"`     // Добавлено поле для орг. юнита
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		// Возвращаем ошибку валидации Gin, которая уже включает детали по полям
		// Убрали предыдущую ошибку, так как ShouldBindJSON предоставляет лучшую информацию
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные", "details": err.Error()})
		return
	}

	// Проверка совпадения паролей
	if input.Password != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пароли не совпадают"})
		return
	}

	// Вызов сервиса регистрации - передаем input.Login как login и OrganizationalUnitID
	user, err := h.authService.Register(input.Login, input.Password, input.FullName, input.PositionID, input.OrganizationalUnitID) // Удален input.Email, Добавлен input.OrganizationalUnitID
	if err != nil {
		// Обработка ошибок сервиса (например, пользователь уже существует)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // Используем 409 Conflict для дубликата
		return
	}

	// Возвращаем созданного пользователя (без пароля)
	c.JSON(http.StatusCreated, user)
}

// GetPositions обработчик для получения списка должностей (публичный)
func (h *AppHandler) GetPositions(c *gin.Context) {
	positions, err := h.userService.GetAllPositions() // Используем userService
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения списка должностей: " + err.Error()})
		return
	}

	// Раньше возвращались группы, теперь просто список
	// Если фронтенд ожидает группы, нужно либо изменить фронтенд,
	// либо сгруппировать здесь (менее предпочтительно).
	// Пока возвращаем плоский список.
	c.JSON(http.StatusOK, positions)
}

// --- Organizational Unit Handlers ---

// CreateOrganizationalUnit обработчик для создания нового орг. юнита (Admin only)
func (h *AppHandler) CreateOrganizationalUnit(c *gin.Context) {
	var unit models.OrganizationalUnit
	if err := c.ShouldBindJSON(&unit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные: " + err.Error()})
		return
	}

	createdUnit, err := h.unitService.CreateUnit(&unit)
	if err != nil {
		// TODO: Различать ошибки валидации (400) и ошибки сервера (500)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания организационного юнита: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdUnit)
}

// GetOrganizationalUnitTree обработчик для получения дерева орг. юнитов
func (h *AppHandler) GetOrganizationalUnitTree(c *gin.Context) {
	tree, err := h.unitService.GetUnitTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения дерева организационных юнитов: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, tree)
}

// GetOrganizationalUnitByID обработчик для получения орг. юнита по ID
func (h *AppHandler) GetOrganizationalUnitByID(c *gin.Context) {
	unitIDStr := c.Param("id")
	unitID, err := strconv.Atoi(unitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID орг. юнита"})
		return
	}

	unit, err := h.unitService.GetUnitByID(unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения организационного юнита: " + err.Error()})
		return
	}
	if unit == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Организационный юнит не найден"})
		return
	}

	c.JSON(http.StatusOK, unit)
}

// UpdateOrganizationalUnit обработчик для обновления орг. юнита (Admin only)
func (h *AppHandler) UpdateOrganizationalUnit(c *gin.Context) {
	unitIDStr := c.Param("id")
	unitID, err := strconv.Atoi(unitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID орг. юнита"})
		return
	}

	var updateData models.OrganizationalUnit
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные для обновления: " + err.Error()})
		return
	}

	updatedUnit, err := h.unitService.UpdateUnit(unitID, &updateData)
	if err != nil {
		// TODO: Различать ошибки (не найдено 404, валидация 400, сервер 500)
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "не может быть родителем") || strings.Contains(err.Error(), "не найден") { // Уточнить ошибки валидации
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": "Ошибка обновления организационного юнита: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUnit)
}

// DeleteOrganizationalUnit обработчик для удаления орг. юнита (Admin only)
func (h *AppHandler) DeleteOrganizationalUnit(c *gin.Context) {
	unitIDStr := c.Param("id")
	unitID, err := strconv.Atoi(unitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID орг. юнита"})
		return
	}

	err = h.unitService.DeleteUnit(unitID)
	if err != nil {
		// TODO: Различать ошибки (не найдено 404, конфликт 409/400, сервер 500)
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound
		}
		// Добавить проверку на конфликт (например, если есть дочерние) -> 409 Conflict или 400 Bad Request
		c.JSON(statusCode, gin.H{"error": "Ошибка удаления организационного юнита: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Организационный юнит успешно удален"})
}

// GetUnitChildrenHandler обработчик для получения дочерних элементов (юнитов и пользователей)
func (h *AppHandler) GetUnitChildrenHandler(c *gin.Context) {
	parentIDStr := c.Query("parentId") // Получаем parentId из query-параметра

	var parentID *int
	if parentIDStr != "" {
		id, err := strconv.Atoi(parentIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный parentId"})
			return
		}
		parentID = &id
	} // Если parentIDStr пустой, parentID остается nil (для корневых элементов)

	// Вызываем новый метод сервиса
	items, err := h.unitService.GetUnitChildrenAndUsers(parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения дочерних элементов: " + err.Error()})
		return
	}

	// Возвращаем список элементов (может быть пустым)
	c.JSON(http.StatusOK, items)
}

// --- User Profile Handler ---

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

// GetMyProfile - обработчик для получения профиля текущего пользователя
func (h *AppHandler) GetMyProfile(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}
	userID := userIDVal.(int)

	// Вызываем сервис для получения профиля
	profile, err := h.userService.GetUserProfile(userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": "Ошибка получения профиля: " + err.Error()})
		return
	}

	// Возвращаем профиль
	c.JSON(http.StatusOK, profile)
}

// GetUnitUsersWithLimitsHandler обработчик для получения пользователей юнита с лимитами отпуска (Admin only)
func (h *AppHandler) GetUnitUsersWithLimitsHandler(c *gin.Context) {
	// Проверяем права администратора
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права администратора"})
		return
	}

	// Получаем ID юнита из URL (теперь параметр называется "id")
	unitIDStr := c.Param("id") // Изменено с "unitId" на "id"
	unitID, err := strconv.Atoi(unitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID орг. юнита"})
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
	usersWithLimits, err := h.unitService.GetUnitUsersWithLimits(unitID, year)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") { // Проверяем, если юнит не найден
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": "Ошибка получения пользователей юнита с лимитами: " + err.Error()})
		return
	}

	// Если лимиты не найдены для некоторых пользователей, поле TotalDays будет nil.
	// Здесь можно установить значение по умолчанию, если это требуется для фронтенда.
	// Например:
	// for i := range usersWithLimits {
	// 	if usersWithLimits[i].TotalDays == nil {
	// 		defaultValue := 28 // Значение по умолчанию
	// 		usersWithLimits[i].TotalDays = &defaultValue
	// 	}
	// }

	c.JSON(http.StatusOK, usersWithLimits)
}

// UpdateUserVacationLimitHandler обработчик для обновления лимита отпуска пользователя (Admin only)
func (h *AppHandler) UpdateUserVacationLimitHandler(c *gin.Context) {
	// Проверяем права администратора
	isAdmin, exists := c.Get("isAdmin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права администратора"})
		return
	}

	// Получаем ID пользователя из URL (исправлено с userId на id)
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя в URL"}) // Уточнено сообщение
		return
	}

	// Структура для данных из тела запроса
	var input struct {
		Year      int `json:"year" binding:"required"`
		TotalDays int `json:"total_days"` // Не используем binding:"required", чтобы позволить 0
	}

	// Привязываем JSON из тела запроса
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные: " + err.Error()})
		return
	}

	// Проверяем total_days на отрицательное значение (0 разрешен)
	if input.TotalDays < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Количество дней отпуска не может быть отрицательным"})
		return
	}

	// Проверяем, существует ли пользователь, для которого устанавливается лимит
	targetUser, errUser := h.userService.FindByID(userID)
	if errUser != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки пользователя: " + errUser.Error()})
		return
	}
	if targetUser == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь с указанным ID не найден"})
		return
	}

	// Вызываем сервис VacationService для установки (создания/обновления) лимита
	err = h.vacationService.SetVacationLimit(userID, input.Year, input.TotalDays)
	if err != nil {
		// Обрабатываем возможные ошибки сервиса (ошибка БД и т.д.)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка установки лимита отпуска: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Лимит отпуска пользователя успешно обновлен"})
}

// --- Dashboard Handler ---

// GetManagerDashboard обработчик для получения данных дашборда руководителя
func (h *AppHandler) GetManagerDashboard(c *gin.Context) {
	// Получаем ID руководителя из контекста
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}
	managerID := userIDVal.(int)

	// Проверяем, является ли пользователь руководителем (дополнительная проверка)
	isManagerVal, exists := c.Get("isManager")
	if !exists || !isManagerVal.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права руководителя"})
		return
	}

	// Вызываем сервис для получения данных дашборда
	dashboardData, err := h.vacationService.GetManagerDashboardData(managerID)
	if err != nil {
		// Обрабатываем возможные ошибки (например, пользователь не руководитель, ошибка БД)
		statusCode := http.StatusInternalServerError
		errMsg := "Ошибка получения данных дашборда: " + err.Error()
		if err.Error() == "пользователь не является руководителем или не привязан к юниту" {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "не найден") {
			// Если ошибка связана с ненайденным руководителем (хотя middleware должен был это проверить)
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	// Возвращаем данные дашборда
	c.JSON(http.StatusOK, dashboardData)
}

// GetVacationConflicts обработчик для получения конфликтов отпусков
func (h *AppHandler) GetVacationConflicts(c *gin.Context) {
	// Получаем ID пользователя из контекста
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}
	userID := userIDVal.(int)

	// Получаем startDate и endDate из query параметров
	// Ожидаемый формат: RFC3339 или YYYY-MM-DD
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	var startDate, endDate time.Time
	var err error

	// Парсим startDate
	if startDateStr != "" {
		// Пытаемся парсить как YYYY-MM-DD сначала
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			// Если не получилось, пробуем RFC3339
			startDate, err = time.Parse(time.RFC3339, startDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат startDate. Используйте YYYY-MM-DD или RFC3339."})
				return
			}
		}
		// Устанавливаем время на начало дня для startDate, если парсился только YYYY-MM-DD
		if len(startDateStr) == 10 { // Формат YYYY-MM-DD
			startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		}
	} else {
		// Если startDate не указан, можно использовать текущую дату или вернуть ошибку
		// startDate = time.Now()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр startDate обязателен"})
		return
	}

	// Парсим endDate
	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			endDate, err = time.Parse(time.RFC3339, endDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат endDate. Используйте YYYY-MM-DD или RFC3339."})
				return
			}
		}
		// Устанавливаем время на конец дня для endDate, если парсился только YYYY-MM-DD
		if len(endDateStr) == 10 { // Формат YYYY-MM-DD
			endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
		}
	} else {
		// Если endDate не указан, можно использовать startDate + N дней или вернуть ошибку
		// endDate = startDate.AddDate(0, 1, 0) // Например, +1 месяц
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр endDate обязателен"})
		return
	}

	// Проверяем, что startDate не позже endDate
	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate не может быть позже endDate"})
		return
	}

	// Вызываем сервисный метод
	conflicts, err := h.vacationService.GetVacationConflicts(userID, startDate, endDate)
	if err != nil {
		// Обрабатываем ошибки (пользователь не найден, ошибка получения юнитов/конфликтов)
		statusCode := http.StatusInternalServerError
		errMsg := "Ошибка получения конфликтов отпусков: " + err.Error()
		if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound // Или StatusUnauthorized, если пользователь из токена не найден
		}
		// Можно добавить другие проверки ошибок
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	// Возвращаем список конфликтов (может быть пустым)
	c.JSON(http.StatusOK, conflicts)
}

// --- Admin User Management Handlers ---

// GetAllUsersHandler обработчик для получения списка всех пользователей (Admin only)
func (h *AppHandler) GetAllUsersHandler(c *gin.Context) {
	// Проверяем права администратора
	isAdminVal, exists := c.Get("isAdmin")
	if !exists || !isAdminVal.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права администратора"})
		return
	}

	// Вызываем сервис для получения всех пользователей
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения списка пользователей: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// UpdateUserAdminHandler обработчик для обновления данных пользователя администратором (Admin only)
func (h *AppHandler) UpdateUserAdminHandler(c *gin.Context) {
	// Проверяем права администратора
	isAdminVal, exists := c.Get("isAdmin")
	if !exists || !isAdminVal.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права администратора"})
		return
	}

	// Получаем ID целевого пользователя из URL
	targetUserIDStr := c.Param("id")
	targetUserID, err := strconv.Atoi(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	// Получаем ID запрашивающего пользователя (админа) из контекста
	requestingUserIDVal, exists := c.Get("userID")
	if !exists {
		// Это не должно произойти, если middleware отработал, но проверим
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не удалось определить администратора"})
		return
	}
	requestingUserID := requestingUserIDVal.(int)

	// Создаем "фиктивного" пользователя для передачи в сервис
	requestingUser := &models.User{
		ID:      requestingUserID,
		IsAdmin: true, // Мы уже проверили права выше
	}

	// Привязываем данные из тела запроса к DTO
	var updateData models.UserUpdateAdminDTO
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные для обновления: " + err.Error()})
		return
	}

	// Вызываем сервис для обновления данных пользователя админом
	err = h.userService.UpdateUserAdmin(requestingUser, targetUserID, &updateData)
	if err != nil {
		// Определяем тип ошибки и возвращаем соответствующий статус
		statusCode := http.StatusInternalServerError
		errMsg := "Ошибка обновления пользователя: " + err.Error()

		if strings.Contains(err.Error(), "недостаточно прав") { // Хотя проверка уже есть выше
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "не найден") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "не предоставлены") || strings.Contains(err.Error(), "нет полей") {
			statusCode = http.StatusBadRequest
		}
		// TODO: Добавить обработку ошибок валидации ID юнита/должности, если сервис их возвращает

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Данные пользователя успешно обновлены администратором"})
}

// ExportVacationsByUnits обработчик для экспорта данных отпусков по юнитам (Admin only)
func (h *AppHandler) ExportVacationsByUnits(c *gin.Context) {
	// Проверяем права администратора
	isAdminVal, exists := c.Get("isAdmin")
	if !exists || !isAdminVal.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права администратора"})
		return
	}

	// Структура для данных из тела запроса
	var input struct {
		UnitIDs []int `json:"unit_ids" binding:"required"`
		Year    *int  `json:"year"` // Год опционален, можно использовать текущий по умолчанию
	}

	// Привязываем JSON из тела запроса
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные: " + err.Error()})
		return
	}

	// Проверяем, что список ID не пустой
	if len(input.UnitIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать хотя бы один ID организационного юнита"})
		return
	}

	// Определяем год для экспорта
	yearToExport := time.Now().Year() // По умолчанию текущий год
	if input.Year != nil {
		yearToExport = *input.Year
		// Можно добавить валидацию года, если нужно (например, не слишком старый/будущий)
	}

	// Вызываем сервис для получения данных для экспорта
	// TODO: Создать метод GetVacationDataForExport в vacationService
	exportData, err := h.vacationService.GetVacationDataForExport(input.UnitIDs, yearToExport)
	if err != nil {
		// Обрабатываем возможные ошибки сервиса (например, юнит не найден, ошибка БД)
		statusCode := http.StatusInternalServerError
		errMsg := "Ошибка получения данных для экспорта: " + err.Error()
		// TODO: Добавить обработку специфичных ошибок, если сервис их возвращает (e.g., StatusNotFound)
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	// Возвращаем данные для экспорта
	// Формат данных должен быть удобен для генерации XLSX на фронтенде
	// Например, массив объектов, где каждый объект - строка в таблице Т-7
	c.JSON(http.StatusOK, exportData)
}
