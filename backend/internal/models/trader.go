package models

import "sync"

type Trader struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Money        float64        `json:"money"`
	InitialMoney float64        `json:"initialMoney"`
	Holdings     map[string]int `json:"holdings"`
	mu           sync.RWMutex
}

func NewTrader(id, name string, money float64) *Trader {
	return &Trader{
		ID:           id,
		Name:         name,
		Money:        money,
		InitialMoney: money, // Store initial amount
		Holdings:     make(map[string]int),
	}
}
