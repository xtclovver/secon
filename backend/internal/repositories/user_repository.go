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
	FindByLogin(login string) (*models.User, error) // Изменено FindByUsername на FindByLogin
	FindByID(id int) (*models.User, error)
	GetUsersByDepartment(departmentID int) ([]models.User, error)
	CreateUser(user *models.User) error
	// UpdateUser теперь принимает ID и DTO для частичного обновления
	UpdateUser(userID int, updateData *models.UserUpdateDTO) error
	GetAllUsersWithLimits(year int) ([]models.UserWithLimitDTO, error)
	GetAllPositionsGrouped() ([]models.PositionGroup, error) // Новый метод
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

// FindByLogin находит пользователя по логину
func (r *UserRepository) FindByLogin(login string) (*models.User, error) { // Изменено FindByUsername на FindByLogin
	// Запрос к БД для поиска пользователя с названием должности
	query := `
		SELECT 
			u.id, u.login, u.password, u.full_name, u.email, 
			u.department_id, u.position_id, p.name AS position_name, 
			u.is_admin, u.is_manager, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN positions p ON u.position_id = p.id
		WHERE u.login = ?` // Добавлен LEFT JOIN и выборка p.name

	row := r.db.QueryRow(query, login) // username -> login
	user := &models.User{}

	// Используем nullable типы для department_id, position_id и position_name
	var departmentID sql.NullInt64
	var positionID sql.NullInt64
	var positionName sql.NullString // Для названия должности

	err := row.Scan(
		&user.ID, &user.Login, &user.Password, &user.FullName, &user.Email, // Username -> Login
		&departmentID, // Сканируем в nullable тип
		&positionID,   // Сканируем в nullable тип
		&positionName, // Сканируем название должности
		&user.IsAdmin, &user.IsManager, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Пользователь не найден, ошибки нет
		}
		// Логирование ошибки может быть полезно
		// log.Printf("Ошибка сканирования пользователя %s: %v", login, err) // username -> login
		return nil, fmt.Errorf("ошибка при поиске пользователя в БД: %w", err)
	}

	// Преобразуем nullable тип в указатель на int
	if departmentID.Valid {
		deptID := int(departmentID.Int64)
		user.DepartmentID = &deptID
	} else {
		user.DepartmentID = nil
	}

	// Преобразуем nullable position_id в указатель на int
	if positionID.Valid {
		posID := int(positionID.Int64)
		user.PositionID = &posID
	} else {
		user.PositionID = nil
	}

	// Преобразуем nullable position_name в указатель на string
	if positionName.Valid {
		user.PositionName = &positionName.String
	} else {
		user.PositionName = nil // Явно устанавливаем nil, если имя должности NULL
	}

	return user, nil
}

// FindByID находит пользователя по ID
func (r *UserRepository) FindByID(id int) (*models.User, error) {
	// Запрос к БД для поиска пользователя с названием должности
	query := `
		SELECT 
			u.id, u.login, u.password, u.full_name, u.email, 
			u.department_id, u.position_id, p.name AS position_name, 
			u.is_admin, u.is_manager, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN positions p ON u.position_id = p.id
		WHERE u.id = ?` // Добавлен LEFT JOIN и выборка p.name

	row := r.db.QueryRow(query, id)
	user := &models.User{}

	// Используем nullable типы для department_id, position_id и position_name
	var (
		departmentID sql.NullInt64
		positionID   sql.NullInt64
		positionName sql.NullString // Для названия должности
	)

	err := row.Scan(
		&user.ID, &user.Login, &user.Password, &user.FullName, &user.Email, // Username -> Login
		&departmentID, // Сканируем в nullable тип
		&positionID,   // Сканируем в nullable тип
		&positionName, // Сканируем название должности
		&user.IsAdmin, &user.IsManager, &user.CreatedAt, &user.UpdatedAt,
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

	if positionID.Valid {
		posID := int(positionID.Int64)
		user.PositionID = &posID
	} else {
		user.PositionID = nil
	}

	// Преобразуем nullable position_name в указатель на string
	if positionName.Valid {
		user.PositionName = &positionName.String
	} else {
		user.PositionName = nil // Явно устанавливаем nil, если имя должности NULL
	}

	return user, nil
}

// GetUsersByDepartment получает список пользователей по ID подразделения
func (r *UserRepository) GetUsersByDepartment(departmentID int) ([]models.User, error) {
	query := `
		SELECT id, login, full_name, email, is_admin, is_manager 
		FROM users 
		WHERE department_id = ?` // username -> login

	rows, err := r.db.Query(query, departmentID)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса пользователей подразделения: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		// Сканируем только нужные поля для этого запроса
		if err := rows.Scan(&user.ID, &user.Login, &user.FullName, &user.Email, &user.IsAdmin, &user.IsManager); err != nil { // Username -> Login
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
		INSERT INTO users (login, password, full_name, email, department_id, position_id, is_admin, is_manager, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)` // username -> login

	result, err := r.db.Exec(query,
		user.Login, string(hashedPassword), user.FullName, user.Email, // Username -> Login
		user.DepartmentID, // Может быть nil
		user.PositionID,   // Может быть nil
		user.IsAdmin, user.IsManager,
	)
	if err != nil {
		// Обработка специфических ошибок БД (например, дубликат login) может быть добавлена здесь
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

// UpdateUser обновляет данные пользователя на основе предоставленных полей в DTO
func (r *UserRepository) UpdateUser(userID int, updateData *models.UserUpdateDTO) error {
	if updateData == nil {
		return errors.New("данные для обновления не предоставлены")
	}

	query := "UPDATE users SET "
	args := []interface{}{}
	argID := 1 // Счетчик для плейсхолдеров

	// Динамически строим запрос
	if updateData.FullName != nil {
		query += "full_name = ?, " // Используем стандартный плейсхолдер '?'
		args = append(args, *updateData.FullName)
		argID++
	}
	if updateData.Password != nil && *updateData.Password != "" { // Обновляем пароль, только если он не пустой
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updateData.Password), 12)
		if err != nil {
			return fmt.Errorf("ошибка хеширования нового пароля: %w", err)
		}
		query += "password = ?, " // Используем стандартный плейсхолдер '?'
		args = append(args, string(hashedPassword))
		argID++
	}
	if updateData.PositionID != nil {
		query += "position_id = ?, " // Используем стандартный плейсхолдер '?'
		args = append(args, *updateData.PositionID)
		argID++
	}

	// Если нечего обновлять (кроме updated_at)
	if argID == 1 {
		return errors.New("нет полей для обновления")
	}

	// Добавляем обновление времени и условие WHERE
	query += "updated_at = CURRENT_TIMESTAMP " // Пробел перед WHERE важен
	query += "WHERE id = ?"                    // Используем стандартный плейсхолдер '?'
	args = append(args, userID)                // Добавляем ID пользователя в конец списка аргументов

	// Заменяем плейсхолдеры MySQL (?) на плейсхолдеры PostgreSQL ($) если нужно
	// query = strings.ReplaceAll(query, "?", "$") // Раскомментировать для PostgreSQL

	// Выполняем запрос
	result, err := r.db.Exec(query, args...)
	if err != nil {
		// TODO: Добавить более специфичную обработку ошибок БД, например, для неверного position_id
		return fmt.Errorf("ошибка выполнения запроса на обновление пользователя: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}
	if rowsAffected == 0 {
		// Это может означать, что пользователь с таким ID не найден
		return errors.New("пользователь для обновления не найден или данные не изменились")
	}

	return nil
}

// GetAllUsersWithLimits получает всех пользователей вместе с их лимитом отпуска на указанный год
func (r *UserRepository) GetAllUsersWithLimits(year int) ([]models.UserWithLimitDTO, error) {
	// Используем LEFT JOIN, чтобы получить всех пользователей, их должности и лимиты на год
	query := `
		SELECT 
			u.id, 
			u.full_name, 
			u.email,         -- Оставляем email на всякий случай, но добавим должность
			p.name AS position_name, -- Добавляем название должности
			vl.total_days    -- Лимит дней
		FROM users u
		LEFT JOIN positions p ON u.position_id = p.id -- Добавляем JOIN для получения должности
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
		var limitDays sql.NullInt64     // Для total_days
		var positionName sql.NullString // Для position_name

		// Сканируем ID, ФИО, Email, Название должности, Лимит дней
		if err := rows.Scan(
			&userDTO.ID,
			&userDTO.FullName,
			&userDTO.Email, // Сканируем email (даже если не используем во фронте сразу)
			&positionName,  // Сканируем название должности
			&limitDays,
		); err != nil {
			// log.Printf("Ошибка сканирования пользователя с лимитом: %v", err)
			// Можно пропустить пользователя или вернуть ошибку
			continue
		}

		// Преобразуем NullInt64 (лимит дней) в *int
		if limitDays.Valid {
			days := int(limitDays.Int64)
			userDTO.VacationLimitDays = &days
		} else {
			userDTO.VacationLimitDays = nil
		}

		// Преобразуем NullString (название должности) в *string
		if positionName.Valid {
			name := positionName.String
			userDTO.Position = &name // Предполагаем, что поле называется Position в DTO
		} else {
			userDTO.Position = nil // Явно указываем nil, если должности нет
		}

		usersWithLimits = append(usersWithLimits, userDTO)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по пользователям с лимитами: %w", err)
	}

	return usersWithLimits, nil
}

// GetAllPositionsGrouped получает все должности, сгруппированные по категориям
func (r *UserRepository) GetAllPositionsGrouped() ([]models.PositionGroup, error) {
	// 1. Получаем все группы, отсортированные по sort_order
	groupQuery := `SELECT id, name, sort_order FROM position_groups ORDER BY sort_order ASC`
	groupRows, err := r.db.Query(groupQuery)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса групп должностей: %w", err)
	}
	defer groupRows.Close()

	groupsMap := make(map[int]*models.PositionGroup) // Карта для быстрого доступа к группам по ID
	var groupOrder []*models.PositionGroup           // Слайс для сохранения порядка групп

	for groupRows.Next() {
		var group models.PositionGroup
		if err := groupRows.Scan(&group.ID, &group.Name, &group.SortOrder); err != nil {
			// log.Printf("Ошибка сканирования группы должностей: %v", err)
			continue // Пропускаем ошибочную строку
		}
		group.Positions = []models.Position{} // Инициализируем слайс должностей
		groupsMap[group.ID] = &group
		groupOrder = append(groupOrder, &group)
	}
	if err = groupRows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по группам должностей: %w", err)
	}

	// 2. Получаем все должности, отсортированные по имени
	positionQuery := `SELECT id, name, group_id FROM positions ORDER BY name ASC`
	positionRows, err := r.db.Query(positionQuery)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса должностей: %w", err)
	}
	defer positionRows.Close()

	for positionRows.Next() {
		var position models.Position
		if err := positionRows.Scan(&position.ID, &position.Name, &position.GroupID); err != nil {
			// log.Printf("Ошибка сканирования должности: %v", err)
			continue // Пропускаем ошибочную строку
		}

		// Добавляем должность в соответствующую группу
		if group, ok := groupsMap[position.GroupID]; ok {
			group.Positions = append(group.Positions, position)
		} else {
			// log.Printf("Предупреждение: Должность %d имеет недействительный group_id %d", position.ID, position.GroupID)
		}
	}
	if err = positionRows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по должностям: %w", err)
	}

	// Преобразуем слайс указателей в слайс значений для возврата
	resultGroups := make([]models.PositionGroup, len(groupOrder))
	for i, groupPtr := range groupOrder {
		resultGroups[i] = *groupPtr
	}

	return resultGroups, nil
}

// TODO: Добавить методы для смены пароля, удаления пользователя и т.д.
