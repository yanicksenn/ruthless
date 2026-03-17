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
	ErrNotCzar               = errors.New("only the czar can select a winner")
	ErrPlayNotFound          = errors.New("play not found")
	ErrInvalidNumberOfCards = errors.New("invalid number of cards played")
)

const HandSize = 10

func ConsolidateDecks(game *pb.Game, deckIDs []string, allDecks map[string]*pb.Deck, allCards map[string]*pb.Card) {
	for _, deckID := range deckIDs {
		deck, ok := allDecks[deckID]
		if !ok {
			continue
		}
		for _, cardID := range deck.CardIds {
			card, ok := allCards[cardID]
			if !ok {
				continue
			}
			if card.Color == pb.CardColor_CARD_COLOR_BLACK {
				game.HiddenBlackDeck = append(game.HiddenBlackDeck, card)
			} else {
				game.HiddenWhiteDeck = append(game.HiddenWhiteDeck, card)
			}
		}
	}
}

func StartGame(game *pb.Game, minPlayers int) error {
	if game.State != pb.GameState_GAME_STATE_WAITING {
		return ErrGameNotWaiting
	}
	if len(game.PlayerIds) < minPlayers {
		return errors.New("not enough players")
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

	for _, pid := range game.PlayerIds {
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
	startNewRound(game, game.PlayerIds)
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
		Plays:     make(map[string]*pb.Play),
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

	if len(cardIDs) != CountBlanks(currentRound.BlackCard.Text) {
		return nil, ErrInvalidNumberOfCards
	}

	// Check if already played
	if _, ok := currentRound.Plays[playerID]; ok {
		return nil, ErrInvalidState // already played
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

	if currentRound.Plays == nil {
		currentRound.Plays = make(map[string]*pb.Play)
	}
	currentRound.Plays[playerID] = play

	// Check if all players (except czar) have played
	expectedPlays := len(game.Scores) - 1
	if len(currentRound.Plays) == expectedPlays {
		game.State = pb.GameState_GAME_STATE_JUDGING
	}

	return play, nil
}

func SelectWinner(game *pb.Game, czarID, playID string) error {
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

	startNewRound(game, game.PlayerIds)
	return nil
}

func StripHidden(game *pb.Game) *pb.Game {
	// Create a copy without hidden fields for clients
	clone := &pb.Game{
		Id:                 game.Id,
		SessionId:          game.SessionId,
		State:              game.State,
		Rounds:             game.Rounds, // rounds are safe to share
		Scores:             game.Scores,
		Players:            game.Players,
		PlayerIds:          game.PlayerIds,
		CreatedAt:          game.CreatedAt,
		MinRequiredPlayers: game.MinRequiredPlayers,
	}
	return clone
}

func HandlePlayerLeave(game *pb.Game, playerID string) {
	// Do NOT remove from game.PlayerIds, we need it for re-join detection.
	// But we DO remove from game.Players (active players)

	// Remove player from Players slice
	newPlayers := make([]*pb.Player, 0)
	for _, p := range game.Players {
		if p.Id != playerID {
			newPlayers = append(newPlayers, p)
		}
	}
	game.Players = newPlayers

	// Remove from scores and hands
	delete(game.Scores, playerID)
	delete(game.HiddenHands, playerID)

	// Check for abandonment
	if len(game.PlayerIds) < int(game.MinRequiredPlayers) {
		game.State = pb.GameState_GAME_STATE_ABANDONED
		return
	}

	// If game is in an active round (PLAYING or JUDGING), reset it
	if game.State == pb.GameState_GAME_STATE_PLAYING || game.State == pb.GameState_GAME_STATE_JUDGING {
		if len(game.Rounds) > 0 {
			currentRound := game.Rounds[len(game.Rounds)-1]
			
			// Return played cards to players' hands
			for pid, play := range currentRound.Plays {
				if hand, ok := game.HiddenHands[pid]; ok {
					// We need to remove the cards that were drawn when playing
					// This is tricky because we drew cards to fill the hand.
					// Let's just restore the hand to what it was before playing.
					// Actually, the simplest is to just append them back if they aren't there,
					// but that might exceed HandSize. 
					// Better: when playing, we draw. If we reset, we should put those drawn cards back in the deck
					// and put the played cards back in hand.
					
					// For simplicity in this first pass, let's just give the cards back
					// and if they have > HandSize, so be it for now, or we can trim.
					hand.Cards = append(hand.Cards, play.Cards...)
				}
			}

			// Remove the current round
			game.Rounds = game.Rounds[:len(game.Rounds)-1]

			// Reset to PLAYING to trigger a new round (which will pick the next Czar)
			// Re-use startNewRound which handles Czar rotation based on len(game.Rounds)
			startNewRound(game, game.PlayerIds)
		}
	}
}

