package ws_socks

import (
	"bytes"
	"encoding/base64"
	"github.com/segmentio/ksuid"
	"sync"
)

// write data to WebSocket server or client
type Base64WSBufferWriter struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

// implement Write interface to write bytes from ssh server into bytes.Buffer.
func (b *Base64WSBufferWriter) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Write(p)
}

// flush all data in this buff into WebSocket.
func (b *Base64WSBufferWriter) Flush(messageType int, id ksuid.KSUID, cws ConcurrentWebSocketInterface) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	length := b.buffer.Len()
	if length != 0 {
		dataBase64 := base64.StdEncoding.EncodeToString(b.buffer.Bytes())
		jsonData := WebSocketMessage{
			Id:   id.String(),
			Type: WsTpData,
			Data: ProxyData{DataBase64: dataBase64},
		}
		if err := cws.WriteWSJSON(&jsonData); err != nil {
			return 0, err
		}
		b.buffer.Reset()
		return length, nil
	}
	return 0, nil
}
