package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var tuiCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start an interactive game session",
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		token, _ := cmd.Flags().GetString("token")

		if token == "" {
			return fmt.Errorf("token is required (use --token)")
		}

		conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()

		p := tea.NewProgram(initialModel(conn, token), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run TUI: %v", err)
		}
		return nil
	},
}

func init() {
	playCmd.AddCommand(tuiCmd)
}

type sessionItem struct {
	id string
}

func (i sessionItem) Title() string       { return "Session " + i.id }
func (i sessionItem) Description() string { return "Join this session" }
func (i sessionItem) FilterValue() string { return i.id }

type state int

const (
	stateSelectingSession state = iota
	stateInGame
)

type model struct {
	conn    *grpc.ClientConn
	token   context.Context
	state   state
	list    list.Model
	session *pb.Session
	game    *pb.Game
	hand    []*pb.Card
	err     error
	width   int
	height  int
}

func initialModel(conn *grpc.ClientConn, token string) model {
	md := metadata.Pairs("authorization", token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a Session"

	return model{
		conn:  conn,
		token: ctx,
		state: stateSelectingSession,
		list:  l,
	}
}

func (m model) Init() tea.Cmd {
	return m.fetchSessions()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.state == stateSelectingSession {
				item, ok := m.list.SelectedItem().(sessionItem)
				if ok {
					return m, m.joinSession(item.id)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)

	case sessionsMsg:
		m.list.SetItems(msg.items)
		return m, nil

	case sessionJoinedMsg:
		m.session = msg.session
		m.state = stateInGame
		return m, tea.Batch(m.fetchGame(), m.fetchHand())

	case gameMsg:
		m.game = msg.game
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return tickMsg{}
		})

	case handMsg:
		m.hand = msg.cards
		return m, nil

	case tickMsg:
		if m.state == stateInGame {
			return m, tea.Batch(m.fetchGame(), m.fetchHand())
		}

	case errorMsg:
		m.err = msg.err
	}

	var cmd tea.Cmd
	if m.state == stateSelectingSession {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

var (
	titleStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("205")).Bold(true)
	infoStyle  = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("250"))
	cardStyle  = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63"))
	scoreStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
	}

	if m.state == stateSelectingSession {
		return "\n" + m.list.View()
	}

	if m.game == nil {
		return "\n  Loading game..."
	}

	var s strings.Builder
	s.WriteString(titleStyle.Render("RUTHLESS - " + m.game.State.String()))
	s.WriteString("\n\n")

	// Scores
	s.WriteString(infoStyle.Render("Scores:"))
	s.WriteString("\n")
	for playerID, score := range m.game.Scores {
		playerName := playerID
		for _, p := range m.session.Players {
			if p.Id == playerID {
				playerName = p.Name
				break
			}
		}
		s.WriteString(fmt.Sprintf("  %s: %d\n", playerName, score))
	}
	s.WriteString("\n")

	// Current Round
	if len(m.game.Rounds) > 0 {
		round := m.game.Rounds[len(m.game.Rounds)-1]
		s.WriteString( lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("--- Current Round ---") )
		s.WriteString("\n")
		
		czarName := round.CzarId
		for _, p := range m.session.Players {
			if p.Id == round.CzarId {
				czarName = p.Name
				break
			}
		}
		s.WriteString(fmt.Sprintf("Czar: %s\n", czarName))
		
		if round.BlackCard != nil {
			s.WriteString(cardStyle.Render(fmt.Sprintf("BLACK CARD:\n%s", round.BlackCard.Text)))
			s.WriteString("\n\n")
		}

		s.WriteString(fmt.Sprintf("Plays: %d / %d players\n", len(round.Plays), len(m.session.Players)-1))
	}

	// Your Hand
	s.WriteString("\n")
	s.WriteString(titleStyle.Render("YOUR HAND:"))
	s.WriteString("\n")
	for i, c := range m.hand {
		s.WriteString(fmt.Sprintf("%d) %s\n", i+1, c.Text))
	}

	s.WriteString("\n\n(q to quit, auto-refreshing every 2s...)\n")

	return s.String()
}

// Messages and Commands

type sessionsMsg struct{ items []list.Item }
type sessionJoinedMsg struct{ session *pb.Session }
type gameMsg struct{ game *pb.Game }
type handMsg struct{ cards []*pb.Card }
type tickMsg struct{}
type errorMsg struct{ err error }

func (m model) fetchSessions() tea.Cmd {
	return func() tea.Msg {
		client := pb.NewSessionServiceClient(m.conn)
		resp, err := client.ListSessions(m.token, &pb.ListSessionsRequest{})
		if err != nil {
			return errorMsg{err}
		}
		var items []list.Item
		for _, s := range resp.Sessions {
			items = append(items, sessionItem{id: s.Id})
		}
		return sessionsMsg{items}
	}
}

func (m model) joinSession(id string) tea.Cmd {
	return func() tea.Msg {
		// Just Fetch it for now since the user might already be joined or we want to use existing session logic
		// But the request said "select a session", and we should join it.
		// However, we don't have a "session_id" associated with a player without joining.
		// For simplicity in TUI, let's assume the user has a name and we join them.
		// But wait, the user just passes a token at the beginning.
		// If we use the token, we can get the user info? 
		// Actually, our JoinSession requires a player_name.
		// Let's just GetSession for now and hope they are already there, or we'll need to ask for a name.
		// Alternatively, we can just GetSession and then try to find the Game.
		client := pb.NewSessionServiceClient(m.conn)
		resp, err := client.GetSession(m.token, &pb.GetSessionRequest{Id: id})
		if err != nil {
			return errorMsg{err}
		}
		return sessionJoinedMsg{session: resp}
	}
}

func (m model) fetchGame() tea.Cmd {
	return func() tea.Msg {
		client := pb.NewGameServiceClient(m.conn)
		resp, err := client.GetGameBySession(m.token, &pb.GetGameBySessionRequest{SessionId: m.session.Id})
		if err != nil {
			// If not found, maybe wait?
			return nil 
		}
		return gameMsg{game: resp}
	}
}

func (m model) fetchHand() tea.Cmd {
	return func() tea.Msg {
		if m.game == nil {
			return nil
		}
		client := pb.NewGameServiceClient(m.conn)
		resp, err := client.GetHand(m.token, &pb.GetHandRequest{GameId: m.game.Id})
		if err != nil {
			return nil
		}
		return handMsg{cards: resp.Cards}
	}
}
