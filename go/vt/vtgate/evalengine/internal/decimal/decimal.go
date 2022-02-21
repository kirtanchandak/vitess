/*
Copyright 2022 The Vitess Authors.
Copyright (c) 2015 Spring, Inc.
Copyright (c) 2013 Oguz Bilgic

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Based on MIT licensed code from https://github.com/shopspring/decimal
// See the LICENSE file in this directory

package decimal

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"vitess.io/vitess/go/hack"
)

// divisionPrecision is the number of decimal places in the result when it
// doesn't divide exactly.
//
// Example:
//
//     d1 := decimal.NewFromFloat(2).div(decimal.NewFromFloat(3))
//     d1.String() // output: "0.6666666666666667"
//     d2 := decimal.NewFromFloat(2).div(decimal.NewFromFloat(30000))
//     d2.String() // output: "0.0000666666666667"
//     d3 := decimal.NewFromFloat(20000).div(decimal.NewFromFloat(3))
//     d3.String() // output: "6666.6666666666666667"
//     decimal.divisionPrecision = 3
//     d4 := decimal.NewFromFloat(2).div(decimal.NewFromFloat(3))
//     d4.String() // output: "0.667"
//
var divisionPrecision = 16

// Zero constant, to make computations faster.
// Zero should never be compared with == or != directly, please use decimal.Equal or decimal.Cmp instead.
var Zero = New(0, 1)

var zeroInt = big.NewInt(0)
var oneInt = big.NewInt(1)
var fiveInt = big.NewInt(5)
var tenInt = big.NewInt(10)

const powTabLen = 20

var pow10bigtab = [powTabLen]*big.Int{
	0:  new(big.Int).SetUint64(1),
	1:  tenInt,
	2:  new(big.Int).SetUint64(100),
	3:  new(big.Int).SetUint64(1000),
	4:  new(big.Int).SetUint64(10000),
	5:  new(big.Int).SetUint64(100000),
	6:  new(big.Int).SetUint64(1000000),
	7:  new(big.Int).SetUint64(10000000),
	8:  new(big.Int).SetUint64(100000000),
	9:  new(big.Int).SetUint64(1000000000),
	10: new(big.Int).SetUint64(10000000000),
	11: new(big.Int).SetUint64(100000000000),
	12: new(big.Int).SetUint64(1000000000000),
	13: new(big.Int).SetUint64(10000000000000),
	14: new(big.Int).SetUint64(100000000000000),
	15: new(big.Int).SetUint64(1000000000000000),
	16: new(big.Int).SetUint64(10000000000000000),
	17: new(big.Int).SetUint64(100000000000000000),
	18: new(big.Int).SetUint64(1000000000000000000),
	19: new(big.Int).SetUint64(10000000000000000000),
}

var limitsBigTab = [powTabLen]*big.Int{
	0:  zeroInt,
	1:  new(big.Int).SetUint64(9),
	2:  new(big.Int).SetUint64(99),
	3:  new(big.Int).SetUint64(999),
	4:  new(big.Int).SetUint64(9999),
	5:  new(big.Int).SetUint64(99999),
	6:  new(big.Int).SetUint64(999999),
	7:  new(big.Int).SetUint64(9999999),
	8:  new(big.Int).SetUint64(99999999),
	9:  new(big.Int).SetUint64(999999999),
	10: new(big.Int).SetUint64(9999999999),
	11: new(big.Int).SetUint64(99999999999),
	12: new(big.Int).SetUint64(999999999999),
	13: new(big.Int).SetUint64(9999999999999),
	14: new(big.Int).SetUint64(99999999999999),
	15: new(big.Int).SetUint64(999999999999999),
	16: new(big.Int).SetUint64(9999999999999999),
	17: new(big.Int).SetUint64(99999999999999999),
	18: new(big.Int).SetUint64(999999999999999999),
	19: new(big.Int).SetUint64(9999999999999999999),
}

// Decimal represents a fixed-point decimal. It is immutable.
// number = value * 10 ^ exp
type Decimal struct {
	value *big.Int

	// NOTE(vadim): this must be an int32, because we cast it to float64 during
	// calculations. If exp is 64 bit, we might lose precision.
	// If we cared about being able to represent every possible decimal, we
	// could make exp a *big.Int but it would hurt performance and numbers
	// like that are unrealistic.
	exp int32
}

// New returns a new fixed-point decimal, value * 10 ^ exp.
func New(value int64, exp int32) Decimal {
	return Decimal{
		value: big.NewInt(value),
		exp:   exp,
	}
}

// NewFromInt converts a int64 to Decimal.
//
// Example:
//
//     NewFromInt(123).String() // output: "123"
//     NewFromInt(-10).String() // output: "-10"
func NewFromInt(value int64) Decimal {
	return Decimal{
		value: big.NewInt(value),
		exp:   0,
	}
}

func NewFromUint(value uint64) Decimal {
	return Decimal{
		value: new(big.Int).SetUint64(value),
		exp:   0,
	}
}

var errOverflow = errors.New("overflow")

func parseDecimal64(s []byte) (Decimal, error) {
	const cutoff = math.MaxUint64/10 + 1
	var n uint64
	var dot = -1

	for i, c := range s {
		var d byte
		switch {
		case c == '.':
			if dot > -1 {
				return Decimal{}, fmt.Errorf("too many .s")
			}
			dot = i
			continue
		case '0' <= c && c <= '9':
			d = c - '0'
		default:
			return Decimal{}, fmt.Errorf("unexpected character %q", c)
		}

		if n >= cutoff {
			// n*base overflows
			return Decimal{}, errOverflow
		}
		n *= 10
		n1 := n + uint64(d)
		if n1 < n {
			return Decimal{}, errOverflow
		}
		n = n1
	}

	var exp int32
	if dot != -1 {
		exp = -int32(len(s) - dot - 1)
	}
	return Decimal{
		value: new(big.Int).SetUint64(n),
		exp:   exp,
	}, nil
}

func NewFromMySQL(s []byte) (Decimal, error) {
	var original = s
	var neg bool

	if len(s) > 0 {
		switch s[0] {
		case '+':
			s = s[1:]
		case '-':
			neg = true
			s = s[1:]
		}
	}

	if len(s) == 0 {
		return Decimal{}, fmt.Errorf("can't convert %q to decimal: too short", original)
	}

	if len(s) <= 18 {
		dec, err := parseDecimal64(s)
		if err == nil {
			if neg {
				dec.value.Neg(dec.value)
			}
			return dec, nil
		}
		if err != errOverflow {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: %v", original, err)
		}
	}

	var intString string
	var exp int32
	if pIndex := bytes.IndexByte(s, '.'); pIndex >= 0 {
		if bytes.IndexByte(s[pIndex+1:], '.') != -1 {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: too many .s", original)
		}
		if pIndex+1 < len(s) {
			var buf strings.Builder
			buf.Grow(len(s))
			buf.Write(s[:pIndex])
			buf.Write(s[pIndex+1:])
			intString = buf.String()
		} else {
			intString = hack.String(s[:pIndex])
		}
		exp = -int32(len(s[pIndex+1:]))
	} else {
		intString = hack.String(s)
	}

	dValue := new(big.Int)
	_, ok := dValue.SetString(intString, 10)
	if !ok {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal", original)
	}
	if neg {
		dValue.Neg(dValue)
	}
	return Decimal{value: dValue, exp: exp}, nil
}

// NewFromString returns a new Decimal from a string representation.
// Trailing zeroes are not trimmed.
//
// Example:
//
//     d, err := NewFromString("-123.45")
//     d2, err := NewFromString(".0001")
//     d3, err := NewFromString("1.47000")
//
func NewFromString(value string) (Decimal, error) {
	originalInput := value
	var intString string
	var exp int64

	// Check if number is using scientific notation
	eIndex := strings.IndexAny(value, "Ee")
	if eIndex != -1 {
		expInt, err := strconv.ParseInt(value[eIndex+1:], 10, 32)
		if err != nil {
			if e, ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
				return Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", value)
			}
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: exponent is not numeric", value)
		}
		value = value[:eIndex]
		exp = expInt
	}

	pIndex := -1
	vLen := len(value)
	for i := 0; i < vLen; i++ {
		if value[i] == '.' {
			if pIndex > -1 {
				return Decimal{}, fmt.Errorf("can't convert %s to decimal: too many .s", value)
			}
			pIndex = i
		}
	}

	if pIndex == -1 {
		// There is no decimal point, we can just parse the original string as
		// an int
		intString = value
	} else {
		if pIndex+1 < vLen {
			intString = value[:pIndex] + value[pIndex+1:]
		} else {
			intString = value[:pIndex]
		}
		expInt := -len(value[pIndex+1:])
		exp += int64(expInt)
	}

	var dValue *big.Int
	// strconv.ParseInt is faster than new(big.Int).SetString so this is just a shortcut for strings we know won't overflow
	if len(intString) <= 18 {
		parsed64, err := strconv.ParseInt(intString, 10, 64)
		if err != nil {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal", value)
		}
		dValue = big.NewInt(parsed64)
	} else {
		dValue = new(big.Int)
		_, ok := dValue.SetString(intString, 10)
		if !ok {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal", value)
		}
	}

	if exp < math.MinInt32 || exp > math.MaxInt32 {
		// NOTE(vadim): I doubt a string could realistically be this long
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", originalInput)
	}

	return Decimal{
		value: dValue,
		exp:   int32(exp),
	}, nil
}

// RequireFromString returns a new Decimal from a string representation
// or panics if NewFromString would have returned an error.
//
// Example:
//
//     d := RequireFromString("-123.45")
//     d2 := RequireFromString(".0001")
//
func RequireFromString(value string) Decimal {
	dec, err := NewFromString(value)
	if err != nil {
		panic(err)
	}
	return dec
}

// NewFromFloat converts a float64 to Decimal.
//
// The converted number will contain the number of significant digits that can be
// represented in a float with reliable roundtrip.
// This is typically 15 digits, but may be more in some cases.
// See https://www.exploringbinary.com/decimal-precision-of-binary-floating-point-numbers/ for more information.
//
// For slightly faster conversion, use NewFromFloatWithExponent where you can specify the precision in absolute terms.
//
// NOTE: this will panic on NaN, +/-inf
func NewFromFloat(value float64) Decimal {
	if value == 0 {
		return New(0, 0)
	}
	dec, err := NewFromMySQL(strconv.AppendFloat(nil, value, 'f', -1, 64))
	if err != nil {
		panic(err)
	}
	return dec
}

func NewFromFloat32(value float32) Decimal {
	if value == 0 {
		return New(0, 0)
	}
	dec, err := NewFromMySQL(strconv.AppendFloat(nil, float64(value), 'f', -1, 32))
	if err != nil {
		panic(err)
	}
	return dec
}

// Copy returns a copy of decimal with the same value and exponent, but a different pointer to value.
func (d Decimal) Copy() Decimal {
	d.ensureInitialized()
	return Decimal{
		value: new(big.Int).Set(d.value),
		exp:   d.exp,
	}
}

func bigPow10(n uint64) *big.Int {
	if n < powTabLen {
		return pow10bigtab[n]
	}

	// Too large for our table.
	// As an optimization, we don't need to start from
	// scratch each time. Start from the largest term we've
	// found so far.
	partial := pow10bigtab[powTabLen-1]
	p := new(big.Int).SetUint64(n - (powTabLen - 1))
	return p.Mul(partial, p.Exp(tenInt, p, nil))
}

// rescale returns a rescaled version of the decimal. Returned
// decimal may be less precise if the given exponent is bigger
// than the initial exponent of the Decimal.
// NOTE: this will truncate, NOT round
//
// Example:
//
// 	d := New(12345, -4)
//	d2 := d.rescale(-1)
//	d3 := d2.rescale(-4)
//	println(d1)
//	println(d2)
//	println(d3)
//
// Output:
//
//	1.2345
//	1.2
//	1.2000
//
func (d Decimal) rescale(exp int32) Decimal {
	d.ensureInitialized()

	value := new(big.Int).Set(d.value)
	if exp > d.exp {
		scale := bigPow10(uint64(exp - d.exp))
		value = value.Quo(value, scale)
	} else if exp < d.exp {
		scale := bigPow10(uint64(d.exp - exp))
		value = value.Mul(value, scale)
	}

	return Decimal{value: value, exp: exp}
}

// abs returns the absolute value of the decimal.
func (d Decimal) abs() Decimal {
	if d.Sign() >= 0 {
		return d
	}
	d.ensureInitialized()
	d2Value := new(big.Int).Abs(d.value)
	return Decimal{
		value: d2Value,
		exp:   d.exp,
	}
}

// Add returns d + d2.
func (d Decimal) Add(d2 Decimal) Decimal {
	rd, rd2 := RescalePair(d, d2)

	d3Value := new(big.Int).Add(rd.value, rd2.value)
	return Decimal{
		value: d3Value,
		exp:   rd.exp,
	}
}

// sub returns d - d2.
func (d Decimal) sub(d2 Decimal) Decimal {
	rd, rd2 := RescalePair(d, d2)
	d3Value := new(big.Int).Sub(rd.value, rd2.value)
	return Decimal{
		value: d3Value,
		exp:   rd.exp,
	}
}

func (d Decimal) Sub(d2 Decimal) Decimal {
	rd := d.sub(d2)
	if rd.value.Sign() == 0 {
		rd.exp = 0
	}
	return rd
}

// Neg returns -d.
func (d Decimal) Neg() Decimal {
	d.ensureInitialized()
	val := new(big.Int).Neg(d.value)
	return Decimal{
		value: val,
		exp:   d.exp,
	}
}

func (d Decimal) NegInPlace() Decimal {
	d.ensureInitialized()
	return Decimal{
		value: d.value.Neg(d.value),
		exp:   d.exp,
	}
}

// mul returns d * d2.
func (d Decimal) mul(d2 Decimal) Decimal {
	d.ensureInitialized()
	d2.ensureInitialized()

	expInt64 := int64(d.exp) + int64(d2.exp)
	if expInt64 > math.MaxInt32 || expInt64 < math.MinInt32 {
		// NOTE(vadim): better to panic than give incorrect results, as
		// Decimals are usually used for money
		panic(fmt.Sprintf("exponent %v overflows an int32!", expInt64))
	}

	d3Value := new(big.Int).Mul(d.value, d2.value)
	return Decimal{
		value: d3Value,
		exp:   int32(expInt64),
	}
}

func (d Decimal) Mul(d2 Decimal) Decimal {
	if d.Sign() == 0 || d2.Sign() == 0 {
		return Zero
	}
	return d.mul(d2)
}

func roundBase10(digits int32) int32 {
	return (digits + 8) / 9
}

func (d Decimal) Div(d2 Decimal, scaleIncr int32) Decimal {
	if d.Sign() == 0 {
		return Zero
	}

	s1 := -d.exp
	s2 := -d2.exp
	fracLeft := roundBase10(s1)
	fracRight := roundBase10(s2)
	scaleIncr -= fracLeft - s1 + fracRight - s2
	if scaleIncr < 0 {
		scaleIncr = 0
	}
	scale := roundBase10(fracLeft+fracRight+scaleIncr) * 9
	q, _ := d.quoRem(d2, scale)
	return q
}

// div returns d / d2. If it doesn't divide exactly, the result will have
// divisionPrecision digits after the decimal point.
func (d Decimal) div(d2 Decimal) Decimal {
	return d.divRound(d2, int32(divisionPrecision))
}

// quoRem does division with remainder
// d.quoRem(d2,precision) returns quotient q and remainder r such that
//   d = d2 * q + r, q an integer multiple of 10^(-precision)
//   0 <= r < abs(d2) * 10 ^(-precision) if d>=0
//   0 >= r > -abs(d2) * 10 ^(-precision) if d<0
// Note that precision<0 is allowed as input.
func (d Decimal) quoRem(d2 Decimal, precision int32) (Decimal, Decimal) {
	d.ensureInitialized()
	d2.ensureInitialized()
	if d2.value.Sign() == 0 {
		panic("decimal division by 0")
	}
	scale := -precision
	e := int64(d.exp - d2.exp - scale)
	if e > math.MaxInt32 || e < math.MinInt32 {
		panic("overflow in decimal quoRem")
	}
	var aa, bb, expo big.Int
	var scalerest int32
	// d = a 10^ea
	// d2 = b 10^eb
	if e < 0 {
		aa = *d.value
		expo.SetInt64(-e)
		bb.Exp(tenInt, &expo, nil)
		bb.Mul(d2.value, &bb)
		scalerest = d.exp
		// now aa = a
		//     bb = b 10^(scale + eb - ea)
	} else {
		expo.SetInt64(e)
		aa.Exp(tenInt, &expo, nil)
		aa.Mul(d.value, &aa)
		bb = *d2.value
		scalerest = scale + d2.exp
		// now aa = a ^ (ea - eb - scale)
		//     bb = b
	}
	var q, r big.Int
	q.QuoRem(&aa, &bb, &r)
	dq := Decimal{value: &q, exp: scale}
	dr := Decimal{value: &r, exp: scalerest}
	return dq, dr
}

// divRound divides and rounds to a given precision
// i.e. to an integer multiple of 10^(-precision)
//   for a positive quotient digit 5 is rounded up, away from 0
//   if the quotient is negative then digit 5 is rounded down, away from 0
// Note that precision<0 is allowed as input.
func (d Decimal) divRound(d2 Decimal, precision int32) Decimal {
	// quoRem already checks initialization
	q, r := d.quoRem(d2, precision)

	// the actual rounding decision is based on comparing r*10^precision and d2/2
	// instead compare 2 r 10 ^precision and d2
	var rv2 big.Int
	rv2.Abs(r.value)
	rv2.Lsh(&rv2, 1)
	// now rv2 = abs(r.value) * 2
	r2 := Decimal{value: &rv2, exp: r.exp + precision}
	// r2 is now 2 * r * 10 ^ precision
	var c = r2.Cmp(d2.abs())

	if c < 0 {
		return q
	}

	if d.value.Sign()*d2.value.Sign() < 0 {
		return q.sub(New(1, -precision))
	}

	return q.Add(New(1, -precision))
}

// mod returns d % d2.
func (d Decimal) mod(d2 Decimal) Decimal {
	quo := d.divRound(d2, -d.exp+1).truncate(0)
	return d.sub(d2.mul(quo))
}

func (d Decimal) truncate(precision int32) Decimal {
	d.ensureInitialized()
	if precision >= 0 && -precision > d.exp {
		return d.rescale(-precision)
	}
	return d
}

// isInteger returns true when decimal can be represented as an integer value, otherwise, it returns false.
func (d Decimal) isInteger() bool {
	// The most typical case, all decimal with exponent higher or equal 0 can be represented as integer
	if d.exp >= 0 {
		return true
	}
	// When the exponent is negative we have to check every number after the decimal place
	// If all of them are zeroes, we are sure that given decimal can be represented as an integer
	var r big.Int
	q := new(big.Int).Set(d.value)
	for z := abs(d.exp); z > 0; z-- {
		q.QuoRem(q, tenInt, &r)
		if r.Cmp(zeroInt) != 0 {
			return false
		}
	}
	return true
}

// abs calculates absolute value of any int32. Used for calculating absolute value of decimal's exponent.
func abs(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}

// Cmp compares the numbers represented by d and d2 and returns:
//
//     -1 if d <  d2
//      0 if d == d2
//     +1 if d >  d2
//
func (d Decimal) Cmp(d2 Decimal) int {
	d.ensureInitialized()
	d2.ensureInitialized()
	if d.exp == d2.exp {
		return d.value.Cmp(d2.value)
	}
	rd, rd2 := RescalePair(d, d2)
	return rd.value.Cmp(rd2.value)
}

func (d Decimal) CmpAbs(d2 Decimal) int {
	d.ensureInitialized()
	d2.ensureInitialized()
	if d.exp == d2.exp {
		return d.value.CmpAbs(d2.value)
	}
	rd, rd2 := RescalePair(d, d2)
	return rd.value.CmpAbs(rd2.value)
}

// Equal returns whether the numbers represented by d and d2 are equal.
func (d Decimal) Equal(d2 Decimal) bool {
	return d.Cmp(d2) == 0
}

// Sign returns:
//
//	-1 if d <  0
//	 0 if d == 0
//	+1 if d >  0
//
func (d Decimal) Sign() int {
	if d.value == nil {
		return 0
	}
	return d.value.Sign()
}

// IsZero return
//
//	true if d == 0
//	false if d > 0
//	false if d < 0
func (d Decimal) IsZero() bool {
	return d.Sign() == 0
}

// Exponent returns the exponent, or scale component of the decimal.
func (d Decimal) Exponent() int32 {
	return d.exp
}

func (d Decimal) Int64() (int64, bool) {
	scaledD := d.rescale(0)
	return scaledD.value.Int64(), scaledD.value.IsInt64()
}

func (d Decimal) Uint64() (uint64, bool) {
	scaledD := d.rescale(0)
	return scaledD.value.Uint64(), scaledD.value.IsUint64()
}

// Float64 returns the nearest float64 value for d and a bool indicating
// whether f represents d exactly.
func (d Decimal) Float64() (f float64, ok bool) {
	f, _ = strconv.ParseFloat(d.String(), 64)
	ok = !math.IsInf(f, 0)
	return
}

// String returns the string representation of the decimal
// with the fixed point.
//
// Example:
//
//     d := New(-12345, -3)
//     println(d.String())
//
// Output:
//
//     -12.345
//
func (d Decimal) String() string {
	return string(d.format(make([]byte, 0, 10), true))
}

// StringFixed returns a rounded fixed-point string with places digits after
// the decimal point.
//
// Example:
//
// 	   NewFromFloat(0).StringFixed(2) // output: "0.00"
// 	   NewFromFloat(0).StringFixed(0) // output: "0"
// 	   NewFromFloat(5.45).StringFixed(0) // output: "5"
// 	   NewFromFloat(5.45).StringFixed(1) // output: "5.5"
// 	   NewFromFloat(5.45).StringFixed(2) // output: "5.45"
// 	   NewFromFloat(5.45).StringFixed(3) // output: "5.450"
// 	   NewFromFloat(545).StringFixed(-1) // output: "550"
//
func (d Decimal) StringFixed(places int32) string {
	rounded := d.Round(places)
	return string(rounded.format(make([]byte, 0, 10), false))
}

func (d Decimal) StringMySQL() string {
	return string(d.format(make([]byte, 0, 10), false))
}

func (d Decimal) FormatMySQL(frac int32) []byte {
	rounded := d.Round(frac)
	return rounded.format(make([]byte, 0, 10), false)
}

// Round rounds the decimal to places decimal places.
// If places < 0, it will round the integer part to the nearest 10^(-places).
//
// Example:
//
// 	   NewFromFloat(5.45).Round(1).String() // output: "5.5"
// 	   NewFromFloat(545).Round(-1).String() // output: "550"
//
func (d Decimal) Round(places int32) Decimal {
	if d.exp == -places {
		return d
	}
	// truncate to places + 1
	ret := d.rescale(-places - 1)

	// add sign(d) * 0.5
	if ret.value.Sign() < 0 {
		ret.value.Sub(ret.value, fiveInt)
	} else {
		ret.value.Add(ret.value, fiveInt)
	}

	// floor for positive numbers, ceil for negative numbers
	_, m := ret.value.DivMod(ret.value, tenInt, new(big.Int))
	ret.exp++
	if ret.value.Sign() < 0 && m.Cmp(zeroInt) != 0 {
		ret.value.Add(ret.value, oneInt)
	}

	return ret
}

func (d Decimal) format(buf []byte, trimTrailingZeros bool) []byte {
	if d.exp >= 0 {
		return d.rescale(0).value.Append(buf, 10)
	}

	rawfmt := d.value.Append(nil, 10)
	if d.value.Sign() < 0 {
		rawfmt = rawfmt[1:]
		buf = append(buf, '-')
	}

	var fractionalPart []byte

	dExpInt := int(d.exp)
	if len(rawfmt) > -dExpInt {
		buf = append(buf, rawfmt[:len(rawfmt)+dExpInt]...)
		fractionalPart = rawfmt[len(rawfmt)+dExpInt:]
	} else {
		buf = append(buf, '0')
		num0s := -dExpInt - len(rawfmt)
		fractionalPart = make([]byte, 0, num0s+len(rawfmt))
		for z := 0; z < num0s; z++ {
			fractionalPart = append(fractionalPart, '0')
		}
		fractionalPart = append(fractionalPart, rawfmt...)
	}

	if trimTrailingZeros {
		i := len(fractionalPart) - 1
		for ; i >= 0; i-- {
			if fractionalPart[i] != '0' {
				break
			}
		}
		fractionalPart = fractionalPart[:i+1]
	}

	if len(fractionalPart) > 0 {
		buf = append(buf, '.')
		buf = append(buf, fractionalPart...)
	}
	return buf
}

func (d *Decimal) ensureInitialized() {
	if d.value == nil {
		d.value = new(big.Int)
	}
}

// RescalePair rescales two decimals to common exponential value (minimal exp of both decimals)
func RescalePair(d1 Decimal, d2 Decimal) (Decimal, Decimal) {
	d1.ensureInitialized()
	d2.ensureInitialized()

	if d1.exp == d2.exp {
		return d1, d2
	}

	baseScale := min(d1.exp, d2.exp)
	if baseScale != d1.exp {
		return d1.rescale(baseScale), d2
	}
	return d1, d2.rescale(baseScale)
}

func min(x, y int32) int32 {
	if x >= y {
		return y
	}
	return x
}

// largestForm returns the largest decimal that can be represented
// with the given amount of integral and fractional digits
// Example:
//	largestForm(1, 1) => 9.9
//	largestForm(5, 0) => 99999
//  largestForm(0, 5) => 0.99999
func largestForm(integral, fractional int32) Decimal {
	// nines is just a very long string of nines; to find the
	// largest form of a large decimal, we parse as many nines
	// as digits are in the form, then adjust the exponent
	// to where the decimal point should be.
	const nines = "99999999999999999999999999999999999999999999999999" +
		"99999999999999999999999999999999999999999999999999" +
		"99999999999999999999999999999999999999999999999999" +
		"99999999999999999999999999999999999999999999999999" +
		"99999999999999999999999999999999999999999999999999" +
		"99999999999999999999999999999999999999999999999999" +
		"99999999999999999999999999999999999999999999999999"

	digits := int(integral + fractional)
	if digits < len(limitsBigTab) {
		return Decimal{value: limitsBigTab[digits], exp: -fractional}
	}
	if digits < len(nines) {
		num, _ := new(big.Int).SetString(nines[:digits], 10)
		return Decimal{value: num, exp: -fractional}
	}
	panic("largestForm: too large")
}

func (d Decimal) Clamp(integral, fractional int32) Decimal {
	limit := largestForm(integral, fractional)
	if d.CmpAbs(limit) <= 0 {
		return d
	}
	if d.value.Sign() < 0 {
		return limit.NegInPlace()
	}
	return limit
}
