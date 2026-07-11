package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrChannelNotFound    = errors.New("channel not found")
	ErrPreferencesNotFound = errors.New("preferences not found")
	ErrDigestNotFound     = errors.New("digest not found")
	ErrInvalidNeurotype   = errors.New("invalid neurotype: must be adhd, autism, anxiety, unspecified, or ally")
	ErrInvalidDigestHour  = errors.New("digest hour must be between 0 and 23")
	ErrInvalidThreshold   = errors.New("focus threshold must be between 1 and 500")
	ErrSlackAPI           = errors.New("slack API error")
	ErrAIService          = errors.New("AI service error")
	ErrMCPServer          = errors.New("MCP server error")
	ErrRateLimited        = errors.New("rate limited")
	ErrUnauthenticated    = errors.New("unauthenticated")
)
