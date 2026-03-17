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
	exited  chan struct{}
	mu      sync.Mutex
	msg     string
}

// NewSpinner creates a new Spinner that writes to out.
func NewSpinner(out io.Writer, noColor bool) *Spinner {
	return &Spinner{out: out, noColor: noColor}
}

// Start begins the spinner animation with the given message.
func (s *Spinner) Start(msg string) {
	done := make(chan struct{})
	exited := make(chan struct{})

	s.mu.Lock()
	s.msg = msg
	s.done = done
	s.exited = exited
	s.mu.Unlock()

	go s.tick(done, exited)
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
	s.mu.Lock()
	_, _ = fmt.Fprintf(s.out, "\r\033[K%s %s\n", prefix, msg)
	s.mu.Unlock()
}

// Warn halts the spinner and prints a warning line.
func (s *Spinner) Warn(msg string) {
	s.stop()
	prefix := "\033[33m⚠\033[0m"
	if s.noColor {
		prefix = "* warn:"
	}
	s.mu.Lock()
	_, _ = fmt.Fprintf(s.out, "\r\033[K%s %s\n", prefix, msg)
	s.mu.Unlock()
}

func (s *Spinner) stop() {
	s.mu.Lock()
	done := s.done
	exited := s.exited
	s.done = nil
	s.exited = nil
	s.mu.Unlock()

	if done != nil {
		close(done)
	}
	if exited != nil {
		<-exited
	}
}

func (s *Spinner) tick(done <-chan struct{}, exited chan<- struct{}) {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	defer close(exited)

	for i := 0; ; i++ {
		s.mu.Lock()
		msg := s.msg
		frame := string(frames[i%len(frames)])
		if s.noColor {
			frame = "*"
		}
		_, _ = fmt.Fprintf(s.out, "\r\033[K%s %s", frame, msg)
		s.mu.Unlock()

		select {
		case <-done:
			return
		case <-ticker.C:
		}
	}
}
