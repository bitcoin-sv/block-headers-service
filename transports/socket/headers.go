package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/libsv/headers-client/config"

	"github.com/libsv/headers-client"
)

type headersSocket struct {
	ws  *centrifuge.Client
	svc headers.BlockheaderService
	cfg *config.WocConfig
}

// NewHeaders will setup a new socket service - used to sync with woc.
// TODO - should this be a data layer?
func NewHeaders(ws *centrifuge.Client, cfg *config.WocConfig, svc headers.BlockheaderService) *headersSocket {
	h := &headersSocket{ws: ws, svc: svc, cfg: cfg}
	ws.OnMessage(h)
	ws.OnError(h)
	ws.OnConnect(h)
	ws.OnConnect(h)
	ws.OnDisconnect(h)
	ws.OnMessage(h)
	ws.OnError(h)

	ws.OnServerPublish(h)
	ws.OnServerJoin(h)
	ws.OnServerLeave(h)

	return h
}

func (h *headersSocket) OnConnect(_ *centrifuge.Client, e centrifuge.ConnectEvent) {
	log.Printf("Connected to chat with ID %s", e.ClientID)
}

func (h *headersSocket) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	log.Printf("Error: %s", e.Message)
}

func (h *headersSocket) OnMessage(_ *centrifuge.Client, e centrifuge.MessageEvent) {
	log.Printf("Message from server: %s", string(e.Data))
}

func (h *headersSocket) OnDisconnect(c *centrifuge.Client, e centrifuge.DisconnectEvent) {
	e.Reconnect = false
	log.Printf("Disconnected from chat: %s", e.Reason)
	height, _ := h.svc.Height(context.Background())
	c.Close()
	c = centrifuge.New(fmt.Sprintf("%s%d", h.cfg.URL, height), centrifuge.DefaultConfig())
	c.Connect()
	c.OnMessage(h)
	c.OnError(h)
	c.OnConnect(h)
	c.OnConnect(h)
	c.OnDisconnect(h)
	c.OnMessage(h)
	c.OnError(h)

	c.OnServerPublish(h)
	c.OnServerJoin(h)
	c.OnServerLeave(h)
	log.Println("Reconnected")
}

func (h *headersSocket) OnServerJoin(_ *centrifuge.Client, e centrifuge.ServerJoinEvent) {
	log.Printf("Server-side join to channel %s: %s (%s)", e.Channel, e.User, e.Client)
}

func (h *headersSocket) OnServerLeave(_ *centrifuge.Client, e centrifuge.ServerLeaveEvent) {
	log.Printf("Server-side leave from channel %s: %s (%s)", e.Channel, e.User, e.Client)
}

func (h *headersSocket) OnServerPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()
	var resp *headers.BlockHeader
	if err := json.Unmarshal(e.Data, &resp); err != nil {
		log.Println(err.Error())
	}
	// TODO - handle errors properly
	if err := h.svc.Create(ctx, *resp); err != nil {
		fmt.Println(err)
	}
}
