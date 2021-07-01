package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/centrifugal/centrifuge-go"

	"github.com/libsv/bitcoin-hc/config"

	"github.com/libsv/bitcoin-hc"
)

type headersPool struct{
	svc headers.BlockheaderService
	headers []*headers.BlockHeader
	rw sync.RWMutex
	syncCompleted bool
	done chan bool
}

func NewHeadersPool( svc headers.BlockheaderService) *headersPool{
	h := &headersPool{
		svc:     svc,
		headers: make([]*headers.BlockHeader,0),
		rw:      sync.RWMutex{},
		done:    make(chan bool),
	}
	go h.Start()
	return h
}

func (h *headersPool) Start(){
	t := time.NewTicker(time.Second*30)
	fmt.Println("starting sync with network...")
	for {
		select {
		case <-h.done:
			t.Stop()
			return
		case <-t.C:
			if !h.syncCompleted {
				height, _ := h.svc.Height(context.Background())
				if height.Synced{
					h.syncCompleted = true
					fmt.Println("sync with network complete, headers all cached, watching for new blocks...")
					h.done<-true
				}
			}
			if len(h.headers) == 0{
				continue
			}
			ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancelFn()
			if err := h.svc.CreateBatch(ctx, h.ReadAll()); err != nil{
				log.Println("failed to add batch " + err.Error())
			}
		}
	}
}

func (h *headersPool) Stop(){
	h.done <- true
	// clear buffer
	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()
	log.Println(fmt.Sprintf("adding %d records", len(h.headers)))
	if err := h.svc.CreateBatch(ctx, h.ReadAll()); err != nil{
		log.Println("failed to add batch " + err.Error())
	}
}

func (h *headersPool) Add(req *headers.BlockHeader){
	h.rw.Lock()
	defer h.rw.Unlock()
	if !h.syncCompleted {
		h.headers = append(h.headers, req)
		return
	}
	if err := h.svc.Create(context.Background(), *req);err != nil{
		fmt.Println(err)
	}
}

func (h *headersPool) ReadAll() []*headers.BlockHeader{
	h.rw.Lock()
	defer h.rw.Unlock()
	bh := make([]*headers.BlockHeader,0, len(h.headers))
	for _, hdr := range h.headers{
		bh = append(bh, hdr)
	}
	h.headers =  make([]*headers.BlockHeader,0)
	return bh
}


type headersSocket struct {
	ws  *centrifuge.Client
	svc headers.BlockheaderService
	cfg *config.WocConfig
	buffer *headersPool
}

// NewHeaders will setup a new socket service - used to sync with woc.
// TODO - should this be a data layer?
func NewHeaders(ws *centrifuge.Client, cfg *config.WocConfig, svc headers.BlockheaderService) *headersSocket {
	h := &headersSocket{ws: ws, svc: svc, cfg: cfg,buffer: NewHeadersPool(svc)}
	h.setup()
	return h
}

func (h *headersSocket) setup(){
	h.ws.OnDisconnect(h)
	h.ws.OnError(h)

	h.ws.OnServerPublish(h)
	h.ws.OnServerJoin(h)
	h.ws.OnServerLeave(h)
}

func (h *headersSocket) OnConnect(_ *centrifuge.Client, e centrifuge.ConnectEvent) {
	log.Printf("Connected to WoC with ID %s", e.ClientID)
}

func (h *headersSocket) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	log.Printf("Socket Error: %s", e.Message)
}


func (h *headersSocket) OnDisconnect(c *centrifuge.Client, e centrifuge.DisconnectEvent) {
	height, err := h.svc.Height(context.Background())
	if err != nil{
		fmt.Println(err)
		h.OnDisconnect(c, e)
		return
	}
	c.URL = fmt.Sprintf("%s%d", h.cfg.URL, height.Height)
}

func (h *headersSocket) OnServerJoin(_ *centrifuge.Client, e centrifuge.ServerJoinEvent) {
	log.Printf("Server-side join to channel %s: %s (%s)", e.Channel, e.User, e.Client)
}

func (h *headersSocket) OnServerLeave(_ *centrifuge.Client, e centrifuge.ServerLeaveEvent) {
	log.Printf("Server-side leave from channel %s: %s (%s)", e.Channel, e.User, e.Client)
}

func (h *headersSocket) OnServerPublish(c *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	go func() {
		var resp *headers.BlockHeader
		if err := json.Unmarshal(e.Data, &resp); err != nil {
			log.Println(err.Error())
		}
		h.buffer.Add(resp)
	}()
}

func (h *headersSocket) Close() {
	h.buffer.Stop()
}
