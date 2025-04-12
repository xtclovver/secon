package services

import (
	"fmt"
	"vacation-scheduler/internal/models"
	"vacation-scheduler/internal/repositories"
)

// UserServiceInterface определяет методы для сервиса пользователей
type UserServiceInterface interface {
	GetAllUsersWithLimits(year int) ([]models.UserWithLimitDTO, error)
	GetAllPositionsGrouped() ([]models.PositionGroup, error) // Добавлен метод для получения должностей
	// TODO: Добавить другие методы сервиса пользователей по мере необходимости
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

// GetAllPositionsGrouped получает список всех должностей, сгруппированных по категориям
func (s *UserService) GetAllPositionsGrouped() ([]models.PositionGroup, error) {
	positions, err := s.userRepo.GetAllPositionsGrouped()
	if err != nil {
		// Можно добавить логирование ошибки здесь
		return nil, fmt.Errorf("ошибка получения должностей из репозитория: %w", err)
	}
	// На данный момент дополнительной бизнес-логики нет
	return positions, nil
}

// TODO: Реализовать другие методы бизнес-логики для пользователей
// Например: GetUserByID, CreateUser, UpdateUser, ChangePassword и т.д.
