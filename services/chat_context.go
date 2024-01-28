package services

// chat_context manages chat history in redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

type ChatIdType string
type Prompts map[string]string

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

// wipes redis instance
func (c *ChatService) ClearAllMessages() error {
	return c.rdb.FlushDB(c.ctx).Err()
}

// returns nil if chatId exists. Otherwise, returns an error
func (c *ChatService) AssertChatIdExists(chatId ChatIdType) error {
	exists, err := c.rdb.Exists(c.ctx, string(chatId)).Result()
	if err != nil {
		return fmt.Errorf("checking if '%s' exists: %w", chatId, err)
	}
	if exists == 0 {
		return fmt.Errorf("chatId '%s' doesn't exist", chatId)
	}
	return nil
}

// gets our serializable message format from the redis instance
func (c *ChatService) GetSerialzableMessages(chatId ChatIdType) ([]SerializableMessage, error) {
	// check if key exists
	exists, err := c.rdb.Exists(c.ctx, string(chatId)).Result()
	if err != nil {
		return nil, fmt.Errorf("checking if '%s' exists: %w", chatId, err)
	}
	// return an empty list if nothing exists
	if exists == 0 {
		return []SerializableMessage{}, nil
	}
	// get the key, assuming it exists
	redisHistory, err := c.rdb.Get(c.ctx, string(chatId)).Result()
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
	err = c.rdb.Set(c.ctx, string(chatId), historyStr, c.chatTTL).Err()
	if err != nil {
		return fmt.Errorf("setting key (%s) in redis: %w", chatId, err)
	}
	return nil
}

const charset = "abcdefghijklmnopqrstuvwxyz"

func randomChatId() ChatIdType {
	chatIdLen := 10
	sb := strings.Builder{}
	sb.Grow(chatIdLen)
	for i := 0; i < chatIdLen; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return ChatIdType(sb.String())
}

func (c *ChatService) StartChat(promptName string) (ChatIdType, error) {
	prompt, ok := c.prompts[promptName]
	if !ok {
		return "", errors.New("prompt not found")
	}
	id := randomChatId()
	err := c.AddMessage(id, schema.SystemChatMessage{Content: prompt})
	if err != nil {
		return "", err
	}
	return id, nil
}
