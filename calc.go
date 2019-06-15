// A Go transliteration of
// https://github.com/xharaken/step2015/blob/master/calculator_modularize_2.py
// This is intended to keep the same shape and style of the
// @xharaken's original Python code, but please keep in mind that
// there are much more Go-friendly ways of writing much of this code.
// Consider watching Rob Pike's <r@google.com> talk on using Go to
// write a lexer (lexical scanner) here:
// https://talks.golang.org/2011/lex.slide recorded here
// https://www.youtube.com/watch?v=HxaD_trXwRE
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"unicode"
)

func main() {
	input := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !input.Scan() || input.Text() == "" { // Reads a line from standard input.
			return // If it's empty, exit the program.
		}
		answer := Calculate(input.Text())
		fmt.Println("answer =", answer)
	}
}

// Calculate turns a string like "1 + 3" into its corresponding
// numerical value (in this case 4).
//メソッド名 引数 戻り値の型
func Calculate(line string) float64 {
	//数式文字列をトークンに分解する
	HEAD, TAIL := tokenize(line)

	//トークンを表示する
	//printToken(HEAD)

	// () を計算してトークンを組み替える
	evaluateStartEnd(TAIL)

	//トークンを表示する
	//printToken(HEAD)

	// * / を計算してトークンを組み替える
	evaluateMulDiv(HEAD)

	//トークンを表示する
	//printToken(HEAD)

	//計算結果を返す
	return evaluatePlusMinus(HEAD)
}

type token struct {
	// Specifies the type of the token. I'm using the word "kind" here
	// rather than "type" because type is a reserved word in Go.
	kind tokenKind

	// If kind is Number, then number is its corresponding numeric
	// value.
	number float64
	prev   *token
	next   *token
}

// TokenKind describes a valid kinds of tokens. This acts kind of
// like an enum in C/C++.
type tokenKind int

// These are the valid kinds of tokens. Each gets automatically
// initialized with a unique value by setting the first one to iota
// like this. https://golang.org/ref/spec#Iota
//0123..と番号が振られていく(https://qiita.com/curepine/items/2ae2f6504f0d28016411)
const (
	Number tokenKind = iota
	Plus
	Minus
	Multiple
	Divide
	Start
	End
)

// Tokenize lexes a given line, breaking it down into its component
// tokens.
//HEAD と TAIL を返す
func tokenize(line string) (*token, *token) {
	// Start with a dummy '+' token
	HEAD := token{Plus, 0, nil, nil}
	prev := &HEAD
	index := 0
	flag := false
	for index < len(line) {
		var tok *token
		switch {
		case unicode.IsDigit(rune(line[index])):
			tok, index = readNumber(line, index)
			if flag {
				tok.number *= -1
				flag = false
			}
		case line[index] == '+':
			tok, index = readPlus(line, index)
		case line[index] == '-':
			if prev.kind != Number {
				flag = true
				index++
				continue
			} else {
				tok, index = readMinus(line, index)
			}
		case line[index] == '*':
			tok, index = readMultiple(line, index)
		case line[index] == '/':
			tok, index = readDivide(line, index)
		case line[index] == '(':
			tok, index = readStart(line, index)
		case line[index] == ')':
			tok, index = readEnd(line, index)
		default:
			//panicとはプログラムの継続的な実行が難しく、どうしよもなくなった時にプログラムを強制的に終了させるために発生するエラーです。
			log.Panicf("invalid character: '%c' at index=%v in %v", line[index], index, line)
		}
		prev = connectToken(prev, tok)
	}
	// means return HEAD and TAIL
	return &HEAD, prev
}

func connectToken(prev *token, tok *token) *token {
	prev.next = tok
	tok.prev = prev
	return tok
}

func printToken(p *token) {
	fmt.Printf("\n")
	for {
		fmt.Printf("%d %f\n", p.kind, p.number)
		p = p.next
		if p == nil {
			break
		}
	}
	fmt.Printf("\n")
}

func evaluateStartEnd(TAIL *token) {
	p := TAIL
	for {
		switch p.kind {
		case Start:
			//(の次の数字
			tmpHead := p.next
			//()のペアを見つける
			for p.next.kind != End {
				p = p.next
			}
			//)の前の数字
			tmpEnd := p
			p = calcStartEnd(tmpHead, tmpEnd)
		default:
			p = p.prev
		}

		if p == nil {
			break
		}
	}
}

func calcStartEnd(tmpHead *token, tmpEnd *token) *token {
	new := &token{Number, 0, nil, nil}
	replaceStartEnd(tmpHead, new, tmpEnd)
	//()の中の式を前後から切り離す
	// Start with a dummy '+' token
	dummy := &token{Plus, 0, nil, tmpHead}
	tmpHead.prev = dummy
	tmpEnd.next = nil
	evaluateMulDiv(dummy)
	new.number = evaluatePlusMinus(dummy)
	return new.prev
}

func replaceStartEnd(tmpHead *token, new *token, tmpEnd *token) {
	//tmpHead.prev.prevは必ずnilにならない(dummyが入ってるから)
	tmpHead.prev.prev.next = new
	new.prev = tmpHead.prev.prev

	if tmpEnd.next.next != nil {
		tmpEnd.next.next.prev = new
		new.next = tmpEnd.next.next
	}
}

func evaluateMulDiv(HEAD *token) {
	p := HEAD
	for {
		switch p.kind {
		case Multiple:
			p = replaceMulDiv(p, calcMultiple(p))
		case Divide:
			p = replaceMulDiv(p, calcDivide(p))
		default:
			p = p.next
		}

		if p == nil {
			break
		}
	}
}

func calcMultiple(p *token) *token {
	return &token{Number, p.prev.number * p.next.number, nil, nil}
}

func calcDivide(p *token) *token {
	return &token{Number, p.prev.number / p.next.number, nil, nil}
}

func replaceMulDiv(p *token, new *token) *token {
	if p.prev.prev != nil {
		p.prev.prev.next = new
		new.prev = p.prev.prev
	}
	if p.next.next != nil {
		p.next.next.prev = new
		new.next = p.next.next
	}
	return new.next
}

// EvaluatePlusMinus computes the numeric value expressed by a series of
// tokens.
func evaluatePlusMinus(p *token) float64 {
	answer := float64(0)
	for {
		switch p.kind {
		case Number:
			switch p.prev.kind {
			case Plus:
				answer += p.number
			case Minus:
				answer -= p.number
			default:
				log.Panicf("invalid syntax for token")
			}
		}
		p = p.next
		if p == nil {
			break
		}
	}
	return answer
}

func readPlus(line string, index int) (*token, int) {
	return &token{Plus, 0, nil, nil}, index + 1
}

func readMinus(line string, index int) (*token, int) {
	return &token{Minus, 0, nil, nil}, index + 1
}

func readMultiple(line string, index int) (*token, int) {
	return &token{Multiple, 0, nil, nil}, index + 1
}

func readDivide(line string, index int) (*token, int) {
	return &token{Divide, 0, nil, nil}, index + 1
}

func readStart(line string, index int) (*token, int) {
	return &token{Start, 0, nil, nil}, index + 1
}

func readEnd(line string, index int) (*token, int) {
	return &token{End, 0, nil, nil}, index + 1
}

func readNumber(line string, index int) (*token, int) {
	number := float64(0)
	flag := false
	keta := float64(1)
DigitLoop:
	for index < len(line) {
		switch {
		case line[index] == '.':
			flag = true
		case unicode.IsDigit(rune(line[index])):
			//'0'をひいて文字を数値に変換
			number = number*10 + float64(line[index]-'0')
			if flag {
				keta *= 0.1
			}
		default:
			// "break DigitLoop" here means to break from the labeled for loop, rather than the switch statement. https://golang.org/ref/spec#Break_statements
			break DigitLoop
		}
		index += 1
	}
	//数値の時はたくさんindexを進める
	return &token{Number, number * keta, nil, nil}, index
}
