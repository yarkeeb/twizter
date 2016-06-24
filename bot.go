package main

import (
	"encoding/json"
	"flag"
	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
)

type Config struct {
	TelegramToken         string `json:"token"`
	TwitterConsumerKey    string `json:"consumer_key"`
	TwitterConsumerSecret string `json:"consumer_secret"`
}

type Twitter struct {
	Api       *anaconda.TwitterApi
	TempCreds *oauth.Credentials
	Verifier  string
	URL       string
}

var twitter Twitter

func handler(w http.ResponseWriter, r *http.Request) {
	twitter.Verifier = r.URL.Query()["oauth_verifier"][0]

	user_creds, _, err := anaconda.GetCredentials(twitter.TempCreds, twitter.Verifier)
	if err != nil {
		log.Panic(err)
	}

	twitter.Api = anaconda.NewTwitterApi(user_creds.Token, user_creds.Secret)
	http.Redirect(w, r, "https://www.twitter.com", 301)
}

func main() {
	configPtr := flag.String("config", "./config/config.json", "Path to config file")
	flag.Parse()

	var config Config
	array, err := ioutil.ReadFile(*configPtr)
	if err != nil {
		log.Panic(err)
	}
	json.Unmarshal(array, &config)

	anaconda.SetConsumerKey(config.TwitterConsumerKey)
	anaconda.SetConsumerSecret(config.TwitterConsumerSecret)
	url, creds, err := anaconda.AuthorizationURL("http://127.0.0.1:8080/twitter_callback")
	if err != nil {
		log.Panic(err)
	}
	twitter.URL = url
	twitter.TempCreds = creds

	http.HandleFunc("/twitter_callback", handler)
	go http.ListenAndServe(":8080", nil)

	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60
	updates, err := bot.GetUpdatesChan(ucfg)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Text == "test_auth" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, twitter.URL)
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			continue
		}

		if twitter.Api == nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You need to give access to twitter to use bot. Follow the link below:\r\n")
			msg.Text += twitter.URL
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			continue
		}

		searchResult, _ := twitter.Api.GetSearch("golang", nil)
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyToMessageID = update.Message.MessageID
		for _, tweet := range searchResult.Statuses {
			msg.Text += tweet.Text + "\r\n=======\r\n"
		}
		bot.Send(msg)
	}
}
