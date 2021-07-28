package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/theflyingcodr/centrifuge-go"

	"github.com/libsv/bitcoin-hc/config"

	headers "github.com/libsv/bitcoin-hc"
)

type headersBuffer struct {
	svc           headers.BlockheaderService
	headers       []*headers.BlockHeader
	rw            sync.RWMutex
	syncCompleted bool
	sync          chan<- bool
	done          chan bool
}

// NewHeadersBuffer will return a new headersBuffer which is used to cache records before
// adding to a database at intervals.
func NewHeadersBuffer(svc headers.BlockheaderService, synced chan<- bool) *headersBuffer {
	h := &headersBuffer{
		svc:     svc,
		headers: make([]*headers.BlockHeader, 0),
		rw:      sync.RWMutex{},
		done:    make(chan bool),
		sync:    synced,
	}
	go h.Start()
	return h
}

func (h *headersBuffer) Start() {
	t := time.NewTicker(time.Second * 30)
	log.Info().Msg("starting sync with network...")
	for {
		select {
		case <-h.done:
			t.Stop()
			return
		case <-t.C:

			if !h.syncCompleted {
				height, err := h.svc.Height(context.Background())
				if err != nil {
					log.Err(err)
					continue
				}
				if height.Synced {
					h.syncCompleted = true
					log.Info().Msg("sync with network complete, headers all cached, watching for new blocks...")
					h.sync <- true
					h.done <- true
				}
			}
			if len(h.headers) == 0 {
				continue
			}

			ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancelFn()
			if err := h.svc.CreateBatch(ctx, h.ReadAll()); err != nil {
				log.Error().Msg("failed to add batch " + err.Error())
			}
			height, _ := h.svc.Height(context.Background())
			log.Info().Msgf("syncing headers, now at header %d of %d",
				height.Height+len(h.headers), height.NetworkNeight)
		}
	}
}

func (h *headersBuffer) Stop() {
	h.done <- true
	// clear buffer
	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()
	log.Info().Msgf("adding %d records", len(h.headers))
	if err := h.svc.CreateBatch(ctx, h.ReadAll()); err != nil {
		log.Error().Msg("failed to add batch " + err.Error())
	}
}

func (h *headersBuffer) Add(req *headers.BlockHeader) {
	h.rw.Lock()
	defer h.rw.Unlock()
	if !h.syncCompleted {
		h.headers = append(h.headers, req)
		return
	}
	log.Info().Msgf("new header received at height %d with hash %s", req.Height, req.Hash)
	if err := h.svc.Create(context.Background(), *req); err != nil {
		log.Err(err)
	}
}

func (h *headersBuffer) ReadAll() []*headers.BlockHeader {
	h.rw.Lock()
	defer h.rw.Unlock()
	bh := make([]*headers.BlockHeader, 0, len(h.headers))
	bh = append(bh, h.headers...)
	h.headers = make([]*headers.BlockHeader, 0)
	return bh
}

type headersSocket struct {
	ws     *centrifuge.Client
	svc    headers.BlockheaderService
	cfg    *config.WocConfig
	buffer *headersBuffer
	synced bool
}

// NewHeaders will setup a new socket service - used to sync with woc.
// TODO - should this be a data layer?
func NewHeaders(ws *centrifuge.Client, cfg *config.WocConfig, svc headers.BlockheaderService) *headersSocket {
	syncChan := make(chan bool)
	h := &headersSocket{
		ws:     ws,
		svc:    svc,
		cfg:    cfg,
		buffer: NewHeadersBuffer(svc, syncChan),
	}
	h.setup()
	go func() {
		for range syncChan {
			h.ws.SetURL("wss://socket.whatsonchain.com/blockheaders")
			_ = h.ws.Connect()
			h.synced = true
			return
		}
	}()
	return h
}

func (h *headersSocket) setup() {
	h.ws.OnDisconnect(h)
	h.ws.OnError(h)

	h.ws.OnServerPublish(h)
	h.ws.OnServerJoin(h)
	h.ws.OnServerLeave(h)
}

func (h *headersSocket) OnConnect(_ *centrifuge.Client, e centrifuge.ConnectEvent) {
	log.Debug().Msgf("Connected to WoC with ID %s", e.ClientID)
}

func (h *headersSocket) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	log.Debug().Msgf("Socket Error: %s", e.Message)
}

func (h *headersSocket) OnDisconnect(c *centrifuge.Client, e centrifuge.DisconnectEvent) {
	log.Debug().Msgf("Disconnected from server Error: %s... reconnecting", e.Reason)
	height, err := h.svc.Height(context.Background())
	if err != nil {
		h.OnDisconnect(c, e)
		return
	}
	log.Debug().Msg("reconnected")
	if h.synced {
		c.SetURL("wss://socket.whatsonchain.com/blockheaders")
		return
	}
	c.SetURL(fmt.Sprintf("%s%d", h.cfg.URL, height.Height))
}

func (h *headersSocket) OnServerJoin(_ *centrifuge.Client, e centrifuge.ServerJoinEvent) {
	log.Debug().Msgf("Server-side join to channel %s: %s (%s)", e.Channel, e.User, e.Client)
}

func (h *headersSocket) OnServerLeave(_ *centrifuge.Client, e centrifuge.ServerLeaveEvent) {
	log.Debug().Msgf("Server-side leave from channel %s: %s (%s)", e.Channel, e.User, e.Client)
}

func (h *headersSocket) OnServerPublish(c *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	go func() {
		var resp *headers.BlockHeader
		if err := json.Unmarshal(e.Data, &resp); err != nil {
			log.Err(err)
		}
		h.buffer.Add(resp)
	}()
}

func (h *headersSocket) Close() {
	h.buffer.Stop()
}
