package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"os"
)

type State int

// The number of points needed to win
const WinnerScore = 2

const (
	NotStarted State = iota
	Initializing
	WaitingOnJudge
	WaitingOnPlayers
	GameOver
)

var templates = template.Must(template.ParseGlob("templates/*"))

var game = CreateGame("cards.json", "players.json")

func main() {
	go h.run()

	r := mux.NewRouter()

	r.HandleFunc("/", mainHandler).Methods("GET")
	r.HandleFunc("/status", statusHandler).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/startitup", startHandler).Methods("GET")
	r.HandleFunc("/ws", serveWs)

	r.HandleFunc("/cards", cardHandler).Methods("GET")
	r.HandleFunc("/answer", answerHandler).Methods("POST")
	r.HandleFunc("/judge", judgeHandler).Methods("POST")

	// Depending on whether or not you're using a proxy, you might need this
	// server to serve static assets, and you'll have to uncomment the next three
	// lines

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))

	http.Handle("/", r)

	log.Println("Starting server...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3100"
	}

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// Rendering the main page
func mainHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", r.Host)
	if err != nil {
		log.Print("Error executing template:", err)
	}
}

func cardHandler(w http.ResponseWriter, r *http.Request) {
	cardPlayer := game.PlayerByHandshake(r.FormValue("handshake"))

	// We're going to send the client some JSON, bruh
	w.Header().Set("Content-Type", "application/json")

	// If we couldn't find the user, we send them an error
	if cardPlayer == nil {
		err := struct {
			Error string
		}{
			"Invalid Login",
		}
		resp, _ := json.Marshal(err)
		w.Write(resp)
		return
	} else {
		ok := struct {
			Cards []Card
		}{
			cardPlayer.Hand,
		}
		resp, _ := json.Marshal(ok)
		w.Write(resp)
		return
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Get password from request, no SSL and plain-text because
	// I like to live dangerously, and this isn't particularly
	// sensitive information
	password := r.PostFormValue("password")

	var newPlayer *Player
	// Look through the list of users we have, see if one of
	// them is allowed in the game
	for _, player := range game.Players {
		if player.Password == password {
			newPlayer = player
			break
		}
	}

	// We're going to send the client some JSON, bruh
	w.Header().Set("Content-Type", "application/json")

	// If we couldn't find the user, we send them an error
	if newPlayer == nil {
		err := struct {
			Error string
		}{
			"Try again, idiot",
		}
		resp, _ := json.Marshal(err)
		w.Write(resp)
		return
	}

	// If they're already logged in, we send them an error
	if newPlayer.LoggedIn {
		err := struct {
			Error string
		}{
			"Already logged In",
		}
		resp, _ := json.Marshal(err)
		w.Write(resp)
		return
	}

	ok := struct {
		Status    string
		Name      string
		Handshake string
	}{
		"Ok",
		newPlayer.Name,
		newPlayer.Handshake,
	}
	resp, _ := json.Marshal(ok)
	w.Write(resp)
}

func answerHandler(w http.ResponseWriter, r *http.Request) {
	if game.GameState == WaitingOnPlayers {
		jsonString := r.PostFormValue("json")
		reqInfo := struct {
			Handshake string
			Cards     []int
		}{}

		json.Unmarshal([]byte(jsonString), &reqInfo)
		player := game.PlayerByHandshake(reqInfo.Handshake)

		ansCards := make([]Card, len(reqInfo.Cards))
		newCards := game.Deck.DrawAnswerCards(len(reqInfo.Cards))

		// Move the cards they chose into a new hand, and overwrite
		// those card with new cards
		for i, index := range reqInfo.Cards {
			ansCards[i] = player.Hand[index]
			player.Hand[index] = newCards[i]
		}

		player.Answer = ansCards

		w.Header().Set("Content-Type", "application/json")
		data := struct {
			Cards []Card
		}{
			player.Hand,
		}

		h.broadcast <- refreshMessageJSON()
		resp, _ := json.Marshal(data)
		w.Write(resp)

		// If everyone has answered
		if game.AllIn() {
			game.GameState = WaitingOnJudge
			game.SendAnswers()
		}
	}
}

func judgeHandler(w http.ResponseWriter, r *http.Request) {
	if game.GameState == WaitingOnJudge {
		jsonString := r.PostFormValue("json")
		reqInfo := struct {
			Handshake string
			Answer    []Card
		}{}

		json.Unmarshal([]byte(jsonString), &reqInfo)
		player := game.PlayerByHandshake(reqInfo.Handshake)

		// Only let them pick the winner if they're the judge
		if player == game.Judge {
			winner := game.PlayerByAnswers(reqInfo.Answer)
			if winner != nil {
				winner.Score++
				if winner.Score == WinnerScore {
					// Give them the code
					game.GameState = GameOver
					h.broadcast <- gameOverJSON()

					for c := range h.connections {
						if c.player == winner {
							c.send <- winnerJSON()
						}
					}
				} else {
					// New round of questions
					game.GameState = WaitingOnPlayers
					game.SetJudge()
					game.SendAnswerAndQuestion(winner)
					game.ClearAnswers()
					h.broadcast <- refreshMessageJSON()
				}
			}
		}
	}
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")

	if code == "letsgetthispartystarted" {
		game.StartGame()
		w.Write([]byte("Here we go"))
	} else {
		w.Write([]byte("GTFO with your cheating shit"))
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	statuses := make([]PlayerStatus, len(game.Players))
	handshake := r.FormValue("handshake")
	me := ""

	state := ""
	for i, player := range game.Players {
		if player.LoggedIn {
			statuses[i] = PlayerStatus{
				player.Name,
				player.Img,
				player.Score,
				player == game.Judge,
				len(player.Answer) != 0,
			}
		} else {
			statuses[i] = PlayerStatus{}
		}

		if player.Handshake == handshake {
			me = player.Name
			state = makeStatus(player)
		}
	}

	data := struct {
		Players     []PlayerStatus
		Me          string
		State       string
		WinnerScore int
	}{
		statuses,
		me,
		state,
		WinnerScore,
	}

	w.Header().Set("Content-Type", "application/json")
	resp, _ := json.Marshal(data)
	w.Write(resp)
	return
}
