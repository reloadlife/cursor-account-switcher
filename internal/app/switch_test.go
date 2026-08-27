package app

import (
	"errors"
	"testing"
)

type stubStarter struct {
	needsRestart bool
	startCalls   int
	startErr     error
	name         string
}

func (s *stubStarter) NeedsRestart() bool { return s.needsRestart }
func (s *stubStarter) Start() error {
	s.startCalls++
	return s.startErr
}
func (s *stubStarter) AppName() string { return s.name }

func TestStartAppIfNeeded(t *testing.T) {
	t.Run("starts when restart required and NoStart is false", func(t *testing.T) {
		stub := &stubStarter{needsRestart: true, name: "Cursor"}
		var steps []string

		err := startAppIfNeeded(stub, SwitchOptions{}, func(label string) {
			steps = append(steps, label)
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stub.startCalls != 1 {
			t.Fatalf("Start calls = %d, want 1", stub.startCalls)
		}
		if len(steps) != 1 || steps[0] != "Starting Cursor..." {
			t.Fatalf("steps = %#v, want [Starting Cursor...]", steps)
		}
	})

	t.Run("skips start when NoStart is true", func(t *testing.T) {
		stub := &stubStarter{needsRestart: true, name: "Cursor"}
		var steps []string

		err := startAppIfNeeded(stub, SwitchOptions{NoStart: true}, func(label string) {
			steps = append(steps, label)
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stub.startCalls != 0 {
			t.Fatalf("Start calls = %d, want 0", stub.startCalls)
		}
		if len(steps) != 0 {
			t.Fatalf("steps = %#v, want none", steps)
		}
	})

	t.Run("skips start when restart not required", func(t *testing.T) {
		stub := &stubStarter{needsRestart: false, name: "Claude Code"}
		var steps []string

		err := startAppIfNeeded(stub, SwitchOptions{}, func(label string) {
			steps = append(steps, label)
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stub.startCalls != 0 {
			t.Fatalf("Start calls = %d, want 0", stub.startCalls)
		}
		if len(steps) != 0 {
			t.Fatalf("steps = %#v, want none", steps)
		}
	})

	t.Run("returns Start error", func(t *testing.T) {
		startErr := errors.New("start failed")
		stub := &stubStarter{needsRestart: true, name: "Cursor", startErr: startErr}

		err := startAppIfNeeded(stub, SwitchOptions{}, nil)
		if !errors.Is(err, startErr) {
			t.Fatalf("error = %v, want %v", err, startErr)
		}
	})
}
