package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/vandad1901/dbfreader/iransystem"
)

func readString(b []byte) string {
	n := len(b)
	for i, v := range b {
		if v == 0 {
			n = i
			break
		}
	}

	return string(b[:n])
}

const (
	CharacterType = 'C'
	DecimalType   = 'D'
	FloatType     = 'F'
	LogicalType   = 'L'
	MemoType      = 'M'
	NumericType   = 'N'
)

type Primitive interface {
	isPrimitive()
	ToString() string
}

type Character string

func (c Character) isPrimitive() {}

func (c *Character) ReadFromDBF(b []byte) error {
	decodedString, err := iransystem.DecodeBytes(b)
	if err != nil {
		return err
	}

	trimmedStr := strings.Trim(decodedString, "\u0000 ")

	*c = Character(trimmedStr)
	return nil
}

func (c Character) ToString() string {
	return string(c)
}

type Decimal decimal.Decimal

func (d Decimal) isPrimitive() {}

func (d *Decimal) ReadFromDBF(b []byte) error {
	s := readString(b)
	dec, err := decimal.NewFromString(s)
	if err != nil {
		return fmt.Errorf("parsing decimal from string %s: %w", s, err)
	}

	*d = Decimal(dec)

	return nil
}

func (d Decimal) ToString() string {
	return decimal.Decimal(d).String()
}

type Float float64

func (f Float) isPrimitive() {}

func (f *Float) ReadFromDBF(b []byte) error {
	s := readString(b)
	var value float64
	_, err := fmt.Sscanf(s, "%f", &value)
	if err != nil {
		return fmt.Errorf("parsing float from string %s: %w", s, err)
	}

	*f = Float(value)

	return nil
}

func (f Float) ToString() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

type Logical bool

func (l Logical) isPrimitive() {}

func (l *Logical) ReadFromDBF(b []byte) error {
	s := readString(b)

	if s == "T" || s == "t" || s == "Y" || s == "y" {
		*l = true
	} else {
		*l = false
	}

	return nil
}

func (l Logical) ToString() string {
	return strconv.FormatBool(bool(l))
}

type Memo string

func (m Memo) isPrimitive() {}

func (m *Memo) ReadFromDBF(_ []byte) error {
	return fmt.Errorf("memo is unsupported")
}

func (m Memo) ToString() string {
	return "MEMO FIELD (NOT SUPPORTED)"
}

type Numeric decimal.Decimal

func (n Numeric) isPrimitive() {}

func (n *Numeric) ReadFromDBF(b []byte) error {
	s := readString(b)
	dec, err := decimal.NewFromString(strings.TrimSpace(s))
	if err != nil {
		return fmt.Errorf("parsing numeric from string %s: %w", s, err)
	}

	*n = Numeric(dec)

	return nil
}

func (n Numeric) ToString() string {
	return decimal.Decimal(n).String()
}
