import React from 'react';
import { Link } from 'react-router-dom';
import { useUser } from '../../context/UserContext';
// Добавляем FaHome в импорт
import { FaUserShield, FaUserTie, FaUser, FaPlusCircle, FaList, FaCalendarAlt, FaHome } from 'react-icons/fa';
// import './UniversalDashboard.css'; // Можно будет создать CSS позже

const UniversalDashboard = () => {
  const { user } = useUser();

  if (!user) {
    // Можно показать лоадер или сообщение об ошибке, если пользователь еще не загружен
    return <div>Загрузка данных пользователя...</div>;
  }

  // Определяем роль пользователя
  const isAdmin = user.role === 'admin'; // Или используйте user.isAdmin если оно есть
  const isManager = user.role === 'manager'; // Или используйте user.isManager
  const isRegularUser = !isAdmin && !isManager;

  return (
    <div className="dashboard-container"> {/* Добавлен класс для стилизации */}
      <h2>
        <FaHome style={{ marginRight: '10px' }} />
        Дашборд {user.fullName || 'Пользователя'}
      </h2>
      <p>Добро пожаловать в систему учета отпусков!</p>

      <div className="dashboard-sections"> {/* Контейнер для секций */}

        {/* Секция для всех пользователей */}
        <section className="dashboard-section common-actions">
          <h3>Основные действия</h3>
          <ul>
            <li>
              <Link to="/profile">
                <FaUser style={{ marginRight: '8px' }} /> Мой профиль
              </Link>
            </li>
            <li>
              <Link to="/vacations/new">
                <FaPlusCircle style={{ marginRight: '8px' }} /> Оформить новый отпуск
              </Link>
            </li>
            <li>
              <Link to="/vacations/list">
                <FaList style={{ marginRight: '8px' }} /> Мои заявки на отпуск
              </Link>
            </li>
            <li>
              <Link to="/vacations/calendar">
                <FaCalendarAlt style={{ marginRight: '8px' }} /> Календарь отпусков
              </Link>
            </li>
          </ul>
        </section>

        {/* Секция для руководителей */}
        {isManager && (
          <section className="dashboard-section manager-actions">
            <h3><FaUserTie style={{ marginRight: '8px' }} /> Инструменты руководителя</h3>
            <ul>
              <li>
                <Link to="/manager/dashboard">Перейти в Дашборд руководителя</Link>
                {/* Сюда можно добавить другие ссылки, специфичные для менеджера */}
              </li>
              {/* <li><Link to="/manager/team-requests">Заявки команды</Link></li> */}
            </ul>
          </section>
        )}

        {/* Секция для администраторов */}
        {isAdmin && (
          <section className="dashboard-section admin-actions">
            <h3><FaUserShield style={{ marginRight: '8px' }} /> Инструменты администратора</h3>
            <ul>
              <li>
                <Link to="/admin/dashboard">Перейти в Админ-панель</Link>
                {/* Сюда можно добавить другие ссылки, специфичные для администратора */}
              </li>
              {/* <li><Link to="/admin/users">Управление пользователями</Link></li> */}
              {/* <li><Link to="/admin/settings">Настройки системы</Link></li> */}
            </ul>
          </section>
        )}
      </div>

      {/* Можно добавить стили прямо здесь или в отдельном CSS файле */}
      <style jsx>{`
        .dashboard-container {
          padding: 20px;
        }
        .dashboard-sections {
          margin-top: 20px;
          display: grid;
          gap: 20px;
          grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); /* Адаптивные колонки */
        }
        .dashboard-section {
          border: 1px solid var(--border-color, #ccc); /* Используем переменную или дефолтный цвет */
          border-radius: 8px;
          padding: 15px;
          background-color: var(--background-secondary, #f9f9f9); /* Используем переменную или дефолтный цвет */
        }
        .dashboard-section h3 {
          margin-top: 0;
          margin-bottom: 15px;
          display: flex;
          align-items: center;
          color: var(--text-color-primary, #333);
        }
        .dashboard-section ul {
          list-style: none;
          padding: 0;
          margin: 0;
        }
        .dashboard-section li {
          margin-bottom: 10px;
        }
        .dashboard-section a {
          text-decoration: none;
          color: var(--primary-color, #007bff); /* Используем переменную или дефолтный цвет */
          display: flex;
          align-items: center;
          transition: color 0.2s ease;
        }
        .dashboard-section a:hover {
          color: var(--primary-color-dark, #0056b3); /* Используем переменную или дефолтный цвет */
        }
        .manager-actions {
            border-left: 5px solid orange; /* Пример выделения секции */
        }
        .admin-actions {
            border-left: 5px solid red; /* Пример выделения секции */
        }
      `}</style>
    </div>
  );
};

export default UniversalDashboard;
