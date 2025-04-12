package services

import (
	"fmt"
	"vacation-scheduler/internal/models"
	"vacation-scheduler/internal/repositories"
)

// OrganizationalUnitServiceInterface определяет методы для сервиса орг. юнитов
type OrganizationalUnitServiceInterface interface {
	CreateUnit(unit *models.OrganizationalUnit) (*models.OrganizationalUnit, error)
	GetUnitByID(id int) (*models.OrganizationalUnit, error)
	UpdateUnit(id int, updateData *models.OrganizationalUnit) (*models.OrganizationalUnit, error)
	DeleteUnit(id int) error
	GetUnitTree() ([]*models.OrganizationalUnit, error)
	GetUnitChildrenAndUsers(parentUnitID *int) ([]models.UnitListItemDTO, error) // Новый метод
}

// OrganizationalUnitService реализует интерфейс
type OrganizationalUnitService struct {
	unitRepo repositories.OrganizationalUnitRepositoryInterface
	userRepo repositories.UserRepositoryInterface // Может понадобиться для проверки manager_id
}

// NewOrganizationalUnitService создает новый сервис
func NewOrganizationalUnitService(unitRepo repositories.OrganizationalUnitRepositoryInterface, userRepo repositories.UserRepositoryInterface) *OrganizationalUnitService {
	return &OrganizationalUnitService{
		unitRepo: unitRepo,
		userRepo: userRepo,
	}
}

// CreateUnit создает новый орг. юнит с валидацией
func (s *OrganizationalUnitService) CreateUnit(unit *models.OrganizationalUnit) (*models.OrganizationalUnit, error) {
	// Валидация входных данных
	if unit.Name == "" {
		return nil, fmt.Errorf("название юнита не может быть пустым")
	}
	if unit.UnitType == "" {
		return nil, fmt.Errorf("тип юнита не может быть пустым")
	}

	// Проверка существования parent_id, если он указан
	if unit.ParentID != nil {
		parent, err := s.unitRepo.GetByID(*unit.ParentID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки родительского юнита: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("родительский юнит с ID %d не найден", *unit.ParentID)
		}
	}

	// Проверка существования manager_id, если он указан
	if unit.ManagerID != nil {
		manager, err := s.userRepo.FindByID(*unit.ManagerID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки пользователя-менеджера: %w", err)
		}
		if manager == nil {
			return nil, fmt.Errorf("пользователь (менеджер) с ID %d не найден", *unit.ManagerID)
		}
		// TODO: Можно добавить проверку, является ли пользователь действительно менеджером (manager.IsManager)
	}

	// Создание юнита через репозиторий
	id, err := s.unitRepo.Create(unit)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения нового орг. юнита: %w", err)
	}

	// Возвращаем созданный юнит (или получаем его снова по ID)
	createdUnit, err := s.unitRepo.GetByID(id)
	if err != nil {
		// Логируем ошибку, но возвращаем ID, т.к. юнит создан
		fmt.Printf("Warning: юнит создан (ID: %d), но не удалось получить его после создания: %v\n", id, err)
		unit.ID = id // Присваиваем ID исходному объекту
		return unit, nil
	}
	return createdUnit, nil
}

// GetUnitByID получает орг. юнит по ID
func (s *OrganizationalUnitService) GetUnitByID(id int) (*models.OrganizationalUnit, error) {
	unit, err := s.unitRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения орг. юнита ID %d из репозитория: %w", id, err)
	}
	// Обработка случая "не найдено" на уровне сервиса не требуется, репозиторий вернет nil, nil
	return unit, nil
}

// UpdateUnit обновляет орг. юнит
func (s *OrganizationalUnitService) UpdateUnit(id int, updateData *models.OrganizationalUnit) (*models.OrganizationalUnit, error) {
	// Получаем существующий юнит
	existingUnit, err := s.unitRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения орг. юнита ID %d для обновления: %w", id, err)
	}
	if existingUnit == nil {
		return nil, fmt.Errorf("орг. юнит ID %d не найден для обновления", id)
	}

	// Применяем изменения из updateData
	if updateData.Name != "" {
		existingUnit.Name = updateData.Name
	}
	if updateData.UnitType != "" {
		existingUnit.UnitType = updateData.UnitType
	}
	// Обновляем ParentID (позволяем установить в nil)
	existingUnit.ParentID = updateData.ParentID
	// Обновляем ManagerID (позволяем установить в nil)
	existingUnit.ManagerID = updateData.ManagerID

	// Валидация обновленных данных
	if existingUnit.ParentID != nil {
		// Предотвращаем установку себя в качестве родителя
		if *existingUnit.ParentID == existingUnit.ID {
			return nil, fmt.Errorf("орг. юнит не может быть родителем для самого себя")
		}
		parent, err := s.unitRepo.GetByID(*existingUnit.ParentID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки нового родительского юнита: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("новый родительский юнит с ID %d не найден", *existingUnit.ParentID)
		}
		// TODO: Добавить проверку на циклические зависимости (A -> B -> C -> A)
	}
	if existingUnit.ManagerID != nil {
		manager, err := s.userRepo.FindByID(*existingUnit.ManagerID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки нового пользователя-менеджера: %w", err)
		}
		if manager == nil {
			return nil, fmt.Errorf("новый пользователь (менеджер) с ID %d не найден", *existingUnit.ManagerID)
		}
	}

	// Обновляем в репозитории
	err = s.unitRepo.Update(existingUnit)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления орг. юнита ID %d в репозитории: %w", id, err)
	}

	return existingUnit, nil
}

// DeleteUnit удаляет орг. юнит
func (s *OrganizationalUnitService) DeleteUnit(id int) error {
	// TODO: Добавить бизнес-логику перед удалением:
	// - Проверить, есть ли дочерние юниты (если есть, запретить или переназначить родителя)
	// - Проверить, есть ли пользователи в этом юните (если есть, запретить или переназначить юнит)

	// Получаем юнит, чтобы убедиться, что он существует
	unit, err := s.unitRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("ошибка проверки орг. юнита ID %d перед удалением: %w", id, err)
	}
	if unit == nil {
		return fmt.Errorf("орг. юнит ID %d не найден для удаления", id)
	}

	// Удаляем через репозиторий
	err = s.unitRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("ошибка удаления орг. юнита ID %d в репозитории: %w", id, err)
	}
	return nil
}

// GetUnitTree получает иерархическое дерево орг. юнитов
func (s *OrganizationalUnitService) GetUnitTree() ([]*models.OrganizationalUnit, error) {
	tree, err := s.unitRepo.GetTree()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения дерева орг. юнитов: %w", err)
	}
	// TODO: Можно добавить логику по заполнению Users и Positions для каждого юнита, если это нужно здесь
	return tree, nil
}

// GetUnitChildrenAndUsers получает список дочерних юнитов и пользователей для заданного родительского ID
func (s *OrganizationalUnitService) GetUnitChildrenAndUsers(parentUnitID *int) ([]models.UnitListItemDTO, error) {
	var listItems []models.UnitListItemDTO

	// 1. Получаем дочерние юниты
	childUnits, err := s.unitRepo.FindByParentID(parentUnitID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения дочерних юнитов для parentID %v: %w", parentUnitID, err)
	}
	for _, unit := range childUnits {
		listItems = append(listItems, models.UnitListItemDTO{
			ID:       unit.ID,
			Name:     unit.Name,
			Type:     "unit",
			UnitType: &unit.UnitType, // Добавляем конкретный тип юнита
		})
	}

	// 2. Получаем пользователей для текущего юнита (только если parentUnitID не nil, пользователи не могут быть в корне)
	//    И только если этот юнит МОЖЕТ содержать пользователей (нужна логика определения, например, по UnitType)
	//    Пока для простоты получаем пользователей для любого НЕ корневого юнита.
	if parentUnitID != nil {
		// Возможно, стоит проверить тип юнита *parentUnitID, чтобы определить, можно ли в нем содержать пользователей
		// currentUnit, _ := s.GetUnitByID(*parentUnitID)
		// if currentUnit != nil && canContainUsers(currentUnit.UnitType) { ... }

		users, err := s.userRepo.FindByOrganizationalUnitID(*parentUnitID)
		if err != nil {
			// Не фатальная ошибка, просто может не быть пользователей или ошибка БД
			fmt.Printf("Warning: ошибка получения пользователей для юнита ID %d: %v\n", *parentUnitID, err)
			// Не прерываем выполнение, просто не будет пользователей в списке
		} else {
			for _, user := range users {
				// У пользователя может не быть должности (PositionName может быть nil)
				listItems = append(listItems, models.UnitListItemDTO{
					ID:       user.ID,
					Name:     user.FullName,
					Type:     "user",
					Position: user.PositionName, // Добавляем должность пользователя (может быть nil)
				})
			}
		}
	}

	// TODO: Возможно, нужна сортировка смешанного списка (например, сначала юниты, потом пользователи, или по имени)

	return listItems, nil
}
