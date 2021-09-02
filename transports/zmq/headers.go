package zmq

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ordishs/go-bitcoin"
	"github.com/pkg/errors"

	headers "github.com/libsv/bitcoin-hc"
)

type headersHandler struct {
	svc    headers.BlockheaderService
	bc     chan []string
	bcdone chan bool
}

func NewHeadersHandler(svc headers.BlockheaderService) *headersHandler {
	return &headersHandler{
		svc: svc,
		bc:  make(chan []string),
	}
}

// Register will setup zmq with a handler
func (h *headersHandler) Register(z *bitcoin.ZMQ) {
	if err := z.Subscribe("hashblock", h.bc); err != nil {
		log.Fatalln(err)
	}
}

func (h *headersHandler) Header() error {
	for {
		select {
		case rawHdr := <-h.bc:
			go func() {
				ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
				defer cancelFn()
				fmt.Println(fmt.Sprintf("%+V", rawHdr))
				hdr, err := h.svc.Header(ctx, headers.HeaderArgs{Blockhash: rawHdr[1]})
				if err != nil {
					log.Println(err)
					return
				}
				if err := h.svc.Create(ctx, *hdr); err != nil {
					log.Println(err)
					return
				}
			}()
		case <-h.bcdone:
			close(h.bc)
			return nil
		}
	}
}

func (h *headersHandler) Close(z *bitcoin.ZMQ) error {
	if err := z.Unsubscribe("rawblock", h.bc); err != nil {
		return errors.WithMessage(err, "failed to unsubscribe rawblock")
	}
	close(h.bc)
	return nil
}
