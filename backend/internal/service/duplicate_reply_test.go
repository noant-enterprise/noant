package service

import "testing"

func TestBeginAIReplyDedupesSameMessage(t *testing.T) {
	svc := NewChatService(nil, nil, nil, nil, nil, nil, nil)

	if !svc.beginAIReply("conv-1", "hi") {
		t.Fatal("first reply should be allowed")
	}
	if svc.beginAIReply("conv-1", "hi") {
		t.Fatal("duplicate reply should be blocked while in flight")
	}

	svc.completeAIReply("conv-1", "hi")
	if svc.beginAIReply("conv-1", "hi") {
		t.Fatal("duplicate reply should be blocked immediately after completion")
	}
}
