package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	GameInProgress = "Game in progress"
	GameOver       = "Game over"
)

type game struct {
	expression string
	result     int
	status     string
}

type client chan<- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	currentGame := initGame("", "", 0)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
			go checkAnswer(msg, len(clients), currentGame)
		case cli := <-entering:
			clients[cli] = true
			go newGame(len(clients), currentGame)
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func initGame(exp, st string, res int) *game {
	return &game{
		expression: exp,
		status:     st,
		result:     res,
	}
}

func generateNewExpression() (string, int) {
	rand.Seed(time.Now().UnixNano())
	operators := []string{"+", "-", "*", "/"}
	op := operators[rand.Intn(len(operators))]
	num1 := rand.Intn(100)
	num2 := 1 + rand.Intn(100)
	exp := fmt.Sprintf("new game: %d%s%d=?", num1, op, num2)
	var result int
	switch op {
	case "+":
		result = num1 + num2
	case "-":
		result = num1 - num2
	case "*":
		result = num1 * num2
	case "/":
		result = num1 / num2
	}

	return exp, result
}

func checkAnswer(msg string, numOfClients int, g *game) {
	if g.status != GameInProgress {
		return
	}
	msgParts := strings.Split(msg, ":")
	if len(msgParts) < 2 {
		return
	}
	answer, err := strconv.Atoi(msgParts[1])
	if err != nil {
		return
	}
	if answer == g.result {
		g.status = GameOver
		messages <- msgParts[0] + " win!"
		newGame(numOfClients, g)
	}
}

func newGame(numOfClients int, g *game) {
	if numOfClients > 1 && g.status != GameInProgress {
		g.expression, g.result = generateNewExpression()
		g.status = GameInProgress
		messages <- g.expression
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	// who := conn.RemoteAddr().String()
	// ch <- "You are " + who
	var who string
	ch <- "Enter your nickname: "
	inputNick := bufio.NewScanner(conn)
	if inputNick.Scan() {
		who = inputNick.Text()
	}
	ch <- "You are " + who
	messages <- who + " has arrived"
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- who + ":" + input.Text()
	}
	leaving <- ch
	messages <- who + " has left"
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
