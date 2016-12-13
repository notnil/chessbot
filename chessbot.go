package main

import (
	"sort"

	"github.com/notnil/chess"
	"github.com/notnil/opening"
)

type Bot struct {
	Debug   bool
	UseBook bool
}

func (b *Bot) Move(game *chess.Game) (*chess.Move, float64) {
	if game == nil {
		return nil, 0.0
	}
	var move *chess.Move
	if b.UseBook {
		move = openingMove(game)
		if move != nil {
			return move, 0.0
		}
	}
	pos := game.Position()
	turn := pos.Turn()
	moves := pos.ValidMoves()
	sort.Sort(byMoveImportance(moves))
	type moveScore struct {
		move  *chess.Move
		score float64
	}
	ch := make(chan moveScore)
	for _, m := range moves {
		go func(scr *scorer, m *chess.Move) {
			newPos := pos.Update(m)
			score := alphaBetaMin(scr, newPos, -10000.0, 10000.0, 5)
			ch <- moveScore{m, score}
		}(&scorer{cache: map[[16]byte]float64{}, maxColor: turn}, m)
	}
	bestScore := -10000.0
	for i := 0; i < len(moves); i++ {
		moveScr := <-ch
		if moveScr.score > bestScore {
			bestScore = moveScr.score
			move = moveScr.move
		}
	}
	return move, bestScore
}

func alphaBetaMax(scr *scorer, pos *chess.Position, alpha, beta float64, depthleft int) float64 {
	hash := pos.Hash()
	score, ok := scr.cache[hash]
	if ok {
		return score
	}
	if depthleft == 0 {
		score = scr.score(pos)
		scr.cache[hash] = score
		return score
	}
	moves := pos.ValidMoves()
	sort.Sort(byMoveImportance(moves))
	for _, m := range moves {
		newPos := pos.Update(m)
		score = alphaBetaMin(scr, newPos, alpha, beta, depthleft-1)
		if score >= beta {
			scr.cache[hash] = beta
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}
	scr.cache[hash] = alpha
	return alpha
}

func alphaBetaMin(scr *scorer, pos *chess.Position, alpha, beta float64, depthleft int) float64 {
	hash := pos.Hash()
	score, ok := scr.cache[hash]
	if ok {
		return score
	}
	if depthleft == 0 {
		score = scr.score(pos)
		scr.cache[hash] = score
		return score
	}
	for _, m := range pos.ValidMoves() {
		newPos := pos.Update(m)
		score = alphaBetaMax(scr, newPos, alpha, beta, depthleft-1)
		if score <= alpha {
			scr.cache[hash] = alpha
			return alpha
		}
		if score < beta {
			beta = score
		}
	}
	scr.cache[hash] = beta
	return beta
}

type scorer struct {
	cache    map[[16]byte]float64
	maxColor chess.Color
}

func (s scorer) score(pos *chess.Position) float64 {
	// check for stalemate and checkmate
	turn := pos.Turn()
	status := pos.Status()
	if status == chess.Stalemate {
		return 0.0
	} else if status == chess.Checkmate {
		if turn == s.maxColor {
			return 1000.0
		}
		return -1000.0
	}
	// compare material
	total := 0.0
	for sq := 0; sq < 64; sq++ {
		p := pos.Board().Piece(chess.Square(sq))
		score := pieceScore(pos, p)
		if p.Color() == s.maxColor {
			total += score
		} else {
			total -= score
		}
	}
	// moveCount := len(pos.ValidMoves())
	// oppenentMoveCount := len(pos.Update(&chess.Move{}).ValidMoves())
	// total += float64(moveCount-oppenentMoveCount) * 0.1
	return total
}

func pieceScore(pos *chess.Position, piece chess.Piece) float64 {
	switch piece.Type() {
	case chess.King:
		return 200.0
	case chess.Queen:
		return 9.0
	case chess.Rook:
		return 5.0
	case chess.Bishop:
		return 3.1
	case chess.Knight:
		return 3.0
	case chess.Pawn:
		return 1.0
	}
	return 0.0
}

func openingMove(game *chess.Game) *chess.Move {
	prevMoves := game.Moves()
	moveIndex := len(prevMoves)
	opennings := opening.Possible(game.Moves())
	sort.Sort(byOpeningLength(opennings))
	for _, op := range opennings {
		moves := op.Game().Moves()
		if len(moves) > moveIndex {
			return moves[moveIndex]
		}
		break
	}
	return nil
}

type byOpeningLength []*opening.Opening

func (a byOpeningLength) Len() int           { return len(a) }
func (a byOpeningLength) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byOpeningLength) Less(i, j int) bool { return len(a[i].PGN()) > len(a[j].PGN()) }

type byMoveImportance []*chess.Move

func (a byMoveImportance) Len() int      { return len(a) }
func (a byMoveImportance) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byMoveImportance) Less(i, j int) bool {
	if a[i].HasTag(chess.Check) && !a[j].HasTag(chess.Check) {
		return true
	}
	if a[i].HasTag(chess.Capture) && !a[j].HasTag(chess.Capture) {
		return true
	}
	return false
}
