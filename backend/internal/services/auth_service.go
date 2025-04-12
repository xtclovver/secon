package services

import (
	"errors"
	"fmt"  // Добавлен для форматирования ошибок
	"time" // Раскомментирован для генерации JWT

	"vacation-scheduler/internal/models"
	"vacation-scheduler/internal/repositories" // Импортируем репозиторий

	"github.com/golang-jwt/jwt/v5" // Раскомментирован для генерации JWT
	"golang.org/x/crypto/bcrypt"   // Раскомментирован для проверки пароля
)

// AuthService предоставляет методы для аутентификации пользователей
type AuthService struct {
	userRepo repositories.UserRepositoryInterface // Используем интерфейс пользователя
	// Используем интерфейс, определенный в repositories/vacation_repository.go (или где он должен быть)
	vacationRepo repositories.VacationRepositoryInterface
	jwtSecret    string // Секрет для JWT
}

// NewAuthService создает новый экземпляр AuthService
// Принимаем интерфейсы репозиториев
func NewAuthService(userRepo repositories.UserRepositoryInterface, vacationRepo repositories.VacationRepositoryInterface, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		vacationRepo: vacationRepo,
		jwtSecret:    jwtSecret,
	}
}

// Login проверяет учетные данные пользователя и возвращает JWT токен
func (s *AuthService) Login(login, password string) (string, *models.User, error) {
	user, err := s.userRepo.FindByLogin(login)
	if err != nil {
		return "", nil, errors.New("ошибка при поиске пользователя")
	}
	if user == nil {
		return "", nil, errors.New("неверный логин или пароль")
	}
	if user.Login != login { // Проверка регистра
		return "", nil, errors.New("неверный логин или пароль")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, errors.New("неверный логин или пароль")
	}

	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"login":      user.Login,
		"is_admin":   user.IsAdmin,
		"is_manager": user.IsManager,
		"exp":        time.Now().Add(time.Hour * 72).Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, errors.New("внутренняя ошибка сервера при генерации токена")
	}

	user.Password = "" // Очищаем пароль
	return tokenString, user, nil
}

// ValidateToken проверяет валидность токена
func (s *AuthService) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неожиданный метод подписи токена")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, errors.New("невалидный токен")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if expFloat, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(expFloat) {
				return nil, errors.New("срок действия токена истек")
			}
		} else {
			return nil, errors.New("некорректный формат срока действия токена")
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return nil, errors.New("некорректный формат ID пользователя в токене")
		}
		userID := int(userIDFloat)

		user, err := s.userRepo.FindByID(userID)
		if err != nil || user == nil {
			return nil, errors.New("пользователь из токена не найден")
		}
		user.Password = ""
		return user, nil
	}

	return nil, errors.New("невалидный токен")
}

// Register создает нового пользователя
func (s *AuthService) Register(login, password, fullName string, positionID *int, organizationalUnitID *int) (*models.User, error) { // Удален параметр email, Добавлен organizationalUnitID
	existingUser, err := s.userRepo.FindByLogin(login)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существующего пользователя: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("пользователь с таким логином уже существует")
	}

	newUser := &models.User{
		Login:    login,
		Password: password,
		FullName: fullName,
		// Email:                email, // Удалено
		PositionID:           positionID,
		OrganizationalUnitID: organizationalUnitID, // Добавлено присваивание
		IsAdmin:              false,
		IsManager:            false,
	}

	err = s.userRepo.CreateUser(newUser)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя в репозитории: %w", err)
	}

	const defaultVacationLimit = 28
	currentYear := time.Now().Year()
	// Используем интерфейс repositories.VacationRepositoryInterface, переданный в конструкторе
	errLimit := s.vacationRepo.CreateOrUpdateVacationLimit(newUser.ID, currentYear, defaultVacationLimit)
	if errLimit != nil {
		fmt.Printf("ВНИМАНИЕ: Пользователь %d создан, но не удалось установить начальный лимит отпуска (%d дней на %d год): %v\n", newUser.ID, defaultVacationLimit, currentYear, errLimit)
	}

	return newUser, nil
}
