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
	HEAD := tokenize(line)

	//トークンを表示する
	//printToken(HEAD)

	// * / を計算してトークンを組み替える
	readTokens(HEAD)

	//トークンを表示する
	//printToken(HEAD)

	//計算結果を返す
	return evaluate(HEAD)
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
)

// Tokenize lexes a given line, breaking it down into its component
// tokens.
func tokenize(line string) *token {
	// Start with a dummy '+' token
	HEAD := token{Plus, 0, nil, nil}
	prev := &HEAD
	index := 0
	for index < len(line) {
		var tok *token
		switch {
		case unicode.IsDigit(rune(line[index])):
			tok, index = readNumber(line, index)
		case line[index] == '+':
			tok, index = readPlus(line, index)
		case line[index] == '-':
			tok, index = readMinus(line, index)
		case line[index] == '*':
			tok, index = readMultiple(line, index)
		case line[index] == '/':
			tok, index = readDivide(line, index)
		default:
			//panicとはプログラムの継続的な実行が難しく、どうしよもなくなった時にプログラムを強制的に終了させるために発生するエラーです。
			log.Panicf("invalid character: '%c' at index=%v in %v", line[index], index, line)
		}
		prev = connectToken(prev, tok)
	}
	return &HEAD
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

func readTokens(HEAD *token) *token {
	p := HEAD
	for {
		switch p.kind {
		case Multiple:
			p = replaceToken(p, calcMultiple(p))
		case Divide:
			p = replaceToken(p, calcDivide(p))
		default:
			p = p.next
		}

		if p == nil {
			break
		}
	}
	return HEAD
}

func calcMultiple(p *token) *token {
	return &token{Number, p.prev.number * p.next.number, nil, nil}
}

func calcDivide(p *token) *token {
	return &token{Number, p.prev.number / p.next.number, nil, nil}
}

func replaceToken(p *token, new *token) *token {
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

// Evaluate computes the numeric value expressed by a series of
// tokens.
func evaluate(p *token) float64 {
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
