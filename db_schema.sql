DROP DATABASE IF EXISTS vacation_scheduler;
CREATE DATABASE IF NOT EXISTS vacation_scheduler;
USE vacation_scheduler;

-- Таблица должностей (без групп)
CREATE TABLE positions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

-- Таблица пользователей
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    login VARCHAR(50) NOT NULL UNIQUE, -- Изменено с username на login
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    organizational_unit_id INT, -- Переименовано с department_id
    position_id INT,
    is_admin BOOLEAN DEFAULT FALSE,
    is_manager BOOLEAN DEFAULT FALSE, -- Роль менеджера может определяться должностью или привязкой к юниту
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE SET NULL
    -- Внешний ключ для organizational_unit_id будет добавлен после создания таблицы organizational_units
);

-- Новая таблица для иерархии организационной структуры
CREATE TABLE organizational_units (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    -- Возможные типы: 'ROOT', 'DEPARTMENT', 'SUB_DEPARTMENT', 'SECTOR', 'OFFICE', 'REPRESENTATION', 'CENTER', etc.
    unit_type VARCHAR(100) NOT NULL COMMENT 'Тип подразделения (Департамент, Отдел, Сектор, Представительство и т.д.)',
    parent_id INT, -- Ссылка на родительский юнит
    manager_id INT, -- Опционально: руководитель конкретно этого юнита
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES organizational_units(id) ON DELETE SET NULL, -- Ссылка на себя для иерархии
    FOREIGN KEY (manager_id) REFERENCES users(id) ON DELETE SET NULL -- Ссылка на пользователя-руководителя
);

-- Добавление внешнего ключа в таблицу пользователей (после создания organizational_units)
ALTER TABLE users
ADD FOREIGN KEY (organizational_unit_id) REFERENCES organizational_units(id) ON DELETE SET NULL;

-- Таблица лимитов отпусков
CREATE TABLE vacation_limits (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    year INT NOT NULL,
    total_days INT NOT NULL DEFAULT 28,
    used_days INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY (user_id, year)
);

-- Таблица заявок на отпуск
CREATE TABLE vacation_requests (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    year INT NOT NULL,
    status_id INT NOT NULL,
    days_requested INT NOT NULL DEFAULT 0, -- Добавлено поле для хранения запрошенных дней
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Таблица периодов отпуска
CREATE TABLE vacation_periods (
    id INT AUTO_INCREMENT PRIMARY KEY,
    request_id INT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    days_count INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (request_id) REFERENCES vacation_requests(id) ON DELETE CASCADE
);

-- Таблица статусов заявок
CREATE TABLE vacation_status (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT
);

-- Добавление внешнего ключа в таблицу заявок
ALTER TABLE vacation_requests
ADD FOREIGN KEY (status_id) REFERENCES vacation_status(id);

-- Таблица уведомлений
CREATE TABLE notifications (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Заполнение таблицы статусов
INSERT INTO vacation_status (name, description) VALUES
('Черновик', 'Заявка создана, но не отправлена'),
('На рассмотрении', 'Заявка отправлена руководителю'),
('Утверждена', 'Заявка утверждена руководителем'),
('Отклонена', 'Заявка отклонена руководителем'),
('Отменена', 'Заявка отменена сотрудником');

-- Заполнение таблицы organizational_units (Иерархия подразделений)
-- Корневые элементы
INSERT INTO organizational_units (id, name, unit_type, parent_id, manager_id) VALUES
(1, 'Руководство', 'MANAGEMENT', NULL, NULL),
(2, 'Бухгалтерия', 'DEPARTMENT', NULL, NULL),
(3, 'Финансовое управление', 'MANAGEMENT', NULL, NULL),
(4, 'Планово-экономическое управление', 'MANAGEMENT', NULL, NULL),
(5, 'Отдел по управлению персоналом', 'DEPARTMENT', NULL, NULL),
(6, 'Отдел информационной политики и внешних коммуникаций', 'DEPARTMENT', NULL, NULL),
(7, 'Управление систем учета электроэнергии', 'MANAGEMENT', NULL, NULL),
(8, 'Департамент обеспечения деятельности', 'DEPARTMENT', NULL, NULL),
(9, 'Департамент по работе с потребителями', 'DEPARTMENT', NULL, NULL),
(10, 'Департамент управления реализацией', 'DEPARTMENT', NULL, NULL),
(11, 'Центральное представительство', 'REPRESENTATION', NULL, NULL),
(12, 'Пензенское представительство', 'REPRESENTATION', NULL, NULL);

-- Подчиненные Бухгалтерии (parent_id = 2)
INSERT INTO organizational_units (id, name, unit_type, parent_id, manager_id) VALUES
(13, 'Отдел бухгалтерского учета и отчетности', 'SUB_DEPARTMENT', 2, NULL),
(14, 'Сектор учета активов, доходов и затрат общества', 'SECTOR', 13, NULL),
(15, 'Сектор учета заработной платы', 'SECTOR', 13, NULL),
(16, 'Отдел налогового учета и отчетности и учета по МСФО', 'SUB_DEPARTMENT', 2, NULL),
(17, 'Сектор налогового учета и отчетности', 'SECTOR', 16, NULL);

-- Подчиненные Планово-экономическому управлению (parent_id = 4)
INSERT INTO organizational_units (id, name, unit_type, parent_id, manager_id) VALUES
(18, 'Отдел бизнес-планирования и тарифообразования', 'SUB_DEPARTMENT', 4, NULL),
(19, 'Сектор по бизнес-планированию', 'SECTOR', 18, NULL),
(20, 'Сектор по тарифам', 'SECTOR', 18, NULL),
(21, 'Сектор по труду и заработной плате', 'SECTOR', 18, NULL);

-- Подчиненные Отделу по управлению персоналом (parent_id = 5)
INSERT INTO organizational_units (id, name, unit_type, parent_id, manager_id) VALUES
(22, 'Сектор воинского учета и бронирования', 'SECTOR', 5, NULL),
(23, 'Сектор кадрового учета', 'SECTOR', 5, NULL),
(24, 'Сектор подбора и развития персонала', 'SECTOR', 5, NULL);

-- Подчиненные Управлению систем учета электроэнергии (parent_id = 7)
INSERT INTO organizational_units (id, name, unit_type, parent_id, manager_id) VALUES
(25, 'Отдел внедрения интеллектуальных систем учета электроэнергии', 'SUB_DEPARTMENT', 7, NULL),
(26, 'Отдел эксплуатации систем учета электроэнергии', 'SUB_DEPARTMENT', 7, NULL);

-- Подчиненные Департаменту обеспечения деятельности (parent_id = 8)
INSERT INTO organizational_units (id, name, unit_type, parent_id, manager_id) VALUES
(27, 'Сектор по охране труда', 'SECTOR', 8, NULL),
(28, 'Отдел документационного обеспечения', 'SUB_DEPARTMENT', 8, NULL),
(29, 'Отдел материально-технического обеспечения', 'SUB_DEPARTMENT', 8, NULL),
(30, 'Управление информационных технологий', 'SUB_MANAGEMENT', 8, NULL),
(31, 'Отдел поддержки автоматизированных систем', 'SUB_DEPARTMENT', 30, NULL),
(32, 'Отдел эксплуатации инфраструктуры информационных технологий', 'SUB_DEPARTMENT', 30, NULL),
(33, 'Сектор телекоммуникаций', 'SECTOR', 32, NULL),
(34, 'Сектор технической поддержки пользователей', 'SECTOR', 32, NULL);

-- Подчиненные Департаменту по работе с потребителями (parent_id = 9)
INSERT INTO organizational_units (id, name, unit_type, parent_id, manager_id) VALUES
(35, 'Сектор качества обслуживания потребителей', 'SECTOR', 9, NULL),
(36, 'Отдел работы с обращениями потребителей и органов власти', 'SUB_DEPARTMENT', 9, NULL),
(37, 'Управление договорной работы', 'SUB_MANAGEMENT', 9, NULL),
(38, 'Отдел договорной работы с юридическими лицами', 'SUB_DEPARTMENT', 37, NULL);

-- Подчиненные Департаменту управления реализацией (parent_id = 10)
INSERT INTO organizational_units (id, name, unit_type, parent_id, manager_id) VALUES
(39, 'Отдел энергоинспекции', 'SUB_DEPARTMENT', 10, NULL),
(40, 'Управление расчетов и аналитики', 'SUB_MANAGEMENT', 10, NULL),
(41, 'Отдел планирования и аналитики', 'SUB_DEPARTMENT', 40, NULL),
(42, 'Отдел работы с сетевыми организациями', 'SUB_DEPARTMENT', 40, NULL),
(43, 'Отдел расчетов с юридическими лицами', 'SUB_DEPARTMENT', 40, NULL),
(44, 'Управление реализацией физических лиц', 'SUB_MANAGEMENT', 10, NULL),
(45, 'Сектор претензионной работы и работы с судебными приказами', 'SECTOR', 44, NULL),
(46, 'Отдел работы с дебиторской задолженностью и контроля ограничений физических лиц', 'SUB_DEPARTMENT', 44, NULL),
(47, 'Управление реализацией юридических лиц', 'SUB_MANAGEMENT', 10, NULL),
(48, 'Отдел досудебной работы с задолженностью юридических лиц', 'SUB_DEPARTMENT', 47, NULL);

-- Подчиненные Центральному представительству (parent_id = 11)
INSERT INTO organizational_units (id, name, unit_type, parent_id, manager_id) VALUES
(49, 'ЦОК Гагарина', 'CENTER', 11, NULL);

-- Подчиненные Пензенскому представительству (parent_id = 12)
INSERT INTO organizational_units (id, name, unit_type, parent_id, manager_id) VALUES
(50, 'Бессоновский офис', 'OFFICE', 12, NULL),
(51, 'Кондольский ЦОК', 'CENTER', 12, NULL),
(52, 'Лунинский офис', 'OFFICE', 12, NULL),
(53, 'Иссинский ЦОК', 'CENTER', 52, NULL), -- Подчинен Лунинскому офису
(54, 'Мокшанский офис', 'OFFICE', 12, NULL),
(55, 'ЦОК Первомайский', 'CENTER', 12, NULL),
(56, 'Шемышейский офис', 'OFFICE', 12, NULL);


-- Заполнение таблицы должностей (без group_id)
INSERT INTO positions (name) VALUES
('Первый заместитель генерального директора'),
('Заместитель генерального директора по правовым вопросам'),
('Заместитель генерального директора по реализации'),
('Директор департамента'),
('Заместитель главного бухгалтера по бухгалтерскому учету и отчетности-начальник отдела'),
('Заместитель главного бухгалтера по налоговому учету и отчетности и учету МСФО-начальник отдела'),
('Заместитель начальника отдела'),
('Заместитель начальника управления'),
('Заместитель руководителя представительства'),
('Начальник отдела'),
('Начальник офиса'),
('Начальник сектора'),
('Начальник управления'),
('Руководитель представительства'),
('Руководитель управления'),
('Ведущий бухгалтер'),
('Ведущий инженер'),
('Ведущий инженер-программист'),
('Ведущий специалист'),
('Ведущий экономист'),
('Главный бухгалтер'),
('Главный специалист'),
('Бухгалтер'),
('Бухгалтер I категории'),
('Бухгалтер II категории'),
('Инженер'),
('Инженер I категории'),
('Инженер II категории'),
('Инженер электросвязи'),
('Инженер-программист III категории'),
('Специалист'),
('Специалист I категории'),
('Специалист II категории'),
('Специалист по связям с общественностью и СМИ'),
('Экономист'),
('Экономист I категории'),
('Экономист II категории'),
('Инспектор энергоинспекции'),
('Делопроизводитель'),
('Документовед'),
('Заведующий складом'),
('Кассир'),
('Референт'),
('Секретарь руководителя'),
('Техник'),
('Техник II категории'),
('Электромонтер связи 5 разряда');

-- Добавление пользователей по умолчанию
-- admin:admin (Начальник отдела, Отдел информационной политики и внешних коммуникаций, is_admin=true, is_manager=true)
INSERT INTO users (login, password, full_name, organizational_unit_id, position_id, is_admin, is_manager) VALUES (
    'admin',
    '$2a$10$0o8J93t0x.QGq4syvzMPnuqwf4vM2UbTbqk7NfN4XNp/F.pvGTw4a',
    'Админ Админов Админович',
    6, -- Отдел информационной политики и внешних коммуникаций (id=6)
    (SELECT id FROM positions WHERE name = 'Начальник отдела'),
    TRUE,
    TRUE
);

-- manager:manager (Руководитель управления, Планово-экономическое управление, is_admin=true, is_manager=true)
INSERT INTO users (login, password, full_name, organizational_unit_id, position_id, is_admin, is_manager) VALUES (
    'manager',
    '$2a$10$qCTjGMYRcS.bYB/ZGlzOmuqtfcjea74VGyE0en0Qu/6Cr9qo0.hnS',
    'Менеджеров Менеджр Менеджерович',
    4, -- Планово-экономическое управление (id=4)
    (SELECT id FROM positions WHERE name = 'Руководитель управления'),
    TRUE,
    TRUE
);

-- user:user (Специалист I категории, Сектор по бизнес-планированию, is_admin=false, is_manager=false)
INSERT INTO users (login, password, full_name, organizational_unit_id, position_id, is_admin, is_manager) VALUES (
    'user',
    '$2a$10$dNyRdXMY4G0gnnjyos7rFOziXUCjqPFVxUROTvwNhZA/440zmtOn6',
    'Юзер Юзер Юзерович',
    19, -- Сектор по бизнес-планированию (id=19, входит в Планово-экономическое управление)
    (SELECT id FROM positions WHERE name = 'Специалист I категории'),
    FALSE,
    FALSE
);
