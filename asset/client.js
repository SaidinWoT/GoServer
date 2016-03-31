
var boardSelectElem = document.getElementById("boardSelect");
var boardElem = document.getElementById("board");

function joinedGame(gameid) {
  // Called when the client is told they joined a game
  // It's expected that a boardUpdate will come in after this.
  window.location.hash = gameid;
  boardSelectElem.className = "hidden";
}

function boardUpdate(board) {
  // Called when the client receives a new board from the server
  // Replace the content of [boardElem] with a newly created table
  // Also wires up board to place stones on clicks.
  var size = board.Size;
  boardElem.className = "board board" + size;

  while(boardElem.firstChild) {
    boardElem.removeChild(boardElem.firstChild);
  }

  var table = buildTable(board, function(r,c){
    socket.emit('place stone', r, c)
  });
  boardElem.appendChild(table);
}

function buildTable(board, onClick) {
  // Construct an HTML table element that is populated from [board].
  // table tableN - classes for a size N table
  // stone stoneN - classes for each stone. N==0(empty), 1(bl), 2(wh)
  // onClick(r,c) - called when a user clicks on the cell at (r,c)
  var table = document.createElement("table");
  var size = board.Size;

  table.className = "table table" + size;
  var idx = 0;
  for(var r = 0 ; r < size ; r++){
    var row = document.createElement("tr");
    table.appendChild(row);

    for(var c = 0 ; c < size ; c++) {
      var cell = document.createElement("td");
      row.appendChild(cell);
      cell.className = "stone stone" + board.Stones[idx++];
      cell.onclick = (function(r,c){return function(){
        onClick(r,c);
      }})(r,c);
    }
  }
  return table;
}

function newgame(size) {
  // Called by the client when they select a board size
  socket.emit('new game', size);
}




var socket = io();
socket.on('joined game', joinedGame);
socket.on('error', console.warn.bind(console));
socket.on('board update', boardUpdate);

var game = parseInt(window.location.hash.substring(1));
if(game)
  socket.emit('join game', game)

