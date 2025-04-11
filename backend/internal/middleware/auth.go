package middleware

import (
	"errors" // Добавлен импорт errors
	"fmt" // Добавлен для форматирования ошибок
	"net/http"
	"strings"
	"time" // Раскомментирован

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5" // Раскомментирован
	
	// Предполагаем, что AuthService будет доступен (например, через DI или глобально - не лучший вариант)
	// Для простоты примера, предположим, что у нас есть доступ к экземпляру AuthService
	// import "vacation-scheduler/internal/services" 
)

// JWTAuth - middleware для проверки JWT токена
// Примечание: Передача secretKey здесь может быть избыточна, если AuthService уже инициализирован с ним.
// Но оставим для совместимости с текущим main.go
func JWTAuth(secretKey string) gin.HandlerFunc { 
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует заголовок Authorization"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Некорректный формат заголовка Authorization"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Парсинг и валидация токена
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверяем метод подписи: убеждаемся, что это HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
			}
			// Возвращаем секретный ключ для проверки подписи
			return []byte(secretKey), nil
		})

		// Обработка ошибок парсинга/валидации
		if err != nil {
			errorMsg := "Невалидный токен"
			if errors.Is(err, jwt.ErrTokenExpired) {
				errorMsg = "Срок действия токена истек"
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
                 errorMsg = "Некорректный формат токена"
            }
			c.JSON(http.StatusUnauthorized, gin.H{"error": errorMsg})
			c.Abort()
			return
		}

		// Проверяем, валиден ли токен и извлекаем данные (claims)
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Проверка срока действия (дополнительно, хотя Parse уже проверяет)
			if expFloat, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(expFloat) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Срок действия токена истек"})
					c.Abort()
					return
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Некорректный формат срока действия токена"})
				c.Abort()
				return
			}

			// Извлечение данных пользователя из claims
			userIDFloat, okUserID := claims["user_id"].(float64)
			isAdmin, okIsAdmin := claims["is_admin"].(bool)
			isManager, okIsManager := claims["is_manager"].(bool)

			// Проверяем, что все необходимые поля присутствуют и имеют правильный тип
			if !okUserID || !okIsAdmin || !okIsManager {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения данных из токена"})
				c.Abort()
				return
			}

			// Сохраняем данные пользователя в контексте Gin
			c.Set("userID", int(userIDFloat)) // Преобразуем float64 в int
			c.Set("isAdmin", isAdmin)
			c.Set("isManager", isManager)

			c.Next() // Передаем управление следующему обработчику
		} else {
			// Если claims не являются jwt.MapClaims или токен не валиден по другим причинам
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Невалидный токен"})
			c.Abort()
		}
	}
}

// AdminOnly - middleware для проверки прав администратора
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права администратора."})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ManagerOrAdminOnly - middleware для проверки прав менеджера или администратора
func ManagerOrAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, adminExists := c.Get("isAdmin")
		isManager, managerExists := c.Get("isManager")

		hasAccess := (adminExists && isAdmin.(bool)) || (managerExists && isManager.(bool))

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права менеджера или администратора."})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ManagerOnly - middleware для проверки прав руководителя (можно добавить)
// func ManagerOnly() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		isManager, exists := c.Get("isManager")
// 		if !exists || !isManager.(bool) {
// 			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права руководителя."})
// 			c.Abort()
// 			return
// 		}
// 		c.Next()
// 	}
// }
