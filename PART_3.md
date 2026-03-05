## Part 3: Code Review Exercise
### Problem
Given a poorly structured Go code and suggest way to refactor and improve the code. 

```golang
package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
)

var result string

func handler(w http.ResponseWriter, r *http.Request) {
    body, _ := ioutil.ReadAll(r.Body) 
    result = string(body)             
    fmt.Fprintf(w, "Saved: %s", result)
    defer r.Body.Close()              
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
```

### Answer
- Improve handler
  - Use defer right after resource opening
  - Declare result as local variable
  - Limit the read buffer 
  - 
- Improve the bootstrap code: 
  - handle error of http.ListenAndServe 
  - add graceful shutdown and listen to system signal
  - configure server with proper timeout
  - set appropriate method for handlers
- Arrange code to meaningful packages
  - server package for bootstraping server
  - handler package for handlers


Codes after:
- handler/save.go
```go
package handler

import (
	"fmt"
	"io"
	"net/http"
)

const maxBodySize = 1 << 20 // 1 MB

// Save reads the request body and echoes it back.
func Save(w http.ResponseWriter, r *http.Request) {
	//Validate requests
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Read with maxBodySize
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		//Handle error
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	//Set headers explicitly
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Saved: %s", string(body))
}

```
- server/main.go
```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zeroboo.webservice/handler"
)

func newRouter() http.Handler {
	mux := http.NewServeMux()

	//Use appropriate method
	mux.HandleFunc("POST /save", handler.Save)
	return mux
}

const (
	serverAddr      = ":8080"
	readTimeout     = 10 * time.Second
	writeTimeout    = 10 * time.Second
	shutdownTimeout = 10 * time.Second
)

func main() {
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      newRouter(),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// Start server in a goroutine.
	go func() {
		log.Printf("starting server on %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Wait for system signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// gracefully shut down
	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server stopped")
}

```


