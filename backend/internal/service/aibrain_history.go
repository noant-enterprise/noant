package service

import (
	"context"
	"encoding/json"
	"fmt"
)

func (b *AIBrain) getConversationHistory(ctx context.Context, conversationID string) ([]MessageTurn, error) {
	if b.redis == nil {
		return nil, nil
	}
	key := fmt.Sprintf("conv:%s:history", conversationID)
	historyJSON, err := b.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var history []MessageTurn
	if err := json.Unmarshal([]byte(historyJSON), &history); err != nil {
		return nil, err
	}
	if len(history) > 10 {
		history = history[len(history)-10:]
	}
	return history, nil
}

func (b *AIBrain) storeConversationTurn(ctx context.Context, conversationID, userQuery, aiResponse string) error {
	if b.redis == nil {
		return nil
	}
	history, err := b.getConversationHistory(ctx, conversationID)
	if err != nil {
		b.logger.Warn("Failed to get conversation history for storage", "error", err)
	}
	history = append(history,
		MessageTurn{Role: "user", Content: userQuery},
		MessageTurn{Role: "assistant", Content: aiResponse},
	)
	if len(history) > 10 {
		history = history[len(history)-10:]
	}
	historyJSON, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}
	return b.redis.Set(ctx, fmt.Sprintf("conv:%s:history", conversationID), string(historyJSON), b.cfg.RedisShortTTL)
}
