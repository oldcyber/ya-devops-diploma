CREATE EXTENSION if not exists pgcrypto;
-- Create Types
DO $$
BEGIN
IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'trans_type') THEN
CREATE TYPE trans_type AS enum ('income', 'outcome');
END IF;
IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ord_status') THEN
    CREATE TYPE ord_status AS enum ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
END IF;
END$$;
--  create Tables
create table if not exists users
(
    user_id  serial primary key,
    login    varchar(255) not null unique,
    password varchar(255) not null
);

create table if not exists "transactions"
(
    transaction_id     serial primary key,
    order_number       bigint   not null,
    order_date         timestamp not null,
    order_status       ord_status,
    transaction_type   trans_type not null,
    transaction_amount decimal(14, 4),
    user_id            integer   not null,
    foreign key (user_id) references users (user_id)
);

create index if not exists "transactions_idx" on "transactions"
    (order_date, order_status, transaction_type, user_id);

