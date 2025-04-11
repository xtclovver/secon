import React, { createContext, useContext } from 'react';

// Создание контекста пользователя
// Инициализируем с null, так как начальное состояние пользователя неизвестно
export const UserContext = createContext({
  user: null,
  setUser: () => {} // Пустая функция-заглушка для обновления пользователя
});

// Провайдер контекста пользователя
// Он будет получать user и setUser из App.js через пропс value
export const UserProvider = ({ children, value }) => {
  return (
    <UserContext.Provider value={value}>
      {children}
    </UserContext.Provider>
  );
};

// Хук для удобного использования контекста пользователя в компонентах
export const useUser = () => {
  const context = useContext(UserContext);
  if (context === undefined) {
    // Эта ошибка поможет отловить использование хука вне провайдера
    throw new Error('useUser должен использоваться внутри UserProvider');
  }
  return context; // Возвращает { user, setUser }
};

// Экспорт по умолчанию для возможного импорта без фигурных скобок
export default UserContext;
