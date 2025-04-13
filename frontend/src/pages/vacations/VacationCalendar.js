import React, { useState, useEffect, useCallback } from 'react'; // Добавлен useCallback
import { motion } from 'framer-motion';
import Calendar from 'react-calendar';
import { toast } from 'react-toastify';
import { FaCalendarAlt, FaUsers, FaInfoCircle, FaExclamationTriangle, FaUser } from 'react-icons/fa'; // Добавлены иконки
// Обновленные импорты API
import { getApprovedVacationsForCalendar, getVacationLimit, getVacationConflicts } from '../../api/vacations'; // Добавлен getVacationConflicts
import Loader from '../../components/ui/Loader/Loader';
import { useUser } from '../../context/UserContext';
import './VacationCalendar.css';

// Вспомогательная функция для форматирования дат
const formatDate = (dateString) => {
  if (!dateString) return '';
  return new Date(dateString).toLocaleDateString('ru-RU');
};

// Вспомогательная функция для проверки, находится ли дата в периоде (сравнение UTC)
// Используется для отпусков (где не было проблем)
const isDateInPeriod = (date, period) => {
    if (!period || !period.start_date || !period.end_date) return false;
    try {
        const startDate = new Date(period.start_date);
        const endDate = new Date(period.end_date);
        startDate.setHours(0, 0, 0, 0);
        endDate.setHours(0, 0, 0, 0);
        const currentDate = new Date(date);
        currentDate.setHours(0, 0, 0, 0);
        return currentDate >= startDate && currentDate <= endDate;
    } catch (e) {
        console.error("Error in isDateInPeriod:", e, date, period);
        return false;
    }
};

// Вспомогательная функция для проверки конфликтов (сравнение UTC)
const isDateInConflictRange = (date, conflict) => {
    if (!conflict || !conflict.overlapStartDate || !conflict.overlapEndDate) return false;
    try {
        // Получаем метку времени UTC для полуночи даты календаря
        const calendarDateUTCTimestamp = Date.UTC(date.getFullYear(), date.getMonth(), date.getDate());

        // Даты конфликта уже в UTC, парсим и получаем метку времени UTC
        const overlapStartDate = new Date(conflict.overlapStartDate);
        const overlapEndDate = new Date(conflict.overlapEndDate);

        // Проверяем валидность дат
        if (isNaN(overlapStartDate.getTime()) || isNaN(overlapEndDate.getTime())) {
            console.error("Invalid conflict date format received:", conflict);
            return false;
        }

        const overlapStartUTCTimestamp = overlapStartDate.getTime();
        // Получаем UTC timestamp для НАЧАЛА ДНЯ *ПОСЛЕ* даты окончания конфликта
        const dayAfterOverlapEnd = new Date(overlapEndDate);
        dayAfterOverlapEnd.setUTCDate(dayAfterOverlapEnd.getUTCDate() + 1);
        const dayAfterOverlapEndUTCTimestamp = Date.UTC(
            dayAfterOverlapEnd.getUTCFullYear(),
            dayAfterOverlapEnd.getUTCMonth(),
            dayAfterOverlapEnd.getUTCDate()
        );

        // Сравниваем: дата >= начала И дата < начала следующего дня после конца
        return calendarDateUTCTimestamp >= overlapStartUTCTimestamp && calendarDateUTCTimestamp < dayAfterOverlapEndUTCTimestamp;
    } catch (e) {
        console.error("Error in isDateInConflictRange:", e, date, conflict);
        return false;
    }
};


const VacationCalendar = () => {
  const { user } = useUser();
  const [year, setYear] = useState(new Date().getFullYear());
  const [approvedVacations, setApprovedVacations] = useState([]); // Состояние для утвержденных отпусков
  const [conflicts, setConflicts] = useState([]); // Состояние для конфликтов
  const [loadingVacations, setLoadingVacations] = useState(false);
  const [loadingConflicts, setLoadingConflicts] = useState(false); // Состояние загрузки конфликтов
  const [errorVacations, setErrorVacations] = useState(null);
  const [errorConflicts, setErrorConflicts] = useState(null); // Состояние ошибки конфликтов
  const [activeStartDate, setActiveStartDate] = useState(new Date()); // Для отслеживания видимого месяца

  // --- Функции загрузки данных ---
  const fetchCalendarData = useCallback(async (currentYear) => {
    if (!user) return;
    setLoadingVacations(true);
    setErrorVacations(null);
    try {
      const filters = { year: currentYear };
      const data = await getApprovedVacationsForCalendar(filters);
      setApprovedVacations(data || []);
    } catch (err) {
      const errorMsg = err.message || 'Не удалось загрузить данные календаря.';
      setErrorVacations(errorMsg);
      toast.error(errorMsg);
    } finally {
      setLoadingVacations(false);
    }
  }, [user]);

  const fetchConflicts = useCallback(async (startDate, endDate) => {
    if (!user) return;
    setLoadingConflicts(true);
    setErrorConflicts(null);
    try {
      const startStr = startDate.toISOString().split('T')[0];
      const endStr = endDate.toISOString().split('T')[0];
      const conflictsData = await getVacationConflicts(startStr, endStr);
      setConflicts(conflictsData || []);
    } catch (err) {
      const errorMsg = err.message || 'Не удалось загрузить конфликты отпусков.';
      setErrorConflicts(errorMsg);
      toast.error(errorMsg);
    } finally {
      setLoadingConflicts(false);
    }
  }, [user]);

  // --- Цветовая палитра для пользователей ---
  const userColorPalette = [
    '#3498db', '#e74c3c', '#2ecc71', '#f1c40f', '#9b59b6',
    '#1abc9c', '#e67e22', '#34495e', '#f39c12', '#d35400'
  ];

  // Функция для получения цвета пользователя
  const getUserColor = (userId) => {
    if (!userId) return '#bdc3c7'; // Серый по умолчанию
    const index = Math.abs(userId) % userColorPalette.length;
    return userColorPalette[index];
  };

  // Загрузка данных при монтировании и смене года
  useEffect(() => {
    fetchCalendarData(year);
  }, [year, fetchCalendarData]);

  // Загрузка конфликтов при изменении видимого месяца
  useEffect(() => {
    const startOfMonth = new Date(activeStartDate.getFullYear(), activeStartDate.getMonth(), 1);
    const endOfMonth = new Date(activeStartDate.getFullYear(), activeStartDate.getMonth() + 1, 0);
    fetchConflicts(startOfMonth, endOfMonth);
  }, [activeStartDate, fetchConflicts]);


  // Функция для отображения маркеров в календаре
  const getTileContent = ({ date, view }) => {
    if (view !== 'month') return null;

    const usersOnVacation = approvedVacations.filter(vacation =>
        Array.isArray(vacation.periods) && vacation.periods.some(period =>
            isDateInPeriod(date, period) // Используем старую функцию для отпусков
        )
    );

    const conflictsOnDate = conflicts.filter(conflict =>
        isDateInConflictRange(date, conflict) // Используем новую функцию для конфликтов
    );

    const markers = [];

    if (conflictsOnDate.length > 0) {
        markers.push(
            <div key="conflict-marker" className="calendar-marker conflict-marker" title="Конфликт!">
                <FaExclamationTriangle />
            </div>
     );
    }

    if (usersOnVacation.length > 0) {
      const maxVisibleMarkers = 3;
      usersOnVacation.slice(0, maxVisibleMarkers).forEach((vacation, index) => {
        const userColor = getUserColor(vacation.user_id);
        const userName = vacation.user_full_name || `ID: ${vacation.user_id}`;
        markers.push(
          <div
            key={`user-${vacation.user_id}-${index}`}
            className="calendar-marker user-dot-marker"
            style={{ backgroundColor: userColor }}
            title={userName}
          >
          </div>
        );
      });
      if (usersOnVacation.length > maxVisibleMarkers) {
         const remainingCount = usersOnVacation.length - maxVisibleMarkers;
         const remainingNames = usersOnVacation.slice(maxVisibleMarkers).map(v => v.user_full_name || `ID: ${v.user_id}`).join('\n');
         markers.push(
           <div key="more-users" className="calendar-marker more-users-marker" title={`Еще ${remainingCount}:\n${remainingNames}`}>
             +{remainingCount}
           </div>
         );
      }
    }

    return markers.length > 0 ? <div className="markers-container">{markers}</div> : null;
  };

  // Функция для определения класса даты в календаре
  const getTileClassName = ({ date, view }) => {
     if (view !== 'month') return '';

    const isOnVacation = approvedVacations.some(vacation =>
        Array.isArray(vacation.periods) && vacation.periods.some(period =>
            isDateInPeriod(date, period) // Используем старую функцию для отпусков
        )
    );

     const hasConflicts = conflicts.some(conflict =>
        isDateInConflictRange(date, conflict) // Используем новую функцию для конфликтов
     );

    let className = '';
    if (hasConflicts) {
      className += ' conflict-date';
    } else if (isOnVacation) {
      className += ' department-vacation-date';
    }

    return className.trim();
  };

  const isLoading = loadingVacations || loadingConflicts;

  return (
    <motion.div
      className="vacation-calendar-container card"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
    >
      <h2><FaCalendarAlt /> Календарь отпусков</h2>

       <div className="controls" style={{ marginBottom: '20px', display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '15px' }}>
         <div className="year-filter" style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
           <label htmlFor="year-select">Год:</label>
           <select
             id="year-select"
             value={year}
             onChange={(e) => {
               const newYear = parseInt(e.target.value, 10);
               setYear(newYear);
               setActiveStartDate(new Date(newYear, 0, 1));
             }}
           >
             {[0, 1, 2, 3].map(offset => {
               const currentYear = new Date().getFullYear();
               const optionYear = currentYear + offset;
               return <option key={optionYear} value={optionYear}>{optionYear}</option>;
             })}
           </select>
         </div>
       </div>

      {isLoading && <Loader text="Загрузка данных..." />}
      {errorVacations && <div className="error-message">{errorVacations}</div>}
      {errorConflicts && <div className="error-message">{errorConflicts}</div>}

      {!isLoading && !errorVacations && (
         <>
             <motion.div
                className="calendar-view card"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.2 }}
             >
                <Calendar
                    value={activeStartDate}
                    onActiveStartDateChange={({ activeStartDate }) => setActiveStartDate(activeStartDate)}
                    tileContent={getTileContent}
                    tileClassName={getTileClassName}
                    locale="ru-RU"
                    className="department-wide-calendar"
                />
                 <div className="calendar-info-section">
                    <div className="calendar-legend">
                        <h4><FaInfoCircle /> Легенда:</h4>
                        <span className="legend-item">
                            <div className="legend-color-box department-vacation-date"></div>
                            Сотрудник(и) в отпуске
                        </span>
                         <span className="legend-item">
                            <div className="legend-color-box conflict-date"></div>
                            Конфликт отпусков
                         </span>
                    </div>
                 </div>
             </motion.div>

             <motion.div
                className="conflicts-display-section card"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.4 }}
             >
                <h3><FaExclamationTriangle /> Конфликты в текущем месяце</h3>
                {loadingConflicts && <Loader text="Загрузка конфликтов..." />}
                {!loadingConflicts && errorConflicts && <div className="error-message">{errorConflicts}</div>}
                {!loadingConflicts && !errorConflicts && (
                    conflicts.length === 0 ? (
                        <p>Конфликтов в этом месяце не найдено.</p>
                    ) : (
                        <div className="conflicts-list-container">
                            {conflicts.map((conflict, index) => (
                                <motion.div
                                    key={`${conflict.original_period_id}-${conflict.conflicting_period_id}-${index}`}
                                    className="conflict-item-calendar"
                                    initial={{ opacity: 0, y: 10 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    transition={{ delay: index * 0.05 }}
                                >
                                    <div className="conflict-item-header">
                                        <FaExclamationTriangle style={{ marginRight: '8px', color: 'var(--warning-color)' }} />
                                        <strong>Пересечение: {formatDate(conflict.overlapStartDate)} - {formatDate(conflict.overlapEndDate)}</strong>
                                    </div>
                                    <div className="conflict-item-users">
                                        <div className="conflict-user-detail">
                                            <FaUser style={{ marginRight: '5px', color: 'var(--text-secondary)' }} />
                                            <span>{conflict.original_user_full_name ?? `ID: ${conflict.original_user_id}`}</span>
                                            <small style={{ marginLeft: '5px', color: 'var(--text-secondary)' }}>(Заявка #{conflict.original_request_id})</small>
                                        </div>
                                        <div className="conflict-separator"> | </div>
                                        <div className="conflict-user-detail">
                                            <FaUser style={{ marginRight: '5px', color: 'var(--text-secondary)' }} />
                                            <span>{conflict.conflicting_user_full_name ?? `ID: ${conflict.conflicting_user_id}`}</span>
                                            <small style={{ marginLeft: '5px', color: 'var(--text-secondary)' }}>(Заявка #{conflict.conflicting_request_id})</small>
                                        </div>
                                    </div>
                                </motion.div>
                            ))}
                        </div>
                    )
                )}
             </motion.div>
         </>
      )}
      {/* Стили встроены для простоты, но лучше вынести в CSS */}
      <style jsx>{`
        .conflicts-list-container {
          margin-top: 15px;
          display: flex;
          flex-direction: column;
          gap: 12px;
        }
        .conflict-item-calendar {
          background-color: var(--bg-secondary);
          border: 1px solid var(--border-color);
          border-left: 4px solid var(--warning-color);
          border-radius: var(--border-radius);
          padding: 12px 15px;
          font-size: 0.9rem;
        }
        .conflict-item-header {
          display: flex;
          align-items: center;
          margin-bottom: 8px;
          font-weight: 500;
        }
        .conflict-item-users {
          display: flex;
          flex-direction: column;
          gap: 5px;
          padding-left: 22px;
        }
        .conflict-user-detail {
          display: flex;
          align-items: center;
        }
        .conflict-separator {
          text-align: center;
          color: var(--text-secondary);
          font-weight: bold;
          margin: 2px 0;
          display: none;
        }

        @media (min-width: 600px) {
          .conflict-item-users {
            flex-direction: row;
            align-items: center;
            justify-content: space-between;
            gap: 10px;
          }
          .conflict-separator {
            display: inline-block;
          }
          .conflict-user-detail {
             flex-basis: 45%;
             justify-content: flex-start;
          }
        }

        .markers-container {
          display: flex;
          justify-content: center;
          align-items: center;
          gap: 3px;
          position: absolute;
          bottom: 2px;
          left: 0;
          right: 0;
          height: 10px;
        }
        .calendar-marker {
          display: inline-flex;
          align-items: center;
          justify-content: center;
          border-radius: 50%;
          font-size: 0.6rem;
          color: white;
          line-height: 1;
          padding: 0;
        }
        .conflict-marker {
          width: 10px;
          height: 10px;
          background-color: transparent;
          color: var(--danger-color);
          font-size: 0.8rem;
          border-radius: 0;
        }
         .user-dot-marker {
          width: 8px;
          height: 8px;
          border: 1px solid rgba(0,0,0,0.1);
        }
        .more-users-marker {
           width: auto;
           min-width: 14px;
           height: 14px;
           padding: 0 3px;
           border-radius: 3px;
           background-color: var(--text-secondary);
           font-weight: bold;
           color: var(--bg-primary);
           font-size: 0.7rem;
           margin-left: 1px;
        }

        .legend-color-box {
          display: inline-block;
          width: 14px;
          height: 14px;
          margin-right: 5px;
          vertical-align: middle;
          border: 1px solid var(--border-color);
          border-radius: 3px;
        }
        .legend-item {
          display: flex;
          align-items: center;
          margin-bottom: 5px;
        }
      `}</style>
    </motion.div>
  );
};

export default VacationCalendar;
