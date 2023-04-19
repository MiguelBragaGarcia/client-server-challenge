package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const QUOTATION_URL = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

type QuotationResult struct {
	USDBRL struct {
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

func main() {
	ctx := context.Background()
	quotationCtx, cancel := context.WithTimeout(ctx, 200*time.Second)

	defer cancel()

	data, err := getDolarQuotation(quotationCtx)

	if err != nil {
		panic(err)
	}

	dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Millisecond)

	defer dbCancel()

	err = saveInDatabase(dbCtx, data)

	if err != nil {
		panic(err)
	}
}

func getDolarQuotation(quotationCtx context.Context) (QuotationResult, error) {

	// Cria um protótipo de requisição
	request, err := http.NewRequestWithContext(quotationCtx, http.MethodGet, QUOTATION_URL, nil)

	if err != nil {
		return QuotationResult{}, err
	}

	// Executa a requisição criada
	data, err := http.DefaultClient.Do(request)

	if err != nil {
		return QuotationResult{}, err
	}

	defer data.Body.Close()

	buffer, err := io.ReadAll(data.Body)

	if err != nil {
		return QuotationResult{}, err
	}

	var quotation QuotationResult

	err = json.Unmarshal(buffer, &quotation)

	if err != nil {
		return QuotationResult{}, err
	}

	return quotation, nil
}

func saveInDatabase(ctx context.Context, data QuotationResult) error {
	db, err := sql.Open("sqlite3", "database")

	if err != nil {
		return err
	}

	defer db.Close()
	// Exec é para utilizar as "querys raw". Query é para executar selects
	_, err = db.Exec(`create table if not exists quotations (
						id integer not null primary key,
						code text,
						codein,
						name,
						high text,
						low text,
						varBid text,
						pctChange text,
						bid text,
						ask text,
						timestamp text
					)`)

	if err != nil {
		return err
	}

	stmt, err := db.Prepare(`
		 insert into quotations
		 (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp)
		 values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx,
		data.USDBRL.Code,
		data.USDBRL.Codein,
		data.USDBRL.Name,
		data.USDBRL.High,
		data.USDBRL.Low,
		data.USDBRL.VarBid,
		data.USDBRL.PctChange,
		data.USDBRL.Bid,
		data.USDBRL.Ask,
		data.USDBRL.Timestamp)

	if err != nil {
		return err
	}

	return nil
}
