package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/LlamaNite/llamalog"
	"github.com/haashemi/tgo"
	"github.com/haashemi/tgo/filters"
	"github.com/haashemi/tgo/routers/message"
	"gopkg.in/yaml.v3"
)

var log = llamalog.NewLogger("ReNamaeBot")

// Config holds the bot's configurations.
type Config struct {
	Host      string  `yaml:"host"`      // Local hosted telegram-bot-api address
	Token     string  `yaml:"token"`     // Bot's token gathered from @BotFather
	Whitelist []int64 `yaml:"whitelist"` // Whitelisted user IDs
}

func main() {
	// Load the config.yaml file
	var config Config
	if f, err := os.ReadFile("config.yaml"); err != nil {
		log.Fatal("Failed to read the config file. > %v", err)
	} else if err = yaml.Unmarshal(f, &config); err != nil {
		log.Fatal("Failed to decode the config file. > %v", err)
	}

	// Initialize a new bot-instance
	bot := tgo.NewBot(config.Token, tgo.Options{Host: config.Host})

	info, err := bot.GetMe()
	if err != nil {
		log.Fatal("Failed to fetch the bot info > %v", err)
	}

	mr := message.NewRouter()
	mr.Handle(filters.Command("start", info.Username), OnStart)
	mr.Handle(filters.And(HasDocument(), filters.Whitelist(config.Whitelist...)), OnDocument)
	bot.AddRouter(mr)

	for {
		log.Info("Bot is running as @%s", info.Username)
		if err := bot.StartPolling(30); err != nil {
			log.Error("Polling stopped > %v", err)

			log.Warn("Sleeping for 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}
}

// OnStart handles /start commands sent to the bot
func OnStart(ctx *message.Context) {
	ctx.Send(&tgo.SendMessage{
		Text: `👋 Hi! 

🤔 I'm ReNamae, a simple bot to rename documents' filenames to something else. My source code is also publicly available; you can run me on your own server.

🪄 https://github.com/haashemi/ReNamaeBot

🫡 I'm a private bot; thus I'll only respond to the whitelisted users. So, if you're @Byfron's friend, tell him to add you to the list of whitelisted users.


「 Reなまえ 」
`,
		LinkPreviewOptions: &tgo.LinkPreviewOptions{IsDisabled: true},
	})

	log.Info("User started the bot. ID: %d | Username: %s | Name: %s", ctx.From.Id, ctx.From.Username, ctx.From.FirstName)
}

// OnDocument handles every documents sent to the bot.
func OnDocument(ctx *message.Context) {
	_, ans, err := ctx.Ask(&tgo.SendMessage{
		Text: "✍️ Send the new filename.",
		ReplyMarkup: tgo.ReplyKeyboardMarkup{
			Keyboard:              [][]*tgo.KeyboardButton{{{Text: "🚫 Cancel"}}},
			ResizeKeyboard:        true,
			OneTimeKeyboard:       true,
			InputFieldPlaceholder: "Write the new filename...",
		},
	}, 5*time.Minute)
	if err != nil {
		ctx.Send(&tgo.SendMessage{Text: "🚫 Failed to get the new filename.\n\n" + err.Error()})
		return
	} else if ans.String() == "🚫 Cancel" {
		ctx.Send(&tgo.SendMessage{Text: "⚠️ Renaming the file has been cancelled."})
		return
	}

	ctx.Send(&tgo.SendMessage{Text: "📥 Downloading the file.\n\n⏳ Please wait..."})

	file, err := ctx.Bot.GetFile(&tgo.GetFile{FileId: ctx.Document.FileId})
	if err != nil {
		ctx.Send(&tgo.SendMessage{Text: "🚫 Failed to GetFile.\n\n" + err.Error()})
		return
	}

	dir, err := os.MkdirTemp(os.TempDir(), "renamae-bot-*")
	if err != nil {
		ctx.Send(&tgo.SendMessage{Text: "🚫 Failed to create a temp-dir.\n\n" + err.Error()})
		return
	}
	defer os.RemoveAll(dir)

	newPath := filepath.Join(dir, ans.String())
	err = os.Rename(file.FilePath, newPath)
	if err != nil {
		ctx.Send(&tgo.SendMessage{Text: "🚫 Failed to update the file.\n\n" + err.Error()})
		return
	}

	ctx.Send(&tgo.SendMessage{Text: "📤 Uploading the new file.\n\n⏳ Please wait..."})
	ctx.Bot.SendChatAction(&tgo.SendChatAction{ChatId: tgo.ID(ctx.Chat.Id), MessageThreadId: ctx.MessageThreadId, Action: "upload_document"})

	_, err = ctx.Send(&tgo.SendDocument{Document: tgo.FileFromPath(newPath), Caption: ctx.Caption})
	if err != nil {
		ctx.Send(&tgo.SendMessage{Text: "🚫 Failed to upload the file.\n\n" + err.Error()})
		return
	}
}

// HasDocument tests if the update is a Message and contains a Document.
func HasDocument() tgo.Filter {
	return filters.NewFilter(func(update *tgo.Update) bool {
		if update.Message != nil {
			return update.Message.Document != nil
		}

		return false
	})
}
