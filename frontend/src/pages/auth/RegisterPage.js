import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { register, getPositions } from '../../api/auth';
import { getOrganizationalUnitTree } from '../../api/units'; // Импортируем API для юнитов
import './RegisterPage.css';

function RegisterPage() {
  const [fullName, setFullName] = useState(''); // Переименовано username в fullName
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState(''); // Добавлено состояние для подтверждения пароля
  // Убираем state для role, добавляем для positionId и positionsData
  const [positionId, setPositionId] = useState('');
  const [positionsData, setPositionsData] = useState([]);
  const [unitTreeData, setUnitTreeData] = useState([]); // Для хранения дерева юнитов
  const [selectedDepartmentId, setSelectedDepartmentId] = useState('');
  const [selectedSubDepartmentId, setSelectedSubDepartmentId] = useState('');
  const [selectedSectorId, setSelectedSectorId] = useState('');
  const [subDepartments, setSubDepartments] = useState([]); // Доступные суб-департаменты
  const [sectors, setSectors] = useState([]); // Доступные секторы
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const [positionsLoading, setPositionsLoading] = useState(true);
  const [unitsLoading, setUnitsLoading] = useState(true); // Состояние загрузки юнитов
  const navigate = useNavigate();

  // Загружаем должности и дерево юнитов при монтировании
  useEffect(() => {
    const fetchData = async () => {
      // Загрузка должностей
      try {
        setPositionsLoading(true);
        const posData = await getPositions();
        setPositionsData(posData || []); // posData теперь []models.Position
        // Устанавливаем ID первой должности по умолчанию, если список не пуст
        if (posData && posData.length > 0) {
          setPositionId(posData[0].id);
        }
      } catch (err) {
        // Отображаем более детальную ошибку из API или общую, если message нет
        const errorMsg = err.message || 'Не удалось загрузить должности (неизвестная ошибка).';
        setError(prev => (prev ? prev + '\n' : '') + errorMsg);
        console.error('Error fetching positions:', err);
      } finally {
        setPositionsLoading(false);
      }

      // Загрузка дерева юнитов
      try {
        setUnitsLoading(true);
        const unitData = await getOrganizationalUnitTree();
        setUnitTreeData(unitData || []);
      } catch (err) {
        setError(prev => prev ? prev + '\nНе удалось загрузить структуру организации.' : 'Не удалось загрузить структуру организации.');
        console.error('Error fetching organizational units:', err);
      } finally {
        setUnitsLoading(false);
      }
    };

    fetchData();
  }, []);

  // Обработчик изменения Департамента
  const handleDepartmentChange = (e) => {
    const depId = e.target.value;
    setSelectedDepartmentId(depId);
    setSelectedSubDepartmentId(''); // Сброс суб-департамента
    setSelectedSectorId(''); // Сброс сектора
    setSectors([]); // Очистка секторов

    if (depId) {
      const selectedDep = unitTreeData.find(unit => unit.id === parseInt(depId));
      setSubDepartments(selectedDep?.children || []);
    } else {
      setSubDepartments([]);
    }
  };

  // Обработчик изменения Суб-департамента
  const handleSubDepartmentChange = (e) => {
    const subDepId = e.target.value;
    setSelectedSubDepartmentId(subDepId);
    setSelectedSectorId(''); // Сброс сектора

    if (subDepId) {
      const selectedSubDep = subDepartments.find(unit => unit.id === parseInt(subDepId));
      setSectors(selectedSubDep?.children || []);
    } else {
      setSectors([]);
    }
  };

  // Обработчик изменения Сектора
  const handleSectorChange = (e) => {
    setSelectedSectorId(e.target.value);
  };
  const handleSubmit = async (event) => {
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
      // Определяем конечный ID юнита
      let finalUnitId = null;
      if (selectedSectorId) {
        finalUnitId = parseInt(selectedSectorId, 10);
      } else if (selectedSubDepartmentId) {
        finalUnitId = parseInt(selectedSubDepartmentId, 10);
      } else if (selectedDepartmentId) {
        finalUnitId = parseInt(selectedDepartmentId, 10);
      }

      // Проверяем, что юнит выбран (хотя бы департамент)
      if (!finalUnitId) {
        throw new Error('Пожалуйста, выберите департамент (и, если применимо, подразделение/сектор).');
      }
      // Проверяем, что должность выбрана
       if (!positionId) {
         throw new Error('Пожалуйста, выберите должность.');
       }

      const newUser = {
        Login: email,
        FullName: fullName,
        Email: email, // Оставляем, если бэкэнд ожидает
        Password: password,
        ConfirmPassword: confirmPassword,
        PositionID: parseInt(positionId, 10),
        OrganizationalUnitID: finalUnitId // Добавляем ID юнита
      };

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
                />
            </div>

            {/* Выбор Департамента */}
            <div className="form-group">
              <label htmlFor="register-department">Департамент/Управление</label>
              <select
                id="register-department"
                value={selectedDepartmentId}
                onChange={handleDepartmentChange}
                required // Делаем обязательным
                disabled={loading || unitsLoading}
              >
                <option value="">-- Выберите --</option>
                {unitsLoading ? (
                  <option value="" disabled>Загрузка...</option>
                ) : (
                  unitTreeData.map(unit => (
                    <option value={unit.id} key={unit.id}>
                      {unit.name} ({unit.unit_type})
                    </option>
                  ))
                )}
              </select>
            </div>

            {/* Выбор Суб-департамента/Отдела */}
            {selectedDepartmentId && subDepartments.length > 0 && (
              <div className="form-group">
                <label htmlFor="register-subdepartment">Отдел/Подотдел</label>
                <select
                  id="register-subdepartment"
                  value={selectedSubDepartmentId}
                  onChange={handleSubDepartmentChange}
                  disabled={loading || unitsLoading || !selectedDepartmentId}
                >
                  <option value="">-- Опционально --</option>
                  {subDepartments.map(unit => (
                    <option value={unit.id} key={unit.id}>
                      {unit.name} ({unit.unit_type})
                    </option>
                  ))}
                </select>
              </div>
            )}

            {/* Выбор Сектора */}
            {selectedSubDepartmentId && sectors.length > 0 && (
              <div className="form-group">
                <label htmlFor="register-sector">Сектор</label>
                <select
                  id="register-sector"
                  value={selectedSectorId}
                  onChange={handleSectorChange}
                  disabled={loading || unitsLoading || !selectedSubDepartmentId}
                >
                  <option value="">-- Опционально --</option>
                  {sectors.map(unit => (
                    <option value={unit.id} key={unit.id}>
                      {unit.name} ({unit.unit_type})
                    </option>
                  ))}
                </select>
              </div>
            )}

            {/* Выбор Должности */}
            <div className="form-group">
              <label htmlFor="register-position">Должность</label>
              <select
                id="register-position"
                value={positionId}
                onChange={(e) => setPositionId(e.target.value)}
                required
                disabled={loading || positionsLoading}
              >
                {positionsLoading ? (
                  <option value="" disabled>Загрузка...</option>
                 ) : positionsData.length === 0 ? (
                   <option value="" disabled>Нет доступных должностей</option>
                 ) : (
                   // Генерируем опции из плоского списка positionsData
                   positionsData.map(position => (
                     <option value={position.id} key={position.id}>
                       {position.name}
                     </option>
                   ))
                   /* Старая логика для групп:
                   positionsData.map(group => (
                     <optgroup label={group.name} key={group.id}>
                       {group.positions && group.positions.map(position => (
                         <option value={position.id} key={position.id}>
                           {position.name}                           
                         </option>
                       ))}
                     </optgroup>
                   ))
                   */
                 )}
               </select>
            </div>
            
            <button
              type="submit"
              disabled={loading || positionsLoading || unitsLoading} // Блокируем во время любой загрузки
            >
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
