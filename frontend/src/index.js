import React from 'react';
import ReactDOM from 'react-dom/client'; // Используем новый API для React 18+
import App from './App';
// import reportWebVitals from './reportWebVitals'; // Можно раскомментировать для измерения производительности

// Импортируем основные стили здесь, чтобы они применялись глобально
import './styles/variables.css'; 
import './styles/App.css'; 

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);

// Если вы хотите измерять производительность вашего приложения, передайте функцию
// для логирования результатов (например: reportWebVitals(console.log))
// или отправки на эндпоинт аналитики. Узнайте больше: https://bit.ly/CRA-vitals
// reportWebVitals();
