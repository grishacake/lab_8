package main

import (
	"database/sql"
	"encoding/json"
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
	dbname = "count"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

// Обработчик GET для получения значения счетчика
func (h *Handlers) GetCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.dbProvider.SelectCount()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Текущий счетчик: %d", count)))
}

// Обработчик POST для увеличения счетчика
func (h *Handlers) PostCount(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Count int `json:"count"`
	}{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}

	if input.Count <= 0 {
		http.Error(w, "Значение count должно быть положительным числом", http.StatusBadRequest)
		return
	}

	err = h.dbProvider.UpdateCount(input.Count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Счетчик увеличен на %d", input.Count)))
}

// Методы для работы с базой данных
func (dp *DatabaseProvider) SelectCount() (int, error) {
	var count int
	row := dp.db.QueryRow("SELECT count FROM counters WHERE id = 1")
	err := row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если записи нет, создаем начальный счетчик
			_, err := dp.db.Exec("INSERT INTO counters (count) VALUES (0)")
			if err != nil {
				return 0, err
			}
			count = 0
		} else {
			return 0, err
		}
	}
	return count, nil
}

func (dp *DatabaseProvider) UpdateCount(increment int) error {
	_, err := dp.db.Exec("UPDATE counters SET count = count + $1 WHERE id = 1", increment)
	if err != nil {
		return err
	}
	return nil
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

	http.HandleFunc("/count/get", h.GetCount)   // Обработчик GET-запроса
	http.HandleFunc("/count/post", h.PostCount) // Обработчик POST-запроса для увеличения

	err = http.ListenAndServe(*address, nil)
	if err != nil {
		log.Fatal(err)
	}
}
