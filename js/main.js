var conn;
var judge = false;
var me = "";
var cards = [];
var handshake = "";
var cardCount = 0;
var qCount = 0;
var winner = false;
var indices = [];

var UNKNOWN = "Unknown";
var NOT_STARTED = "Not started";
var WAITING_ON_JUDGE = "Waiting on Judge";
var WAITING_ON_OTHER_PLAYERS = "Waiting on other players";
var WAITING_ON_YOU = "Waiting on you";
var GAME_OVER = "Game is over";


var state = UNKNOWN;

// Foundation likes you to hold it's hand
$(document).foundation();

$(function() {
  $('.leaderboard').on('mouseenter', '.section', function() {
    $(this).stop();
    $(this).animate({'background-position-x': '0%'}, 5000, "linear");
  })
  .on('mouseleave', '.section', function() {
    $(this).stop();
    $(this).animate({'background-position-x': '50%'}, 1000, "linear");
  });

  $(window).resize(renderCards);

  $('#loginForm').ajaxForm({
    success: function(response, statusText, xhr, form) {
      handshake = response.Handshake;
      if (typeof handshake !== "undefined") {
        // Load their hand if we expect the handshake to succeed
        $.ajax('/cards', {
          data: {handshake: handshake},
          success: function(data) {
            if (typeof data.Cards !== "undefined" && data.Cards != null) {
              cards = data.Cards;
              renderCards();
            } else {
              cards = [];
            }
          }
        });
        if (Modernizr.websockets) {
          var loc = window.location, newUri;
          if (loc.protocol === "https:") {
              newUri = "wss:";
          } else {
              newUri = "ws:";
          }
          newUri += "//" + loc.host;
          newUri += loc.pathname + "ws?handshake=";
          conn = new WebSocket(newUri + handshake);
          conn.onerror = function(evt) {
            showError("Trying to sneak past the handshake asshole?");
          }
          conn.onclose = function(evt) {
            console.log(evt);
          }
          conn.onmessage = function(evt) { // Message received. evt.data is something
            $('#loginModal').foundation('reveal', 'close');
            var data = JSON.parse(evt.data);
            
            switch (data.Action) {
              case "refresh":
                updateBoard();
                break;
              // Cards have just been dealt
              case "dealt":
                cards = data.Data.Cards;
                renderCards();
                break;
              // A question has just been issued
              case "question":
                // We're the judge if we're the judge, duh
                judge = (me == data.Data.Judge);
                qCount = data.Data.Question.NumAnswers;
                $('.answers').empty();
                $('.question').text(data.Data.Question.Text);
                if (!judge) {
                  setState(WAITING_ON_YOU, "Pick a card");
                } else {
                  setState(WAITING_ON_OTHER_PLAYERS, "Waiting for people to pick cards");
                }
                break;
              case "answerquestion":
                var text = "";
                for (var i = 0; i < data.Data.Answer.length; i++) {
                  text += data.Data.Answer[i].Text;
                }

                var correctButton;
                $('.answers .button').each(function() {
                  if (text == $(this).text()) {
                    correctButton = $(this);
                    return false;
                  }
                });

                correctButton.css({backgroundColor: 'green'});
                setTimeout(function() {
                  judge = (me == data.Data.Judge);
                  qCount = data.Data.Question.NumAnswers;
                  $('.answers').empty();
                  $('.question').text(data.Data.Question.Text);
                  if (!judge) {
                    setState(WAITING_ON_YOU, "Pick a card");
                  } else {
                    setState(WAITING_ON_OTHER_PLAYERS, "Waiting for people to pick cards");
                  }
                }, 1000);
                break;
              case "answers":
                showAnswers(data.Data.Answers);
                if (!judge) {
                  setState(WAITING_ON_JUDGE, "Waiting on the judge to choose");
                } else {
                  setState(WAITING_ON_YOU, "Pick a winner");
                }
                break;
              case "gameover":
                if (!winner) {
                  state = GAME_OVER;
                  $('.question').text("");
                  $('.answers').empty();
                  $('.cards').empty();
                  updateBoard(function() {
                    setState(null, "Game Over");
                  });
                }
                break;
              case "winner":
                winner = true;
                state = GAME_OVER;
                $('.question').text("The code is: " + data.Data.Code);
                $('.answers').empty();
                $('.cards').empty();
                updateBoard(function() {
                  setState(null, "Winner");
                });
                break;
            }
          }
        } else {
          showError("Somehow, you don't have WebSockets and can't play. Are you using IE 6 or some shit?");
        }
      } else {
        showError(response.Error);
      }
    }
  });
  $('#loginModal').foundation('reveal', 'open');

  $('.cards').on('click', '.card-inner', function() {
    var card = $(this);
    if (!judge &&
        !card.hasClass('selected') &&
        indices.length < qCount &&
        state == WAITING_ON_YOU) {
      card.addClass('selected');
      indices.push($('.card-inner').index($(this)));
      if (indices.length == qCount) {
        setState(WAITING_ON_OTHER_PLAYERS, "Waiting on other players");
        var jsonString = JSON.stringify({handshake: handshake, cards: indices});
        $.post('/answer', {json: jsonString}, function(data) {
          indices = []; 
          cards = data.Cards;
          renderCards();
        });
      }
    }
  });
});

// Updates the score of each player, and who the judge is
function updateBoard(callback) {
  $.get('/status', {handshake: handshake}, function(data) {
    var players = data.Players;
    if (data.State != "") {
      $('.status').text(data.State);
      if (state == UNKNOWN) {
        state = data.State;
      }
    }
    if (data.Me !== "") {
      me = data.Me;
    }
    var leaderboard = $('.leaderboard').empty();
    for (var i = 0; i < players.length; i++) {
      var section = $('<div class="section"><div class="prog"></div><div class="score"></div><div class="judge hide">Judge</div><div class="name"></div></div>');
      section.css({
        width: (100/players.length) + "%",
      });

      
      if (players[i].Img != "") {
        section.css({
          'background-image': 'url(img/' + players[i].Img + '.png)'
        });

        if(players[i].Judge) {
          section.find('.judge').removeClass('hide');
        } else {
          // If you aren't the judge, check if you've voted
          if(players[i].Answered) {
            section.addClass('answered');
          } else {
            section.addClass('not-answered');
          }
        }
        var progress = section.find('.prog');
        var name = section.find('.name');
        var score = section.find('.score');
        var points = players[i].Score
        progress.css({height: points*100/data.WinnerScore + '%'});
        score.text(points + "/" + data.WinnerScore);
        name.text(players[i].Name);
      }
      leaderboard.append(section);
    }

    if (typeof callback !== "undefined") {
      callback();
    }
  });
}

function showError(text) {
  $('.error-text').text(text);
  $('.alert-box').removeClass("hide");
}

function renderCards() {
  var cardHolder = $('.cards');
  cardHolder.empty();
  for (var i = 0; i < cards.length; i++) {
    var card = $('<div class="card-outer"><div class="card-inner"><div class="card-text"></div></div></div>');
    var cardText = card.find('.card-text');
    cardText.text(cards[i].Text);
    cardHolder.append(card);
    card.width(card.height()*5/7);
  }
}

function setState(newState, stat) {
  if (state != UNKNOWN && state != null) {
    state = newState
  }
  $('.status').text(stat);
}

function showAnswers(answers) {
  var holder = $('.answers');

  for (var i = 0; i < answers.length; i++) {
    var text = "";
    var card = $('<div class="button expand"></div>');
    for (var j = 0; j < answers[i].length; j++) {
      text += answers[i][j].Text;
      text += "\n";
    }
    card.text(text);
    card.html(card.html().replace(/\n/g,'<br/>'));
    holder.append(card);
    (function(x) {
      card.click(function() {
        var jsonString = JSON.stringify({handshake: handshake, answer: answers[x]});
        $.post('/judge', {json: jsonString}, function(data) {
          // Error checking? Or do you even give a shit?
        });
      });
    })(i);
  }
}
