package services

import (
	"errors"
	"fmt" // Добавлен импорт fmt
	"log" // Добавляем импорт log
	"time"

	"vacation-scheduler/internal/models"
	"vacation-scheduler/internal/repositories" // Добавлен импорт repositories
)

// VacationServiceInterface определяет методы для сервиса отпусков
type VacationServiceInterface interface {
	GetVacationLimit(userID int, year int) (*models.VacationLimit, error)
	SetVacationLimit(userID int, year int, totalDays int) error
	ValidateVacationRequest(request *models.VacationRequest) error
	SaveVacationRequest(request *models.VacationRequest) error
	SubmitVacationRequest(requestID int, userID int) error
	CheckIntersections(departmentID int, year int) ([]models.Intersection, error)
	NotifyManager(managerID int, intersections []models.Intersection) error
	GetUserVacations(userID int, year int, statusFilter *int) ([]models.VacationRequest, error)                                                                          // Добавлен statusFilter
	GetDepartmentVacations(departmentID int, year int, statusFilter *int) ([]models.VacationRequest, error)                                                              // Добавлен statusFilter
	GetAllUserVacations(requestingUserID int, yearFilter *int, statusFilter *int, userIDFilter *int, departmentIDFilter *int) ([]models.VacationRequestAdminView, error) // Добавлен метод для админов/менеджеров
	CancelVacationRequest(requestID int, cancellingUserID int) error                                                                                                     // Изменен параметр
	ApproveVacationRequest(requestID int, approverID int) error                                                                                                          // Новый метод
	RejectVacationRequest(requestID int, rejecterID int, reason string) error                                                                                            // Новый метод
}

// VacationRepositoryInterface определяет методы для работы с данными отпусков.
type VacationRepositoryInterface interface {
	// --- Лимиты ---
	GetVacationLimit(userID int, year int) (*models.VacationLimit, error)
	CreateOrUpdateVacationLimit(userID int, year int, totalDays int) error
	UpdateVacationLimitUsedDays(userID int, year int, daysDelta int) error // Добавлен метод для изменения использованных дней

	// --- Заявки ---
	GetVacationRequestByID(requestID int) (*models.VacationRequest, error) // Добавлен метод получения заявки по ID
	SaveVacationRequest(request *models.VacationRequest) error
	UpdateVacationRequest(request *models.VacationRequest) error                                                                                      // Для обновления комментария и т.д. пользователем
	UpdateRequestStatusByID(requestID int, newStatusID int) error                                                                                     // Изменен: обновляет статус по ID заявки (без userID)
	GetVacationRequestsByUser(userID int, year int, statusFilter *int) ([]models.VacationRequest, error)                                              // Добавлен statusFilter (указатель на int)
	GetVacationRequestsByDepartment(departmentID int, year int, statusFilter *int) ([]models.VacationRequest, error)                                  // Добавлен statusFilter
	GetAllVacationRequests(yearFilter *int, statusFilter *int, userIDFilter *int, departmentIDFilter *int) ([]models.VacationRequestAdminView, error) // Добавлен метод для админов/менеджеров с фильтрами

	// --- Уведомления ---
	CreateNotification(notification *models.Notification) error

	// --- Периоды (при необходимости) ---
	// GetPeriodsByRequestID(requestID int) ([]models.VacationPeriod, error) // Пример
	// DeletePeriodsByRequestID(requestID int) error // Пример
}

// VacationService реализует VacationServiceInterface
type VacationService struct {
	vacationRepo VacationRepositoryInterface          // Используем интерфейс репозитория
	userRepo     repositories.UserRepositoryInterface // Используем интерфейс репозитория пользователей
}

// Обновляем конструктор, чтобы принимать интерфейсы
func NewVacationService(vacationRepo VacationRepositoryInterface, userRepo repositories.UserRepositoryInterface) *VacationService {
	return &VacationService{
		vacationRepo: vacationRepo,
		userRepo:     userRepo,
	}
}

// GetVacationLimit получает лимит отпуска для пользователя
func (s *VacationService) GetVacationLimit(userID int, year int) (*models.VacationLimit, error) {
	// Вызываем метод репозитория отпусков
	limit, err := s.vacationRepo.GetVacationLimit(userID, year)
	if err != nil {
		// Если лимит не найден, можно создать дефолтный (опционально)
		// Если лимит не найден, возвращаем дефолтное значение лимита (например, 28 дней с 0 использованных)
		if err.Error() == "лимит отпуска не найден для данного пользователя и года" {
			defaultLimit := &models.VacationLimit{UserID: userID, Year: year, TotalDays: 28, UsedDays: 0}
			return defaultLimit, nil
		}
		return nil, err // Возвращаем другие ошибки БД
	}
	return limit, nil
}

// SetVacationLimit устанавливает (создает или обновляет) лимит отпуска для пользователя
func (s *VacationService) SetVacationLimit(userID int, year int, totalDays int) error {
	// TODO: Добавить валидацию входных данных (например, totalDays > 0)
	if totalDays < 0 {
		return errors.New("количество дней отпуска не может быть отрицательным")
	}
	return s.vacationRepo.CreateOrUpdateVacationLimit(userID, year, totalDays)
}

// ValidateVacationRequest проверяет условия отпуска
func (s *VacationService) ValidateVacationRequest(request *models.VacationRequest) error {
	// Проверка на наличие части отпуска не менее 14 дней
	hasLongPeriod := false
	totalDays := 0

	if len(request.Periods) == 0 {
		return errors.New("необходимо указать хотя бы один период отпуска")
	}

	// Сортируем периоды по дате начала для упрощения проверки пересечений
	// sort.Slice(request.Periods, func(i, j int) bool {
	//  return request.Periods[i].StartDate.Time.Before(request.Periods[j].StartDate.Time)
	// })
	// Примечание: Сортировка не обязательна для логики ниже, но может быть полезна. Пока оставим без сортировки.

	for i, period := range request.Periods {
		// 1. Проверка корректности дат внутри одного периода
		if period.StartDate.IsZero() || period.EndDate.IsZero() || period.EndDate.Time.Before(period.StartDate.Time) {
			return fmt.Errorf("некорректные даты в периоде %d: дата начала %s, дата окончания %s",
				i+1, period.StartDate.Format("2006-01-02"), period.EndDate.Format("2006-01-02"))
		}

		// 2. Проверка пересечений с *другими* периодами в *этой же* заявке
		for j := i + 1; j < len(request.Periods); j++ {
			otherPeriod := request.Periods[j]
			// Используем существующую вспомогательную функцию doPeriodIntersect
			if doPeriodIntersect(period, otherPeriod) {
				return fmt.Errorf("периоды %d (%s - %s) и %d (%s - %s) в заявке пересекаются",
					i+1, period.StartDate.Format("2006-01-02"), period.EndDate.Format("2006-01-02"),
					j+1, otherPeriod.StartDate.Format("2006-01-02"), otherPeriod.EndDate.Format("2006-01-02"))
			}
		}
		// Доверяем DaysCount из запроса
		totalDays += period.DaysCount
		if period.DaysCount >= 14 {
			hasLongPeriod = true
		}
	}

	if !hasLongPeriod {
		return errors.New("Одна из частей отпуска должна быть не менее 14 календарных дней")
	}

	// Проверка лимита дней
	limit, err := s.GetVacationLimit(request.UserID, request.Year)
	availableDays := 0 // Инициализируем доступные дни

	if err != nil {
		// Проверяем, является ли ошибка "лимит не найден"
		if err.Error() == "лимит отпуска не найден для данного пользователя и года" {
			// Лимит не найден, используем стандартный лимит (например, 28) и считаем, что ничего не использовано
			// TODO: Возможно, стоит вынести стандартный лимит в конфигурацию
			const defaultLimit = 28
			availableDays = defaultLimit
			// Не возвращаем ошибку, просто используем дефолтное значение
			// Возвращаем другие ошибки (например, ошибка БД)
			log.Printf("[Validation Error] UserID: %d, Year: %d - Error getting limit: %v", request.UserID, request.Year, err) // LOGGING
			return fmt.Errorf("ошибка получения лимита отпуска: %w", err)
		}
	} else {
		// Лимит найден, используем его
		availableDays = limit.TotalDays - limit.UsedDays
	}

	// LOGGING: Log values just before the final check
	log.Printf("[Validation Check] UserID: %d, Year: %d, TotalDaysLimit: %d, UsedDaysLimit: %d, CalculatedAvailable: %d, DaysRequestedInThisRequest: %d",
		request.UserID, request.Year, limit.TotalDays, limit.UsedDays, availableDays, totalDays)

	// НОВАЯ ПРОВЕРКА: Запрошенные дни должны ТОЧНО соответствовать доступным дням
	if totalDays != availableDays {
		// LOGGING: Log the failure reason
		log.Printf("[Validation Failed] UserID: %d, Year: %d - Days mismatch: available %d, requested %d",
			request.UserID, request.Year, availableDays, totalDays)
		return fmt.Errorf("необходимо использовать все доступные дни отпуска: доступно %d, запрошено %d", availableDays, totalDays)
	}

	// Старая проверка (на всякий случай, хотя новая её покрывает)
	// if totalDays > availableDays {
	//  log.Printf("[Validation Failed] UserID: %d, Year: %d - Limit exceeded: available %d, requested %d",
	//   request.UserID, request.Year, availableDays, totalDays)
	//  return fmt.Errorf("превышен доступный лимит дней отпуска: доступно %d, запрошено %d", availableDays, totalDays)
	// }

	log.Printf("[Validation OK] UserID: %d, Year: %d - Exact days match passed: available %d, requested %d",
		request.UserID, request.Year, availableDays, totalDays) // LOGGING
	return nil // Все проверки пройдены
}

// SaveVacationRequest сохраняет заявку на отпуск
func (s *VacationService) SaveVacationRequest(request *models.VacationRequest) error {
	// Устанавливаем статус черновика, если не указан
	if request.StatusID == 0 {
		request.StatusID = models.StatusDraft
	}

	// Рассчитываем общее количество дней перед сохранением
	// Репозиторий также делает это на всякий случай, но лучше здесь.
	request.DaysRequested = 0 // Сбрасываем на случай повторного сохранения
	for _, p := range request.Periods {
		request.DaysRequested += p.DaysCount
	}

	// Используем vacationRepo
	return s.vacationRepo.SaveVacationRequest(request)
}

// SubmitVacationRequest отправляет заявку руководителю
func (s *VacationService) SubmitVacationRequest(requestID int, userID int) error {
	// 1. Получаем заявку
	req, err := s.vacationRepo.GetVacationRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("ошибка получения заявки для отправки: %w", err)
	}
	if req == nil {
		return errors.New("заявка не найдена")
	}
	// 2. Проверяем владельца
	if req.UserID != userID {
		return errors.New("нет прав на отправку этой заявки")
	}
	// 3. Проверяем текущий статус (можно отправить только черновик)
	if req.StatusID != models.StatusDraft {
		return errors.New("можно отправить только заявку в статусе 'Черновик'")
	}

	// 4. **Валидируем заявку перед отправкой**
	err = s.ValidateVacationRequest(req) // Вызываем существующую функцию валидации
	if err != nil {
		// Если валидация не прошла, возвращаем ошибку
		return fmt.Errorf("ошибка валидации заявки перед отправкой: %w", err)
	}

	// 5. **Проверяем лимит и списываем дни ПРИ ОТПРАВКЕ (ПЕРЕД сменой статуса)**
	// Используем DaysRequested, которое должно быть рассчитано при сохранении/получении
	if req.DaysRequested <= 0 {
		// Если дней 0 или меньше, лимит проверять и списывать не нужно
		fmt.Printf("Предупреждение: Заявка ID %d отправляется с %d запрошенными днями.\n", requestID, req.DaysRequested)
	} else {
		limit, errLimit := s.vacationRepo.GetVacationLimit(req.UserID, req.Year)
		limitNotFoundErrorMsg := "лимит отпуска не найден для данного пользователя и года"

		if errLimit != nil && errLimit.Error() != limitNotFoundErrorMsg {
			// Если ошибка не "не найден", это проблема с БД
			return fmt.Errorf("ошибка получения лимита отпуска пользователя %d (год %d) перед отправкой заявки %d: %w", req.UserID, req.Year, requestID, errLimit)
		}

		availableDays := 0
		if errLimit != nil && errLimit.Error() == limitNotFoundErrorMsg {
			// Лимит не найден, используем дефолтный total для проверки
			const defaultTotalDays = 28 // TODO: Вынести в конфиг
			availableDays = defaultTotalDays
		} else if limit != nil {
			// Лимит найден
			availableDays = limit.TotalDays - limit.UsedDays
		}

		if req.DaysRequested > availableDays {
			return fmt.Errorf("недостаточно дней отпуска у пользователя %d (год %d) для отправки заявки %d: доступно %d, запрошено %d", req.UserID, req.Year, availableDays, req.DaysRequested, requestID)
		}

		// Пытаемся списать дни (увеличить used_days)
		errSpend := s.vacationRepo.UpdateVacationLimitUsedDays(req.UserID, req.Year, req.DaysRequested)
		if errSpend != nil {
			// Если списать не удалось, НЕ отправляем заявку
			return fmt.Errorf("ошибка списания %d дней из лимита пользователя %d (год %d) при отправке заявки %d: %w", req.DaysRequested, req.UserID, req.Year, requestID, errSpend)
		}
		fmt.Printf("Успешно списаны дни (%d) для заявки %d пользователя %d (год %d) при отправке.\n", req.DaysRequested, requestID, req.UserID, req.Year)

	} // Конец блока if req.DaysRequested > 0

	// 6. Обновляем статус на "На рассмотрении" (ТОЛЬКО если все предыдущие шаги успешны)
	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusPending)
	if err != nil {
		// Если не удалось обновить статус, НО дни уже были списаны, нужно их вернуть!
		if req.DaysRequested > 0 { // Проверяем, были ли дни для списания/возврата
			fmt.Printf("КРИТИЧЕСКАЯ ОШИБКА: Дни (%d) для заявки %d пользователя %d (год %d) были списаны, но НЕ удалось установить статус 'На рассмотрении'. Попытка вернуть дни...\n", req.DaysRequested, requestID, req.UserID, req.Year)
			revertErr := s.vacationRepo.UpdateVacationLimitUsedDays(req.UserID, req.Year, -req.DaysRequested) // Возвращаем дни
			if revertErr != nil {
				fmt.Printf("КРИТИЧЕСКАЯ ОШИБКА: НЕ удалось вернуть списанные дни (%d) для заявки %d пользователя %d (год %d) после неудачной отправки: %v\n", req.DaysRequested, requestID, req.UserID, req.Year, revertErr)
				// Алертинг!
			} else {
				fmt.Printf("Успешно возвращены дни (%d) для заявки %d пользователя %d (год %d) после неудачной отправки.\n", req.DaysRequested, requestID, req.UserID, req.Year)
			}
		}
		return fmt.Errorf("ошибка установки статуса 'На рассмотрении' для заявки %d (дни могли быть списаны): %w", requestID, err)
	}

	// TODO: Опционально: Уведомить менеджера о новой заявке
	return nil
}

// CheckIntersections проверяет пересечения отпусков в подразделении
func (s *VacationService) CheckIntersections(departmentID int, year int) ([]models.Intersection, error) {
	// Получаем все заявки на отпуск в подразделении за указанный год (без фильтра статуса)
	// Используем vacationRepo
	requests, err := s.vacationRepo.GetVacationRequestsByDepartment(departmentID, year, nil) // Добавляем nil для statusFilter
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заявок подразделения: %w", err)
	}

	// Получаем всех сотрудников подразделения
	// Используем userRepo
	users, err := s.userRepo.GetUsersByDepartment(departmentID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователей подразделения: %w", err)
	}

	// Создаем мапу для быстрого поиска пользователей по ID
	userMap := make(map[int]models.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// Ищем пересечения
	var intersections []models.Intersection

	for i, req1 := range requests {
		for j := i + 1; j < len(requests); j++ {
			req2 := requests[j]

			// Пропускаем, если заявки от одного пользователя
			if req1.UserID == req2.UserID {
				continue
			}

			// Проверяем пересечения периодов
			for _, period1 := range req1.Periods {
				for _, period2 := range req2.Periods {
					if doPeriodIntersect(period1, period2) {
						// Находим период пересечения, используя .Time для max/min
						start := max(period1.StartDate.Time, period2.StartDate.Time)
						end := min(period1.EndDate.Time, period2.EndDate.Time)
						daysCount := int(end.Sub(start).Hours()/24) + 1

						intersection := models.Intersection{
							UserID1:   req1.UserID,
							UserName1: userMap[req1.UserID].FullName,
							UserID2:   req2.UserID,
							UserName2: userMap[req2.UserID].FullName,
							// Оборачиваем time.Time в models.CustomDate при присваивании
							StartDate: models.CustomDate{Time: start},
							EndDate:   models.CustomDate{Time: end},
							DaysCount: daysCount,
						}

						intersections = append(intersections, intersection)
					}
				}
			}
		}
	}

	return intersections, nil
}

// NotifyManager уведомляет руководителя о пересечениях отпусков
func (s *VacationService) NotifyManager(managerID int, intersections []models.Intersection) error {
	if len(intersections) == 0 {
		return nil
	}

	notification := &models.Notification{
		UserID:    managerID,
		Title:     "Обнаружено пересечение отпусков",
		Message:   "В подразделении обнаружены пересечения отпусков сотрудников. Требуется ваше внимание.",
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	// Используем vacationRepo
	return s.vacationRepo.CreateNotification(notification)
}

// GetUserVacations получает заявки конкретного пользователя с возможностью фильтрации по статусу
func (s *VacationService) GetUserVacations(userID int, year int, statusFilter *int) ([]models.VacationRequest, error) {
	// Вызываем соответствующий метод репозитория
	return s.vacationRepo.GetVacationRequestsByUser(userID, year, statusFilter)
}

// GetDepartmentVacations получает заявки подразделения с возможностью фильтрации по статусу
func (s *VacationService) GetDepartmentVacations(departmentID int, year int, statusFilter *int) ([]models.VacationRequest, error) {
	// Используем vacationRepo
	return s.vacationRepo.GetVacationRequestsByDepartment(departmentID, year, statusFilter)
}

// GetAllUserVacations получает все заявки (для админов) или заявки своего отдела (для менеджеров)
func (s *VacationService) GetAllUserVacations(requestingUserID int, yearFilter *int, statusFilter *int, userIDFilter *int, departmentIDFilter *int) ([]models.VacationRequestAdminView, error) {
	// 1. Получаем информацию о запрашивающем пользователе
	requestingUser, err := s.userRepo.FindByID(requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения данных запрашивающего пользователя: %w", err)
	}
	if requestingUser == nil {
		return nil, errors.New("запрашивающий пользователь не найден")
	}

	// 2. Определяем права доступа и фильтр по отделу
	actualDeptFilter := departmentIDFilter // Используем переданный фильтр по отделу, если он есть
	if !requestingUser.IsAdmin {
		if requestingUser.IsManager {
			// Менеджер может видеть только свой отдел. Игнорируем departmentIDFilter, если он не совпадает.
			if actualDeptFilter != nil && *actualDeptFilter != *requestingUser.DepartmentID {
				// Менеджер пытается посмотреть чужой отдел
				return nil, errors.New("менеджер может просматривать заявки только своего отдела")
			}
			// Устанавливаем фильтр по отделу менеджера, если он не был передан
			if actualDeptFilter == nil {
				if requestingUser.DepartmentID == nil {
					// Менеджер без отдела? Странная ситуация.
					return []models.VacationRequestAdminView{}, nil // Возвращаем пустой список
				}
				actualDeptFilter = requestingUser.DepartmentID
			}
		} else {
			// Обычный пользователь не должен вызывать этот метод напрямую через API
			// Но если вызвал, возвращаем ошибку прав доступа
			return nil, errors.New("недостаточно прав для просмотра всех заявок")
		}
	}
	// Если пользователь админ, actualDeptFilter остается таким, каким был передан (может быть nil)

	// 3. Вызываем метод репозитория с определенными фильтрами
	return s.vacationRepo.GetAllVacationRequests(yearFilter, statusFilter, userIDFilter, actualDeptFilter)
}

// CancelVacationRequest отменяет заявку пользователя (или админ/менеджер)
func (s *VacationService) CancelVacationRequest(requestID int, cancellingUserID int) error {
	// 1. Получаем заявку
	req, err := s.vacationRepo.GetVacationRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("ошибка получения заявки для отмены: %w", err)
	}
	if req == nil {
		return errors.New("заявка не найдена")
	}

	// 2. Проверяем права доступа на отмену
	canCancel := false
	// Владелец может отменить свою заявку в определенных статусах
	if req.UserID == cancellingUserID && (req.StatusID == models.StatusDraft || req.StatusID == models.StatusPending) {
		canCancel = true
	} else {
		// Проверяем, является ли отменяющий админом или менеджером отдела сотрудника
		cancellingUser, err := s.userRepo.FindByID(cancellingUserID)
		if err != nil || cancellingUser == nil {
			return errors.New("не удалось проверить права пользователя на отмену")
		}
		if cancellingUser.IsAdmin {
			canCancel = true // Админ может отменить любую заявку
		} else if cancellingUser.IsManager {
			// Менеджер может отменить заявку сотрудника своего отдела
			employee, err := s.userRepo.FindByID(req.UserID)
			if err == nil && employee != nil && employee.DepartmentID != nil && cancellingUser.DepartmentID != nil && *employee.DepartmentID == *cancellingUser.DepartmentID {
				canCancel = true
			}
		}
	}

	if !canCancel {
		return errors.New("нет прав на отмену этой заявки")
	}

	// 3. Проверяем текущий статус заявки (можно ли ее отменить)
	// Отменять можно Черновик, На рассмотрении, Утверждена. Нельзя отменять уже Отклоненную или Отмененную.
	if req.StatusID == models.StatusRejected || req.StatusID == models.StatusCancelled {
		return fmt.Errorf("нельзя отменить заявку ID %d в статусе '%d'", requestID, req.StatusID)
	}

	originalStatus := req.StatusID // Запоминаем исходный статус для возврата дней

	// 4. Обновляем статус на "Отменена"
	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusCancelled)
	if err != nil {
		return fmt.Errorf("ошибка установки статуса 'Отменена' для заявки %d: %w", requestID, err)
	}

	// 5. Возвращаем дни в лимит, если отменяется заявка, которая была НА РАССМОТРЕНИИ или УТВЕРЖДЕНА
	// (т.е. дни были списаны при отправке)
	if originalStatus == models.StatusPending || originalStatus == models.StatusApproved {
		daysToReturn := req.DaysRequested // Используем сохраненное значение
		if daysToReturn > 0 {
			errReturn := s.vacationRepo.UpdateVacationLimitUsedDays(req.UserID, req.Year, -daysToReturn) // Возвращаем дни
			if errReturn != nil {
				// Логируем ошибку, но сама отмена уже произошла.
				fmt.Printf("ВНИМАНИЕ: Заявка %d отменена, но не удалось вернуть %d дней в лимит пользователя %d (год %d): %v\n", requestID, daysToReturn, req.UserID, req.Year, errReturn)
				// TODO: Механизм компенсации или уведомления.
			} else {
				fmt.Printf("Успешно возвращены дни (%d) для заявки %d пользователя %d (год %d) при отмене.\n", daysToReturn, requestID, req.UserID, req.Year)
			}
		}
	} // Конец if originalStatus == Pending || Approved

	// TODO: Отправить уведомление пользователю об отмене заявки (особенно если отменил не он сам)

	return nil
}

// ApproveVacationRequest утверждает заявку (менеджер/админ)
func (s *VacationService) ApproveVacationRequest(requestID int, approverID int) error {
	// 1. Получаем заявку
	req, err := s.vacationRepo.GetVacationRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("ошибка получения заявки ID %d для утверждения: %w", requestID, err)
	}
	if req == nil {
		return fmt.Errorf("заявка ID %d не найдена", requestID)
	}

	// 2. Проверяем права доступа утверждающего
	approver, err := s.userRepo.FindByID(approverID)
	if err != nil {
		return fmt.Errorf("ошибка проверки прав пользователя ID %d: %w", approverID, err)
	}
	if approver == nil {
		return fmt.Errorf("утверждающий пользователь ID %d не найден", approverID)
	}

	canApprove := false
	if approver.IsAdmin {
		canApprove = true // Админ может все
	} else if approver.IsManager {
		// Менеджер может утвердить заявку сотрудника своего отдела
		employee, err := s.userRepo.FindByID(req.UserID)
		// Проверяем наличие ID отдела у обоих и их совпадение
		if err == nil && employee != nil && employee.DepartmentID != nil && approver.DepartmentID != nil && *employee.DepartmentID == *approver.DepartmentID {
			canApprove = true
		} else if err != nil {
			// Логируем ошибку получения данных сотрудника, но не прерываем, т.к. админ все равно может утвердить
			fmt.Printf("Предупреждение: не удалось получить данные сотрудника %d для проверки отдела при утверждении заявки %d: %v\n", req.UserID, requestID, err)
		}
	}

	if !canApprove {
		return fmt.Errorf("пользователь ID %d не имеет прав для утверждения заявки ID %d", approverID, requestID)
	}

	// 3. Проверяем текущий статус (утвердить можно только "На рассмотрении")
	if req.StatusID != models.StatusPending {
		return fmt.Errorf("можно утвердить только заявку ID %d в статусе 'На рассмотрении' (текущий статус: %d)", requestID, req.StatusID)
	}

	// 4. Обновляем статус заявки на "Утверждена"
	// Дни уже были списаны при отправке (SubmitVacationRequest), поэтому здесь только меняем статус.
	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusApproved)
	if err != nil {
		// Если не удалось обновить статус, дни остаются списанными.
		// Это не идеальная ситуация, но и откат дней здесь был бы неверным,
		// так как заявка фактически могла быть утверждена (например, устно),
		// а проблема только в записи статуса. Логируем ошибку.
		fmt.Printf("ВНИМАНИЕ: Не удалось установить статус 'Утверждена' для заявки %d (пользователь %d, год %d), хотя дни были списаны при отправке: %v\n", requestID, req.UserID, req.Year, err)
		return fmt.Errorf("ошибка установки статуса 'Утверждена' для заявки %d: %w", requestID, err)
	}

	// TODO: Отправить уведомление пользователю об утверждении заявки

	return nil // Все успешно
}

// RejectVacationRequest отклоняет заявку (менеджер/админ)
// При отклонении заявки из статуса "На рассмотрении" (Pending), ранее списанные дни возвращаются пользователю.
func (s *VacationService) RejectVacationRequest(requestID int, rejecterID int, reason string) error {
	// 1. Получаем заявку
	req, err := s.vacationRepo.GetVacationRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("ошибка получения заявки ID %d для отклонения: %w", requestID, err)
	}
	if req == nil {
		return fmt.Errorf("заявка ID %d не найдена", requestID)
	}

	// 2. Проверяем права доступа отклоняющего
	rejecter, err := s.userRepo.FindByID(rejecterID)
	if err != nil {
		return fmt.Errorf("ошибка проверки прав пользователя ID %d: %w", rejecterID, err)
	}
	if rejecter == nil {
		return fmt.Errorf("отклоняющий пользователь ID %d не найден", rejecterID)
	}

	canReject := false
	if rejecter.IsAdmin {
		canReject = true
	} else if rejecter.IsManager {
		employee, err := s.userRepo.FindByID(req.UserID)
		if err == nil && employee != nil && employee.DepartmentID != nil && rejecter.DepartmentID != nil && *employee.DepartmentID == *rejecter.DepartmentID {
			canReject = true
		} else if err != nil {
			fmt.Printf("Предупреждение: не удалось получить данные сотрудника %d для проверки отдела при отклонении заявки %d: %v\n", req.UserID, requestID, err)
		}
	}

	if !canReject {
		return fmt.Errorf("пользователь ID %d не имеет прав для отклонения заявки ID %d", rejecterID, requestID)
	}

	// 3. Проверяем текущий статус (отклонить можно только "На рассмотрении")
	if req.StatusID != models.StatusPending {
		return fmt.Errorf("можно отклонить только заявку ID %d в статусе 'На рассмотрении' (текущий статус: %d)", requestID, req.StatusID)
	}

	// 4. Обновляем статус заявки на "Отклонена" ПЕРЕД возвратом дней
	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusRejected)
	if err != nil {
		return fmt.Errorf("ошибка установки статуса 'Отклонена' для заявки %d: %w", requestID, err)
	}

	// 5. Возвращаем списанные при отправке дни
	daysToReturn := req.DaysRequested
	if daysToReturn > 0 {
		fmt.Printf("RejectVacationRequest: Попытка возврата %d дней для заявки %d (пользователь %d, год %d)\n", daysToReturn, requestID, req.UserID, req.Year) // Добавлено логирование
		errReturn := s.vacationRepo.UpdateVacationLimitUsedDays(req.UserID, req.Year, -daysToReturn)                                                           // Возвращаем дни
		if errReturn != nil {
			// Статус уже "Отклонена", но дни вернуть не удалось. Логируем подробнее.
			fmt.Printf("КРИТИЧЕСКАЯ ОШИБКА: Заявка %d отклонена, но НЕ удалось вернуть %d дней в лимит пользователя %d (год %d): %v\n", requestID, daysToReturn, req.UserID, req.Year, errReturn)
			// ВАЖНО: Возвращаем ошибку пользователю, чтобы он знал о проблеме с возвратом дней!
			return fmt.Errorf("заявка отклонена, но произошла ошибка при возврате дней в лимит: %w", errReturn)
		} else {
			fmt.Printf("Успешно возвращены дни (%d) для заявки %d пользователя %d (год %d) при отклонении.\n", daysToReturn, requestID, req.UserID, req.Year)
		}
	}

	// 6. Добавляем комментарий с причиной отклонения (если нужно)
	// TODO: Реализовать сохранение причины отклонения, если она передана.

	// TODO: Отправить уведомление пользователю об отклонении заявки

	return nil
}

// Вспомогательные функции

// doPeriodIntersect проверяет, пересекаются ли два периода отпуска, используя .Time
func doPeriodIntersect(period1, period2 models.VacationPeriod) bool {
	// Используем .Time для доступа к значениям time.Time для сравнения
	return period1.StartDate.Time.Before(period2.EndDate.Time) && period2.StartDate.Time.Before(period1.EndDate.Time)
}

// max возвращает более позднюю из двух дат
func max(date1, date2 time.Time) time.Time {
	if date1.After(date2) {
		return date1
	}
	return date2
}

// min возвращает более раннюю из двух дат
func min(date1, date2 time.Time) time.Time {
	if date1.Before(date2) {
		return date1
	}
	return date2
}
