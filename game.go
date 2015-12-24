package main

import (
	"encoding/gob"
	"fmt"
	"github.com/WouterBeets/gen"
	"log"
	"math"
	"os"
)

const (
	P1   = 1
	P2   = 2
	DRAW = 3
)

type Game struct {
	move      chan int
	gameBoard []float64
	size      int
	caller    int
	winner    int
}

type lastRound struct {
	AiNames []string
	Weights [][]float64
}

func NewGame() *Game {
	g := &Game{
		move:      make(chan int),
		gameBoard: make([]float64, 9, 9),
		size:      3,
	}
	return g
}

func checkRow(row []float64) int {
	p := row[0]
	for i := range row {
		if row[i] != p {
			return 0
		}
	}
	return int(p)
}

func (g *Game) hasEmpty() bool {
	for i, _ := range g.gameBoard {
		if g.gameBoard[i] == 0 {
			return true
		}
	}
	return false
}

func (g *Game) String() string {
	str := fmt.Sprintln(g.gameBoard[0:3])
	str += fmt.Sprintln(g.gameBoard[3:6])
	str += fmt.Sprintln(g.gameBoard[6:9])
	return str
}

func (g *Game) checkWin() (winner int) {
	checks := [8]int{}
	checks[0] = checkRow(g.gameBoard[0:3])
	checks[1] = checkRow(g.gameBoard[3:6])
	checks[2] = checkRow(g.gameBoard[6:9])
	checks[3] = checkRow([]float64{g.gameBoard[0], g.gameBoard[4], g.gameBoard[8]})
	checks[4] = checkRow([]float64{g.gameBoard[2], g.gameBoard[4], g.gameBoard[6]})
	checks[5] = checkRow([]float64{g.gameBoard[0], g.gameBoard[3], g.gameBoard[6]})
	checks[6] = checkRow([]float64{g.gameBoard[1], g.gameBoard[4], g.gameBoard[7]})
	checks[7] = checkRow([]float64{g.gameBoard[2], g.gameBoard[5], g.gameBoard[8]})
	for i, _ := range checks {
		if checks[i] != 0 {
			return checks[i]
		}
	}
	if g.hasEmpty() == false {
		return DRAW
	}
	return 0
}

func clean(g *Game) {
	for i := range g.gameBoard {
		g.gameBoard[i] = 0
	}
}

func (g *Game) Start(p1, p2 *gen.Ai) (s1, s2 float64) {
	defer clean(g)
	for {
		g.caller = P1
		_, m := g.miniMax(p1, 2, P1)
		g.gameBoard[m] = P1
		log.Println(g)
		if g.winner = g.checkWin(); g.winner != 0 {
			break
		}
		g.caller = P2
		_, m = g.miniMax(p2, 2, P2)
		g.gameBoard[m] = P2
		log.Println(g)
		if g.winner = g.checkWin(); g.winner != 0 {
			break
		}
	}
	log.Println("winner is", g.winner)
	if g.winner == DRAW {
		s1 = 0.50
		s2 = 0.50
	} else if g.winner == P1 {
		s1 = 1
		s2 = 0
	} else {
		s1 = 0
		s2 = 1
	}
	return
}

func (g *Game) getMoves() (m []int) {
	m = make([]int, 0, 9)
	for i, _ := range g.gameBoard {
		if g.gameBoard[i] == 0 {
			m = append(m, i)
		}
	}
	return m
}

func (g *Game) miniMax(ai *gen.Ai, depth int, turn int) (best float64, move int) {
	if depth == 0 || g.hasEmpty() == false {
		if g.caller == P1 {
			ai.In(g.gameBoard)
		} else {
			for i := range g.gameBoard {
				if g.gameBoard[i] == P1 {
					g.gameBoard[i] = P2
				} else if g.gameBoard[i] == P2 {
					g.gameBoard[i] = P1
				}
			}
			ai.In(g.gameBoard)
			for i := range g.gameBoard {
				if g.gameBoard[i] == P1 {
					g.gameBoard[i] = P2
				} else if g.gameBoard[i] == P2 {
					g.gameBoard[i] = P1
				}
			}
		}
		score := ai.Out()
		return score[0], 0
	}
	if turn == P1 {
		best = 0.0
		moves := g.getMoves()
		for _, m := range moves {
			g.gameBoard[m] = P1
			newBest, _ := g.miniMax(ai, depth-1, P2)
			if newBest >= best {
				best, move = newBest, m
			}
			g.gameBoard[m] = 0
		}
	} else {
		best = math.MaxFloat64
		moves := g.getMoves()
		for _, m := range moves {
			g.gameBoard[m] = P2
			newBest, _ := g.miniMax(ai, depth-1, P1)
			if newBest <= best {
				best, move = newBest, m
			}
			g.gameBoard[m] = 0
		}
	}
	return
}

func main() {
	f, err := os.Create("log.txt")
	if err != nil {
		log.Panic(err)
	}
	log.SetFlags(0)
	log.SetOutput(f)
	g := NewGame()
	file, err := os.Open("bestAi.gob")
	ais := []lastRound{}
	if err == nil {
		dec := gob.NewDecoder(file)
		oldAis := new(lastRound)
		for {
			err := dec.Decode(oldAis)
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println("decoding")
			fmt.Println(oldAis)
			ais = append(ais, *oldAis)
		}
		file.Close()
	}
	nAis := 0
	for _, v := range ais {
		nAis += len(v.AiNames)
	}
	if nAis == 0 {
		nAis = 500
	}
	p := gen.CreatePool(nAis*2, 0.05, 1, 9, 9, 4, 1)
	p.Chal = g
	nAis = 0
	for _, ai := range ais {
		for _, Name := range ai.AiNames {
			fmt.Println(Name)
			p.Ai[nAis].Name = Name
			p.Ai[nAis].SetWeights(ai.Weights[nAis])
			nAis++
		}
	}
	fmt.Println()
	p.Evolve(100, nil, nil)
	file, err = os.OpenFile("bestAi.gob", os.O_APPEND|os.O_RDWR, 666)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	save := lastRound{
		AiNames: []string{
			p.Ai[0].Name,
			p.Ai[1].Name,
			p.Ai[2].Name,
			p.Ai[3].Name,
			p.Ai[4].Name,
		},
		Weights: [][]float64{
			p.Ai[0].GetWeights(),
			p.Ai[1].GetWeights(),
			p.Ai[2].GetWeights(),
			p.Ai[3].GetWeights(),
			p.Ai[4].GetWeights(),
		},
	}
	enc := gob.NewEncoder(file)
	err = enc.Encode(&save)
	if err != nil {
		panic(err)
	}
}
