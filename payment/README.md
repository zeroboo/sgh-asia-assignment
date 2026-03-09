
## Overview
Project demonstrate an REST API server.
API:

Technologies:
- [Gin Web Framework](https://gin-gonic.com/en/) to handle HTTP requests
- MySQL: [schema](./schema.sql])

## Build

## Test
```sql 
CREATE DATABASE IF NOT EXISTS payment_test;
USE payment_test;
SOURCE schema.sql;
```

```shell
cd D:\Projects\zeroboo\interview\sgh-asia\payment
go test ./test/ -v -count=1
```
## Run



