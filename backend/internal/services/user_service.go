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
	GetUserProfile(userID int) (*models.UserProfileDTO, error) // Новый метод для получения профиля
	// TODO: Добавить другие методы сервиса пользователей по мере необходимости (GetUserByID и т.д.)
}

// UserService реализует UserServiceInterface
type UserService struct {
	userRepo repositories.UserRepositoryInterface // Зависимость от интерфейса репозитория пользователей
	// Можно добавить другие зависимости, например, от репозитория лимитов, если нужно
}

// NewUserService создает новый экземпляр UserService
func NewUserService(userRepo repositories.UserRepositoryInterface) *UserService { // Принимаем интерфейс
	return &UserService{
		userRepo: userRepo, // Сохраняем интерфейс
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

// TODO: Реализовать другие методы бизнес-логики для пользователей
// Например: GetUserByID, CreateUser, ChangePassword и т.д.
