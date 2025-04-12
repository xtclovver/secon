-- Создание базы данных (если она еще не существует)
CREATE DATABASE IF NOT EXISTS vacation_db
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

-- Переключение на созданную базу данных
USE vacation_db;

-- Удаление существующих таблиц (для чистого старта)
DROP TABLE IF EXISTS `vacation_requests`;
DROP TABLE IF EXISTS `users`;

-- Создание таблицы пользователей (users)
CREATE TABLE IF NOT EXISTS `users` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `first_name` VARCHAR(255) NOT NULL,
    `last_name` VARCHAR(255) NOT NULL,
    `email` VARCHAR(191) NOT NULL UNIQUE, -- 191 для совместимости с utf8mb4 индексами
    `password` VARCHAR(255) NOT NULL, -- Хранить только хеш пароля
    `is_admin` BOOLEAN NOT NULL DEFAULT FALSE, -- Роль пользователя (администратор или нет)
    `vacation_limit` INT NOT NULL DEFAULT 28, -- Примерное значение по умолчанию
    INDEX `idx_users_email` (`email`) -- Индекс для быстрого поиска по email
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Создание таблицы заявок на отпуск (vacation_requests)
CREATE TABLE IF NOT EXISTS `vacation_requests` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT UNSIGNED NOT NULL,
    `start_date` DATE NOT NULL, -- Используем DATE, так как время обычно не важно для отпуска
    `end_date` DATE NOT NULL,
    `status` ENUM('pending', 'approved', 'rejected', 'submitted') NOT NULL DEFAULT 'pending',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_vacation_requests_user_id` (`user_id`),
    INDEX `idx_vacation_requests_status` (`status`),
    INDEX `idx_vacation_requests_start_date` (`start_date`),
    INDEX `idx_vacation_requests_end_date` (`end_date`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE -- Удалять заявки при удалении пользователя
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Добавление начальных пользователей
-- Пароль для admin: adminpass
INSERT INTO `users` (`first_name`, `last_name`, `email`, `password`, `is_admin`, `vacation_limit`) VALUES
('Admin', 'User', 'admin@example.com', '$2a$10$XCOQO2SsW9ejkTeMdlaRJuZjt/MEj8QAfH3TrUj/C2HGNTwUl8rx2', TRUE, 28);

-- Пароль для manager: managerpass (is_admin = FALSE, так как отдельной роли менеджера в схеме нет)
INSERT INTO `users` (`first_name`, `last_name`, `email`, `password`, `is_admin`, `vacation_limit`) VALUES
('Manager', 'User', 'manager@example.com', '$2a$10$OhN1EkW2WB.pFB.Lsns2XO6FcI66yxlSRvNAC/3QJpeBwXbRl3oDC', FALSE, 28);

-- Пароль для user: userpass
INSERT INTO `users` (`first_name`, `last_name`, `email`, `password`, `is_admin`, `vacation_limit`) VALUES
('Regular', 'User', 'user@example.com', '$2a$10$0EVTAr4B7tKVgFjUR2UeHO.aWKgreWogYOKXIxXzbE7c5sBLcY4bi', FALSE, 28);

-- Вывод сообщения об успешном завершении (опционально, для интерактивного выполнения)
-- SELECT 'Database initialization complete.' AS status;
