-- Убедитесь, что вы используете правильную базу данных
USE vacation_scheduler;

-- Вставка администратора
INSERT INTO users (username, password, full_name, email, is_admin, is_manager)
VALUES ('admin', '$2a$10$AXmdtgiIW31LNEuDOJhupevhAvjgkXESIl0XGO5tI3qvrvd.PT9eu', 'Admin User', 'admin@example.com', TRUE, FALSE);

-- Вставка менеджера
INSERT INTO users (username, password, full_name, email, is_admin, is_manager)
VALUES ('manager', '$2a$10$MfGQ5oDtYx8jtAUzp1ewXuRucnAe7LYqNXvY877yTz0V.gXzgJxuq', 'Manager User', 'manager@example.com', FALSE, TRUE);

-- Вставка обычного пользователя
INSERT INTO users (username, password, full_name, email, is_admin, is_manager)
VALUES ('user', '$2a$10$lGVfIXM6oJDbE.Jdd4ItPOz9bAR8ZFiys0Xst1gF7vfJ5fFCldauO', 'Regular User', 'user@example.com', FALSE, FALSE);
