package domain

import (
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
	pb "github.com/yanicksenn/ruthless/api/v1"
)

var (
	ErrGameNotWaiting = errors.New("game is not waiting for players")
	ErrNotEnoughCards = errors.New("not enough cards to start")
	ErrInvalidState   = errors.New("invalid game state")
	ErrNotYourTurn    = errors.New("not your turn or you are the czar")
	ErrCzarCannotPlay = errors.New("czar cannot play white cards")
	ErrPlayerNotFound = errors.New("player not found")
	ErrNotCzar        = errors.New("only the czar can select a winner")
	ErrPlayNotFound   = errors.New("play not found")
)

const HandSize = 10

func StartGame(game *pb.Game, session *pb.Session, allCards map[string]*pb.Card, allDecks map[string]*pb.Deck) error {
	if game.State != pb.GameState_GAME_STATE_WAITING {
		return ErrGameNotWaiting
	}
	if len(session.PlayerIds) < 3 {
		// Just for testing, maybe no strict player limit but typical is 3
		// We'll allow 2 for easier testing
	}

	game.HiddenBlackDeck = make([]*pb.Card, 0)
	game.HiddenWhiteDeck = make([]*pb.Card, 0)

	// Consolidate decks
	for _, deckID := range session.DeckIds {
		deck, ok := allDecks[deckID]
		if !ok {
			continue
		}
		for _, cardID := range deck.CardIds {
			card, ok := allCards[cardID]
			if !ok {
				continue
			}
			if card.Blanks > 0 {
				game.HiddenBlackDeck = append(game.HiddenBlackDeck, card)
			} else {
				game.HiddenWhiteDeck = append(game.HiddenWhiteDeck, card)
			}
		}
	}

	if len(game.HiddenBlackDeck) == 0 || len(game.HiddenWhiteDeck) == 0 {
		return ErrNotEnoughCards
	}

	// Shuffle
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(game.HiddenBlackDeck), func(i, j int) {
		game.HiddenBlackDeck[i], game.HiddenBlackDeck[j] = game.HiddenBlackDeck[j], game.HiddenBlackDeck[i]
	})
	rand.Shuffle(len(game.HiddenWhiteDeck), func(i, j int) {
		game.HiddenWhiteDeck[i], game.HiddenWhiteDeck[j] = game.HiddenWhiteDeck[j], game.HiddenWhiteDeck[i]
	})

	// Deal hands
	game.HiddenHands = make(map[string]*pb.PlayerHand)
	game.Scores = make(map[string]uint32)

	for _, pid := range session.PlayerIds {
		game.Scores[pid] = 0
		hand := &pb.PlayerHand{Cards: make([]*pb.Card, 0)}
		for i := 0; i < HandSize; i++ {
			if len(game.HiddenWhiteDeck) > 0 {
				hand.Cards = append(hand.Cards, game.HiddenWhiteDeck[0])
				game.HiddenWhiteDeck = game.HiddenWhiteDeck[1:]
			}
		}
		game.HiddenHands[pid] = hand
	}

	// Start round 1
	startNewRound(game, session.PlayerIds)
	return nil
}

func startNewRound(game *pb.Game, playerIDs []string) {
	if len(game.HiddenBlackDeck) == 0 {
		game.State = pb.GameState_GAME_STATE_FINISHED
		return
	}

	bc := game.HiddenBlackDeck[0]
	game.HiddenBlackDeck = game.HiddenBlackDeck[1:]

	czarIdx := len(game.Rounds) % len(playerIDs)
	czar := playerIDs[czarIdx]

	round := &pb.Round{
		Id:        uuid.New().String(),
		CzarId:    czar,
		BlackCard: bc,
		Plays:     make([]*pb.Play, 0),
	}
	game.Rounds = append(game.Rounds, round)
	game.State = pb.GameState_GAME_STATE_PLAYING
}

func PlayCards(game *pb.Game, playerID string, cardIDs []string) (*pb.Play, error) {
	if game.State != pb.GameState_GAME_STATE_PLAYING {
		return nil, ErrInvalidState
	}

	currentRound := game.Rounds[len(game.Rounds)-1]
	if currentRound.CzarId == playerID {
		return nil, ErrCzarCannotPlay
	}

	// Check if already played
	for _, p := range currentRound.Plays {
		if p.PlayerId == playerID {
			return nil, ErrInvalidState // already played
		}
	}

	hand, ok := game.HiddenHands[playerID]
	if !ok {
		return nil, ErrPlayerNotFound
	}

	// Find cards in hand
	var playedCards []*pb.Card
	var remainingCards []*pb.Card

	for _, cid := range cardIDs {
		var found *pb.Card
		for _, hc := range hand.Cards {
			if hc.Id == cid {
				found = hc
				break
			}
		}
		if found == nil {
			return nil, ErrCardNotFound
		}
		playedCards = append(playedCards, found)
	}

	// Remove played cards from hand and draw new ones
	handMap := make(map[string]bool)
	for _, fc := range playedCards {
		handMap[fc.Id] = true
	}

	for _, hc := range hand.Cards {
		if !handMap[hc.Id] {
			remainingCards = append(remainingCards, hc)
		}
	}

	for len(remainingCards) < HandSize && len(game.HiddenWhiteDeck) > 0 {
		remainingCards = append(remainingCards, game.HiddenWhiteDeck[0])
		game.HiddenWhiteDeck = game.HiddenWhiteDeck[1:]
	}
	game.HiddenHands[playerID].Cards = remainingCards

	play := &pb.Play{
		Id:       uuid.New().String(),
		PlayerId: playerID,
		Cards:    playedCards,
	}

	currentRound.Plays = append(currentRound.Plays, play)

	// Check if all players (except czar) have played
	expectedPlays := len(game.Scores) - 1
	if len(currentRound.Plays) == expectedPlays {
		game.State = pb.GameState_GAME_STATE_JUDGING
	}

	return play, nil
}

func SelectWinner(game *pb.Game, session *pb.Session, czarID, playID string) error {
	if game.State != pb.GameState_GAME_STATE_JUDGING {
		return ErrInvalidState
	}

	currentRound := game.Rounds[len(game.Rounds)-1]
	if currentRound.CzarId != czarID {
		return ErrNotCzar
	}

	var winningPlay *pb.Play
	for _, p := range currentRound.Plays {
		if p.Id == playID {
			winningPlay = p
			break
		}
	}

	if winningPlay == nil {
		return ErrPlayNotFound
	}

	currentRound.WinningPlayId = winningPlay.Id
	game.Scores[winningPlay.PlayerId]++

	startNewRound(game, session.PlayerIds)
	return nil
}

func StripHidden(game *pb.Game) *pb.Game {
	// Create a copy without hidden fields for clients
	clone := &pb.Game{
		Id:        game.Id,
		SessionId: game.SessionId,
		State:     game.State,
		Rounds:    game.Rounds, // rounds are safe to share
		Scores:    game.Scores,
	}
	return clone
}
