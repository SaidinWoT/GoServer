package main

import (
	"fmt"
	"github.com/googollee/go-socket.io"
	"log"
	"net/http"
)

var gameid = 0

func main() {
	games = make(map[int]*Game)
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", handleClient)
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

type Stone int

const (
	EMPTY Stone = iota
	BLACK
	WHITE
)

type Board struct {
	Size   int
	Stones []Stone
}

type Game struct {
	board Board
	turn  Stone
}

var games map[int]*Game

func handleClient(so socketio.Socket) {
	var currentGame = -1
	var channel = ""

	so.On("new game", func(size int) {
		log.Println("new game", size)
		switch size {
		case 9, 13, 19:
			// Actually make a game and stuff.
			gameid += 1
			currentGame = gameid

			games[currentGame] = &Game{
				fakeBoard(size),
				BLACK,
			}

			channel = fmt.Sprint("game", gameid)
			so.Join(channel)
			so.Emit("joined game", currentGame)
			so.Emit("board update", games[currentGame].board)
		default:
			so.Emit("error", "Bad game size")
		}
	})

	so.On("join game", func(gameid int) {
		log.Println("join game", gameid)
		// black goes first, then white, then spectators?
		game, ok := games[gameid]
		if ok {
			currentGame = gameid
			channel = fmt.Sprint("game", gameid)
			so.Join(channel)
			so.Emit("joined game", currentGame)
			so.Emit("board update", game.board)
		} else {
			so.Emit("error", "Bad game id")
		}

	})

	so.On("place stone", func(row, col int) {
		if currentGame == -1 {
			so.Emit("error", "Not in a game")
			return
		}
		game, ok := games[currentGame]
		if !ok {
			so.Emit("error", "Bad game id")
			return
		}
		board := game.board
		log.Println("Place stone", row, col, currentGame)
		pos := board.Size*row + col
		if pos < 0 || pos >= board.Size*board.Size {
			so.Emit("error", "Invalid position")
			return
		}

		if board.Stones[pos] != EMPTY {
			so.Emit("error", "Position already taken")
			return
		}

		board.Stones[board.Size*row+col] = game.turn
		if game.turn == WHITE {
			game.turn = BLACK
		} else {
			game.turn = WHITE
		}
		// Send to everyone
		so.BroadcastTo(channel, "board update", board)
		so.Emit("board update", board)
	})
	so.On("disconnection", func() {
		log.Println("on disconnect")
	})
}

func fakeBoard(size int) Board {
	b := Board{size, make([]Stone, size*size)}
	/*b.Stones[1] = WHITE
	b.Stones[2] = BLACK
	b.Stones[size] = WHITE
	b.Stones[size*size-1] = BLACK*/
	return b
}
