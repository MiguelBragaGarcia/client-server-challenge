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

const SERVER_URL = "http://localhost:8080/cotacao"

type QuotationResultFromServer struct {
	USDBRL struct {
		Bid string
	}
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)

	defer cancel()

	quotation, err := getDolarQuotationFromServer(ctx)
	if err != nil {
		panic(err)
	}

	saveQuotationInFile(quotation)

}

func getDolarQuotationFromServer(ctx context.Context) (*QuotationResultFromServer, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, SERVER_URL, nil)
	if err != nil {
		return &QuotationResultFromServer{}, err
	}

	data, err := http.DefaultClient.Do(request)

	if err != nil {
		return &QuotationResultFromServer{}, err
	}

	defer data.Body.Close()

	body, err := io.ReadAll(data.Body)
	if err != nil {
		return &QuotationResultFromServer{}, err
	}

	var quotationResult QuotationResultFromServer

	json.Unmarshal(body, &quotationResult)

	return &quotationResult, nil
}

func saveQuotationInFile(data *QuotationResultFromServer) {
	file, err := os.Create("cotacao.txt")

	if err != nil {
		panic(err)
	}

	defer file.Close()

	formattedQuotation := fmt.Sprintf("Dólar: %s", data.USDBRL.Bid)

	file.WriteString(formattedQuotation)
}
