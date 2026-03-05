# sgh-asia-assignment
Submission for [assignment from SGH Asia](https://better-airport-510.notion.site/Advanced-Test-263f007ae57480a8b051c87b500b0e04)


## Part 1
### 3. Code Review – Bad Go Code

You are given the following code:

```go
var data = ""

func handler(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		data = string(body)
		fmt.Fprintf(w, "Saved: %s", data)
		defer r.Body.Close()
}
```
### Problems of the code  
1. This code silently omits the error returns by ioutil.ReadAll. This will lead to undetermined result (_data_ partialy load or empty) and limit the traceability when error occurs.  
Fix: 
```go
body, err := ioutil.ReadAll(r.Body)
if err != nil {
    //Log error...
    
    //Response to client...
    http.Error(w, "failed to read body", http.StatusBadRequest)
    //Fail fast and stop processing
    return
}
```
2. _defer_ should be called right after the resource opening to make sure the closing is scheduled. Placing _defer r.Body.Close()_ at the end of function:
  - Potential bug: if the code has a fatal before reaching _defer_
  - Make the maintaining easier to produce bug: the maintainer may places a return before reaching _defer_  
in both cases the resource will not be closed and leads to memory leaks  
Fix:
```go
body, err := ioutil.ReadAll(r.Body)
defer r.Body.Close()
```

3. _data_ is declared as global variable, it then modified inside function _handler_. It can lead to:
- Race condition: multiple goroutines update the variable, it makes the program behave unpredictably. 
- Memory waste: data may hold a huge buffer and will not be freed until next read

If the intentional is to hold a copy of read data, use a mechanism of synchonization like Mutex or channel, however declaring _data_ as local variable is more reasonable.

4. Writing whole read data into response doesn't make sense, it is wasting resource.
Fix: only write amount of read data
```go
fmt.Fprintf(w, "Saved: %d bytes", len(data))
```
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

## Part 3: Code Review Exercise

[Answer](./PART_3.md)