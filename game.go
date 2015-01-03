package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
)

type Game struct {
	Players   []*Player
	Deck      *Deck
	Judge     *Player
	GameState State
	Question  Card
}

func CreateGame(deckFile, playerFile string) *Game {
	players := LoadPlayers(playerFile)
	deck := NewDeck(deckFile)
	game := Game{players, deck, nil, NotStarted, Card{}}
	return &game
}

func (g *Game) StartGame() error {
	if g.GameState != NotStarted {
		return errors.New("Game has already started")
	}

	g.GameState = Initializing

	g.SetJudge()

	numOn := 0
	for _, player := range g.Players {
		if player.LoggedIn {
			numOn++
		}
	}

	cards := g.Deck.DrawAnswerCards(numOn * 10)

	i := 0
	for _, player := range g.Players {
		if player.LoggedIn {
			player.Hand = cards[i*10 : (i+1)*10]
			i++
		}
	}

	h.broadcast <- refreshMessageJSON()
	sendCardsToPlayers()
	g.SendQuestion()

	g.GameState = WaitingOnPlayers

	return nil
}

func (g *Game) SendQuestion() {
	card := g.Deck.DrawQuestionCard()
	g.Question = card
	res, err := g.questionCardJSON(card)
	if err != nil {
		log.Println("Error with qCard JSON", err)
		return
	}
	h.broadcast <- res
}

func (g *Game) SendAnswerAndQuestion(player *Player) {
	card := g.Deck.DrawQuestionCard()
	g.Question = card
	res, err := g.answerQuestionCardJSON(player.Answer, card)
	if err != nil {
		log.Println("Error with aqCard JSON", err)
		return
	}
	h.broadcast <- res
}

func (g *Game) questionCardJSON(card Card) ([]byte, error) {
	if card.CardType == "Q" {
		dat := SocketAction{
			"question",
			map[string]interface{}{
				"Question": card,
				"Judge":    g.Judge.Name,
			},
		}
		resp, err := json.Marshal(dat)
		return resp, err
	} else {
		return []byte(""), errors.New("Not a question card")
	}
}

func (g *Game) answerQuestionCardJSON(answers []Card, question Card) ([]byte, error) {
	dat := SocketAction{
		"answerquestion",
		map[string]interface{}{
			"Question": question,
			"Answer":   answers,
			"Judge":    g.Judge.Name,
		},
	}
	resp, err := json.Marshal(dat)
	return resp, err
}

func (g *Game) AnswerJSON() ([]byte, error) {
	answers := make([][]Card, 0)

	for _, player := range g.Players {
		if player.LoggedIn &&
			len(player.Answer) > 0 {
			answers = append(answers, player.Answer)
		}
	}

	for i := range answers {
		j := rand.Intn(i + 1)
		answers[i], answers[j] = answers[j], answers[i]
	}

	dat := SocketAction{
		"answers",
		map[string]interface{}{
			"Answers": answers,
		},
	}

	resp, err := json.Marshal(dat)
	return resp, err
}

func (g *Game) AllIn() bool {
	for _, player := range g.Players {
		if player.LoggedIn &&
			len(player.Answer) == 0 &&
			player != g.Judge {
			return false
		}
	}
	return true
}

func (g *Game) SendAnswers() {
	res, err := g.AnswerJSON()
	if err != nil {
		log.Println("Error with answer JSON", err)
		return
	}
	h.broadcast <- res
}

func (g *Game) PlayerByHandshake(handshake string) *Player {
	// Look through the list of users we have, see if one of
	// them is allowed in the game
	for _, player := range g.Players {
		if player.Handshake == handshake {
			return player
		}
	}
	return nil
}

func (g *Game) PlayerByAnswers(answer []Card) *Player {
	// Look through the list of users we have, see if one of
	// them is allowed in the game
	for _, player := range g.Players {
		if cardsEqual(player.Answer, answer) {
			return player
		}
	}
	return nil
}

func cardsEqual(c1, c2 []Card) bool {
	if len(c1) != len(c2) {
		return false
	}

	for i, card := range c1 {
		if c2[i].Text != card.Text {
			return false
		}
	}
	return true
}

func (g *Game) ClearAnswers() {
	for _, player := range g.Players {
		player.Answer = []Card{}
	}
}

func (g *Game) SetJudge() {
	if g.Judge == nil {
		for _, player := range g.Players {
			if player.LoggedIn {
				g.Judge = player
				return
			}
		}
	}
	seen := false
	for {
		for _, player := range g.Players {
			if seen && player.LoggedIn {
				g.Judge = player
				return
			}
			if player == g.Judge {
				seen = true
			}
		}
	}
}

func (g *Game) JudgeCheck(player *Player) {
	if player == g.Judge && g.GameState != GameOver {
		game.GameState = WaitingOnPlayers
		game.SetJudge()
		game.ClearAnswers()
		game.SendQuestion()
	}
}
