package socket_io

import (
	"net/http"

	"github.com/golite/litego/logger"
	"github.com/golite/litego/util"
	"github.com/googollee/go-socket.io"
)

type SocketioClient interface {
	OnMessage(msg string)
	OnDisconnect()
}

type SocketioServer interface {
	CreateClient(so socketio.Socket) *SocketioClient
}

func HandleSocketIOClient(so socketio.Socket, client SocketioClient) {
	logger.Debug("on connection")
	so.On("msg", func(msg string) { client.OnMessage(msg) })
	so.On("disconnection", func() { client.OnDisconnect() })
}

func ListenSocketIOServer(port int, soserver SocketioServer) {
	server, err := socketio.NewServer(nil)
	util.CheckError(err)
	server.SetMaxConnection(100000)

	server.On("connection", func(so socketio.Socket) {
		client := soserver.CreateClient(so)
		if client != nil {
			HandleSocketIOClient(so, client)
		}
	})

	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./assert")))
	logger.Debug("Serving at localhost:", port)
	http.ListenAndServe(":5000", nil)
}
