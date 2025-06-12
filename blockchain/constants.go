package blockchain

import (
	"math"
	"time"
)

const (
	// TargetBlockTime is the desired time between blocks (30 seconds)
	TargetBlockTime = 30 * time.Second

	// DifficultyAdjustmentInterval is the number of blocks between difficulty adjustments
	DifficultyAdjustmentInterval = 2016 // About 2 weeks with 30-second blocks

	// MaxBlockSize is the maximum size of a block in bytes (1 MB)
	MaxBlockSize = 1024 * 1024 // 1 MB

	// InitialDifficulty is the starting difficulty for the blockchain
	InitialDifficulty = 24

	// MinDifficulty is the minimum difficulty allowed
	MinDifficulty = 1

	// MaxDifficulty is the maximum difficulty allowed
	MaxDifficulty = 64

	// MaxTimeDeviation is the maximum time difference allowed between blocks
	MaxTimeDeviation = 2 * time.Hour
)

// CalculateNextDifficulty calculates the next difficulty based on the time taken to mine the previous blocks
func CalculateNextDifficulty(currentDifficulty int, actualTimespan time.Duration) int {
	// Constrain the actual timespan to prevent extreme adjustments
	minTimespan := TargetBlockTime.Seconds() / 4
	maxTimespan := TargetBlockTime.Seconds() * 4
	actualSeconds := actualTimespan.Seconds()

	if actualSeconds < minTimespan {
		actualSeconds = minTimespan
	}
	if actualSeconds > maxTimespan {
		actualSeconds = maxTimespan
	}

	// Calculate adjustment factor
	adjustment := TargetBlockTime.Seconds() / actualSeconds

	// Calculate new difficulty
	newDifficulty := float64(currentDifficulty) * adjustment

	// Ensure the difficulty stays within bounds
	if newDifficulty < MinDifficulty {
		newDifficulty = MinDifficulty
	}
	if newDifficulty > MaxDifficulty {
		newDifficulty = MaxDifficulty
	}

	return int(math.Round(newDifficulty))
}
