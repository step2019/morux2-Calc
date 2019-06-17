// To run these tests run: go test *.go
package main

import (
	"fmt"
	"log"
	"math"
	"runtime/debug"
	"testing"
)

// Floating point values have limited precision when adding/dividing,
// so 1.0 + 1.0 might not always be exactly 2.0, depending on the
// value of 1.0 (but it should be pretty close).
const tolerance float64 = 0.000000001

func TestCalculate(t *testing.T) {
	for _, test := range []struct {
		in   string
		want float64
	}{
		//数値の入力(整数)
		{"0", 0},

		//数値の入力(小数)
		{"1.11111111111", 1.11111111111},

		//演算子1つ(整数)
		{"1+2", 3},
		{"2-1", 1},
		{"2*4", 8},
		{"4/2", 2},

		//演算子1つ(小数)
		{"1.1+2.2", 3.3},
		{"2.2-1.1", 1.1},
		{"1.1*2.2", 2.42},
		{"2.2/1.1", 2.0},

		//演算子1つ(整数, 小数)
		{"2+2.2", 4.2},
		{"2.2-2", 0.2},
		{"2*2.2", 4.4},
		{"2.2/2", 1.1},

		//答えが負
		{"1-2.2", -1.2},

		//入力に負の整数
		{"1+-2", -1},

		//入力に負の小数
		//-1をかける実装をした時にこの誤差が許容範囲を超えてしまう
		//Calculate(1+-2.2) = -1.2000000000000002 but want 1.2
		{"1+-2.2", -1.2},

		//連続計算
		{"2+3+4", 9},
		{"4-3-2", -1},
		{"2*3*4", 24},
		{"4/2/1", 2},

		//四則演算
		{"1+2*3+4/2-5", 4},

		//(整数)
		{"(1)", 1},

		//(小数)
		{"(1.1)", 1.1},

		//((整数))
		{"((2))", 2},

		//((小数))
		{"((2.1))", 2.1},

		//(整数+整数)
		{"(1+2)", 3},

		//(小数+小数)
		{"(1.1+2.1)", 3.2},

		//(整数+小数)
		{"(1+2.1)", 3.1},

		//(整数*整数)
		{"(2*3)", 6},

		//(小数*小数)
		{"(2.1*3.0)", 6.3},

		//(整数*小数)
		{"(2*3.1)", 6.2},

		//((式))
		{"((2+3))", 5},

		//((式)+整数)
		{"((2+3)+4)", 9},

		//((式)+小数)
		{"((2+3)+4.1)", 9.1},

		//((式)*整数)
		{"((2+3)*4)", 20},

		//((式)*小数)
		{"((2+3)+4.1)", 9.1},

		//(式)+(式)
		{"(1+2)+(3+4)", 10},

		//(式)*(式)
		{"(1+2)*(3+4)", 21},
	} {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Calculate(%v) panicked(%v) but wanted %v", test.in, r, test.want)
				t.Errorf("stacktrace: %s", debug.Stack())
			}
		}()
		got := Calculate(test.in)
		// floatだと完全には一致しないのでとっても近いかどうかを判定している。
		if math.Abs(got-test.want) > tolerance {
			t.Errorf("Calculate(%v) = %v but want %v", test.in, got, test.want)
		}
	}
}

func TestCalculatePanics(t *testing.T) {
	for _, test := range []struct {
		in   string
		want string
	}{
		//0で割った時に正しいエラー文が出るかどうかのチェックを追加したい
		{"0/0", "can't divide by 0"},
		{"1/0", "this doesn't break but it should."},
	} {
		// When panic is called the whole calling function is
		// terminated. So unless we're calling Calculate from within
		// another function (missing in this case!), we would (and
		// do!) stop the for loop at the first panic (so we won't
		// actually catch a test case if it were wrong).

		// NOTE: This example is wrong. In this case we've removed the
		// surrounding function, so when the first test case panics,
		// the whole test case terminates without reporting a failure.
		defer func() {
			panicked := fmt.Sprint(recover())
			if panicked != test.want {
				t.Errorf("Calculate(%v) had panicked = `%v` but wanted panic: %v", test.in, panicked, test.want)
			}
		}()
		log.Printf("run %v", test.in)
		Calculate(test.in)
		// The defered function above executes anyway and will
		// report an error unless panicked matches the expected
		// value.
	}
}
