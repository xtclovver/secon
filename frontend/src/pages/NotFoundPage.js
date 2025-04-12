import React from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { FaExclamationTriangle, FaHome } from 'react-icons/fa';
// import './NotFoundPage.css'; // Можно добавить стили

const NotFoundPage = () => {
  return (
    <motion.div
      className="not-found-container card" // Используем card для общего стиля
      style={{ textAlign: 'center', maxWidth: '600px', margin: '50px auto' }}
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.5, type: 'spring' }}
    >
      <FaExclamationTriangle size={50} style={{ color: 'var(--warning-color)', marginBottom: '20px' }} />
      <h1>404 - Страница не найдена</h1>
      <p style={{ color: 'var(--text-secondary)', marginBottom: '30px' }}>
        К сожалению, страница, которую вы ищете, не существует или была перемещена.
      </p>
      <Link to="/" className="btn btn-primary">
        <FaHome /> Вернуться на главную
      </Link>
    </motion.div>
  );
};

export default NotFoundPage;
