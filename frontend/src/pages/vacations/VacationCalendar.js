import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import Calendar from 'react-calendar';
import { toast } from 'react-toastify';
import { FaCalendarAlt, FaUser, FaUsers } from 'react-icons/fa';
import { getDepartmentVacations } from '../../api/vacations'; // Используем API для получения данных
import Loader from '../../components/ui/Loader/Loader';
import { useUser } from '../../context/UserContext'; // Для получения ID подразделения
import 'react-calendar/dist/Calendar.css';
// import './VacationCalendar.css'; // Можно добавить стили

const VacationCalendar = () => {
  const { user } = useUser(); // Получаем текущего пользователя
  const [year, setYear] = useState(new Date().getFullYear());
  const [vacations, setVacations] = useState([]); // Данные об отпусках
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [calendarDate, setCalendarDate] = useState(new Date()); // Для управления выбранной датой/месяцем

  // Загрузка данных об отпусках
  useEffect(() => {
    const fetchCalendarData = async () => {
      // Определяем ID подразделения (для руководителя - его подразделение, для обычного - тоже его)
      // В реальном приложении логика может быть сложнее (например, админ видит все)
      const departmentId = user?.departmentId || 1; // Заглушка ID=1, если у пользователя нет departmentId

      if (!departmentId) {
          setError("Не удалось определить подразделение для загрузки календаря.");
          return;
      }

      setLoading(true);
      setError(null);
      try {
        const data = await getDepartmentVacations(departmentId, year); // Используем реальный API вызов
        setVacations(data);
      } catch (err) {
        setError(err.message || 'Не удалось загрузить данные календаря.');
        toast.error(err.message || 'Не удалось загрузить данные календаря.');
      } finally {
        setLoading(false);
      }
    };

    if (user) { // Загружаем только если есть данные пользователя
        fetchCalendarData();
    }

  }, [year, user]); // Перезагружаем при смене года или пользователя

   // Функция для отображения маркеров в календаре
  const getTileContent = ({ date, view }) => {
    if (view !== 'month') return null;
    
    const dateString = date.toISOString().split('T')[0]; 

    // Находим сотрудников в отпуске на эту дату
    const usersOnVacation = vacations.filter(vacation => 
        vacation.periods.some(period => dateString >= period.startDate && dateString <= period.endDate)
    );

    if (usersOnVacation.length > 0) {
      // Показываем иконку и количество сотрудников
      return (
        <div className="calendar-marker user-marker" title={usersOnVacation.map(v => v.userName).join(', ')}>
          <FaUsers /> 
          {/* <span>{usersOnVacation.length}</span> */}
        </div>
      );
    }
    
    return null;
  };

  // Функция для определения класса даты в календаре
  const getTileClassName = ({ date, view }) => {
    if (view !== 'month') return '';
    
    const dateString = date.toISOString().split('T')[0]; 

    const isOnVacation = vacations.some(vacation => 
        vacation.periods.some(period => dateString >= period.startDate && dateString <= period.endDate)
    );
    
    if (isOnVacation) {
      return 'department-vacation-date'; // Класс для дней с отпусками
    }
    
    return '';
  };


  return (
    <motion.div
      className="vacation-calendar-container card"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
    >
      <h2><FaCalendarAlt /> Календарь отпусков {user?.departmentId ? `подразделения #${user.departmentId}` : ''}</h2>

       <div className="controls" style={{ marginBottom: '20px', display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '10px' }}>
         <label htmlFor="calendar-year">Год:</label>
         <select 
            id="calendar-year" 
            value={year} 
            onChange={(e) => setYear(parseInt(e.target.value))}
            disabled={loading}
          >
            {[...Array(5)].map((_, i) => {
              const y = new Date().getFullYear() + 2 - i;
              return <option key={y} value={y}>{y}</option>;
            })}
          </select>
       </div>

      {loading && <Loader text="Загрузка календаря..." />}
      {error && <div className="error-message">{error}</div>}

      {!loading && !error && (
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
                className="department-wide-calendar" // Класс для возможных стилей
            />
             <div className="calendar-legend" style={{ marginTop: '15px', textAlign: 'center', fontSize: '0.9rem' }}>
                <span style={{ display: 'inline-flex', alignItems: 'center', marginRight: '15px' }}>
                    <div style={{ width: '15px', height: '15px', backgroundColor: 'rgba(13, 202, 240, 0.3)', marginRight: '5px', borderRadius: '3px' }}></div>
                    Сотрудник в отпуске
                </span>
             </div>
         </motion.div>
      )}
      
       {/* Стили для маркеров и дат */}
       <style jsx>{`
        .calendar-marker { position: absolute; bottom: 4px; left: 50%; transform: translateX(-50%); display: flex; align-items: center; gap: 3px; font-size: 0.7rem; padding: 1px 3px; border-radius: 3px; color: white; background-color: var(--info-color); }
        .department-vacation-date { background-color: rgba(13, 202, 240, 0.15) !important; /* Легкий фон для дней с отпусками */ }
        .react-calendar__tile--now.department-vacation-date { background-color: rgba(13, 202, 240, 0.3) !important; /* Выделение сегодняшнего дня в отпуске */ }
        .department-wide-calendar { width: 100%; border: none !important; }
        .department-wide-calendar .react-calendar__tile { height: 90px; /* Увеличим высоту ячейки */ } 
      `}</style>

    </motion.div>
  );
};

export default VacationCalendar;
