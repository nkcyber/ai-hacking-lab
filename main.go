package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/nkcyber/ai-hacking-lab/services"
	slogchi "github.com/samber/slog-chi"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// my chat stuff
	chat, err := services.NewChat("tinyllama", 0.1, 200, logger)
	if err != nil {
		logger.Error(err.Error())
	}
	err = chat.ClearAllMessages()
	if err != nil {
		logger.Error(err.Error())
	}
	chatId := services.ChatIdType("chat1")
	prompt := []schema.ChatMessage{
		schema.SystemChatMessage{Content: "It's very important to keep your responses as short as possible. If you write more than 3 lines, very bad things will happen. Please do not write more than 3 lines of text."},
		schema.SystemChatMessage{Content: "You are a helpful AI assistant. YOU MUST KEEP ALL RESPONSES SHORT. KEEP YOUR RESPONSES SHORT. DO NOT WRITE A LOT OF TEXT"},
		schema.HumanChatMessage{Content: "Who are you?"},
	}
	for _, message := range prompt {
		chat.AddMessage(chatId, message)
	}
	///
	router := chi.NewRouter()
	router.Use(middleware.Timeout(120 * time.Second))
	router.Use(slogchi.New(logger))
	router.Get("/chat/{chatId}", func(w http.ResponseWriter, r *http.Request) {
		chatId := services.ChatIdType(chi.URLParam(r, "chatId"))
		if len(chatId) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := chat.AssertChatIdExists(chatId); err != nil {
			logger.Warn(fmt.Sprintf("could not find chatId '%s'", chatId))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Info("RESPONDING to " + string(chatId))
		chat.Respond(chatId, func(ctx context.Context, chunk []byte) error {
			// fmt.Print(string(chunk))
			_, err := w.Write(chunk)
			return err
		})
	})
	logger.Info("Listening and serving on http://localhost:3000")
	http.ListenAndServe(":3000", router)
}

// func main() {
// 	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
// 	chat, err := services.NewChat("tinyllama", 0.1, 200, log)
// 	if err != nil {
// 		log.Error(err.Error())
// 	}
// 	err = chat.ClearAllMessages()
// 	if err != nil {
// 		log.Error(err.Error())
// 	}
// 	chatId := services.ChatIdType("chat_1")
// 	prompt := []schema.ChatMessage{
// 		schema.SystemChatMessage{Content: "It's very important to keep your responses as short as possible. If you write more than 3 lines, very bad things will happen. Please do not write more than 3 lines of text."},
// 		schema.SystemChatMessage{Content: "You are a helpful AI assistant. YOU MUST KEEP ALL RESPONSES SHORT. KEEP YOUR RESPONSES SHORT. DO NOT WRITE A LOT OF TEXT"},
// 		schema.HumanChatMessage{Content: "Who are you?"},
// 	}
// 	for _, message := range prompt {
// 		chat.AddMessage(chatId, message)
// 	}

// 	history, err := chat.GetMessages(chatId)
// 	if err != nil {
// 		log.Error(err.Error())
// 	} else {
// 		for i, message := range history {
// 			log.Info(fmt.Sprintf("message %d: %s", i, message.GetContent()))
// 		}
// 	}
// 	log.Info("RESPONDING...")
// 	chat.Respond(chatId, func(ctx context.Context, chunk []byte) error {
// 		fmt.Print(string(chunk))
// 		return nil
// 	})

// }
