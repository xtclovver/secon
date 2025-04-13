import React, { useState, useEffect, useContext } from 'react'; // Добавили useContext
import { motion, AnimatePresence } from 'framer-motion';
import DatePicker, { registerLocale } from 'react-datepicker';
import { toast } from 'react-toastify';
import { FaCalendarAlt, FaPlus, FaTrash, FaSave, FaPaperPlane } from 'react-icons/fa';
import ru from 'date-fns/locale/ru';
import { useUser } from '../../context/UserContext'; // Импортируем хук useUser
import { createVacationRequest } from '../../api/vacations'; // Убран submitVacationRequest
import Loader from '../../components/ui/Loader/Loader'; // Импортируем Loader
import 'react-datepicker/dist/react-datepicker.css';
import './VacationForm.css';

// Регистрация русской локали для DatePicker
registerLocale('ru', ru);

const VacationForm = () => {
  // Используем хук useUser для получения данных и функций из контекста
  const { user, refreshUserVacationLimits, limitsLoading } = useUser();
  const [year, setYear] = useState(new Date().getFullYear() + 1); // По умолчанию следующий год
  const [periods, setPeriods] = useState([{ startDate: null, endDate: null, daysCount: 0 }]);
  // const [status, setStatus] = useState('draft'); // Удалено состояние status
  const [submitting, setSubmitting] = useState(false);
  const [errors, setErrors] = useState({});
  const [formSubmitted, setFormSubmitted] = useState(false); // Новое состояние для индикации успешной отправки
  // const [requestId, setRequestId] = useState(null); // Удалено состояние requestId

  // useEffect для принудительной перезагрузки данных при монтировании или изменении ключевых зависимостей
  useEffect(() => {
    // Цель: Загрузить/обновить лимиты для ТЕКУЩЕГО выбранного года (`year`)
    // когда компонент монтируется ИЛИ когда пользователь/год меняется.
    // Всегда вызываем refreshUserVacationLimits, если пользователь доступен и загрузка не идет.
    if (user && !limitsLoading) {
      console.log(`useEffect[user, year, limitsLoading]: Triggering refreshUserVacationLimits for year ${year}...`);
      refreshUserVacationLimits(year);
    } else if (!user) {
      console.warn("useEffect[user, year, limitsLoading]: Cannot refresh limits: user context is not available.");
    } else if (limitsLoading) {
      console.log(`useEffect[user, year, limitsLoading]: Limits are already loading for year ${year}.`);
    }
    // Зависимости: year (чтобы сработало при смене года),
    // refreshUserVacationLimits (стабильная функция из context).
    // Убрали user и limitsLoading, чтобы избежать бесконечного цикла.
    // Проверка user и limitsLoading выполняется внутри эффекта.
  }, [year, refreshUserVacationLimits]); // Измененный массив зависимостей


  // --- Получение данных о лимитах из контекста ---
  const limitsData = user?.vacationLimits?.[year];
  const limitError = limitsData?.error; // Ошибка загрузки лимита для этого года

  // !!! ЛОГИРОВАНИЕ ДАННЫХ ИЗ КОНТЕКСТА !!!
  console.log(`VacationForm - Selected Year: ${year}`);
  // console.log(`VacationForm - Full User Context:`, user); // Раскомментируйте для полной информации о пользователе
  console.log(`VacationForm - Limits Data for ${year} from context:`, limitsData);
  console.log(`VacationForm - limitsLoading state: ${limitsLoading}`);
  console.log(`VacationForm - Limit Error for ${year}: ${limitError}`);

  // Используем null как индикатор, что лимит не загружен или не существует
  const currentLimit = limitsData?.loaded && !limitError ? limitsData.totalDays : null;
  const currentUsedDays = limitsData?.loaded && !limitError ? limitsData.usedDays : null;
  const currentAvailableDays = limitsData?.loaded && !limitError ? limitsData.availableDays : null;

  // !!! ЛОГИРОВАНИЕ ВЫЧИСЛЕННЫХ ЗНАЧЕНИЙ !!!
  console.log(`VacationForm - Calculated Values: currentLimit=${currentLimit}, currentUsedDays=${currentUsedDays}, currentAvailableDays=${currentAvailableDays}`);


  // Вычисление запрошенных дней
  const totalDaysRequested = periods.reduce((sum, period) => sum + period.daysCount, 0);

  // Вычисление оставшихся дней (только если лимиты загружены и без ошибок)
  const remainingDays = currentAvailableDays !== null
    ? currentAvailableDays - totalDaysRequested
    : null;

  // Функция для подсчета дней между датами
  const calculateDays = (startDate, endDate) => {
    if (!startDate || !endDate || startDate > endDate) return 0; // Добавлена проверка startDate > endDate
    const diffTime = Math.abs(endDate.getTime() - startDate.getTime());
    return Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
  };

  // Обработка изменения года
  const handleYearChange = (e) => {
    const newYear = parseInt(e.target.value);
    setYear(newYear);
    // Сброс формы при смене года
    setPeriods([{ startDate: null, endDate: null, daysCount: 0 }]);
    setErrors({});
    // setRequestId(null); // Сброс ID не нужен
    // setStatus('draft'); // Сброс статуса не нужен
    setFormSubmitted(false); // Сбрасываем флаг отправки при смене года

    // --- ВСЕГДА вызываем обновление лимитов при смене года ---
    if (user && !limitsLoading) { // Проверяем только, что есть user и не идет загрузка
        console.log(`Year changed to ${newYear}. Triggering refreshUserVacationLimits...`);
        refreshUserVacationLimits(newYear); // Игнорируем флаг 'loaded'
    } else if (!user) {
        console.warn("Cannot refresh limits on year change: user context is not available.");
    } else if (limitsLoading) {
        console.log(`Year changed to ${newYear}, but limits are already loading.`);
    }
    // Старая проверка с 'loaded' удалена
  };

  // Обработка изменения дат
  const handleDateChange = (index, field, date) => {
    const newPeriods = [...periods];
    newPeriods[index][field] = date;
    
    // Автоматический расчет дней, если обе даты выбраны и start <= end
    if (newPeriods[index].startDate && newPeriods[index].endDate && newPeriods[index].startDate <= newPeriods[index].endDate) {
      newPeriods[index].daysCount = calculateDays(
        newPeriods[index].startDate,
        newPeriods[index].endDate
      );
    } else {
      newPeriods[index].daysCount = 0; // Сброс, если даты некорректны
    }
    
    setPeriods(newPeriods);
    validateForm(newPeriods); // Валидация при каждом изменении
  };

  // Добавление нового периода
  const addPeriod = () => {
    setPeriods([...periods, { startDate: null, endDate: null, daysCount: 0 }]);
  };

  // Удаление периода
  const removePeriod = (index) => {
    const newPeriods = periods.filter((_, i) => i !== index);
    setPeriods(newPeriods);
    validateForm(newPeriods); // Валидация после удаления
  };

  // Валидация формы (переработанная логика)
  // Убран useCallback
  const validateForm = (periodsToValidate = periods) => { // Снова принимаем аргумент или используем состояние
    // Всегда начинаем с пустого объекта ошибок
    const currentErrors = {};
    // const periodsToValidate = periods; // Используем переданный аргумент или состояние

    // 1. Проверка базовой корректности дат и наличия периодов
    if (periodsToValidate.length === 0) {
      currentErrors.noPeriods = 'Необходимо добавить хотя бы один период отпуска.';
      // Устанавливаем ошибки и выходим, если периодов нет
      setErrors(currentErrors);
      return false;
    }

    const hasInvalidDates = periodsToValidate.some(
      period => !period.startDate || !period.endDate || period.startDate > period.endDate
    );
    if (hasInvalidDates) {
      currentErrors.invalidDates = 'Все даты должны быть заполнены корректно (дата начала <= дата окончания).';
      // Не выходим сразу, чтобы проверить и другие ошибки, если возможно
    }

    // 2. Проверка общих дней и правила 14 дней (только если базовые даты корректны)
    if (!errors.invalidDates && !errors.noDays && !errors.noPeriods) { // Проверяем только если базовые вещи ОК
        const totalRequested = periodsToValidate.reduce((sum, period) => sum + period.daysCount, 0);
        const hasLongPeriod = periodsToValidate.some(period => period.daysCount >= 14);

        if (!hasLongPeriod && totalRequested > 0) { // Добавляем проверку totalRequested > 0
            currentErrors.longPeriod = 'Одна из частей отпуска должна быть не менее 14 календарных дней.';
        }
        // Проверка лимита дней (только если лимиты загружены и без ошибок)
        if (currentAvailableDays !== null) {
            if (totalRequested > currentAvailableDays) {
                currentErrors.limit = `Превышен доступный лимит дней отпуска (${currentAvailableDays} дн.). Запрошено: ${totalRequested} дн.`;
            }
            // Проверка на точное соответствие доступным дням
            if (totalRequested !== currentAvailableDays && totalRequested > 0) {
                currentErrors.exactDays = `Необходимо использовать все доступные дни отпуска (${currentAvailableDays} дн.). Запрошено: ${totalRequested} дн.`;
            }
        } else if (limitsData?.loaded && !limitError && currentLimit === null) { // Явно проверяем, что лимит null после загрузки
             // Лимит был запрошен, но не найден в БД
             currentErrors.limitNotSet = `Лимит отпуска на ${year} год не найден в системе. Пожалуйста, обратитесь к администратору для установки лимита.`;
        } else if (limitError) {
             // Если была ошибка при загрузке лимита
             currentErrors.limitLoadError = `Ошибка при загрузке лимита на ${year} год: ${limitError}`;
        } else if (!limitsData?.loaded && !limitsLoading) {
             // Данные еще не загружены и не в процессе загрузки
             currentErrors.limitNotLoaded = 'Ожидание данных о лимите отпуска...';
        }
        // Если есть ошибка limitError, она будет показана в блоке лимитов, здесь можно не дублировать
    }

    // Логирование и обновление состояния ошибок
    console.log("ValidateForm - Calculated Errors:", JSON.stringify(currentErrors));
    
    // Простое обновление состояния
    setErrors({ ...currentErrors }); 

    // Log *after* setting state (хотя это может не показать обновленное значение немедленно)
    // console.log("Called setErrors with:", JSON.stringify(currentErrors)); // Можно убрать этот лог
    
    // ИСПРАВЛЕНО: Проверяем 'currentErrors', а не 'errors' из предыдущего состояния
    const isValid = Object.keys(currentErrors).length === 0; 
    console.log(`validateForm returning: ${isValid}`); // Log the return value
    return isValid; // Возвращаем результат проверки текущих ошибок
  };
  
  // Убран useEffect для валидации

  // Отправка заявки руководителю (объединяет сохранение и отправку)
  const handleSubmit = async (e) => {
    // Предотвращаем стандартное поведение формы, если вызвано через onSubmit
    if (e && typeof e.preventDefault === 'function') {
        e.preventDefault();
    }

    // Вызываем validateForm() непосредственно перед отправкой для актуальной проверки
    if (!validateForm()) {
        console.log("Submit blocked by validateForm() check.");
        toast.warn('Пожалуйста, исправьте ошибки в форме перед отправкой');
        return;
    }

    console.log("Validation passed in handleSubmit, proceeding.");
    setSubmitting(true);
    setFormSubmitted(false); // Сбрасываем флаг перед попыткой
    try {
      // Создаем заявку сразу со статусом "На рассмотрении" (ID 2)
      const vacationRequest = {
        year,
        periods: periods.map(period => ({
          start_date: period.startDate.toISOString(),
          end_date: period.endDate.toISOString(),
          days_count: period.daysCount
        })),
        // Ключ statusId удален, бэкенд теперь сам ставит StatusPending по умолчанию
        // statusId: 2 // На рассмотрении
      };

      // Вызываем только createVacationRequest
      await createVacationRequest(vacationRequest);
      // ID заявки больше не хранится в состоянии
      toast.success('Заявка успешно создана и отправлена на рассмотрение');
      setFormSubmitted(true); // Устанавливаем флаг успешной отправки

      // !!! ОБНОВЛЯЕМ ЛИМИТЫ ПОСЛЕ УСПЕШНОЙ ОТПРАВКИ !!!
      console.log("Request created and submitted successfully. Refreshing limits...");
      await refreshUserVacationLimits(year); // Вызываем обновление лимитов для текущего года формы

    } catch (error) {
      toast.error(error.message || 'Ошибка при сохранении или отправке заявки');
      // Не сбрасываем submitting в случае ошибки, чтобы пользователь мог попробовать еще раз?
      // Или сбрасываем, чтобы кнопка стала активной? Решаем сбросить.
      // setSubmitting(false); // Убираем сброс из catch, он будет в finally
    } finally {
      // Блок finally выполняется всегда, поэтому просто сбрасываем submitting здесь
      setSubmitting(false);
    }
  };

  return (
    <motion.div
      className="vacation-form-container"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
    >
      <h2>Оформление отпуска на {year} год</h2>
      
      <div className="year-selector">
        {/* Убираем disabled={loading} */}
        <select id="year" value={year} onChange={handleYearChange} disabled={submitting}>
          {[...Array(4)].map((_, i) => { // Генерируем 4 года: текущий + 3 следующих
            const currentYear = new Date().getFullYear();
            const y = currentYear + i; // Начинаем с текущего года и добавляем смещение
            return <option key={y} value={y}>{y}</option>;
          })}
        </select>
        {/* Убираем индикатор загрузки лимита */}
      </div>
      {/* --- Блок отображения лимитов --- */}
      <motion.div
        className="vacation-limits card"
        initial={{ x: -20, opacity: 0 }}
        animate={{ x: 0, opacity: 1 }}
        transition={{ delay: 0.2 }}
      >
        {limitsLoading ? (
           <div className="loading-container"> <Loader size="small" /> Загрузка лимитов...</div>
        ) : limitError ? (
           <div className="error-message">{limitError}</div>
        ) : currentLimit !== null ? (
          <>
            <div className="limit-item">
              <span>Лимит на {year}:</span>
              <span className="limit-value">{currentLimit}</span>
            </div>
            <div className="limit-item">
              <span>Использовано:</span>
              <span className="limit-value">{currentUsedDays ?? 'N/A'}</span>
            </div>
            <div className="limit-item">
              <span>Доступно:</span>
              <span className="limit-value">{currentAvailableDays ?? 'N/A'}</span>
            </div>
            {/* <hr /> был здесь */}
            <div className="limit-item">
              <span>Запрошено в этой заявке:</span>
              <span className="limit-value">{totalDaysRequested}</span>
            </div>
            {/* Блок "Останется после заявки" удален полностью */}
          </>
        ) : (
          <div className="info-message">Лимит отпуска на {year} год не найден или не установлен.</div>
        )}
      </motion.div>
      
      {/* --- Отображение ошибок валидации формы --- */}
      {Object.values(errors).map((error, index) => (
        error && <div key={index} className="error-message">{error}</div>
      ))}
      {/* Убираем явные проверки, используем map */}
      
      {/* Форма теперь вызывает handleSubmit при отправке */}
      <form onSubmit={handleSubmit}> 
        {/* <AnimatePresence> */} {/* Временно убираем AnimatePresence */}
          {periods.map((period, index) => (
            <motion.div
              key={index}
              className="vacation-period card" // Добавлен класс card
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              transition={{ delay: index * 0.1 }}
            >
              <h3>Часть отпуска {index + 1}</h3>
              
              <div className="date-inputs">
                <div className="date-field">
                  <label htmlFor={`start-date-${index}`}>Дата начала</label>
                  <div className="date-picker-container">
                    <DatePicker
                      id={`start-date-${index}`}
                      selected={period.startDate}
                      onChange={(date) => handleDateChange(index, 'startDate', date)}
                      selectsStart
                      startDate={period.startDate}
                      endDate={period.endDate}
                      minDate={new Date(year, 0, 1)}
                      maxDate={new Date(year, 11, 31)}
                      dateFormat="dd.MM.yyyy"
                      locale="ru"
                      placeholderText="Выберите дату"
                      className="date-input"
                      disabled={submitting || formSubmitted} // Используем formSubmitted
                    />
                    <FaCalendarAlt className="date-icon" />
                  </div>
                </div>
                
                <div className="date-field">
                  <label htmlFor={`end-date-${index}`}>Дата окончания</label>
                  <div className="date-picker-container">
                    <DatePicker
                      id={`end-date-${index}`}
                      selected={period.endDate}
                      onChange={(date) => handleDateChange(index, 'endDate', date)}
                      selectsEnd
                      startDate={period.startDate}
                      endDate={period.endDate}
                      minDate={period.startDate || new Date(year, 0, 1)}
                      maxDate={new Date(year, 11, 31)}
                      dateFormat="dd.MM.yyyy"
                      locale="ru"
                      placeholderText="Выберите дату"
                      className="date-input"
                      disabled={submitting || formSubmitted || !period.startDate} // Используем formSubmitted
                    />
                    <FaCalendarAlt className="date-icon" />
                  </div>
                </div>
              </div>
              
              <div className="days-count">
                Количество дней: <strong>{period.daysCount}</strong>
              </div>
              
              {periods.length > 1 && (
                <motion.button
                  type="button"
                  className="remove-period btn btn-danger" // Добавлены классы btn
                  onClick={() => removePeriod(index)}
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  disabled={submitting || formSubmitted} // Используем formSubmitted
                >
                  <FaTrash /> Удалить часть
                </motion.button>
              )}
            </motion.div>
          ))}
        {/* </AnimatePresence> */} {/* Временно убираем AnimatePresence */}
        
        <div className="form-actions">
          <motion.button
            type="button"
            className="add-period btn" // Добавлен класс btn
            onClick={addPeriod}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            disabled={submitting || formSubmitted} // Используем formSubmitted
          >
            <FaPlus /> Добавить часть отпуска
          </motion.button>
          
          {/* Кнопка "Сохранить черновик" удалена */}
          
          {/* Кнопка "Отправить" теперь основная, тип submit */}
          <motion.button
            type="submit" // Изменен тип на submit
            className="submit-request btn btn-success" // Оставляем классы
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            disabled={
              limitsLoading || // Блокируем во время загрузки лимитов
              currentAvailableDays === null || // Блокируем, если лимит не загружен/не найден
              submitting ||
              formSubmitted || // Используем formSubmitted
              Object.values(errors).some(e => e) // Блокируем, если есть хотя бы одна ошибка
            }
          >
            {submitting ? <Loader size="small" inline /> : <FaPaperPlane />}
            {submitting ? ' Отправка...' : ' Отправить на рассмотрение'}
          </motion.button>
        </div>
      </form>
      
      {/* Используем formSubmitted для отображения сообщения об успехе */}
      {formSubmitted && (
          <motion.div
            className="success-message" // Нужен CSS для этого класса
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
          >
            Заявка успешно отправлена!
          </motion.div>
      )}
    </motion.div>
  );
};

export default VacationForm;
