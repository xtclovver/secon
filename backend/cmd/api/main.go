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

	// Создание сервисов
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret)
	// Передаем оба репозитория в NewVacationService
	vacationService := services.NewVacationService(vacationRepo, userRepo)
	// Создаем UserService
	userService := services.NewUserService(userRepo)

	// Создание обработчиков
	authHandler := handlers.NewAuthHandler(authService)
	// Создаем AppHandler вместо VacationHandler и передаем оба сервиса
	appHandler := handlers.NewAppHandler(vacationService, userService)

	// Настройка маршрутизатора Gin
	router := gin.Default()

	// Настройка CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Публичные маршруты
	router.POST("/api/auth/login", authHandler.Login)

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

			// Маршруты для менеджеров и администраторов
			vacationsMgmt := vacations.Group("")
			vacationsMgmt.Use(middleware.ManagerOrAdminOnly()) // Доступ только для менеджеров или админов
			{
				vacationsMgmt.GET("/all", appHandler.GetAllVacations)                          // Получение всех заявок (с фильтрами)
				vacationsMgmt.GET("/department/:id", appHandler.GetDepartmentVacations)        // Менеджер может получить заявки своего отдела (ID отдела игнорируется для менеджера)
				vacationsMgmt.GET("/intersections", appHandler.GetVacationIntersections)       // Проверка пересечений (доступна менеджерам)
				vacationsMgmt.POST("/requests/:id/approve", appHandler.ApproveVacationRequest) // Утверждение заявки
				vacationsMgmt.POST("/requests/:id/reject", appHandler.RejectVacationRequest)   // Отклонение заявки
			}
		}

		// Маршруты только для администраторов (используем appHandler)
		admin := api.Group("/admin")
		admin.Use(middleware.AdminOnly()) // Доступ только для админов
		{
			// Маршруты для управления лимитами отпусков
			admin.POST("/vacation-limits", appHandler.SetVacationLimit)
			// Маршрут для получения пользователей с лимитами
			admin.GET("/users", appHandler.GetAllUsersWithLimits)
			// TODO: Добавить другие админские маршруты (управление пользователями, отделами и т.д.)
		}
	}

	// Запуск сервера
	if err := router.Run(cfg.Server.Port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
