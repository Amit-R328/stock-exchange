package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("🧪 Testing Random Seed Behavior")
	fmt.Println("=================================")

	// Test 1: Without seed (should be deterministic)
	fmt.Println("\n1. Without seed (deterministic):")
	fmt.Println("   Run 1:")
	for i := 0; i < 5; i++ {
		change := (rand.Float64() - 0.5) * 0.04
		fmt.Printf("   %.4f ", change)
	}

	// Reset to default state
	rand.Seed(1)
	fmt.Println("\n   Run 2 (same seed):")
	for i := 0; i < 5; i++ {
		change := (rand.Float64() - 0.5) * 0.04
		fmt.Printf("   %.4f ", change)
	}

	// Test 2: With time-based seed (should be different)
	fmt.Println("\n\n2. With time-based seed (random):")
	fmt.Println("   Run 1:")
	seed1 := time.Now().UnixNano()
	rng1 := rand.New(rand.NewSource(seed1))
	for i := 0; i < 5; i++ {
		change := (rng1.Float64() - 0.5) * 0.04
		fmt.Printf("   %.4f ", change)
	}

	time.Sleep(1 * time.Millisecond) // Ensure different seed
	fmt.Println("\n   Run 2 (different seed):")
	seed2 := time.Now().UnixNano()
	rng2 := rand.New(rand.NewSource(seed2))
	for i := 0; i < 5; i++ {
		change := (rng2.Float64() - 0.5) * 0.04
		fmt.Printf("   %.4f ", change)
	}

	fmt.Printf("\n\n📊 Seed 1: %d\n", seed1)
	fmt.Printf("📊 Seed 2: %d\n", seed2)
	fmt.Printf("🎯 Seeds are different: %t\n", seed1 != seed2)

	fmt.Println("\n✅ This demonstrates that our price updater will generate")
	fmt.Println("   different price sequences each time the server starts!")
}
