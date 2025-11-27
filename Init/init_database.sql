-- Добавление таблицы пользователей
create table if not exists users (
    id bigserial primary key,
    password varchar(255),
    username varchar(50),
    email varchar(50)
);
