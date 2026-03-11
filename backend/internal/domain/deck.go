package domain

import (
	"errors"

	"github.com/google/uuid"
	pb "github.com/yanicksenn/ruthless/api/v1"
)

var (
	ErrUnauthorized = errors.New("user is not authorized to modify this deck")
	ErrCardNotFound = errors.New("card not found in deck")
)

func NewDeck(name string, ownerID string) *pb.Deck {
	return &pb.Deck{
		Id:           uuid.New().String(),
		Name:         name,
		OwnerId:      ownerID,
		Contributors: []string{},
		Cards:        []*pb.Card{},
	}
}

func CanModifyDeck(d *pb.Deck, userID string) bool {
	if d.OwnerId == userID {
		return true
	}
	for _, contributorID := range d.Contributors {
		if contributorID == userID {
			return true
		}
	}
	return false
}

func AddContributorToDeck(d *pb.Deck, ownerID, contributorID string) error {
	if d.OwnerId != ownerID {
		return ErrUnauthorized
	}
	for _, id := range d.Contributors {
		if id == contributorID {
			return nil // Already a contributor
		}
	}
	d.Contributors = append(d.Contributors, contributorID)
	return nil
}

func RemoveContributorFromDeck(d *pb.Deck, ownerID, contributorID string) error {
	if d.OwnerId != ownerID {
		return ErrUnauthorized
	}
	for i, id := range d.Contributors {
		if id == contributorID {
			d.Contributors = append(d.Contributors[:i], d.Contributors[i+1:]...)
			return nil
		}
	}
	return nil
}

func AddCardToDeck(d *pb.Deck, userID string, card *pb.Card) error {
	if !CanModifyDeck(d, userID) {
		return ErrUnauthorized
	}
	d.Cards = append(d.Cards, card)
	return nil
}

func RemoveCardFromDeck(d *pb.Deck, userID string, cardID string) error {
	if !CanModifyDeck(d, userID) {
		return ErrUnauthorized
	}
	for i, card := range d.Cards {
		if card.Id == cardID {
			d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
			return nil
		}
	}
	return ErrCardNotFound
}
