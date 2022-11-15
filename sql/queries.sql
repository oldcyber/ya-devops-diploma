-- Запросы для API
-- POST /api/user/register
insert into users (login, password)
values ('$1', crypt('$2', gen_salt('bf')));

-- POST /api/user/login
SELECT *
FROM users
WHERE login = '$1'
  AND password = '$2';

-- POST /api/user/orders
INSERT INTO transactions (user_id, order_number, order_date, transaction_type)
values ('$1', '$2', now(), true);

-- GET /api/user/orders
SELECT (order_number, order_status, transaction_amount, order_date)
FROM transactions
WHERE user_id = '$1'
order by order_date desc;

-- GET /api/user/balance
-- приход - расход
-- вариант 1
SELECT (select COALESCE(SUM(transaction_amount), 0) AS income
        from transactions
        WHERE user_id = 1
          AND order_status = 'PROCESSED'
          AND transaction_type = true) -
       (select COALESCE(SUM(transaction_amount), 0) AS outcome
        FROM transactions
        WHERE user_id = 1
          AND order_status = 'PROCESSED'
          AND transaction_type = false)
FROM transactions
group by user_id;
-- расход
select COALESCE(SUM(transaction_amount), 0) AS outcome
FROM transactions
WHERE user_id = 1
  AND order_status = 'PROCESSED'
  AND transaction_type = false;

-- POST /api/user/balance/withdraw
INSERT INTO transactions (user_id, order_number, order_date, transaction_type)
values ('$1', '$2', now(), false);

-- GET /api/user/withdrawals
SELECT (order_number, transaction_amount, order_date)
FROM transactions
WHERE user_id = '$1'
  AND transaction_type = false
  AND order_status = 'PROCESSED'
order by order_date DESC;

