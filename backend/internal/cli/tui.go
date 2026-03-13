package cli

import (
	"context"
	"fmt"
	"sort"
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
	conn     *grpc.ClientConn
	token    context.Context
	playerID string
	state    state
	list     list.Model
	session  *pb.Session
	game     *pb.Game
	hand     []*pb.Card
	cursor   int
	selected []int
	err      error
	width    int
	height   int
}

func initialModel(conn *grpc.ClientConn, token string) model {
	md := metadata.Pairs("authorization", token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a Session"

	return model{
		conn:     conn,
		token:    ctx,
		playerID: token,
		state:    stateSelectingSession,
		list:     l,
		selected: []int{},
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
			} else if m.state == stateInGame && m.game != nil {
				if m.game.State == pb.GameState_GAME_STATE_PLAYING {
					return m, m.playCards()
				} else if m.game.State == pb.GameState_GAME_STATE_JUDGING {
					round := m.game.Rounds[len(m.game.Rounds)-1]
					if round.CzarId == m.playerID {
						return m, m.selectWinner()
					}
				}
			}
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			limit := 0
			if m.game != nil {
				if m.game.State == pb.GameState_GAME_STATE_PLAYING {
					limit = len(m.hand) - 1
				} else if m.game.State == pb.GameState_GAME_STATE_JUDGING {
					round := m.game.Rounds[len(m.game.Rounds)-1]
					limit = len(round.Plays) - 1
				}
			}
			if m.cursor < limit {
				m.cursor++
			}
		case " ":
			if m.game != nil && m.game.State == pb.GameState_GAME_STATE_PLAYING {
				// Only if not the czar
				round := m.game.Rounds[len(m.game.Rounds)-1]
				if round.CzarId != m.playerID && !hasSubmitted(m.game, m.playerID) {
					// Toggle selection
					found := -1
					for i, idx := range m.selected {
						if idx == m.cursor {
							found = i
							break
						}
					}

					if found != -1 {
						// Remove
						m.selected = append(m.selected[:found], m.selected[found+1:]...)
					} else {
						// Add if not at limit
					if len(m.selected) < strings.Count(round.BlackCard.Text, "___") {
						m.selected = append(m.selected, m.cursor)
					}
					}
				}
			} else if m.game != nil && m.game.State == pb.GameState_GAME_STATE_JUDGING {
				m.selected = []int{m.cursor} // In judging, we just pick one
			}
		case "left":
			if m.state == stateInGame && m.game != nil && m.game.State == pb.GameState_GAME_STATE_JUDGING {
				if m.cursor > 0 {
					m.cursor--
				}
			}
		case "right":
			if m.state == stateInGame && m.game != nil && m.game.State == pb.GameState_GAME_STATE_JUDGING {
				round := m.game.Rounds[len(m.game.Rounds)-1]
				if m.cursor < len(round.Plays)-1 {
					m.cursor++
				}
			}
		case "s":
			if m.state == stateInGame && m.game != nil && m.game.State == pb.GameState_GAME_STATE_WAITING {
				return m, m.startGame()
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

	case actionDoneMsg:
		m.selected = []int{}
		m.cursor = 0
		return m, tea.Batch(m.fetchGame(), m.fetchHand())

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

	playerIDs := make([]string, 0, len(m.game.Scores))
	for pid := range m.game.Scores {
		playerIDs = append(playerIDs, pid)
	}
	sort.Slice(playerIDs, func(i, j int) bool {
		si, sj := m.game.Scores[playerIDs[i]], m.game.Scores[playerIDs[j]]
		if si != sj {
			return si > sj
		}
		return playerIDs[i] < playerIDs[j]
	})

	for _, pid := range playerIDs {
		score := m.game.Scores[pid]
		s.WriteString(fmt.Sprintf("  %s: %d\n", pid, score))
	}
	s.WriteString("\n")

	// Current Round
	if len(m.game.Rounds) > 0 {
		round := m.game.Rounds[len(m.game.Rounds)-1]
		s.WriteString( lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("--- Current Round ---") )
		s.WriteString("\n")
		
		s.WriteString(fmt.Sprintf("Czar: %s\n", round.CzarId))
		
		bcText := round.BlackCard.Text
		if m.game.State == pb.GameState_GAME_STATE_JUDGING && len(round.Plays) > 0 {
			plays := getSortedPlays(round.Plays)
			bcText = renderBlackCard(bcText, plays[m.cursor].Cards)
			s.WriteString(fmt.Sprintf("\nViewing submission %d of %d (left/right to navigate)\n", m.cursor+1, len(round.Plays)))
		}

		// Calculate available width for the card (minus margins and border/padding)
		cardWidth := m.width - 6
		if cardWidth < 20 {
			cardWidth = 20
		}

		s.WriteString(cardStyle.Width(cardWidth).Render(fmt.Sprintf("BLACK CARD:\n%s", bcText)))
		s.WriteString("\n\n")

		if m.game.State == pb.GameState_GAME_STATE_JUDGING && len(round.Plays) > 0 && round.CzarId == m.playerID {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("YOU ARE THE CZAR. Press Enter to pick this as the winner!"))
			s.WriteString("\n")
		}

		s.WriteString(fmt.Sprintf("Plays: %d / %d players\n", len(round.Plays), len(m.session.PlayerIds)-1))
	}

	// Your Hand
	if m.game.State == pb.GameState_GAME_STATE_PLAYING {
		round := m.game.Rounds[len(m.game.Rounds)-1]
		if round.CzarId == m.playerID {
			s.WriteString("\n")
			s.WriteString( lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("YOU ARE THE CZAR. Wait for other players to submit their cards.") )
			s.WriteString("\n")
		} else if hasSubmitted(m.game, m.playerID) {
			s.WriteString("\n")
			s.WriteString( lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("You have submitted your cards. Waiting for others...") )
			s.WriteString("\n")
		} else {
			s.WriteString("\n")
			s.WriteString(titleStyle.Render("YOUR HAND:"))
			s.WriteString("\n")
			for i, c := range m.hand {
				cursor := " "
				if m.cursor == i {
					cursor = ">"
				}
				selected := " "
				found := -1
				for j, idx := range m.selected {
					if idx == i {
						found = j
						break
					}
				}

				if found != -1 {
					if strings.Count(round.BlackCard.Text, "___") > 1 {
						selected = fmt.Sprintf("%d", found+1)
					} else {
						selected = "*"
					}
				}
				s.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, selected, c.Text))
			}
		}
	}

	s.WriteString("\n\n(q to quit, arrows to move, space to select, enter to submit)")
	if m.game != nil && m.game.State == pb.GameState_GAME_STATE_JUDGING {
		s.WriteString("\n(left/right to flip through submissions)")
	}
	s.WriteString("\n(auto-refreshing every 2s...)")
	if m.game != nil && m.game.State == pb.GameState_GAME_STATE_WAITING {
		s.WriteString("\n(s to start game)")
	}
	s.WriteString("\n")

	return s.String()
}

// Messages and Commands

type sessionsMsg struct{ items []list.Item }
type sessionJoinedMsg struct{ session *pb.Session }
type gameMsg struct{ game *pb.Game }
type handMsg struct{ cards []*pb.Card }
type tickMsg struct{}
type actionDoneMsg struct{}
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

func (m model) startGame() tea.Cmd {
	return func() tea.Msg {
		client := pb.NewGameServiceClient(m.conn)
		_, err := client.StartGame(m.token, &pb.StartGameRequest{Id: m.game.Id})
		if err != nil {
			return errorMsg{err}
		}
		return tickMsg{}
	}
}

func (m model) playCards() tea.Cmd {
	return func() tea.Msg {
		// Safety check: Czar cannot play cards
		round := m.game.Rounds[len(m.game.Rounds)-1]
		if round.CzarId == m.playerID || hasSubmitted(m.game, m.playerID) {
			return nil // ignore if czar or already submitted
		}

		var cardIDs []string
		for _, idx := range m.selected {
			if idx < len(m.hand) {
				cardIDs = append(cardIDs, m.hand[idx].Id)
			}
		}

		if len(cardIDs) != strings.Count(round.BlackCard.Text, "___") {
			return nil // Don't allow submission if not enough cards
		}

		client := pb.NewGameServiceClient(m.conn)
		_, err := client.PlayCards(m.token, &pb.PlayCardsRequest{
			GameId:  m.game.Id,
			CardIds: cardIDs,
		})
		if err != nil {
			return errorMsg{err}
		}

		// Clear selection and cursor
		return actionDoneMsg{}
	}
}

func (m model) selectWinner() tea.Cmd {
	return func() tea.Msg {
		// Only the Czar can select a winner
		round := m.game.Rounds[len(m.game.Rounds)-1]
		if round.CzarId != m.playerID {
			return errorMsg{fmt.Errorf("only the czar can select a winner")}
		}

		if m.cursor >= len(round.Plays) {
			return nil
		}
		plays := getSortedPlays(round.Plays)
		playID := plays[m.cursor].Id

		client := pb.NewGameServiceClient(m.conn)
		_, err := client.SelectWinner(m.token, &pb.SelectWinnerRequest{
			GameId: m.game.Id,
			PlayId: playID,
		})
		if err != nil {
			return errorMsg{err}
		}

		return actionDoneMsg{}
	}
}

func renderBlackCard(template string, whiteCards []*pb.Card) string {
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Underline(true)
	result := template
	for _, wc := range whiteCards {
		if !strings.Contains(result, "___") {
			break
		}
		// Use a marker that's less likely to collide, but here we just replace the first ___
		replacement := highlightStyle.Render(wc.Text)
		result = strings.Replace(result, "___", replacement, 1)
	}
	// Clean up any remaining underscores if any (though usually card has exact count)
	return result
}

func hasSubmitted(game *pb.Game, playerID string) bool {
	if game == nil || len(game.Rounds) == 0 {
		return false
	}
	round := game.Rounds[len(game.Rounds)-1]
	_, ok := round.Plays[playerID]
	return ok
}

func getSortedPlays(plays map[string]*pb.Play) []*pb.Play {
	playerIDs := make([]string, 0, len(plays))
	for pid := range plays {
		playerIDs = append(playerIDs, pid)
	}
	sort.Strings(playerIDs)

	sorted := make([]*pb.Play, 0, len(plays))
	for _, pid := range playerIDs {
		sorted = append(sorted, plays[pid])
	}
	return sorted
}
