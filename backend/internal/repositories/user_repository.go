package repositories

import (
	"database/sql" // Оставляем один импорт
	"errors"
	"fmt" // Для форматирования ошибок
	"strings"

	"vacation-scheduler/internal/models"

	"golang.org/x/crypto/bcrypt" // Раскомментирован
)

// UserRepositoryInterface определяет методы для репозитория пользователей
type UserRepositoryInterface interface {
	FindByLogin(login string) (*models.User, error)
	FindByID(id int) (*models.User, error)
	GetUsersByOrganizationalUnit(unitID int) ([]models.User, error) // Изменено GetUsersByDepartment
	CreateUser(user *models.User) error
	UpdateUser(userID int, updateData *models.UserUpdateDTO) error
	GetUsersByUnitIDs(unitIDs []int) ([]models.User, error) // Добавлен метод для получения пользователей по списку ID юнитов
	GetAllUsersWithLimits(year int) ([]models.UserWithLimitDTO, error)
	GetAllPositions() ([]models.Position, error)                                                         // Восстановлен метод для получения всех должностей
	GetUserProfileByID(userID int) (*models.UserProfileDTO, error)                                       // Новый метод для профиля
	FindByOrganizationalUnitID(unitID int) ([]*models.User, error)                                       // Найти пользователей по ID орг. юнита
	GetUsersWithLimitsByOrganizationalUnit(unitID int, year int) ([]models.UserWithLimitAdminDTO, error) // Новый метод
	// TODO: Добавить интерфейсы для работы с OrganizationalUnit
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
			u.id, u.login, u.password, u.full_name, 
			u.organizational_unit_id, u.position_id, p.name AS position_name, 
			u.is_admin, u.is_manager, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN positions p ON u.position_id = p.id
		WHERE u.login = ?` // Добавлен LEFT JOIN и выборка p.name

	row := r.db.QueryRow(query, login) // username -> login
	user := &models.User{}

	// Используем nullable типы для organizational_unit_id, position_id и position_name
	var organizationalUnitID sql.NullInt64 // departmentID -> organizationalUnitID
	var positionID sql.NullInt64
	var positionName sql.NullString

	err := row.Scan(
		&user.ID, &user.Login, &user.Password, &user.FullName,
		&organizationalUnitID, // Сканируем в nullable тип
		&positionID,
		&positionName,
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

	// Преобразуем nullable organizational_unit_id в указатель на int
	if organizationalUnitID.Valid {
		unitID := int(organizationalUnitID.Int64)
		user.OrganizationalUnitID = &unitID
	} else {
		user.OrganizationalUnitID = nil
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
			u.id, u.login, u.password, u.full_name, 
			u.organizational_unit_id, u.position_id, p.name AS position_name, 
			u.is_admin, u.is_manager, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN positions p ON u.position_id = p.id
		WHERE u.id = ?` // Добавлен LEFT JOIN и выборка p.name

	row := r.db.QueryRow(query, id)
	user := &models.User{}

	// Используем nullable типы для organizational_unit_id, position_id и position_name
	var (
		organizationalUnitID sql.NullInt64 // departmentID -> organizationalUnitID
		positionID           sql.NullInt64
		positionName         sql.NullString
	)

	err := row.Scan(
		&user.ID, &user.Login, &user.Password, &user.FullName,
		&organizationalUnitID, // Сканируем в nullable тип
		&positionID,
		&positionName,
		&user.IsAdmin, &user.IsManager, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Пользователь не найден
		}
		return nil, fmt.Errorf("ошибка при поиске пользователя по ID: %w", err)
	}

	// Преобразуем nullable organizational_unit_id в указатель на int
	if organizationalUnitID.Valid {
		unitID := int(organizationalUnitID.Int64)
		user.OrganizationalUnitID = &unitID
	} else {
		user.OrganizationalUnitID = nil
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

// GetUsersByOrganizationalUnit получает список пользователей по ID орг. юнита
func (r *UserRepository) GetUsersByOrganizationalUnit(unitID int) ([]models.User, error) { // Изменено GetUsersByDepartment
	query := `
		SELECT u.id, u.login, u.full_name, u.position_id, p.name as position_name, u.is_admin, u.is_manager 
		FROM users u
		LEFT JOIN positions p ON u.position_id = p.id
		WHERE u.organizational_unit_id = ?` // Добавлен JOIN для должности

	rows, err := r.db.Query(query, unitID)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса пользователей орг. юнита: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var positionID sql.NullInt64    // Для nullable position_id
		var positionName sql.NullString // Для nullable position_name
		// Сканируем поля, включая название должности
		if err := rows.Scan(&user.ID, &user.Login, &user.FullName, &positionID, &positionName, &user.IsAdmin, &user.IsManager); err != nil {
			// log.Printf("Ошибка сканирования пользователя орг. юнита: %v", err)
			continue
		}
		// Устанавливаем PositionID
		if positionID.Valid {
			posID := int(positionID.Int64)
			user.PositionID = &posID
		} else {
			user.PositionID = nil
		}
		// Устанавливаем PositionName
		if positionName.Valid {
			posName := positionName.String
			user.PositionName = &posName
		} else {
			user.PositionName = nil
		}
		users = append(users, user) // Добавляем пользователя один раз
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по пользователям орг. юнита: %w", err)
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
		INSERT INTO users (login, password, full_name, organizational_unit_id, position_id, is_admin, is_manager, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)` // email уже удален

	result, err := r.db.Exec(query,
		user.Login, string(hashedPassword), user.FullName,
		user.OrganizationalUnitID, // Может быть nil
		user.PositionID,           // Может быть nil
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
	argID := 1

	// Строим запрос динамически
	updates := []string{}
	if updateData.FullName != nil {
		updates = append(updates, "full_name = ?")
		args = append(args, *updateData.FullName)
		argID++
	}
	if updateData.Password != nil && *updateData.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updateData.Password), 12)
		if err != nil {
			return fmt.Errorf("ошибка хеширования нового пароля: %w", err)
		}
		updates = append(updates, "password = ?")
		args = append(args, string(hashedPassword))
		argID++
	}
	if updateData.PositionID != nil {
		updates = append(updates, "position_id = ?")
		args = append(args, *updateData.PositionID)
		argID++
	}
	// Добавлено обновление organizational_unit_id
	if updateData.OrganizationalUnitID != nil {
		updates = append(updates, "organizational_unit_id = ?")
		args = append(args, *updateData.OrganizationalUnitID)
		argID++
		argID++
	}

	if len(updates) == 0 {
		return errors.New("нет полей для обновления")
	}

	// Добавляем updated_at и формируем запрос
	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	query += strings.Join(updates, ", ")
	query += " WHERE id = ?"
	args = append(args, userID)

	// Выполняем запрос (плейсхолдеры '?' подходят для MySQL)
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
			p.name AS position_name, -- Добавляем название должности
			vl.total_days,    -- Лимит дней
			ou.name AS organizational_unit_name -- Добавляем название юнита
		FROM users u
		LEFT JOIN organizational_units ou ON u.organizational_unit_id = ou.id -- Добавляем JOIN для юнита
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
		var unitName sql.NullString     // <--- Добавлено: объявление переменной для имени юнита

		// Сканируем ID, ФИО, Название должности, Лимит дней
		if err := rows.Scan(
			&userDTO.ID,
			&userDTO.FullName,
			&positionName, // Сканируем название должности
			&limitDays,
			&unitName, // Сканируем название юнита
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

		// <--- Добавлено: Добавляем имя юнита в DTO
		if unitName.Valid {
			name := unitName.String
			userDTO.OrganizationalUnitName = &name
		} else {
			userDTO.OrganizationalUnitName = nil
		}
		// --->

		usersWithLimits = append(usersWithLimits, userDTO)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по пользователям с лимитами: %w", err)
	}

	// Удален дублирующийся блок кода

	return usersWithLimits, nil
}

// GetAllPositionsGrouped удален, так как группы должностей больше не используются

// GetAllPositions получает плоский список всех должностей из БД
func (r *UserRepository) GetAllPositions() ([]models.Position, error) {
	query := `SELECT id, name FROM positions ORDER BY name` // Просто получаем ID и имя, сортируем по имени

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса списка должностей: %w", err)
	}
	defer rows.Close()

	var positions []models.Position
	for rows.Next() {
		var pos models.Position
		if err := rows.Scan(&pos.ID, &pos.Name); err != nil {
			// Можно логировать ошибку сканирования
			// log.Printf("Ошибка сканирования должности: %v", err)
			continue // Пропускаем строку с ошибкой
		}
		positions = append(positions, pos)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по списку должностей: %w", err)
	}

	return positions, nil
}

// GetUserProfileByID получает данные пользователя для профиля, включая иерархию орг. юнитов
func (r *UserRepository) GetUserProfileByID(userID int) (*models.UserProfileDTO, error) {
	// 1. Получаем основные данные пользователя и ID его юнита + имя должности
	queryUser := `
		SELECT
			u.id, u.login, u.full_name, u.organizational_unit_id, 
			p.name AS position_name,
			u.is_admin, u.is_manager, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN positions p ON u.position_id = p.id
		WHERE u.id = ?`

	row := r.db.QueryRow(queryUser, userID)
	profile := &models.UserProfileDTO{}
	var unitID sql.NullInt64
	var positionName sql.NullString

	err := row.Scan(
		&profile.ID, &profile.Login, &profile.FullName, &unitID,
		&positionName,
		&profile.IsAdmin, &profile.IsManager, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Пользователь не найден
		}
		return nil, fmt.Errorf("ошибка получения основных данных пользователя ID %d: %w", userID, err)
	}

	// Устанавливаем имя должности
	if positionName.Valid {
		profile.PositionName = &positionName.String
	} else {
		profile.PositionName = nil
	}

	// 2. Если у пользователя есть юнит, получаем иерархию
	if unitID.Valid {
		currentUnitID := int(unitID.Int64)
		unitHierarchy := make([]string, 0, 3) // Предполагаем макс. 3 уровня для Department, SubDepartment, Sector

		queryUnit := `SELECT name, parent_id, unit_type FROM organizational_units WHERE id = ?`

		for currentUnitID != 0 { // Цикл до корневого элемента (или ошибки)
			var unitName string
			var parentID sql.NullInt64
			var unitType string // Пока не используем unit_type для назначения, но можем получить

			err := r.db.QueryRow(queryUnit, currentUnitID).Scan(&unitName, &parentID, &unitType)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					// Юнит не найден, это странно, если currentUnitID валиден
					// log.Printf("Предупреждение: Юнит с ID %d не найден при построении иерархии для пользователя %d", currentUnitID, userID)
					break // Прерываем цикл
				}
				return nil, fmt.Errorf("ошибка получения юнита ID %d для пользователя %d: %w", currentUnitID, userID, err)
			}

			unitHierarchy = append(unitHierarchy, unitName) // Добавляем имя текущего юнита

			if parentID.Valid {
				currentUnitID = int(parentID.Int64)
			} else {
				currentUnitID = 0 // Дошли до корня
			}
		}

		// 3. Заполняем поля DTO в обратном порядке (от корня к листу)
		numLevels := len(unitHierarchy)
		if numLevels > 0 {
			profile.Department = &unitHierarchy[numLevels-1] // Последний элемент - самый верхний (Department)
		}
		if numLevels > 1 {
			profile.SubDepartment = &unitHierarchy[numLevels-2] // Предпоследний - средний (SubDepartment)
		}
		if numLevels > 2 {
			profile.Sector = &unitHierarchy[numLevels-3] // Третий с конца - самый нижний (Sector)
			// Если уровней больше 3, самые нижние будут в Sector. Можно изменить логику при необходимости.
			// Или можно было бы ориентироваться на unit_type, если он строго задан ('DEPARTMENT', 'SUB_DEPARTMENT', 'SECTOR').
		}
		// Если уровней меньше, соответствующие поля останутся nil, что корректно
	}

	return profile, nil
}

// FindByOrganizationalUnitID находит всех пользователей, принадлежащих указанному орг. юниту
func (r *UserRepository) FindByOrganizationalUnitID(unitID int) ([]*models.User, error) {
	// Запрос выбирает пользователей и их должности
	query := `
		SELECT
			u.id, u.login, u.password, u.full_name, 
			u.organizational_unit_id, u.position_id, p.name AS position_name, 
			u.is_admin, u.is_manager, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN positions p ON u.position_id = p.id
		WHERE u.organizational_unit_id = ?
		ORDER BY u.full_name ASC` // Сортируем для консистентности

	rows, err := r.db.Query(query, unitID)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса пользователей по ID орг. юнита %d: %w", unitID, err)
	}
	defer rows.Close()

	users := []*models.User{}
	for rows.Next() {
		user := &models.User{}
		var organizationalUnitID sql.NullInt64 // Используем NullInt64
		var positionID sql.NullInt64
		var positionName sql.NullString

		err := rows.Scan(
			&user.ID, &user.Login, &user.Password, &user.FullName,
			&organizationalUnitID, // Сканируем в nullable типы
			&positionID,
			&positionName,
			&user.IsAdmin, &user.IsManager, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			// log.Printf("Ошибка сканирования пользователя для юнита %d: %v", unitID, err)
			continue // Пропускаем пользователя с ошибкой
		}

		// Устанавливаем ID юнита (хотя он должен быть равен unitID)
		if organizationalUnitID.Valid {
			uID := int(organizationalUnitID.Int64)
			user.OrganizationalUnitID = &uID
		} else {
			user.OrganizationalUnitID = nil // Маловероятно, но для полноты
		}

		// Устанавливаем ID должности
		if positionID.Valid {
			posID := int(positionID.Int64)
			user.PositionID = &posID
		} else {
			user.PositionID = nil
		}

		// Устанавливаем название должности
		if positionName.Valid {
			user.PositionName = &positionName.String
		} else {
			user.PositionName = nil
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по пользователям для юнита %d: %w", unitID, err)
	}

	return users, nil
}

// GetUsersWithLimitsByOrganizationalUnit получает пользователей конкретного юнита с их лимитами на год
func (r *UserRepository) GetUsersWithLimitsByOrganizationalUnit(unitID int, year int) ([]models.UserWithLimitAdminDTO, error) {
	query := `
		SELECT
			u.id,
			u.full_name,
			p.name AS position_name,
			vl.total_days
		FROM users u
		LEFT JOIN positions p ON u.position_id = p.id
		LEFT JOIN vacation_limits vl ON u.id = vl.user_id AND vl.year = ?
		WHERE u.organizational_unit_id = ?
		ORDER BY u.full_name ASC`

	rows, err := r.db.Query(query, year, unitID)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса пользователей юнита %d с лимитами на год %d: %w", unitID, year, err)
	}
	defer rows.Close()

	var usersWithLimits []models.UserWithLimitAdminDTO
	for rows.Next() {
		var userDTO models.UserWithLimitAdminDTO
		var positionName sql.NullString
		var totalDays sql.NullInt64

		if err := rows.Scan(&userDTO.ID, &userDTO.FullName, &positionName, &totalDays); err != nil {
			// log.Printf("Ошибка сканирования пользователя юнита %d с лимитом: %v", unitID, err)
			continue // Пропускаем пользователя с ошибкой
		}

		// Устанавливаем PositionName
		if positionName.Valid {
			name := positionName.String
			userDTO.PositionName = &name
		} else {
			userDTO.PositionName = nil
		}

		// Устанавливаем TotalDays
		if totalDays.Valid {
			days := int(totalDays.Int64)
			userDTO.TotalDays = &days
		} else {
			// Если лимита нет, можно оставить nil или установить дефолтное значение (например, 0 или 28, если требуется)
			// Оставляем nil, чтобы фронтенд мог решить, как это отображать
			userDTO.TotalDays = nil
		}

		usersWithLimits = append(usersWithLimits, userDTO)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по пользователям юнита %d с лимитами: %w", unitID, err)
	}

	return usersWithLimits, nil
}

// GetUsersByUnitIDs получает список пользователей для заданного списка ID организационных юнитов
func (r *UserRepository) GetUsersByUnitIDs(unitIDs []int) ([]models.User, error) {
	if len(unitIDs) == 0 {
		return []models.User{}, nil // Возвращаем пустой срез, если ID юнитов не переданы
	}

	// Генерируем плейсхолдеры (?, ?, ...) для IN клаузы
	placeholders := sqlRepeatParams(len(unitIDs))
	query := fmt.Sprintf(`
		SELECT
			u.id, u.login, u.full_name, u.organizational_unit_id, u.position_id,
			p.name AS position_name, u.is_admin, u.is_manager
		FROM users u
		LEFT JOIN positions p ON u.position_id = p.id
		WHERE u.organizational_unit_id IN (?%s)
		ORDER BY u.full_name ASC`, placeholders) // Добавлен LEFT JOIN для должности и сортировка

	// Создаем срез аргументов []interface{} для передачи в Query
	args := make([]interface{}, len(unitIDs))
	for i, id := range unitIDs {
		args[i] = id
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса пользователей по списку ID юнитов: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var organizationalUnitID sql.NullInt64
		var positionID sql.NullInt64
		var positionName sql.NullString

		// Сканируем данные пользователя
		if err := rows.Scan(
			&user.ID, &user.Login, &user.FullName, &organizationalUnitID, &positionID,
			&positionName, &user.IsAdmin, &user.IsManager,
		); err != nil {
			// log.Printf("Ошибка сканирования пользователя при запросе по списку юнитов: %v", err)
			continue // Пропускаем пользователя с ошибкой
		}

		// Устанавливаем ID юнита
		if organizationalUnitID.Valid {
			unitID := int(organizationalUnitID.Int64)
			user.OrganizationalUnitID = &unitID
		} else {
			user.OrganizationalUnitID = nil
		}

		// Устанавливаем ID должности
		if positionID.Valid {
			posID := int(positionID.Int64)
			user.PositionID = &posID
		} else {
			user.PositionID = nil
		}

		// Устанавливаем название должности
		if positionName.Valid {
			name := positionName.String
			user.PositionName = &name
		} else {
			user.PositionName = nil
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по пользователям для списка юнитов: %w", err)
	}

	return users, nil
}

// TODO: Добавить репозиторий и методы для работы с organizational_units (CRUD, получение дерева)
