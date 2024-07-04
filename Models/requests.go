package models

import "sync"

// Shared state to hold OTP and manage synchronization
type OTPState struct {
	sync.Mutex
	Otp       string
	WaitGroup sync.WaitGroup
}

type ShareRequest struct {
	Varimg      [][]byte
	Contacts    []string `json:"contacts"`
	PhoneNumber string   `json:"phone"`
}
