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
