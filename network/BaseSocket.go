package network

var ID uint64

type BaseSocket interface {
	OnCheckPackage(data *[]byte) (bool, []byte)
	OnConnect(pClient *TcpClient) // 主动链接
	OnDataIn(data []byte)
	OnClose()
}

type BaseListen interface {
	OnAccept(pClient *TcpClient) BaseSocket
}
