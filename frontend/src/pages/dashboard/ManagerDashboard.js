import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { toast } from 'react-toastify';
import Calendar from 'react-calendar';
import { FaExclamationTriangle, FaCheck, FaUser, FaCalendarAlt } from 'react-icons/fa';
import { getUnitVacations, getVacationIntersections } from '../../api/vacations'; // <<< Исправлен импорт
import 'react-calendar/dist/Calendar.css';
import './ManagerDashboard.css'; // Предполагается, что CSS файл будет создан

const ManagerDashboard = () => {
  const [year, setYear] = useState(new Date().getFullYear() + 1);
  const [unitId, setUnitId] = useState(1); // <<< Переименована переменная состояния
  const [vacations, setVacations] = useState([]);
  const [intersections, setIntersections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('calendar'); // 'calendar' или 'intersections'
  const [calendarDate, setCalendarDate] = useState(new Date());


  // Загрузка данных при изменении года или подразделения
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // Получение отпусков сотрудников подразделения
        const vacationsData = await getUnitVacations(unitId, year); // <<< Исправлен вызов и параметр
        setVacations(vacationsData);
        
        // Получение пересечений отпусков
        const intersectionsData = await getVacationIntersections(unitId, year); // <<< Исправлен параметр
        setIntersections(intersectionsData);
      } catch (error) {
        toast.error('Ошибка при загрузке данных');
        console.error(error);
      } finally {
        setLoading(false);
      }
    };
    
    // Получение ID подразделения руководителя (в реальном приложении из user context)
    // const currentUser = JSON.parse(localStorage.getItem('user'));
    // if (currentUser && currentUser.departmentId) {
    //   setDepartmentId(currentUser.departmentId);
    // } else {
    //   toast.error('Не удалось определить подразделение руководителя');
    // }

    fetchData();
  }, [year, unitId]); // <<< Исправлена зависимость useEffect

  // Функция для отображения отпусков в календаре
  const getTileContent = ({ date, view }) => {
    if (view !== 'month') return null;
    
    const dateString = date.toISOString().split('T')[0]; // Формат YYYY-MM-DD

    // Находим все отпуска на эту дату
    const vacationsOnDate = vacations.filter(vacation => {
      return vacation.periods.some(period => {
        return dateString >= period.startDate && dateString <= period.endDate;
      });
    });
    
    // Находим пересечения на эту дату
    const intersectionsOnDate = intersections.filter(intersection => {
       return dateString >= intersection.startDate && dateString <= intersection.endDate;
    });
    
    if (intersectionsOnDate.length > 0) {
      return (
        <div className="calendar-marker intersection-marker">
          <FaExclamationTriangle />
          {/* <span>{intersectionsOnDate.length * 2}</span> Можно показывать кол-во сотрудников */}
        </div>
      );
    }
    
    if (vacationsOnDate.length > 0) {
      return (
        <div className="calendar-marker vacation-marker">
          <FaUser />
          {/* <span>{vacationsOnDate.length}</span> */}
        </div>
      );
    }
    
    return null;
  };

  // Функция для определения класса даты в календаре
  const getTileClassName = ({ date, view }) => {
    if (view !== 'month') return '';
    
    const dateString = date.toISOString().split('T')[0]; // Формат YYYY-MM-DD

    // Проверяем, есть ли пересечения на эту дату
    const hasIntersections = intersections.some(intersection => {
      return dateString >= intersection.startDate && dateString <= intersection.endDate;
    });
    
    if (hasIntersections) {
      return 'intersection-date';
    }
    
    // Проверяем, есть ли отпуска на эту дату
    const hasVacations = vacations.some(vacation => {
      return vacation.periods.some(period => {
        return dateString >= period.startDate && dateString <= period.endDate;
      });
    });
    
    if (hasVacations) {
      return 'vacation-date';
    }
    
    return '';
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
        <div className="year-selector">
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
        </div>
        
        <div className="tab-selector">
          <button 
            className={`tab-button btn ${activeTab === 'calendar' ? 'active' : ''}`}
            onClick={() => setActiveTab('calendar')}
          >
            <FaCalendarAlt /> Календарь
          </button>
          <button 
            className={`tab-button btn ${activeTab === 'intersections' ? 'active' : ''}`}
            onClick={() => setActiveTab('intersections')}
          >
            <FaExclamationTriangle /> Пересечения ({intersections.length})
          </button>
        </div>
      </div>
      
      {loading ? (
        <div className="loading-spinner">Загрузка данных...</div>
      ) : (
        <>
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
                <div className="legend-item">
                  <div className="legend-color vacation"></div>
                  <span>Сотрудники в отпуске</span>
                </div>
                <div className="legend-item">
                  <div className="legend-color intersection"></div>
                  <span>Пересечения отпусков</span>
                </div>
              </div>
            </motion.div>
          )}
          
          {activeTab === 'intersections' && (
            <motion.div 
              className="intersections-container card"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
            >
              <h3>Пересечения отпусков</h3>
              
              {intersections.length === 0 ? (
                <div className="no-intersections">
                  <FaCheck />
                  <p>Пересечений отпусков не найдено</p>
                </div>
              ) : (
                <div className="intersections-list">
                  {intersections.map((intersection, index) => (
                    <motion.div 
                      key={index}
                      className="intersection-item"
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: 0.1 * index }}
                    >
                      <div className="intersection-header">
                        <FaExclamationTriangle className="intersection-icon" />
                        <span>Пересечение #{index + 1}</span>
                      </div>
                      <div className="intersection-details">
                        <div className="intersection-users">
                          <div className="intersection-user">
                            <FaUser />
                            <span>{intersection.userName1} (ID: {intersection.userId1})</span>
                          </div>
                          <div className="intersection-user">
                            <FaUser />
                            <span>{intersection.userName2} (ID: {intersection.userId2})</span>
                          </div>
                        </div>
                        <div className="intersection-dates">
                          <div>
                            <strong>Период пересечения:</strong> 
                            {new Date(intersection.startDate).toLocaleDateString('ru-RU')} - 
                            {new Date(intersection.endDate).toLocaleDateString('ru-RU')}
                          </div>
                          <div>
                            <strong>Количество дней:</strong> {intersection.daysCount}
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
