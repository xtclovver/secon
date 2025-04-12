import React, { useState, useEffect, Suspense, lazy, useCallback } from 'react'; // Добавлен useCallback
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
import { Outlet } from 'react-router-dom'; // Импортируем Outlet

// Сервисы (перемещены выше)
import { isAuthenticated, getCurrentUser, logout } from './api/auth';
import { getVacationLimit } from './api/vacations'; // Импортируем API для получения лимита

// Auth pages (lazy loading)
const LoginPage = lazy(() => import('./pages/auth/LoginPage'));
const RegisterPage = lazy(() => import('./pages/auth/RegisterPage'));

// Other pages (lazy loading)
// const UserDashboard = lazy(() => import('./pages/dashboard/UserDashboard')); // Больше не используется напрямую
const UniversalDashboard = lazy(() => import('./pages/dashboard/UniversalDashboard')); // Добавляем универсальный дашборд
const ManagerDashboard = lazy(() => import('./pages/dashboard/ManagerDashboard'));
const AdminDashboard = lazy(() => import('./pages/dashboard/AdminDashboard'));
const VacationForm = lazy(() => import('./pages/vacations/VacationForm'));
const VacationsList = lazy(() => import('./pages/vacations/VacationsList')); // Будет создан позже
const VacationCalendar = lazy(() => import('./pages/vacations/VacationCalendar')); // Будет создан позже
const UserProfilePage = lazy(() => import('./pages/profile/UserProfilePage')); // Добавляем страницу профиля
const NotFoundPage = lazy(() => import('./pages/NotFoundPage')); // Будет создан позже

// Компонент-обертка для основного макета приложения
const MainLayout = () => (
  <>
    <Header />
    <div className="app-container">
      <Sidebar />
      <main className="app-content">
        {/* Suspense нужен здесь, если дочерние компоненты тяжелые */}
        <Suspense fallback={<Loader />}>
           <Outlet /> {/* Здесь будут рендериться дочерние защищенные маршруты */}
        </Suspense>
      </main>
    </div>
    <Footer />
  </>
);

const App = () => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true); // Состояние загрузки данных пользователя

  // Функция для обновления лимитов отпуска пользователя
  const refreshUserVacationLimits = useCallback(async () => {
    // Проверяем наличие токена, а не объекта user, который может быть старым
    if (!isAuthenticated()) {
        console.log("Not authenticated, skipping limits refresh.");
        return;
    }

    const currentYear = new Date().getFullYear();
    console.log(`Refreshing vacation limits for year ${currentYear}`);
    try {
      // API getVacationLimit использует токен для идентификации пользователя
      const limits = await getVacationLimit(currentYear); // API returns { total_days, used_days, ... }

      // ИСПРАВЛЕНО: Используем правильные имена полей из API
      const total_days = limits.total_days ?? 0;
      const used_days = limits.used_days ?? 0;
      const availableDays = total_days - used_days; // Correct calculation

      // Логирование с правильными именами
      console.log(`Fetched limits: total_days=${total_days}, used_days=${used_days}, Available=${availableDays}`);

      // Используем функциональную форму setUser, чтобы не зависеть от 'user'
      setUser(prevUser => {
        // Если по какой-то причине предыдущего пользователя нет, но мы аутентифицированы,
        // возможно, стоит вернуть null или базовый объект, но пока просто проверим.
        if (!prevUser) {
            console.log("setUser in refresh: prevUser is null, returning null.");
            return null;
        }
        const updatedUser = {
          ...prevUser,
          vacationLimits: { // Добавляем или обновляем объект с лимитами
            ...(prevUser.vacationLimits || {}), // Сохраняем другие года, если они есть
            [currentYear]: {
              // ИСПРАВЛЕНО: Сохраняем с правильными именами, но можно и с camelCase, если консистентно
              totalDays: total_days, // Сохраним как camelCase в стейте для удобства
              usedDays: used_days,   // Сохраним как camelCase в стейте для удобства
              availableDays: availableDays,
            }
          },
          // Обновляем поля верхнего уровня (camelCase)
          currentAvailableDays: availableDays,
          currentTotalDays: total_days,
          currentUsedDays: used_days,
        };
        console.log("Updating user state with new limits:", updatedUser);
        localStorage.setItem('user', JSON.stringify(updatedUser));
        return updatedUser;
      });
    } catch (error) {
      console.error("Failed to refresh vacation limits:", error);
      // Возможно, стоит разлогинить пользователя, если запрос лимитов критичен или вернул 401/403
      if (error.response?.status === 401 || error.response?.status === 403) {
          logout();
          setUser(null); // Сбрасываем пользователя в стейте
      }
    }
  }, []); // <-- УБРАЛИ 'user' из зависимостей

  // Основной useEffect для загрузки пользователя
   useEffect(() => {
    let isMounted = true; // Флаг для предотвращения обновления состояния в размонтированном компоненте
    const fetchUser = async () => {
        if (!isMounted) return; // Прерываем, если компонент размонтирован
        setLoading(true);
        if (isAuthenticated()) {
            const storedUserString = localStorage.getItem('user');
            if (storedUserString) {
                try {
                    const storedUser = JSON.parse(storedUserString);
                    if (isMounted) {
                        setUser(storedUser);
                        // Вызываем обновление лимитов ПОСЛЕ установки пользователя.
                        // refreshUserVacationLimits теперь стабильна (нет зависимостей).
                        await refreshUserVacationLimits(); // Убрали setTimeout, вызываем напрямую
                    }
                } catch (parseError) {
                    console.error("Error parsing user from localStorage:", parseError);
                    logout(); // Разлогиниваем при ошибке парсинга
                    if (isMounted) setUser(null);
                }
            } else if (localStorage.getItem('token')) {
                 // Есть токен, но нет пользователя в localStorage - некорректное состояние
                 console.log("Token exists, but no user data in localStorage. Logging out.");
                 logout();
                 if (isMounted) setUser(null);
            } else {
                 // Нет токена и нет пользователя в localStorage (не залогинен)
                 if (isMounted) setUser(null);
            }
        } else {
             // Не аутентифицирован (нет токена)
             if (isMounted) setUser(null);
        }
        if (isMounted) {
             setLoading(false);
        }
    };

    fetchUser();

    return () => {
      isMounted = false; // Устанавливаем флаг при размонтировании
    };
    // Зависимость от refreshUserVacationLimits остается, но функция теперь стабильна
  }, [refreshUserVacationLimits]);
  // Отображение глобального загрузчика во время проверки аутентификации
  if (loading) {
    return <Loader />; // Отображаем лоадер на весь экран
  }

  return (
    <ThemeProvider>
      {/* Передаем user, setUser и refreshUserVacationLimits */}
      <UserProvider value={{ user, setUser, refreshUserVacationLimits }}>
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
            
            {/* Suspense для обработки ленивой загрузки страниц */}
            <Suspense fallback={<Loader />}>
               <AnimatePresence mode="wait">
                  <Routes>
                    {/* Auth Routes (вне MainLayout) */}
                    <Route
                      path="/login"
                      element={isAuthenticated() && user ? <Navigate to="/profile" replace /> : <LoginPage />}
                    />
                    <Route
                      path="/register"
                      element={isAuthenticated() && user ? <Navigate to="/profile" replace /> : <RegisterPage />}
                    />

                    {/* Protected Routes (внутри MainLayout) */}
                    {/* Обертка ProtectedRoute проверяет аутентификацию */}
                    <Route element={<ProtectedRoute />}>
                      {/* MainLayout применяется ко всем вложенным маршрутам */}
                      <Route element={<MainLayout />}>
                         {/* Перенаправление с главной на дашборд */}
                         <Route 
                           path="/" 
                           element={
                             <Navigate
                               to="/dashboard" // Всегда на /dashboard если залогинен и прошел ProtectedRoute
                               replace
                             />
                           } 
                         />
                         {/* Остальные защищенные маршруты */}
                         <Route path="/dashboard" element={<UniversalDashboard />} />
                         <Route path="/vacations/new" element={<VacationForm />} />
                         <Route path="/vacations/list" element={<VacationsList />} />
                         <Route path="/vacations/calendar" element={<VacationCalendar />} />
                         <Route path="/profile" element={<UserProfilePage />} />
                         {/* Маршруты для всех аутентифицированных пользователей */}
                         <Route path="/dashboard" element={<UniversalDashboard />} />
                         <Route path="/vacations/new" element={<VacationForm />} />
                         <Route path="/vacations/list" element={<VacationsList />} />
                         <Route path="/vacations/calendar" element={<VacationCalendar />} />
                         <Route path="/profile" element={<UserProfilePage />} />

                         {/* Маршруты для руководителей (дополнительная проверка роли) */}
                         <Route
                           path="/manager/dashboard"
                           element={
                             user?.role === 'manager' ? <ManagerDashboard /> : <Navigate to="/dashboard" replace />
                           }
                         />

                         {/* Маршруты для администраторов (дополнительная проверка роли) */}
                         <Route
                           path="/admin/dashboard"
                           element={
                             user?.isAdmin ? <AdminDashboard /> : <Navigate to="/dashboard" replace />
                           }
                         />
                      </Route> {/* Конец MainLayout */}
                    </Route> {/* Конец ProtectedRoute */}
                      
                    {/* Маршрут 404 (вне MainLayout, но можно и внутри, если нужен хедер/футер) */}
                    <Route path="*" element={<NotFoundPage />} />
                  </Routes>
               </AnimatePresence>
             </Suspense>
          </div>
        </Router>
      </UserProvider>
    </ThemeProvider>
  );
};

export default App;
