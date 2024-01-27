package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
)

type ChatIdType string

type ChatService struct {
	Ctx context.Context
	Llm *ollama.Chat
	log *slog.Logger
	// There's community support for Redis chat history for TypeScript, but not Go.
	// So, we have to roll our own, for the moment.
	// TypeScript suppport:
	//   <https://api.js.langchain.com/classes/langchain_community_stores_message_ioredis.RedisChatMessageHistory.html>
	// We just store the array of messages in Redis according to the chatId.
	// It's not a fancy solution, but this is not a fancy app.
	rdb     *redis.Client
	chatTTL time.Duration
	temp    float64
}

type SerializableMessage struct {
	Content string                 `json:"content"`
	Type    schema.ChatMessageType `json:"type"`
}

// implement schema.ChatMessage interface
func (sm SerializableMessage) GetContent() string {
	return sm.Content
}

// implement schema.ChatMessage interface
func (sm SerializableMessage) GetType() schema.ChatMessageType {
	return sm.Type
}

// initalizes ollama & redis
func NewChat(model_name string, log *slog.Logger) (ChatService, error) {
	llm, err := ollama.NewChat(ollama.WithLLMOptions(
		ollama.WithModel(model_name), ollama.WithPredictPenalizeNewline(true)))
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
		temp:    1,
		log:     log,
	}, nil
}

// wipes redis instance
func (c *ChatService) ClearAllMessages() error {
	return c.rdb.FlushDB(c.Ctx).Err()
}

// gets our serializable message format from the redis instance
func (c *ChatService) GetSerialzableMessages(chatId ChatIdType) ([]SerializableMessage, error) {
	// check if key exists
	exists, err := c.rdb.Exists(c.Ctx, string(chatId)).Result()
	if err != nil {
		return nil, fmt.Errorf("checking if '%s' exists: %w", chatId, err)
	}
	// return an empty list if nothing exists
	if exists == 0 {
		return []SerializableMessage{}, nil
	}
	// get the key, assuming it exists
	redisHistory, err := c.rdb.Get(c.Ctx, string(chatId)).Result()
	if err != nil {
		return nil, fmt.Errorf("getting '%s' from redis: %w", chatId, err)
	}
	// unmarshal redis contents
	var history []SerializableMessage
	err = json.Unmarshal([]byte(redisHistory), &history)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling history for %s: %w", chatId, err)
	}
	return history, nil
}

// gets the official message format from the redis instance
func (c *ChatService) GetMessages(chatId ChatIdType) ([]schema.ChatMessage, error) {
	history, err := c.GetSerialzableMessages(chatId)
	if err != nil {
		return nil, err
	}
	// convert from slice of structs to slice of interfaces
	chatMessages := make([]schema.ChatMessage, len(history))
	for i, m := range history {
		chatMessages[i] = m
	}
	return chatMessages, nil
}

// adds a message to the Redis instance
func (c *ChatService) AddMessage(chatId ChatIdType, message schema.ChatMessage) error {
	c.log.Info("adding message to " + string(chatId))
	history, err := c.GetSerialzableMessages(chatId)
	if err != nil {
		return fmt.Errorf("adding message: %w", err)
	}

	history = append(history, SerializableMessage{
		Content: message.GetContent(),
		Type:    message.GetType(),
	})

	historyStr, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("marshalling history: %w", err)
	}
	err = c.rdb.Set(c.Ctx, string(chatId), historyStr, c.chatTTL).Err()
	if err != nil {
		return fmt.Errorf("setting key (%s) in redis: %w", chatId, err)
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
