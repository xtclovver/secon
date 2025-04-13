import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { toast } from 'react-toastify';
import Calendar from 'react-calendar';
import { FaExclamationTriangle, FaCheck, FaUser, FaCalendarAlt, FaUsers, FaClock, FaThumbsUp, FaThumbsDown } from 'react-icons/fa'; // Добавлены иконки
import { getManagerDashboardData } from '../../api/vacations'; // Используем новую функцию
import 'react-calendar/dist/Calendar.css';
import './ManagerDashboard.css'; // Предполагается, что CSS файл будет создан

// Вспомогательная функция для форматирования дат
const formatDate = (dateString) => {
  if (!dateString) return '';
  return new Date(dateString).toLocaleDateString('ru-RU');
};

const ManagerDashboard = () => {
  const [year, setYear] = useState(new Date().getFullYear()); // По умолчанию текущий год
  // const [unitId, setUnitId] = useState(1); // Больше не нужно, бэкенд определяет по токену
  // const [vacations, setVacations] = useState([]); // Заменено на dashboardData
  // const [intersections, setIntersections] = useState([]); // Заменено на dashboardData.upcomingConflicts
  const [dashboardData, setDashboardData] = useState(null); // Новое состояние для данных дашборда
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('stats'); // 'stats', 'calendar' или 'conflicts'
  const [calendarDate, setCalendarDate] = useState(new Date());


  // Загрузка данных при изменении года
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        // Получение данных дашборда
        const data = await getManagerDashboardData(); // Год пока не передаем, бэкенд использует текущий
        setDashboardData(data);
      } catch (error) {
        toast.error(`Ошибка при загрузке данных дашборда: ${error.message}`);
        console.error(error);
        setDashboardData(null); // Сбрасываем данные при ошибке
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []); // Загружаем один раз при монтировании (год пока не используется в API)

  // Функция для отображения отпусков в календаре
  // TODO: Для отображения ВСЕХ отпусков (FaUser) нужно будет отдельно загружать
  //       данные getApprovedVacationsForCalendar или getAllVacations с фильтром.
  //       Пока оставим только маркеры конфликтов.
  const getTileContent = ({ date, view }) => {
    if (view !== 'month' || !dashboardData?.upcomingConflicts) return null;

    const dateString = date.toISOString().split('T')[0]; // Формат YYYY-MM-DD

    // Находим конфликты на эту дату
    const conflictsOnDate = dashboardData.upcomingConflicts.filter(conflict => {
       // Проверяем пересечение с датой календаря
       const overlapStart = new Date(conflict.overlapStartDate);
       const overlapEnd = new Date(conflict.overlapEndDate);
       const currentDate = new Date(dateString);
       // Устанавливаем время на 00:00:00 для корректного сравнения дат
       overlapStart.setHours(0, 0, 0, 0);
       overlapEnd.setHours(0, 0, 0, 0);
       currentDate.setHours(0, 0, 0, 0);
       return currentDate >= overlapStart && currentDate <= overlapEnd;
    });

    if (conflictsOnDate.length > 0) {
      return (
        <div className="calendar-marker intersection-marker">
          <FaExclamationTriangle />
        </div>
      );
    }

    // TODO: Добавить логику для маркера обычного отпуска (FaUser),
    //       когда будут загружаться все утвержденные отпуска.

    return null;
  };

  // Функция для определения класса даты в календаре
  const getTileClassName = ({ date, view }) => {
    if (view !== 'month' || !dashboardData?.upcomingConflicts) return '';

    const dateString = date.toISOString().split('T')[0]; // Формат YYYY-MM-DD

    // Проверяем, есть ли конфликты на эту дату
    const hasConflicts = dashboardData.upcomingConflicts.some(conflict => {
       const overlapStart = new Date(conflict.overlapStartDate);
       const overlapEnd = new Date(conflict.overlapEndDate);
       const currentDate = new Date(dateString);
       overlapStart.setHours(0, 0, 0, 0);
       overlapEnd.setHours(0, 0, 0, 0);
       currentDate.setHours(0, 0, 0, 0);
       return currentDate >= overlapStart && currentDate <= overlapEnd;
    });

    if (hasConflicts) {
      return 'intersection-date';
    }

    // TODO: Добавить класс 'vacation-date' для обычных отпусков,
    //       когда будут загружаться все утвержденные отпуска.

    return '';
  };

  // Компонент для отображения статистики
  const StatsDisplay = () => {
    if (!dashboardData) return <p>Нет данных для отображения.</p>;

    return (
      <motion.div
        className="stats-container card"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2 }}
      >
        <h3>Статистика ({year})</h3>
        <div className="stats-grid">
          <div className="stat-item">
            <FaClock className="stat-icon pending" />
            <span className="stat-value">{dashboardData.pendingRequestsCount ?? 'N/A'}</span>
            <span className="stat-label">Заявок на рассмотрении</span>
          </div>
           <div className="stat-item">
            <FaUsers className="stat-icon users" />
            <span className="stat-value">{dashboardData.subordinateUserCount ?? 'N/A'}</span>
            <span className="stat-label">Подчиненных</span>
          </div>
          <div className="stat-item">
            <FaThumbsUp className="stat-icon approved" />
            <span className="stat-value">{dashboardData.approvedDaysCountYear ?? 'N/A'}</span>
            <span className="stat-label">Утверждено дней</span>
          </div>
          <div className="stat-item">
            <FaThumbsDown className="stat-icon rejected" />
            <span className="stat-value">{dashboardData.rejectedDaysCountYear ?? 'N/A'}</span>
            <span className="stat-label">Отклонено дней</span>
          </div>
          <div className="stat-item">
            <FaClock className="stat-icon pending-days" />
            <span className="stat-value">{dashboardData.pendingDaysCountYear ?? 'N/A'}</span>
            <span className="stat-label">Дней на рассмотрении</span>
          </div>
        </div>
      </motion.div>
    );
  };

  return (
    <motion.div
      className="manager-dashboard"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
    >
      <h2>Дашборд руководителя</h2>

      <div className="dashboard-controls card">
        {/* <div className="year-selector">
          <label htmlFor="manager-year">Год:</label>
          <select
            id="manager-year"
            value={year}
            onChange={(e) => setYear(parseInt(e.target.value))}
            disabled={loading}
          >
            <option value={new Date().getFullYear()}>Текущий год</option>
            <option value={new Date().getFullYear() + 1}>Следующий год</option>
          </select>
        </div> */}
        {/* Селектор года пока убран, т.к. API его не использует */}

        <div className="tab-selector">
           <button
            className={`tab-button btn ${activeTab === 'stats' ? 'active' : ''}`}
            onClick={() => setActiveTab('stats')}
          >
             Статистика
          </button>
          <button
            className={`tab-button btn ${activeTab === 'calendar' ? 'active' : ''}`}
            onClick={() => setActiveTab('calendar')}
          >
            <FaCalendarAlt /> Календарь
          </button>
          <button
            className={`tab-button btn ${activeTab === 'conflicts' ? 'active' : ''}`}
            onClick={() => setActiveTab('conflicts')}
          >
            <FaExclamationTriangle /> Конфликты ({dashboardData?.upcomingConflicts?.length ?? 0})
          </button>
        </div>
      </div>

      {loading ? (
        <div className="loading-spinner">Загрузка данных...</div>
      ) : (
        <>
          {activeTab === 'stats' && <StatsDisplay />}

          {activeTab === 'calendar' && (
            <motion.div
              className="calendar-container card"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
            >
              <Calendar
                onChange={setCalendarDate}
                value={calendarDate}
                tileContent={getTileContent}
                tileClassName={getTileClassName}
                locale="ru-RU" // Установка русской локали
                className="department-calendar"
              />

              <div className="calendar-legend">
                {/* <div className="legend-item">
                  <div className="legend-color vacation"></div>
                  <span>Сотрудники в отпуске</span>
                </div> */}
                <div className="legend-item">
                  <div className="legend-color intersection"></div>
                  <span>Конфликты отпусков</span>
                </div>
              </div>
            </motion.div>
          )}

          {activeTab === 'conflicts' && (
            <motion.div
              className="conflicts-container card" // Переименован класс
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
            >
              <h3>Предстоящие конфликты отпусков</h3>

              {!dashboardData || !dashboardData.upcomingConflicts || dashboardData.upcomingConflicts.length === 0 ? (
                <div className="no-conflicts"> {/* Переименован класс */}
                  <FaCheck />
                  <p>Конфликтов отпусков не найдено</p>
                </div>
              ) : (
                <div className="conflicts-list"> {/* Переименован класс */}
                  {dashboardData.upcomingConflicts.map((conflict, index) => (
                    <motion.div
                      key={`${conflict.originalPeriodID}-${conflict.conflictingPeriodID}-${index}`} // Более надежный ключ
                      className="conflict-item" // Переименован класс
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: 0.1 * index }}
                    >
                      <div className="conflict-header"> {/* Переименован класс */}
                        <FaExclamationTriangle className="conflict-icon" /> {/* Переименован класс */}
                        <span>Конфликт #{index + 1}</span>
                      </div>
                      <div className="conflict-details"> {/* Переименован класс */}
                        <div className="conflict-users"> {/* Переименован класс */}
                          <div className="conflict-user"> {/* Переименован класс */}
                            <FaUser />
                            {/* Используем поля из ConflictingPeriod */}
                            <span>{conflict.originalUserFullName ?? `User ID: ${conflict.originalUserID}`}</span>
                            <small> (Заявка #{conflict.originalRequestID})</small>
                          </div>
                          <div className="conflict-user"> {/* Переименован класс */}
                            <FaUser />
                            <span>{conflict.conflictingUserFullName ?? `User ID: ${conflict.conflictingUserID}`}</span>
                             <small> (Заявка #{conflict.conflictingRequestID})</small>
                          </div>
                        </div>
                        <div className="conflict-dates"> {/* Переименован класс */}
                           <div>
                            <strong>Период 1:</strong> {formatDate(conflict.originalStartDate)} - {formatDate(conflict.originalEndDate)}
                          </div>
                           <div>
                            <strong>Период 2:</strong> {formatDate(conflict.conflictingStartDate)} - {formatDate(conflict.conflictingEndDate)}
                          </div>
                          <div>
                            <strong>Пересечение:</strong>
                            {formatDate(conflict.overlapStartDate)} - {formatDate(conflict.overlapEndDate)}
                          </div>
                        </div>
                      </div>
                    </motion.div>
                  ))}
                </div>
              )}
            </motion.div>
          )}
        </>
      )}
    </motion.div>
  );
};

export default ManagerDashboard;
