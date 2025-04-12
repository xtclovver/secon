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
	CheckIntersections(unitID int, year int) ([]models.Intersection, error) // departmentID -> unitID
	NotifyManager(managerID int, intersections []models.Intersection) error
	GetUserVacations(userID int, year int, statusFilter *int) ([]models.VacationRequest, error)
	GetOrganizationalUnitVacations(unitID int, year int, statusFilter *int) ([]models.VacationRequest, error)                                                      // GetDepartmentVacations -> GetOrganizationalUnitVacations, departmentID -> unitID
	GetAllUserVacations(requestingUserID int, yearFilter *int, statusFilter *int, userIDFilter *int, unitIDFilter *int) ([]models.VacationRequestAdminView, error) // departmentIDFilter -> unitIDFilter
	CancelVacationRequest(requestID int, cancellingUserID int) error
	// Изменена сигнатура: возвращает список конфликтов и ошибку
	ApproveVacationRequest(requestID int, approverID int) ([]models.ConflictingPeriod, error)
	RejectVacationRequest(requestID int, rejecterID int, reason string) error
}

// VacationRepositoryInterface определяет методы для работы с данными отпусков.
// Перемещено из vacation_repository.go для корректной компиляции сервиса
type VacationRepositoryInterface interface {
	// --- Лимиты ---
	GetVacationLimit(userID int, year int) (*models.VacationLimit, error)
	CreateOrUpdateVacationLimit(userID int, year int, totalDays int) error
	UpdateVacationLimitUsedDays(userID int, year int, daysDelta int) error

	// --- Заявки ---
	GetVacationRequestByID(requestID int) (*models.VacationRequest, error)
	SaveVacationRequest(request *models.VacationRequest) error
	UpdateVacationRequest(request *models.VacationRequest) error
	UpdateRequestStatusByID(requestID int, newStatusID int) error
	GetVacationRequestsByUser(userID int, year int, statusFilter *int) ([]models.VacationRequest, error)
	GetVacationRequestsByOrganizationalUnit(unitID int, year int, statusFilter *int) ([]models.VacationRequest, error)
	GetAllVacationRequests(yearFilter *int, statusFilter *int, userIDFilter *int, unitIDsFilter []int) ([]models.VacationRequestAdminView, error) // Изменен тип unitIDsFilter на []int

	// --- Уведомления ---
	CreateNotification(notification *models.Notification) error

	// --- Проверка конфликтов ---
	GetUserPositionByID(userID int) (*int, error)                                                                                                         // Добавлен метод получения должности
	GetApprovedVacationConflictsByPosition(positionID int, excludeUserID int, periodsToCheck []models.VacationPeriod) ([]models.ConflictingPeriod, error) // Добавлен метод поиска конфликтов
}

// VacationService реализует VacationServiceInterface
type VacationService struct {
	vacationRepo VacationRepositoryInterface                        // Используем интерфейс репозитория отпусков
	userRepo     repositories.UserRepositoryInterface               // Используем интерфейс репозитория пользователей
	unitRepo     repositories.OrganizationalUnitRepositoryInterface // Добавляем интерфейс репозитория юнитов
}

// Обновляем конструктор, чтобы принимать интерфейсы
func NewVacationService(vacationRepo VacationRepositoryInterface, userRepo repositories.UserRepositoryInterface, unitRepo repositories.OrganizationalUnitRepositoryInterface) *VacationService { // Добавлен unitRepo
	return &VacationService{
		vacationRepo: vacationRepo,
		userRepo:     userRepo,
		unitRepo:     unitRepo, // Сохраняем unitRepo
	}
}

// GetVacationLimit получает лимит отпуска для пользователя.
// Если лимит не найден, пытается создать лимит по умолчанию (28 дней) и возвращает его.
func (s *VacationService) GetVacationLimit(userID int, year int) (*models.VacationLimit, error) {
	limit, err := s.vacationRepo.GetVacationLimit(userID, year)
	if err != nil {
		// Проверяем, является ли ошибка "не найдено" с помощью errors.Is и экспортированной ошибки
		if errors.Is(err, repositories.ErrLimitNotFound) {
			log.Printf("[GetVacationLimit] Limit not found for UserID: %d, Year: %d. Attempting to create default limit.", userID, year)
			// Пытаемся создать лимит по умолчанию
			defaultTotalDays := 28 // Лимит по умолчанию
			createErr := s.vacationRepo.CreateOrUpdateVacationLimit(userID, year, defaultTotalDays)
			if createErr != nil {
				log.Printf("[GetVacationLimit] Failed to create default limit for UserID: %d, Year: %d. Error: %v", userID, year, createErr)
				// Возвращаем исходную ошибку "не найдено", т.к. создать не удалось
				return nil, err
			}
			log.Printf("[GetVacationLimit] Default limit created successfully for UserID: %d, Year: %d.", userID, year)
			// Повторно пытаемся получить только что созданный лимит
			limit, err = s.vacationRepo.GetVacationLimit(userID, year)
			if err != nil {
				log.Printf("[GetVacationLimit] Failed to retrieve the newly created default limit for UserID: %d, Year: %d. Error: %v", userID, year, err)
				// Возвращаем ошибку получения после создания
				return nil, fmt.Errorf("ошибка получения созданного по умолчанию лимита: %w", err)
			}
			// Успешно получили созданный лимит
			return limit, nil
		}
		// Если ошибка не "не найдено", возвращаем ее
		log.Printf("[GetVacationLimit] Error retrieving limit for UserID: %d, Year: %d. Error: %v", userID, year, err)
		return nil, err
	}
	// Лимит найден с первого раза
	return limit, nil
}

// SetVacationLimit устанавливает (создает или обновляет) лимит отпуска для пользователя
func (s *VacationService) SetVacationLimit(userID int, year int, totalDays int) error {
	if totalDays < 0 {
		return errors.New("количество дней отпуска не может быть отрицательным")
	}
	return s.vacationRepo.CreateOrUpdateVacationLimit(userID, year, totalDays)
}

// ValidateVacationRequest проверяет условия отпуска
func (s *VacationService) ValidateVacationRequest(request *models.VacationRequest) error {
	hasLongPeriod := false
	totalDays := 0
	if len(request.Periods) == 0 {
		return errors.New("необходимо указать хотя бы один период отпуска")
	}

	for i, period := range request.Periods {
		if period.StartDate.IsZero() || period.EndDate.IsZero() || period.EndDate.Time.Before(period.StartDate.Time) {
			return fmt.Errorf("некорректные даты в периоде %d: дата начала %s, дата окончания %s",
				i+1, period.StartDate.Format("2006-01-02"), period.EndDate.Format("2006-01-02"))
		}
		for j := i + 1; j < len(request.Periods); j++ {
			if doPeriodIntersect(period, request.Periods[j]) {
				return fmt.Errorf("периоды %d и %d в заявке пересекаются", i+1, j+1)
			}
		}
		totalDays += period.DaysCount
		if period.DaysCount >= 14 {
			hasLongPeriod = true
		}
	}
	if !hasLongPeriod {
		return errors.New("Одна из частей отпуска должна быть не менее 14 календарных дней")
	}

	limit, err := s.GetVacationLimit(request.UserID, request.Year) // Эта функция теперь пытается создать лимит, если его нет
	if err != nil {
		// Если ошибка именно в том, что лимит не найден (и не удалось создать), даем понятное сообщение
		if errors.Is(err, repositories.ErrLimitNotFound) {
			log.Printf("[Validation Error] UserID: %d, Year: %d - Limit not found and could not be created: %v", request.UserID, request.Year, err)
			return fmt.Errorf("лимит отпуска для пользователя %d на %d год не найден и не может быть создан", request.UserID, request.Year)
		}
		// Для других ошибок получения/создания лимита
		log.Printf("[Validation Error] UserID: %d, Year: %d - Failed to get/create vacation limit: %v", request.UserID, request.Year, err)
		return fmt.Errorf("ошибка при получении/создании лимита отпуска: %w", err)
	}
	availableDays := limit.TotalDays - limit.UsedDays
	log.Printf("[Validation Check] UserID: %d, Year: %d, Limit: %d, Used: %d, Available: %d, Requested: %d", request.UserID, request.Year, limit.TotalDays, limit.UsedDays, availableDays, totalDays)
	if totalDays != availableDays {
		log.Printf("[Validation Failed] UserID: %d, Year: %d - Days mismatch: available %d, requested %d", request.UserID, request.Year, availableDays, totalDays)
		return fmt.Errorf("необходимо использовать все доступные дни отпуска: доступно %d, запрошено %d", availableDays, totalDays)
	}
	log.Printf("[Validation OK] UserID: %d, Year: %d - Exact days match passed: available %d, requested %d", request.UserID, request.Year, availableDays, totalDays)
	return nil
}

// checkUserUnitAccess проверяет иерархический доступ менеджера к сотруднику
func (s *VacationService) checkUserUnitAccess(accessor *models.User, targetUser *models.User) (bool, error) {
	if accessor.IsAdmin {
		return true, nil
	}
	if accessor.IsManager {
		if accessor.OrganizationalUnitID == nil || targetUser.OrganizationalUnitID == nil {
			log.Printf("[Access Check] Denied: Accessor (%d) or Target (%d) has no unit ID.", accessor.ID, targetUser.ID)
			return false, nil
		}
		subtreeIDs, err := s.unitRepo.GetSubtreeIDs(*accessor.OrganizationalUnitID)
		if err != nil {
			log.Printf("[Access Check] Error getting subtree for accessor %d (unit %d): %v", accessor.ID, *accessor.OrganizationalUnitID, err)
			return false, fmt.Errorf("ошибка получения поддерева юнитов руководителя: %w", err)
		}
		targetUnitID := *targetUser.OrganizationalUnitID
		for _, id := range subtreeIDs {
			if id == targetUnitID {
				return true, nil
			}
		}
		log.Printf("[Access Check] Denied: Target user %d (unit %d) is not in accessor %d's subtree (unit %d)", targetUser.ID, targetUnitID, accessor.ID, *accessor.OrganizationalUnitID)
	}
	return false, nil
}

// SaveVacationRequest сохраняет заявку на отпуск
func (s *VacationService) SaveVacationRequest(request *models.VacationRequest) error {
	if request.StatusID == 0 {
		request.StatusID = models.StatusDraft
	}
	request.DaysRequested = 0
	for _, p := range request.Periods {
		request.DaysRequested += p.DaysCount
	}
	log.Printf("[Service SaveVacationRequest] Calculated DaysRequested: %d for UserID: %d", request.DaysRequested, request.UserID)
	return s.vacationRepo.SaveVacationRequest(request)
}

// SubmitVacationRequest отправляет заявку руководителю
func (s *VacationService) SubmitVacationRequest(requestID int, userID int) error {
	req, err := s.vacationRepo.GetVacationRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("ошибка получения заявки для отправки: %w", err)
	}
	if req == nil {
		return errors.New("заявка не найдена")
	}
	if req.UserID != userID {
		return errors.New("нет прав на отправку этой заявки")
	}
	if req.StatusID != models.StatusDraft {
		return errors.New("можно отправить только заявку в статусе 'Черновик'")
	}
	if err = s.ValidateVacationRequest(req); err != nil {
		return fmt.Errorf("ошибка валидации заявки перед отправкой: %w", err)
	}

	if req.DaysRequested > 0 {
		limit, errLimit := s.vacationRepo.GetVacationLimit(req.UserID, req.Year)
		// Используем errors.Is для сравнения с экспортированной ошибкой
		// Переменная limitNotFoundErrorMsg больше не нужна
		if errLimit != nil && !errors.Is(errLimit, repositories.ErrLimitNotFound) {
			return fmt.Errorf("ошибка получения лимита пользователя %d (год %d) перед отправкой заявки %d: %w", req.UserID, req.Year, requestID, errLimit)
		}
		// Проверяем на любую ошибку (включая ErrLimitNotFound после попытки создания)
		if errLimit != nil {
			log.Printf("[Submit Error] UserID: %d, Year: %d, RequestID: %d - Limit not found: %v", req.UserID, req.Year, requestID, errLimit)
			return fmt.Errorf("невозможно отправить заявку: лимит отпуска для пользователя %d на %d год не установлен", req.UserID, req.Year)
		}
		availableDays := limit.TotalDays - limit.UsedDays
		if req.DaysRequested > availableDays {
			return fmt.Errorf("недостаточно дней отпуска у пользователя %d (год %d): доступно %d, запрошено %d", req.UserID, req.Year, availableDays, req.DaysRequested)
		}
		errSpend := s.vacationRepo.UpdateVacationLimitUsedDays(req.UserID, req.Year, req.DaysRequested)
		if errSpend != nil {
			log.Printf("[Service SubmitVacationRequest] Failed to spend days. UserID: %d, Year: %d, RequestID: %d, Days: %d, Error: %v", req.UserID, req.Year, requestID, req.DaysRequested, errSpend)
			return fmt.Errorf("ошибка списания %d дней из лимита пользователя %d (год %d) при отправке заявки %d: %w", req.DaysRequested, req.UserID, req.Year, requestID, errSpend)
		}
		log.Printf("[Service SubmitVacationRequest] Successfully spent days. UserID: %d, Year: %d, RequestID: %d, Days: %d", req.UserID, req.Year, requestID, req.DaysRequested)
	}

	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusPending)
	if err != nil {
		if req.DaysRequested > 0 {
			log.Printf("CRITICAL ERROR: Days (%d) for request %d (user %d, year %d) were spent, but failed to set status to Pending. Attempting to revert days...", req.DaysRequested, requestID, req.UserID, req.Year)
			revertErr := s.vacationRepo.UpdateVacationLimitUsedDays(req.UserID, req.Year, -req.DaysRequested)
			if revertErr != nil {
				log.Printf("CRITICAL ERROR: Failed to revert spent days (%d) for request %d (user %d, year %d) after failed submission: %v", req.DaysRequested, requestID, req.UserID, req.Year, revertErr)
			} else {
				log.Printf("Successfully reverted days (%d) for request %d (user %d, year %d) after failed submission.", req.DaysRequested, requestID, req.UserID, req.Year)
			}
		}
		return fmt.Errorf("ошибка установки статуса 'На рассмотрении' для заявки %d: %w", requestID, err)
	}
	// TODO: Notify manager
	return nil
}

// CheckIntersections проверяет пересечения отпусков (с учетом правила: только внутри отдела/сектора)
func (s *VacationService) CheckIntersections(unitID int, year int) ([]models.Intersection, error) {
	targetUnit, err := s.unitRepo.GetByID(unitID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации о юните %d: %w", unitID, err)
	}
	if targetUnit == nil {
		return nil, fmt.Errorf("юнит %d не найден", unitID)
	}

	var unitsToCheck []int
	// Определяем, для каких юнитов нужно проверять пересечения
	// Если это Сектор или Отдел без секторов (предполагаем, что Отдел - конечный уровень, если нет дочерних секторов)
	// TODO: Уточнить типы юнитов, которые являются "конечными" для проверки пересечений
	if targetUnit.UnitType == "SECTOR" || targetUnit.UnitType == "DEPARTMENT" || targetUnit.UnitType == "SUB_DEPARTMENT" || targetUnit.UnitType == "OFFICE" || targetUnit.UnitType == "CENTER" {
		// Проверяем только этот конкретный юнит
		unitsToCheck = append(unitsToCheck, unitID)
	} else {
		// Для юнитов более высокого уровня (MANAGEMENT, REPRESENTATION) нужно найти все конечные юниты в поддереве
		subtreeIDs, err := s.unitRepo.GetSubtreeIDs(unitID)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения поддерева для юнита %d: %w", unitID, err)
		}
		// Получаем информацию о всех юнитах поддерева, чтобы определить конечные
		allUnitsInSubtree := make(map[int]*models.OrganizationalUnit)
		allUnits, err := s.unitRepo.GetAll() // Получаем все юниты для построения карты
		if err != nil {
			return nil, fmt.Errorf("ошибка получения всех юнитов для определения конечных: %w", err)
		}
		tempUnitMap := make(map[int]*models.OrganizationalUnit)
		for _, u := range allUnits {
			tempUnitMap[u.ID] = u
		}
		for _, id := range subtreeIDs {
			if u, ok := tempUnitMap[id]; ok {
				allUnitsInSubtree[id] = u
			}
		}

		// Определяем юниты, внутри которых нужно проверять пересечения
		potentialLeafUnits := make(map[int]bool)
		hasChildren := make(map[int]bool)
		for _, u := range allUnitsInSubtree {
			if u.ParentID != nil {
				hasChildren[*u.ParentID] = true
			}
			// Кандидаты: Секторы, Отделы, Офисы, ЦОКи
			if u.UnitType == "SECTOR" || u.UnitType == "DEPARTMENT" || u.UnitType == "SUB_DEPARTMENT" || u.UnitType == "OFFICE" || u.UnitType == "CENTER" {
				potentialLeafUnits[u.ID] = true
			}
		}
		// Исключаем те, у которых есть дочерние юниты (из кандидатов)
		for id := range potentialLeafUnits {
			if !hasChildren[id] {
				unitsToCheck = append(unitsToCheck, id)
			}
		}
		if len(unitsToCheck) == 0 && len(allUnitsInSubtree) > 0 { // Если нет явных листьев, берем сам юнит
			unitsToCheck = append(unitsToCheck, unitID)
		}
	}

	log.Printf("[CheckIntersections] Units identified for intersection check: %v (for initial unit %d)", unitsToCheck, unitID)

	var allIntersections []models.Intersection
	// Проверяем пересечения для каждого определенного юнита
	for _, checkUnitID := range unitsToCheck {
		requests, err := s.vacationRepo.GetVacationRequestsByOrganizationalUnit(checkUnitID, year, nil) // Заявки только этого юнита
		if err != nil {
			log.Printf("Ошибка получения заявок для юнита %d при проверке пересечений: %v", checkUnitID, err)
			continue // Пропускаем этот юнит при ошибке
		}
		if len(requests) < 2 {
			continue
		} // Нет смысла проверять, если меньше 2 заявок

		users, err := s.userRepo.GetUsersByOrganizationalUnit(checkUnitID) // Пользователи только этого юнита
		if err != nil {
			log.Printf("Ошибка получения пользователей для юнита %d при проверке пересечений: %v", checkUnitID, err)
			continue
		}
		userMap := make(map[int]models.User)
		for _, user := range users {
			userMap[user.ID] = user
		}

		// Логика поиска пересечений внутри checkUnitID (как было раньше, но для requests/userMap этого юнита)
		for i, req1 := range requests {
			for j := i + 1; j < len(requests); j++ {
				req2 := requests[j]
				if req1.UserID == req2.UserID {
					continue
				}
				for _, period1 := range req1.Periods {
					for _, period2 := range req2.Periods {
						if doPeriodIntersect(period1, period2) {
							start := max(period1.StartDate.Time, period2.StartDate.Time)
							end := min(period1.EndDate.Time, period2.EndDate.Time)
							daysCount := int(end.Sub(start).Hours()/24) + 1
							intersection := models.Intersection{
								UserID1: req1.UserID, UserName1: userMap[req1.UserID].FullName,
								UserID2: req2.UserID, UserName2: userMap[req2.UserID].FullName,
								StartDate: models.CustomDate{Time: start}, EndDate: models.CustomDate{Time: end},
								DaysCount: daysCount,
							}
							allIntersections = append(allIntersections, intersection)
						}
					}
				}
			}
		}
	}

	return allIntersections, nil
}

// NotifyManager уведомляет руководителя о пересечениях отпусков
func (s *VacationService) NotifyManager(managerID int, intersections []models.Intersection) error {
	if len(intersections) == 0 {
		return nil
	}
	notification := &models.Notification{
		UserID: managerID, Title: "Обнаружено пересечение отпусков",
		Message: "В подразделении обнаружены пересечения отпусков сотрудников. Требуется ваше внимание.",
		IsRead:  false, CreatedAt: time.Now(),
	}
	return s.vacationRepo.CreateNotification(notification)
}

// GetUserVacations получает заявки конкретного пользователя
func (s *VacationService) GetUserVacations(userID int, year int, statusFilter *int) ([]models.VacationRequest, error) {
	return s.vacationRepo.GetVacationRequestsByUser(userID, year, statusFilter)
}

// GetOrganizationalUnitVacations получает заявки орг. юнита
func (s *VacationService) GetOrganizationalUnitVacations(unitID int, year int, statusFilter *int) ([]models.VacationRequest, error) {
	return s.vacationRepo.GetVacationRequestsByOrganizationalUnit(unitID, year, statusFilter)
}

// GetAllUserVacations получает все заявки (для админов) или заявки своего поддерева (для менеджеров)
func (s *VacationService) GetAllUserVacations(requestingUserID int, yearFilter *int, statusFilter *int, userIDFilter *int, unitIDFilter *int) ([]models.VacationRequestAdminView, error) {
	requestingUser, err := s.userRepo.FindByID(requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения данных запрашивающего пользователя: %w", err)
	}
	if requestingUser == nil {
		return nil, errors.New("запрашивающий пользователь не найден")
	}

	var unitIDsFilterForRepo []int
	if !requestingUser.IsAdmin {
		if requestingUser.IsManager {
			if requestingUser.OrganizationalUnitID == nil {
				return []models.VacationRequestAdminView{}, nil
			}
			subtreeIDs, err := s.unitRepo.GetSubtreeIDs(*requestingUser.OrganizationalUnitID)
			if err != nil {
				log.Printf("[GetAllUserVacations] Error getting subtree for manager %d (unit %d): %v", requestingUser.ID, *requestingUser.OrganizationalUnitID, err)
				return nil, fmt.Errorf("ошибка получения подчиненных юнитов: %w", err)
			}
			unitIDsFilterForRepo = subtreeIDs
			if unitIDFilter != nil { // Если менеджер дополнительно фильтрует по юниту
				requestedUnitID := *unitIDFilter
				found := false
				for _, allowedID := range subtreeIDs {
					if allowedID == requestedUnitID {
						found = true
						break
					}
				}
				if !found {
					return nil, errors.New("менеджер может фильтровать заявки только по своему юниту или подчиненным")
				}
				unitIDsFilterForRepo = []int{requestedUnitID} // Используем только запрошенный ID
			}
		} else {
			return nil, errors.New("недостаточно прав для просмотра всех заявок")
		}
	} else { // Админ
		if unitIDFilter != nil {
			unitIDsFilterForRepo = []int{*unitIDFilter}
		}
		// Если админ не передал фильтр, unitIDsFilterForRepo остается nil (без фильтрации по юнитам)
	}

	return s.vacationRepo.GetAllVacationRequests(yearFilter, statusFilter, userIDFilter, unitIDsFilterForRepo)
}

// CancelVacationRequest отменяет заявку
func (s *VacationService) CancelVacationRequest(requestID int, cancellingUserID int) error {
	req, err := s.vacationRepo.GetVacationRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("ошибка получения заявки для отмены: %w", err)
	}
	if req == nil {
		return errors.New("заявка не найдена")
	}

	canCancel := false
	if req.UserID == cancellingUserID && (req.StatusID == models.StatusDraft || req.StatusID == models.StatusPending) {
		canCancel = true
	} else {
		cancellingUser, err := s.userRepo.FindByID(cancellingUserID)
		if err != nil || cancellingUser == nil {
			return errors.New("не удалось проверить права пользователя на отмену")
		}
		if cancellingUser.IsAdmin {
			canCancel = true
		} else if cancellingUser.IsManager {
			employee, err := s.userRepo.FindByID(req.UserID)
			accessGranted, accessErr := s.checkUserUnitAccess(cancellingUser, employee)
			if accessErr != nil {
				return fmt.Errorf("ошибка проверки доступа для отмены: %w", accessErr)
			}
			if err == nil && employee != nil && accessGranted {
				canCancel = true
			}
		}
	}
	if !canCancel {
		return errors.New("нет прав на отмену этой заявки")
	}
	if req.StatusID == models.StatusRejected || req.StatusID == models.StatusCancelled {
		return fmt.Errorf("нельзя отменить заявку ID %d в статусе '%d'", requestID, req.StatusID)
	}

	originalStatus := req.StatusID
	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusCancelled)
	if err != nil {
		return fmt.Errorf("ошибка установки статуса 'Отменена' для заявки %d: %w", requestID, err)
	}

	if (originalStatus == models.StatusPending || originalStatus == models.StatusApproved) && req.DaysRequested > 0 {
		log.Printf("[Service CancelVacationRequest] Attempting to return days. UserID: %d, Year: %d, RequestID: %d, Days: %d", req.UserID, req.Year, requestID, req.DaysRequested)
		errReturn := s.vacationRepo.UpdateVacationLimitUsedDays(req.UserID, req.Year, -req.DaysRequested)
		if errReturn != nil {
			log.Printf("[Service CancelVacationRequest] CRITICAL ERROR: Failed to return days! UserID: %d, Year: %d, RequestID: %d, Days: %d, Error: %v", req.UserID, req.Year, requestID, req.DaysRequested, errReturn)
		} else {
			log.Printf("[Service CancelVacationRequest] Successfully returned days. UserID: %d, Year: %d, RequestID: %d, Days: %d", req.UserID, req.Year, requestID, req.DaysRequested)
		}
	}
	// TODO: Notify user
	return nil
}

// ApproveVacationRequest утверждает заявку и возвращает список конфликтов (предупреждений)
func (s *VacationService) ApproveVacationRequest(requestID int, approverID int) ([]models.ConflictingPeriod, error) {
	req, err := s.vacationRepo.GetVacationRequestByID(requestID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заявки ID %d для утверждения: %w", requestID, err)
	}
	if req == nil {
		return nil, fmt.Errorf("заявка ID %d не найдена", requestID)
	}

	// --- Проверка прав на утверждение (как и раньше) ---
	approver, err := s.userRepo.FindByID(approverID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки прав пользователя ID %d: %w", approverID, err)
	}
	if approver == nil {
		return nil, fmt.Errorf("утверждающий пользователь ID %d не найден", approverID)
	}

	canApprove := false
	if approver.IsAdmin {
		canApprove = true
	} else if approver.IsManager {
		employee, errUser := s.userRepo.FindByID(req.UserID) // Переименована переменная ошибки
		if errUser != nil {
			log.Printf("[ApproveVacationRequest] Warning: could not get employee %d data to check unit access: %v", req.UserID, errUser)
			return nil, fmt.Errorf("ошибка получения данных сотрудника %d для проверки доступа: %w", req.UserID, errUser)
		}
		if employee == nil {
			return nil, fmt.Errorf("сотрудник %d, подавший заявку, не найден", req.UserID)
		}
		accessGranted, accessErr := s.checkUserUnitAccess(approver, employee)
		if accessErr != nil {
			return nil, fmt.Errorf("ошибка проверки доступа для утверждения: %w", accessErr)
		}
		if accessGranted {
			canApprove = true
		}
	}
	if !canApprove {
		log.Printf("[ApproveVacationRequest] Access denied: User %d (admin: %t, manager: %t, unit: %v) cannot approve request %d for user %d", approver.ID, approver.IsAdmin, approver.IsManager, approver.OrganizationalUnitID, requestID, req.UserID)
		return nil, fmt.Errorf("пользователь ID %d не имеет прав для утверждения заявки ID %d", approverID, requestID)
	}
	if req.StatusID != models.StatusPending {
		return nil, fmt.Errorf("можно утвердить только заявку ID %d в статусе 'На рассмотрении' (текущий статус: %d)", requestID, req.StatusID)
	}

	// --- Проверка конфликтов ПЕРЕД утверждением ---
	var conflicts []models.ConflictingPeriod
	positionID, errPos := s.vacationRepo.GetUserPositionByID(req.UserID)
	if errPos != nil {
		// Логируем ошибку, но не прерываем утверждение, если должность не удалось получить
		log.Printf("[ApproveVacationRequest] Warning: could not get position for user %d while checking conflicts for request %d: %v", req.UserID, requestID, errPos)
	} else if positionID != nil {
		// Если должность есть, проверяем конфликты
		log.Printf("[ApproveVacationRequest] Checking conflicts for request %d (user %d, position %d)", requestID, req.UserID, *positionID)
		conflicts, err = s.vacationRepo.GetApprovedVacationConflictsByPosition(*positionID, req.UserID, req.Periods)
		if err != nil {
			// Логируем ошибку проверки конфликтов, но не прерываем утверждение
			log.Printf("[ApproveVacationRequest] Error checking conflicts for request %d: %v", requestID, err)
			// Очищаем конфликты на всякий случай, чтобы не вернуть ошибочные данные
			conflicts = []models.ConflictingPeriod{}
			// Можно вернуть ошибку, если считаем это критичным:
			// return nil, fmt.Errorf("ошибка проверки конфликтов отпусков: %w", err)
		} else if len(conflicts) > 0 {
			log.Printf("[ApproveVacationRequest] Found %d conflicts for request %d", len(conflicts), requestID)
		}
	} else {
		log.Printf("[ApproveVacationRequest] Skipping conflict check for request %d as user %d has no position assigned.", requestID, req.UserID)
	}

	// --- Утверждение заявки (установка статуса) ---
	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusApproved)
	if err != nil {
		log.Printf("WARNING: Failed to set status 'Approved' for request %d (user %d, year %d), but days were likely spent on submission: %v\n", requestID, req.UserID, req.Year, err)
		// В случае ошибки утверждения, не возвращаем конфликты, а только ошибку
		return nil, fmt.Errorf("ошибка установки статуса 'Утверждена' для заявки %d: %w", requestID, err)
	}

	log.Printf("[ApproveVacationRequest] Successfully approved request %d. Returning %d conflicts as warnings.", requestID, len(conflicts))

	// TODO: Notify user об утверждении

	// Возвращаем найденные конфликты (если есть) и nil в качестве ошибки, т.к. утверждение прошло успешно
	return conflicts, nil
}

// RejectVacationRequest отклоняет заявку
func (s *VacationService) RejectVacationRequest(requestID int, rejecterID int, reason string) error {
	req, err := s.vacationRepo.GetVacationRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("ошибка получения заявки ID %d для отклонения: %w", requestID, err)
	}
	if req == nil {
		return fmt.Errorf("заявка ID %d не найдена", requestID)
	}

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
		accessGranted, accessErr := s.checkUserUnitAccess(rejecter, employee)
		if accessErr != nil {
			return fmt.Errorf("ошибка проверки доступа для отклонения: %w", accessErr)
		}
		if err == nil && employee != nil && accessGranted {
			canReject = true
		} else if err != nil {
			log.Printf("[RejectVacationRequest] Warning: could not get employee %d data to check unit access: %v", req.UserID, err)
		}
	}
	if !canReject {
		return fmt.Errorf("пользователь ID %d не имеет прав для отклонения заявки ID %d", rejecterID, requestID)
	}
	if req.StatusID != models.StatusPending {
		return fmt.Errorf("можно отклонить только заявку ID %d в статусе 'На рассмотрении' (текущий статус: %d)", requestID, req.StatusID)
	}

	err = s.vacationRepo.UpdateRequestStatusByID(requestID, models.StatusRejected)
	if err != nil {
		return fmt.Errorf("ошибка установки статуса 'Отклонена' для заявки %d: %w", requestID, err)
	}

	if req.DaysRequested > 0 {
		log.Printf("[Service RejectVacationRequest] Attempting to return days. UserID: %d, Year: %d, RequestID: %d, Days: %d", req.UserID, req.Year, requestID, req.DaysRequested)
		errReturn := s.vacationRepo.UpdateVacationLimitUsedDays(req.UserID, req.Year, -req.DaysRequested)
		if errReturn != nil {
			log.Printf("[Service RejectVacationRequest] CRITICAL ERROR: Failed to return days! UserID: %d, Year: %d, RequestID: %d, Days: %d, Error: %v", req.UserID, req.Year, requestID, req.DaysRequested, errReturn)
			return fmt.Errorf("заявка отклонена, но произошла ошибка при возврате дней в лимит: %w", errReturn)
		} else {
			log.Printf("[Service RejectVacationRequest] Successfully returned days. UserID: %d, Year: %d, RequestID: %d, Days: %d", req.UserID, req.Year, requestID, req.DaysRequested)
		}
	} else {
		log.Printf("[Service RejectVacationRequest] No days to return for RequestID: %d (DaysRequested: %d)", requestID, req.DaysRequested)
	}

	// TODO: Save rejection reason
	// TODO: Notify user
	return nil
}

// --- Вспомогательные функции ---
func doPeriodIntersect(p1, p2 models.VacationPeriod) bool {
	return p1.StartDate.Time.Before(p2.EndDate.Time) && p2.StartDate.Time.Before(p1.EndDate.Time)
}
func max(t1, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}
func min(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}
	return t2
}
