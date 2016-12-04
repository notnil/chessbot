package main

import (
	"fmt"
	"time"

	"github.com/notnil/chess"
)

type inputPlayer struct{}

func (p inputPlayer) Move(gs *chess.Position) string {
	fmt.Println("Enter Move in Algebraic Notation (ex. 'e4'):")
	input := ""
	fmt.Scanln(&input)
	return input
}

func main() {
	g := chess.NewGame()
	p := inputPlayer{}
	cpu := Bot{UseBook: true}
	count := 0
	for g.Outcome() == chess.NoOutcome {
		fmt.Println(g.Position().Board().Draw())
		fmt.Println(g.Position().String())
		if count%2 == 0 {
			alg := p.Move(g.Position())
			if err := g.MoveStr(alg); err != nil {
				fmt.Println("invalid move")
				continue
			}
		} else {
			t1 := time.Now()
			move, score := cpu.Move(g)
			g.Move(move)
			secs := time.Now().Sub(t1).Seconds()
			fmt.Printf("CPU moved %s in %.f secs, scored %.f\n", move.String(), secs, score)
		}
		count++
	}
	fmt.Println(g.Position().Board().Draw())
	fmt.Printf("Game completed. %s by %s.\n", g.Outcome(), g.Method())
	fmt.Println(g.String())
}
