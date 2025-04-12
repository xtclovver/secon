import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { isAuthenticated } from '../../api/auth'; // Импортируем функцию проверки аутентификации
import { useUser } from '../../context/UserContext'; // Импортируем хук пользователя

/**
 * Компонент для защиты маршрутов, требующих аутентификации.
 * 
 * Проверяет, аутентифицирован ли пользователь с помощью `isAuthenticated`.
 * Если нет, перенаправляет на страницу `/login`, сохраняя исходный путь
 * для возможного редиректа после успешного входа.
 * 
 * Если пользователь аутентифицирован, отображает дочерние маршруты (`<Outlet />`).
 */
const ProtectedRoute = ({ allowedRoles }) => { 
  const location = useLocation();
  const { user } = useUser(); 

  // Логируем пользователя, которого видит ProtectedRoute
  console.log("ProtectedRoute received user:", user); 
  console.log("ProtectedRoute checking path:", location.pathname);
  console.log("Allowed roles:", allowedRoles);

  // Проверяем аутентификацию
  if (!isAuthenticated() || !user) {
    // Пользователь не вошел в систему или данные пользователя еще не загружены
    // Перенаправляем на страницу логина, передавая текущий путь в state
    // чтобы можно было вернуться сюда после логина
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // Проверяем роли, если они указаны
  if (allowedRoles && allowedRoles.length > 0) {
    const userHasRequiredRole = 
        (allowedRoles.includes('admin') && user.isAdmin) ||
        (allowedRoles.includes('manager') && user.isManager) ||
        (allowedRoles.includes('user') && !user.isAdmin && !user.isManager); // Предполагаем, что 'user' - это обычный пользователь

    if (!userHasRequiredRole) {
      // У пользователя нет необходимой роли, перенаправляем на его дашборд или страницу ошибки доступа
      console.warn(`Доступ запрещен: Пользователь ${user.username} не имеет роли из списка [${allowedRoles.join(', ')}] для доступа к ${location.pathname}`);
      // Редирект на основной дашборд пользователя как fallback
       const fallbackPath = user.isAdmin ? "/admin/dashboard" : user.isManager ? "/manager/dashboard" : "/dashboard";
      return <Navigate to={fallbackPath} replace />;
      // Или можно редиректить на специальную страницу "Доступ запрещен"
      // return <Navigate to="/unauthorized" replace />; 
    }
  }

  // Если пользователь аутентифицирован и (если роли указаны) имеет нужную роль,
  // отображаем запрошенный компонент/страницу
  return <Outlet />;
};

export default ProtectedRoute;
