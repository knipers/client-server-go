package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
	"time"
)

const API_URL = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

type Quotation struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

var DB *sql.DB

func main() {
	DB = ConfigureDatabase()
	defer DB.Close()
	http.HandleFunc("/cotacao", GetNewDollarQuotation)
	http.HandleFunc("/test", listAllTest)
	http.ListenAndServe(":8080", nil)
}

func ConfigureDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./quotation.db")
	if err != nil {
		panic(err)
	}

	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS quotation (id text PRIMARY KEY, code text, code_in text, bid text, create_date text, quotation_date datetime)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}
	return db
}

func GetNewDollarQuotation(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/cotacao" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	quotation, err := getExternalQuotation()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = persistQuotation(quotation)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(CotacaoResponse{quotation.Usdbrl.Bid})
}

func getExternalQuotation() (*Quotation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", API_URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var q Quotation
	json.Unmarshal(body, &q)

	return &q, err
}

func persistQuotation(q *Quotation) error {
	stmt, err := DB.Prepare("INSERT INTO quotation (id, code, code_in, bid, create_date, quotation_date) values (?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Nanosecond)
	defer cancel()

	_, err = stmt.ExecContext(ctx, uuid.NewString(), q.Usdbrl.Code, q.Usdbrl.Codein, q.Usdbrl.Bid, q.Usdbrl.CreateDate, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func listAllTest(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/test" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	rows, err := DB.Query("Select id, code, code_in, bid, create_date, quotation_date from quotation")

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var id, code, codeIn, bid, createDate string
	var quotationDate time.Time

	for rows.Next() {
		err := rows.Scan(&id, &code, &codeIn, &bid, &createDate, &quotationDate)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Printf("ID: %s, code: %s, codeIn: %s, bid: %s, createDate: %s, quotationDate: %+v \n", id, code, codeIn, bid, createDate, quotationDate)
	}
	writer.WriteHeader(http.StatusOK)
}
