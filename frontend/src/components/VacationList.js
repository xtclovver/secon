import React from 'react';
import {
  List,
  ListItem,
  ListItemText,
  Typography,
  Chip,
  Box,
  Divider,
  Fade, // Для анимации появления
  ListItemIcon,
  Tooltip,
  IconButton, // Для кнопки удаления
  ButtonGroup, // Для группировки кнопок Approve/Reject
  Button,      // Для кнопок Approve/Reject
  CircularProgress, // Для индикации загрузки при действии
} from '@mui/material';
import EventIcon from '@mui/icons-material/Event';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import CancelIcon from '@mui/icons-material/Cancel';
import HourglassEmptyIcon from '@mui/icons-material/HourglassEmpty';
import SendIcon from '@mui/icons-material/Send';
import DeleteIcon from '@mui/icons-material/Delete';
import CheckIcon from '@mui/icons-material/Check'; // Для Approve
import CloseIcon from '@mui/icons-material/Close'; // Для Reject
import PersonIcon from '@mui/icons-material/Person'; // Для имени пользователя

import { updateVacationStatus, deleteVacation } from '../services/api'; // Импорт API функций
import { useState } from 'react'; // Для управления состоянием загрузки кнопок

// Функция для форматирования даты
const formatDate = (dateString) => {
  if (!dateString) return '';
  const date = new Date(dateString);
  // Проверяем, валидна ли дата
  if (isNaN(date.getTime())) {
      return 'Invalid Date';
  }
  return date.toLocaleDateString('ru-RU', { // Русский формат даты
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
};

// Функция для получения цвета и иконки статуса
const getStatusChipProps = (status) => {
  switch (status) {
    case 'approved':
      return { color: 'success', icon: <CheckCircleIcon />, label: 'Одобрено' };
    case 'rejected':
      return { color: 'error', icon: <CancelIcon />, label: 'Отклонено' };
    case 'submitted':
      return { color: 'info', icon: <SendIcon />, label: 'На рассмотрении' };
    case 'pending':
    default:
      return { color: 'warning', icon: <HourglassEmptyIcon />, label: 'Черновик' };
  }
};

// Компонент списка заявок
const VacationList = ({ requests, isAdminView, onUpdateRequest, onDeleteRequest }) => {
  const [loadingAction, setLoadingAction] = useState({}); // Состояние загрузки для каждой кнопки { [requestId]: { action: 'approve' | 'reject' | 'delete' } }

  const handleAction = async (action, requestId, status = null) => {
    setLoadingAction({ [requestId]: { action } }); // Устанавливаем лоадер для конкретной кнопки
    try {
      if (action === 'delete') {
        await deleteVacation(requestId);
        if (onDeleteRequest) onDeleteRequest(); // Вызываем колбэк для обновления списка
      } else if (action === 'approve' || action === 'reject') {
        await updateVacationStatus(requestId, status);
        if (onUpdateRequest) onUpdateRequest(); // Вызываем колбэк для обновления списка
      }
    } catch (error) {
      console.error(`Failed to ${action} request ${requestId}:`, error);
      // TODO: Показать сообщение об ошибке пользователю (например, через Snackbar)
      alert(`Ошибка при выполнении действия: ${error.response?.data?.error || error.message}`);
    } finally {
      setLoadingAction({}); // Сбрасываем лоадер
    }
  };


  if (!requests || requests.length === 0) {
    return (
      <Typography variant="subtitle1" align="center" color="textSecondary">
        {isAdminView ? 'Нет заявок для отображения.' : 'У вас пока нет заявок на отпуск.'}
      </Typography>
    );
  }

  return (
    <List>
      {requests.map((request, index) => (
        // Используем Fade для плавной анимации появления каждого элемента
        <Fade in={true} key={request.id} timeout={300 + index * 50}>
          <Box> {/* Используем Box вместо div для лучшей интеграции с MUI */}
            <ListItem
              alignItems="flex-start"
              secondaryAction={
                isAdminView && ( // Показываем кнопки только для админа
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                    {loadingAction[request.id] ? (
                      <CircularProgress size={24} />
                    ) : (
                      <>
                        {/* Кнопки Approve/Reject показываем только для pending/submitted */}
                        {(request.status === 'pending' || request.status === 'submitted') && (
                          <ButtonGroup size="small" variant="outlined" aria-label="approve reject button group">
                            <Tooltip title="Approve">
                              <Button
                                color="success"
                                onClick={() => handleAction('approve', request.id, 'approved')}
                                disabled={!!loadingAction[request.id]}
                              >
                                <CheckIcon fontSize="small" />
                              </Button>
                            </Tooltip>
                            <Tooltip title="Reject">
                              <Button
                                color="error"
                                onClick={() => handleAction('reject', request.id, 'rejected')}
                                disabled={!!loadingAction[request.id]}
                              >
                                <CloseIcon fontSize="small" />
                              </Button>
                            </Tooltip>
                          </ButtonGroup>
                        )}
                        <Tooltip title="Delete">
                          {/* Отдельная кнопка удаления */}
                          <IconButton
                            edge="end"
                            aria-label="delete"
                            color="error"
                            onClick={() => handleAction('delete', request.id)}
                            disabled={!!loadingAction[request.id]}
                            size="small"
                            sx={{ ml: 1 }} // Небольшой отступ слева
                          >
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </>
                    )}
                  </Box>
                )
              }
            >
              <ListItemIcon sx={{ mt: 1, minWidth: 40 }}>
                {isAdminView ? <PersonIcon color="action" /> : <EventIcon color="action" />}
              </ListItemIcon>
              <ListItemText
                primary={
                  <Typography variant="body1">
                    {/* Показываем имя пользователя для админа */}
                    {isAdminView && request.User && (
                      <Typography component="span" variant="subtitle2" color="text.secondary" sx={{ mr: 1 }}>
                        {`${request.User.first_name || ''} ${request.User.last_name || ''}:`}
                      </Typography>
                    )}
                    {`С ${formatDate(request.start_date)} по ${formatDate(request.end_date)}`}
                  </Typography>
                }
                secondary={
                  <Typography variant="body2" color="textSecondary">
                    {`Создано: ${formatDate(request.created_at)}`}
                  </Typography>
                }
              />
              {/* Статус показываем всегда */}
              <Box sx={{ ml: 2, mt: 1, mr: isAdminView ? 15 : 0 /* Доп. отступ справа для админа из-за кнопок */ }}>
                <Tooltip title={getStatusChipProps(request.status).label} arrow>
                  <Chip
                    {...getStatusChipProps(request.status)}
                    size="small"
                    variant="outlined"
                  />
                </Tooltip>
              </Box>
            </ListItem>
            {index < requests.length - 1 && <Divider variant="inset" component="li" />}
          </Box>
        </Fade>
      ))}
    </List>
  );
};

export default VacationList;
