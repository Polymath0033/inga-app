package generator

import (
	"math/rand/v2"
	"strconv"
)

type StringGenerator interface {
	Generate() string
}

type numberGenerator struct {
	max int
}

func NewNumberGenerator(s string) StringGenerator {
	max, err := strconv.Atoi(s)
	if err != nil || max <= 0 {
		max = 1000000
	}
	return &numberGenerator{
		max: max,
	}
}

func (s *numberGenerator) Generate() string {
	value := rand.IntN(s.max)
	return strconv.Itoa(value)
}
