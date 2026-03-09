package test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
	"zeroboo.payment/handler"
	"zeroboo.payment/service"
)

// Suite-level shared state
var (
	suiteDB      *sql.DB
	suiteBaseURL string
	suiteSrv     *httptest.Server
)

// TestMain – suite setup & teardown
func TestMain(m *testing.M) {
	// ---- Setup ----
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/payment_test?parseTime=true"
	}

	var err error
	suiteDB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("skipping integration tests: cannot open db: %v", err)
		return
	}
	if err = suiteDB.Ping(); err != nil {
		log.Printf("skipping integration tests: cannot ping db: %v", err)
		suiteDB.Close()
		return
	}
	log.Println("integration test suite: database connected")

	// Start the HTTP server once for all tests.
	gin.SetMode(gin.TestMode)
	svc := service.NewPaymentService(suiteDB)
	r := gin.New()
	h := handler.NewPaymentHandler(svc, nil)
	h.RegisterRoutes(r)
	suiteSrv = httptest.NewServer(r)
	suiteBaseURL = suiteSrv.URL
	log.Printf("integration test suite: server started at %s", suiteBaseURL)

	// Run tests
	code := m.Run()

	// ---- Teardown ----
	suiteSrv.CloseClientConnections()
	suiteSrv.Close()
	suiteDB.Close()
	// Give the OS a moment to release file handles (Windows file-lock workaround).
	time.Sleep(1000 * time.Millisecond)
	log.Println("integration test suite: teardown complete")

	os.Exit(code)
}

// ---------------------------------------------------------------------------
// Per-test helpers
// ---------------------------------------------------------------------------

// cleanTables truncates all tables before each test for isolation.
func cleanTables(t *testing.T) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM transaction_locks",
		"DELETE FROM events",
		"DELETE FROM transactions",
		"DELETE FROM user_balances",
	} {
		if _, err := suiteDB.Exec(stmt); err != nil {
			t.Fatalf("cleanup failed (%s): %v", stmt, err)
		}
	}
}

// seedUserBalance inserts a user balance row for testing.
func seedUserBalance(t *testing.T, userID string, balance int64) {
	t.Helper()
	_, err := suiteDB.Exec(
		"INSERT INTO user_balances (user_id, balance) VALUES (?, ?) ON DUPLICATE KEY UPDATE balance = ?",
		userID, balance, balance,
	)
	if err != nil {
		t.Fatalf("seed user balance: %v", err)
	}
}

// postPay sends a real HTTP POST /pay to the suite server.
func postPay(t *testing.T, body interface{}) *http.Response {
	t.Helper()
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	resp, err := http.Post(suiteBaseURL+"/pay", "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	return resp
}

// decodeBody reads and JSON-decodes the response body into dest.
func decodeBody(t *testing.T, resp *http.Response, dest interface{}) {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if err := json.Unmarshal(body, dest); err != nil {
		t.Fatalf("decode response JSON: %v\nbody: %s", err, string(body))
	}
}
