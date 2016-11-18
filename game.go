/*
	a tic tac toe ai generator

	This program uses neuro-evolution to generate an ai for tic tac toe.
*/
package main

import (
	"encoding/gob"
	"fmt"
	gen "github.com/Wouterbeets/genetic"
	flag "gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
)

const (
	P1             = 1
	P2             = -1
	DRAW           = 3
	INPUT_NEURONS  = 10
	OUTPUT_NEURONS = 9
)

//Game is a struct that implements the challenge interface from the genetic package
// it holds the tic tac toe board.
type Game struct {
	gameBoard [9]float64 //the tic tac toe board
	caller    int        //the player whose turn is
	winner    int        //defines winner at the end of a game
}

//String allows Game to implement the stringer interface
func (g *Game) String() string {
	str := fmt.Sprintln(g.gameBoard[0:3])
	str += fmt.Sprintln(g.gameBoard[3:6])
	str += fmt.Sprintln(g.gameBoard[6:9])
	return str
}

//checkWin checks the game board for a winner and returns the winning player
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

//checkRow is used to check for a winner
func checkRow(row []float64) int {
	p := row[0]
	for i := range row {
		if row[i] != p {
			return 0
		}
	}
	return int(p)
}

//hasEmpty is used to check for a draw
func (g *Game) hasEmpty() bool {
	for i, _ := range g.gameBoard {
		if g.gameBoard[i] == 0 {
			return true
		}
	}
	return false
}

func isAllowed(allowedMoves []int, move int) bool {
	for _, v := range allowedMoves {
		if v == move {
			return true
		}
	}
	return false
}

func getBestMove(scores []float64, allowedMoves []int) (move int) {
	bestScore := -1.0
	bestMove := 0
	for move, score := range scores {
		if score > bestScore /* && isAllowed(allowedMoves, move)*/ {
			bestScore = score
			bestMove = move
		}
	}
	return bestMove
}

func (g *Game) move(ai gen.Intel, rounds int) []float64 {
	if g.caller == P2 {
		g.switchBoard()
		input := append(g.gameBoard[:], float64(rounds))
		ai.In(input)
		g.switchBoard()
	} else {
		input := append(g.gameBoard[:], float64(rounds))
		ai.In(input)
	}
	scores := ai.Out()
	return scores
}

//Start makes the two ais passed as parameters fight and returns scores based on their performance
/*
	TODO right now package genetic calls Start 2X so each player can play as P1.
	That logic should be here since genetic should work with any 2p game and not all of them have a P1 advantage
*/

type randomPlayer struct {
	board []float64
}

func (r *randomPlayer) In(board []float64) error {
	r.board = board
	return nil
}

func (r *randomPlayer) Out() []float64 {
	validMoves := []int{}
	for i, val := range r.board {
		if val == 0 && i != 9 {
			validMoves = append(validMoves, i)
		}
	}
	move := validMoves[rand.Intn(len(validMoves))]
	ret := []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	ret[move] = 1
	return ret
}

func (r *randomPlayer) GetName() string {
	return "random Player"
}

func (g *Game) Teach(p1 gen.Intel) float64 {
	var s1, s4 float64
	play1, _ := g.Combat(p1, &randomPlayer{})
	_, play4 := g.Combat(&randomPlayer{}, p1)
	s1 += play1
	s4 += play4
	return (s1 + s4)
}

func (g *Game) Combat(p1, p2 gen.Intel) (s1, s2 float64) {
	defer clean(g)
	log.Printf("%20s vs %20s ", p1.GetName(), p2.GetName())
	bestMove := 0
	var scores []float64
	rounds := 0
	for {
		g.caller = P1
		scores = g.move(p1, rounds)
		bestMove = getBestMove(scores, g.getMoves())
		if g.gameBoard[bestMove] != 0 {
			g.winner = 2
			break
		}
		g.gameBoard[bestMove] = P1
		log.Println(g)
		if g.winner = g.checkWin(); g.winner != 0 {
			break
		}
		rounds++
		g.caller = P2
		scores = g.move(p2, rounds)
		bestMove = getBestMove(scores, g.getMoves())
		if g.gameBoard[bestMove] != 0 {
			g.winner = -2
			break
		}
		g.gameBoard[bestMove] = P2
		log.Println(g)
		if g.winner = g.checkWin(); g.winner != 0 {
			break
		}
		rounds++
	}
	if g.winner == DRAW {
		log.Println("draw")
		s1 = float64(0.4) + (float64(rounds))
		s2 = float64(0.4) + (float64(rounds))
		fmt.Print(".")
	} else if g.winner == P1 {
		log.Println("winner is", p1.GetName())
		s1 = float64(1) + (float64(rounds))
		s2 = float64(0) + (float64(rounds))
		fmt.Print(".")
	} else if g.winner == P2 {
		log.Println("winner is", p2.GetName())
		s1 = float64(0) + (float64(rounds))
		s2 = float64(1) + (float64(rounds))
		fmt.Print(".")
	} else if g.winner == 2 {
		log.Println("p1 broke the rules")
		s1 = float64(0) + (float64(rounds))
		s2 = float64(0) + (float64(9))
	} else if g.winner == -2 {
		log.Println("p2 broke the rules")
		s1 = float64(0) + (float64(9))
		s2 = float64(0) + (float64(rounds))
	}
	return
}

//clean resets the game baard
func clean(g *Game) {
	for i := range g.gameBoard {
		g.gameBoard[i] = 0
	}
}

//getMoves returns a []int containing the indexes of playable tiles
func (g *Game) getMoves() (m []int) {
	m = make([]int, 0, 9)
	for i, _ := range g.gameBoard {
		if g.gameBoard[i] == 0 {
			m = append(m, i)
		}
	}
	return m
}

//switchBoard transforms all P1 values to P2 values and all P2 values to P1 values
// it is used so that ai's player always ahs the value of P1, which might increase learning speed
func (g *Game) switchBoard() {
	for i := range g.gameBoard {
		if g.gameBoard[i] == P1 {
			g.gameBoard[i] = P2
		} else if g.gameBoard[i] == P2 {
			g.gameBoard[i] = P1
		}
	}
}

//recursive minimax function. Depth is set to 1 in the caller so the intelligence is in the ai
// and not in the minimax
//func (g *Game) miniMax(ai *gen.Ai, depth int, turn int) (bestScore float64, move int) {
//	if depth == 0 || g.hasEmpty() == false {
//		if g.caller == P1 {
//			ai.In(g.gameBoard[:])
//		} else if g.caller == P2 {
//			g.switchBoard()
//			ai.In(g.gameBoard[:])
//			g.switchBoard()
//		} else {
//			panic("caller of miniMax != P1 && different from P2")
//		}
//		score := ai.Out()
//		return score[0], 0
//	}
//	moves := g.getMoves()
//	if turn == P1 {
//		bestScore = 0.0
//		for _, m := range moves {
//			g.gameBoard[m] = P1
//			newBest, _ := g.miniMax(ai, depth-1, P2)
//			if newBest >= bestScore {
//				bestScore, move = newBest, m
//			}
//			g.gameBoard[m] = 0
//		}
//	} else {
//		bestScore = math.MaxFloat64
//		for _, m := range moves {
//			g.gameBoard[m] = P2
//			newBest, _ := g.miniMax(ai, depth-1, P1)
//			if newBest <= bestScore {
//				bestScore, move = newBest, m
//			}
//			g.gameBoard[m] = 0
//		}
//	}
//	return
//}

//importAisInfo is the struct used to save ais for later use or parse ais into the program
type importAisInfo struct {
	AiNames []string
	Weights [][]float64
}

//function that fetches  ais from files
func importAis(files []string) (ais importAisInfo) {
	for _, fileName := range files {
		file, err := os.Open(fileName)
		fmt.Println(fileName)
		if err != nil {
			fmt.Println(err, "\nno file with name:", fileName)
			err = nil
			continue
		}
		dec := gob.NewDecoder(file)
		batchOfAis := importAisInfo{}
		err = dec.Decode(&batchOfAis)
		if err != nil {
			fmt.Println("on reading file", fileName, err)
			file.Close()
			continue
		}
		ais.AiNames = append(ais.AiNames, batchOfAis.AiNames[:]...)
		ais.Weights = append(ais.Weights, batchOfAis.Weights[:]...)
		file.Close()
	}
	fmt.Println("total imported ais:", len(ais.AiNames))
	fmt.Println("names:", ais.AiNames)
	//	for _, name := range ais.AiNames {
	//		fmt.Println(name)
	//	}
	return
}

func saveAis(ais []*gen.Ai, saveFile string) {
	//saveFile += strconv.Itoa(*saveSize) + "_npl" + strconv.Itoa(*neuronsPerLayer) + "_layers" + strconv.Itoa(*hiddenLayers) + "_pool" + strconv.Itoa(*poolSize) + "_gen" + strconv.Itoa(*generations)
	file, err := os.OpenFile(saveFile, os.O_CREATE|os.O_RDWR, 0666)
	defer (*file).Close()
	if err != nil {
		panic(err)
	}
	if *saveSize > *poolSize {
		*saveSize = *poolSize
	}
	save := importAisInfo{
		AiNames: make([]string, *saveSize),
		Weights: make([][]float64, *saveSize),
	}
	for i := 0; i < *saveSize; i++ {
		save.AiNames[i] = ais[i].Name
		save.Weights[i] = ais[i].GetWeights()
	}
	enc := gob.NewEncoder(file)
	err = enc.Encode(&save)
	if err != nil {
		panic(err)
	}
}

func play(ais importAisInfo) {
	fmt.Println("choose your opponent")
	for i, name := range ais.AiNames {
		fmt.Println(i, name)
	}
	fmt.Println("type the number")
	num := 0
	fmt.Scanln(&num)
	fmt.Println("you chose", ais.AiNames[num])
	*poolSize = len(ais.AiNames)
	p := gen.CreatePool(*poolSize, *mutation, *mutationStrength, INPUT_NEURONS, *neuronsPerLayer, *hiddenLayers+2, OUTPUT_NEURONS)
	for i, name := range ais.AiNames {
		log.Println("imported ai", i, name)
		p.Ai[i].Name = name
		p.Ai[i].SetWeights(ais.Weights[i])
	}
	op := p.Ai[num]
	g := new(Game)
	p.FightingPit = g
	ex := new(Game)
	fmt.Println(ex)
	fmt.Println("when its your turn type the index of square you want to play")
	ex.gameBoard = [9]float64{0, 1, 2, 3, 4, 5, 6, 7, 8}
	rounds := 0
	for {
		moves := g.move(op, rounds)
		allowedMoves := g.getMoves()
		for _, m := range allowedMoves {
			moves[m] += 1
		}
		fmt.Println(moves)
		move := getBestMove(moves, g.getMoves())
		fmt.Println(op.Name, "plays", move)
		g.gameBoard[move] = P1
		fmt.Println("your turn")
		fmt.Println(g)
		rounds++
		fmt.Scanln(&num)
		g.gameBoard[num] = P2
		rounds++
	}
}

func main() {
	fmt.Println(Files)
	ais := importAis(*Files)
	if *inter {
		play(ais)
	} else {
		*poolSize += len(ais.AiNames)
		p := gen.CreatePool(*poolSize, *mutation, *mutationStrength, INPUT_NEURONS, *neuronsPerLayer, *hiddenLayers+2, OUTPUT_NEURONS)
		g := new(Game)
		//p.FightingPit = g
		p.School = g
		for i, name := range ais.AiNames {
			log.Println("imported ai", i, name)
			p.Ai[i].Name = name
			p.Ai[i].SetWeights(ais.Weights[i])
		}
		p.Evolve(*generations, nil, nil)
		saveAis(p.Ai, *saveFile)
	}
}

//flags
var (
	app              = flag.New("game", "Generate tic tac toe AI's using neuro-evolution")
	Files            = app.Flag("file", "path/to/file1 path/to/file2").Short('f').Default("bestAi").Strings()
	saveFile         = app.Flag("saveFile", "path/to/file1 path/to/file2").Short('s').Default("bestAi").String()
	saveSize         = app.Flag("saveSize", "amount of ais to be saved, may not exceed poolSize").Short('a').Default("10").Int()
	generations      = app.Flag("generations", "generations to train network. ex: 50").Default("500").Short('g').Int()
	mutation         = app.Flag("mutation", "fraction of  genes to be mutated. ex: 0.05").Default("0.05").Short('m').Float64()
	mutationStrength = app.Flag("mutationStrength", "strength of the aplied mutation. ex: 1").Default("1").Short('t').Float64()
	hiddenLayers     = app.Flag("hiddenLayers", "amount of hidden layers in neural network. ex: 1").Default("3").Short('l').Int()
	neuronsPerLayer  = app.Flag("neuronsPerLayer", "neurons per layer. ex: 9").Short('n').Default("9").Int()
	poolSize         = app.Flag("poolSize", "number of neuralNetworks in generation pool").Short('p').Default("100").Int()
	logFile          = app.Flag("log", "path/to/logFile").Short('r').Default("log.txt").String()
	logging          = app.Flag("logging", "should logging be turned on?").Short('z').Default("true").Bool()
	mMDepth          = app.Flag("depth", "miniMax depth, how many moves the algorithm looks ahead before it feeds the board positions to the neural network").Short('d').Default("1").Int()
	inter            = app.Flag("interactive", "play against a evolved ai").Short('i').Default("false").Bool()
)

//initialise logfiles and command line flags
func init() {
	app.Parse(os.Args[1:])
	if *logging {
		f, err := os.Create(*logFile)
		if err != nil {
			panic(err)
		}
		log.SetFlags(0)
		log.SetOutput(f)
	} else {
		dis := ioutil.Discard
		log.SetOutput(dis)
	}
	fmt.Println("Game starting\nPoolsize :\t", *poolSize)
	fmt.Println("Generations :\t", *generations)
	fmt.Println("neuronsPL :\t", *neuronsPerLayer)
}
