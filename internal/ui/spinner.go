package ui

import (
	"fmt"
	"io"
	"sync"
	"time"
)

var frames = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")

// Spinner displays an animated spinner with a status message.
type Spinner struct {
	out     io.Writer
	noColor bool
	done    chan struct{}
	mu      sync.Mutex
	msg     string
}

// NewSpinner creates a new Spinner that writes to out.
func NewSpinner(out io.Writer, noColor bool) *Spinner {
	return &Spinner{out: out, noColor: noColor}
}

// Start begins the spinner animation with the given message.
func (s *Spinner) Start(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.done = make(chan struct{})
	s.mu.Unlock()

	go s.tick()
}

// Update changes the spinner message while it's running.
func (s *Spinner) Update(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()
}

// Stop halts the spinner and prints a success line.
func (s *Spinner) Stop(msg string) {
	s.stop()
	prefix := "\033[32m✓\033[0m"
	if s.noColor {
		prefix = "* done:"
	}
	_, _ = fmt.Fprintf(s.out, "\r\033[K%s %s\n", prefix, msg)
}

// Warn halts the spinner and prints a warning line.
func (s *Spinner) Warn(msg string) {
	s.stop()
	prefix := "\033[33m⚠\033[0m"
	if s.noColor {
		prefix = "* warn:"
	}
	_, _ = fmt.Fprintf(s.out, "\r\033[K%s %s\n", prefix, msg)
}

func (s *Spinner) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done != nil {
		close(s.done)
		s.done = nil
	}
}

func (s *Spinner) tick() {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	i := 0
	for {
		s.mu.Lock()
		done := s.done
		msg := s.msg
		s.mu.Unlock()

		frame := string(frames[i%len(frames)])
		if s.noColor {
			frame = "*"
		}
		_, _ = fmt.Fprintf(s.out, "\r\033[K%s %s", frame, msg)

		i++
		select {
		case <-done:
			return
		case <-ticker.C:
		}
	}
}
