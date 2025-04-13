package services

import (
	"fmt"
	"vacation-scheduler/internal/models"
	"vacation-scheduler/internal/repositories"
)

// UserServiceInterface определяет методы для сервиса пользователей
type UserServiceInterface interface {
	GetAllUsersWithLimits(year int) ([]models.UserWithLimitDTO, error)
	// GetAllPositionsGrouped удален
	GetUsersByOrganizationalUnit(unitID int) ([]models.User, error) // Добавлен метод
	GetAllPositions() ([]models.Position, error)                    // Добавлен метод для получения должностей
	// UpdateUserProfile обновляет профиль пользователя с проверкой прав доступа
	UpdateUserProfile(requestingUser *models.User, targetUserID int, updateData *models.UserUpdateDTO) error
	GetUserProfile(userID int) (*models.UserProfileDTO, error)                                                  // Новый метод для получения профиля
	GetAllUsers() ([]models.UserProfileDTO, error)                                                              // Новый метод для получения всех пользователей (админ)
	UpdateUserAdmin(requestingUser *models.User, targetUserID int, updateData *models.UserUpdateAdminDTO) error // Новый метод для обновления админом
	FindByID(id int) (*models.User, error)                                                                      // Добавлен метод для поиска по ID
	// TODO: Добавить другие методы сервиса пользователей по мере необходимости
}

// UserService реализует UserServiceInterface
type UserService struct {
	userRepo repositories.UserRepositoryInterface
	unitRepo repositories.OrganizationalUnitRepositoryInterface // Добавлена зависимость от репозитория юнитов
	// TODO: Добавить зависимость от репозитория должностей, если он будет отдельным
}

// NewUserService создает новый экземпляр UserService
func NewUserService(userRepo repositories.UserRepositoryInterface, unitRepo repositories.OrganizationalUnitRepositoryInterface) *UserService { // Добавлен unitRepo
	return &UserService{
		userRepo: userRepo,
		unitRepo: unitRepo, // Сохраняем unitRepo
	}
}

// GetAllUsersWithLimits получает список всех пользователей с их лимитами на указанный год
func (s *UserService) GetAllUsersWithLimits(year int) ([]models.UserWithLimitDTO, error) {
	users, err := s.userRepo.GetAllUsersWithLimits(year)
	if err != nil {
		// Можно добавить логирование ошибки здесь
		return nil, fmt.Errorf("ошибка получения пользователей с лимитами из репозитория: %w", err)
	}
	// На данный момент дополнительной бизнес-логики нет, просто возвращаем результат репозитория.
	// В будущем здесь можно добавить проверки, фильтрацию и т.д.
	return users, nil
}

// GetUsersByOrganizationalUnit получает пользователей по ID орг. юнита
func (s *UserService) GetUsersByOrganizationalUnit(unitID int) ([]models.User, error) {
	return s.userRepo.GetUsersByOrganizationalUnit(unitID)
}

// GetAllPositions получает список всех должностей
func (s *UserService) GetAllPositions() ([]models.Position, error) {
	positions, err := s.userRepo.GetAllPositions()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения должностей из репозитория: %w", err)
	}
	// Пока просто возвращаем результат репозитория
	return positions, nil
}

// GetAllPositionsGrouped удален

// UpdateUserProfile обновляет профиль пользователя с проверкой прав доступа
func (s *UserService) UpdateUserProfile(requestingUser *models.User, targetUserID int, updateData *models.UserUpdateDTO) error {
	if requestingUser == nil {
		return fmt.Errorf("не удалось определить запрашивающего пользователя")
	}
	if updateData == nil {
		return fmt.Errorf("данные для обновления не предоставлены")
	}

	isSelfUpdate := requestingUser.ID == targetUserID
	canManageUsers := requestingUser.IsAdmin || requestingUser.IsManager

	// Проверка прав на обновление должности
	if updateData.PositionID != nil {
		if !canManageUsers {
			return fmt.Errorf("недостаточно прав для изменения должности пользователя")
		}
		// Дополнительно можно проверить, существует ли такая должность, но это лучше делать на уровне репозитория или БД
	}

	// Проверка прав на обновление ФИО и пароля
	if updateData.FullName != nil || (updateData.Password != nil && *updateData.Password != "") {
		if !isSelfUpdate && !canManageUsers {
			return fmt.Errorf("недостаточно прав для изменения данных другого пользователя")
		}
	}

	// Если обновляется только должность, а пользователь не админ/менеджер и не обновляет себя - это уже отсечено выше.
	// Если обновляется только ФИО/пароль, и пользователь не админ/менеджер, но обновляет себя - это разрешено.
	// Если обновляется только ФИО/пароль, и пользователь админ/менеджер - это разрешено.

	// Проверяем, есть ли вообще что обновлять (кроме PositionID, если его обновляет не админ/менеджер)
	hasUpdates := updateData.FullName != nil || (updateData.Password != nil && *updateData.Password != "")
	if updateData.PositionID != nil && canManageUsers {
		hasUpdates = true
	}
	// Добавлено обновление OrganizationalUnitID
	if updateData.OrganizationalUnitID != nil {
		if !canManageUsers { // Только админ/менеджер могут менять юнит
			return fmt.Errorf("недостаточно прав для изменения организационного юнита пользователя")
		}
		hasUpdates = true
	}

	if !hasUpdates {
		return fmt.Errorf("нет допустимых полей для обновления")
	}

	// Вызов репозитория для обновления
	err := s.userRepo.UpdateUser(targetUserID, updateData)
	if err != nil {
		// Можно добавить логирование ошибки здесь
		return fmt.Errorf("ошибка обновления пользователя в репозитории: %w", err)
	}

	return nil
}

// GetUserProfile получает профиль пользователя по ID, включая иерархию юнитов
func (s *UserService) GetUserProfile(userID int) (*models.UserProfileDTO, error) {
	profile, err := s.userRepo.GetUserProfileByID(userID)
	if err != nil {
		// Можно добавить логирование здесь
		return nil, fmt.Errorf("ошибка получения профиля пользователя ID %d из репозитория: %w", userID, err)
	}
	if profile == nil {
		// Репозиторий вернул nil без ошибки, значит пользователь не найден
		return nil, fmt.Errorf("пользователь с ID %d не найден", userID) // Возвращаем ошибку, а не nil, nil
	}
	// Дополнительная бизнес-логика, если нужна
	return profile, nil
}

// GetAllUsers получает список всех пользователей для админ-панели
func (s *UserService) GetAllUsers() ([]models.UserProfileDTO, error) {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		// Логирование ошибки может быть полезно
		return nil, fmt.Errorf("ошибка получения всех пользователей из репозитория: %w", err)
	}
	// Дополнительная бизнес-логика (фильтрация, обогащение данных) может быть добавлена здесь
	return users, nil
}

// UpdateUserAdmin обновляет данные пользователя от имени администратора
func (s *UserService) UpdateUserAdmin(requestingUser *models.User, targetUserID int, updateData *models.UserUpdateAdminDTO) error {
	if requestingUser == nil {
		return fmt.Errorf("не удалось определить запрашивающего пользователя")
	}
	if !requestingUser.IsAdmin {
		return fmt.Errorf("недостаточно прав для выполнения этой операции") // Только админ может использовать этот метод
	}
	if updateData == nil {
		return fmt.Errorf("данные для обновления не предоставлены")
	}

	// Проверяем, есть ли что обновлять (хотя бы одно поле не nil)
	hasUpdate := updateData.PositionID != nil || updateData.OrganizationalUnitID != nil || updateData.IsAdmin != nil || updateData.IsManager != nil
	if !hasUpdate {
		return fmt.Errorf("нет полей для обновления")
	}

	// Проверка существования целевого пользователя
	targetUser, err := s.userRepo.FindByID(targetUserID)
	if err != nil {
		return fmt.Errorf("ошибка проверки целевого пользователя ID %d: %w", targetUserID, err)
	}
	if targetUser == nil {
		return fmt.Errorf("целевой пользователь с ID %d не найден", targetUserID)
	}

	// Проверка существования Юнита, если он указан
	if updateData.OrganizationalUnitID != nil {
		unitID := *updateData.OrganizationalUnitID
		unit, err := s.unitRepo.GetByID(unitID) // Используем unitRepo
		if err != nil {
			return fmt.Errorf("ошибка проверки организационного юнита ID %d: %w", unitID, err)
		}
		if unit == nil {
			return fmt.Errorf("организационный юнит с ID %d не найден", unitID)
		}
	}

	// Проверка существования Должности, если она указана
	if updateData.PositionID != nil {
		positionID := *updateData.PositionID
		// Предполагаем, что у userRepo есть метод для проверки должности
		// TODO: Убедиться, что метод GetPositionByID существует и работает
		position, err := s.userRepo.GetPositionByID(positionID) // Используем userRepo (или отдельный repo)
		if err != nil {
			return fmt.Errorf("ошибка проверки должности ID %d: %w", positionID, err)
		}
		if position == nil {
			return fmt.Errorf("должность с ID %d не найдена", positionID)
		}
	}

	// Вызов репозитория для обновления
	err = s.userRepo.UpdateUserAdmin(targetUserID, updateData)
	if err != nil {
		// Логирование ошибки
		return fmt.Errorf("ошибка обновления пользователя (админ) в репозитории: %w", err)
	}

	return nil
}

// FindByID находит пользователя по его ID
func (s *UserService) FindByID(id int) (*models.User, error) {
	user, err := s.userRepo.FindByID(id) // Предполагаем, что у репозитория есть этот метод
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска пользователя ID %d в репозитории: %w", id, err)
	}
	// Репозиторий должен вернуть nil, nil если пользователь не найден,
	// поэтому дополнительная проверка на nil здесь не обязательна, если репозиторий следует этому контракту.
	return user, nil
}

// TODO: Реализовать другие методы бизнес-логики для пользователей
// Например: CreateUser, ChangePassword и т.д.
