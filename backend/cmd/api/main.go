package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"vacation-scheduler/internal/config"
	"vacation-scheduler/internal/database"
	"vacation-scheduler/internal/handlers"
	"vacation-scheduler/internal/middleware"
	"vacation-scheduler/internal/repositories"
	"vacation-scheduler/internal/services"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация подключения к базе данных
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// Создание репозиториев
	userRepo := repositories.NewUserRepository(db)
	vacationRepo := repositories.NewVacationRepository(db)
	unitRepo := repositories.NewOrganizationalUnitRepository(db) // Добавлен репозиторий юнитов

	// Создание сервисов
	// Передаем оба репозитория в NewAuthService
	authService := services.NewAuthService(userRepo, vacationRepo, cfg.JWT.Secret)
	// Передаем все три репозитория в NewVacationService
	vacationService := services.NewVacationService(vacationRepo, userRepo, unitRepo) // Добавлен unitRepo
	// Создаем UserService
	userService := services.NewUserService(userRepo)
	unitService := services.NewOrganizationalUnitService(unitRepo, userRepo) // Добавлен сервис юнитов

	// Создание обработчиков
	authHandler := handlers.NewAuthHandler(authService)
	// Создаем AppHandler и передаем все три сервиса
	appHandler := handlers.NewAppHandler(vacationService, userService, unitService) // Добавлен unitService

	// Настройка маршрутизатора Gin
	router := gin.Default()

	// Настройка CORS
	// ВАЖНО: Для продакшена лучше использовать AllowOrigins с переменной окружения,
	// содержащей URL вашего фронтенда в Cloud Run, вместо AllowAllOrigins: true.
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true, // Разрешаем все источники для простоты
		// AllowOrigins:     []string{"http://localhost:3000"}, // Старая настройка
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Публичные маршруты
	router.POST("/api/auth/login", authHandler.Login)
	router.POST("/api/auth/register", authHandler.Register)              // Новый маршрут для регистрации
	router.GET("/api/units/tree", appHandler.GetOrganizationalUnitTree)  // ПУБЛИЧНЫЙ маршрут для дерева юнитов
	router.GET("/api/positions", appHandler.GetPositions)                // ПУБЛИЧНЫЙ маршрут для должностей
	router.GET("/api/units/children", appHandler.GetUnitChildrenHandler) // ПУБЛИЧНЫЙ маршрут для получения дочерних элементов

	// Защищенные маршруты
	api := router.Group("/api")
	api.Use(middleware.JWTAuth(cfg.JWT.Secret))
	{
		// Маршруты для работы с отпусками (используем appHandler)
		vacations := api.Group("/vacations")
		{
			vacations.GET("/limits/:year", appHandler.GetVacationLimit)
			vacations.POST("/requests", appHandler.CreateVacationRequest)
			vacations.POST("/requests/:id/submit", appHandler.SubmitVacationRequest)
			vacations.POST("/requests/:id/cancel", appHandler.CancelVacationRequest) // Доступен всем аутентифицированным (проверка прав внутри)
			vacations.GET("/my", appHandler.GetMyVacations)                          // Получение своих заявок
			// Новый маршрут для получения конфликтов (доступен всем аутентифицированным)
			vacations.GET("/conflicts", appHandler.GetVacationConflicts)

			// Маршруты для менеджеров и администраторов
			vacationsMgmt := vacations.Group("")
			vacationsMgmt.Use(middleware.ManagerOrAdminOnly()) // Доступ только для менеджеров или админов
			{
				vacationsMgmt.GET("/all", appHandler.GetAllVacations)                          // Получение всех заявок (с фильтрами)
				vacationsMgmt.GET("/unit/:id", appHandler.GetOrganizationalUnitVacations)      // Маршрут обновлен: /department/:id -> /unit/:id, обработчик изменен
				vacationsMgmt.GET("/intersections", appHandler.GetVacationIntersections)       // Проверка пересечений (доступна менеджерам)
				vacationsMgmt.POST("/requests/:id/approve", appHandler.ApproveVacationRequest) // Утверждение заявки
				vacationsMgmt.POST("/requests/:id/reject", appHandler.RejectVacationRequest)   // Отклонение заявки
			}
		}

		// Маршрут для дашборда руководителя
		dashboard := api.Group("/dashboard")
		dashboard.Use(middleware.ManagerOrAdminOnly()) // Доступ только для менеджеров или админов
		{
			dashboard.GET("/manager", appHandler.GetManagerDashboard) // Новый маршрут для дашборда
		}

		// Маршруты только для администраторов (используем appHandler)
		admin := api.Group("/admin")
		admin.Use(middleware.AdminOnly()) // Доступ только для админов
		{
			// Маршруты для управления лимитами отпусков
			admin.POST("/vacation-limits", appHandler.SetVacationLimit)
			// Переименован маршрут для избежания конфликта с GET /api/admin/users
			admin.GET("/users-with-limits", appHandler.GetAllUsersWithLimits) // GET /api/admin/users-with-limits?year=...

			// Маршруты для управления организационной структурой
			units := admin.Group("/units")
			{
				units.POST("", appHandler.CreateOrganizationalUnit) // Создать юнит
				// units.GET("/tree", appHandler.GetOrganizationalUnitTree)  // ПЕРЕМЕЩЕНО В ПУБЛИЧНУЮ СЕКЦИЮ
				units.GET("/:id", appHandler.GetOrganizationalUnitByID)   // Получить юнит по ID (остается админским)
				units.PUT("/:id", appHandler.UpdateOrganizationalUnit)    // Обновить юнит
				units.DELETE("/:id", appHandler.DeleteOrganizationalUnit) // Удалить юнит
				// Новый маршрут для получения пользователей юнита с лимитами (Используем :id вместо :unitId)
				units.GET("/:id/users-with-limits", appHandler.GetUnitUsersWithLimitsHandler) // GET /api/admin/units/{id}/users-with-limits?year=...
			}

			// Маршруты для управления пользователями (Admin only)
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", appHandler.GetAllUsersHandler)         // GET /api/admin/users - Получить всех пользователей
				adminUsers.PUT("/:id", appHandler.UpdateUserAdminHandler) // PUT /api/admin/users/{id} - Обновить пользователя админом
				// Маршрут обновления лимита перенесен сюда и использует :id
				adminUsers.PUT("/:id/vacation-limit", appHandler.UpdateUserVacationLimitHandler) // PUT /api/admin/users/{id}/vacation-limit
				// TODO: Добавить маршруты для создания/удаления пользователей админом, если нужно
			}
		}

		// Маршрут для обновления профиля пользователя (доступен всем аутентифицированным, права проверяются в обработчике)
		api.PUT("/users/:id", appHandler.UpdateUserProfile)
		// Маршрут для получения профиля текущего пользователя
		api.GET("/profile", appHandler.GetMyProfile) // Добавлен новый маршрут
	}

	// Запуск сервера
	if err := router.Run(cfg.Server.Port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
