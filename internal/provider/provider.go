package provider

import (
	"context"
	"time"
)

type Service string

const (
	Platform          Service = "platform"
	ChatGPTEnterprise Service = "chatgpt_enterprise"
)

type Cost struct {
	Amount    float64
	Currency  string
	ProjectID string
	LineItem  string
	StartTime time.Time
	EndTime   time.Time
}

// Provider retrieves costs for a half-open time range [start, end).
type Provider interface {
	Service() Service
	Costs(ctx context.Context, start, end time.Time) ([]Cost, error)
}
