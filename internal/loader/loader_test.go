package loader

import (
	"testing"

	"auto-reply-evaluator/internal/model"
)

func TestLoadAutoReplies(t *testing.T) {
	replies, err := LoadAutoReplies("../../task/task3_auto_replies.json")
	if err != nil {
		t.Fatalf("failed to load auto replies: %v", err)
	}

	if len(replies) != 20 {
		t.Errorf("expected 20 auto replies, got %d", len(replies))
	}

	for _, r := range replies {
		if r.ID == "" {
			t.Error("auto reply has empty ID")
		}
		if r.UserQuestion == "" {
			t.Errorf("auto reply %s has empty UserQuestion", r.ID)
		}
		if r.AutoReply == "" {
			t.Errorf("auto reply %s has empty AutoReply", r.ID)
		}
	}
}

func TestLoadHumanRefs(t *testing.T) {
	refs, err := LoadHumanRefs("../../task/task3_human_ref.json")
	if err != nil {
		t.Fatalf("failed to load human refs: %v", err)
	}

	if len(refs) != 20 {
		t.Errorf("expected 20 human refs, got %d", len(refs))
	}

	for _, r := range refs {
		if r.ID == "" {
			t.Error("human ref has empty ID")
		}
		if r.HumanReference == "" {
			t.Errorf("human ref %s has empty HumanReference", r.ID)
		}
		if r.AnnotatorNotes == "" {
			t.Errorf("human ref %s has empty AnnotatorNotes", r.ID)
		}
	}
}

func TestLoadAutoReplies_FileNotFound(t *testing.T) {
	_, err := LoadAutoReplies("nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadHumanRefs_FileNotFound(t *testing.T) {
	_, err := LoadHumanRefs("nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestDataConsistency(t *testing.T) {
	replies, err := LoadAutoReplies("../../task/task3_auto_replies.json")
	if err != nil {
		t.Fatalf("failed to load auto replies: %v", err)
	}
	refs, err := LoadHumanRefs("../../task/task3_human_ref.json")
	if err != nil {
		t.Fatalf("failed to load human refs: %v", err)
	}

	replyIDs := make(map[string]*model.AutoReply)
	for i := range replies {
		replyIDs[replies[i].ID] = &replies[i]
	}

	for _, ref := range refs {
		if _, ok := replyIDs[ref.ID]; !ok {
			t.Errorf("human ref %s has no matching auto reply", ref.ID)
		}
	}
}
