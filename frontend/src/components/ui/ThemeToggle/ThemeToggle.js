import React, { useContext } from 'react';
import { motion } from 'framer-motion';
import { ThemeContext } from '../../../context/ThemeContext'; // Исправлен путь
import { FaSun, FaMoon } from 'react-icons/fa';
import './ThemeToggle.css';

const ThemeToggle = () => {
  const { darkMode, toggleTheme } = useContext(ThemeContext);

  return (
    <motion.div
      className="theme-toggle"
      onClick={toggleTheme}
      whileTap={{ scale: 0.95 }}
      whileHover={{ scale: 1.05 }}
      transition={{ type: 'spring', stiffness: 400, damping: 10 }}
    >
      <motion.div 
        className="toggle-track" 
        initial={false}
        animate={{ backgroundColor: darkMode ? '#2d3748' : '#cbd5e0' }}
      >
        <motion.div 
          className="toggle-thumb"
          initial={false}
          animate={{ 
            // x: darkMode ? 28 : 0, // Старое значение
            x: darkMode ? 24 : 2 // Исправлено: 48(track) - 20(thumb) - 2(padding) - 2(left offset) = 24px; 2px для светлой темы
            // backgroundColor: ... - Удалено, управляется CSS
          }}
          transition={{ type: 'spring', stiffness: 500, damping: 30 }}
        />
        
        <motion.div 
          className="toggle-icon sun"
          initial={false}
          animate={{ opacity: darkMode ? 0.2 : 1 }}
        >
          <FaSun />
        </motion.div>
        
        <motion.div 
          className="toggle-icon moon"
          initial={false}
          animate={{ opacity: darkMode ? 1 : 0.2 }}
        >
          <FaMoon />
        </motion.div>
      </motion.div>
    </motion.div>
  );
};

export default ThemeToggle;
