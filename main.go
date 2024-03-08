package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/haashemi/tgo"
	"github.com/haashemi/tgo/filters"
	"github.com/haashemi/tgo/routers/message"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Host   string  `yaml:"host"`
	Token  string  `yaml:"token"`
	Admins []int64 `yaml:"admins"`
}

func main() {
	var config Config
	if f, err := os.ReadFile("config.yaml"); err != nil {
		panic(err)
	} else if err = yaml.Unmarshal(f, &config); err != nil {
		panic(err)
	}

	bot := tgo.NewBot(config.Token, tgo.Options{Host: config.Host})

	mr := message.NewRouter()
	mr.Handle(filters.And(HasDocument(), filters.Whitelist(config.Admins...)), OnDocument)
	bot.AddRouter(mr)

	for {
		fmt.Println("Running...")
		if err := bot.StartPolling(30); err != nil {
			fmt.Println("Polling failed", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func OnDocument(ctx *message.Context) {
	ctx.Send(&tgo.SendMessage{Text: "📥 Downloading the file.\n\n⏳ Please wait..."})

	file, err := ctx.Bot.GetFile(&tgo.GetFile{FileId: ctx.Document.FileId})
	if err != nil {
		ctx.Send(&tgo.SendMessage{Text: "🚫 Failed to GetFile.\n\n" + err.Error()})
		return
	}

	_, ans, err := ctx.Ask(&tgo.SendMessage{Text: "✍️ Send the new file name."}, 5*time.Minute)
	if err != nil {
		ctx.Send(&tgo.SendMessage{Text: "🚫 Failed to get the new file name.\n\n" + err.Error()})
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

	ctx.Bot.SendChatAction(&tgo.SendChatAction{ChatId: tgo.ID(ctx.Chat.Id), MessageThreadId: ctx.MessageThreadId, Action: "upload_document"})

	_, err = ctx.Send(&tgo.SendDocument{Document: tgo.FileFromPath(newPath), Caption: ctx.Caption})
	if err != nil {
		ctx.Send(&tgo.SendMessage{Text: "🚫 Failed to upload the file.\n\n" + err.Error()})
		return
	}
}

func HasDocument() tgo.Filter {
	return filters.NewFilter(func(update *tgo.Update) bool {
		if update.Message != nil {
			return update.Message.Document != nil
		}

		return false
	})
}
