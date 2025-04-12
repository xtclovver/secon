package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"log" // Убедимся, что log импортирован
	"strings"

	"vacation-scheduler/internal/models"
)

// ErrLimitNotFound - Ошибка, возвращаемая, когда лимит отпуска не найден.
var ErrLimitNotFound = errors.New("лимит отпуска не найден для данного пользователя и года")

// statusIDToNameMap - Вспомогательная карта для получения имени статуса по ID
var statusIDToNameMap = map[int]string{
	models.StatusDraft:     "Черновик",
	models.StatusPending:   "На рассмотрении",
	models.StatusApproved:  "Утверждена",
	models.StatusRejected:  "Отклонена",
	models.StatusCancelled: "Отменена",
}

// VacationRepositoryInterface определяет методы для работы с данными отпусков.
// (Интерфейс перемещен сюда для лучшей читаемости)
type VacationRepositoryInterface interface {
	// --- Лимиты ---
	GetVacationLimit(userID int, year int) (*models.VacationLimit, error)
	CreateOrUpdateVacationLimit(userID int, year int, totalDays int) error
	UpdateVacationLimitUsedDays(userID int, year int, daysDelta int) error // Добавлен метод для изменения использованных дней

	// --- Заявки ---
	GetVacationRequestByID(requestID int) (*models.VacationRequest, error) // Добавлен метод получения заявки по ID
	SaveVacationRequest(request *models.VacationRequest) error
	UpdateVacationRequest(request *models.VacationRequest) error // Для обновления комментария и т.д. пользователем
	UpdateRequestStatusByID(requestID int, newStatusID int) error
	GetVacationRequestsByUser(userID int, year int, statusFilter *int) ([]models.VacationRequest, error)
	GetVacationRequestsByOrganizationalUnit(unitID int, year int, statusFilter *int) ([]models.VacationRequest, error)
	// Изменен тип unitIDsFilter на []int
	GetAllVacationRequests(yearFilter *int, statusFilter *int, userIDFilter *int, unitIDsFilter []int) ([]models.VacationRequestAdminView, error)

	// --- Уведомления ---
	CreateNotification(notification *models.Notification) error

	// --- Периоды (при необходимости) ---
	// GetPeriodsByRequestID(requestID int) ([]models.VacationPeriod, error) // Пример
	// DeletePeriodsByRequestID(requestID int) error // Пример
}

// VacationRepository предоставляет методы для работы с данными отпусков в БД
type VacationRepository struct {
	db *sql.DB
}

// NewVacationRepository создает новый экземпляр VacationRepository
func NewVacationRepository(db *sql.DB) *VacationRepository {
	return &VacationRepository{db: db}
}

// --- Лимиты ---

// GetVacationLimit получает лимит отпуска для пользователя на указанный год
func (r *VacationRepository) GetVacationLimit(userID int, year int) (*models.VacationLimit, error) {
	query := `
		SELECT id, user_id, year, total_days, used_days, created_at, updated_at
		FROM vacation_limits
		WHERE user_id = ? AND year = ?`

	row := r.db.QueryRow(query, userID, year)
	limit := &models.VacationLimit{}

	err := row.Scan(
		&limit.ID, &limit.UserID, &limit.Year, &limit.TotalDays,
		&limit.UsedDays, &limit.CreatedAt, &limit.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Возвращаем стандартизированную ошибку
			return nil, ErrLimitNotFound
		}
		return nil, fmt.Errorf("ошибка получения лимита отпуска из БД: %w", err)
	}
	return limit, nil
}

// CreateOrUpdateVacationLimit создает или обновляет лимит отпуска для пользователя
func (r *VacationRepository) CreateOrUpdateVacationLimit(userID int, year int, totalDays int) error {
	query := `
		INSERT INTO vacation_limits (user_id, year, total_days, used_days, created_at, updated_at)
		VALUES (?, ?, ?, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON DUPLICATE KEY UPDATE
			total_days = VALUES(total_days),
			used_days = 0,
			updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query, userID, year, totalDays)
	if err != nil {
		return fmt.Errorf("ошибка создания/обновления лимита отпуска: %w", err)
	}
	return nil
}

// UpdateVacationLimitUsedDays атомарно обновляет использованные дни для СУЩЕСТВУЮЩЕГО лимита.
func (r *VacationRepository) UpdateVacationLimitUsedDays(userID int, year int, daysDelta int) error {
	log.Printf("[Repo UpdateUsedDays] Attempting update. UserID: %d, Year: %d, Delta: %d", userID, year, daysDelta) // LOGGING
	query := `
		UPDATE vacation_limits
		SET
			used_days = GREATEST(0, used_days + ?), -- Используем параметр (?) для UPDATE
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND year = ?`

	result, err := r.db.Exec(query, daysDelta, userID, year)
	if err != nil {
		log.Printf("[Repo UpdateUsedDays] DB Exec Error. UserID: %d, Year: %d, Delta: %d, Error: %v", userID, year, daysDelta, err) // LOGGING
		return fmt.Errorf("ошибка обновления used_days (user: %d, year: %d, delta: %d): %w", userID, year, daysDelta, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[Repo UpdateUsedDays] Error getting rows affected. UserID: %d, Year: %d, Delta: %d, Error: %v", userID, year, daysDelta, err) // LOGGING
		return fmt.Errorf("ошибка получения кол-ва строк при обновлении used_days (user: %d, year: %d, delta: %d): %w", userID, year, daysDelta, err)
	}

	if rowsAffected == 0 {
		log.Printf("[Repo UpdateUsedDays] Update failed - Limit not found. UserID: %d, Year: %d", userID, year) // LOGGING
		return errors.New("лимит отпуска не найден для данного пользователя и года")
	}
	log.Printf("[Repo UpdateUsedDays] DB Exec Success. UserID: %d, Year: %d, Delta: %d", userID, year, daysDelta) // LOGGING

	// READ-AFTER-WRITE CHECK
	var currentUsedDays int
	checkQuery := "SELECT used_days FROM vacation_limits WHERE user_id = ? AND year = ?"
	checkErr := r.db.QueryRow(checkQuery, userID, year).Scan(&currentUsedDays)
	if checkErr != nil {
		log.Printf("[Repo UpdateUsedDays] Read-after-write CHECK FAILED. UserID: %d, Year: %d, Error: %v", userID, year, checkErr)
	} else {
		log.Printf("[Repo UpdateUsedDays] Read-after-write CHECK PASSED. UserID: %d, Year: %d, CurrentUsedDaysInDB: %d", userID, year, currentUsedDays)
	}

	return nil // Успешно
}

// --- Заявки ---

// SaveVacationRequest сохраняет новую заявку на отпуск и ее периоды в транзакции
func (r *VacationRepository) SaveVacationRequest(request *models.VacationRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	var txErr error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if txErr != nil {
			log.Printf("Rolling back transaction due to error: %v", txErr)
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error during transaction rollback: %v", rbErr)
			}
		} else {
			txErr = tx.Commit()
			if txErr != nil {
				log.Printf("Ошибка коммита транзакции сохранения заявки: %v", txErr)
			}
		}
	}()

	if request.DaysRequested == 0 {
		for _, p := range request.Periods {
			request.DaysRequested += p.DaysCount
		}
	}

	queryReq := `INSERT INTO vacation_requests (user_id, year, status_id, days_requested, comment, created_at, updated_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	result, errExec := tx.Exec(queryReq, request.UserID, request.Year, request.StatusID, request.DaysRequested, request.Comment)
	if errExec != nil {
		txErr = fmt.Errorf("ошибка сохранения заявки: %w", errExec)
		return txErr
	}
	requestID, errID := result.LastInsertId()
	if errID != nil {
		txErr = fmt.Errorf("ошибка получения ID сохраненной заявки: %w", errID)
		return txErr
	}
	request.ID = int(requestID)

	if len(request.Periods) > 0 {
		queryPeriod := `INSERT INTO vacation_periods (request_id, start_date, end_date, days_count, created_at, updated_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
		stmt, errPrepare := tx.Prepare(queryPeriod)
		if errPrepare != nil {
			txErr = fmt.Errorf("ошибка подготовки запроса для периодов: %w", errPrepare)
			return txErr
		}
		defer stmt.Close()
		for i := range request.Periods {
			if request.Periods[i].StartDate.IsZero() || request.Periods[i].EndDate.IsZero() || request.Periods[i].StartDate.Time.After(request.Periods[i].EndDate.Time) {
				txErr = fmt.Errorf("некорректные даты в периоде %d", i+1)
				return txErr
			}
			_, errStmtExec := stmt.Exec(request.ID, request.Periods[i].StartDate, request.Periods[i].EndDate, request.Periods[i].DaysCount)
			if errStmtExec != nil {
				txErr = fmt.Errorf("ошибка сохранения периода %d: %w", i+1, errStmtExec)
				return txErr
			}
		}
	}
	return txErr
}

// UpdateVacationRequest обновляет существующую заявку (комментарий) пользователем
func (r *VacationRepository) UpdateVacationRequest(request *models.VacationRequest) error {
	query := `UPDATE vacation_requests SET comment = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
	result, err := r.db.Exec(query, request.Comment, request.ID, request.UserID)
	if err != nil {
		return fmt.Errorf("ошибка обновления комментария заявки: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк при обновлении комментария: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("заявка для обновления комментария не найдена или не принадлежит пользователю")
	}
	return nil
}

// UpdateRequestStatusByID обновляет только статус заявки по ее ID
func (r *VacationRepository) UpdateRequestStatusByID(requestID int, newStatusID int) error {
	query := `UPDATE vacation_requests SET status_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := r.db.Exec(query, newStatusID, requestID)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса заявки: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк при смене статуса: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("заявка для обновления статуса не найдена")
	}
	return nil
}

// GetVacationRequestByID получает одну заявку по ее ID вместе с периодами
func (r *VacationRepository) GetVacationRequestByID(requestID int) (*models.VacationRequest, error) {
	queryRequest := `SELECT id, user_id, year, status_id, days_requested, comment, created_at, updated_at FROM vacation_requests WHERE id = ?`
	row := r.db.QueryRow(queryRequest, requestID)
	var req models.VacationRequest
	var comment sql.NullString
	err := row.Scan(&req.ID, &req.UserID, &req.Year, &req.StatusID, &req.DaysRequested, &comment, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка сканирования заявки по ID: %w", err)
	}
	if comment.Valid {
		req.Comment = comment.String
	}
	req.Periods, err = r.getPeriodsByRequestID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения периодов для заявки %d: %w", req.ID, err)
	}
	return &req, nil
}

// getPeriodsByRequestID - вспомогательный метод для получения периодов заявки
func (r *VacationRepository) getPeriodsByRequestID(requestID int) ([]models.VacationPeriod, error) {
	queryPeriods := `SELECT id, request_id, start_date, end_date, days_count, created_at, updated_at FROM vacation_periods WHERE request_id = ?`
	rows, err := r.db.Query(queryPeriods, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var periods []models.VacationPeriod
	for rows.Next() {
		var period models.VacationPeriod
		if err := rows.Scan(&period.ID, &period.RequestID, &period.StartDate, &period.EndDate, &period.DaysCount, &period.CreatedAt, &period.UpdatedAt); err != nil {
			log.Printf("Ошибка сканирования периода для заявки %d: %v\n", requestID, err)
			continue
		}
		periods = append(periods, period)
	}
	if err = rows.Err(); err != nil {
		return periods, err
	}
	return periods, nil
}

// GetVacationRequestsByUser получает заявки пользователя с фильтрацией по статусу
func (r *VacationRepository) GetVacationRequestsByUser(userID int, year int, statusFilter *int) ([]models.VacationRequest, error) {
	baseQuery := `SELECT id, user_id, year, status_id, days_requested, comment, created_at, updated_at FROM vacation_requests WHERE user_id = ? AND year = ?`
	args := []interface{}{userID, year}
	if statusFilter != nil {
		baseQuery += " AND status_id = ?"
		args = append(args, *statusFilter)
	}
	baseQuery += " ORDER BY created_at DESC"
	rowsReq, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса заявок пользователя %d: %w", userID, err)
	}
	defer rowsReq.Close()
	requestsMap := make(map[int]*models.VacationRequest)
	var requestIDs []interface{}
	for rowsReq.Next() {
		var req models.VacationRequest
		var comment sql.NullString
		if err := rowsReq.Scan(&req.ID, &req.UserID, &req.Year, &req.StatusID, &req.DaysRequested, &comment, &req.CreatedAt, &req.UpdatedAt); err != nil {
			log.Printf("Ошибка сканирования заявки пользователя %d: %v\n", userID, err)
			continue
		}
		if comment.Valid {
			req.Comment = comment.String
		}
		req.Periods = []models.VacationPeriod{}
		requestsMap[req.ID] = &req
		requestIDs = append(requestIDs, req.ID)
	}
	if err = rowsReq.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по заявкам пользователя %d: %w", userID, err)
	}
	if len(requestIDs) == 0 {
		return []models.VacationRequest{}, nil
	}
	periods, err := r.getPeriodsByRequestIDs(requestIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения периодов для заявок пользователя %d: %w", userID, err)
	}
	for _, period := range periods {
		if req, ok := requestsMap[period.RequestID]; ok {
			req.Periods = append(req.Periods, period)
		}
	}
	result := make([]models.VacationRequest, 0, len(requestsMap))
	for _, req := range requestsMap {
		result = append(result, *req)
	}
	return result, nil
}

// GetVacationRequestsByOrganizationalUnit получает заявки орг. юнита с фильтрацией по статусу
func (r *VacationRepository) GetVacationRequestsByOrganizationalUnit(unitID int, year int, statusFilter *int) ([]models.VacationRequest, error) {
	baseQuery := `SELECT vr.id, vr.user_id, vr.year, vr.status_id, vr.days_requested, vr.comment, vr.created_at, vr.updated_at FROM vacation_requests vr JOIN users u ON vr.user_id = u.id WHERE u.organizational_unit_id = ? AND vr.year = ?`
	args := []interface{}{unitID, year}
	if statusFilter != nil {
		baseQuery += " AND vr.status_id = ?"
		args = append(args, *statusFilter)
	}
	baseQuery += " ORDER BY vr.created_at DESC"
	rowsReq, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса заявок орг. юнита %d: %w", unitID, err)
	}
	defer rowsReq.Close()
	requestsMap := make(map[int]*models.VacationRequest)
	var requestIDs []interface{}
	for rowsReq.Next() {
		var req models.VacationRequest
		var comment sql.NullString
		if err := rowsReq.Scan(&req.ID, &req.UserID, &req.Year, &req.StatusID, &req.DaysRequested, &comment, &req.CreatedAt, &req.UpdatedAt); err != nil {
			log.Printf("Ошибка сканирования заявки орг. юнита %d: %v\n", unitID, err) // Исправлено departmentID -> unitID
			continue
		}
		if comment.Valid {
			req.Comment = comment.String
		}
		req.Periods = []models.VacationPeriod{}
		requestsMap[req.ID] = &req
		requestIDs = append(requestIDs, req.ID)
	}
	if err = rowsReq.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по заявкам орг. юнита %d: %w", unitID, err)
	}
	if len(requestIDs) == 0 {
		return []models.VacationRequest{}, nil
	}
	periods, err := r.getPeriodsByRequestIDs(requestIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения периодов для заявок орг. юнита %d: %w", unitID, err)
	}
	for _, period := range periods {
		if req, ok := requestsMap[period.RequestID]; ok {
			req.Periods = append(req.Periods, period)
		}
	}
	result := make([]models.VacationRequest, 0, len(requestsMap))
	for _, req := range requestsMap {
		result = append(result, *req)
	}
	return result, nil
}

// GetAllVacationRequests получает все заявки для админов/менеджеров с фильтрами (использует срез unitIDs)
func (r *VacationRepository) GetAllVacationRequests(yearFilter *int, statusFilter *int, userIDFilter *int, unitIDsFilter []int) ([]models.VacationRequestAdminView, error) { // unitIDFilter *int -> unitIDsFilter []int
	queryBase := `
		SELECT
			vr.id, vr.user_id, vr.year, vr.status_id, vr.days_requested, vr.comment, vr.created_at, vr.updated_at,
			u.full_name
		FROM vacation_requests vr
		JOIN users u ON vr.user_id = u.id`

	conditions := []string{}
	args := []interface{}{}

	if yearFilter != nil {
		conditions = append(conditions, "vr.year = ?")
		args = append(args, *yearFilter)
	}
	if statusFilter != nil {
		conditions = append(conditions, "vr.status_id = ?")
		args = append(args, *statusFilter)
	}
	if userIDFilter != nil {
		conditions = append(conditions, "vr.user_id = ?")
		args = append(args, *userIDFilter)
	}
	// Изменено: фильтр по срезу ID юнитов
	if len(unitIDsFilter) > 0 {
		// Генерируем плейсхолдеры (?, ?, ...)
		placeholders := sqlRepeatParams(len(unitIDsFilter))
		conditions = append(conditions, fmt.Sprintf("u.organizational_unit_id IN (?%s)", placeholders))
		// Добавляем ID юнитов в аргументы
		for _, id := range unitIDsFilter {
			args = append(args, id)
		}
	}

	query := queryBase
	if len(conditions) > 0 {
		query += " WHERE " + sqlJoinStrings(conditions, " AND ")
	}
	query += " ORDER BY vr.created_at DESC"

	rowsReq, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса всех заявок: %w", err)
	}
	defer rowsReq.Close()

	requestsMap := make(map[int]*models.VacationRequestAdminView)
	var requestIDs []interface{}

	for rowsReq.Next() {
		var req models.VacationRequestAdminView
		var comment sql.NullString
		err := rowsReq.Scan(&req.ID, &req.UserID, &req.Year, &req.StatusID, &req.DaysRequested, &comment, &req.CreatedAt, &req.UpdatedAt, &req.UserFullName)
		if err != nil {
			log.Printf("Ошибка сканирования AdminView заявки: %v\n", err)
			continue
		}
		if comment.Valid {
			req.Comment = comment.String
		}
		if name, ok := statusIDToNameMap[req.StatusID]; ok {
			req.StatusName = name
		} else {
			req.StatusName = "Неизвестно"
		}
		req.Periods = []models.VacationPeriod{}
		requestsMap[req.ID] = &req
		requestIDs = append(requestIDs, req.ID)
	}
	if err = rowsReq.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по всем заявкам: %w", err)
	}

	if len(requestIDs) == 0 {
		return []models.VacationRequestAdminView{}, nil
	}

	periods, err := r.getPeriodsByRequestIDs(requestIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения периодов для всех заявок: %w", err)
	}

	totalDaysMap := make(map[int]int)
	for _, period := range periods {
		if req, ok := requestsMap[period.RequestID]; ok {
			req.Periods = append(req.Periods, period)
			totalDaysMap[period.RequestID] += period.DaysCount
		}
	}

	result := make([]models.VacationRequestAdminView, 0, len(requestsMap))
	for _, req := range requestsMap {
		req.TotalDays = totalDaysMap[req.ID]
		if req.DaysRequested != req.TotalDays {
			log.Printf("Warning: DaysRequested (%d) in DB differs from calculated TotalDays (%d) for request ID %d", req.DaysRequested, req.TotalDays, req.ID)
		}
		result = append(result, *req)
	}

	return result, nil
}

// getPeriodsByRequestIDs - вспомогательный метод для получения периодов для списка ID заявок
func (r *VacationRepository) getPeriodsByRequestIDs(requestIDs []interface{}) ([]models.VacationPeriod, error) {
	if len(requestIDs) == 0 {
		return []models.VacationPeriod{}, nil
	}
	query := fmt.Sprintf(`SELECT id, request_id, start_date, end_date, days_count, created_at, updated_at FROM vacation_periods WHERE request_id IN (?%s)`, sqlRepeatParams(len(requestIDs)-1))
	rows, err := r.db.Query(query, requestIDs...)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса периодов по IDs: %w", err)
	}
	defer rows.Close()
	var periods []models.VacationPeriod
	for rows.Next() {
		var period models.VacationPeriod
		if err := rows.Scan(&period.ID, &period.RequestID, &period.StartDate, &period.EndDate, &period.DaysCount, &period.CreatedAt, &period.UpdatedAt); err != nil {
			log.Printf("Ошибка сканирования периода (множественный запрос): %v\n", err)
			continue
		}
		periods = append(periods, period)
	}
	if err = rows.Err(); err != nil {
		return periods, fmt.Errorf("ошибка итерации по периодам: %w", err)
	}
	return periods, nil
}

// --- Уведомления ---

// CreateNotification создает новое уведомление
func (r *VacationRepository) CreateNotification(notification *models.Notification) error {
	query := `INSERT INTO notifications (user_id, title, message, is_read, created_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`
	_, err := r.db.Exec(query, notification.UserID, notification.Title, notification.Message, notification.IsRead)
	if err != nil {
		return fmt.Errorf("ошибка создания уведомления: %w", err)
	}
	return nil
}

// --- Новые методы для проверки конфликтов ---

// GetUserPositionByID получает ID должности пользователя. Возвращает nil, если должность не установлена.
func (r *VacationRepository) GetUserPositionByID(userID int) (*int, error) {
	query := `SELECT position_id FROM users WHERE id = ?`
	var positionID sql.NullInt64 // Используем sql.NullInt64 для обработки NULL

	err := r.db.QueryRow(query, userID).Scan(&positionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пользователь с ID %d не найден", userID)
		}
		return nil, fmt.Errorf("ошибка получения position_id для пользователя %d: %w", userID, err)
	}

	if !positionID.Valid {
		return nil, nil // Должность не установлена (NULL)
	}

	posID := int(positionID.Int64)
	return &posID, nil
}

// GetApprovedVacationConflictsByPosition ищет конфликты утвержденных отпусков по должности
// periodsToCheck: периоды заявки, которую пытаются утвердить
func (r *VacationRepository) GetApprovedVacationConflictsByPosition(positionID int, excludeUserID int, periodsToCheck []models.VacationPeriod) ([]models.ConflictingPeriod, error) {
	if len(periodsToCheck) == 0 {
		return []models.ConflictingPeriod{}, nil // Нет периодов для проверки
	}

	var conflicts []models.ConflictingPeriod

	// Собираем условия WHERE для дат для всех проверяемых периодов
	dateConditions := []string{}
	dateArgs := []interface{}{}
	for _, p := range periodsToCheck {
		// Ищем существующие периоды, которые ПЕРЕСЕКАЮТСЯ с проверяемым периодом p
		// Пересечение: (Existing.Start <= p.End) AND (Existing.End >= p.Start)
		dateConditions = append(dateConditions, "(vp.start_date <= ? AND vp.end_date >= ?)")
		dateArgs = append(dateArgs, p.EndDate, p.StartDate)
	}
	dateConditionString := strings.Join(dateConditions, " OR ")

	// Основной запрос
	query := `
			SELECT
				u.id as conflicting_user_id,
				u.full_name as conflicting_user_full_name,
				vr.id as conflicting_request_id,
				vp.id as conflicting_period_id,
				vp.start_date as conflicting_start_date,
				vp.end_date as conflicting_end_date
			FROM vacation_periods vp
			JOIN vacation_requests vr ON vp.request_id = vr.id
			JOIN users u ON vr.user_id = u.id
			WHERE vr.status_id = ? -- Только утвержденные
			  AND u.position_id = ? -- Та же должность
			  AND vr.user_id != ?   -- Кроме самого пользователя
			  AND (` + dateConditionString + `) -- Пересечение дат
		`

	// Собираем аргументы: StatusApproved, positionID, excludeUserID, затем все start/end даты из dateArgs
	args := []interface{}{models.StatusApproved, positionID, excludeUserID}
	args = append(args, dateArgs...)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска конфликтующих отпусков: %w", err)
	}
	defer rows.Close()

	foundConflicts := make(map[int]models.ConflictingPeriod) // Используем map для временного хранения, ключ - ID конфликтного периода

	for rows.Next() {
		var conflict models.ConflictingPeriod
		err := rows.Scan(
			&conflict.ConflictingUserID,
			&conflict.ConflictingUserFullName,
			&conflict.ConflictingRequestID,
			&conflict.ConflictingPeriodID,
			&conflict.ConflictingStartDate,
			&conflict.ConflictingEndDate,
		)
		if err != nil {
			log.Printf("Ошибка сканирования конфликтующего периода: %v", err)
			continue // Пропускаем эту строку, но продолжаем обработку
		}
		foundConflicts[conflict.ConflictingPeriodID] = conflict // Сохраняем найденный конфликтный период
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по конфликтующим периодам: %w", err)
	}

	// Теперь для каждого найденного конфликтующего периода нужно определить,
	// с каким из ИСХОДНЫХ периодов (periodsToCheck) он пересекается и вычислить даты пересечения.
	for _, conflictingPeriod := range foundConflicts {
		for _, originalPeriod := range periodsToCheck {
			// Проверяем пересечение еще раз (хотя запрос уже должен был отфильтровать)
			// Условие пересечения: StartA <= EndB AND EndA >= StartB
			if conflictingPeriod.ConflictingStartDate.Time.Unix() <= originalPeriod.EndDate.Time.Unix() &&
				conflictingPeriod.ConflictingEndDate.Time.Unix() >= originalPeriod.StartDate.Time.Unix() {

				// Находим даты пересечения
				overlapStart := conflictingPeriod.ConflictingStartDate.Time
				if originalPeriod.StartDate.Time.After(overlapStart) {
					overlapStart = originalPeriod.StartDate.Time
				}

				overlapEnd := conflictingPeriod.ConflictingEndDate.Time
				if originalPeriod.EndDate.Time.Before(overlapEnd) {
					overlapEnd = originalPeriod.EndDate.Time
				}

				// Создаем и добавляем детализированный конфликт
				detailedConflict := conflictingPeriod                         // Копируем базовую информацию
				detailedConflict.OriginalRequestID = originalPeriod.RequestID // Заполняем детали исходного периода
				detailedConflict.OriginalPeriodID = originalPeriod.ID
				detailedConflict.OriginalStartDate = originalPeriod.StartDate
				detailedConflict.OriginalEndDate = originalPeriod.EndDate
				detailedConflict.OverlapStartDate = models.CustomDate{Time: overlapStart} // Заполняем даты пересечения
				detailedConflict.OverlapEndDate = models.CustomDate{Time: overlapEnd}

				conflicts = append(conflicts, detailedConflict)
				// Если один конфликтный период пересекается с несколькими исходными,
				// будет создано несколько записей в `conflicts`. Это нормально.
			}
		}
	}

	return conflicts, nil
}

// --- Вспомогательные функции ---

// sqlRepeatParams генерирует строку плейсхолдеров (?, ?, ...)
// count - количество плейсхолдеров ПОСЛЕ первого (т.е. для n параметров нужно передать n-1)
func sqlRepeatParams(count int) string {
	if count < 0 { // Исправлено на < 0, так как 0 значит один параметр уже есть
		return ""
	}
	return strings.Repeat(", ?", count)
}

// sqlJoinStrings соединяет строки с разделителем
func sqlJoinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}
