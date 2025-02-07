package ai

import (
	"github.com/Scrimzay/blackjacksimulator/deck"
	"errors"
)

// Represents the current state of the game using an int8 type.
type state int8

const (
	statePlayerTurn state = iota  // Player's turn
	stateDealerTurn               // Dealer's turn
	stateHandOver                 // Round is over
)

// Options struct defines configuration parameters for the game.
type Options struct {
	Decks           int     // Number of decks used in the game
	Hands           int     // Number of hands to be played
	BlackjackPayout float64 // Payout ratio for blackjack
}

// New initializes a Game instance with default values if options are not provided.
func New(opts Options) Game {
	g := Game{
		state:    statePlayerTurn,
		dealerAI: dealerAI{},
		balance:  0,
	}
	// Set default values if none are provided
	if opts.Decks == 0 {
		opts.Decks = 3
	}
	if opts.Hands == 0 {
		opts.Hands = 100
	}
	if opts.BlackjackPayout == 0.0 {
		opts.BlackjackPayout = 1.5
	}
	g.nDecks = opts.Decks
	g.nHands = opts.Hands
	g.blackjackPayout = opts.BlackjackPayout
	return g
}

// Game represents the state of the game.
type Game struct {
	nDecks          int     // Number of decks
	nHands          int     // Number of hands
	blackjackPayout float64 // Payout ratio for blackjack

	deck     []deck.Card // The deck of cards
	state    state       // Current game state

	player   []hand // Player's hands
	handIdx  int    // Index of the active hand
	playerBet int   // Current bet amount
	balance   int   // Player's balance

	dealer   []deck.Card // Dealer's hand
	dealerAI AI          // AI logic for dealer's moves
}

// currentHand returns a pointer to the current active hand's cards.
func (g *Game) currentHand() *[]deck.Card {
	switch g.state {
	case statePlayerTurn:
		return &g.player[g.handIdx].cards
	case stateDealerTurn:
		return &g.dealer
	default:
		panic("It isn't currently any players' turn")
	}
}

// hand represents a single hand played by the player.
type hand struct {
	cards []deck.Card // Cards in the hand
	bet   int         // Bet placed on the hand
}

// bet places a bet for the player using the AI logic.
func bet(g *Game, ai AI, shuffled bool) {
	bet := ai.Bet(shuffled)
	if bet < 100 {
		panic("Bet must be at least 100")
	}
	g.playerBet = bet
}

// deal distributes two cards to the player and dealer at the beginning of a round.
func deal(g *Game) {
	playerHand := make([]deck.Card, 0, 5) // Player's hand initialized with capacity of 5
	g.handIdx = 0
	g.dealer = make([]deck.Card, 0, 5) // Dealer's hand initialized

	var card deck.Card
	for i := 0; i < 2; i++ {
		card, g.deck = draw(g.deck)
		playerHand = append(playerHand, card)
		card, g.deck = draw(g.deck)
		g.dealer = append(g.dealer, card)
	}
	g.player = []hand{
		{
			cards: playerHand,
			bet:   g.playerBet,
		},
	}
	g.state = statePlayerTurn
}

// Play runs the game loop for the specified number of hands.
func (g *Game) Play(ai AI) int {
	g.deck = nil
	min := 52 * g.nDecks / 3 // Minimum deck size before reshuffling

	for i := 0; i < g.nHands; i++ {
		shuffled := false
		if len(g.deck) < min {
			g.deck = deck.New(deck.Deck(g.nDecks), deck.Shuffle)
			shuffled = true
		}
		bet(g, ai, shuffled)
		deal(g)

		// Check for dealer blackjack immediately
		if Blackjack(g.dealer...) {
			endRound(g, ai)
			continue
		}

		// Player's turn
		for g.state == statePlayerTurn {
			hand := make([]deck.Card, len(*g.currentHand()))
			copy(hand, *g.currentHand())
			move := ai.Play(hand, g.dealer[0])
			err := move(g)
			switch err {
			case errBust:
				MoveStand(g) // If player busts, automatically stand
			case nil:
				// No error, continue
			default:
				panic(err)
			}
		}

		// Dealer's turn
		for g.state == stateDealerTurn {
			hand := make([]deck.Card, len(g.dealer))
			copy(hand, g.dealer)
			move := g.dealerAI.Play(hand, g.dealer[0])
			move(g)
		}

		endRound(g, ai)
	}
	return g.balance
}

// Error representing a busted hand.
var (
	errBust = errors.New("Hand score exceeded 21")
)

// Move represents a function that executes a player's move.
type Move func(*Game) error

// MoveHit allows the player to draw a card.
func MoveHit(g *Game) error {
	hand := g.currentHand()
	var card deck.Card
	card, g.deck = draw(g.deck)
	*hand = append(*hand, card)
	if Score(*hand...) > 21 {
		return errBust
	}
	return nil
}

// MoveSplit allows the player to split their hand if they have two identical cards.
func MoveSplit(g *Game) error {
	cards := g.currentHand()
	if len(*cards) != 2 {
		return errors.New("You can only split with two cards in your hand")
	}
	if (*cards)[0].Rank != (*cards)[1].Rank {
		return errors.New("Both cards must have the same rank to split")
	}
	g.player = append(g.player, hand{
		cards: []deck.Card{(*cards)[1]},
		bet:   g.player[g.handIdx].bet,
	})
	g.player[g.handIdx].cards = (*cards)[:1]
	return nil
}

// MoveDouble allows the player to double their bet and draw one final card.
func MoveDouble(g *Game) error {
	if len(*g.currentHand()) != 2 {
		return errors.New("Can only double on a hand with 2 cards")
	}
	g.playerBet *= 2
	MoveHit(g)
	return MoveStand(g)
}

// MoveStand ends the player's turn.
func MoveStand(g *Game) error {
	if g.state == stateDealerTurn {
		g.state++
		return nil
	}
	if g.state == statePlayerTurn {
		g.handIdx++
		if g.handIdx >= len(g.player) {
			g.state++
		}
		return nil
	}
	return errors.New("Invalid state")
}

// draw removes and returns the top card from the deck.
func draw(cards []deck.Card) (deck.Card, []deck.Card) {
	return cards[0], cards[1:]
}

// endRound evaluates the results of the round and updates the balance.
func endRound(g *Game, ai AI) {
	dScore := Score(g.dealer...)
	dBlackjack := Blackjack(g.dealer...)

	allHands := make([][]deck.Card, len(g.player))
	for hi, hand := range g.player {
		cards := hand.cards
		allHands[hi] = cards

		pScore, pBlackjack := Score(cards...), Blackjack(cards...)
		winnings := hand.bet

		switch {
		case pBlackjack && dBlackjack:
			winnings = 0
		case dBlackjack, pScore > 21:
			winnings = -winnings
		case pBlackjack:
			winnings = int(float64(winnings) * g.blackjackPayout)
		case dScore > 21, pScore > dScore:
			// Win
		case dScore == pScore:
			winnings = 0
		default:
			winnings = -winnings
		}
		g.balance += winnings
	}
	ai.Results(allHands, g.dealer)
	g.player = nil
	g.dealer = nil
}

// Score calculates the best possible score for a hand.
func Score(hand ...deck.Card) int {
	minScore := minScore(hand...)
	if minScore > 11 {
		return minScore
	}
	for _, c := range hand {
		if c.Rank == deck.Ace {
			return minScore + 10
		}
	}
	return minScore
}

// soft 17 score for dealer
func Soft(hand ...deck.Card) bool {
	minScore := minScore(hand...)
	score := Score(hand...)
	return minScore != score
}

// identifies a blackjack
func Blackjack(hand ...deck.Card) bool {
	return len(hand) == 2 && Score(hand...) == 21
}

func minScore(hand ...deck.Card) int {
	score := 0
	for _, c := range hand {
		score += min(int(c.Rank), 10)
	}

	return score
}

// helper func
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}