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
const (
	StatusDraft     = 1 // Черновик
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
	ID           int       `json:"id" db:"id"`
	Login        string    `json:"login" db:"login"` // Изменено с Username на Login
	Password     string    `json:"-" db:"password"`
	FullName     string    `json:"full_name" db:"full_name"`
	Email        string    `json:"email" db:"email"`
	DepartmentID *int      `json:"department_id" db:"department_id"`
	PositionID   *int      `json:"position_id,omitempty" db:"position_id"`    // Добавлено поле для должности, omitempty если не будет имени
	PositionName *string   `json:"positionName,omitempty" db:"position_name"` // Изменен JSON тег на positionName
	IsAdmin      bool      `json:"is_admin" db:"is_admin"`
	IsManager    bool      `json:"is_manager" db:"is_manager"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserUpdateDTO - структура для обновления данных пользователя
type UserUpdateDTO struct {
	FullName   *string `json:"full_name"`   // Указатель, чтобы различать пустую строку и отсутствие значения
	Password   *string `json:"password"`    // Указатель для опционального обновления пароля
	PositionID *int    `json:"position_id"` // Указатель для опционального обновления должности (только для админа/менеджера)
}

// PositionGroup - модель группы должностей
type PositionGroup struct {
	ID        int        `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	SortOrder int        `json:"sort_order" db:"sort_order"`
	Positions []Position `json:"positions"` // Список должностей в этой группе
}

// Position - модель должности
type Position struct {
	ID      int    `json:"id" db:"id"`
	Name    string `json:"name" db:"name"`
	GroupID int    `json:"-" db:"group_id"` // Скрываем group_id в JSON, так как он будет в структуре группы
}

// Department - модель подразделения
type Department struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	ManagerID *int      `json:"manager_id" db:"manager_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
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
	ID                int     `json:"id"`
	FullName          string  `json:"full_name"`
	Email             string  `json:"email"`               // Оставляем email, вдруг понадобится
	Position          *string `json:"position,omitempty"`  // Добавлено поле для должности (указатель для NULL)
	VacationLimitDays *int    `json:"vacation_limit_days"` // Pointer to handle null/absence of limit
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
