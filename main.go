package main

import (
	env "CustomPhotoGenerator/m-v0/ENV"
	server "CustomPhotoGenerator/m-v0/Server"
)

func init() {
	env.SetENV()
}

func main() {
	server.Server()
}
