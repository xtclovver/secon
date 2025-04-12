import React, { useState, useEffect, useContext } from 'react'; // Добавили useContext
import { motion, AnimatePresence } from 'framer-motion';
import DatePicker, { registerLocale } from 'react-datepicker';
import { toast } from 'react-toastify';
import { FaCalendarAlt, FaPlus, FaTrash, FaSave, FaPaperPlane } from 'react-icons/fa';
import ru from 'date-fns/locale/ru';
import { UserContext } from '../../context/UserContext'; // Импортируем UserContext
// Убираем getVacationLimit из импортов API, т.к. он больше не нужен здесь
import { createVacationRequest, submitVacationRequest } from '../../api/vacations';
import 'react-datepicker/dist/react-datepicker.css';
import './VacationForm.css';

// Регистрация русской локали для DatePicker
registerLocale('ru', ru);

const VacationForm = () => {
  const { user } = useContext(UserContext); // Получаем пользователя из контекста
  const [year, setYear] = useState(new Date().getFullYear() + 1);
  const [periods, setPeriods] = useState([{ startDate: null, endDate: null, daysCount: 0 }]);
  const [status, setStatus] = useState('draft');
  // Убираем локальные состояния limit и usedDays
  // const [limit, setLimit] = useState(28);
  // const [usedDays, setUsedDays] = useState(0);
  // const [loading, setLoading] = useState(false); // Больше не загружаем лимит здесь
  const [submitting, setSubmitting] = useState(false);
  const [errors, setErrors] = useState({});
  const [requestId, setRequestId] = useState(null);

  // Получаем актуальные лимиты из контекста пользователя
  // Используем значения для ТЕКУЩЕГО ВЫБРАННОГО года (year), если они есть в контексте
  // Если нет, используем дефолтные значения
  const limitsForSelectedYear = user?.vacationLimits?.[year];
  const currentLimit = limitsForSelectedYear?.totalDays ?? 28; // Дефолт 28, если нет данных
  const currentUsedDays = limitsForSelectedYear?.usedDays ?? 0; // Дефолт 0

  // Убираем useEffect для загрузки лимита
  // useEffect(() => { ... fetchLimit ... }, [year]);

  // Вычисление общего количества дней (используем currentLimit и currentUsedDays из контекста)
  const totalDaysRequested = periods.reduce((sum, period) => sum + period.daysCount, 0);
  const remainingDays = currentLimit - currentUsedDays - totalDaysRequested; // Используем актуальные данные

  // Функция для подсчета дней между датами
  const calculateDays = (startDate, endDate) => {
    if (!startDate || !endDate || startDate > endDate) return 0; // Добавлена проверка startDate > endDate
    const diffTime = Math.abs(endDate.getTime() - startDate.getTime());
    return Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
  };

  // Обработка изменения года
  const handleYearChange = (e) => {
    setYear(parseInt(e.target.value));
    // Сброс периодов и ошибок при смене года
    setPeriods([{ startDate: null, endDate: null, daysCount: 0 }]);
    setErrors({});
    setRequestId(null);
    setStatus('draft');
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
        // Используем currentLimit и currentUsedDays из контекста для проверки лимита
        if (totalRequested > (currentLimit - currentUsedDays)) {
            currentErrors.limit = `Превышен доступный лимит дней отпуска (${currentLimit - currentUsedDays} дн.).`;
        }
    }
    
    // Log the errors object *before* setting state
    console.log("Errors object calculated:", JSON.stringify(currentErrors));
    
    // Простое обновление состояния
    setErrors({ ...currentErrors }); 

    // Log *after* setting state (хотя это может не показать обновленное значение немедленно)
    // console.log("Called setErrors with:", JSON.stringify(currentErrors)); // Можно убрать этот лог
    
    const isValid = Object.keys(currentErrors).length === 0;
    console.log(`validateForm returning: ${isValid}`); // Log the return value
    return isValid; // Проверяем собранный объект
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
    try {
      let currentRequestId = requestId;

      // 1. Создаем/обновляем черновик, если необходимо
      // (Предполагаем, что API createVacationRequest может обрабатывать обновление,
      // или всегда создает новую запись, что тоже приемлемо для упрощения)
      const vacationRequest = {
        year,
        periods: periods.map(period => ({
          // Отправляем дату в полном формате ISO 8601 (RFC3339), который Go понимает по умолчанию
          // Ключи должны соответствовать JSON-тегам в Go модели (snake_case)
          start_date: period.startDate.toISOString(), 
          end_date: period.endDate.toISOString(),
          days_count: period.daysCount 
        })),
        statusId: 1 // Черновик (или статус для отправки, если API требует)
      };

      // Всегда вызываем create, чтобы получить актуальный ID перед отправкой
      // (или используем update, если API позволяет и requestId есть)
      const createResponse = await createVacationRequest(vacationRequest);
      currentRequestId = createResponse.id; // Получаем ID созданной/обновленной заявки
      setRequestId(currentRequestId); // Обновляем состояние ID
      toast.info('Черновик сохранен/обновлен.'); // Информируем пользователя

      // ADD DELAY before submitting
      await new Promise(resolve => setTimeout(resolve, 200)); // Add 200ms delay

      // 2. Отправляем заявку с полученным ID
      await submitVacationRequest(currentRequestId);
      toast.success('Заявка успешно отправлена руководителю');
      setStatus('submitted'); // Меняем статус в UI

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
          <option value={new Date().getFullYear()}>Текущий год</option>
          <option value={new Date().getFullYear() + 1}>Следующий год</option>
        </select>
        {/* Убираем индикатор загрузки лимита */}
      </div>
      
      <motion.div 
        className="vacation-limits card" // Добавлен класс card
        initial={{ x: -20, opacity: 0 }}
        animate={{ x: 0, opacity: 1 }}
        transition={{ delay: 0.2 }}
      >
        <div className="limit-item">
          <span>Доступно дней:</span>
          {/* Используем currentLimit */}
          <span className="limit-value">{currentLimit}</span>
        </div>
        <div className="limit-item">
          <span>Использовано:</span>
           {/* Используем currentUsedDays */}
          <span className="limit-value">{currentUsedDays}</span>
        </div>
        <div className="limit-item">
          <span>Запрошено:</span>
          <span className="limit-value">{totalDaysRequested}</span>
        </div>
        <div className="limit-item">
          <span>Осталось:</span>
           {/* remainingDays теперь рассчитывается на основе данных контекста */}
          <span className={`limit-value ${remainingDays < 0 ? 'error' : ''}`}>
            {remainingDays}
          </span>
        </div>
      </motion.div>
      
      {/* Отображение ошибок (изменено на явные проверки) */}
      {errors.invalidDates && <div className="error-message">{errors.invalidDates}</div>}
      {errors.noPeriods && <div className="error-message">{errors.noPeriods}</div>}
      {errors.noDays && <div className="error-message">{errors.noDays}</div>}
      {/* Упрощенное отображение ошибки без анимации */}
      {/* Убран console.log из рендера */}
      {errors.longPeriod ? <div className="error-message">{errors.longPeriod}</div> : null} 
      {errors.limit && <div className="error-message">{errors.limit}</div>}
      
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
                      disabled={submitting || status === 'submitted'}
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
                      disabled={submitting || status === 'submitted' || !period.startDate} // Блокируем, пока не выбрана дата начала
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
                  disabled={submitting || status === 'submitted'}
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
            disabled={submitting || status === 'submitted'}
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
              submitting || 
              status === 'submitted' || 
              Object.keys(errors).length > 0 // Блокируем при ошибках валидации
            } 
          >
            <FaPaperPlane /> Отправить руководителю
          </motion.button>
        </div>
      </form>
      
      {status === 'submitted' && (
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
