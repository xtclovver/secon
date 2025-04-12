import React, { useState, useEffect, Suspense, lazy } from 'react'; // Добавлен Suspense и lazy
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ToastContainer } from 'react-toastify';
import { AnimatePresence } from 'framer-motion';

// Context
import { ThemeProvider } from './context/ThemeContext';
import { UserProvider } from './context/UserContext'; 

// Стили (перемещены наверх)
import 'react-toastify/dist/ReactToastify.css';
import './styles/App.css'; 
import './styles/variables.css'; 

// Компоненты
import ProtectedRoute from './components/ProtectedRoute/ProtectedRoute'; // Исправлен путь
import Header from './components/Header/Header'; // Исправлен путь
import Sidebar from './components/Sidebar/Sidebar'; // Исправлен путь
import Footer from './components/Footer/Footer'; // Исправлен путь
import Loader from './components/ui/Loader/Loader'; 

// Сервисы (перемещены выше)
import { isAuthenticated, getCurrentUser, logout } from './api/auth';

// Страницы (используем lazy loading)
const LoginPage = lazy(() => import('./pages/auth/LoginPage'));
const RegisterPage = lazy(() => import('./pages/auth/RegisterPage')); // Добавляем страницу регистрации
const UserDashboard = lazy(() => import('./pages/dashboard/UserDashboard')); // Будет создан позже
const ManagerDashboard = lazy(() => import('./pages/dashboard/ManagerDashboard'));
const AdminDashboard = lazy(() => import('./pages/dashboard/AdminDashboard')); // Будет создан позже
const VacationForm = lazy(() => import('./pages/vacations/VacationForm'));
const VacationsList = lazy(() => import('./pages/vacations/VacationsList')); // Будет создан позже
const VacationCalendar = lazy(() => import('./pages/vacations/VacationCalendar')); // Будет создан позже
const NotFoundPage = lazy(() => import('./pages/NotFoundPage')); // Будет создан позже


const App = () => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true); // Состояние загрузки данных пользователя

  useEffect(() => {
    const fetchUser = async () => {
      console.log("App useEffect running..."); // Лог 1: Проверяем запуск useEffect
      if (isAuthenticated()) {
        console.log("User is authenticated (token found)."); // Лог 2: Проверяем аутентификацию
        try {
          const storedUser = localStorage.getItem('user');
          console.log("Stored user data from localStorage:", storedUser); // Лог 3: Смотрим, что в localStorage
          if (storedUser) {
             setUser(JSON.parse(storedUser));
          } else {
             // Если пользователя нет в localStorage, но есть токен, разлогиниваем
             // Если пользователя нет в localStorage, но есть токен, разлогиниваем
             console.log("Token exists, but no user data in localStorage. Logging out.");
             logout();
             // Важно выйти из try/catch или функции после logout, чтобы не пытаться парсить null
             setLoading(false); // Завершаем загрузку в этом случае тоже
             return; 
          }
          
          // Пытаемся парсить только если storedUser не null
          const userData = JSON.parse(storedUser); 
          setUser(userData); // Устанавливаем состояние
          console.log("User data set in App state:", userData); // Лог 4: Выводим данные пользователя ПОСЛЕ setUser
          
          // --- Закомментированный блок для реального API ---
          // const realUserData = await getCurrentUser(); 
          // console.log("User data from API:", realUserData); 
          // if (realUserData) {
          //    setUser(realUserData);
          // } else {
          //    logout(); // Разлогиниваем, если API не вернуло пользователя
          // }
          // --- Конец закомментированного блока ---

         } catch (error) {
          console.error('Error fetching/parsing user data:', error); // Лог 5: Ловим ошибки в try
          // Если произошла ошибка (например, токен невалиден или JSON некорректен), разлогиниваем
          logout();
        }
      } else {
          console.log("User is not authenticated (no token found)."); // Лог 6: Если токена нет
      }
      setLoading(false); // Завершаем загрузку
      console.log("App useEffect finished."); // Лог 7: Проверяем завершение useEffect
    };

    fetchUser();
  }, []); // Пустой массив зависимостей, чтобы выполнилось один раз при монтировании

  // Отображение глобального загрузчика во время проверки аутентификации
  if (loading) {
    return <Loader />; // Отображаем лоадер на весь экран
  }

  return (
    <ThemeProvider>
      {/* Передаем user и setUser в провайдер */}
      <UserProvider value={{ user, setUser }}>
        <Router>
          <div className="app">
            <ToastContainer
              position="top-right"
              autoClose={5000}
              hideProgressBar={false}
              newestOnTop
              closeOnClick
              rtl={false}
              pauseOnFocusLoss
              draggable
              pauseOnHover
              theme="colored" // Используем цветные уведомления
            />

            {/* Отображаем Header и Sidebar только если пользователь аутентифицирован */}
            {isAuthenticated() && user && <Header />} 
            
            <div className="app-container">
              {isAuthenticated() && user && <Sidebar />}
              
              <main className="app-content">
                {/* Suspense для обработки ленивой загрузки компонентов */}
                <Suspense fallback={<Loader />}> 
                  <AnimatePresence mode="wait">
                    <Routes>
                      {/* Общедоступные маршруты */}
                      <Route 
                        path="/login"
                        element={isAuthenticated() && user ? <Navigate to="/dashboard" replace /> : <LoginPage />}
                      />
                      <Route
                        path="/register"
                        element={isAuthenticated() && user ? <Navigate to="/dashboard" replace /> : <RegisterPage />} // Добавляем маршрут регистрации
                      />

                      {/* Защищенные маршруты */}
                      <Route element={<ProtectedRoute />}>
                        {/* Перенаправление на соответствующий дашборд */}
                        <Route 
                          path="/" 
                          element={
                            <Navigate 
                              to={
                                !user // Если user еще null (маловероятно из-за ProtectedRoute, но для надежности)
                                  ? "/login" 
                                  : user.isAdmin 
                                    ? "/admin/dashboard" 
                                    : user.isManager 
                                      ? "/manager/dashboard" 
                                      : "/dashboard"
                              } 
                              replace 
                            />
                          } 
                        />
                        
                        {/* Маршруты для всех аутентифицированных пользователей */}
                        <Route path="/dashboard" element={<UserDashboard />} />
                        <Route path="/vacations/new" element={<VacationForm />} />
                        <Route path="/vacations/list" element={<VacationsList />} />
                        <Route path="/vacations/calendar" element={<VacationCalendar />} />
                        
                        {/* Маршруты для руководителей (дополнительная проверка роли) */}
                        <Route 
                          path="/manager/dashboard" 
                          element={
                            user?.isManager ? <ManagerDashboard /> : <Navigate to="/dashboard" replace />
                          } 
                        />
                        
                        {/* Маршруты для администраторов (дополнительная проверка роли) */}
                        <Route 
                          path="/admin/dashboard" 
                          element={
                            user?.isAdmin ? <AdminDashboard /> : <Navigate to="/dashboard" replace />
                          } 
                        />
                      </Route>
                      
                      {/* Маршрут для страницы 404 */}
                      <Route path="*" element={<NotFoundPage />} />
                    </Routes>
                  </AnimatePresence>
                </Suspense>
              </main>
            </div>
            
            {isAuthenticated() && user && <Footer />}
          </div>
        </Router>
      </UserProvider>
    </ThemeProvider>
  );
};

export default App;
