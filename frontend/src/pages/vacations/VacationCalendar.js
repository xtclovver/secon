import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import Calendar from 'react-calendar';
import { toast } from 'react-toastify';
import { FaCalendarAlt, FaUsers, FaInfoCircle } from 'react-icons/fa';
// Обновленные импорты API
import { getApprovedVacationsForCalendar, getVacationLimit } from '../../api/vacations';
import Loader from '../../components/ui/Loader/Loader';
import { useUser } from '../../context/UserContext';
import './VacationCalendar.css';

const VacationCalendar = () => {
  const { user } = useUser();
  const [year, setYear] = useState(new Date().getFullYear());
  const [approvedVacations, setApprovedVacations] = useState([]); // Состояние для утвержденных отпусков
  const [vacationLimitInfo, setVacationLimitInfo] = useState(null); // Состояние для лимита отпуска
  const [loadingVacations, setLoadingVacations] = useState(false);
  const [loadingLimit, setLoadingLimit] = useState(false);
  const [errorVacations, setErrorVacations] = useState(null);
  const [errorLimit, setErrorLimit] = useState(null);
  const [calendarDate, setCalendarDate] = useState(new Date());

  // Загрузка утвержденных данных для календаря
  useEffect(() => {
    const fetchCalendarData = async () => {
      if (!user) return; // Не грузим, если нет пользователя

      setLoadingVacations(true);
      setErrorVacations(null);
      try {
        // Получаем утвержденные отпуска для рендеринга в календаре
        // Можно добавить фильтр по unitId, если нужно показывать только свой юнит: { year, unitId: user.organizational_unit_id }
        const filters = { year };
        const data = await getApprovedVacationsForCalendar(filters);
        setApprovedVacations(data || []); // Убедимся, что это массив
      } catch (err) {
        const errorMsg = err.message || 'Не удалось загрузить данные календаря.';
        setErrorVacations(errorMsg);
        toast.error(errorMsg);
      } finally {
        setLoadingVacations(false);
      }
    };

    fetchCalendarData();
  }, [year, user]); // Перезагружаем при смене года или пользователя

  // Загрузка лимита отпуска текущего пользователя
  useEffect(() => {
    const fetchVacationLimit = async () => {
       if (!user) return; // Не грузим, если нет пользователя

        setLoadingLimit(true);
        setErrorLimit(null);
        try {
            const limitData = await getVacationLimit(year);
            setVacationLimitInfo(limitData);
        } catch (err) {
            const errorMsg = err.message || 'Не удалось загрузить лимит отпуска.';
            setErrorLimit(errorMsg);
            // Можно не показывать toast, если лимит не критичен для основной функции
            // toast.error(errorMsg);
             console.error("Ошибка загрузки лимита отпуска:", errorMsg); // Логируем ошибку
             setVacationLimitInfo(null); // Сбрасываем лимит при ошибке
        } finally {
            setLoadingLimit(false);
        }
    };

     fetchVacationLimit();
  }, [year, user]); // Перезагружаем при смене года или пользователя


  // Вспомогательная функция для проверки, находится ли дата в периоде
  const isDateInPeriod = (dateStr, period) => {
    // Убедимся, что period.start_date и period.end_date существуют и корректны
    if (!period || !period.start_date || !period.end_date) {
        // console.warn("Некорректный период отпуска:", period);
        return false;
    }
    // Извлекаем только дату YYYY-MM-DD из ISO строки
    const startDateStr = period.start_date.split('T')[0];
    const endDateStr = period.end_date.split('T')[0];

    return dateStr >= startDateStr && dateStr <= endDateStr;
  };

  // Функция для отображения маркеров в календаре
  const getTileContent = ({ date, view }) => {
    if (view !== 'month' || !Array.isArray(approvedVacations)) return null;

    const dateString = date.toISOString().split('T')[0];

    // Находим сотрудников в отпуске на эту дату
    const usersOnVacation = approvedVacations.filter(vacation =>
        Array.isArray(vacation.periods) && vacation.periods.some(period => isDateInPeriod(dateString, period))
    );

    if (usersOnVacation.length > 0) {
      // Формируем строку с ФИО для всплывающей подсказки
      const names = usersOnVacation.map(v => v.user_full_name || `ID: ${v.user_id}`).join(', '); // Используем user_full_name
      return (
        <div className="calendar-marker user-marker" title={names}>
          <FaUsers />
          {/* Можно добавить счетчик, если нужно: <span>{usersOnVacation.length}</span> */}
        </div>
      );
    }

    return null;
  };

  // Функция для определения класса даты в календаре
  const getTileClassName = ({ date, view }) => {
    if (view !== 'month' || !Array.isArray(approvedVacations)) return '';

    const dateString = date.toISOString().split('T')[0];

    const isOnVacation = approvedVacations.some(vacation =>
        Array.isArray(vacation.periods) && vacation.periods.some(period => isDateInPeriod(dateString, period))
    );

    if (isOnVacation) {
      // Используем тот же класс, что и в легенде
      return 'department-vacation-date';
    }

    return '';
  };

  const isLoading = loadingVacations || loadingLimit;

  return (
    <motion.div
      className="vacation-calendar-container card"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
    >
      <h2><FaCalendarAlt /> Календарь отпусков (Утвержденные)</h2>

       <div className="controls" style={{ marginBottom: '20px', display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '10px' }}>
         <label htmlFor="calendar-year">Год:</label>
         <select
            id="calendar-year"
            value={year}
            onChange={(e) => {
              const newYear = parseInt(e.target.value);
              setYear(newYear);
              setCalendarDate(new Date(newYear, 0, 1));
            }}
            disabled={isLoading} // Блокируем во время любой загрузки
          >
            {[...Array(4)].map((_, i) => {
              const currentYear = new Date().getFullYear();
              const y = currentYear + i;
              return <option key={y} value={y}>{y}</option>;
            })}
          </select>
       </div>

      {isLoading && <Loader text="Загрузка данных..." />}
      {errorVacations && <div className="error-message">{errorVacations}</div>}
      {/* Можно добавить отображение errorLimit, если нужно */}

      {!isLoading && !errorVacations && (
         <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.2 }}
         >
            <Calendar
                onChange={setCalendarDate}
                value={calendarDate}
                tileContent={getTileContent}
                tileClassName={getTileClassName}
                locale="ru-RU"
                className="department-wide-calendar"
            />
             {/* Блок Легенды и Статуса Отпуска */}
             <div className="calendar-info-section">
                <div className="calendar-legend">
                    <h4><FaInfoCircle /> Легенда:</h4>
                    <span>
                        {/* Класс .department-vacation-date должен быть определен в CSS */}
                        <div className="legend-color-box department-vacation-date"></div>
                        Сотрудник в отпуске
                    </span>
                    {/* Сюда можно добавлять другие элементы легенды динамически */}
                </div>

                {/* Отображение статуса отпуска */}
                 {vacationLimitInfo ? (
                    <div className="vacation-status-info">
                        <h4>Ваш статус отпуска ({year}):</h4>
                         <p>
                            Доступно дней:
                            <strong> {vacationLimitInfo.total_days - vacationLimitInfo.used_days} </strong>
                             / {vacationLimitInfo.total_days}
                        </p>
                    </div>
                ) : (
                     !loadingLimit && errorLimit && ( // Показываем ошибку только если не грузится и есть ошибка лимита
                        <div className="vacation-status-info error-message">
                            Не удалось загрузить информацию о лимите отпуска.
                        </div>
                     )
                 )}
             </div>
         </motion.div>
      )}

    </motion.div>
  );
};

export default VacationCalendar;
