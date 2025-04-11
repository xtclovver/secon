-- Создание базы данных (если она еще не существует)
CREATE DATABASE IF NOT EXISTS vacation_db
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

-- Переключение на созданную базу данных
USE vacation_db;

-- Создание таблицы пользователей (users)
CREATE TABLE IF NOT EXISTS `users` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `first_name` VARCHAR(255) NOT NULL,
    `last_name` VARCHAR(255) NOT NULL,
    `email` VARCHAR(191) NOT NULL UNIQUE, -- 191 для совместимости с utf8mb4 индексами
    `password` VARCHAR(255) NOT NULL, -- Хранить только хеш пароля
    `is_admin` BOOLEAN NOT NULL DEFAULT FALSE, -- Роль пользователя (администратор или нет)
    `vacation_limit` INT NOT NULL DEFAULT 28, -- Примерное значение по умолчанию
    -- Добавьте другие поля при необходимости, например:
    -- `department_id` INT UNSIGNED,
    -- `position` VARCHAR(255),
    -- `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
    -- `comment` TEXT, -- Раскомментируйте, если нужно поле для комментариев

    INDEX `idx_vacation_requests_user_id` (`user_id`),
    INDEX `idx_vacation_requests_status` (`status`),
    INDEX `idx_vacation_requests_start_date` (`start_date`),
    INDEX `idx_vacation_requests_end_date` (`end_date`),

    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE -- Удалять заявки при удалении пользователя
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Примечание:
-- 1. Тип ENUM для статуса удобен, но может потребовать изменения при добавлении новых статусов. Альтернатива - VARCHAR.
-- 2. Тип DATE для дат отпуска обычно достаточен. Если нужно учитывать время, используйте DATETIME или TIMESTAMP.
-- 3. Подумайте о добавлении дополнительных индексов в зависимости от частых запросов.
-- 4. Убедитесь, что ваш MySQL сервер поддерживает InnoDB и utf8mb4.
