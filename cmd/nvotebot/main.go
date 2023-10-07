package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/goccy/go-json"
	"github.com/iwalfy/nvotebot/api"
	. "github.com/iwalfy/nvotebot/util"
	ju "github.com/iwalfy/nvotebot/util/json"
	tg "github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

var (
	debugLogPath = Getenv("NVOTE_DEBUG_LOG_PATH", "./debug.log")
	apiUrl       = Getenv("NVOTE_API_URL", "https://nvote.lebedinets.ru/vote_bot.php")
	apiToken     = os.Getenv("NVOTE_API_TOKEN")
	botToken     = os.Getenv("NVOTE_BOT_TOKEN")
)

func init() {
	if apiToken == "" {
		panic("NeuralOpenNet token is not set in NVOTE_API_TOKEN")
	}
	if botToken == "" {
		panic("bot token is not set in NVOTE_BOT_TOKEN")
	}
}

var (
	ctx      context.Context
	neuroapi *api.Client
)

func main() {
	logger.Info("Starting...")

	stopChan := make(chan struct{})
	var ctxCancel context.CancelFunc
	ctx, ctxCancel = context.WithCancel(context.Background())

	neuroapi = api.NewClient(apiUrl, apiToken)

	logger.Info("Creating a Telegram Bot API client")
	bot := Must(tg.NewBot(botToken, tg.WithExtendedDefaultLogger(false, true, nil)))
	logger.Info("Running a long update poll")
	updates := Must(bot.UpdatesViaLongPolling(nil))

	handler := Must(th.NewBotHandler(bot, updates))

	defer func() {
		ctxCancel()
		handler.Stop()
		bot.StopLongPolling()
		logger.Info("Exit")
		_ = debugLogFile.Close()
	}()

	handler.Use(func(bot *tg.Bot, update tg.Update, next th.Handler) {
		_, _ = debugLogFile.WriteString(string(Must(json.Marshal(update))))
		_, _ = debugLogFile.WriteString("\n")

		next(bot, update)
	})

	handler.HandleMessage(nextCommand, th.CommandEqual("start"))
	handler.HandleMessage(nextCommand, th.CommandEqual("next"))
	handler.HandleCallbackQuery(voteCallback, th.AnyCallbackQueryWithMessage())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		for range signals {
			fmt.Print("\r")
			stopChan <- struct{}{}
		}
	}()

	go handler.Start()
	logger.Info("Started!")

	<-stopChan
}

func nextCommand(bot *tg.Bot, message tg.Message) {
	sendVote(bot, message.Chat.ID)
}

func voteCallback(bot *tg.Bot, query tg.CallbackQuery) {
	id := query.Message.Chat.ID

	var meta buttonMetadata

	if err := json.Unmarshal([]byte(query.Data), &meta); err != nil {
		answerCallback(bot, query.ID, "???")
		logger.Error("failed to deserialize meta", zap.Error(err))
		return
	}

	if meta.Voted == 1 {
		answerCallback(bot, query.ID, "Вы уже проголосовали")
		return
	}

	if meta.Skipped == 1 {
		_ = bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		sendVote(bot, id)
		return
	}

	if meta.Like == 0 && meta.Skipped == 0 {
		meta.Voted = 1

		_, err := bot.EditMessageText(&tg.EditMessageTextParams{
			ChatID:    tu.ID(id),
			MessageID: query.Message.MessageID,
			Text:      query.Message.Text,
			ReplyMarkup: tu.InlineKeyboard(
				tu.InlineKeyboardRow(
					tu.InlineKeyboardButton("👍").WithCallbackData(ju.Stringify(meta)),
					tu.InlineKeyboardButton("➡️").WithCallbackData(ju.Stringify(buttonMetadata{Like: 0, Voted: 0, Skipped: 1})),
				),
			),
		})
		handleMessageEditError(query.Message.Chat.ID, query.Message.MessageID, err)

		sendVote(bot, id)
		return
	}

	res, err := neuroapi.Vote(ctx, id, meta.UUID)
	if err != nil {
		var apiError *api.Error
		if errors.As(err, &apiError) {
			answerCallback(bot, query.ID, apiError.Error())
		} else {
			answerCallback(bot, query.ID, "Сервер временно недоступен")
		}
		logger.Error("received an error from api", zap.Error(err))
		return
	}

	meta.Voted = 1

	_, err = bot.EditMessageText(&tg.EditMessageTextParams{
		ChatID:    tu.ID(id),
		MessageID: query.Message.MessageID,
		Text:      query.Message.Text,
		ReplyMarkup: tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton(fmt.Sprintf("👍 (%d)", res.Votes)).WithCallbackData(ju.Stringify(meta)),
				tu.InlineKeyboardButton("➡️").WithCallbackData(ju.Stringify(buttonMetadata{Like: 0, Voted: 0, Skipped: 1})),
			),
		),
	})
	handleMessageEditError(query.Message.Chat.ID, query.Message.MessageID, err)
	answerCallback(bot, query.ID, "Успешно")
	sendVote(bot, id)
}

func sendVote(bot *tg.Bot, id int64) {
	res, err := neuroapi.Get(ctx, id, 1)
	if err != nil {
		var apiError *api.Error
		if errors.As(err, &apiError) {
			sendMessage(bot, id, apiError.Error())
		} else {
			sendMessage(bot, id, "Сервер временно недоступен")
		}
		logger.Error("received an error from api", zap.Error(err))
		return
	}

	if len(res.ToVote) < 1 {
		sendMessage(bot, id, "Сервер отдал неверный ответ")
		return
	}

	variant := res.ToVote[0]

	metadata := ju.Stringify(buttonMetadata{
		Like:    1,
		Voted:   0,
		Skipped: 0,
		UUID:    variant.UUID,
	})

	_, err = bot.SendMessage(
		tu.Message(
			tu.ID(id),
			variant.Text,
		).WithReplyMarkup(tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("👍").WithCallbackData(metadata),
				tu.InlineKeyboardButton("➡️").WithCallbackData(ju.Stringify(buttonMetadata{Like: 0, Skipped: 1})),
			),
		)),
	)
	handleMessageSendError(id, err)

}

func sendMessage(bot *tg.Bot, id int64, text string) {
	_, err := bot.SendMessage(tu.Message(
		tu.ID(id),
		text,
	))
	handleMessageSendError(id, err)
}

func answerCallback(bot *tg.Bot, id string, text string) {
	if err := bot.AnswerCallbackQuery(tu.CallbackQuery(id).WithText(text)); err != nil {
		logger.Error(
			"An error occured while sending answering to callback query",
			zap.String("query", id),
			zap.String("text", text),
			zap.Error(err),
		)
	}
}

func handleMessageSendError(id int64, err error) {
	if err != nil {
		logger.Error(
			"An error occured while sending message",
			zap.Int64("id", id),
			zap.Error(err),
		)
	}
}

func handleMessageEditError(chatId int64, messageId int, err error) {
	if err != nil {
		logger.Error(
			"An error occured while editing message",
			zap.Int64("chat_id", chatId),
			zap.Int("message_id", messageId),
			zap.Error(err),
		)
	}
}

type buttonMetadata struct {
	Like    int    `json:"l"`
	Voted   int    `json:"v"`
	Skipped int    `json:"s"`
	UUID    string `json:"u,omitempty"`
}
