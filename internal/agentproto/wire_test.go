package agentproto

import (
	"encoding/json"
	"testing"
)

// The same file exists in the other repo (MCP service and API). If this test
// has to change, change it there too: the two sides talk over Redis in exactly
// this JSON.
func TestWireFormatIsStable(t *testing.T) {
	req := Request{RequestID: "r1", UserID: "u1", ProjectID: "p1", PageID: "pg1", Tool: "apply_edits", Args: json.RawMessage(`{"operations":[]}`)}
	b, _ := json.Marshal(req)
	if got, want := string(b), `{"requestId":"r1","userId":"u1","projectId":"p1","pageId":"pg1","tool":"apply_edits","args":{"operations":[]}}`; got != want {
		t.Fatalf("request JSON changed:\n got %s\nwant %s", got, want)
	}
	resp := Response{RequestID: "r1", Error: &Error{Code: CodeInvalidEdit, Message: "nothing was applied", Issues: []Issue{{Path: "operations[0].id", Message: "unknown element"}}}}
	b, _ = json.Marshal(resp)
	if got, want := string(b), `{"requestId":"r1","ok":false,"error":{"code":"invalid_edit","message":"nothing was applied","issues":[{"path":"operations[0].id","message":"unknown element"}]}}`; got != want {
		t.Fatalf("response JSON changed:\n got %s\nwant %s", got, want)
	}
	if RequestChannelPrefix != "agent:req:" || AckChannelPrefix != "agent:ack:" || ResponseChannelPrefix != "agent:res:" {
		t.Fatal("channel names changed")
	}
}
