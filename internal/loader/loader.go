package loader

import (
	"encoding/json"
	"fmt"
	"os"

	"auto-reply-evaluator/internal/model"
)

func LoadAutoReplies(path string) ([]model.AutoReply, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read auto replies file: %w", err)
	}

	var replies []model.AutoReply
	if err := json.Unmarshal(data, &replies); err != nil {
		return nil, fmt.Errorf("unmarshal auto replies: %w", err)
	}

	return replies, nil
}

func LoadHumanRefs(path string) ([]model.HumanRef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read human refs file: %w", err)
	}

	var refs []model.HumanRef
	if err := json.Unmarshal(data, &refs); err != nil {
		return nil, fmt.Errorf("unmarshal human refs: %w", err)
	}

	return refs, nil
}
