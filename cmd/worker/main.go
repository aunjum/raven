package main

import (
	"bytes"
	_ "bytes"
	_ "context"
	_ "encoding/json"
	_ "fmt"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
	"html/template"
	_ "html/template"
	"log"
	_ "log"
	"net/http"
	_ "net/http"
	"os"
	_ "os"
	_ "time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fs := http.FileServer(http.Dir("assets"))

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.HandleFunc("/", indexHandler)
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		return
	}
	if err != nil {
		return
	}
}

var tpl = template.Must(template.ParseFiles("assets/index.html"))

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	buf := &bytes.Buffer{}
	err := tpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		return
	}
}
