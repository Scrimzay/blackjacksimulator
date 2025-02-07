package ai

import (
	"github.com/Scrimzay/blackjacksimulator/deck"
	"fmt"
)

// AI interface defines the behavior for different types of players (human or dealer).
type AI interface {
	// Bet determines the amount a player wants to bet, considering if the deck was shuffled.
	Bet(shuffled bool) int
	
	// Play takes the player's current hand and the dealer's visible card, returning the player's move.
	Play(hand []deck.Card, dealer deck.Card) Move
	
	// Results provides feedback at the end of the round, showing the final hands.
	Results(hand [][]deck.Card, dealer []deck.Card)
}

// dealerAI is the built-in AI for the dealer's moves.
type dealerAI struct {}

// Bet is a no-op for the dealer since the dealer doesn't bet.
func (ai dealerAI) Bet(shuffled bool) int {
	return 1 // Returns a dummy value since the dealer doesn't bet.
}

// Play determines the dealer's move based on blackjack rules:
// - Hit on 16 or lower
// - Hit on soft 17 (an Ace counted as 11)
// - Otherwise, stand
func (ai dealerAI) Play(hand []deck.Card, dealer deck.Card) Move {
	dScore := Score(hand...)
	if dScore <= 16 || (dScore == 17 && Soft(hand...)) {
		return MoveHit
	} 
	return MoveStand
}

// Results is a no-op for the dealer AI since it doesn’t need to process results.
func (ai dealerAI) Results(hand [][]deck.Card, dealer []deck.Card) {}

// humanAI represents a human player, requiring user input for actions.
type humanAI struct {}

// HumanAI initializes and returns a human-controlled AI.
func HumanAI() AI {
	return humanAI{}
}

// Bet prompts the player to enter their bet amount. If the deck was shuffled, it notifies the player.
func (ai humanAI) Bet(shuffled bool) int {
	if shuffled {
		fmt.Println("The deck was just shuffled")
	}
	fmt.Println("What would you like to bet?")
	var bet int
	fmt.Scanf("%d\n", &bet)
	return bet
}

// Play prompts the player to choose an action: hit, stand, double, or split.
func (ai humanAI) Play(hand []deck.Card, dealer deck.Card) Move {
	for {
		fmt.Println("Player:", hand)
		fmt.Println("Dealer:", dealer)
		fmt.Println("What will you do? (h)it, (s)tand, (d)ouble or s(p)lit")
		var input string
		fmt.Scanf("%s\n", &input)
		switch input {
		case "h":
			return MoveHit
		case "s":
			return MoveStand
		case "d":
			return MoveDouble
		case "p":
			return MoveSplit
		default:
			fmt.Println("Not a valid option.")
		}
	}
}

// Results displays the final hands of both the player and dealer at the end of the round.
func (ai humanAI) Results(hands [][]deck.Card, dealer []deck.Card) {
	fmt.Println("=== FINAL HANDS ===")
	fmt.Println("Player:")
	for _, h := range hands {
		fmt.Println(" ", h)
	}
	fmt.Println("Dealer:", dealer)
}