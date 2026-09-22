package users

import (
	"fmt"
	"math"
	"math/rand/v2"
	"regexp"
	"strings"

	"github.com/sebnow/chud/platform/apperr"
)

var colorPattern = regexp.MustCompile(`^#[0-9a-f]{6}$`)

// normalizeColor lowercases the value and validates it as #rrggbb.
func normalizeColor(color string) (string, *apperr.ServiceError) {
	color = strings.ToLower(strings.TrimSpace(color))
	if !colorPattern.MatchString(color) {
		return "", apperr.NewBadRequestError("color must be in #rrggbb format")
	}
	return color, nil
}

// randomColor picks a random hue with fixed saturation and lightness, so it stays readable.
func randomColor() string {
	const saturation, lightness = 0.65, 0.45
	hue := rand.Float64() * 360

	chroma := (1 - math.Abs(2*lightness-1)) * saturation
	x := chroma * (1 - math.Abs(math.Mod(hue/60, 2)-1))
	m := lightness - chroma/2

	var r, g, b float64
	switch {
	case hue < 60:
		r, g, b = chroma, x, 0
	case hue < 120:
		r, g, b = x, chroma, 0
	case hue < 180:
		r, g, b = 0, chroma, x
	case hue < 240:
		r, g, b = 0, x, chroma
	case hue < 300:
		r, g, b = x, 0, chroma
	default:
		r, g, b = chroma, 0, x
	}

	toByte := func(v float64) int { return int(math.Round((v + m) * 255)) }
	return fmt.Sprintf("#%02x%02x%02x", toByte(r), toByte(g), toByte(b))
}
