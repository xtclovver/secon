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
      generateXLSX(vacationData, selectedYear); // Передаем год в функцию генерации

    } catch (error) {
      console.error("Ошибка экспорта отпусков:", error);
      toast.error(`Произошла ошибка во время экспорта: ${error.message}`);
    } finally {
      setIsExporting(false);
    }
  };

  // Функция для генерации XLSX файла
  const generateXLSX = (data, year) => { // Принимаем год как аргумент
    console.log("Генерация XLSX с данными:", data);
    if (!data || data.length === 0) {
      toast.warn("Нет данных для генерации XLSX файла.");
      return;
    }

    // --- Стили ---
    const thinBorder = { style: "thin", color: { rgb: "000000" } };
    const allBorders = { top: thinBorder, bottom: thinBorder, left: thinBorder, right: thinBorder };
    const topBorderOnly = { top: thinBorder };

    const styles = {
        header: {
            font: { bold: true },
            alignment: { horizontal: 'center', vertical: 'center', wrapText: true },
            border: allBorders
        },
        center: {
            alignment: { horizontal: 'center', vertical: 'center' },
            border: allBorders
        },
        dataCell: {
            border: allBorders,
            alignment: { vertical: 'center', wrapText: true }
        },
        dataCellCenter: {
            border: allBorders,
            alignment: { horizontal: 'center', vertical: 'center', wrapText: true }
        },
        dataCellLeft: {
             border: allBorders,
            alignment: { horizontal: 'left', vertical: 'center', wrapText: true }
        },
        rightAlign: {
            alignment: { horizontal: 'right', vertical: 'center' },
            border: allBorders
        },
        leftAlign: {
            alignment: { horizontal: 'left', vertical: 'center' },
            border: allBorders
        },
        topHeaderCenter: {
            font: { bold: false },
            alignment: { horizontal: 'center', vertical: 'center' },
            border: allBorders
        },
        signature: {
            font: { sz: 8 },
            alignment: { horizontal: 'center', vertical: 'top' },
            border: topBorderOnly
        },
        mainTitle: {
            font: { bold: true, sz: 12 },
            alignment: { horizontal: 'center', vertical: 'center' }
        },
        simpleText: { // Стиль для текста без границ
             alignment: { horizontal: 'left', vertical: 'center' }
        },
         simpleTextRight: { // Стиль для текста без границ, выровненный вправо
             alignment: { horizontal: 'right', vertical: 'center' }
        },
         simpleTextCenter: { // Стиль для текста без границ, выровненный по центру
             alignment: { horizontal: 'center', vertical: 'center' }
        }
    };

    // --- Данные ---
    const worksheetData = [];
    const merges = [];

    // --- Шапка Т-7 ---
    worksheetData.push(['', '', '', '', '', '', '', '', '', '', 'Форма по ОКУД', '', '0301020']);
    merges.push({ s: { r: 0, c: 10 }, e: { r: 0, c: 11 } });
    worksheetData.push(['Наименование организации', '', '', '', '', '', '', '', '', '', 'по ОКПО', '', '']); // TODO: Заполнить
    merges.push({ s: { r: 1, c: 0 }, e: { r: 1, c: 9 } });
    merges.push({ s: { r: 1, c: 10 }, e: { r: 1, c: 11 } });
    worksheetData.push([]); // Пустая строка
    worksheetData.push(['', '', '', '', '', '', '', '', 'УТВЕРЖДАЮ', '', '', '', '']);
    merges.push({ s: { r: 3, c: 8 }, e: { r: 3, c: 12 } });
    worksheetData.push(['', '', '', '', '', '', '', '', 'Руководитель', '', '', 'Дата составления', '']); // TODO: Заполнить
    merges.push({ s: { r: 4, c: 8 }, e: { r: 4, c: 9 } });
    merges.push({ s: { r: 4, c: 11 }, e: { r: 4, c: 12 } });
    worksheetData.push(['', '', '', '', '', '', '', '', '(должность)', '', '(личная подпись)', '', '(расшифровка подписи)']);
    merges.push({ s: { r: 5, c: 8 }, e: { r: 5, c: 9 } });
    merges.push({ s: { r: 5, c: 10 }, e: { r: 5, c: 11 } });
    worksheetData.push([]); // Пустая строка
    worksheetData.push(['', '', '', '', 'ГРАФИК ОТПУСКОВ', '', '', '', '', '', 'Номер документа', '', `на ${year} год`]);
    merges.push({ s: { r: 7, c: 4 }, e: { r: 7, c: 9 } });
    merges.push({ s: { r: 7, c: 10 }, e: { r: 7, c: 11 } });
    worksheetData.push([]); // Пустая строка

    // --- Заголовки таблицы ---
    const headerRow1 = [
        '№ п/п', 'Структурное подразделение', 'Должность (специальность, профессия) по штатному расписанию',
        'Фамилия, имя, отчество', 'Табельный номер', 'Количество календарных дней ежегодного отпуска', null, null,
        'ОТПУСК', null, 'перенесение отпуска', null, 'Примечание'
    ];
    const headerRow2 = [
        null, null, null, null, null, 'основного', 'дополнительного', 'итого',
        'дата запланированная', 'дата фактическая', 'основание (документ)', 'дата предполагаемого отпуска', null
    ];
    const headerRow3 = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10', '11', '12', '13'];

    worksheetData.push(headerRow1);
    worksheetData.push(headerRow2);
    worksheetData.push(headerRow3);

    // Объединения для заголовков таблицы
    merges.push({ s: { r: 9, c: 0 }, e: { r: 10, c: 0 } }); // № п/п
    merges.push({ s: { r: 9, c: 1 }, e: { r: 10, c: 1 } }); // Структурное подразделение
    merges.push({ s: { r: 9, c: 2 }, e: { r: 10, c: 2 } }); // Должность
    merges.push({ s: { r: 9, c: 3 }, e: { r: 10, c: 3 } }); // ФИО
    merges.push({ s: { r: 9, c: 4 }, e: { r: 10, c: 4 } }); // Табельный номер
    merges.push({ s: { r: 9, c: 5 }, e: { r: 9, c: 7 } });  // Количество дней
    merges.push({ s: { r: 9, c: 8 }, e: { r: 9, c: 9 } });  // ОТПУСК
    merges.push({ s: { r: 9, c: 10 }, e: { r: 9, c: 11 } }); // перенесение отпуска
    merges.push({ s: { r: 9, c: 12 }, e: { r: 10, c: 12 } }); // Примечание

    // --- Данные отпусков ---
    const dataStartIndex = worksheetData.length;
    data.forEach((row, index) => {
        worksheetData.push([
            index + 1,
            row.unit_name || '',
            row.position_name || '',
            row.full_name || '',
            row.employee_number || '',
            row.planned_days_main || 0,
            row.planned_days_additional || 0,
            row.planned_days_total || 0,
            row.planned_date ? new Date(row.planned_date).toLocaleDateString('ru-RU') : '',
            row.actual_date ? new Date(row.actual_date).toLocaleDateString('ru-RU') : '',
            row.transfer_reason || '',
            row.transfer_date ? new Date(row.transfer_date).toLocaleDateString('ru-RU') : '',
            row.note || ''
        ]);
    });

    // --- Создание листа ---
    const worksheet = XLSX.utils.aoa_to_sheet(worksheetData);
    worksheet['!merges'] = merges;

    // --- Применение стилей ---
    // Функция для применения стиля к ячейке
    const applyStyle = (r, c, style) => {
        const cellRef = XLSX.utils.encode_cell({ r, c });
        worksheet[cellRef] = worksheet[cellRef] || {}; // Создаем ячейку, если ее нет
        worksheet[cellRef].s = style;
    };

    // Стили шапки
    applyStyle(0, 10, styles.simpleText); // "Форма по ОКУД"
    applyStyle(0, 12, styles.simpleTextRight); // Код ОКУД
    applyStyle(1, 0, styles.simpleText); // "Наименование организации"
    applyStyle(1, 10, styles.simpleText); // "по ОКПО"
    applyStyle(1, 12, styles.simpleTextRight); // Код ОКПО
    applyStyle(3, 8, styles.topHeaderCenter); // "УТВЕРЖДАЮ"
    applyStyle(4, 8, styles.simpleTextCenter); // "Руководитель"
    applyStyle(4, 11, styles.simpleTextCenter); // "Дата составления"
    applyStyle(5, 8, styles.signature); // (должность)
    applyStyle(5, 10, styles.signature); // (личная подпись)
    applyStyle(5, 12, styles.signature); // (расшифровка подписи)
    applyStyle(7, 4, styles.mainTitle); // "ГРАФИК ОТПУСКОВ"
    applyStyle(7, 10, styles.simpleText); // "Номер документа"
    applyStyle(7, 12, styles.simpleText); // "на ... год"

    // Стили заголовков таблицы
    for (let R = 9; R <= 11; ++R) {
        for (let C = 0; C <= 12; ++C) {
            // Пропускаем пустые ячейки во второй строке заголовка, чтобы не переопределять стиль объединенных
            if (R === 10 && C < 5 || R === 10 && C > 11) continue;
            applyStyle(R, C, (R === 11) ? styles.center : styles.header);
        }
    }

    // Стили данных
    for (let R = dataStartIndex; R < worksheetData.length; ++R) {
        for (let C = 0; C <= 12; ++C) {
            let style = styles.dataCellLeft; // По умолчанию выравнивание влево
            if (C === 0 || (C >= 4 && C <= 7)) { // №, Таб. номер, Дни
                style = styles.dataCellCenter;
            } else if (C === 8 || C === 9 || C === 11) { // Даты
                 style = styles.dataCellCenter;
            }
            applyStyle(R, C, style);
        }
    }

    // --- Ширина колонок ---
    worksheet['!cols'] = [
      { wch: 5 },  // A (№ п/п)
      { wch: 30 }, // B (Структурное подразделение)
      { wch: 40 }, // C (Должность)
      { wch: 35 }, // D (ФИО)
      { wch: 15 }, // E (Табельный номер)
      { wch: 10 }, // F (Дни осн)
      { wch: 10 }, // G (Дни доп)
      { wch: 10 }, // H (Итого)
      { wch: 15 }, // I (Дата план)
      { wch: 15 }, // J (Дата факт)
      { wch: 30 }, // K (Перенос осн)
      { wch: 15 }, // L (Перенос дата)
      { wch: 30 }  // M (Примечание)
    ];

    // --- Создание и скачивание книги ---
    const workbook = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(workbook, worksheet, "График отпусков Т-7");
    XLSX.writeFile(workbook, `grafik_otpuskov_${year}_T-7.xlsx`); // Добавляем год в имя файла
    toast.success(`Экспорт ${data.length} записей за ${year} год завершен.`);
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
