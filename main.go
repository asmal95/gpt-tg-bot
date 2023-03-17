package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var randomSource = rand.NewSource(time.Now().UnixNano())
var randGenerator = rand.New(randomSource)

var tgToken string
var openAiToken string
var debug = false
var Bot *tgbotapi.BotAPI

var cache = make(map[int][]string, 0)

func init() {
	tgToken = os.Getenv("BOT_TOKEN")
	openAiToken = os.Getenv("OPEN_AI_TOKEN")

	debugEnv := os.Getenv("BOT_DEBUG")
	if debugEnv != "" {
		debug, _ = strconv.ParseBool(debugEnv)
	}
}

func main() {
	Start()
}

func Start() {
	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		panic(err)
	}

	bot.Debug = debug
	Bot = bot

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil || update.Message.Text == "" {
			continue
		}

		message := update.Message

		openAiRequest := OpenAIRequest{
			Model: "gpt-3.5-turbo",
			Messages: []OpenAIMessage{ /*OpenAIMessage{
					Role:    "assistant",
					Content: "context text",
				},*/{
					Role:    "user",
					Content: message.Text,
				}},
		}

		json_data, _ := json.Marshal(openAiRequest)

		req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(json_data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+openAiToken)
		openAiResponse, _ := http.DefaultClient.Do(req)

		var res OpenAIResponse

		json.NewDecoder(openAiResponse.Body).Decode(&res)

		fmt.Println(res.Choices[0].Message.Content)

		responseMessage := tgbotapi.NewMessage(message.Chat.ID, string(res.Choices[0].Message.Content))
		_, _ = Bot.Send(responseMessage)
	}
}

func contains(slice []string, elem string) bool {
	for _, s := range slice {
		if s == elem {
			return true
		}
	}
	return false
}

type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Id      string          `json:"id"`
	Choices []OpenAIChoices `json:"choices"`
}

type OpenAIChoices struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}
