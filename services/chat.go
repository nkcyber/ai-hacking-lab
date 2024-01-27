package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type ChatService struct {
	// There's community support for Redis chat history for TypeScript, but not Go.
	// So, we have to roll our own, for the moment.
	// TypeScript suppport:
	//   <https://api.js.langchain.com/classes/langchain_community_stores_message_ioredis.RedisChatMessageHistory.html>
	// We just store the array of messages in Redis according to the chatId.
	// It's not a fancy solution, but this is not a fancy app.
	rdb     *redis.Client
	Llm     *ollama.Chat
	log     *slog.Logger
	Ctx     context.Context
	chatTTL time.Duration // used when adding keys to redis
	temp    float64       // passed to LLM options
	maxLen  int           // max len in tokens - passed to LLM options
}

// initalizes ollama & redis
func NewChat(modelName string, modelTemp float64, maxLen int, log *slog.Logger) (ChatService, error) {
	llm, err := ollama.NewChat(ollama.WithLLMOptions(
		ollama.WithModel(modelName), ollama.WithPredictPenalizeNewline(true)))
	if err != nil {
		return ChatService{}, fmt.Errorf("initalizing chat: %w", err)
	}
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	log.Info("creating chat sevice")
	return ChatService{
		Ctx:     ctx,
		Llm:     llm,
		rdb:     rdb,
		chatTTL: 10 * time.Minute,
		temp:    modelTemp,
		log:     log,
		maxLen:  maxLen,
	}, nil
}

func (c *ChatService) Respond(chatId ChatIdType, callback func(ctx context.Context, chunk []byte) error) error {
	messageHistory, err := c.GetMessages(chatId)
	if err != nil {
		return err
	}
	completion, err := c.Llm.Call(c.Ctx, messageHistory,
		llms.WithStreamingFunc(callback),
		llms.WithTemperature(c.temp),
		llms.WithMaxTokens(c.maxLen),
	)
	if err != nil {
		return fmt.Errorf("llm response: %w", err)
	}
	err = c.AddMessage(chatId, completion)
	if err != nil {
		return err
	}
	return nil
}

// TODO: load prompts
// TODO: func (c *ChatService) NewChat(prompt PromptId) (ChatId, err)
// TODO: namespace redis keys: message_ vs prompt_

// TODO: figure out return type
// func (c *ChatService) RespondToUserMessage(string) {
// }
// TODO: make chat service after I get message history figured out
