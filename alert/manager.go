package alert

import (
	"fmt"
	"log/slog"

	"github.com/boar-d-white-foundation/drone/tg"
)

type Manager struct {
	telegram tg.Client
}

func NewManager(
	telegram tg.Client,
) *Manager {
	return &Manager{
		telegram: telegram,
	}
}

func (m *Manager) send(msg string) {
	chunkLen := 4096
	for i := 0; i < len(msg); i += chunkLen {
		chunk := msg[i:min(i+chunkLen, len(msg))]
		if _, err := m.telegram.SendMonospace(0, chunk); err != nil {
			slog.Error(
				"err sending alert chunk",
				slog.Any("err", err), slog.String("msg", msg), slog.String("chunk", chunk),
			)
		}
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
