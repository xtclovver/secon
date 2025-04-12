package repositories

import (
	"database/sql" // Оставляем один импорт
	"errors" 
	"fmt" // Для форматирования ошибок

	"vacation-scheduler/internal/models"
	"golang.org/x/crypto/bcrypt" // Раскомментирован
)

// UserRepositoryInterface определяет методы для репозитория пользователей
type UserRepositoryInterface interface {
	FindByUsername(username string) (*models.User, error)
	FindByID(id int) (*models.User, error)
	GetUsersByDepartment(departmentID int) ([]models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	GetAllUsersWithLimits(year int) ([]models.UserWithLimitDTO, error)
	// Добавьте другие методы по мере необходимости
}

// UserRepository реализует UserRepositoryInterface
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByUsername находит пользователя по имени пользователя
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	// Запрос к БД для поиска пользователя
	query := `
		SELECT id, username, password, full_name, email, department_id, is_admin, is_manager, created_at, updated_at 
		FROM users 
		WHERE username = ?`
		
	row := r.db.QueryRow(query, username)
	user := &models.User{}
	
	// Используем nullable типы для department_id, если он может быть NULL в БД
	var departmentID sql.NullInt64 

	err := row.Scan(
		&user.ID, &user.Username, &user.Password, &user.FullName, &user.Email, 
		&departmentID, // Сканируем в nullable тип
		&user.IsAdmin, &user.IsManager, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Пользователь не найден, ошибки нет
		}
		// Логирование ошибки может быть полезно
		// log.Printf("Ошибка сканирования пользователя %s: %v", username, err)
		return nil, fmt.Errorf("ошибка при поиске пользователя в БД: %w", err) 
	}
	
	// Преобразуем nullable тип в указатель на int
	if departmentID.Valid {
		deptID := int(departmentID.Int64)
		user.DepartmentID = &deptID
	} else {
		user.DepartmentID = nil
	}

	return user, nil
}

// FindByID находит пользователя по ID
func (r *UserRepository) FindByID(id int) (*models.User, error) {
	query := `
		SELECT id, username, password, full_name, email, department_id, is_admin, is_manager, created_at, updated_at 
		FROM users 
		WHERE id = ?`
		
	row := r.db.QueryRow(query, id)
	user := &models.User{}
	var departmentID sql.NullInt64

	err := row.Scan(
		&user.ID, &user.Username, &user.Password, &user.FullName, &user.Email, 
		&departmentID, &user.IsAdmin, &user.IsManager, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Пользователь не найден
		}
		return nil, fmt.Errorf("ошибка при поиске пользователя по ID: %w", err)
	}
	
	if departmentID.Valid {
		deptID := int(departmentID.Int64)
		user.DepartmentID = &deptID
	} else {
		user.DepartmentID = nil
	}

	return user, nil
}

// GetUsersByDepartment получает список пользователей по ID подразделения
func (r *UserRepository) GetUsersByDepartment(departmentID int) ([]models.User, error) {
	query := `
		SELECT id, username, full_name, email, is_admin, is_manager 
		FROM users 
		WHERE department_id = ?`
		
	rows, err := r.db.Query(query, departmentID)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса пользователей подразделения: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		// Сканируем только нужные поля для этого запроса
		if err := rows.Scan(&user.ID, &user.Username, &user.FullName, &user.Email, &user.IsAdmin, &user.IsManager); err != nil {
			// Логирование ошибки сканирования может быть полезно
			// log.Printf("Ошибка сканирования пользователя подразделения: %v", err)
			// Продолжаем сканировать остальных, но можно и вернуть ошибку
			continue 
		}
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
         return nil, fmt.Errorf("ошибка при итерации по пользователям подразделения: %w", err)
    }

	return users, nil
}

// CreateUser создает нового пользователя (с хешированием пароля)
func (r *UserRepository) CreateUser(user *models.User) error {
	// Хешируем пароль перед сохранением с cost = 12
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12) // Используем cost 12
	if err != nil {
		return fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	query := `
		INSERT INTO users (username, password, full_name, email, department_id, is_admin, is_manager, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
		
	result, err := r.db.Exec(query, 
		user.Username, string(hashedPassword), user.FullName, user.Email, 
		user.DepartmentID, // Может быть nil
		user.IsAdmin, user.IsManager,
	)
	if err != nil {
		// Обработка специфических ошибок БД (например, дубликат username) может быть добавлена здесь
		return fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	// Получаем ID созданного пользователя
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("ошибка получения ID нового пользователя: %w", err)
	}
	user.ID = int(id)
	user.Password = "" // Очищаем пароль после сохранения

	return nil
}

// UpdateUser обновляет данные пользователя (кроме пароля)
func (r *UserRepository) UpdateUser(user *models.User) error {
	query := `
		UPDATE users 
		SET full_name = ?, email = ?, department_id = ?, is_admin = ?, is_manager = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`
		
	result, err := r.db.Exec(query, 
		user.FullName, user.Email, user.DepartmentID, 
		user.IsAdmin, user.IsManager, user.ID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления пользователя: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("пользователь для обновления не найден")
	}

	return nil
}


// GetAllUsersWithLimits получает всех пользователей вместе с их лимитом отпуска на указанный год
func (r *UserRepository) GetAllUsersWithLimits(year int) ([]models.UserWithLimitDTO, error) {
	// Используем LEFT JOIN, чтобы получить всех пользователей, даже если у них нет лимита на этот год
	query := `
		SELECT 
			u.id, 
			u.full_name, 
			u.email, 
			vl.total_days 
		FROM users u
		LEFT JOIN vacation_limits vl ON u.id = vl.user_id AND vl.year = ?
		ORDER BY u.full_name` // Сортируем по имени для удобства отображения

	rows, err := r.db.Query(query, year)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса пользователей с лимитами: %w", err)
	}
	defer rows.Close()

	var usersWithLimits []models.UserWithLimitDTO
	for rows.Next() {
		var userDTO models.UserWithLimitDTO
		var limitDays sql.NullInt64 // Используем NullInt64 для total_days, так как JOIN может вернуть NULL

		if err := rows.Scan(&userDTO.ID, &userDTO.FullName, &userDTO.Email, &limitDays); err != nil {
			// log.Printf("Ошибка сканирования пользователя с лимитом: %v", err)
			// Можно пропустить пользователя или вернуть ошибку
			continue 
		}

		// Преобразуем NullInt64 в *int
		if limitDays.Valid {
			days := int(limitDays.Int64)
			userDTO.VacationLimitDays = &days
		} else {
			userDTO.VacationLimitDays = nil // Явно указываем nil, если лимита нет
		}

		usersWithLimits = append(usersWithLimits, userDTO)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по пользователям с лимитами: %w", err)
	}

	return usersWithLimits, nil
}


// TODO: Добавить методы для смены пароля, удаления пользователя и т.д.
