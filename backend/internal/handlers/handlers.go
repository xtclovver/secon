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

	// Получаем ID пользователя из URL
	userIDStr := c.Param("userId") // Используем userId из роута
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
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

	// Вызываем сервис VacationService для установки (создания/обновления) лимита
	err = h.vacationService.SetVacationLimit(userID, input.Year, input.TotalDays)
	if err != nil {
		// Обрабатываем возможные ошибки сервиса
		// (ошибка БД, но валидация на отрицательные дни уже сделана)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка установки лимита отпуска: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Лимит отпуска пользователя успешно обновлен"})
}
