package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const API_URL = "http://localhost:8080/cotacao"

type ResponseBody struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", API_URL, nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var response ResponseBody
	json.Unmarshal(body, &response)
	if err := saveDollarInFile(response.Bid); err != nil {
		panic(err)
	}
}

func saveDollarInFile(dolar string) error {
	os.Remove("cotacao.txt")
	file, err := os.Create("cotacao.txt")
	defer file.Close()
	if err != nil {
		return err
	}

	if _, err := file.Write([]byte(fmt.Sprintf("Dolar: %s", dolar))); err != nil {
		return err
	}

	return nil
}
