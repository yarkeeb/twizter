package main

import (
	"encoding/json"
	"flag"
	"github.com/ChimeraCoder/anaconda"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"strings"
)

//Configuration struct for setting up Twizter
type Config struct {
	TelegramToken         string `json:"token"`
	TwitterConsumerKey    string `json:"consumer_key"`
	TwitterConsumerSecret string `json:"consumer_secret"`
	TwitterAccessToken    string `json:"access_token"`
	TwitterAccessSecret   string `json:"acccess_secret"`
}

//Wrapper for Telegram message
type message struct {
	Message *tgbotapi.Message
	Cmd     string
	Args    []string
}

//Handler for a bot command
type CmdFunc func(m *message)

//Map for storing handlers to bot commands
type CmdMap map[string]CmdFunc

//Basic bot structure
type Twizter struct {
	botApi   *tgbotapi.BotAPI
	client   *anaconda.TwitterApi
	commands CmdMap
}

//Create new Twizter bot
func New(arr []byte) *Twizter {
	var config Config
	err := json.Unmarshal(arr, &config)
	if err != nil {
		log.Panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	anaconda.SetConsumerKey(config.TwitterConsumerKey)
	anaconda.SetConsumerSecret(config.TwitterConsumerSecret)

	Twizter := new(Twizter)
	Twizter.botApi = bot
	Twizter.client = anaconda.NewTwitterApi(config.TwitterAccessToken, config.TwitterAccessSecret)
	Twizter.commands = Twizter.getDefaultCommands()
	return Twizter
}

//Init default bot commads
func (tb *Twizter) getDefaultCommands() CmdMap {
	return CmdMap{
		"/start":  tb.Start,
		"/search": tb.Search,
	}

}

//Process messages from telegram
func (tb *Twizter) ProcessUpdate(m *tgbotapi.Message) {
	if m == nil {
		return
	}

	if strings.HasPrefix(m.Text, "/") {
		tb.handleCommand(m)
		return
	}
}

//Routing commands
func (tb *Twizter) handleCommand(m *tgbotapi.Message) {

	tbmsg := tb.parseMessage(m)
	exec := tb.commands[tbmsg.Cmd]
	if exec != nil {
		go exec(tbmsg)
	}
}

//Create message wrapper
func (tb *Twizter) parseMessage(msg *tgbotapi.Message) *message {
	tokens := strings.Fields(msg.Text)
	cmd, args := strings.ToLower(tokens[0]), tokens[1:]
	return &message{Cmd: cmd, Args: args, Message: msg}
}

//Function to handle bot /start command
func (tb *Twizter) Start(m *message) {
	msg := tgbotapi.NewMessage(m.Message.Chat.ID, "Hello there! I'm twitter bot.\r\nTo search tweets, type: /search <tweet>")
	msg.ReplyToMessageID = m.Message.MessageID
	tb.botApi.Send(msg)
}

//Function to handle bot /search command
func (tb *Twizter) Search(m *message) {
	for _, str := range m.Args {
		searchResult, _ := tb.client.GetSearch(str, nil)
		log.Printf("[%s] %s", m.Message.From.UserName, m.Message.Text)
		msg := tgbotapi.NewMessage(m.Message.Chat.ID, "")
		msg.ReplyToMessageID = m.Message.MessageID
		for _, tweet := range searchResult.Statuses {
			msg.Text += tweet.Text + "\r\n=======\r\n"
		}
		tb.botApi.Send(msg)
	}
	return
}

func main() {
	configPtr := flag.String("config", "./config/config.json", "Path to config file")
	flag.Parse()

	array, err := ioutil.ReadFile(*configPtr)
	if err != nil {
		log.Panic(err)
	}

	twitBot := New(array)

	var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60
	updates, err := twitBot.botApi.GetUpdatesChan(ucfg)

	for update := range updates {
		twitBot.ProcessUpdate(update.Message)
	}

}
