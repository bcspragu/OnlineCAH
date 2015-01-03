package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Player struct {
	Name      string
	Hand      []Card
	Password  string
	LoggedIn  bool
	Score     int
	Handshake string
	Img       string
	Answer    []Card
}

func LoadPlayers(filename string) []*Player {
	var players []*Player

	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Print("Error reading json:", err)
	}

	json.Unmarshal(dat, &players)

	for _, player := range players {
		player.Handshake = randSeq(20)
		player.Answer = []Card{}
	}

	return players
}
