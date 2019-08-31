package wss

import (
	"encoding/base64"
	"github.com/genshen/wssocks/wss/term_view"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"io"
)

// proxy client handle one connection, send data to proxy server vai websocket.
type ProxyClient struct {
	Conn     io.ReadWriteCloser
	Id       ksuid.KSUID
	isClosed bool
}

// can't do large compute or communication here
func (p *ProxyClient) DispatchData(data *ProxyData) error {
	// decode base64
	if decodeBytes, err := base64.StdEncoding.DecodeString(data.DataBase64); err != nil { // todo ignore error
		log.Error("base64 decode error,", err)
		return err // skip error
	} else {
		// just write data back
		if _, err := p.Conn.Write(decodeBytes); err != nil {
			return err
		}
	}
	return nil
}

// close (tcp) connection
// the close command can be from server
func (p *ProxyClient) Close() {
	if p.isClosed {
		return
	}
	p.Conn.Close()
	p.isClosed = true
}

// handel socket dial results processing
// copy income connection data to proxy serve via websocket
func (p *ProxyClient) Serve(plog *term_view.ProgressLog, wsc *WebSocketClient,
	firstSendData []byte, proxyType int, addr string) error {
	plog.Update(term_view.Status{IsNew: true, Address: addr})
	defer plog.Update(term_view.Status{IsNew: false, Address: addr})
	defer wsc.Close(p.Id)

	estMsg := ProxyEstMessage{
		Type:     proxyType,
		Addr:     addr,
		WithData: false,
	}
	if firstSendData != nil {
		estMsg.WithData = true
		estMsg.DataBase64 = base64.StdEncoding.EncodeToString(firstSendData)
	}
	addrSend := WebSocketMessage{Type: WsTpEst, Id: p.Id.String(), Data: estMsg}
	if err := wsc.WriteWSJSON(&addrSend); err != nil {
		log.Error("json error:", err)
		return err
	}

	var buffer = make([]byte, 1024*64)
	for {
		if n, err := p.Conn.Read(buffer); err != nil {
			break
			// log.Println("read error:", err)
		} else if n > 0 {
			if err := wsc.WriteProxyMessage(p.Id, buffer[:n]); err != nil {
				log.Error("write error:", err)
				break
			}
		}
	}
	return nil
}
