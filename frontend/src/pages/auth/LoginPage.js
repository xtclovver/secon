import React, { useState, useContext } from 'react'; // Добавлен useContext
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { FaUser, FaLock } from 'react-icons/fa';
import { login } from '../../api/auth'; // Предполагается, что API файл будет создан
import { useUser } from '../../context/UserContext'; // Импортируем хук useUser
import './LoginPage.css';

const LoginPage = () => {
  const { setUser } = useUser(); // Получаем setUser из контекста
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    
    try {
      // Вызываем реальную функцию API для входа
      const data = await login(username, password); 
      
      // API должен вернуть токен и данные пользователя при успехе
      if (data && data.token && data.user) {
        // Сохраняем токен и пользователя (API уже должен был сохранить в localStorage)
        // localStorage.setItem('token', data.token); // Это делается внутри api/auth.js
        // localStorage.setItem('user', JSON.stringify(data.user)); // Это делается внутри api/auth.js
        
        // Обновляем состояние пользователя в контексте
        setUser(data.user); 
        
        // Определяем путь для редиректа на основе роли пользователя
        let redirectPath = '/dashboard'; // Путь по умолчанию
        if (data.user.isAdmin) {
          redirectPath = '/admin/dashboard';
        } else if (data.user.isManager) {
          redirectPath = '/manager/dashboard';
        }
        
        navigate(redirectPath, { replace: true }); // Перенаправляем пользователя
      } else {
         // Если API не вернуло ожидаемые данные
         throw new Error('Не удалось получить данные пользователя после входа.');
      }
    } catch (err) {
      setError(err.message || 'Ошибка при входе. Пожалуйста, проверьте учетные данные.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <motion.div 
      className="login-page"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
    >
      <motion.div 
        className="login-container"
        initial={{ y: -50, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ 
          type: 'spring', 
          stiffness: 300, 
          damping: 20,
          delay: 0.2
        }}
      >
        <motion.div 
          className="login-header"
          initial={{ y: -20, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          transition={{ delay: 0.3 }}
        >
          <h1>Система учета отпусков</h1>
          <p>Введите учетные данные для входа</p>
        </motion.div>
        
        {error && (
          <motion.div 
            className="error-message"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
          >
            {error}
          </motion.div>
        )}
        
        <form onSubmit={handleSubmit}>
          <motion.div 
            className="input-group"
            initial={{ x: -20, opacity: 0 }}
            animate={{ x: 0, opacity: 1 }}
            transition={{ delay: 0.4 }}
          >
            <FaUser className="input-icon" />
            <input
              type="text"
              id="username"
              placeholder="Имя пользователя (admin/manager/user)"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
          </motion.div>
          
          <motion.div 
            className="input-group"
            initial={{ x: -20, opacity: 0 }}
            animate={{ x: 0, opacity: 1 }}
            transition={{ delay: 0.5 }}
          >
            <FaLock className="input-icon" />
            <input
              type="password"
              id="password"
              placeholder="Пароль (admin/manager/user)"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </motion.div>
          
          <motion.button
            type="submit"
            className="login-button"
            disabled={loading}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            initial={{ y: 20, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            transition={{ delay: 0.6 }}
          >
            {loading ? 'Выполняется вход...' : 'Войти'}
          </motion.button>
        </form>
      </motion.div>
    </motion.div>
  );
};

export default LoginPage;
