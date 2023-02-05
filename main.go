package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/dimchansky/utfbom"
	"github.com/xuri/excelize/v2"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Println("Start server at", addr)
	log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(handler)))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); ct != "text/csv" {
		http.Error(w, "Invalid Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")

	cr := csv.NewReader(utfbom.SkipOnly(r.Body))
	index := 1
	for {
		rs, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			http.Error(w, "Invalid CSV", http.StatusBadRequest)
			return
		}

		cells := make([]any, len(rs))
		for i, v := range rs {
			cells[i] = excelize.Cell{Value: v}
		}
		err = sw.SetRow(fmt.Sprintf("A%d", index), cells)
		if err != nil {
			log.Println(err)
		}
		index++
	}
	sw.Flush()
	f.Write(w)
}
