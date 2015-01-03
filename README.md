# Online Cards Against Humanity

This is a Go-based version of Cards Against Humanity

## To Use:

1. Supply your own `cards.json` file in the following format:

```JSON
[
  {"id":1,"cardType":"A","text":"Flying sex snakes.","numAnswers":0,"expansion": "Base"},
  {"id":2,"cardType":"A","text":"Michelle Obama's arms.","numAnswers":0,"expansion": "Base"},
  {"id":3,"cardType":"A","text":"German dungeon porn.","numAnswers":0,"expansion": "Base"},
  {"id":4,"cardType":"Q","text":"Patient presents with _. Likely a result of _.","numAnswers":2,"expansion":"CAHe5"},
  {"id":5,"cardType":"Q","text":"Hi MTV! My name is Kendra, I live in Malibu, I'm into _, and I love to have a good time.","numAnswers":1,"expansion":"CAHe5"},
  {"id":6,"cardType":"Q","text":"Help me doctor, I've got _ in my butt!","numAnswers":1,"expansion":"CAHe5"},
  {"id":7,"cardType":"Q","text":"Why am I broke?","numAnswers":1,"expansion":"CAHe5"},
]
```

2. Supply your own `players.json` file in the following format:

```JSON
[
  {"name":"Pat", "password":"Jim", "img":"pk"},
  {"name":"Brandon", "password":"Biz", "img":"bs"},
  {"name":"Evan", "password":"Donut", "img":"eb"},
  {"name":"John", "password":"Juan", "img":"jc"},
]
```
3. Add an `img` directory with images corresponding to the names above (with a png extension)

4. If you want the winner to receive a message (or code or something), place it in the `code` file

5. Set the WinnerScore to the number of point required to win.

6. If you're running this standalone (not behind Apache or Nginx), uncomment these lines in `main.go`

```Go
http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))
```

## TODO

* Separate CAH portion from server to its own library
* Clean up API endpoints for transition to Angular
* Clean up state diagram
* Fix infinite loop issue in extreme edge case
* Make card dealing "smarter"
