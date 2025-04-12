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
  DRAFT: 1,
  PENDING: 2,
  APPROVED: 3,
  REJECTED: 4,
  CANCELLED: 5,
};

// Карта статусов для отображения
const STATUS_MAP = {
  [STATUS.DRAFT]: { text: 'Черновик', class: 'status-draft', Icon: FaEdit },
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

  // Обработчики событий
  const handleYearChange = (e) => setYear(parseInt(e.target.value));
  const handleStatusChange = (e) => setStatusFilter(e.target.value);

  const handleAction = async (actionFunc, id, successMsg, errorMsgPrefix) => {
      setLoading(true);
      try {
          await actionFunc(id);
          toast.success(successMsg);
          // Обновляем список заявок И лимиты пользователя
           await fetchVacations(year, statusFilter);
           if (refreshUserVacationLimits) { // Вызываем обновление лимитов, если функция доступна
              console.log(`Action successful, refreshing limits for year: ${year}`); // Лог
              await refreshUserVacationLimits(year); // <-- ИСПРАВЛЕНО: Передаем год
           }
       } catch (err) {
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

  // Обработчик утверждения (вызывает handleAction)
  const handleApprove = (id) => {
    if (window.confirm('Вы уверены, что хотите утвердить эту заявку?')) {
        // Передаем approveVacationRequest в handleAction
        handleAction(approveVacationRequest, id, 'Заявка успешно утверждена', 'Ошибка утверждения');
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
            {[...Array(5)].map((_, i) => {
              const y = new Date().getFullYear() + 2 - i;
              return <option key={y} value={y}>{y}</option>;
            })}
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
                      {vacation.periods.map((p, index) => (
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
                      {/* Пользователь может отменить свой Черновик или На рассмотрении */}
                      {user && vacation.userId === user.id && (vacation.statusId === STATUS.DRAFT || vacation.statusId === STATUS.PENDING) && (
                        <button onClick={() => handleCancel(vacation.id)} className="btn btn-sm btn-warning" title={vacation.statusId === STATUS.DRAFT ? "Удалить черновик" : "Отменить заявку"} disabled={loading}>
                          {vacation.statusId === STATUS.DRAFT ? <FaTrashAlt /> : <FaBan />}
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
                       {! (user && vacation.userId === user.id && (vacation.statusId === STATUS.DRAFT || vacation.statusId === STATUS.PENDING)) &&
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
