// Package main simulates the dice game Pig to find the best strategy.
//
// Source: https://go.dev/doc/codewalk/functions/
//
// ── What this codewalk teaches ──────────────────────────────────────────────
//
// The central Go idea here is that FUNCTIONS ARE FIRST-CLASS VALUES.
// You can assign them to variables, store them in data structures, pass them
// as arguments, and return them from other functions — the same as int or string.
//
// This program uses that idea to model game strategies:
//
//   type action   func(score) (score, bool)   — a single move (roll or stay)
//   type strategy func(score) action          — a policy that picks a move
//
// Both are just function types. stayAtK returns a strategy (a function) by
// returning another function — a closure that "remembers" k.
//
// ── The game of Pig ─────────────────────────────────────────────────────────
//
// Two players take turns. On your turn you repeatedly roll a die:
//   - Roll 1 → you lose all points accumulated this turn; opponent's turn
//   - Roll 2-6 → add roll value to your turn total; choose to roll again or stay
//   - Stay → add turn total to your score; opponent's turn
// First player to reach 100 wins.
//
// Run:
//
//	go run pig/pig.go
package main

import (
	"fmt"
	"math/rand"
)

const (
	win            = 100 // winning score
	gamesPerSeries = 10  // games played between each pair of strategies
)

// score holds the complete game state at any point in time.
// Using a struct (not separate variables) makes it easy to pass the full
// state to functions and return new states without mutation.
type score struct {
	player   int // current player's banked score
	opponent int // opponent's banked score
	thisTurn int // current player's unbanked points this turn
}

// action is a FUNCTION TYPE.
// Any function that takes a score and returns (score, bool) satisfies this type.
// The bool indicates whether the turn ended (roll of 1, or player chose to stay).
//
// This is the key insight: we can store functions in variables, pass them around,
// and call them without knowing which concrete function we have.
type action func(current score) (result score, turnIsOver bool)

// roll is an action: simulate one die roll.
// Roll of 1 → forfeit thisTurn, swap players (turn is over).
// Roll of 2-6 → add to thisTurn, same player continues (turn not over).
func roll(s score) (score, bool) {
	outcome := rand.Intn(6) + 1 // [1, 6]
	if outcome == 1 {
		// Lose thisTurn points; it's now the opponent's turn.
		// Note how player and opponent swap: opponent becomes the new player.
		return score{s.opponent, s.player, 0}, true
	}
	return score{s.player, s.opponent, outcome + s.thisTurn}, false
}

// stay is an action: bank thisTurn points and end the turn.
// thisTurn is added to player's score, then players swap.
func stay(s score) (score, bool) {
	return score{s.opponent, s.player + s.thisTurn, 0}, true
}

// strategy is a FUNCTION TYPE that returns an action.
// Given the current score, a strategy decides what to do next.
// Returning an action (rather than a bool) keeps the decision open-ended —
// future strategies could return custom actions, not just roll/stay.
type strategy func(score) action

// stayAtK returns a strategy that rolls until thisTurn >= k, then stays.
//
// This is a CLOSURE: the returned function "closes over" k, capturing it
// from the enclosing scope. Each call to stayAtK(k) produces a new,
// independent strategy function that remembers its own k value.
//
// stayAtK(1)  → stays after any successful roll (very cautious)
// stayAtK(25) → keeps rolling until 25 points banked (aggressive)
func stayAtK(k int) strategy {
	return func(s score) action {
		if s.thisTurn >= k {
			return stay // function VALUE, not a call — no ()
		}
		return roll
	}
}

// play simulates one complete Pig game between two strategies.
// Returns 0 if strategy0 wins, 1 if strategy1 wins.
//
// strategies is a SLICE OF FUNCTIONS — we index into it with currentPlayer
// to call whichever strategy is currently playing.
func play(strategy0, strategy1 strategy) int {
	strategies := []strategy{strategy0, strategy1}
	var s score
	var turnIsOver bool
	currentPlayer := rand.Intn(2) // random first player

	for s.player+s.thisTurn < win {
		// strategies[currentPlayer] is a function value — call it with (s)
		// to get back an action (another function value), then call that with (s).
		action := strategies[currentPlayer](s)
		s, turnIsOver = action(s)
		if turnIsOver {
			currentPlayer = (currentPlayer + 1) % 2 // swap players
		}
	}
	return currentPlayer
}

// roundRobin runs every pair of strategies against each other gamesPerSeries
// times and tallies wins. Returns a wins slice and games-per-strategy count.
func roundRobin(strategies []strategy) ([]int, int) {
	wins := make([]int, len(strategies))
	for i := 0; i < len(strategies); i++ {
		for j := i + 1; j < len(strategies); j++ {
			for k := 0; k < gamesPerSeries; k++ {
				winner := play(strategies[i], strategies[j])
				if winner == 0 {
					wins[i]++
				} else {
					wins[j]++
				}
			}
		}
	}
	gamesPerStrategy := gamesPerSeries * (len(strategies) - 1) // no self-play
	return wins, gamesPerStrategy
}

// ratioString formats a list of win counts as "wins/total (pct%), ..." strings.
// The variadic ...int parameter accepts any number of int arguments.
func ratioString(vals ...int) string {
	total := 0
	for _, val := range vals {
		total += val
	}
	s := ""
	for _, val := range vals {
		if s != "" {
			s += ", "
		}
		pct := 100 * float64(val) / float64(total)
		s += fmt.Sprintf("%d/%d (%0.1f%%)", val, total, pct)
	}
	return s
}

func main() {
	// Build 100 strategies: stayAtK(1) through stayAtK(100).
	// strategies is a []strategy — a slice of function values.
	strategies := make([]strategy, win)
	for k := range strategies {
		strategies[k] = stayAtK(k + 1)
	}

	wins, games := roundRobin(strategies)

	// Print each strategy's win/loss record.
	// The output shows that staying around k=20-25 tends to win most often.
	for k := range strategies {
		fmt.Printf("Wins, losses staying at k =% 4d: %s\n",
			k+1, ratioString(wins[k], games-wins[k]))
	}
}
