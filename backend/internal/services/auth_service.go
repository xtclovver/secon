package services

import (
	"errors"
	"time" // Раскомментирован для генерации JWT

	"vacation-scheduler/internal/models" 
	"vacation-scheduler/internal/repositories" // Импортируем репозиторий

	"github.com/golang-jwt/jwt/v5" // Раскомментирован для генерации JWT
	"golang.org/x/crypto/bcrypt" // Раскомментирован для проверки пароля
)

// AuthService предоставляет методы для аутентификации пользователей
type AuthService struct {
	userRepo *repositories.UserRepository // Зависимость от репозитория пользователей
	jwtSecret string                     // Секрет для JWT
}

// NewAuthService создает новый экземпляр AuthService
func NewAuthService(userRepo *repositories.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtSecret: jwtSecret,
	}
}

// Login проверяет учетные данные пользователя и возвращает JWT токен (заглушка)
func (s *AuthService) Login(username, password string) (string, *models.User, error) {
	// 1. Найти пользователя по имени пользователя
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		// Ошибка при запросе к БД
		return "", nil, errors.New("ошибка при поиске пользователя")
	}
	if user == nil {
		// Пользователь не найден
		return "", nil, errors.New("неверное имя пользователя или пароль")
	}

	// 2. Сравнить хеш пароля из БД с предоставленным паролем
	// ЗАГЛУШКА: В реальном приложении здесь будет сравнение хешей
	// err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	// if err != nil {
	// 	// Пароль не совпадает
	// 	return "", nil, errors.New("неверное имя пользователя или пароль")
	// }
	
	// Сравниваем хеш пароля из БД с предоставленным паролем
	// Примечание: user.Password должен содержать хеш из БД (репозиторий-заглушка должен это имитировать)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Пароль не совпадает или другая ошибка bcrypt
		return "", nil, errors.New("неверное имя пользователя или пароль")
	}


	// 3. Сгенерировать JWT токен
	claims := jwt.MapClaims{
		"user_id":    user.ID, // Используем ID пользователя
		"username":   user.Username,
		"is_admin":   user.IsAdmin,
		"is_manager": user.IsManager,
		"exp":        time.Now().Add(time.Hour * 72).Unix(), // Токен действителен 72 часа
		"iat":        time.Now().Unix(),                     // Время создания токена
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		// Логирование ошибки генерации токена может быть полезно
		// log.Printf("Ошибка генерации JWT: %v", err)
		return "", nil, errors.New("внутренняя ошибка сервера при генерации токена")
	}

	// Убираем хеш пароля перед возвратом данных пользователя
	user.Password = ""

	return tokenString, user, nil
}

// ValidateToken проверяет валидность токена (заглушка)
// Этот метод может понадобиться для middleware или других проверок
func (s *AuthService) ValidateToken(tokenString string) (*models.User, error) {
	// ЗАГЛУШКА: Реальная проверка токена
	// Парсинг и валидация JWT токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неожиданный метод подписи токена")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		// Ошибка парсинга или невалидная подпись
		return nil, errors.New("невалидный токен")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Проверяем срок действия (exp)
		if expFloat, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(expFloat) {
				return nil, errors.New("срок действия токена истек")
			}
		} else {
			return nil, errors.New("некорректный формат срока действия токена")
		}

		// Извлекаем ID пользователя
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return nil, errors.New("некорректный формат ID пользователя в токене")
		}
		userID := int(userIDFloat)

		// Получаем пользователя по ID (используем метод репозитория)
		user, err := s.userRepo.FindByID(userID)
		if err != nil || user == nil {
			return nil, errors.New("пользователь из токена не найден")
		}
		user.Password = "" // Убираем пароль
		return user, nil
	}

	return nil, errors.New("невалидный токен")
}
