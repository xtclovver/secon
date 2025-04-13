import React, { useState, useEffect, useContext } from 'react';
import { motion } from 'framer-motion';
import { toast } from 'react-toastify';
import { FaList, FaFilter, FaSync, FaEdit, FaTrashAlt, FaPaperPlane, FaCheckCircle, FaTimesCircle, FaHourglassHalf, FaBan, FaUserTie } from 'react-icons/fa';
import {
  getMyVacations,
  getAllVacations,
  cancelVacationRequest,
  approveVacationRequest,
  rejectVacationRequest
} from '../../api/vacations';
import Loader from '../../components/ui/Loader/Loader';
import { UserContext } from '../../context/UserContext'; // Импортируем UserContext
// import './VacationsList.css'; // Раскомментируйте, если есть стили

// Константы статусов
const STATUS = {
  // DRAFT: 1, // Удалено
  PENDING: 2,
  APPROVED: 3,
  REJECTED: 4,
  CANCELLED: 5,
};

// Карта статусов для отображения
const STATUS_MAP = {
  // [STATUS.DRAFT]: { text: 'Черновик', class: 'status-draft', Icon: FaEdit }, // Удалено
  [STATUS.PENDING]: { text: 'На рассмотрении', class: 'status-pending', Icon: FaHourglassHalf },
  [STATUS.APPROVED]: { text: 'Утверждена', class: 'status-approved', Icon: FaCheckCircle },
  [STATUS.REJECTED]: { text: 'Отклонена', class: 'status-rejected', Icon: FaTimesCircle },
  [STATUS.CANCELLED]: { text: 'Отменена', class: 'status-cancelled', Icon: FaBan },
};

// Функция для форматирования даты
const formatDate = (dateString) => {
  if (!dateString) return '-';
  try {
    return new Date(dateString).toLocaleDateString('ru-RU');
  } catch (e) {
    console.error("Error formatting date:", dateString, e);
    return dateString; // Возвращаем исходную строку в случае ошибки
  }
};

// Компонент для отображения статуса
const StatusBadge = ({ statusId, statusNameFromData }) => {
  const statusInfo = STATUS_MAP[statusId] || { text: statusNameFromData || 'Неизвестно', class: 'status-unknown', Icon: FaHourglassHalf };
  const { text, class: statusClass, Icon } = statusInfo;

  return (
    <span className={`status-badge ${statusClass}`}>
      <Icon style={{ marginRight: '5px', verticalAlign: 'middle' }} />
      {text}
    </span>
  );
};

const VacationsList = () => {
  // Получаем user и refreshUserVacationLimits из контекста
  const { user, refreshUserVacationLimits } = useContext(UserContext);
  const [vacations, setVacations] = useState([]);
  const [loading, setLoading] = useState(false);
  const [year, setYear] = useState(new Date().getFullYear());
  const [availableYears, setAvailableYears] = useState([]); // Состояние для доступных годов
  const [statusFilter, setStatusFilter] = useState('');
  const [error, setError] = useState(null);

  const isAdmin = user?.isAdmin; // Предполагаем camelCase в контексте
  const isManager = user?.isManager; // Предполагаем camelCase в контексте

  // Функция загрузки заявок
  const fetchVacations = async (selectedYear, selectedStatus) => {
    if (!user) return; // Не загружаем, если нет данных пользователя

    setLoading(true);
    setError(null);
    const currentStatusFilter = selectedStatus === '' ? null : parseInt(selectedStatus);

    try {
      let rawData;
      const filters = { year: selectedYear };
      if (currentStatusFilter !== null) {
        filters.status = currentStatusFilter;
      }

      if (isAdmin || isManager) {
        rawData = await getAllVacations(filters);
      } else {
        rawData = await getMyVacations(selectedYear, currentStatusFilter);
      }

      console.log("Raw Vacations Data from API:", rawData); // Отладка сырых данных

      // Обрабатываем данные для приведения к camelCase и вычисления нужных полей
      const processedData = (rawData || []).map(v => {
        const periods = v.periods || [];
        const totalDays = periods.reduce((sum, p) => sum + (p.days_count || 0), 0); // Используем days_count
        const statusId = v.status_id; // Используем status_id
        const statusInfo = STATUS_MAP[statusId] || { text: v.status_name || 'Неизвестно', Icon: FaHourglassHalf }; // Используем status_name если есть

        return {
          id: v.id,
          userId: v.user_id,
          userFullName: v.user_full_name,
          year: v.year,
          statusId: statusId,
          statusName: statusInfo.text,
          comment: v.comment,
          createdAt: v.created_at,
          periods: periods.map(p => ({ // Приводим периоды к camelCase
            id: p.id,
            requestId: p.request_id,
            startDate: p.start_date,
            endDate: p.end_date,
            daysCount: p.days_count
          })),
          totalDays: totalDays
        };
      });

      console.log("Processed Vacations Data:", processedData); // Отладка обработанных данных
      setVacations(processedData);

    } catch (err) {
      console.error("Error fetching vacations:", err);
      const errorMsg = err.response?.data?.error || err.message || 'Не удалось загрузить список заявок.';
      setError(errorMsg);
      toast.error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  // Загрузка данных при монтировании и смене зависимостей
  useEffect(() => {
    fetchVacations(year, statusFilter);
  }, [year, statusFilter, user]); // Зависим от user, чтобы перезагрузить при смене пользователя

  // Эффект для обновления списка доступных годов
  useEffect(() => {
    const currentSystemYear = new Date().getFullYear();
    const futureYears = Array.from({ length: 6 }, (_, i) => currentSystemYear + i); // Текущий + 5 следующих
    const pastYearsFromData = [...new Set(vacations.map(v => v.year))]; // Уникальные годы из загруженных данных

    // Объединяем все годы, добавляем текущий выбранный год на всякий случай,
    // убираем дубликаты и сортируем по убыванию
    const allYears = [...new Set([year, ...futureYears, ...pastYearsFromData])]
                        .sort((a, b) => b - a); // Сортировка по убыванию

    setAvailableYears(allYears);

  }, [vacations, year]); // Пересчитываем при изменении заявок или выбранного года

  // Обработчики событий
  const handleYearChange = (e) => setYear(parseInt(e.target.value));
  const handleStatusChange = (e) => setStatusFilter(e.target.value);

  // Обновляем handleAction для обработки ответа с warnings
  const handleAction = async (actionFunc, id, successMsg, errorMsgPrefix, isApproveAction = false) => {
      setLoading(true);
      try {
          // Выполняем действие (например, approveVacationRequest)
          const response = await actionFunc(id); // Теперь получаем ответ

          // Проверяем наличие предупреждений (конфликтов) при утверждении
          if (isApproveAction && response && response.warnings && response.warnings.length > 0) {
              // Формируем сообщение о конфликтах
              const conflictMessages = response.warnings.map((warn, index) => (
                 `  ${index + 1}. ${warn.conflicting_user_full_name} (${formatDate(warn.conflicting_start_date)} - ${formatDate(warn.conflicting_end_date)}), Пересечение: ${formatDate(warn.overlap_start_date)} - ${formatDate(warn.overlap_end_date)}`
              )).join('\n');

              // Показываем предупреждение пользователю (можно использовать modal или расширенный toast)
              toast.warn(
                  <div>
                      <p><strong>{successMsg}, НО ОБНАРУЖЕНЫ КОНФЛИКТЫ:</strong></p>
                      <pre style={{ whiteSpace: 'pre-wrap', textAlign: 'left', fontSize: '0.9em' }}>{conflictMessages}</pre>
                  </div>,
                  { autoClose: 15000 } // Увеличиваем время отображения
              );
          } else {
              // Стандартное сообщение об успехе, если нет конфликтов
              toast.success(successMsg);
          }

          // Обновляем список заявок И лимиты пользователя
           await fetchVacations(year, statusFilter);
           if (refreshUserVacationLimits) {
              console.log(`Action successful (${actionFunc.name}), refreshing limits for year: ${year}`); // Лог с именем функции
              await refreshUserVacationLimits(year);
           }
       } catch (err) {
          // Обработка ошибок остается прежней
          const errMsg = err.response?.data?.error || err.message || `Не удалось выполнить действие.`;
          toast.error(`${errorMsgPrefix}: ${errMsg}`);
          // setLoading(false) не нужен здесь, так как fetchVacations/refreshUserVacationLimits его сбросят
      } finally {
          // Убедимся, что loading сбрасывается, даже если refreshUserVacationLimits нет или он упал
           setLoading(false);
      }
  };

  const handleCancel = (id) => {
      if (window.confirm('Вы уверены, что хотите отменить эту заявку?')) {
          handleAction(cancelVacationRequest, id, 'Заявка успешно отменена', 'Ошибка отмены');
      }
  };

  // Обработчик утверждения с проверкой конфликтов
  const handleApprove = async (id) => {
    setLoading(true); // Устанавливаем loading в начале
    try {
      // 1. Первая попытка утверждения без force
      const response = await approveVacationRequest(id, false);
      // Если дошли сюда без ошибки 409, значит конфликтов не было
      toast.success(response.message || 'Заявка успешно утверждена');
      await fetchVacations(year, statusFilter); // Обновляем список
      if (refreshUserVacationLimits) {
        await refreshUserVacationLimits(year); // Обновляем лимиты
      }
    } catch (error) {
      if (error.isConflict) {
        // 2. Обработка ошибки конфликта (409)
        console.warn("Approval conflict detected:", error.conflicts);
        // Формируем сообщение для пользователя
        const conflictMessages = error.conflicts.map((c, index) => (
          `  ${index + 1}. ${c.conflictingUserFullName} (${formatDate(c.overlapStartDate)} - ${formatDate(c.overlapEndDate)})`
        )).join('\n');
        const confirmationMessage = `Внимание! Обнаружены конфликты отпусков с:\n${conflictMessages}\n\nВсе равно утвердить заявку?`;

        // 3. Запрашиваем подтверждение у пользователя
        if (window.confirm(confirmationMessage)) {
          // 4. Повторная попытка с force=true
          try {
            const forceResponse = await approveVacationRequest(id, true);
            toast.success(forceResponse.message || 'Заявка принудительно утверждена');
             // Обновляем список заявок И лимиты пользователя
            await fetchVacations(year, statusFilter);
            if (refreshUserVacationLimits) {
                await refreshUserVacationLimits(year);
            }
          } catch (forceError) {
            // Ошибка при принудительном утверждении
            const errMsg = forceError.response?.data?.error || forceError.message || 'Не удалось принудительно утвердить заявку.';
            toast.error(`Ошибка принудительного утверждения: ${errMsg}`);
          }
        } else {
          // Пользователь отменил утверждение
          toast.info('Утверждение заявки отменено из-за конфликтов.');
        }
      } else {
        // 5. Обработка других ошибок
        const errMsg = error.response?.data?.error || error.message || 'Не удалось утвердить заявку.';
        toast.error(`Ошибка утверждения: ${errMsg}`);
      }
    } finally {
      setLoading(false); // Сбрасываем loading в конце
    }
  };

  // Обработчик отклонения (немного изменен для использования refreshUserVacationLimits)
  const handleReject = async (id) => {
    const reason = prompt('Укажите причину отклонения (необязательно):');
    if (reason !== null) { // Убедимся, что пользователь не нажал "Отмена" в prompt
        if (window.confirm(`Вы уверены, что хотите отклонить эту заявку? ${reason ? `Причина: ${reason}` : ''}`)) {
            setLoading(true);
            try {
                await rejectVacationRequest(id, reason || '');
                toast.success('Заявка успешно отклонена');
                 // Обновляем список заявок И лимиты пользователя
                 await fetchVacations(year, statusFilter);
                 if (refreshUserVacationLimits) {
                    console.log(`Rejection successful, refreshing limits for year: ${year}`); // Лог
                    await refreshUserVacationLimits(year); // <-- ИСПРАВЛЕНО: Передаем год
                 }
             } catch (err) {
                const errMsg = err.response?.data?.error || err.message || `Не удалось отклонить заявку.`;
                toast.error(`Ошибка отклонения: ${errMsg}`);
            } finally {
                 setLoading(false); // Сбрасываем loading в finally
            }
        }
    }
  };

  // Проверка прав на управление (используем camelCase userId)
  const canManageRequest = (vacationOwnerId) => {
      if (!user) return false;
      if (isAdmin) return true;
      if (isManager) return true; // Бэкенд проверит принадлежность к отделу
      return false;
  };

  return (
    <motion.div
      className="vacations-list-container card"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
    >
      <h2><FaList /> {isAdmin || isManager ? 'Заявки на отпуск' : 'Мои заявки на отпуск'}</h2>

      <div className="controls" style={{ marginBottom: '20px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '15px', flexWrap: 'wrap' }}>
        <div className="year-filter" style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
          <label htmlFor="vacation-year">Год:</label>
          <select id="vacation-year" value={year} onChange={handleYearChange} disabled={loading}>
             {/* Генерируем опции на основе динамического списка годов */}
            {availableYears.map(y => (
              <option key={y} value={y}>{y}</option>
            ))}
          </select>
        </div>
         <div className="status-filter" style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <label htmlFor="vacation-status">Статус:</label>
            <select id="vacation-status" value={statusFilter} onChange={handleStatusChange} disabled={loading}>
              <option value="">Все</option>
              {Object.entries(STATUS_MAP).map(([id, { text }]) => (
                <option key={id} value={id}>{text}</option>
              ))}
            </select>
          </div>
        <button onClick={() => fetchVacations(year, statusFilter)} className="btn btn-secondary" disabled={loading}>
          <FaSync className={loading ? 'spin-icon' : ''} /> Обновить
        </button>
      </div>

      {loading && <Loader text="Загрузка заявок..." />}
      
      {error && <div className="error-message" style={{ color: 'var(--danger-color)', marginBottom: '15px' }}>{error}</div>}

      {!loading && !error && vacations.length === 0 && (
        <p style={{ textAlign: 'center', color: 'var(--text-secondary)', marginTop: '30px' }}>
          Заявок на {year} год не найдено.
        </p>
      )}

      {!loading && !error && vacations.length > 0 && (
        <div className="vacation-table-container" style={{ overflowX: 'auto' }}>
          <table className="vacation-table" style={{ width: '100%', borderCollapse: 'collapse', marginTop: '20px' }}>
            <thead>
              <tr>
                <th>ID</th>
                 {(isAdmin || isManager) && <th><FaUserTie /> Сотрудник</th>}
                <th>Год</th>
                <th>Статус</th>
                <th>Периоды</th>
                <th>Дней</th>
                <th>Комментарий</th>
                <th>Создана</th>
                <th style={{ minWidth: '120px' }}>Действия</th>
              </tr>
            </thead>
            <tbody>
              {/* Используем обработанные данные в camelCase */}
              {vacations.map((vacation) => (
                  <motion.tr
                    key={vacation.id}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ duration: 0.3 }}
                  >
                    <td>{vacation.id}</td>
                    {(isAdmin || isManager) && <td>{vacation.userFullName || `User ${vacation.userId || 'N/A'}`}</td>}
                    <td>{vacation.year}</td>
                    <td><StatusBadge statusId={vacation.statusId} statusNameFromData={vacation.statusName} /></td>
                    <td>
                      {vacation.periods?.map((p, index) => ( // Добавлена проверка на vacation.periods
                        <div key={p.id || index} style={{ whiteSpace: 'nowrap' }}>
                          {formatDate(p.startDate)} - {formatDate(p.endDate)} ({p.daysCount || 0} дн.)
                        </div>
                      ))}
                      {vacation.periods.length === 0 && '-'}
                    </td>
                    <td>{vacation.totalDays}</td>
                    <td title={vacation.comment}>{vacation.comment ? `${vacation.comment.substring(0, 30)}${vacation.comment.length > 30 ? '...' : ''}` : '-'}</td>
                    <td style={{ whiteSpace: 'nowrap' }}>{formatDate(vacation.createdAt)}</td>
                    <td className="action-buttons" style={{ whiteSpace: 'nowrap' }}>
                      {/* Пользователь может отменить свою заявку "На рассмотрении" */}
                      {user && vacation.userId === user.id && vacation.statusId === STATUS.PENDING && (
                        <button onClick={() => handleCancel(vacation.id)} className="btn btn-sm btn-warning" title="Отменить заявку" disabled={loading}>
                           <FaBan /> {/* Всегда иконка отмены */}
                        </button>
                      )}
                      {/* Менеджер/Админ могут управлять заявкой "На рассмотрении" */}
                      {canManageRequest(vacation.userId) && vacation.statusId === STATUS.PENDING && (
                        <>
                          <button onClick={() => handleApprove(vacation.id)} className="btn btn-sm btn-success" title="Утвердить" disabled={loading}> <FaCheckCircle /> </button>
                          <button onClick={() => handleReject(vacation.id)} className="btn btn-sm btn-danger" title="Отклонить" disabled={loading}> <FaTimesCircle /> </button>
                        </>
                      )}
                      {/* Менеджер/Админ могут отменить Утвержденную заявку */}
                      {canManageRequest(vacation.userId) && vacation.statusId === STATUS.APPROVED && (
                        <button onClick={() => handleCancel(vacation.id)} className="btn btn-sm btn-danger" title="Отменить утвержденную заявку" disabled={loading}> <FaBan /> </button>
                      )}
                      {/* Отображаем прочерк, если действий нет */}
                       {! (user && vacation.userId === user.id && vacation.statusId === STATUS.PENDING) && // Убрана проверка на DRAFT
                        ! (canManageRequest(vacation.userId) && vacation.statusId === STATUS.PENDING) &&
                        ! (canManageRequest(vacation.userId) && vacation.statusId === STATUS.APPROVED) &&
                        '-'}
                    </td>
                  </motion.tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      
      <style jsx>{`
        .vacation-table { border: 1px solid var(--border-color); margin-top: 20px; width: 100%; border-collapse: collapse; }
        .vacation-table th, .vacation-table td { padding: 10px 12px; border: 1px solid var(--border-color); text-align: left; vertical-align: middle; }
        .vacation-table th { background-color: var(--bg-tertiary); font-weight: 500; white-space: nowrap; }
        .vacation-table tbody tr:hover { background-color: var(--bg-tertiary); }
        .status-badge { display: inline-flex; align-items: center; padding: 4px 8px; border-radius: var(--border-radius-md); font-size: 0.85rem; white-space: nowrap; }
        .status-draft { background-color: #6c757d; color: white; }
        .status-pending { background-color: var(--warning-color); color: #333; }
        .status-approved { background-color: var(--success-color); color: white; }
        .status-rejected { background-color: var(--danger-color); color: white; }
        .status-cancelled { background-color: #adb5bd; color: #333; }
        .action-buttons { display: flex; gap: 5px; flex-wrap: nowrap; }
        .btn-sm { padding: 4px 8px; font-size: 0.9rem; display: inline-flex; align-items: center; justify-content: center; line-height: 1; }
        .btn-sm svg { margin-right: 0 !important; /* Убираем отступ у иконок маленьких кнопок */ }
        .spin-icon { animation: spin 1.5s linear infinite; }
        @keyframes spin { to { transform: rotate(360deg); } }
        .error-message { color: var(--danger-color); margin-bottom: 15px; padding: 10px; background-color: var(--danger-bg-light); border: 1px solid var(--danger-border); border-radius: var(--border-radius); }
      `}</style>

    </motion.div>
  );
};

export default VacationsList;
