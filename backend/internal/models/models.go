package models

import (
	"database/sql/driver" // Import driver package
	"encoding/json"
	"fmt" // Import fmt for error formatting
	"log"
	"strings"
	"time"
)

// --- Vacation Status Constants ---
// StatusDraft (ID=1) был удален. Остальные статусы НЕ сдвигаются.
const (
	// StatusDraft     = 1 // Черновик - УДАЛЕНО
	StatusPending   = 2 // На рассмотрении
	StatusApproved  = 3 // Утверждена
	StatusRejected  = 4 // Отклонена
	StatusCancelled = 5 // Отменена
)

// CustomDate is a wrapper around time.Time to handle specific JSON format and database scanning/valuing
type CustomDate struct {
	time.Time
}

const customDateFormat = time.RFC3339 // Use the standard RFC3339 constant

// UnmarshalJSON implements the json.Unmarshaler interface.
func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"") // Remove quotes
	if s == "null" || s == "" {
		cd.Time = time.Time{} // Handle null or empty string as zero time
		return nil
	}
	// Parse using the expected format RFC3339
	log.Printf("CustomDate UnmarshalJSON: Attempting to parse '%s'", s) // Лог перед парсингом
	t, err := time.Parse(customDateFormat, s)
	if err != nil {
		log.Printf("CustomDate UnmarshalJSON: Error parsing date string '%s': %v", s, err) // Лог ошибки парсинга
		return err                                                                         // Return error if parsing fails
	}
	log.Printf("CustomDate UnmarshalJSON: Successfully parsed '%s' into %v", s, t) // Лог успешного парсинга
	cd.Time = t
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if cd.Time.IsZero() {
		return json.Marshal(nil) // Marshal zero time as null
	}
	// Format back to RFC3339 when sending JSON responses
	return json.Marshal(cd.Time.Format(customDateFormat))
}

// Value implements the driver.Valuer interface.
// This method defines how CustomDate should be converted to a database value.
func (cd CustomDate) Value() (driver.Value, error) {
	if cd.Time.IsZero() {
		return nil, nil // Return nil for zero time
	}
	// Return the underlying time.Time, which the database driver understands.
	return cd.Time, nil
}

// Scan implements the sql.Scanner interface.
// This method defines how to scan a database value into CustomDate.
func (cd *CustomDate) Scan(value interface{}) error {
	if value == nil {
		cd.Time = time.Time{} // Handle NULL from database as zero time
		return nil
	}
	// Check if the value is already time.Time
	if t, ok := value.(time.Time); ok {
		cd.Time = t
		return nil
	}
	// If not time.Time, attempt to convert from string (less common for date types, but possible)
	// Or handle other potential database types if necessary.
	// For now, we assume the database returns a time.Time compatible type.
	return fmt.Errorf("не удалось сканировать тип %T в CustomDate", value)
}

// User - модель пользователя
type User struct {
	ID       int    `json:"id" db:"id"`
	Login    string `json:"login" db:"login"` // Изменено с Username на Login
	Password string `json:"-" db:"password"`
	FullName string `json:"full_name" db:"full_name"`
	// Email                string    `json:"email" db:"email"` // Удалено
	OrganizationalUnitID *int      `json:"organizational_unit_id,omitempty" db:"organizational_unit_id"` // Переименовано с department_id
	PositionID           *int      `json:"position_id,omitempty" db:"position_id"`
	PositionName         *string   `json:"positionName,omitempty" db:"position_name"` // Use pointer for nullable position name
	IsAdmin              bool      `json:"is_admin" db:"is_admin"`
	IsManager            bool      `json:"is_manager" db:"is_manager"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// UserProfileDTO - DTO для отображения профиля пользователя с иерархией юнитов
type UserProfileDTO struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	FullName string `json:"full_name"`
	// Email         string    `json:"email"` // Удалено, так как не выбирается в GetUserProfileByID
	PositionName  *string   `json:"positionName,omitempty"`  // Optional position name
	Department    *string   `json:"department,omitempty"`    // Название департамента (верхний уровень)
	SubDepartment *string   `json:"subDepartment,omitempty"` // Название подотдела (средний уровень)
	Sector        *string   `json:"sector,omitempty"`        // Название сектора (нижний уровень)
	IsAdmin       bool      `json:"is_admin"`
	IsManager     bool      `json:"is_manager"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserUpdateDTO - структура для обновления данных пользователя
type UserUpdateDTO struct {
	FullName             *string `json:"full_name"`              // Указатель, чтобы различать пустую строку и отсутствие значения
	Password             *string `json:"password"`               // Указатель для опционального обновления пароля
	PositionID           *int    `json:"position_id"`            // Указатель для опционального обновления должности
	OrganizationalUnitID *int    `json:"organizational_unit_id"` // Указатель для опционального обновления юнита
}

// UserUpdateAdminDTO - структура для обновления данных пользователя администратором
type UserUpdateAdminDTO struct {
	PositionID           *int  `json:"position_id"`            // Указатель для опционального обновления должности
	OrganizationalUnitID *int  `json:"organizational_unit_id"` // Указатель для опционального обновления юнита
	IsAdmin              *bool `json:"is_admin"`               // Указатель для опционального обновления статуса админа
	IsManager            *bool `json:"is_manager"`             // Указатель для опционального обновления статуса менеджера
}

// Position - модель должности (без GroupID)
type Position struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	// GroupID удалено
}

// OrganizationalUnit - модель для иерархии подразделений, отделов, секторов и т.д.
type OrganizationalUnit struct {
	ID        int                   `json:"id" db:"id"`
	Name      string                `json:"name" db:"name"`
	UnitType  string                `json:"unit_type" db:"unit_type"`             // 'DEPARTMENT', 'SUB_DEPARTMENT', 'SECTOR', etc.
	ParentID  *int                  `json:"parent_id,omitempty" db:"parent_id"`   // Указатель для NULL parent_id (корень)
	ManagerID *int                  `json:"manager_id,omitempty" db:"manager_id"` // Указатель на ID руководителя
	CreatedAt time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt time.Time             `json:"updated_at" db:"updated_at"`
	Children  []*OrganizationalUnit `json:"children,omitempty" db:"-"`  // Дочерние юниты (для построения дерева), не маппится на БД
	Positions []Position            `json:"positions,omitempty" db:"-"` // Должности в этом юните (если применимо), не маппится на БД
	Users     []User                `json:"users,omitempty" db:"-"`     // Пользователи в этом юните (если применимо), не маппится на БД
}

// VacationRequest - модель заявки на отпуск
type VacationRequest struct {
	ID            int              `json:"id" db:"id"`
	UserID        int              `json:"user_id" db:"user_id"`
	Year          int              `json:"year" db:"year"`
	StatusID      int              `json:"status_id" db:"status_id"`
	DaysRequested int              `json:"days_requested" db:"days_requested"` // Добавлено поле
	Comment       string           `json:"comment" db:"comment"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
	Periods       []VacationPeriod `json:"periods"`
}

// VacationPeriod - модель периода отпуска
type VacationPeriod struct {
	ID        int        `json:"id" db:"id"`
	RequestID int        `json:"request_id" db:"request_id"`
	StartDate CustomDate `json:"start_date" db:"start_date"` // Use CustomDate
	EndDate   CustomDate `json:"end_date" db:"end_date"`     // Use CustomDate
	DaysCount int        `json:"days_count" db:"days_count"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"` // Keep time.Time for DB timestamps
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"` // Keep time.Time for DB timestamps
}

// VacationLimit - модель лимита отпуска
type VacationLimit struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Year      int       `json:"year" db:"year"`
	TotalDays int       `json:"total_days" db:"total_days"`
	UsedDays  int       `json:"used_days" db:"used_days"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Notification - модель уведомления
type Notification struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	Message   string    `json:"message" db:"message"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// VacationStatus - модель статуса отпуска
type VacationStatus struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

// Intersection - модель пересечения отпусков
type Intersection struct {
	UserID1   int        `json:"user_id_1"`
	UserName1 string     `json:"user_name_1"`
	UserID2   int        `json:"user_id_2"`
	UserName2 string     `json:"user_name_2"`
	StartDate CustomDate `json:"start_date"` // Use CustomDate for consistency if needed
	EndDate   CustomDate `json:"end_date"`   // Use CustomDate for consistency if needed
	DaysCount int        `json:"days_count"`
}

// UserWithLimitDTO represents user data along with their vacation limit for a specific year.
type UserWithLimitDTO struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	// Email             string  `json:"email"`            // Удалено, так как больше не используется
	Position               *string `json:"position,omitempty"`               // Добавлено поле для должности (указатель для NULL)
	OrganizationalUnitName *string `json:"organizationalUnitName,omitempty"` // <--- Добавлено: Название орг. юнита
	VacationLimitDays      *int    `json:"vacation_limit_days"`              // Pointer to handle null/absence of limit
}

// --- New DTO for Admin/Manager View ---
// VacationRequestAdminView includes user and status details for admin/manager displays.
type VacationRequestAdminView struct {
	ID            int              `json:"id" db:"id"`
	UserID        int              `json:"user_id" db:"user_id"`
	UserFullName  string           `json:"user_full_name" db:"full_name"` // Added user's full name
	Year          int              `json:"year" db:"year"`
	StatusID      int              `json:"status_id" db:"status_id"`
	StatusName    string           `json:"status_name" db:"status_name"`       // Added status name
	DaysRequested int              `json:"days_requested" db:"days_requested"` // Добавлено поле (уже было в схеме, добавляем сюда для согласованности)
	Comment       string           `json:"comment" db:"comment"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
	Periods       []VacationPeriod `json:"periods"`    // Populated separately
	TotalDays     int              `json:"total_days"` // Calculated total days across periods (остается для отображения, но не используется для логики списания)
}

// --- DTO for Unit/User List ---

// UnitListItemDTO - Элемент списка для отображения юнитов или пользователей в иерархии
type UnitListItemDTO struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`                // "unit" or "user"
	UnitType *string `json:"unit_type,omitempty"` // Specific type if Type is "unit" ('DEPARTMENT', 'SECTOR', etc.)
	Position *string `json:"position,omitempty"`  // Position name if Type is "user"
	// Можно добавить другие поля по необходимости, например, ManagerID для юнита
}

// UserWithLimitAdminDTO - DTO для отображения пользователя в админ-панели управления подразделениями, включая лимит отпуска
type UserWithLimitAdminDTO struct {
	ID           int     `json:"id" db:"id"`
	FullName     string  `json:"full_name" db:"full_name"`
	PositionName *string `json:"position_name,omitempty" db:"position_name"` // Указатель для NULL
	TotalDays    *int    `json:"total_days" db:"total_days"`                 // Указатель, так как лимит может отсутствовать для года
	// Возможно, добавим organizational_unit_id, если потребуется для других целей
}

// ConflictingPeriod - Структура для описания конфликта отпусков
type ConflictingPeriod struct {
	ConflictingUserID       int        `json:"conflicting_user_id"`        // ID второго пользователя в конфликте
	ConflictingUserFullName string     `json:"conflicting_user_full_name"` // ФИО второго пользователя
	ConflictingRequestID    int        `json:"conflicting_request_id"`     // ID заявки второго пользователя
	ConflictingPeriodID     int        `json:"conflicting_period_id"`      // ID периода второго пользователя
	ConflictingStartDate    CustomDate `json:"conflicting_start_date"`     // Начало периода второго пользователя
	ConflictingEndDate      CustomDate `json:"conflicting_end_date"`       // Конец периода второго пользователя
	OriginalUserID          int        `json:"original_user_id"`           // ID первого пользователя в конфликте
	OriginalUserFullName    string     `json:"original_user_full_name"`    // ФИО первого пользователя
	OriginalRequestID       int        `json:"original_request_id"`        // ID заявки первого пользователя
	OriginalPeriodID        int        `json:"original_period_id"`         // ID периода первого пользователя
	OriginalStartDate       CustomDate `json:"original_start_date"`        // Начало периода первого пользователя
	OriginalEndDate         CustomDate `json:"original_end_date"`          // Конец периода первого пользователя
	OverlapStartDate        CustomDate `json:"overlap_start_date"`         // Начало пересечения
	OverlapEndDate          CustomDate `json:"overlap_end_date"`           // Конец пересечения
}

// --- Dashboard DTOs ---

// ManagerDashboardData - DTO для дашборда руководителя
type ManagerDashboardData struct {
	PendingRequestsCount  int                 `json:"pending_requests_count"`   // Количество заявок "На рассмотрении"
	ApprovedDaysCountYear int                 `json:"approved_days_count_year"` // Сумма дней в утвержденных заявках за год (или выбранный период)
	RejectedDaysCountYear int                 `json:"rejected_days_count_year"` // Сумма дней в отклоненных заявках за год (или выбранный период)
	PendingDaysCountYear  int                 `json:"pending_days_count_year"`  // Сумма дней в заявках "На рассмотрении" за год (или выбранный период)
	UpcomingConflicts     []ConflictingPeriod `json:"upcoming_conflicts"`       // Список ближайших конфликтов (пересечений утвержденных отпусков)
	SubordinateUserCount  int                 `json:"subordinate_user_count"`   // Общее количество подчиненных пользователей
	// Можно добавить другие счетчики при необходимости
}

// VacationExportRow структура для строки данных в экспорте Т-7
// Поля соответствуют столбцам формы Т-7 из примера
type VacationExportRow struct {
	SequenceNumber        int         `json:"sequence_number"`         // 1. № п/п (будет присвоен при генерации)
	UnitName              string      `json:"unit_name"`               // 2. Структурное подразделение
	PositionName          string      `json:"position_name"`           // 3. Должность
	FullName              string      `json:"full_name"`               // 4. Фамилия, имя, отчество
	EmployeeNumber        string      `json:"employee_number"`         // 5. Табельный номер (используем UserID или Login)
	PlannedDaysMain       int         `json:"planned_days_main"`       // 6. Дни основного отпуска (из периода)
	PlannedDaysAdditional int         `json:"planned_days_additional"` // 7. Дни дополнительного отпуска (пока 0)
	PlannedDaysTotal      int         `json:"planned_days_total"`      // 8. Итого дней (сумма)
	PlannedDate           CustomDate  `json:"planned_date"`            // 9. Дата запланированная (StartDate периода)
	ActualDate            *CustomDate `json:"actual_date,omitempty"`   // 10. Дата фактическая (StartDate, если утвержден?) - Используем указатель
	TransferReason        string      `json:"transfer_reason"`         // 11. Основание переноса (пока пусто)
	TransferDate          *CustomDate `json:"transfer_date,omitempty"` // 12. Дата предполагаемого отпуска (пока пусто) - Используем указатель
	Note                  string      `json:"note"`                    // 13. Примечание (пока пусто)
}
