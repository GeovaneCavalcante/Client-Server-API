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

type Cotacao struct {
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
}

func checkCancelContext(ctx context.Context) {
	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Contexto errror: ", ctx.Err())
	}
}

func main() {
	ctx := context.Background()
	cotacao, err := fetchCotacao(ctx)
	if err != nil {
		fmt.Println("Erro: ", err)
		return
	}
	err = saveCotacao(cotacao)
	if err != nil {
		fmt.Println("Erro: ", err)
		return
	}
}

func fetchCotacao(ctx context.Context) (*Cotacao, error) {

	ctx, cancelCtx := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancelCtx()
	go checkCancelContext(ctx)

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		return nil, err
	}

	return &cotacao, nil
}

func saveCotacao(cotacao *Cotacao) error {
	f, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(fmt.Sprintf("Dólar:{%s}", cotacao.Bid)))
	if err != nil {
		return err
	}
	return nil
}
