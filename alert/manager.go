package alert

import (
	"fmt"
	"log/slog"

	"github.com/boar-d-white-foundation/drone/config"
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

func NewManagerFromConfig(cfg config.Config) (*Manager, error) {
	telegram, err := tg.NewAdminClientFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create tg admin client: %w", err)
	}

	return NewManager(telegram), nil
}

func (m *Manager) Errorxf(err error, msg string, args ...any) {
	errMsg := fmt.Sprintf(msg, args...)
	errMsg = fmt.Sprintf("Error:\n%s\n\n%s", errMsg, err.Error())
	chunkLen := 4096
	for i := 0; i < len(errMsg); i += chunkLen {
		chunk := errMsg[i:min(i+chunkLen, len(errMsg))]
		if _, err := m.telegram.SendMonospace(0, chunk); err != nil {
			slog.Error(
				"err sending alert chunk",
				slog.Any("err", err), slog.String("msg", errMsg), slog.String("chunk", chunk),
			)
		}
	}
}
