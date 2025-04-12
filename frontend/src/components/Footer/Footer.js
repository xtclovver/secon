import React, { useContext } from 'react';
import { ThemeContext } from '../../context/ThemeContext';
import './Footer.css'; // CSS для футера

const Footer = () => {
  const { darkMode } = useContext(ThemeContext);
  const currentYear = new Date().getFullYear();

  return (
    <footer className={`app-footer ${darkMode ? 'dark' : 'light'}`}>
      <div className="footer-container">
        <p>
          &copy; {currentYear} Система учета отпусков. Все права защищены.
        </p>
        {/* Можно добавить ссылки или другую информацию */}
        {/* <nav>
          <a href="/privacy">Политика конфиденциальности</a>
          <a href="/terms">Условия использования</a>
        </nav> */}
      </div>
    </footer>
  );
};

export default Footer;
