package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Shared state to hold OTP and manage synchronization
type OTPState struct {
	sync.Mutex
	Otp       string
	WaitGroup *sync.WaitGroup
}

type ShareRequest struct {
	Varimg      [][]byte
	Contacts    []string `json:"contacts"`
	PhoneNumber string   `json:"phone"`
}

type ContactDetails struct {
	Contacts []string `json:"contacts"`
	Phone    string   `json:"phone"`
}

type WebsocketServer struct {
	Conns map[*websocket.Conn]bool
	Lock  sync.Mutex
}
