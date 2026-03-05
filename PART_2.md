## Part 2: SQL

1. Write a query to list all users and their total order amount, including users with no orders.
```sql
select users.id, users.name, count(orders.id) as 'order_amount'
from users left join orders on users.id=orders.user_id
group by users.id;
```

2. Optimization/Indexing

Query: Last 10 transactions for user_id = 123

```sql
SELECT id, user_id, amount, created_at
FROM transactions
WHERE user_id = 123
ORDER BY created_at DESC
LIMIT 10;
```

Suggested Index

```sql
CREATE INDEX idx_transactions_user_id_created_at
    ON transactions (user_id, created_at DESC);
```