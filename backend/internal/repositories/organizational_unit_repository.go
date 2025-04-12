package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	// "strings" // Удален неиспользуемый импорт
	"vacation-scheduler/internal/models"
)

// OrganizationalUnitRepositoryInterface определяет методы для работы с орг. юнитами
type OrganizationalUnitRepositoryInterface interface {
	Create(unit *models.OrganizationalUnit) (int, error)
	GetByID(id int) (*models.OrganizationalUnit, error)
	Update(unit *models.OrganizationalUnit) error
	Delete(id int) error
	GetAll() ([]*models.OrganizationalUnit, error)
	GetTree() ([]*models.OrganizationalUnit, error)                     // Метод для получения всей иерархии деревом
	GetSubtreeIDs(unitID int) ([]int, error)                            // Метод для получения ID юнита и всех его дочерних юнитов
	FindByParentID(parentID *int) ([]*models.OrganizationalUnit, error) // Ищет прямых потомков (юниты)
	// TODO: Добавить методы для поиска, получения пользователей юнита/поддерева и т.д., если нужно
}

// OrganizationalUnitRepository реализует интерфейс
type OrganizationalUnitRepository struct {
	db *sql.DB
}

// NewOrganizationalUnitRepository создает новый репозиторий
func NewOrganizationalUnitRepository(db *sql.DB) *OrganizationalUnitRepository {
	return &OrganizationalUnitRepository{db: db}
}

// Create создает новый организационный юнит
func (r *OrganizationalUnitRepository) Create(unit *models.OrganizationalUnit) (int, error) {
	query := `
		INSERT INTO organizational_units (name, unit_type, parent_id, manager_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := r.db.Exec(query, unit.Name, unit.UnitType, unit.ParentID, unit.ManagerID)
	if err != nil {
		// TODO: Обработка специфических ошибок БД, например, неверный parent_id или manager_id
		return 0, fmt.Errorf("ошибка создания орг. юнита: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID нового орг. юнита: %w", err)
	}

	return int(id), nil
}

// GetByID получает орг. юнит по ID
func (r *OrganizationalUnitRepository) GetByID(id int) (*models.OrganizationalUnit, error) {
	query := `
		SELECT id, name, unit_type, parent_id, manager_id, created_at, updated_at
		FROM organizational_units
		WHERE id = ?`
	row := r.db.QueryRow(query, id)

	unit := &models.OrganizationalUnit{}
	var parentID sql.NullInt64
	var managerID sql.NullInt64

	err := row.Scan(
		&unit.ID, &unit.Name, &unit.UnitType,
		&parentID, &managerID,
		&unit.CreatedAt, &unit.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Не найдено
		}
		return nil, fmt.Errorf("ошибка получения орг. юнита по ID %d: %w", id, err)
	}

	if parentID.Valid {
		pID := int(parentID.Int64)
		unit.ParentID = &pID
	}
	if managerID.Valid {
		mID := int(managerID.Int64)
		unit.ManagerID = &mID
	}

	return unit, nil
}

// Update обновляет существующий орг. юнит
func (r *OrganizationalUnitRepository) Update(unit *models.OrganizationalUnit) error {
	query := `
		UPDATE organizational_units
		SET name = ?, unit_type = ?, parent_id = ?, manager_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	result, err := r.db.Exec(query, unit.Name, unit.UnitType, unit.ParentID, unit.ManagerID, unit.ID)
	if err != nil {
		// TODO: Обработка специфических ошибок БД
		return fmt.Errorf("ошибка обновления орг. юнита ID %d: %w", unit.ID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения кол-ва строк при обновлении орг. юнита ID %d: %w", unit.ID, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("орг. юнит ID %d не найден для обновления", unit.ID)
	}

	return nil
}

// Delete удаляет орг. юнит по ID
func (r *OrganizationalUnitRepository) Delete(id int) error {
	// TODO: Подумать о каскадном удалении или запрете удаления, если есть дочерние юниты или пользователи
	// Пока просто удаляем
	query := `DELETE FROM organizational_units WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления орг. юнита ID %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения кол-ва строк при удалении орг. юнита ID %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("орг. юнит ID %d не найден для удаления", id)
	}

	return nil
}

// GetAll получает плоский список всех орг. юнитов
func (r *OrganizationalUnitRepository) GetAll() ([]*models.OrganizationalUnit, error) {
	query := `
		SELECT id, name, unit_type, parent_id, manager_id, created_at, updated_at
		FROM organizational_units
		ORDER BY name ASC` // Или по другому полю

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех орг. юнитов: %w", err)
	}
	defer rows.Close()

	units := []*models.OrganizationalUnit{}
	for rows.Next() {
		unit := &models.OrganizationalUnit{}
		var parentID sql.NullInt64
		var managerID sql.NullInt64

		err := rows.Scan(
			&unit.ID, &unit.Name, &unit.UnitType,
			&parentID, &managerID,
			&unit.CreatedAt, &unit.UpdatedAt,
		)
		if err != nil {
			log.Printf("Ошибка сканирования орг. юнита: %v", err)
			continue // Пропускаем ошибочный юнит
		}

		if parentID.Valid {
			pID := int(parentID.Int64)
			unit.ParentID = &pID
		}
		if managerID.Valid {
			mID := int(managerID.Int64)
			unit.ManagerID = &mID
		}
		units = append(units, unit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по орг. юнитам: %w", err)
	}

	return units, nil
}

// GetTree строит иерархическое дерево орг. юнитов
func (r *OrganizationalUnitRepository) GetTree() ([]*models.OrganizationalUnit, error) {
	units, err := r.GetAll() // Получаем плоский список
	if err != nil {
		return nil, fmt.Errorf("ошибка получения плоского списка юнитов для дерева: %w", err)
	}

	if len(units) == 0 {
		return []*models.OrganizationalUnit{}, nil
	}

	// Карта для быстрого поиска юнитов по ID
	unitsMap := make(map[int]*models.OrganizationalUnit, len(units))
	for _, unit := range units {
		unitsMap[unit.ID] = unit
		unit.Children = []*models.OrganizationalUnit{} // Инициализируем слайс детей
	}

	// Строим дерево
	var rootUnits []*models.OrganizationalUnit
	for _, unit := range units {
		if unit.ParentID == nil {
			// Это корневой юнит
			rootUnits = append(rootUnits, unit)
		} else {
			// Ищем родителя
			if parent, ok := unitsMap[*unit.ParentID]; ok {
				parent.Children = append(parent.Children, unit)
			} else {
				// Родитель не найден (осиротевший юнит?) - можно логировать или добавить в корень
				log.Printf("Warning: Parent unit with ID %d not found for unit ID %d", *unit.ParentID, unit.ID)
				// Добавим его пока в корень для отображения
				rootUnits = append(rootUnits, unit)
			}
		}
	}

	// TODO: Можно добавить сортировку детей на каждом уровне, если нужно

	return rootUnits, nil
}

// GetSubtreeIDs возвращает ID переданного юнита и всех его дочерних юнитов
func (r *OrganizationalUnitRepository) GetSubtreeIDs(unitID int) ([]int, error) {
	allUnits, err := r.GetAll() // Получаем все юниты
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех юнитов для построения поддерева: %w", err)
	}

	if len(allUnits) == 0 {
		return []int{}, nil
	}

	// Строим карту parentID -> []childID для быстрого обхода
	childrenMap := make(map[int][]int)
	unitExists := false
	for _, unit := range allUnits {
		if unit.ID == unitID {
			unitExists = true
		}
		if unit.ParentID != nil {
			childrenMap[*unit.ParentID] = append(childrenMap[*unit.ParentID], unit.ID)
		}
	}

	if !unitExists {
		return nil, fmt.Errorf("юнит с ID %d не найден", unitID)
	}

	// Рекурсивная функция для сбора ID
	var subtreeIDs []int
	var collectIDs func(currentUnitID int)
	collectIDs = func(currentUnitID int) {
		subtreeIDs = append(subtreeIDs, currentUnitID) // Добавляем текущий ID
		// Добавляем всех детей рекурсивно
		if children, ok := childrenMap[currentUnitID]; ok {
			for _, childID := range children {
				collectIDs(childID)
			}
		}
	}

	// Начинаем сбор с переданного unitID
	collectIDs(unitID)

	return subtreeIDs, nil
}

// FindByParentID находит все дочерние юниты для заданного parentID
// Если parentID = nil, возвращает корневые юниты (parent_id IS NULL)
func (r *OrganizationalUnitRepository) FindByParentID(parentID *int) ([]*models.OrganizationalUnit, error) {
	var rows *sql.Rows
	var err error
	query := `
		SELECT id, name, unit_type, parent_id, manager_id, created_at, updated_at
		FROM organizational_units `

	if parentID == nil {
		query += "WHERE parent_id IS NULL ORDER BY name ASC"
		rows, err = r.db.Query(query)
	} else {
		query += "WHERE parent_id = ? ORDER BY name ASC"
		rows, err = r.db.Query(query, *parentID)
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка запроса дочерних орг. юнитов для parentID %v: %w", parentID, err)
	}
	defer rows.Close()

	units := []*models.OrganizationalUnit{}
	for rows.Next() {
		unit := &models.OrganizationalUnit{}
		var pID sql.NullInt64 // Используем sql.NullInt64 для parent_id
		var mID sql.NullInt64 // Используем sql.NullInt64 для manager_id

		scanErr := rows.Scan(
			&unit.ID, &unit.Name, &unit.UnitType,
			&pID, &mID,
			&unit.CreatedAt, &unit.UpdatedAt,
		)
		if scanErr != nil {
			log.Printf("Ошибка сканирования дочернего орг. юнита: %v", scanErr)
			continue // Пропускаем ошибочный юнит
		}

		if pID.Valid {
			parentIntValue := int(pID.Int64)
			unit.ParentID = &parentIntValue
		} else {
			unit.ParentID = nil // Явно устанавливаем nil, если в БД NULL
		}
		if mID.Valid {
			managerIntValue := int(mID.Int64)
			unit.ManagerID = &managerIntValue
		} else {
			unit.ManagerID = nil // Явно устанавливаем nil, если в БД NULL
		}
		units = append(units, unit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по дочерним орг. юнитам для parentID %v: %w", parentID, err)
	}

	return units, nil
}

// --- Вспомогательные функции, если потребуются ---
