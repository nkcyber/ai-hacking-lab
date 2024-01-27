package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nkcyber/ai-hacking-lab/services"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	chat, err := services.NewChat("tinyllama", 0, log)
	if err != nil {
		log.Error(err.Error())
	}
	err = chat.ClearAllMessages()
	if err != nil {
		log.Error(err.Error())
	}
	chatId := services.ChatIdType("chat_1")
	prompt := []schema.ChatMessage{
		schema.SystemChatMessage{Content: "It's very important to keep your responses as short as possible. If you write more than 3 lines, very bad things will happen. Please do not write more than 3 lines of text."},
		schema.SystemChatMessage{Content: "You are a hyper-creative rhyming machine. KEEP ALL RESPONSES SHORTER THAN THREE LINES. KEEP YOUR RESPONSES SHORT. DO NOT WRITE A LOT OF TEXT"},
		schema.HumanChatMessage{Content: "Tell me a poem about Trees"},
	}
	for _, message := range prompt {
		chat.AddMessage(chatId, message)
	}

	history, err := chat.GetMessages(chatId)
	if err != nil {
		log.Error(err.Error())
	} else {
		for i, message := range history {
			log.Info(fmt.Sprintf("message %d: %s", i, message.GetContent()))
		}
	}
	log.Info("RESPONDING...")
	chat.Respond(chatId, func(ctx context.Context, chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	})

	// log.Info("running Llm.Call")
	// completion, err := chat.Llm.Call(chat.Ctx, []schema.ChatMessage{
	// 	schema.SystemChatMessage{Content: "It's very important to keep your responses as short as possible. If you write more than 3 lines, very bad things will happen. Please do not write more than 3 lines of text."},
	// 	schema.SystemChatMessage{Content: "You are a hyper-creative rhyming machine. KEEP ALL RESPONSES SHORTER THAN THREE LINES. KEEP YOUR RESPONSES SHORT. DO NOT WRITE A LOT OF TEXT"},
	// 	schema.HumanChatMessage{Content: "Tell me a poem about Trees"},
	// }, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
	// 	fmt.Print(string(chunk))
	// 	return nil
	// }),
	// )
	// if err != nil {
	// 	log.Error(err.Error())
	// }
	// _ = completion

	// ORIGINAL
	// llm, err := ollama.NewChat(ollama.WithLLMOptions(ollama.WithModel("orca-mini")))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// ctx := context.Background()
	// completion, err := llm.Call(ctx, []schema.ChatMessage{
	// 	schema.SystemChatMessage{Content: "You are a hyper-creative rhyming machine."},
	// 	schema.HumanChatMessage{Content: "Tell me a poem about Sam Altman"},
	// }, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
	// 	fmt.Print(string(chunk))
	// 	return nil
	// }),
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _ = completion
}
