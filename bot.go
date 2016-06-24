package main

import (
	"encoding/json"
	"flag"
	"github.com/ChimeraCoder/anaconda"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
  "strconv"
)

type Config struct {
  TelegramToken         string `json:"token"`
  TwitterConsumerKey    string `json:"consumer_key"`
  TwitterConsumerSecret string `json:"consumer_secret"`
}

type TwitterBot struct {
  botApi     *tgbotapi.BotAPI
  client *anaconda.TwitterApi
}

func New(token string, consumer_key string, consumer_secret string) *TwitterBot {
  bot, err := tgbotapi.NewBotAPI(token)
  if err != nil {
    log.Panic(err)
  }
  bot.Debug = true
  log.Printf("Authorized on account %s", bot.Self.UserName)
  
  anaconda.SetConsumerKey(consumer_key)
  anaconda.SetConsumerSecret(consumer_secret)
  
  twitterBot := new(TwitterBot)
  twitterBot.botApi = bot
  twitterBot.client = nil
  return twitterBot
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

  twitBot := New(config.TelegramToken, config.TwitterConsumerKey, config.TwitterConsumerSecret)

	url, creds, err := anaconda.AuthorizationURL("oob")
	if err != nil {
		log.Panic(err)
	}

	var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60
	updates, err := twitBot.botApi.GetUpdatesChan(ucfg)

	for update := range updates {
		if update.Message == nil {
			continue
		}

    if _, err := strconv.Atoi(update.Message.Text); err == nil && twitBot.client == nil {
      user_creds, _, err := anaconda.GetCredentials(creds, update.Message.Text)
      if err != nil {
        log.Panic(err)
      }
      if user_creds != nil {
        twitBot.client = anaconda.NewTwitterApi(user_creds.Token, user_creds.Secret)
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Successfuly authorized. Enjoy your time :)")
        msg.ReplyToMessageID = update.Message.MessageID
        twitBot.botApi.Send(msg)
      }
      continue
    }

		if twitBot.client == nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You need to give access to twitter to use bot. Follow the link below:\r\n")
			msg.Text += url
			msg.ReplyToMessageID = update.Message.MessageID
			twitBot.botApi.Send(msg)
			continue
		}

		searchResult, _ := twitBot.client.GetSearch("golang", nil)
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyToMessageID = update.Message.MessageID
		for _, tweet := range searchResult.Statuses {
			msg.Text += tweet.Text + "\r\n=======\r\n"
		}
		twitBot.botApi.Send(msg)
	}
}
