import React, { useState, useContext } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion'; // Добавлена AnimatePresence
import { 
  FaHome, 
  FaCalendarAlt, 
  FaList, 
  FaPlusCircle, 
  FaUsersCog, // Иконка для руководителя/админа
  FaChevronLeft,
  FaChevronRight,
  FaUserShield, // Иконка для админа
  FaUserTie // Иконка для руководителя
} from 'react-icons/fa';
import { ThemeContext } from '../../context/ThemeContext';
import { useUser } from '../../context/UserContext';
import './Sidebar.css'; // CSS для сайдбара

const Sidebar = () => {
  const { darkMode } = useContext(ThemeContext);
  const { user } = useUser(); 
  const [collapsed, setCollapsed] = useState(false); 
  const location = useLocation();

  // Логируем пользователя при рендере сайдбара
  console.log("Sidebar rendering with user:", user); 

  // Переключение состояния сворачивания сайдбара
  const toggleCollapse = () => {
    setCollapsed(!collapsed);
  };

  // Анимация для сайдбара
  const sidebarVariants = {
    open: { width: 'var(--sidebar-width)', transition: { duration: 0.3, ease: 'easeInOut' } },
    closed: { width: 'var(--sidebar-width-collapsed)', transition: { duration: 0.3, ease: 'easeInOut' } }
  };

  // Анимация для текста пунктов меню
  const itemTextVariants = {
    open: { opacity: 1, x: 0, display: 'inline', transition: { duration: 0.2, delay: 0.1 } },
    closed: { opacity: 0, x: -10, transitionEnd: { display: 'none' }, transition: { duration: 0.1 } }
  };
  
   // Анимация для заголовка "Меню"
  const menuTitleVariants = {
    open: { opacity: 1, transition: { duration: 0.3 } },
    closed: { opacity: 0, transition: { duration: 0.1 } }
  };

  return (
    <motion.aside 
      className={`sidebar ${darkMode ? 'dark' : 'light'} ${collapsed ? 'collapsed' : ''}`}
      variants={sidebarVariants}
      initial={false} // Не анимируем при первой загрузке
      animate={collapsed ? 'closed' : 'open'}
    >
      <div className="sidebar-header">
         <AnimatePresence>
          {!collapsed && (
            <motion.h3
              key="menu-title" // Ключ для AnimatePresence
              variants={menuTitleVariants}
              initial="closed"
              animate="open"
              exit="closed"
            >
              Меню
            </motion.h3>
          )}
        </AnimatePresence>
        
        <button className="collapse-btn" onClick={toggleCollapse} aria-label={collapsed ? "Развернуть меню" : "Свернуть меню"}>
          {collapsed ? <FaChevronRight /> : <FaChevronLeft />}
        </button>
      </div>
      
      <nav className="sidebar-nav">
        <ul>
          {/* Общие пункты меню */}
          <li>
            <NavLink to="/dashboard" className={({isActive}) => isActive ? 'active' : ''}>
              <FaHome />
              <motion.span variants={itemTextVariants} animate={collapsed ? 'closed' : 'open'}>
                Дашборд
              </motion.span>
            </NavLink>
          </li>
          <li>
            <NavLink to="/vacations/new" className={({isActive}) => isActive ? 'active' : ''}>
              <FaPlusCircle />
              <motion.span variants={itemTextVariants} animate={collapsed ? 'closed' : 'open'}>
                Оформить отпуск
              </motion.span>
            </NavLink>
          </li>
          <li>
            <NavLink to="/vacations/list" className={({isActive}) => isActive ? 'active' : ''}>
              <FaList />
              <motion.span variants={itemTextVariants} animate={collapsed ? 'closed' : 'open'}>
                Мои заявки
              </motion.span>
            </NavLink>
          </li>
          <li>
            <NavLink to="/vacations/calendar" className={({isActive}) => isActive ? 'active' : ''}>
              <FaCalendarAlt />
              <motion.span variants={itemTextVariants} animate={collapsed ? 'closed' : 'open'}>
                Календарь отпусков
              </motion.span>
            </NavLink>
          </li>
          
          {/* Пункты меню для руководителя */}
          {user?.isManager && (
            <li className="role-section manager-section">
              <NavLink to="/manager/dashboard" className={({isActive}) => isActive ? 'active' : ''}>
                <FaUserTie /> {/* Иконка руководителя */}
                <motion.span variants={itemTextVariants} animate={collapsed ? 'closed' : 'open'}>
                  Дашборд руководителя
                </motion.span>
              </NavLink>
              {/* Другие пункты для руководителя */}
            </li>
          )}
          
          {/* Пункты меню для администратора (временно без анимации для отладки) */}
          {user?.isAdmin && (
            <li className="role-section admin-section">
               <NavLink to="/admin/dashboard" className={({isActive}) => isActive ? 'active' : ''}>
                <FaUserShield /> {/* Иконка админа */}
                {/* Всегда показываем текст, если isAdmin true, для отладки */}
                <span>Админ-панель</span> 
              </NavLink>
               {/* Другие пункты для админа */}
            </li>
          )}
        </ul>
      </nav>
    </motion.aside>
  );
};

export default Sidebar;
