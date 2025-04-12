import React from 'react';
import { motion } from 'framer-motion';
import { useUser } from '../../context/UserContext';
import { Link } from 'react-router-dom';
import { FaPlusCircle, FaList, FaCalendarAlt, FaCalendarCheck } from 'react-icons/fa'; // Добавили FaCalendarCheck
// import './UserDashboard.css'; // Можно добавить стили при необходимости

const UserDashboard = () => {
  // Получаем user из контекста, теперь он должен содержать лимиты
  const { user } = useUser();
  const currentYear = new Date().getFullYear();

  // ЛОГИРОВАНИЕ: Проверяем объект пользователя и его лимиты при рендере
  console.log("UserDashboard rendering. User object from context:", user);
  console.log("Current available days from context:", user?.currentAvailableDays);

  // Получаем данные о лимитах за текущий год из объекта пользователя
  const limits = user?.vacationLimits?.[currentYear];
  const availableDays = user?.currentAvailableDays; // Используем поле верхнего уровня
  const totalDays = user?.currentTotalDays;
  const usedDays = user?.currentUsedDays;

  return (
    <motion.div
      className="user-dashboard card" // Используем класс card для общего стиля
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <h2>Добро пожаловать, {user?.fullName || 'Пользователь'}!</h2>
      <p>Это ваша панель управления системой учета отпусков.</p>

      {/* Отображение лимитов отпуска */}
      <div className="vacation-limits-info card" style={{ marginTop: '20px', padding: '15px', backgroundColor: 'var(--bg-secondary)' }}>
        <h3><FaCalendarCheck /> Ваш лимит отпуска на {currentYear} год</h3>
        {limits ? (
          <div style={{ display: 'flex', justifyContent: 'space-around', flexWrap: 'wrap', gap: '10px' }}>
            <p><strong>Всего дней:</strong> {totalDays ?? 'N/A'}</p>
            <p><strong>Использовано:</strong> {usedDays ?? 'N/A'}</p>
            <p><strong>Доступно:</strong> <strong style={{ color: 'var(--success-color)', fontSize: '1.1em' }}>{availableDays ?? 'N/A'}</strong></p>
          </div>
        ) : (
          <p>Информация о лимите загружается или недоступна...</p>
        )}
      </div>

      {/* Секция с быстрыми действиями */}
      <div className="quick-actions" style={{ marginTop: '30px', display: 'flex', gap: '15px', flexWrap: 'wrap' }}>
        <Link to="/vacations/new" className="btn btn-primary">
          <FaPlusCircle /> Оформить отпуск
        </Link>
        <Link to="/vacations/list" className="btn btn-secondary">
          <FaList /> Мои заявки
        </Link>
        <Link to="/vacations/calendar" className="btn btn-secondary">
          <FaCalendarAlt /> Календарь отпусков
        </Link>
      </div>

    </motion.div>
  );
};

export default UserDashboard;
