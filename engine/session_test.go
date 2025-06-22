package engine

import "testing"

func TestSessionManager_VariablesPersist(t *testing.T) {
	mgr := NewSessionManager()
	user := "user1"

	mgr.SetVar(user, "foo", "bar")
	if got := mgr.GetVar(user, "foo"); got != "bar" {
		t.Errorf("Expected 'bar', got '%s'", got)
	}

	mgr.SetVar(user, "foo", "baz")
	if got := mgr.GetVar(user, "foo"); got != "baz" {
		t.Errorf("Expected 'baz', got '%s'", got)
	}
}

func TestSessionManager_TopicAndThatPersist(t *testing.T) {
	mgr := NewSessionManager()
	user := "user2"

	if topic := mgr.GetTopic(user); topic != "*" {
		t.Errorf("Expected default topic '*', got '%s'", topic)
	}
	if that := mgr.GetThat(user); that != "*" {
		t.Errorf("Expected default that '*', got '%s'", that)
	}

	mgr.UpdateTopic(user, "JOKES")
	mgr.UpdateThat(user, "Why did the chicken cross the road?")

	if topic := mgr.GetTopic(user); topic != "JOKES" {
		t.Errorf("Expected topic 'JOKES', got '%s'", topic)
	}
	if that := mgr.GetThat(user); that != "Why did the chicken cross the road?" {
		t.Errorf("Expected updated that, got '%s'", that)
	}
}

func TestSessionManager_MultipleUsers(t *testing.T) {
	mgr := NewSessionManager()
	userA := "alice"
	userB := "bob"

	mgr.SetVar(userA, "color", "red")
	mgr.SetVar(userB, "color", "blue")

	if got := mgr.GetVar(userA, "color"); got != "red" {
		t.Errorf("Expected 'red' for alice, got '%s'", got)
	}
	if got := mgr.GetVar(userB, "color"); got != "blue" {
		t.Errorf("Expected 'blue' for bob, got '%s'", got)
	}

	mgr.UpdateTopic(userA, "SPORTS")
	mgr.UpdateTopic(userB, "MUSIC")

	if topic := mgr.GetTopic(userA); topic != "SPORTS" {
		t.Errorf("Expected topic 'SPORTS' for alice, got '%s'", topic)
	}
	if topic := mgr.GetTopic(userB); topic != "MUSIC" {
		t.Errorf("Expected topic 'MUSIC' for bob, got '%s'", topic)
	}
} 