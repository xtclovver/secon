package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings" // Убедимся, что strings импортирован

	"vacation-scheduler/internal/models"
)

// statusIDToNameMap - Вспомогательная карта для получения имени статуса по ID
var statusIDToNameMap = map[int]string{
	models.StatusDraft:     "Черновик",
	models.StatusPending:   "На рассмотрении",
	models.StatusApproved:  "Утверждена",
	models.StatusRejected:  "Отклонена",
	models.StatusCancelled: "Отменена",
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
			return nil, errors.New("лимит отпуска не найден для данного пользователя и года")
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

// UpdateVacationLimitUsedDays обновляет количество использованных дней отпуска для пользователя
func (r *VacationRepository) UpdateVacationLimitUsedDays(userID int, year int, daysDelta int) error {
	query := `
		UPDATE vacation_limits 
		SET used_days = used_days + ?, updated_at = CURRENT_TIMESTAMP 
		WHERE user_id = ? AND year = ?`

	result, err := r.db.Exec(query, daysDelta, userID, year)
	if err != nil {
		return fmt.Errorf("ошибка обновления использованных дней лимита: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк при изменении лимита: %w", err)
	}
	if rowsAffected == 0 {
		// Вместо ошибки, можно просто ничего не делать, если лимита нет
		// Или создать лимит по умолчанию? Пока возвращаем ошибку, но можно изменить логику.
		return fmt.Errorf("лимит отпуска для пользователя %d на год %d не найден для обновления", userID, year)
	}
	return nil
}

// --- Заявки ---

// SaveVacationRequest сохраняет новую заявку на отпуск и ее периоды в транзакции
func (r *VacationRepository) SaveVacationRequest(request *models.VacationRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				fmt.Printf("Ошибка коммита транзакции сохранения заявки: %v\n", err)
			}
		}
	}()

	queryReq := `
		INSERT INTO vacation_requests (user_id, year, status_id, comment, created_at, updated_at) 
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	result, err := tx.Exec(queryReq, request.UserID, request.Year, request.StatusID, request.Comment)
	if err != nil {
		return fmt.Errorf("ошибка сохранения заявки: %w", err)
	}
	requestID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("ошибка получения ID сохраненной заявки: %w", err)
	}
	request.ID = int(requestID)

	if len(request.Periods) > 0 {
		queryPeriod := `
			INSERT INTO vacation_periods (request_id, start_date, end_date, days_count, created_at, updated_at) 
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
		stmt, err := tx.Prepare(queryPeriod)
		if err != nil {
			return fmt.Errorf("ошибка подготовки запроса для периодов: %w", err)
		}
		defer stmt.Close()

		for i := range request.Periods {
			if request.Periods[i].StartDate.IsZero() || request.Periods[i].EndDate.IsZero() || request.Periods[i].StartDate.Time.After(request.Periods[i].EndDate.Time) {
				return fmt.Errorf("некорректные даты в периоде %d", i+1)
			}
			_, err := stmt.Exec(request.ID, request.Periods[i].StartDate, request.Periods[i].EndDate, request.Periods[i].DaysCount)
			if err != nil {
				return fmt.Errorf("ошибка сохранения периода %d: %w", i+1, err)
			}
		}
	}

	return err
}

// UpdateVacationRequest обновляет существующую заявку (комментарий) пользователем
func (r *VacationRepository) UpdateVacationRequest(request *models.VacationRequest) error {
	query := `
		UPDATE vacation_requests 
		SET comment = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ? AND user_id = ?`

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
	query := `
		UPDATE vacation_requests 
		SET status_id = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`

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
	queryRequest := `
		SELECT id, user_id, year, status_id, comment, created_at, updated_at
		FROM vacation_requests 
		WHERE id = ?`
	row := r.db.QueryRow(queryRequest, requestID)

	var req models.VacationRequest
	var comment sql.NullString
	err := row.Scan(
		&req.ID, &req.UserID, &req.Year, &req.StatusID,
		&comment,
		&req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is not an error here
		}
		return nil, fmt.Errorf("ошибка сканирования заявки по ID: %w", err)
	}
	if comment.Valid {
		req.Comment = comment.String
	}

	req.Periods, err = r.getPeriodsByRequestID(req.ID)
	if err != nil {
		return &req, fmt.Errorf("ошибка получения периодов для заявки %d: %w", req.ID, err)
	}

	return &req, nil
}

// getPeriodsByRequestID - вспомогательный метод для получения периодов заявки
func (r *VacationRepository) getPeriodsByRequestID(requestID int) ([]models.VacationPeriod, error) {
	queryPeriods := `
		SELECT id, request_id, start_date, end_date, days_count, created_at, updated_at
		FROM vacation_periods
		WHERE request_id = ?`
	rows, err := r.db.Query(queryPeriods, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var periods []models.VacationPeriod
	for rows.Next() {
		var period models.VacationPeriod
		err := rows.Scan(
			&period.ID, &period.RequestID, &period.StartDate, &period.EndDate,
			&period.DaysCount, &period.CreatedAt, &period.UpdatedAt,
		)
		if err != nil {
			fmt.Printf("Ошибка сканирования периода для заявки %d: %v\n", requestID, err)
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
	baseQuery := `
		SELECT id, user_id, year, status_id, comment, created_at, updated_at
		FROM vacation_requests 
		WHERE user_id = ? AND year = ?`
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
		err := rowsReq.Scan(
			&req.ID, &req.UserID, &req.Year, &req.StatusID,
			&comment,
			&req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			fmt.Printf("Ошибка сканирования заявки пользователя %d: %v\n", userID, err)
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


// GetVacationRequestsByDepartment получает заявки подразделения с фильтрацией по статусу
func (r *VacationRepository) GetVacationRequestsByDepartment(departmentID int, year int, statusFilter *int) ([]models.VacationRequest, error) {
	baseQuery := `
		SELECT vr.id, vr.user_id, vr.year, vr.status_id, vr.comment, vr.created_at, vr.updated_at
		FROM vacation_requests vr
		JOIN users u ON vr.user_id = u.id
		WHERE u.department_id = ? AND vr.year = ?`
	args := []interface{}{departmentID, year}

	if statusFilter != nil {
		baseQuery += " AND vr.status_id = ?"
		args = append(args, *statusFilter)
	}
	baseQuery += " ORDER BY vr.created_at DESC"

	rowsReq, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса заявок подразделения %d: %w", departmentID, err)
	}
	defer rowsReq.Close()

	requestsMap := make(map[int]*models.VacationRequest)
	var requestIDs []interface{}

	for rowsReq.Next() {
		var req models.VacationRequest
		var comment sql.NullString
		err := rowsReq.Scan(
			&req.ID, &req.UserID, &req.Year, &req.StatusID,
			&comment,
			&req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			fmt.Printf("Ошибка сканирования заявки отдела %d: %v\n", departmentID, err)
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
		return nil, fmt.Errorf("ошибка итерации по заявкам отдела %d: %w", departmentID, err)
	}

	if len(requestIDs) == 0 {
		return []models.VacationRequest{}, nil
	}

	periods, err := r.getPeriodsByRequestIDs(requestIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения периодов для заявок отдела %d: %w", departmentID, err)
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


// GetAllVacationRequests получает все заявки для админов/менеджеров с фильтрами
func (r *VacationRepository) GetAllVacationRequests(yearFilter *int, statusFilter *int, userIDFilter *int, departmentIDFilter *int) ([]models.VacationRequestAdminView, error) {
	// Убрали JOIN к vacation_statuses
	queryBase := `
		SELECT 
			vr.id, vr.user_id, vr.year, vr.status_id, vr.comment, vr.created_at, vr.updated_at,
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
	if departmentIDFilter != nil {
		conditions = append(conditions, "u.department_id = ? AND u.department_id IS NOT NULL")
		args = append(args, *departmentIDFilter)
	}

	query := queryBase
	if len(conditions) > 0 {
		query += " WHERE " + sqlJoinStrings(conditions, " AND ")
	}
	query += " ORDER BY vr.created_at DESC"

	rowsReq, err := r.db.Query(query, args...)
	if err != nil {
		// Проверяем ошибку на отсутствие таблицы `vacation_statuses` явно
		// (Хотя мы убрали JOIN, оставим проверку на всякий случай, если проблема глубже)
		if strings.Contains(err.Error(), "vacation_statuses") {
			return nil, fmt.Errorf("ошибка структуры базы данных: таблица статусов недоступна. %w", err)
		}
		return nil, fmt.Errorf("ошибка запроса всех заявок: %w", err)
	}
	defer rowsReq.Close()

	requestsMap := make(map[int]*models.VacationRequestAdminView)
	var requestIDs []interface{}

	for rowsReq.Next() {
		var req models.VacationRequestAdminView
		var comment sql.NullString
		// Убрали сканирование statusName
		err := rowsReq.Scan(
			&req.ID, &req.UserID, &req.Year, &req.StatusID,
			&comment,
			&req.CreatedAt, &req.UpdatedAt,
			&req.UserFullName,
		)
		if err != nil {
			fmt.Printf("Ошибка сканирования AdminView заявки: %v\n", err)
			continue
		}
		if comment.Valid {
			req.Comment = comment.String
		}
		// Определяем StatusName на основе StatusID
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
		result = append(result, *req)
	}

	return result, nil
}

// getPeriodsByRequestIDs - вспомогательный метод для получения периодов для списка ID заявок
func (r *VacationRepository) getPeriodsByRequestIDs(requestIDs []interface{}) ([]models.VacationPeriod, error) {
	if len(requestIDs) == 0 {
		return []models.VacationPeriod{}, nil
	}
	query := fmt.Sprintf(`
		SELECT id, request_id, start_date, end_date, days_count, created_at, updated_at
		FROM vacation_periods
		WHERE request_id IN (?%s)`, sqlRepeatParams(len(requestIDs)-1))

	rows, err := r.db.Query(query, requestIDs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var periods []models.VacationPeriod
	for rows.Next() {
		var period models.VacationPeriod
		err := rows.Scan(
			&period.ID, &period.RequestID, &period.StartDate, &period.EndDate,
			&period.DaysCount, &period.CreatedAt, &period.UpdatedAt,
		)
		if err != nil {
			fmt.Printf("Ошибка сканирования периода (множественный запрос): %v\n", err)
			continue
		}
		periods = append(periods, period)
	}
	if err = rows.Err(); err != nil {
		return periods, err
	}
	return periods, nil
}


// --- Уведомления ---

// CreateNotification создает новое уведомление
func (r *VacationRepository) CreateNotification(notification *models.Notification) error {
	query := `
		INSERT INTO notifications (user_id, title, message, is_read, created_at) 
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`

	_, err := r.db.Exec(query, notification.UserID, notification.Title, notification.Message, notification.IsRead)
	if err != nil {
		return fmt.Errorf("ошибка создания уведомления: %w", err)
	}
	return nil
}


// --- Вспомогательные функции ---

// sqlRepeatParams генерирует строку плейсхолдеров (?, ?, ...)
func sqlRepeatParams(count int) string {
	if count < 1 {
		return ""
	}
	return strings.Repeat(", ?", count)
}

// sqlJoinStrings соединяет строки с разделителем
func sqlJoinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}
