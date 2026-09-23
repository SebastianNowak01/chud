package clock

import (
	"errors"
	"fmt"
)

const MaxRangeDays = 400

func ValidateRange(from, to Date) error {
	if from.IsZero() || to.IsZero() {
		return errors.New("zakres musi mieć początek i koniec")
	}
	if to.Before(from) {
		return errors.New("początek zakresu nie może być po końcu")
	}
	if from.DaysUntil(to) >= MaxRangeDays {
		return fmt.Errorf("zakres może mieć maksymalnie %d dni", MaxRangeDays)
	}
	return nil
}
