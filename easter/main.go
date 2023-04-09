package main

import (
	"flag"
	"fmt"
	"time"
)

// GaussEasterFormula calculates the day and month of Easter for a given year
// using the Gauss Easter formula. It returns the day and month as integers.
func GaussEasterFormula(year int) (int, int) {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	easterMonth := (h + l - 7*m + 114) / 31
	easterDay := ((h + l - 7*m + 114) % 31) + 1

	return easterDay, easterMonth
}

func main() {
	var year int
	var currentYear = time.Now().Year()
	flag.IntVar(&year, "year", currentYear, "year")
	flag.Parse()

	easterDay, easterMonth := GaussEasterFormula(year)
	fmt.Printf("Easter in %d is on %d.%d\n", year, easterDay, easterMonth)
}
