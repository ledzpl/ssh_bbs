package server

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"ag/internal/bbs"
	"ag/internal/ui"
)

// New creates a new SSH server configured with the BBS application.
func New(addr string, hostKeyPath string, board *bbs.BBS) (*ssh.Server, error) {
	s, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler(board)),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func teaHandler(board *bbs.BBS) func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		_, _, active := s.Pty()
		if !active {
			fmt.Println("no active terminal, skipping")
			return nil, nil
		}

		username := s.User()
		if username == "" {
			username = "guest"
		}

		m := ui.NewModel(board, username)
		return m, []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
	}
}
