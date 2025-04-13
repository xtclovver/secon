package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"log" // Убедимся, что log импортирован
	"strings"
	"time" // Убедимся, что time импортирован

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

	// --- Dashboard Data ---
	CountPendingRequestsByUnitIDs(unitIDs []int) (int, error)
	SumRequestedDaysByStatusAndUnitIDs(unitIDs []int, statusIDs []int, year int) (int, error) // Новый метод для суммирования дней
	GetUpcomingApprovedConflictsByUnitIDs(unitIDs []int, startDate time.Time, endDate time.Time) ([]models.ConflictingPeriod, error)
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

// --- Dashboard Data Methods ---

// CountPendingRequestsByUnitIDs подсчитывает заявки в статусе "На рассмотрении" для заданных юнитов
func (r *VacationRepository) CountPendingRequestsByUnitIDs(unitIDs []int) (int, error) {
	if len(unitIDs) == 0 {
		return 0, nil // Нет юнитов для поиска
	}

	query := `
		SELECT COUNT(vr.id)
		FROM vacation_requests vr
		JOIN users u ON vr.user_id = u.id
		WHERE vr.status_id = ?
		  AND u.organizational_unit_id IN (?` + sqlRepeatParams(len(unitIDs)-1) + `)`

	args := []interface{}{models.StatusPending}
	for _, id := range unitIDs {
		args = append(args, id)
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка подсчета ожидающих заявок по юнитам %v: %w", unitIDs, err)
	}
	return count, nil
}

// GetUpcomingApprovedConflictsByUnitIDs ищет утвержденные конфликты в заданном диапазоне дат для юнитов
func (r *VacationRepository) GetUpcomingApprovedConflictsByUnitIDs(unitIDs []int, startDate time.Time, endDate time.Time) ([]models.ConflictingPeriod, error) {
	if len(unitIDs) == 0 {
		return []models.ConflictingPeriod{}, nil
	}

	// 1. Найти все УТВЕРЖДЕННЫЕ периоды отпусков в ЗАДАННОМ ДИАПАЗОНЕ ДАТ для пользователей из ЗАДАННЫХ ЮНИТОВ
	queryPeriods := `
		SELECT
			vp.id as period_id, vp.request_id, vp.start_date, vp.end_date,
			vr.user_id, u.full_name, u.position_id
		FROM vacation_periods vp
		JOIN vacation_requests vr ON vp.request_id = vr.id
		JOIN users u ON vr.user_id = u.id
		WHERE vr.status_id = ? -- Только утвержденные
		  AND u.organizational_unit_id IN (?` + sqlRepeatParams(len(unitIDs)-1) + `)
		  AND vp.start_date <= ? -- Периоды, которые начинаются до конца диапазона
		  AND vp.end_date >= ?   -- Периоды, которые заканчиваются после начала диапазона
		ORDER BY u.position_id, vp.start_date
	`
	args := []interface{}{models.StatusApproved}
	for _, id := range unitIDs {
		args = append(args, id)
	}
	args = append(args, endDate, startDate)

	rows, err := r.db.Query(queryPeriods, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения утвержденных периодов для юнитов %v: %w", unitIDs, err)
	}
	defer rows.Close()

	type periodInfo struct {
		period     models.VacationPeriod
		userID     int
		userFName  string
		positionID *int // Используем указатель, так как должность может быть NULL
	}
	var approvedPeriods []periodInfo

	for rows.Next() {
		var pi periodInfo
		var posID sql.NullInt64 // Для сканирования position_id
		err := rows.Scan(
			&pi.period.ID, &pi.period.RequestID, &pi.period.StartDate, &pi.period.EndDate,
			&pi.userID, &pi.userFName, &posID,
		)
		if err != nil {
			log.Printf("Ошибка сканирования утвержденного периода при поиске конфликтов для дашборда: %v", err)
			continue
		}
		if posID.Valid {
			pInt := int(posID.Int64)
			pi.positionID = &pInt
		}
		approvedPeriods = append(approvedPeriods, pi)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по утвержденным периодам для дашборда: %w", err)
	}

	// 2. Найти пересечения между полученными периодами С УЧЕТОМ ДОЛЖНОСТИ
	var conflicts []models.ConflictingPeriod
	for i := 0; i < len(approvedPeriods); i++ {
		p1Info := approvedPeriods[i]
		if p1Info.positionID == nil {
			continue // Пропускаем, если должность не указана (не с кем сравнивать)
		}
		for j := i + 1; j < len(approvedPeriods); j++ {
			p2Info := approvedPeriods[j]

			// Проверяем, что пользователи разные, на одной должности и их периоды пересекаются
			if p1Info.userID != p2Info.userID &&
				p2Info.positionID != nil && // У второго тоже должна быть должность
				*p1Info.positionID == *p2Info.positionID && // Должности совпадают
				p1Info.period.StartDate.Time.Unix() <= p2Info.period.EndDate.Time.Unix() && // Периоды пересекаются
				p1Info.period.EndDate.Time.Unix() >= p2Info.period.StartDate.Time.Unix() {

				// Находим даты пересечения
				overlapStart := p1Info.period.StartDate.Time
				if p2Info.period.StartDate.Time.After(overlapStart) {
					overlapStart = p2Info.period.StartDate.Time
				}

				overlapEnd := p1Info.period.EndDate.Time
				if p2Info.period.EndDate.Time.Before(overlapEnd) {
					overlapEnd = p2Info.period.EndDate.Time
				}

				// Создаем конфликт (в обе стороны, но потом можно будет уникализировать, если надо)
				// Важно: Заполняем поля так, чтобы было понятно, кто с кем конфликтует
				conflict := models.ConflictingPeriod{
					ConflictingUserID:       p2Info.userID,
					ConflictingUserFullName: p2Info.userFName,
					ConflictingRequestID:    p2Info.period.RequestID,
					ConflictingPeriodID:     p2Info.period.ID,
					ConflictingStartDate:    p2Info.period.StartDate,
					ConflictingEndDate:      p2Info.period.EndDate,
					OriginalUserID:          p1Info.userID,           // <-- Добавлено ID первого пользователя
					OriginalUserFullName:    p1Info.userFName,        // <-- Добавлено ФИО первого пользователя
					OriginalRequestID:       p1Info.period.RequestID, // Используем поля p1 как "оригинальные" для этой записи
					OriginalPeriodID:        p1Info.period.ID,
					OriginalStartDate:       p1Info.period.StartDate,
					OriginalEndDate:         p1Info.period.EndDate,
					OverlapStartDate:        models.CustomDate{Time: overlapStart},
					OverlapEndDate:          models.CustomDate{Time: overlapEnd},
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	// TODO: Возможно, стоит добавить логику для удаления дублирующихся конфликтов (если A+B и B+A считаются одним и тем же)

	return conflicts, nil
}

// SumRequestedDaysByStatusAndUnitIDs суммирует поле days_requested для заявок с заданными статусами и юнитами за год
func (r *VacationRepository) SumRequestedDaysByStatusAndUnitIDs(unitIDs []int, statusIDs []int, year int) (int, error) {
	if len(unitIDs) == 0 || len(statusIDs) == 0 {
		return 0, nil // Нет юнитов или статусов для поиска
	}

	unitPlaceholders := sqlRepeatParams(len(unitIDs))
	statusPlaceholders := sqlRepeatParams(len(statusIDs))

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(vr.days_requested), 0)
		FROM vacation_requests vr
		JOIN users u ON vr.user_id = u.id
		WHERE vr.year = ?
		  AND vr.status_id IN (?%s)
		  AND u.organizational_unit_id IN (?%s)`, statusPlaceholders, unitPlaceholders)

	args := []interface{}{year}
	for _, sID := range statusIDs {
		args = append(args, sID)
	}
	for _, uID := range unitIDs {
		args = append(args, uID)
	}

	var totalDays int
	err := r.db.QueryRow(query, args...).Scan(&totalDays)
	if err != nil {
		// Если ошибок нет, но сумма NULL (нет заявок), COALESCE вернет 0, ошибки Scan не будет
		// Обрабатываем только реальные ошибки запроса/сканирования
		return 0, fmt.Errorf("ошибка суммирования запрошенных дней по статусам %v, юнитам %v, году %d: %w", statusIDs, unitIDs, year, err)
	}

	return totalDays, nil
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
