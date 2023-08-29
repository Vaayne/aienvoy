package tgbot

import (
	"aienvoy/internal/pkg/config"
	"aienvoy/internal/pkg/logger"
	"aienvoy/internal/ports/tgbot/handler"

	"github.com/pocketbase/pocketbase"
	tb "gopkg.in/telebot.v3"
)

type TeleBot struct {
	*tb.Bot
	// app is use for db usage
	app *pocketbase.PocketBase
}

var bot *TeleBot

func New(token string, app *pocketbase.PocketBase) *TeleBot {
	b, err := tb.NewBot(tb.Settings{
		Token: token,
		// Poller:  WebHook,
		// Verbose: false,
	})
	if err != nil {
		logger.SugaredLogger.Fatalf("Init telebot error", "err", err)
	}

	teleBot := &TeleBot{
		Bot: b,
		app: app,
	}
	// teleBot.registerHandlers()

	return teleBot
}

func DefaultBot(app *pocketbase.PocketBase) *TeleBot {
	if bot == nil {
		bot = New(config.GetConfig().Telegram.Token, app)
	}
	return bot
}

func appMiddleware(next tb.HandlerFunc) tb.HandlerFunc {
	return func(c tb.Context) error {
		c.Set("app", bot.app)
		return next(c) // continue execution chain
	}
}

func registerHandlers(b *TeleBot) {
	b.Handle(tb.OnText, handler.OnText)
}

func Serve(app *pocketbase.PocketBase) {
	b := DefaultBot(app)
	b.Use(appMiddleware)
	registerHandlers(b)
	b.Start()
}