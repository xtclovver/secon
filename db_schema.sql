DROP DATABASE IF EXISTS vacation_scheduler;
CREATE DATABASE IF NOT EXISTS vacation_scheduler;
USE vacation_scheduler;

-- Таблица групп должностей
CREATE TABLE position_groups (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    sort_order INT DEFAULT 0 -- Для сортировки групп
);

-- Таблица должностей
CREATE TABLE positions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    group_id INT NOT NULL,
    FOREIGN KEY (group_id) REFERENCES position_groups(id) ON DELETE CASCADE
);

-- Таблица пользователей
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    department_id INT,
    position_id INT, -- Добавлено поле для должности
    is_admin BOOLEAN DEFAULT FALSE,
    is_manager BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE SET NULL -- Добавлен внешний ключ
);

-- Таблица подразделений
CREATE TABLE departments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    manager_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (manager_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Добавление внешнего ключа в таблицу пользователей
ALTER TABLE users
ADD FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL;

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

-- Заполнение таблицы групп должностей
INSERT INTO position_groups (id, name, sort_order) VALUES
(1, 'Руководители высшего звена', 1),
(2, 'Руководители подразделений и их заместители', 2),
(3, 'Главные и ведущие специалисты', 3),
(4, 'Специалисты (включая инженеров, экономистов, бухгалтеров)', 4),
(5, 'Инспекторы', 5),
(6, 'Технический и административный персонал', 6);

-- Заполнение таблицы должностей
INSERT INTO positions (name, group_id) VALUES
-- Руководители высшего звена
('Первый заместитель генерального директора', 1),
('Заместитель генерального директора по правовым вопросам', 1),
('Заместитель генерального директора по реализации', 1),
-- Руководители подразделений и их заместители
('Директор департамента', 2),
('Заместитель главного бухгалтера по бухгалтерскому учету и отчетности-начальник отдела', 2),
('Заместитель главного бухгалтера по налоговому учету и отчетности и учету МСФО-начальник отдела', 2),
('Заместитель начальника отдела', 2),
('Заместитель начальника управления', 2),
('Заместитель руководителя представительства', 2),
('Начальник отдела', 2),
('Начальник офиса', 2),
('Начальник сектора', 2),
('Начальник управления', 2),
('Руководитель представительства', 2),
('Руководитель управления', 2),
-- Главные и ведущие специалисты
('Ведущий бухгалтер', 3),
('Ведущий инженер', 3),
('Ведущий инженер-программист', 3),
('Ведущий специалист', 3),
('Ведущий экономист', 3),
('Главный бухгалтер', 3),
('Главный специалист', 3),
-- Специалисты
('Бухгалтер', 4),
('Бухгалтер I категории', 4),
('Бухгалтер II категории', 4),
('Инженер', 4),
('Инженер I категории', 4),
('Инженер II категории', 4),
('Инженер электросвязи', 4),
('Инженер-программист III категории', 4),
('Специалист', 4),
('Специалист I категории', 4),
('Специалист II категории', 4),
('Специалист по связям с общественностью и СМИ', 4),
('Экономист', 4),
('Экономист I категории', 4),
('Экономист II категории', 4),
-- Инспекторы
('Инспектор энергоинспекции', 5),
-- Технический и административный персонал
('Делопроизводитель', 6),
('Документовед', 6),
('Заведующий складом', 6),
('Кассир', 6),
('Референт', 6),
('Секретарь руководителя', 6),
('Техник', 6),
('Техник II категории', 6),
('Электромонтер связи 5 разряда', 6);

-- Добавление пользователей по умолчанию
-- admin:admin (Начальник отдела, is_admin=true)
INSERT INTO users (username, password, full_name, email, position_id, is_admin, is_manager) VALUES (
    'admin',
    '$2a$10$0o8J93t0x.QGq4syvzMPnuqwf4vM2UbTbqk7NfN4XNp/F.pvGTw4a', -- Хеш для 'admin'
    'Admin User',
    'admin@example.com',
    (SELECT id FROM positions WHERE name = 'Начальник отдела'),
    TRUE, -- is_admin
    TRUE  -- is_manager (Начальник отдела - руководитель)
);

-- manager:manager (Руководитель управления, is_admin=true)
INSERT INTO users (username, password, full_name, email, position_id, is_admin, is_manager) VALUES (
    'manager',
    '$2a$10$qCTjGMYRcS.bYB/ZGlzOmuqtfcjea74VGyE0en0Qu/6Cr9qo0.hnS', -- Хеш для 'manager'
    'Manager User',
    'manager@example.com',
    (SELECT id FROM positions WHERE name = 'Руководитель управления'),
    TRUE, -- is_admin
    TRUE  -- is_manager (Руководитель управления - руководитель)
);

-- user:user (Специалист I категории, is_admin=false)
INSERT INTO users (username, password, full_name, email, position_id, is_admin, is_manager) VALUES (
    'user',
    '$2a$10$dNyRdXMY4G0gnnjyos7rFOziXUCjqPFVxUROTvwNhZA/440zmtOn6', -- Хеш для 'user'
    'User User',
    'user@example.com',
    (SELECT id FROM positions WHERE name = 'Специалист I категории'),
    FALSE, -- is_admin
    FALSE  -- is_manager
);
