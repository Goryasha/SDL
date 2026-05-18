\connect POSTGRES_DB;

INSERT INTO POSTGRES_SCHEMA.product_categories(title) VALUES
('Электроника'),
('Одежда'),
('Спортивные товары');


INSERT INTO POSTGRES_SCHEMA.products(title, price, comment, category_id) VALUES
('Смартфон', 39999.99, 'Современный смартфон', 1),
('Джинсы', 2999.99, 'Классические джинсы', 2),
('Футболка', 999.99, 'Хлопковая футболка', 2),
('Кроссовки', 4999.99, 'Легкие кроссовки', 3),
('Игровая приставка', 29999.99, 'Новая игровая приставка', 1);

INSERT INTO POSTGRES_SCHEMA.users(first_name, last_name, email, phone_number) VALUES
('Иван', 'Иванов', 'ivan@example.com', '+79111111111'),
('Анна', 'Петрова', 'anna@example.com', '+79222222222'),
('Сергей', 'Смирнов', 'sergey@example.com', '+79333333333');

INSERT INTO POSTGRES_SCHEMA.orders(user_id, total_amount) VALUES
(1, 49999.98),
(2, 5999.98),
(3, 39999.99);

INSERT INTO POSTGRES_SCHEMA.order_details(quantity, order_id, product_id) VALUES
(1, 1, 1),
(1, 1, 5),
(1, 2, 2),
(1, 2, 3),
(1, 3, 5);