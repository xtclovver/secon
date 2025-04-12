package services

import (
	"errors"
	"fmt" // Добавлен импорт fmt

	// "log" // Убираем неиспользуемый импорт log
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
		} else {
			// Возвращаем другие ошибки (например, ошибка БД)
			return fmt.Errorf("ошибка получения лимита отпуска: %w", err)
		}
	} else {
		// Лимит найден, используем его
		availableDays = limit.TotalDays - limit.UsedDays
	}

	// Изменено: Проверка, что запрошено ТОЧНО столько дней, сколько доступно
	// (Исходя из интерпретации "все назначенные дни были израсходованы")
	// ВНИМАНИЕ: Это может быть не стандартным поведением. Обычно проверяют totalDays <= availableDays.
	// Если нужно стандартное поведение (не превышать лимит), верните проверку if totalDays > availableDays
	if totalDays != availableDays {
		return fmt.Errorf("запрошенное количество дней (%d) не совпадает с доступным лимитом (%d)", totalDays, availableDays)
	}

	return nil // Все проверки пройдены
}

// SaveVacationRequest сохраняет заявку на отпуск
func (s *VacationService) SaveVacationRequest(request *models.VacationRequest) error {
	// Устанавливаем статус черновика, если не указан
	if request.StatusID == 0 {
		request.StatusID = 1 // Черновик
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

	// 5. Обновляем статус на "На рассмотрении"
	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusPending)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса заявки на 'На рассмотрении': %w", err)
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
		return fmt.Errorf("нельзя отменить заявку в статусе '%d'", req.StatusID)
	}

	originalStatus := req.StatusID // Запоминаем исходный статус для возврата дней

	// 4. Обновляем статус на "Отменена"
	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusCancelled)
	if err != nil {
		return fmt.Errorf("ошибка установки статуса 'Отменена': %w", err)
	}

	// 5. Возвращаем дни в лимит, если отменяется УТВЕРЖДЕННАЯ заявка
	if originalStatus == models.StatusApproved {
		totalDaysToReturn := 0
		for _, p := range req.Periods {
			totalDaysToReturn += p.DaysCount
		}
		if totalDaysToReturn > 0 {
			err = s.vacationRepo.UpdateVacationLimitUsedDays(req.UserID, req.Year, -totalDaysToReturn)
			if err != nil {
				// Логируем ошибку, но сама отмена уже произошла.
				fmt.Printf("ВНИМАНИЕ: Не удалось вернуть %d дней в лимит пользователя %d (год %d) при отмене заявки %d: %v\n", totalDaysToReturn, req.UserID, req.Year, requestID, err)
				// Возможно, стоит добавить механизм повторной попытки или уведомление администратору.
			}
		}
	}

	// TODO: Отправить уведомление пользователю об отмене заявки (особенно если отменил не он сам)

	return nil
}

// ApproveVacationRequest утверждает заявку (менеджер/админ)
func (s *VacationService) ApproveVacationRequest(requestID int, approverID int) error {
	// 1. Получаем заявку
	req, err := s.vacationRepo.GetVacationRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("ошибка получения заявки для утверждения: %w", err)
	}
	if req == nil {
		return errors.New("заявка не найдена")
	}

	// 2. Проверяем права доступа утверждающего
	approver, err := s.userRepo.FindByID(approverID)
	if err != nil || approver == nil {
		return errors.New("не удалось проверить права пользователя на утверждение")
	}

	canApprove := false
	if approver.IsAdmin {
		canApprove = true // Админ может все
	} else if approver.IsManager {
		// Менеджер может утвердить заявку сотрудника своего отдела
		employee, err := s.userRepo.FindByID(req.UserID)
		if err == nil && employee != nil && employee.DepartmentID != nil && approver.DepartmentID != nil && *employee.DepartmentID == *approver.DepartmentID {
			canApprove = true
		}
	}

	if !canApprove {
		return errors.New("недостаточно прав для утверждения этой заявки")
	}

	// 3. Проверяем текущий статус (утвердить можно только "На рассмотрении")
	if req.StatusID != models.StatusPending {
		return fmt.Errorf("можно утвердить только заявку в статусе 'На рассмотрении' (текущий статус: %d)", req.StatusID)
	}

	// 4. Проверяем доступный лимит дней у сотрудника ПЕРЕД утверждением
	totalDaysRequested := 0
	for _, p := range req.Periods {
		totalDaysRequested += p.DaysCount
	}

	limit, err := s.GetVacationLimit(req.UserID, req.Year) // Используем существующий метод сервиса
	if err != nil {
		// Если GetVacationLimit вернул ошибку (не "лимит не найден"), то проблема
		if err.Error() != "лимит отпуска не найден для данного пользователя и года" {
			return fmt.Errorf("ошибка получения лимита отпуска пользователя %d: %w", req.UserID, err)
		}
		// Если лимит не найден, GetVacationLimit возвращает дефолтный, проверка ниже сработает
	}

	availableDays := limit.TotalDays - limit.UsedDays
	if totalDaysRequested > availableDays {
		return fmt.Errorf("недостаточно дней отпуска у сотрудника: доступно %d, запрошено %d", availableDays, totalDaysRequested)
	}

	// --- Начало транзакции (гипотетически, если бы репозиторий поддерживал) ---
	// tx, err := s.vacationRepo.BeginTx() // Псевдокод
	// if err != nil { return err }
	// defer tx.Rollback() // Откат по умолчанию

	// 5. Обновляем статус заявки на "Утверждена"
	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusApproved) // Используем метод репо
	if err != nil {
		return fmt.Errorf("ошибка установки статуса 'Утверждена': %w", err)
	}

	// 6. Уменьшаем количество доступных дней в лимите
	if totalDaysRequested > 0 {
		err = s.vacationRepo.UpdateVacationLimitUsedDays(req.UserID, req.Year, totalDaysRequested) // Используем метод репо
		if err != nil {
			// Откатываем статус заявки, если не удалось обновить лимит? Или оставляем как есть и логируем?
			// Решение: Оставляем статус "Утверждена", но логируем критическую ошибку.
			fmt.Printf("КРИТИЧЕСКАЯ ОШИБКА: Заявка %d утверждена, но НЕ удалось списать %d дней из лимита пользователя %d (год %d): %v\n", requestID, totalDaysRequested, req.UserID, req.Year, err)
			// В реальном приложении здесь нужна система компенсации или уведомления.
			// Пока просто возвращаем ошибку, но статус уже изменен.
			return fmt.Errorf("заявка утверждена, но произошла ошибка при списании дней из лимита: %w", err)
			// Если бы была транзакция:
			// tx.Rollback()
			// return fmt.Errorf("ошибка списания дней из лимита: %w", err)
		}
	}

	// --- Коммит транзакции ---
	// err = tx.Commit() // Псевдокод
	// if err != nil { return err }

	// TODO: Отправить уведомление пользователю об утверждении заявки

	return nil
}

// RejectVacationRequest отклоняет заявку (менеджер/админ)
func (s *VacationService) RejectVacationRequest(requestID int, rejecterID int, reason string) error {
	// 1. Получаем заявку
	req, err := s.vacationRepo.GetVacationRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("ошибка получения заявки для отклонения: %w", err)
	}
	if req == nil {
		return errors.New("заявка не найдена")
	}

	// 2. Проверяем права доступа отклоняющего
	rejecter, err := s.userRepo.FindByID(rejecterID)
	if err != nil || rejecter == nil {
		return errors.New("не удалось проверить права пользователя на отклонение")
	}

	canReject := false
	if rejecter.IsAdmin {
		canReject = true
	} else if rejecter.IsManager {
		employee, err := s.userRepo.FindByID(req.UserID)
		if err == nil && employee != nil && employee.DepartmentID != nil && rejecter.DepartmentID != nil && *employee.DepartmentID == *rejecter.DepartmentID {
			canReject = true
		}
	}

	if !canReject {
		return errors.New("недостаточно прав для отклонения этой заявки")
	}

	// 3. Проверяем текущий статус (отклонить можно только "На рассмотрении")
	if req.StatusID != models.StatusPending {
		return fmt.Errorf("можно отклонить только заявку в статусе 'На рассмотрении' (текущий статус: %d)", req.StatusID)
	}

	// 4. Обновляем статус заявки на "Отклонена"
	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusRejected)
	if err != nil {
		return fmt.Errorf("ошибка установки статуса 'Отклонена': %w", err)
	}

	// 5. Добавляем комментарий с причиной отклонения (если нужно)
	// TODO: Решить, как хранить причину отклонения. Пока можно добавить в comment заявки.
	// Если reason не пустая, можно обновить комментарий заявки.
	// Но UpdateVacationRequest требует UserID владельца. Нужно либо изменить его,
	// либо добавить новый метод в репозиторий UpdateRequestCommentByID(requestID, comment).
	// Пока оставляем без обновления комментария.

	// 6. Возвращаем дни, если заявка была *ранее утверждена* (маловероятно в текущем потоке, но для полноты)
	// В текущей логике мы отклоняем только StatusPending, поэтому возврат дней не нужен.
	// Если бы логика позволяла отклонять StatusApproved, здесь был бы код возврата дней,
	// аналогичный CancelVacationRequest.

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
