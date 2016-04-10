package main

import (
  "errors"
	"log"
	"net/http"
  "strconv"

	"github.com/googollee/go-socket.io"
)

var gameid = 0

func main() {
	games = make(map[int]*Game)
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

  ider := LinearIDer{}
	server.On("connection", connect(ider))
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
	Empty Stone = iota
	Black
	White
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

type IDer interface {
  ID() int
}

type LinearIDer struct{
  int
}

func (l LinearIDer) ID() int {
  l.int++
  return l.int
}

type SocketListener struct {
  channel string
  IDer
  *Game
  socketio.Socket
}

func connect(ider IDer) func (socketio.Socket) {
  sl := &SocketListener{
    IDer: ider,
  }
  return func(so socketio.Socket) {
    sl.Socket = so
    sl.On("new game", sl.MakeGame)
    sl.On("join game", sl.JoinGame)
    sl.On("place stone", sl.PlaceStone)
  }
}

func (sl *SocketListener) Error(err string) {
  sl.Emit("error", err)
}

func (sl *SocketListener) MakeGame(size int) {
  if size != 9 && size != 13 && size != 19 {
    sl.Error("Bad game size")
    return
  }
  gameID := sl.IDer.ID()
  games[gameID] = &Game{
    board: fakeBoard(size),
    turn:  Black,
  }
  
  sl.JoinGame(gameID)
}

func (sl *SocketListener) JoinGame(gameID int) {
  game, ok := games[gameID]
  if !ok {
    sl.Error("Bad game ID")
    return
  }
  channelID := "game" + strconv.Itoa(gameID)
  sl.channel = channelID
  sl.Join(channelID)
  sl.Emit("joined game", gameID)

  sl.Game = game
  sl.updateBoard(false)
}

func (sl *SocketListener) PlaceStone(row, col int) {
  if err := sl.Game.PlaceStone(row, col); err != nil {
    sl.Error(err.Error())
    return
  }
  sl.updateBoard(true)
}

func (sl *SocketListener) updateBoard(global bool) {
  if global {
    sl.BroadcastTo(sl.channel, "board update", sl.Game.board)
  }
  sl.Emit("board update", sl.Game.board)
}

func (g *Game) PlaceStone(row, col int) error {
  if g == nil {
    return errors.New("Nil game.")
  }
  pos := g.board.Size*row + col
  if pos < 0 || pos >= g.board.Size * g.board.Size {
    return errors.New("Invalid position")
  }
  if g.board.Stones[pos] != Empty {
    return errors.New("Position already taken")
  }
  g.board.Stones[pos] = g.turn
  if g.turn == White {
    g.turn = Black
  } else {
    g.turn = White
  }
  return nil
}

func fakeBoard(size int) Board {
	b := Board{size, make([]Stone, size*size)}
	/*b.Stones[1] = WHITE
	b.Stones[2] = BLACK
	b.Stones[size] = WHITE
	b.Stones[size*size-1] = BLACK*/
	return b
}
