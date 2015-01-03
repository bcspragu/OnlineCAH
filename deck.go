package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

type Card struct {
	CardType   string
	Text       string
	NumAnswers int
}

type Deck struct {
	AnswerCards []Card
	AnswerDraw  []int
	AnswerIndex int

	QuestionCards []Card
	QuestionDraw  []int
	QuestionIndex int

	r *rand.Rand
}

func NewDeck(filename string) *Deck {
	var cards []Card
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Print("Error reading json:", err)
	}
	answerCards := make([]Card, 0, len(cards))
	questionCards := make([]Card, 0, len(cards))

	err = json.Unmarshal(dat, &cards)
	if err != nil {
		log.Printf("Apparently you can't unmarshal cards", err)
	}

	for _, card := range cards {
		if card.CardType == "A" {
			answerCards = append(answerCards, card)
		} else {
			questionCards = append(questionCards, card)
		}
	}

	deck := &Deck{
		AnswerCards:   answerCards,
		QuestionCards: questionCards,
	}

	deck.r = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	deck.ShuffleQuestionCards()
	deck.ShuffleAnswerCards()

	return deck
}

func (d *Deck) DrawAnswerCard() Card {
	return d.DrawAnswerCards(1)[0]
}

func (d *Deck) DrawQuestionCard() Card {
	card := d.QuestionCards[d.QuestionDraw[d.QuestionIndex]]
	d.QuestionIndex++
	return card
}

func (d *Deck) DrawAnswerCards(n int) []Card {
	cards := make([]Card, n)

	// Shuffle the deck before we deal them if we don't have
	// enough left
	if (len(d.AnswerCards) - d.AnswerIndex) < n {
		d.ShuffleAnswerCards()
	}

	// Deal out n cards
	for i := 0; i < n; i++ {
		cards[i] = d.AnswerCards[d.AnswerDraw[d.AnswerIndex]]
		d.AnswerIndex++
	}

	return cards
}

func (d *Deck) ShuffleQuestionCards() {
	d.QuestionIndex = 0
	d.QuestionDraw = d.r.Perm(len(d.QuestionCards))
}

func (d *Deck) ShuffleAnswerCards() {
	d.AnswerIndex = 0
	d.AnswerDraw = d.r.Perm(len(d.AnswerCards))
}
