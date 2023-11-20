package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

type Moeda struct {
	USDBRL Cotacao `json:"USDBRL"`
}

func checkCancelContext(ctx context.Context) {
	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Contexto errror: ", ctx.Err())
	}
}

func initDatabase() (*gorm.DB, error) {
	dsn := "cotacoes.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Cotacao{})
	return db, nil
}

func main() {
	db, err := initDatabase()
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		CotacacaoHandler(w, r, db)
	})
	http.ListenAndServe(":8080", mux)
}

func CotacacaoHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	ctx := r.Context()

	cotacao, err := fetchCotacao(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = storeCotacao(ctx, db, cotacao)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao)
}

func fetchCotacao(ctx context.Context) (*Cotacao, error) {

	ctx, cancelCtx := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancelCtx()
	go checkCancelContext(ctx)

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
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

	var moeda Moeda
	err = json.Unmarshal(body, &moeda)
	if err != nil {
		return nil, err
	}

	return &moeda.USDBRL, nil
}

func storeCotacao(ctx context.Context, db *gorm.DB, cotacao *Cotacao) error {
	ctx, cancelCtx := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancelCtx()
	go checkCancelContext(ctx)
	result := db.WithContext(ctx).Create(&cotacao)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
