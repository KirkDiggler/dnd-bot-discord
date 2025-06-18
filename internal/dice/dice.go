package dice

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
)

type RollResult struct {
	Used    bool
	Total   int
	Highest int
	Lowest  int
	Rolls   []int
	Bonus   int
}

func Roll(count, size, bonus int) (*RollResult, error) {
	if count < 1 {
		return nil, errors.New("invalid dice count")
	}

	if size < 1 {
		return nil, errors.New("invalid dice size")
	}

	maxValue, minValue, total := 0, 0, 0

	out := make([]int, count)
	for i := 0; i < count; i++ {
		roll := rand.Intn(size) + 1
		total += roll
		if i == 0 {
			minValue = roll
			maxValue = roll
		}

		if minValue > roll {
			minValue = roll
		}

		if maxValue < roll {
			maxValue = roll
		}

		out[i] = roll
	}

	log.Println("Rolling", count, "d", size, ":", out, "total:", total, "min:", minValue, "max:", maxValue)
	return &RollResult{
		Total:   total + bonus,
		Highest: maxValue,
		Lowest:  minValue,
		Rolls:   out,
		Bonus:   bonus,
	}, nil
}

func RollString(diceString string) (*RollResult, error) {
	a := strings.Split(diceString, "+")
	var dice = diceString
	var bonus, diceSize, diceCount int
	var err error
	if len(a) == 2 {
		bonus, err = strconv.Atoi(a[1])
		if err != nil {
			return nil, errors.New("invalid dice string")
		}
		dice = a[0]
	}

	diceParts := strings.Split(dice, "d")
	if len(diceParts) != 2 {
		return nil, errors.New("invalid dice string")
	}

	strCount := diceParts[0]
	strSize := diceParts[1]

	diceCount, err = strconv.Atoi(strCount)
	if err != nil {
		return nil, errors.New("invalid dice string")
	}
	diceSize, err = strconv.Atoi(strSize)
	if err != nil {
		return nil, errors.New("invalid dice string")
	}

	return Roll(diceCount, diceSize, bonus)
}

func (r *RollResult) String() string {
	compact := strings.ReplaceAll(fmt.Sprintf("%v", r.Rolls), " ", "")
	return fmt.Sprintf("**%d** : %s", r.Total-r.Lowest, compact)
}
