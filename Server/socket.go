package server

import (
	models "CustomPhotoGenerator/m-v0/Models"
	telegramclient "CustomPhotoGenerator/m-v0/TelegramClient"
	"CustomPhotoGenerator/m-v0/utils"
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WebsocketServer struct {
	conns map[*websocket.Conn]bool
	lock  sync.Mutex
}

func newserver() *WebsocketServer {
	return &WebsocketServer{conns: make(map[*websocket.Conn]bool)}
}

func (s *WebsocketServer) handleWS(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error while upgrading connection:", err)
		return
	}
	fmt.Println("Incoming request from:", conn.RemoteAddr())
	var wg sync.WaitGroup
	s.procedure(conn, &wg)
	wg.Wait()
	s.lock.Lock()
	s.conns[conn] = true
	s.lock.Unlock()

	defer func() {
		s.lock.Lock()
		delete(s.conns, conn)
		s.lock.Unlock()
		conn.Close()
	}()

}

func (s *WebsocketServer) procedure(conn *websocket.Conn, wg *sync.WaitGroup) {
	otpState := &models.OTPState{ // Initialize otpState
		WaitGroup: &sync.WaitGroup{},
	}
	baseImg := s.readPhoto(conn)
	bulkPosterInst := s.readBulkPosterData(conn)
	varImgs, err := utils.PrintVarContent(baseImg, bulkPosterInst)
	if err != nil {
		log.Println(err)
		return // Return early if there's an error
	}
	cd := s.readContactDetails(conn)
	otpState.WaitGroup.Add(1)
	wg.Add(1)
	go func() {
		log.Println("Starting authentication and message sending process.")
		err := telegramclient.AuthenticateAndSend(cd, varImgs, otpState)
		if err != nil {
			log.Println("Error during authentication and sending message:", err)
			conn.WriteMessage(websocket.TextMessage, []byte("Error sending images"))
		}
		conn.WriteMessage(websocket.TextMessage, []byte("Successfully sent the images"))
		wg.Done()
	}()
	conn.WriteMessage(websocket.TextMessage, []byte("Please send otp"))
	otp := s.readOTP(conn)
	otpState.Lock()
	otpState.Otp = otp
	otpState.Unlock()
	otpState.WaitGroup.Done()
}

func (s *WebsocketServer) readPhoto(conn *websocket.Conn) image.Image {
	_, imgBytes, err := conn.ReadMessage()
	if err != nil {
		log.Println("Error reading message:", err)
		return nil
	}
	buf := bytes.NewBuffer(imgBytes)
	fmt.Println("Received binary data")
	img, _, err := image.Decode(buf)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return img
}

func (s *WebsocketServer) readBulkPosterData(conn *websocket.Conn) models.IshareRequest {
	var data models.IshareRequest
	err := conn.ReadJSON(&data)
	if err != nil {
		log.Fatal(err)
		return models.IshareRequest{}
	}
	return data
}

func (s *WebsocketServer) readContactDetails(conn *websocket.Conn) models.ContactDetails {
	var data models.ContactDetails
	err := conn.ReadJSON(&data)
	if err != nil {
		log.Fatal(err)
		return models.ContactDetails{}
	}
	return data
}

func (s *WebsocketServer) readOTP(conn *websocket.Conn) string {
	_, otp, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		return ""
	}
	fmt.Println("OTP :", string(otp))
	return string(otp)
}

func RunSocketServer() {
	s := newserver()
	http.HandleFunc("/ws", s.handleWS)
	fmt.Println("WebSocket server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
