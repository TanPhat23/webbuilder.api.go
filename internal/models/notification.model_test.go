package models

import (
	"encoding/json"
	"testing"
)

func TestNotificationTableName(t *testing.T) {
	if got := (Notification{}).TableName(); got != `public."Notification"` {
		t.Fatalf("table name = %s", got)
	}
}

func TestNotificationJSON(t *testing.T) {
	b, err := json.Marshal(Notification{Id: "n", UserId: "u", Type: "alert", Title: "T", Description: "D", Read: true})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["id"] != "n" || got["userId"] != "u" || got["type"] != "alert" || got["title"] != "T" || got["description"] != "D" || got["read"] != true {
		t.Fatalf("unexpected json: %s", b)
	}
}

func TestNotificationListJSONShape(t *testing.T) {
	list := NotificationList{
		Notifications: []Notification{{Id: "n"}},
		Stats:         NotificationStats{Total: 3, Unread: 2, ReadToday: 1},
		FilterCounts:  NotificationFilterCounts{All: 3, Unread: 2, Messages: 1, Followers: 4, Alerts: 5, Updates: 6},
	}
	b, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Notifications []map[string]any `json:"notifications"`
		Stats         map[string]int64 `json:"stats"`
		FilterCounts  map[string]int64 `json:"filterCounts"`
	}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Notifications) != 1 || got.Stats["total"] != 3 || got.Stats["unread"] != 2 || got.Stats["readToday"] != 1 {
		t.Fatalf("stats wrong: %s", b)
	}
	want := map[string]int64{"all": 3, "unread": 2, "messages": 1, "followers": 4, "alerts": 5, "updates": 6}
	for k, v := range want {
		if got.FilterCounts[k] != v {
			t.Fatalf("filterCounts[%s] = %d, want %d", k, got.FilterCounts[k], v)
		}
	}
}

func TestProjectAccessJSON(t *testing.T) {
	b, err := json.Marshal(ProjectAccess{UserId: "u", ProjectId: "p", Role: CollaboratorRole("editor"), IsCollaborator: true})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	_ = json.Unmarshal(b, &got)
	if got["userId"] != "u" || got["projectId"] != "p" || got["isOwner"] != false || got["isCollaborator"] != true || got["role"] != "editor" {
		t.Fatalf("unexpected json: %s", b)
	}
}
