package services

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/tmc/langchaingo/llms/ollama"
)

type ChatService struct {
	ctx context.Context
	rdb *redis.Client
	llm *ollama.Chat
}

func NewChatService(model_name string) (ChatService, error) {
	// initalize ollama connection
	llm, err := ollama.NewChat(ollama.WithLLMOptions(ollama.WithModel(model_name)))
	if err != nil {
		return ChatService{}, fmt.Errorf("initalizing chat: %w", err)
	}
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return ChatService{
		ctx: ctx,
		rdb: rdb,
		llm: llm,
	}, nil
}

// TODO: load prompts
// TODO: func (c *ChatService) NewChat(prompt PromptId) (ChatId, err)
// TODO: namespace redis keys: message_ vs prompt_

// TODO: figure out return type
// func (c *ChatService) RespondToUserMessage(string) {
// }
// TODO: make chat service after I get message history figured out
