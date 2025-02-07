package main

import (
	"fmt"

	"github.com/Scrimzay/blackjacksimulator/ai"
	"github.com/Scrimzay/blackjacksimulator/deck"
)

// basicAI represents a simple card-counting AI that adjusts bets and decisions 
// based on the number of high/low cards seen in the game.
type basicAI struct {
	score int // Running count of the card values seen
	seen  int // Number of cards seen so far
	decks int // Number of decks in play
}

// Bet calculates the betting amount based on the true count (score adjusted for unseen cards).
// If the deck is shuffled, it resets the counting variables.
func (bi *basicAI) Bet(shuffled bool) int {
	if shuffled {
		bi.score = 0
		bi.seen = 0
	}
	// Calculate the true count: running count divided by the number of remaining decks
	trueScore := bi.score / ((bi.decks*52 - bi.seen) / 52)

	// Adjust bet size based on the true count value
	switch {
	case trueScore >= 14:
		return 100000 // Very high confidence in a favorable deck
	case trueScore >= 8:
		return 5000   // Medium confidence
	default:
		return 100    // Default minimal bet
	}
}

// Play determines the AI's move based on basic blackjack strategy and card counting.
func (bi *basicAI) Play(hand []deck.Card, dealer deck.Card) ai.Move {
	score := ai.Score(hand...)

	// If the player has two cards
	if len(hand) == 2 {
		// Check for pair splitting strategy
		if hand[0] == hand[1] {
			cardScore := ai.Score(hand[0])
			if cardScore >= 8 && cardScore != 10 {
				return ai.MoveSplit // Split pairs if the value is favorable
			}
		}

		// Double down strategy for hands with a total of 10 or 11 (excluding soft hands)
		if score == 10 || (score == 11 && !ai.Soft(hand...)) {
			return ai.MoveDouble
		}
	}

	// Dealer strategy influences the decision
	dScore := ai.Score(dealer)
	if dScore >= 5 && dScore <= 6 {
		return ai.MoveStand // Favorable situation, stand
	}

	// If the player's score is low, hit
	if score < 13 {
		return ai.MoveHit
	}

	// Otherwise, stand
	return ai.MoveStand
}

// Results processes the final hands of the round and updates the card count.
func (bi *basicAI) Results(hands [][]deck.Card, dealer []deck.Card) {
	// Count the dealer's cards
	for _, card := range dealer {
		bi.count(card)
	}
	// Count all player hands
	for _, hand := range hands {
		for _, card := range hand {
			bi.count(card)
		}
	}
}

// count updates the running card count based on the value of a given card.
// - High-value cards (10, J, Q, K, A) decrease the count
// - Low-value cards (2-6) increase the count
func (bi *basicAI) count(card deck.Card) {
	score := ai.Score(card)
	switch {
	case score >= 10:
		bi.score-- // High-value cards are bad for the player
	case score <= 6:
		bi.score++ // Low-value cards are good for the player
	}
	bi.seen++ // Increment the total number of seen cards
}

func main() {
	// Define game options
	opts := ai.Options{
		Decks:          4,       // Number of decks used
		Hands:          999999,  // Number of hands to simulate
		BlackjackPayout: 1.5,    // Standard blackjack payout ratio
	}

	// Create and run the game simulation using the basicAI strategy
	game := ai.New(opts)
	winnings := game.Play(&basicAI{
		decks: 4, // Initialize AI with 4 decks
	})

	// Print the total winnings from the simulation
	fmt.Println(winnings)
}