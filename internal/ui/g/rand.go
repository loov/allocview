package g

import (
	"math/rand"
	"time"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func Rand() float32 {
	return random.Float32()
}

func RandBetween(min, max float32) float32 {
	return random.Float32()*(max-min) + min
}

func RandColor(saturation, lightness float32) Color {
	return HSL(Rand(), saturation, lightness)
}
