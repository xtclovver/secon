import React, { useState, useEffect } from 'react'; // Добавляем useEffect
import { useNavigate, Link } from 'react-router-dom';
import { register, getPositions } from '../../api/auth'; // Импортируем getPositions
import './RegisterPage.css'; // Импортируем новые стили

function RegisterPage() {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  // Убираем state для role, добавляем для positionId и positionsData
  const [positionId, setPositionId] = useState('');
  const [positionsData, setPositionsData] = useState([]); // Для хранения списка должностей
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const [positionsLoading, setPositionsLoading] = useState(true); // Состояние загрузки должностей
  const navigate = useNavigate();

  // Загружаем должности при монтировании компонента
  useEffect(() => {
    const fetchPositions = async () => {
      try {
        setPositionsLoading(true);
        const data = await getPositions();
        setPositionsData(data || []); // Устанавливаем данные или пустой массив
         // Устанавливаем ID первой должности по умолчанию, если список не пуст
        if (data && data.length > 0 && data[0].positions && data[0].positions.length > 0) {
          setPositionId(data[0].positions[0].id); 
        }
      } catch (err) {
        setError('Не удалось загрузить список должностей.');
        console.error('Error fetching positions:', err);
      } finally {
        setPositionsLoading(false);
      }
    };

    fetchPositions();
  }, []); // Пустой массив зависимостей для выполнения один раз

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    try {
       // Передаем position_id вместо role
      const newUser = { username, email, password, position_id: parseInt(positionId, 10) }; // Убедимся, что ID - число
      if (!positionId) {
          throw new Error('Пожалуйста, выберите должность.'); // Добавим проверку
      }
      await register(newUser);
      setSuccess('Регистрация прошла успешно! Теперь вы можете войти.');
      // Очистить поля формы после успешной регистрации (опционально)
      // setUsername('');
      // setEmail('');
      // setPassword('');
      // setRole('employee');
      // Можно добавить небольшую задержку перед перенаправлением на страницу входа
      setTimeout(() => {
        navigate('/login');
      }, 2000); // 2 секунды задержки
    } catch (err) {
      // console.error('Registration error:', err); // Лог ошибки
      setError(err.message || 'Ошибка регистрации. Пожалуйста, попробуйте еще раз.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-page register-page"> {/* Добавляем класс register-page */}
      <div className="auth-container register-container-size"> {/* Новый класс для управления размером */}
        <div className="form-container"> {/* Убираем register-container класс отсюда */}
          <h2>Регистрация</h2>
          <form onSubmit={handleSubmit}>
            {error && <p className="error-message">{error}</p>}
            {success && <p className="success-message">{success}</p>}
            <div className="form-group">
              <label htmlFor="register-username">Имя пользователя</label>
              <input
                type="text"
                id="register-username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                disabled={loading}
              />
            </div>
            <div className="form-group">
              <label htmlFor="register-email">Email</label>
              <input
                type="email"
                id="register-email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                disabled={loading}
              />
            </div>
            <div className="form-group">
              <label htmlFor="register-password">Пароль</label>
              <input
                type="password"
                id="register-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={loading}
              />
             </div>
             {/* Заменяем select для role на select для positionId */}
             <div className="form-group">
               <label htmlFor="register-position">Должность</label>
               <select
                 id="register-position"
                 value={positionId}
                 onChange={(e) => setPositionId(e.target.value)}
                 required
                 disabled={loading || positionsLoading} // Блокируем во время загрузки должностей
               >
                 {positionsLoading ? (
                   <option value="" disabled>Загрузка должностей...</option>
                 ) : positionsData.length === 0 ? (
                     <option value="" disabled>Нет доступных должностей</option>
                 ) : (
                   // Генерируем группы и опции из positionsData
                   positionsData.map(group => (
                     <optgroup label={group.name} key={group.id}>
                       {group.positions && group.positions.map(position => (
                         <option value={position.id} key={position.id}>
                           {position.name}
                         </option>
                       ))}
                     </optgroup>
                   ))
                 )}
               </select>
             </div>
            <button type="submit" disabled={loading || positionsLoading}>
              {loading ? 'Регистрация...' : 'Зарегистрироваться'}
            </button>
          </form>
           <Link to="/login" className="toggle-link"> {/* Используем Link */}
             Уже есть аккаунт? Войдите
           </Link>
        </div>
      </div>
    </div>
  );
}

export default RegisterPage;
