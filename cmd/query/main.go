package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "grishka"
	dbname = "query"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

// Обработчик GET для получения приветствия по имени
func (h *Handlers) GetGreeting(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Нет параметра 'name'", http.StatusBadRequest)
		return
	}

	greeting, err := h.dbProvider.SelectGreeting(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(greeting))
}

// Методы для работы с базой данных
func (dp *DatabaseProvider) SelectGreeting(name string) (string, error) {
	var greeting string
	row := dp.db.QueryRow("SELECT greeting FROM greetings WHERE name = $1", name)
	err := row.Scan(&greeting)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err := dp.db.Exec("INSERT INTO greetings (name, greeting) VALUES ($1, $2)", name, fmt.Sprintf("Hello, %s!", name))
			if err != nil {
				return "", err
			}
			greeting = fmt.Sprintf("Hello, %s!", name)
		} else {
			return "", err
		}
	}
	return greeting, nil
}

func main() {
	address := flag.String("address", "127.0.0.1:8081", "адрес для запуска сервера")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	// Регистрируем обработчик для /api/user
	http.HandleFunc("/api/user", h.GetGreeting)

	// Запускаем веб-сервер на указанном адресе
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		log.Fatal(err)
	}
}
