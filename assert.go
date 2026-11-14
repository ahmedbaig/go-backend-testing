package testkit

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func NoError(t testing.TB, err error, msg ...string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%sunexpected error: %v", prefix(msg), err)
	}
}

func ErrorIs(t testing.TB, err, target error, msg ...string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("%serror %v is not %v", prefix(msg), err, target)
	}
}

func Equal[T any](t testing.TB, got, want T, msg ...string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%sgot  %#v\nwant %#v", prefix(msg), got, want)
	}
}

func True(t testing.TB, cond bool, msg ...string) {
	t.Helper()
	if !cond {
		t.Fatalf("%sexpected true", prefix(msg))
	}
}

func Contains(t testing.TB, s, sub string, msg ...string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("%s%q does not contain %q", prefix(msg), s, sub)
	}
}

func NotContains(t testing.TB, s, sub string, msg ...string) {
	t.Helper()
	if strings.Contains(s, sub) {
		t.Fatalf("%s%q should not contain %q", prefix(msg), s, sub)
	}
}

func Eventually(t testing.TB, timeout time.Duration, cond func() bool, msg ...string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if cond() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%scondition not met within %s", prefix(msg), timeout)
		}
		time.Sleep(timeout / 50)
	}
}

func prefix(msg []string) string {
	if len(msg) == 0 {
		return ""
	}
	return strings.Join(msg, " ") + ": "
}
