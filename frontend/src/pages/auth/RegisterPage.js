import React, { useState, useEffect } from 'react'; // Добавляем useEffect
import { useNavigate, Link } from 'react-router-dom';
import { register, getPositions } from '../../api/auth'; // Импортируем getPositions
import './RegisterPage.css'; // Импортируем новые стили

function RegisterPage() {
  const [fullName, setFullName] = useState(''); // Переименовано username в fullName
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState(''); // Добавлено состояние для подтверждения пароля
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
    // Убираем логи из useEffect
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
        console.error('Error fetching positions:', err); // Оставляем лог ошибки
      } finally {
        setPositionsLoading(false);
      }
    };

    fetchPositions();
  }, []); // Пустой массив зависимостей для выполнения один раз

  const handleSubmit = async (event) => {
    console.log('[RegisterPage] handleSubmit triggered'); // <-- Оставляем только первый лог
    event.preventDefault(); // Возвращаем стандартный preventDefault
    setError('');
    setSuccess('');

    // Проверка совпадения паролей
    if (password !== confirmPassword) {
      setError('Пароли не совпадают.');
      return; // Прерываем выполнение, если пароли не совпадают
    }

    setLoading(true);

    try {
      // Передаем поля в PascalCase, включая Login (из состояния email)
      const newUser = { 
        Login: email,                       // Добавлено поле Login со значением из email state
        FullName: fullName,                 // PascalCase, как в ошибке
        Email: email,                       // PascalCase, как в ошибке (оставляем, т.к. бэкэнд его тоже ожидает)
        Password: password,                 // PascalCase, по аналогии
        ConfirmPassword: confirmPassword,   // PascalCase, как в ошибке
        PositionID: parseInt(positionId, 10) // PascalCase, по аналогии
      }; 
      
      // Проверяем, что должность выбрана
      if (!positionId) {
        throw new Error('Пожалуйста, выберите должность.');
      }
      console.log('[RegisterPage] FINAL newUser object being sent:', newUser); // <-- Log the object just before sending
      await register(newUser); 
      setSuccess('Регистрация прошла успешно! Теперь вы можете войти.');
      // Очистить поля формы после успешной регистрации (опционально)
      setFullName(''); // Очищаем fullName
      setEmail('');
      setPassword('');
      setConfirmPassword(''); // Очищаем confirmPassword
      // Можно добавить небольшую задержку перед перенаправлением на страницу входа
      setTimeout(() => {
        navigate('/login');
      }, 2000); // 2 секунды задержки
    } catch (err) {
      console.error('[RegisterPage] Registration error:', err); // Оставляем лог ошибки
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
          {/* Возвращаем onSubmit на форму */}
          <form onSubmit={handleSubmit}> 
            {error && <p className="error-message">{error}</p>}
            {success && <p className="success-message">{success}</p>}
            <div className="form-group">
              <label htmlFor="register-fullname">ФИО</label> 
              <input
                type="text"
                id="register-fullname"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                required
                disabled={loading}
              />
            </div>
            <div className="form-group">
              <label htmlFor="register-email">Логин</label>
              <input
                type="text" // Changed from email to text
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
            <div className="form-group">
              <label htmlFor="register-confirm-password">Повторите пароль</label>
              <input
                type="password"
                id="register-confirm-password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
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
            {/* Возвращаем type="submit", onClick убран, возвращаем disabled */}
            <button 
              type="submit" // Возвращаем "submit"
              disabled={loading || positionsLoading} // Возвращаем disabled
            >
              {loading ? 'Регистрация...' : 'Зарегистрироваться'} {/* Убираем "(Debug)" */}
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
