package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type bResponce struct {
	Symbol string `json:"symbol"`
	Price float64 `json:"price,string"`
}

type wallet map[string]float64

var db = map[int]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI("YourToken")
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		command:= strings.Split(update.Message.Text, " ")
		userID := update.Message.From.ID
		switch command[0] {
		case "ADD":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный аргумент"))
				continue
			}
			_, err := getPrice(command[1])
			if err !=nil{
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "неверная валюта"))
				continue
			}
			money, err := strconv.ParseFloat(command[2], 64)
			if err != nil{
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}
			if _, ok := db[userID]; !ok {
				db[userID] = make(wallet)
			}

			db[userID][command[1]] += money
		case "SUB":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный аргумент"))
				continue
			}
			money, err := strconv.ParseFloat(command[2], 64)
			if err != nil{
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}
			if _, ok := db[userID]; !ok {
				db[userID] = make(wallet)
			}

			db[userID][command[1]] -= money
		case "SHOW":
			//fmt.Println(db)
			resp := ""
			for key, value := range db[userID]{
				usdPrice, err := getPrice(key)
				if err !=nil{
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					continue
				}
				resp += fmt.Sprintf("%s : RUB %.2f\n", key, value * usdPrice)
			}
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, resp))
		case "DEL":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный аргумент"))
				continue
			}
			delete(db[userID], command[1])
		default: bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))
		}
	}
}
func getPrice(symbol string) (float64, error){
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol)
	resp2, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=USDTRUB")
	if err != nil {
		return 0, err
	}
	var bRes2 bResponce
	err = json.NewDecoder(resp2.Body).Decode((&bRes2))
	if err != nil {
		return 0, err
	}
	if bRes2.Symbol == ""{
		return 0, errors.New("Отсутствует валютная пара USD-RUB")
	}
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	var bRes bResponce
	err = json.NewDecoder(resp.Body).Decode(&bRes)
	if err != nil {
		return 0, err
	}
	if bRes.Symbol == ""{
		return 0, errors.New("Неверная валюта")
	}
	return bRes.Price * bRes2.Price, nil
}