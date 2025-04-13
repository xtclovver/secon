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
const UniversalDashboard = lazy(() => import('./pages/dashboard/UniversalDashboard')); // Добавляем универсальный дашборд
const ManagerDashboard = lazy(() => import('./pages/dashboard/ManagerDashboard'));
const AdminDashboard = lazy(() => import('./pages/dashboard/AdminDashboard'));
const VacationForm = lazy(() => import('./pages/vacations/VacationForm'));
const VacationsList = lazy(() => import('./pages/vacations/VacationsList')); // Будет создан позже
const VacationCalendar = lazy(() => import('./pages/vacations/VacationCalendar')); // Будет создан позже
const UserProfilePage = lazy(() => import('./pages/profile/UserProfilePage'));
const DepartmentManagementPage = lazy(() => import('./pages/admin/DepartmentManagementPage'));
const UserManagementPage = lazy(() => import('./pages/admin/UserManagementPage'));
const ExportVacationsPage = lazy(() => import('./pages/admin/ExportVacationsPage')); // <-- Добавляем импорт страницы экспорта
const NotFoundPage = lazy(() => import('./pages/NotFoundPage'));

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
  const [limitsLoading, setLimitsLoading] = useState(false); // Новое состояние для загрузки лимитов

  // Функция для обновления лимитов отпуска пользователя (принимает год)
  const refreshUserVacationLimits = useCallback(async (yearToRefresh) => { // <-- Принимает год
    // Проверяем наличие токена
    if (!isAuthenticated()) {
        console.log("Not authenticated, skipping limits refresh.");
        return; // Выходим, если не аутентифицирован
    }
    setLimitsLoading(true); // <-- Устанавливаем флаг загрузки лимитов
    console.log(`Refreshing vacation limits for year ${yearToRefresh}`);
    try {
      // Используем переданный год
      const limits = await getVacationLimit(yearToRefresh); // API returns { total_days, used_days, ... }

      // Используем правильные имена полей из API
      const total_days = limits.total_days ?? null; // <-- Используем null как индикатор отсутствия данных
      const used_days = limits.used_days ?? null;

      // !!! ДОБАВЛЕНО ЛОГИРОВАНИЕ СЫРОГО ОТВЕТА API !!!
      console.log(`>>> Raw API response for getVacationLimit(${yearToRefresh}):`, JSON.stringify(limits));

      // Рассчитываем доступные дни только если total_days не null
      const availableDays = total_days !== null && used_days !== null ? total_days - used_days : null;

      // Логирование обработанных значений
      console.log(`Fetched limits for ${yearToRefresh}: total_days=${total_days}, used_days=${used_days}, Available=${availableDays}`);


      // Используем функциональную форму setUser, чтобы не зависеть от 'user'
      setUser(prevUser => {
        if (!prevUser) {
            console.warn("setUser in refresh: prevUser is null, but authenticated. Setting partial user data.");
             // Возможно, создать базовый объект пользователя, если он должен быть?
             // Или просто не обновлять, если prevUser null. Пока не обновляем.
            return null;
        }
        const updatedLimits = {
           ...(prevUser.vacationLimits || {}),
           [yearToRefresh]: { // <-- Используем переданный год как ключ
               totalDays: total_days, // Сохраняем как camelCase
               usedDays: used_days,   // Сохраняем как camelCase
               availableDays: availableDays,
               // Добавим флаг, что данные были загружены (даже если они null)
               loaded: true
           }
        };
        const updatedUser = {
          ...prevUser,
          vacationLimits: updatedLimits,
        };

        // Опционально: Обновлять поля верхнего уровня только если обновляется текущий год
        const currentSystemYear = new Date().getFullYear();
        if (yearToRefresh === currentSystemYear) {
            updatedUser.currentAvailableDays = availableDays;
            updatedUser.currentTotalDays = total_days;
            updatedUser.currentUsedDays = used_days;
        }

        console.log(`Updating user state with limits for year ${yearToRefresh}:`, updatedUser);
        localStorage.setItem('user', JSON.stringify(updatedUser)); // Сохраняем обновленного пользователя
        return updatedUser;
      });

    } catch (error) {
        console.error(`Failed to refresh vacation limits for year ${yearToRefresh}:`, error);
        // Сохраняем информацию об ошибке в состоянии пользователя для данного года
        setUser(prevUser => {
          if (!prevUser) return null;
          const updatedLimits = {
             ...(prevUser.vacationLimits || {}),
             [yearToRefresh]: {
                 loaded: true, // Отмечаем, что попытка загрузки была
                 error: error.message || 'Не удалось загрузить лимит' // Сохраняем текст ошибки
             }
          };
           const updatedUser = { ...prevUser, vacationLimits: updatedLimits };
           localStorage.setItem('user', JSON.stringify(updatedUser));
           return updatedUser;
        });
        // Разлогиниваем только при 401/403
        if (error.response?.status === 401 || error.response?.status === 403) {
            logout();
            setUser(null);
        }
    } finally {
         setLimitsLoading(false); // <-- Сбрасываем флаг загрузки лимитов
    }
  }, []); // <-- Зависимостей нет, функция стабильна

  // Основной useEffect для загрузки пользователя
  useEffect(() => {
    let isMounted = true;
    const fetchUser = async () => {
        if (!isMounted) return;
        setLoading(true);
        if (isAuthenticated()) {
            const storedUserString = localStorage.getItem('user');
            let userToSet = null; // Временная переменная для пользователя
            if (storedUserString) {
                try {
                    userToSet = JSON.parse(storedUserString);
                } catch (parseError) {
                    console.error("Error parsing user from localStorage:", parseError);
                    logout();
                }
            } else if (localStorage.getItem('token')) {
                 console.warn("Token exists, but no user data in localStorage. Logging out.");
                 logout();
            }

            if (isMounted) {
                setUser(userToSet); // Устанавливаем пользователя (может быть null)

                // Если пользователь успешно загружен/установлен, обновляем лимиты для ТЕКУЩЕГО года
                if (userToSet) {
                    const currentSystemYear = new Date().getFullYear();
                    // Проверяем, были ли лимиты для текущего года уже загружены (из localStorage)
                    if (!userToSet.vacationLimits?.[currentSystemYear]?.loaded) {
                         await refreshUserVacationLimits(currentSystemYear);
                    } else {
                         console.log(`Limits for year ${currentSystemYear} already loaded.`);
                    }
                }
                setLoading(false);
            }

        } else {
             // Не аутентифицирован
             if (isMounted) {
                 setUser(null);
                 setLoading(false);
             }
        }
    };
    fetchUser();


    return () => { isMounted = false; };
  }, [refreshUserVacationLimits]); // Зависимость остается, но функция стабильна

  // Отображение глобального загрузчика
  // ИСПРАВЛЕНО: Используем loading И limitsLoading для более точного отображения
  // Но пока оставим только loading для первоначальной загрузки
  if (loading) {
    return <Loader />;
  }

  return (
    <ThemeProvider>
      {/* Передаем user, setUser, refreshUserVacationLimits и limitsLoading */}
      <UserProvider value={{ user, setUser, refreshUserVacationLimits, limitsLoading }}>
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
                          {/* Новый маршрут для управления подразделениями (только админ) */}
                          <Route
                            path="/admin/units"
                            element={
                               user?.isAdmin ? <DepartmentManagementPage /> : <Navigate to="/dashboard" replace />
                            }
                          />
                          {/* Новый маршрут для управления пользователями (только админ) */}
                          <Route
                            path="/admin/users"
                            element={
                              user?.isAdmin ? <UserManagementPage /> : <Navigate to="/dashboard" replace />
                            }
                          />
                          {/* Новый маршрут для экспорта отпусков (только админ) */}
                          <Route
                            path="/admin/export-vacations"
                            element={
                              user?.isAdmin ? <ExportVacationsPage /> : <Navigate to="/dashboard" replace />
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
