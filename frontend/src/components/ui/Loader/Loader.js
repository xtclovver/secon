import React from 'react';
import './Loader.css'; // CSS для лоадера

/**
 * Компонент для отображения индикатора загрузки.
 * Может использоваться как полноэкранный оверлей или встроенный индикатор.
 * 
 * @param {boolean} overlay - Если true, отображается как полноэкранный оверлей.
 * @param {string} text - Текст, отображаемый под индикатором (опционально).
 */
const Loader = ({ overlay = false, text = 'Загрузка...' }) => {
  if (overlay) {
    return (
      <div className="loader-overlay">
        <div className="loader-container">
          <div className="loader-spinner"></div>
          {text && <p className="loader-text">{text}</p>}
        </div>
      </div>
    );
  }

  return (
    <div className="loader-inline-container">
      <div className="loader-spinner small"></div>
      {text && <p className="loader-text small">{text}</p>}
    </div>
  );
};

export default Loader;
