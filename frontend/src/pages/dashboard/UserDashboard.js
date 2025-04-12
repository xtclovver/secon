import React from 'react';
import { motion } from 'framer-motion';
import { useUser } from '../../context/UserContext'; // Получаем данные пользователя
import { Link } from 'react-router-dom'; // Для ссылок
import { FaPlusCircle, FaList, FaCalendarAlt } from 'react-icons/fa';
// import './UserDashboard.css'; // Можно добавить стили при необходимости

const UserDashboard = () => {
  const { user } = useUser(); // Получаем информацию о текущем пользователе

  return (
    <motion.div
      className="user-dashboard card" // Используем класс card для общего стиля
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <h2>Добро пожаловать, {user?.fullName || 'Пользователь'}!</h2>
      <p>Это ваша панель управления системой учета отпусков.</p>
      
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

      {/* Здесь можно добавить другую информацию, например: */}
      {/* - Оставшиеся дни отпуска */}
      {/* - Статус последних заявок */}
      {/* - Уведомления */}
      
    </motion.div>
  );
};

export default UserDashboard;
