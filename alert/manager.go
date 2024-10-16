package alert

import (
	"fmt"
	"log/slog"
)

type Sender interface {
	SendAlert(msg string) error
}

type Manager struct {
	sender Sender
}

func NewManager(
	sender Sender,
) *Manager {
	return &Manager{
		sender: sender,
	}
}

func (m *Manager) send(msg string) {
	slog.Error("sending alert", slog.String("msg", msg))
	if err := m.sender.SendAlert(msg); err != nil {
		slog.Error("err sending alert", slog.String("msg", msg), slog.Any("err", err))
	}
}

func (m *Manager) Errorf(msg string, args ...any) {
	errMsg := fmt.Sprintf(msg, args...)
	m.send(errMsg)
}

func (m *Manager) Errorxf(err error, msg string, args ...any) {
	errMsg := fmt.Sprintf(msg, args...)
	errMsg = fmt.Sprintf("Error:\n%s\n\n%s", errMsg, err.Error())
	m.send(errMsg)
}
