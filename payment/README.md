
## Overview
Project demonstrate a REST API server. Implemented using [Gin Web Framework](https://gin-gonic.com/en/) to handle HTTP requests and MySQL for storage.

### 1. APIs:
  - POST /pay: Process the payment.
    - Request: userID, amount, transactionID as POST request with JSON content
    - Response: transactionId, amount, result of payment
    - API is idempotent using the transactionId as idempotent key. 
    - API prevents duplication in processing by issuing a temporary lock on transactionId 

### 2. Database: 
#### Schema: [schema.sql](./schema.sql)

#### Tables

| Table | Purpose |
|-------|---------|
| **transactions** | Stores every payment transaction. Each row tracks the transaction ID, user, amount, type (`debit`/`credit`), and lifecycle status (`created` → `processing` → `completed`/`failed`/`refunded`). The `transaction_id` is the primary key and also serves as the idempotency key. |
| **user_balances** | Holds the current wallet balance for each user (`user_id` → `balance`). Updated atomically inside the same DB transaction that completes a payment. |
| **events** | Append-only event log for auditing and event-sourcing. Records events such as `payment.created`, `payment.completed`, and `payment.failed` with a JSON `payload` linked to a `transaction_id`. |
| **transaction_locks** | Advisory lock table to prevent concurrent duplicate processing of the same transaction. A row is inserted when processing begins (with an `expires_at` TTL) and deleted on completion. |


### 3. Project Folder Structure

```
payment/
├── server/                 # Application entry point
│   └── main.go             #   HTTP server bootstrap, DB connection, graceful shutdown
│
├── handler/                # HTTP / Gin request handlers
│   ├── dto.go              #   Request & response structs (PayRequest, PayResponse, BaseResponse)
│   ├── payment.go          #   PaymentHandler – validates input, orchestrates service calls
│   └── route.go            #   Route registration helper
│
├── service/                # Business logic layer
├── repository/             # Data-access layer (MySQL queries)
├── model/                  # Domain structs shared across layers
├── middleware/             # HTTP middleware (auth, rate-limiting, etc.)
├── transfer/               # Transfer-related logic
├── test/                   # Integration tests (real HTTP + real MySQL)
└── deployments/            # Docker / docker-compose files for test environment
    ├── docker-compose.yml  #   Spins up MySQL 8.0 for integration tests
    └── mysql-init/
        └── 01-init.sql     #   Creates payment_test DB & tables on first start
```


## Build
Under `payment` folder:
```shell

```

## Test
Tests of project placed under folder `payment/test`

### 1. Start test database
```shell
cd payment/deployments
docker compose up -d
```
This starts a MySQL 8.0 container on port **3306** with root password `root`.
The init script automatically creates the `payment_test` database and all tables.

### 2. Run tests
```shell
cd payment
go test ./test/ -v -count=1
```
The test suite connects to `root:root@tcp(127.0.0.1:3306)/payment_test` by default.
Override with the `MYSQL_TEST_DSN` environment variable if needed.

### 3. Tear down
```shell
cd payment/deployments
docker compose down -v
```



