import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { useUser } from '../../context/UserContext';
import { FaUsersCog, FaBuilding, FaUserShield, FaCalendarCheck, FaSave } from 'react-icons/fa';
import { toast } from 'react-toastify';
import { setVacationLimit } from '../../api/vacations'; // API для установки лимита
import { getUsersWithLimits } from '../../api/users'; // API для получения пользователей с лимитами
// import './AdminDashboard.css'; // Можно добавить стили при необходимости

const AdminDashboard = () => {
  const { user } = useUser();
  const [users, setUsers] = useState([]); // Состояние для списка пользователей (UserWithLimitDTO)
  const [selectedYear, setSelectedYear] = useState(new Date().getFullYear() + 1); // Выбранный год
  const [limits, setLimits] = useState({}); // { userId: limitValue (string) } - Лимиты для инпутов
  const [loadingUsers, setLoadingUsers] = useState(false); // Флаг загрузки пользователей
  const [savingLimit, setSavingLimit] = useState({}); // { userId: boolean } - Флаг сохранения лимита для конкретного пользователя

  // Загрузка пользователей с лимитами при монтировании и при смене года
  useEffect(() => {
    const fetchUsersAndLimits = async () => {
      setLoadingUsers(true);
      setUsers([]); // Очищаем список перед загрузкой
      setLimits({}); // Очищаем лимиты
      try {
        const fetchedUsersWithLimits = await getUsersWithLimits(selectedYear);
        setUsers(fetchedUsersWithLimits);

        // Инициализируем состояние limits на основе полученных данных
        const initialLimits = {};
        fetchedUsersWithLimits.forEach(u => {
          // Если лимит есть (не null), устанавливаем его как строку, иначе пустая строка
          // Используем u.vacation_limit_days, как определено в DTO на бэкенде
          initialLimits[u.id] = u.vacation_limit_days !== null ? String(u.vacation_limit_days) : '';
        });
        setLimits(initialLimits);

      } catch (error) {
        toast.error(`Ошибка загрузки пользователей: ${error.message}`);
        console.error("Ошибка загрузки пользователей:", error);
      } finally {
        setLoadingUsers(false);
      }
    };

    fetchUsersAndLimits();
  }, [selectedYear]); // Добавляем selectedYear в зависимости для перезагрузки при смене года

  // Обработчик изменения значения лимита в input
  const handleLimitChange = (userId, value) => {
    setLimits(prevLimits => ({
      ...prevLimits,
      [userId]: value // Сохраняем как строку, преобразуем при сохранении
    }));
  };

  // Обработчик сохранения лимита для пользователя
  const handleSaveLimit = async (userId) => {
    const limitValue = limits[userId]; // Берем значение из состояния limits
    if (limitValue === undefined || limitValue === null || limitValue === '') {
      toast.warn('Введите значение лимита.');
      return;
    }

    const totalDays = parseInt(limitValue, 10);
    if (isNaN(totalDays) || totalDays < 0) {
      toast.error('Лимит должен быть неотрицательным числом.');
      return;
    }

    setSavingLimit(prev => ({ ...prev, [userId]: true }));
    try {
      // Вызываем API для установки лимита
      await setVacationLimit(userId, selectedYear, totalDays);
      toast.success(`Лимит для пользователя ${users.find(u => u.id === userId)?.fullName || userId} на ${selectedYear} год успешно установлен.`);
      // Обновляем состояние limits локально после успешного сохранения (не обязательно, т.к. данные перезагрузятся при смене года)
      // Но можно сделать для мгновенного отображения, если нужно
      // setLimits(prev => ({ ...prev, [userId]: String(totalDays) }));
    } catch (error) {
      toast.error(`Ошибка установки лимита: ${error.message}`);
      console.error("Ошибка установки лимита:", error);
    } finally {
      setSavingLimit(prev => ({ ...prev, [userId]: false }));
    }
  };

  return (
    <motion.div
      className="admin-dashboard card"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <h2><FaUserShield /> Панель администратора</h2>
      <p>Добро пожаловать, {user?.fullName || 'Администратор'}!</p>
      <p>Здесь вы можете управлять пользователями, подразделениями и настройками системы.</p>

      {/* Раздел управления (примеры) */}
      <div className="admin-actions" style={{ marginTop: '30px', display: 'grid', gap: '15px', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))' }}>
         <motion.div className="action-card card" whileHover={{ scale: 1.03 }}>
           <FaUsersCog size={30} style={{ marginBottom: '10px', color: 'var(--accent-color)' }} />
           <h3>Управление пользователями</h3>
           <p>Добавление, редактирование и удаление пользователей.</p>
           <button className="btn btn-secondary" disabled>Перейти (в разработке)</button>
         </motion.div>

         <motion.div className="action-card card" whileHover={{ scale: 1.03 }}>
           <FaBuilding size={30} style={{ marginBottom: '10px', color: 'var(--success-color)' }} />
           <h3>Управление подразделениями</h3>
           <p>Создание и настройка структуры подразделений.</p>
            <button className="btn btn-secondary" disabled>Перейти (в разработке)</button>
         </motion.div>
      </div>

      {/* Раздел управления лимитами отпусков */}
      <motion.div
        className="vacation-limit-management card"
        style={{ marginTop: '30px' }}
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.3 }}
      >
        <h3><FaCalendarCheck /> Управление лимитами отпусков</h3>

        <div className="year-selector" style={{ marginBottom: '20px' }}>
          <label htmlFor="limit-year">Год:</label>
          <select
            id="limit-year"
            value={selectedYear}
            onChange={(e) => setSelectedYear(parseInt(e.target.value))}
            style={{ marginLeft: '10px' }}
            disabled={loadingUsers} // Блокируем выбор года во время загрузки
          >
            {/* Генерируем несколько лет вокруг текущего */}
            {[...Array(5)].map((_, i) => {
              const yearOption = new Date().getFullYear() - 2 + i;
              return <option key={yearOption} value={yearOption}>{yearOption}</option>;
            })}
            <option value={new Date().getFullYear() + 1}>Следующий год ({new Date().getFullYear() + 1})</option>
          </select>
        </div>

        {loadingUsers ? (
          <p>Загрузка пользователей...</p>
        ) : users.length === 0 ? (
           <p>Пользователи не найдены.</p>
        ) : (
          <div style={{ overflowX: 'auto' }}> {/* Добавляем обертку для скролла на маленьких экранах */}
            <table className="user-limits-table" style={{ width: '100%', borderCollapse: 'collapse', minWidth: '600px' }}>
              <thead>
                <tr>
                  <th style={tableHeaderStyle}>Пользователь</th>
                  <th style={tableHeaderStyle}>Email</th>
                  <th style={tableHeaderStyle}>Лимит дней ({selectedYear})</th>
                  <th style={tableHeaderStyle}>Действие</th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id}>
                    <td style={tableCellStyle}>{u.fullName}</td>
                    <td style={tableCellStyle}>{u.email}</td>
                    <td style={tableCellStyle} title={limits[u.id] === '' ? 'Лимит не установлен' : ''}>
                      <input
                        type="number"
                        min="0"
                        value={limits[u.id] || ''} // Используем значение из состояния limits
                        onChange={(e) => handleLimitChange(u.id, e.target.value)}
                        style={{ width: '80px', padding: '5px' }}
                        disabled={savingLimit[u.id]} // Блокируем во время сохранения
                        placeholder="Нет" // Показываем плейсхолдер, если лимит не задан
                      />
                    </td>
                    <td style={tableCellStyle}>
                      <button
                        className="btn btn-primary btn-sm"
                        onClick={() => handleSaveLimit(u.id)}
                        disabled={savingLimit[u.id] || limits[u.id] === undefined || limits[u.id] === '' || parseInt(limits[u.id]) < 0} // Блокируем, если сохраняется, не введено или отрицательное
                        style={{ display: 'flex', alignItems: 'center', gap: '5px', minWidth: '110px' }} // Добавляем minWidth
                      >
                        <FaSave /> {savingLimit[u.id] ? 'Сохранение...' : 'Сохранить'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {/* TODO: Добавить пагинацию или поиск, если пользователей много */}
        {/* TODO: Добавить возможность массового обновления */}
      </motion.div>

    </motion.div>
  );
};

// Стили для таблицы (можно вынести в CSS)
const tableHeaderStyle = {
  borderBottom: '2px solid var(--border-color)',
  padding: '10px 5px',
  textAlign: 'left',
  fontWeight: 'bold',
  whiteSpace: 'nowrap', // Предотвращаем перенос заголовков
};

const tableCellStyle = {
  borderBottom: '1px solid var(--border-color-light)',
  padding: '10px 5px',
  verticalAlign: 'middle',
};

export default AdminDashboard;
