import React, { useState, useContext } from 'react';
import { Link, NavLink } from 'react-router-dom'; // Используем NavLink для активных ссылок
import { motion, AnimatePresence } from 'framer-motion';
import { FaUser, FaBell, FaSignOutAlt, FaBars, FaTimes } from 'react-icons/fa'; // Добавлена иконка FaTimes
import ThemeToggle from '../ui/ThemeToggle/ThemeToggle'; // Путь к переключателю темы
import { ThemeContext } from '../../context/ThemeContext';
import { useUser } from '../../context/UserContext';
import { logout } from '../../api/auth';
import './Header.css'; // CSS для хедера

const Header = () => {
  const { darkMode } = useContext(ThemeContext);
  const { user } = useUser(); 
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const [isNotificationsOpen, setIsNotificationsOpen] = useState(false);
  
  // Логируем пользователя при рендере хедера
  console.log("Header rendering with user:", user); 
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  // Обработчик выхода из системы
  const handleLogout = () => {
    logout();
    // Дополнительно можно сбросить состояние пользователя в контексте, если необходимо
    // setUser(null); 
  };

  // Закрытие всех выпадающих меню при клике вне них
  // useEffect(() => {
  //   const handleClickOutside = () => {
  //     setIsProfileOpen(false);
  //     setIsNotificationsOpen(false);
  //   };
  //   document.addEventListener('click', handleClickOutside);
  //   return () => document.removeEventListener('click', handleClickOutside);
  // }, []); // Этот useEffect может вызывать проблемы, если клик происходит на кнопке открытия

  // Переключение меню профиля
  const toggleProfileMenu = (e) => {
    e.stopPropagation(); // Предотвращаем всплытие события
    setIsProfileOpen(prev => !prev);
    setIsNotificationsOpen(false); // Закрываем другое меню
  };

  // Переключение меню уведомлений
  const toggleNotificationsMenu = (e) => {
    e.stopPropagation(); // Предотвращаем всплытие события
    setIsNotificationsOpen(prev => !prev);
    setIsProfileOpen(false); // Закрываем другое меню
  };

  // Переключение мобильного меню
  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(prev => !prev);
  };
  
  // Закрытие мобильного меню при клике на ссылку
  const closeMobileMenu = () => {
    setIsMobileMenuOpen(false);
  };

  // Анимация для выпадающих меню
  const dropdownVariants = {
    hidden: { opacity: 0, y: -10, scale: 0.95 },
    visible: { opacity: 1, y: 0, scale: 1, transition: { duration: 0.2 } },
    exit: { opacity: 0, y: -10, scale: 0.95, transition: { duration: 0.15 } }
  };
  
  // Анимация для мобильного меню
  const mobileMenuVariants = {
      hidden: { opacity: 0, height: 0 },
      visible: { opacity: 1, height: 'auto', transition: { duration: 0.3, ease: "easeInOut" } },
      exit: { opacity: 0, height: 0, transition: { duration: 0.2, ease: "easeInOut" } }
  };

  return (
    <header className={`app-header ${darkMode ? 'dark' : 'light'}`}>
      <div className="header-container">
        <div className="header-left">
          <Link to="/" className="logo">
            <motion.span
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
            >
              Система учета отпусков
            </motion.span>
          </Link>
          
          {/* Кнопка мобильного меню */}
          <button className="mobile-menu-button" onClick={toggleMobileMenu}>
            {isMobileMenuOpen ? <FaTimes /> : <FaBars />}
          </button>
        </div>
        
        {/* Мобильное меню */}
        <AnimatePresence>
          {isMobileMenuOpen && (
            <motion.div 
              className="mobile-menu"
              variants={mobileMenuVariants}
              initial="hidden"
              animate="visible"
              exit="exit"
            >
              <nav className="mobile-nav">
                <NavLink to="/dashboard" onClick={closeMobileMenu} className={({isActive}) => isActive ? "active" : ""}>Дашборд</NavLink>
                <NavLink to="/vacations/new" onClick={closeMobileMenu} className={({isActive}) => isActive ? "active" : ""}>Оформить отпуск</NavLink>
                <NavLink to="/vacations/list" onClick={closeMobileMenu} className={({isActive}) => isActive ? "active" : ""}>Мои заявки</NavLink>
                <NavLink to="/vacations/calendar" onClick={closeMobileMenu} className={({isActive}) => isActive ? "active" : ""}>Календарь</NavLink>
                {user?.isManager && (
                  <NavLink to="/manager/dashboard" onClick={closeMobileMenu} className={({isActive}) => isActive ? "active" : ""}>Дашборд руководителя</NavLink>
                )}
                {/* Возвращаем isAdmin */}
                {user?.isAdmin && ( 
                  <NavLink to="/admin/dashboard" onClick={closeMobileMenu} className={({isActive}) => isActive ? "active" : ""}>Админ-панель</NavLink>
                )}
                 <button onClick={() => { handleLogout(); closeMobileMenu(); }} className="mobile-logout-button">
                    <FaSignOutAlt /> Выйти
                 </button>
              </nav>
            </motion.div>
          )}
        </AnimatePresence>
        
        {/* Правая часть хедера (скрыта на мобильных, если меню открыто) */}
        <div className={`header-right ${isMobileMenuOpen ? 'hidden-on-mobile' : ''}`}>
          <ThemeToggle />
          
          {/* Уведомления */}
          <div className="notifications-dropdown">
            <button 
              className="icon-button notifications-button" // Добавлен класс icon-button
              onClick={toggleNotificationsMenu}
              aria-label="Уведомления"
            >
              <FaBell />
              {/* Заглушка для количества уведомлений */}
              <span className="badge">0</span> 
            </button>
            
            <AnimatePresence>
              {isNotificationsOpen && (
                <motion.div 
                  className="dropdown-menu notifications-menu" // Добавлен класс notifications-menu
                  variants={dropdownVariants}
                  initial="hidden"
                  animate="visible"
                  exit="exit"
                  onClick={(e) => e.stopPropagation()} // Предотвращаем закрытие при клике внутри меню
                >
                  <div className="menu-header">Уведомления</div>
                  <div className="empty-notifications">
                    <p>Нет новых уведомлений</p>
                    {/* Здесь будет список уведомлений */}
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
          
          {/* Профиль пользователя */}
          <div className="profile-dropdown">
            <button 
              className="profile-button" 
              onClick={toggleProfileMenu}
              aria-label="Профиль пользователя"
            >
              <FaUser />
              {/* Используем fullName (после исправления transformUserKeys) */}
              <span className="user-name">{user?.fullName || 'Пользователь'}</span>
            </button>

            <AnimatePresence>
              {isProfileOpen && (
                <motion.div 
                  className="dropdown-menu profile-menu" // Добавлен класс profile-menu
                   variants={dropdownVariants}
                   initial="hidden"
                   animate="visible"
                   exit="exit"
                   onClick={(e) => e.stopPropagation()} // Предотвращаем закрытие при клике внутри меню
                >
                  <div className="menu-header">
                    <div className="user-info">
                      {/* Используем fullName */}
                      <strong>{user?.fullName}</strong>
                      <span className="user-email">{user?.email}</span>
                      {/* --- DEBUGGING ROLES --- */}
                      {/* Используем isAdmin и isManager */}
                      {console.log(`Header Role Check: isAdmin=${user?.isAdmin}, isManager=${user?.isManager}`)} 
                      {/* Отображение ролей */}
                      {user?.isManager && <span className="user-role manager">Руководитель</span>} 
                      {user?.isAdmin && <span className="user-role admin">Администратор</span>} 
                      {/* Проверяем оба флага перед отображением "Сотрудник" */}
                      {!user?.isManager && !user?.isAdmin && <span className="user-role">Сотрудник</span>} 
                    </div>
                  </div>
                  <ul className="menu-list">
                    <li>
                      {/* Добавляем ссылку на профиль */}
                      <Link to="/profile" className="menu-link" onClick={() => setIsProfileOpen(false)}>
                        <FaUser /> {/* Можно использовать другую иконку, если есть */}
                        <span>Профиль</span>
                      </Link>
                    </li>
                    <li>
                      <button onClick={handleLogout} className="logout-button">
                        <FaSignOutAlt />
                        <span>Выйти</span>
                      </button>
                    </li>
                    {/* Другие пункты меню профиля */}
                  </ul>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
