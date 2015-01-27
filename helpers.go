package main

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type PlayerStatus struct {
	Name     string
	Img      string
	Score    int
	Judge    bool
	Answered bool
}

type SocketAction struct {
	Action string
	Data   map[string]interface{}
}

func sendCardsToPlayers() {
	for c := range h.connections {
		if c.player.LoggedIn {
			c.send <- cardMessageJSON(c.player.Hand)
		}
	}
}

func refreshMessageJSON() []byte {
	dat := SocketAction{
		"refresh",
		map[string]interface{}{},
	}
	resp, _ := json.Marshal(dat)
	return resp
}

func cardMessageJSON(Hand []Card) []byte {
	dat := SocketAction{
		"dealt",
		map[string]interface{}{
			"Cards": Hand,
		},
	}
	resp, _ := json.Marshal(dat)
	return resp
}

// Clean up shop, we're done here
func gameOverJSON() []byte {
	dat := SocketAction{
		"gameover",
		map[string]interface{}{},
	}
	resp, _ := json.Marshal(dat)
	return resp
}

// Clean up shop, we're done here
func winnerJSON() []byte {
	code, _ := ioutil.ReadFile("code")
	dat := SocketAction{
		"winner",
		map[string]interface{}{
			"Code": string(code),
		},
	}
	resp, _ := json.Marshal(dat)
	return resp
}

func makeStatus(player *Player) string {
	if game.GameState == NotStarted {
		return "Not started"
	}
	if game.GameState == WaitingOnJudge {
		if player == game.Judge {
			return "Pick a card"
		} else {
			return "Waiting on Judge"
		}
	} else {
		if player == game.Judge {
			return "Waiting on players"
		} else if len(player.Answer) == 0 {
			return "Waiting on you"
		} else {
			return "Waiting on other players"
		}
	}
}
