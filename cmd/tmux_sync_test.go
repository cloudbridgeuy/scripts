package cmd

import (
	"reflect"
	"testing"
)

func TestReverseSyncPlan(t *testing.T) {
	t.Parallel()

	history := []string{"/a", "web1:/a/b", "/b.c", "/c", "web2:/x"}
	sessions := []string{"/a", "web1_/a/b", "/b_c", "/d"}

	toCreate, toKill := reverseSyncPlan(history, sessions)

	if !reflect.DeepEqual(toCreate, []string{"/c", "web2:/x"}) {
		t.Fatalf("unexpected toCreate: %#v", toCreate)
	}

	if !reflect.DeepEqual(toKill, []string{"/d"}) {
		t.Fatalf("unexpected toKill: %#v", toKill)
	}
}

func TestLastSession(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		session, ok := lastSession(nil)
		if ok {
			t.Fatalf("expected no session, got %q", session)
		}
	})

	t.Run("returns last", func(t *testing.T) {
		t.Parallel()

		session, ok := lastSession([]string{"one", "two", "three"})
		if !ok {
			t.Fatalf("expected session")
		}

		if session != "three" {
			t.Fatalf("expected three, got %q", session)
		}
	})
}

func TestRotateHistoryPrev(t *testing.T) {
	t.Parallel()

	rotated := rotateHistoryPrev([]string{"a", "b", "c"})
	if !reflect.DeepEqual(rotated, []string{"c", "a", "b"}) {
		t.Fatalf("unexpected prev rotation: %#v", rotated)
	}
}

func TestRotateHistoryNext(t *testing.T) {
	t.Parallel()

	rotated := rotateHistoryNext([]string{"a", "b", "c"})
	if !reflect.DeepEqual(rotated, []string{"b", "c", "a"}) {
		t.Fatalf("unexpected next rotation: %#v", rotated)
	}
}
