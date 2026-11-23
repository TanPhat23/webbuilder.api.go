package entity

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestTableNames(t *testing.T) {
	tests := []struct {
		model interface{ TableName() string }
		want  string
	}{
		{ElementComment{}, "ElementComment"},
		{Element{}, "Element"},
		{Page{}, "Page"},
		{User{}, "User"},
		{Collaborator{}, "Collaborator"},
		{Invitation{}, "Invitation"},
		{Project{}, "Project"},
		{Snapshot{}, "Snapshot"},
		{EventWorkflow{}, "EventWorkflow"},
		{ElementEventWorkflow{}, "ElementEventWorkflow"},
	}
	for _, tt := range tests {
		if got := tt.model.TableName(); got != tt.want {
			t.Errorf("%T.TableName()=%q want %q", tt.model, got, tt.want)
		}
	}
}

func TestNewMessage(t *testing.T) {
	before := time.Now().UnixMilli()
	msg, err := NewMessage(TypeElementUpdate, "proj", "page", map[string]any{"id": "e1", "n": 2})
	after := time.Now().UnixMilli()
	if err != nil {
		t.Fatal(err)
	}
	if msg.Type != TypeElementUpdate || msg.ProjectID != "proj" || msg.PageID != "page" {
		t.Fatalf("envelope mismatch: %+v", msg)
	}
	if msg.Timestamp < before || msg.Timestamp > after {
		t.Fatalf("timestamp %d not in [%d,%d]", msg.Timestamp, before, after)
	}
	var got map[string]any
	if err := json.Unmarshal(msg.Payload, &got); err != nil {
		t.Fatalf("payload not valid JSON: %v", err)
	}
	if got["id"] != "e1" || got["n"] != float64(2) {
		t.Fatalf("payload mismatch: %v", got)
	}
	if msg.UserID != "" || msg.Seq != 0 || msg.ClientOpID != "" || msg.RequestID != "" {
		t.Fatalf("optional fields must be zero: %+v", msg)
	}
}

func TestNewMessageNilPayloadIsJSONNull(t *testing.T) {
	msg, err := NewMessage(TypeLeave, "p", "g", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(msg.Payload) != "null" {
		t.Fatalf("payload=%q", msg.Payload)
	}
}

func TestNewMessageMarshalError(t *testing.T) {
	msg, err := NewMessage(TypeError, "p", "g", make(chan int))
	if err == nil || msg != nil {
		t.Fatalf("expected marshal error and nil message, got %v, %v", msg, err)
	}
	var ute *json.UnsupportedTypeError
	if !errors.As(err, &ute) {
		t.Fatalf("unexpected error type %T", err)
	}
}
