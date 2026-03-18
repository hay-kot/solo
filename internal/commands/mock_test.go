package commands

import (
	"context"
	"fmt"

	"github.com/hay-kot/solo/internal/tmux"
)

type sendKeysCall struct {
	Target string
	Keys   []string
}

type mockClient struct {
	inTmux       bool
	windows      []tmux.Window
	currentWin   string
	createdWins  []string
	sentKeys     []sendKeysCall
	selectedWins []string
	killedWins   []string
	nextWindowID int // auto-increments to generate window IDs

	// listPaneCommandFn allows tests to control ListPaneCommand behavior.
	// When nil, returns ("", nil).
	listPaneCommandFn func(target string) (string, error)
}

func (m *mockClient) InTmux() bool { return m.inTmux }

func (m *mockClient) CurrentWindow(_ context.Context) (string, error) {
	return m.currentWin, nil
}

func (m *mockClient) ListWindows(_ context.Context) ([]tmux.Window, error) {
	return m.windows, nil
}

func (m *mockClient) NewWindow(_ context.Context, name string) (string, error) {
	m.createdWins = append(m.createdWins, name)
	id := fmt.Sprintf("@%d", m.nextWindowID)
	m.nextWindowID++
	return id, nil
}

func (m *mockClient) SendKeys(_ context.Context, target string, keys ...string) error {
	m.sentKeys = append(m.sentKeys, sendKeysCall{Target: target, Keys: keys})
	return nil
}

func (m *mockClient) SelectWindow(_ context.Context, target string) error {
	m.selectedWins = append(m.selectedWins, target)
	return nil
}

func (m *mockClient) ListPaneCommand(_ context.Context, target string) (string, error) {
	if m.listPaneCommandFn != nil {
		return m.listPaneCommandFn(target)
	}
	return "", nil
}

func (m *mockClient) KillWindow(_ context.Context, target string) error {
	m.killedWins = append(m.killedWins, target)
	return nil
}
