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
    const bottomBorderOnly = { bottom: thinBorder }; // Добавлен стиль для нижней границы

    const styles = {
        header: { // Стиль для заголовков таблицы (жирный, центр, границы, перенос)
            font: { bold: true },
            alignment: { horizontal: 'center', vertical: 'center', wrapText: true },
            border: allBorders
        },
        center: { // Стиль для центрированных ячеек с границами (например, цифры 1-13)
            alignment: { horizontal: 'center', vertical: 'center' },
            border: allBorders
        },
        dataCellLeft: { // Стиль для данных, выровненных влево (текст)
             border: allBorders,
            alignment: { horizontal: 'left', vertical: 'center', wrapText: true }
        },
        dataCellCenter: { // Стиль для данных, выровненных по центру (даты, числа)
            border: allBorders,
            alignment: { horizontal: 'center', vertical: 'center', wrapText: true }
        },
        mainTitle: { // Стиль для "ГРАФИК ОТПУСКОВ"
            font: { bold: true, sz: 12 },
            alignment: { horizontal: 'center', vertical: 'center' }
            // Без границ
        },
        simpleTextRight: { // Стиль для текста с границами, выровненный вправо
             alignment: { horizontal: 'right', vertical: 'center' },
             border: allBorders
        },
        simpleTextCenter: { // Стиль для текста с границами, выровненный по центру
             alignment: { horizontal: 'center', vertical: 'center' },
             border: allBorders
        },
        // Стили для текста без границ
        noBorderText: {
             alignment: { horizontal: 'left', vertical: 'center', wrapText: true }
        },
        noBorderTextRight: {
             alignment: { horizontal: 'right', vertical: 'center' }
        },
        noBorderTextCenter: {
             alignment: { horizontal: 'center', vertical: 'center' }
        },
        // Стиль для подписи руководителя (подчеркивание снизу)
        signatureLine: {
            alignment: { horizontal: 'center', vertical: 'bottom' },
            border: bottomBorderOnly // Только нижняя граница
        },
        // Стиль для текста постановления (мелкий, правый, без границ)
        decreeText: {
            font: { sz: 8 },
            alignment: { horizontal: 'right', vertical: 'top', wrapText: true }
        }
    };

    // --- Данные ---
    const worksheetData = [];
    const merges = [];
    let currentRowIndex = 0; // Начинаем с 0

    // --- Шапка Т-7 ---
    // Строка 0: Постановление Госкомстата
    worksheetData.push(['', '', '', '', '', '', '', '', '', 'Унифицированная форма № Т-7']);
    merges.push({ s: { r: currentRowIndex, c: 9 }, e: { r: currentRowIndex, c: 12 } });
    currentRowIndex++; // 1

    // Строка 1: Постановление Госкомстата (продолжение)
    worksheetData.push(['', '', '', '', '', '', '', '', '', 'Утверждена постановлением Госкомстата']);
    merges.push({ s: { r: currentRowIndex, c: 9 }, e: { r: currentRowIndex, c: 12 } });
    currentRowIndex++; // 2

    // Строка 2: Постановление Госкомстата (окончание)
    worksheetData.push(['', '', '', '', '', '', '', '', '', 'России от 06.04.01 № 26']);
    merges.push({ s: { r: currentRowIndex, c: 9 }, e: { r: currentRowIndex, c: 12 } });
    currentRowIndex++; // 3
    worksheetData.push([]); currentRowIndex++; // 4 Пустая строка

    // Строка 5: Коды ОКУД/ОКПО
    worksheetData.push(['Наименование организации', '', '', '', '', '', '', '', '', '', 'Код']); // TODO: Заполнить организацию
    merges.push({ s: { r: currentRowIndex, c: 0 }, e: { r: currentRowIndex, c: 9 } });
    currentRowIndex++; // 6
    worksheetData.push(['', '', '', '', '', '', '', '', 'Форма по ОКУД', '', '', '0301020']);
    merges.push({ s: { r: currentRowIndex, c: 8 }, e: { r: currentRowIndex, c: 10 } });
    currentRowIndex++; // 7
    worksheetData.push(['', '', '', '', '', '', '', '', 'по ОКПО', '', '', '']); // TODO: Заполнить ОКПО
    merges.push({ s: { r: currentRowIndex, c: 8 }, e: { r: currentRowIndex, c: 10 } });
    currentRowIndex++; // 8
    worksheetData.push([]); currentRowIndex++; // 9 Пустая строка

    // Строка 10: УТВЕРЖДАЮ
    worksheetData.push(['', '', '', '', '', '', '', '', 'УТВЕРЖДАЮ']);
    merges.push({ s: { r: currentRowIndex, c: 8 }, e: { r: currentRowIndex, c: 12 } });
    currentRowIndex++; // 11

    // Строка 11-13: Должность утверждающего
    // TODO: Эти данные должны приходить извне или быть константами
    const approvingTitleLines = [
        'Заместитель Генерального директора',
        'ПАО ГК "ТНС энерго" - Управляющий',
        'директор ООО "ТНС энерго Пенза"'
    ];
    approvingTitleLines.forEach(line => {
        worksheetData.push(['', '', '', '', '', '', '', '', line]);
        merges.push({ s: { r: currentRowIndex, c: 8 }, e: { r: currentRowIndex, c: 12 } });
        currentRowIndex++;
    }); // Стало 14

    // Строка 14: Пустая строка перед подписью
    worksheetData.push([]); currentRowIndex++; // 15

    // Строка 15: Мнение профсоюза и Подпись
    // TODO: Данные профсоюза и ФИО должны приходить извне или быть константами
    const unionOpinionDate = 'от "29" ноября 2024 г. № 88 учтено';
    const approverName = 'Р.Б.Чернов';
    worksheetData.push(['Мнение выборного профсоюзного органа', '', '', '', '', '', '', '', '', '', '', approverName]);
    merges.push({ s: { r: currentRowIndex, c: 0 }, e: { r: currentRowIndex + 1, c: 5 } }); // Объединение для текста профсоюза
    merges.push({ s: { r: currentRowIndex, c: 8 }, e: { r: currentRowIndex, c: 10 } }); // Место для подписи (линия будет применена стилем)
    merges.push({ s: { r: currentRowIndex, c: 11 }, e: { r: currentRowIndex, c: 12 } }); // ФИО
    currentRowIndex++; // 16

    // Строка 16: Мнение профсоюза (продолжение) и Дата утверждения
    worksheetData.push([unionOpinionDate, '', '', '', '', '', '', '', '"__" декабря', '', `20${year.toString().slice(-2)} г.`]);
    merges.push({ s: { r: currentRowIndex, c: 8 }, e: { r: currentRowIndex, c: 9 } }); // Дата месяц
    merges.push({ s: { r: currentRowIndex, c: 10 }, e: { r: currentRowIndex, c: 12 } }); // Год
    currentRowIndex++; // 17
    worksheetData.push([]); currentRowIndex++; // 18 Пустая строка

    // Строка 19: Номер документа, Дата составления, Год (Заголовки)
    worksheetData.push(['', '', '', '', '', '', '', '', 'Номер документа', '', 'Дата составления', '', 'Год']);
    merges.push({ s: { r: currentRowIndex, c: 8 }, e: { r: currentRowIndex, c: 9 } });
    merges.push({ s: { r: currentRowIndex, c: 10 }, e: { r: currentRowIndex, c: 11 } });
    currentRowIndex++; // 20

    // Строка 20: Номер документа, Дата составления, Год (Значения)
    // TODO: Номер и дату составления брать извне или использовать константы/текущую дату
    const documentNumber = '1';
    const creationDate = new Date().toLocaleDateString('ru-RU'); // Пример: текущая дата
    worksheetData.push(['', '', '', '', '', '', '', '', documentNumber, '', creationDate, '', year]);
    merges.push({ s: { r: currentRowIndex, c: 8 }, e: { r: currentRowIndex, c: 9 } });
    merges.push({ s: { r: currentRowIndex, c: 10 }, e: { r: currentRowIndex, c: 11 } });
    currentRowIndex++; // 21
    worksheetData.push([]); currentRowIndex++; // 22 Пустая строка

    // Строка 23: ГРАФИК ОТПУСКОВ
    worksheetData.push(['', '', '', '', 'ГРАФИК ОТПУСКОВ']);
    merges.push({ s: { r: currentRowIndex, c: 4 }, e: { r: currentRowIndex, c: 9 } });
    currentRowIndex++; // 24
    worksheetData.push([]); currentRowIndex++; // 25 Пустая строка

    // --- Заголовки таблицы ---
    const tableHeaderStartIndex = currentRowIndex; // Индекс начала заголовков таблицы
    const headerRow1 = [
        '№ п/п', 'Структурное подразделение', 'Должность (специальность, профессия) по штатному расписанию',
        'Фамилия, имя, отчество', 'Табельный номер', 'ОТПУСК', null, null, null, null, null, null, 'Примечание'
    ];
    const headerRow2 = [
        null, null, null, null, null, 'Количество календарных дней ежегодного отпуска', null, null,
        'дата', null, 'перенесение отпуска', null, null
    ];
    const headerRow3 = [
        null, null, null, null, null, 'основного', 'дополнительного', 'итого',
        'запланированная', 'фактическая', 'основание', 'дата предполагаемого отпуска', null
    ];
    const headerRow4 = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10', '11', '12', '13'];

    worksheetData.push(headerRow1);
    worksheetData.push(headerRow2);
    worksheetData.push(headerRow3);
    worksheetData.push(headerRow4);
    currentRowIndex += 4; // Обновляем индекс текущей строки

    // Объединения для заголовков таблицы
    const r1 = tableHeaderStartIndex; // Индекс первой строки заголовка
    merges.push({ s: { r: r1, c: 0 }, e: { r: r1 + 2, c: 0 } }); // № п/п
    merges.push({ s: { r: r1, c: 1 }, e: { r: r1 + 2, c: 1 } }); // Структурное подразделение
    merges.push({ s: { r: r1, c: 2 }, e: { r: r1 + 2, c: 2 } }); // Должность
    merges.push({ s: { r: r1, c: 3 }, e: { r: r1 + 2, c: 3 } }); // ФИО
    merges.push({ s: { r: r1, c: 4 }, e: { r: r1 + 2, c: 4 } }); // Табельный номер
    merges.push({ s: { r: r1, c: 5 }, e: { r: r1, c: 11 } });  // ОТПУСК (верхний уровень)
    merges.push({ s: { r: r1 + 1, c: 5 }, e: { r: r1 + 1, c: 7 } }); // Количество дней (средний уровень)
    merges.push({ s: { r: r1 + 1, c: 8 }, e: { r: r1 + 1, c: 9 } }); // дата (средний уровень)
    merges.push({ s: { r: r1 + 1, c: 10 }, e: { r: r1 + 1, c: 11 } }); // перенесение отпуска (средний уровень)
    merges.push({ s: { r: r1, c: 12 }, e: { r: r1 + 2, c: 12 } }); // Примечание

    // --- Данные отпусков ---
    const dataStartIndex = currentRowIndex; // Используем обновленный индекс
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
        worksheet[cellRef] = worksheet[cellRef] || { t: 's', v: '' }; // Создаем ячейку, если ее нет, тип 's' (string)
        // Если ячейка существует, но не имеет типа, устанавливаем 's'
        if (!worksheet[cellRef].t) {
            worksheet[cellRef].t = 's';
        }
        // Применяем стиль
        worksheet[cellRef].s = style;
    };

    // Стили для шапки документа (строки 0-24)
    // Постановление (строки 0-2)
    applyStyle(0, 9, styles.decreeText);
    applyStyle(1, 9, styles.decreeText);
    applyStyle(2, 9, styles.decreeText);
    // Организация и Код (строка 5)
    applyStyle(5, 0, styles.noBorderText); // Наименование организации
    applyStyle(5, 10, styles.simpleTextCenter); // "Код"
    // Коды ОКУД/ОКПО (строки 6-7)
    applyStyle(6, 8, styles.simpleTextRight); // "Форма по ОКУД"
    applyStyle(6, 11, styles.center); // Код ОКУД (значение)
    applyStyle(7, 8, styles.simpleTextRight); // "по ОКПО"
    applyStyle(7, 11, styles.center); // Код ОКПО (значение)
    // Утверждение (строки 10-13)
    applyStyle(10, 8, styles.noBorderTextCenter); // "УТВЕРЖДАЮ"
    applyStyle(11, 8, styles.noBorderTextCenter); // Должность 1
    applyStyle(12, 8, styles.noBorderTextCenter); // Должность 2
    applyStyle(13, 8, styles.noBorderTextCenter); // Должность 3
    // Профсоюз и подпись (строки 15-16)
    applyStyle(15, 0, styles.noBorderText); // Мнение профсоюза (начало)
    applyStyle(16, 0, styles.noBorderText); // Мнение профсоюза (конец)
    applyStyle(15, 8, styles.signatureLine); // Линия подписи
    applyStyle(15, 11, styles.noBorderTextCenter); // ФИО
    applyStyle(16, 8, styles.noBorderTextCenter); // Дата утверждения (день, месяц)
    applyStyle(16, 10, styles.noBorderTextCenter); // Дата утверждения (год)
    // Информация о документе (строки 19-20)
    applyStyle(19, 8, styles.center); // "Номер документа" (заголовок)
    applyStyle(19, 10, styles.center); // "Дата составления" (заголовок)
    applyStyle(19, 12, styles.center); // "Год" (заголовок)
    applyStyle(20, 8, styles.center); // Номер документа (значение)
    applyStyle(20, 10, styles.center); // Дата составления (значение)
    applyStyle(20, 12, styles.center); // Год (значение)
    // Заголовок графика (строка 23)
    applyStyle(23, 4, styles.mainTitle); // "ГРАФИК ОТПУСКОВ"

    // Стили заголовков таблицы (начиная с tableHeaderStartIndex)
    for (let R = tableHeaderStartIndex; R < tableHeaderStartIndex + 4; ++R) {
        for (let C = 0; C <= 12; ++C) {
            // Пропускаем ячейки, которые будут перекрыты объединением
             const isMerged = merges.some(m => R >= m.s.r && R <= m.e.r && C >= m.s.c && C <= m.e.c && !(R === m.s.r && C === m.s.c));
             if (isMerged) continue;

            // Применяем стиль заголовка (жирный, центрированный, с границами)
            // Для последней строки (цифры) используем обычный центрированный стиль
            applyStyle(R, C, (R === tableHeaderStartIndex + 3) ? styles.center : styles.header);
        }
    }

    // Стили данных (начиная с dataStartIndex)
    for (let R = dataStartIndex; R < worksheetData.length; ++R) {
        for (let C = 0; C <= 12; ++C) {
            let style = styles.dataCellLeft; // По умолчанию выравнивание влево
            // Добавлены скобки для явного указания порядка операций (no-mixed-operators)
            if ((C === 0) || (C >= 4 && C <= 7)) { // №, Таб. номер, Дни
                style = styles.dataCellCenter;
            } else if (C === 8 || C === 9 || C === 11) { // Даты
                 style = styles.dataCellCenter;
            }
            // Применяем стиль к существующей ячейке данных
             const cellRef = XLSX.utils.encode_cell({ r: R, c: C });
             if (worksheet[cellRef]) { // Убедимся, что ячейка существует
                 worksheet[cellRef].s = style;
             }
        }
    }

    // --- Ширина колонок ---
    // (Оставляем как было, можно скорректировать при необходимости)
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
