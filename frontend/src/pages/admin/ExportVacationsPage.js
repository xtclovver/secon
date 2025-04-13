import React, { useState, useEffect } from 'react';
import { toast } from 'react-toastify';
import { getOrganizationalUnitTree } from '../../api/units';
import { exportVacationsByUnits } from '../../api/vacations';
import './ExportVacationsPage.css';
import * as XLSX from 'xlsx';

// Вспомогательная функция для преобразования дерева юнитов в плоский список
const flattenUnitTree = (nodes) => {
  let flatList = [];
  nodes.forEach(node => {
    // Добавляем сам узел (если это не корень без ID или если нужно включить все уровни)
    // Проверяем, что у узла есть ID и имя
    if (node.id && node.name) {
       flatList.push({ id: node.id, name: node.name, unit_type: node.unit_type }); // Добавляем тип юнита
    }
    // Рекурсивно обходим дочерние узлы
    if (node.children && node.children.length > 0) {
      flatList = flatList.concat(flattenUnitTree(node.children));
    }
  });
  return flatList;
};

const ExportVacationsPage = () => {
  const currentYear = new Date().getFullYear();
  const [departments, setDepartments] = useState([]);
  const [selectedDepartments, setSelectedDepartments] = useState([]);
  const [selectedYear, setSelectedYear] = useState(currentYear); // Состояние для выбранного года
  const [isLoading, setIsLoading] = useState(false);
  const [isExporting, setIsExporting] = useState(false);

  // Загрузка дерева юнитов
  useEffect(() => {
    const fetchDepartments = async () => {
      setIsLoading(true);
      try {
        const treeData = await getOrganizationalUnitTree(); // Получаем дерево
        const flatList = flattenUnitTree(treeData || []); // Преобразуем в плоский список
        setDepartments(flatList);
      } catch (error) {
        console.error("Ошибка загрузки отделов:", error);
        toast.error("Не удалось загрузить список отделов.");
        setDepartments([]); // Установить пустой массив в случае ошибки
      } finally {
        setIsLoading(false);
      }
    };
    fetchDepartments();
  }, []);

  const handleCheckboxChange = (event) => {
    const departmentId = parseInt(event.target.value, 10);
    const isChecked = event.target.checked;

    setSelectedDepartments(prevSelected => {
      if (isChecked) {
        return [...prevSelected, departmentId];
      } else {
        return prevSelected.filter(id => id !== departmentId);
      }
    });
  };

  const handleSelectAll = (event) => {
    if (event.target.checked) {
      setSelectedDepartments(departments.map(dep => dep.id));
    } else {
      setSelectedDepartments([]);
    }
  };

  // Обработчик изменения года
  const handleYearChange = (event) => {
    const year = parseInt(event.target.value, 10);
    if (!isNaN(year)) {
      setSelectedYear(year);
    }
  };

  const handleExport = async () => {
    if (selectedDepartments.length === 0) {
      toast.warn("Пожалуйста, выберите хотя бы один отдел для экспорта.");
      return;
    }
    if (!selectedYear || selectedYear < 2000 || selectedYear > 2100) {
        toast.warn("Пожалуйста, введите корректный год для экспорта.");
        return;
    }

    setIsExporting(true);
    toast.info(`Начинаем экспорт за ${selectedYear} год...`);

    try {
      // 1. Вызвать API бэкенда, передавая выбранные отделы и ГОД
      const vacationData = await exportVacationsByUnits(selectedDepartments, selectedYear);
      console.log(`Выбранные отделы для экспорта (${selectedYear}):`, selectedDepartments);
      console.log(`Получены данные для экспорта (${selectedYear}):`, vacationData);

      // 2. Сформировать XLSX файл на основе полученных данных и шаблона
      generateXLSX(vacationData); // Передаем реальные данные

    } catch (error) {
      console.error("Ошибка экспорта отпусков:", error);
      toast.error(`Произошла ошибка во время экспорта: ${error.message}`);
    } finally {
      setIsExporting(false);
    }
  };

  // Функция для генерации XLSX файла
  const generateXLSX = (data) => {
    console.log("Генерация XLSX с данными:", data);
    if (!data || data.length === 0) {
      toast.warn("Нет данных для генерации XLSX файла.");
      return;
    }

    // --- Подготовка данных ---
    // Заголовки таблицы Т-7
    const headers = [
      "№ п/п", // 1
      "Структурное подразделение", // 2
      "Должность (специальность, профессия) по штатному расписанию", // 3
      "Фамилия, имя, отчество", // 4
      "Табельный номер", // 5
      "Количество календарных дней ежегодного отпуска - основной", // 6
      "Количество календарных дней ежегодного отпуска - дополнительный", // 7
      "Количество календарных дней ежегодного отпуска - итого", // 8
      "Дата отпуска - запланированная", // 9
      "Дата отпуска - фактическая", // 10
      "Перенесение отпуска - основание (документ)", // 11
      "Перенесение отпуска - дата предполагаемого отпуска", // 12
      "Примечание" // 13
    ];

    // Преобразование данных API в массив массивов для XLSX
    const sheetData = data.map((row, index) => [
      index + 1, // № п/п
      row.unit_name || '',
      row.position_name || '',
      row.full_name || '',
      row.employee_number || '',
      row.planned_days_main || 0,
      row.planned_days_additional || 0,
      row.planned_days_total || 0,
      row.planned_date ? new Date(row.planned_date).toLocaleDateString('ru-RU') : '', // Форматируем дату
      row.actual_date ? new Date(row.actual_date).toLocaleDateString('ru-RU') : '', // Форматируем дату
      row.transfer_reason || '',
      row.transfer_date ? new Date(row.transfer_date).toLocaleDateString('ru-RU') : '', // Форматируем дату
      row.note || ''
    ]);

    // Добавляем заголовки в начало данных
    sheetData.unshift(headers);

    // --- Создание XLSX ---
    const worksheet = XLSX.utils.aoa_to_sheet(sheetData);

    // TODO: Добавить статические заголовки формы Т-7 (УТВЕРЖДАЮ, шапка и т.д.)
    // Это потребует более сложной работы с ячейками worksheet['!merges'], worksheet['A1'] = { v: '...', t: 's', ... }
    // Пока экспортируем только таблицу с данными.

    const workbook = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(workbook, worksheet, "График отпусков Т-7");

    // Генерация и скачивание файла
    XLSX.writeFile(workbook, "grafik_otpuskov_T-7.xlsx");
    toast.success(`Экспорт ${data.length} записей завершен.`);
  };


  return (
    <div className="export-vacations-page">
      <h2>Экспорт графика отпусков (Форма Т-7)</h2>

      <div className="year-selection">
        <label htmlFor="export-year">Год для экспорта:</label>
        <input
          type="number"
          id="export-year"
          value={selectedYear}
          onChange={handleYearChange}
          min="2000"
          max="2100"
          className="year-input"
        />
      </div>

      {isLoading ? (
        <p>Загрузка отделов...</p>
      ) : departments.length > 0 ? (
        <div className="department-selection">
          <h3>Выберите отделы для экспорта:</h3>
          <div className="select-all-container">
             <input
               type="checkbox"
               id="select-all"
               onChange={handleSelectAll}
               checked={selectedDepartments.length === departments.length && departments.length > 0}
               disabled={departments.length === 0}
             />
             <label htmlFor="select-all">Выбрать все</label>
          </div>
          <div className="department-list">
            {departments.map(dep => (
              <div key={dep.id} className="department-item">
                <input
                  type="checkbox"
                  id={`dep-${dep.id}`}
                  value={dep.id}
                  checked={selectedDepartments.includes(dep.id)}
                  onChange={handleCheckboxChange}
                />
                <label htmlFor={`dep-${dep.id}`}>{dep.name}</label>
              </div>
            ))}
          </div>
        </div>
      ) : (
         <p>Не удалось загрузить отделы или отделы отсутствуют.</p>
      )}

      <button
        onClick={handleExport}
        disabled={isLoading || isExporting || selectedDepartments.length === 0 || !selectedYear}
        className="export-button"
      >
        {isExporting ? `Экспорт (${selectedYear})...` : `Экспортировать за ${selectedYear} год`}
      </button>
    </div>
  );
};

export default ExportVacationsPage; // Убедимся, что экспорт по умолчанию присутствует
