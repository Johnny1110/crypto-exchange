package ws

import (
	"context"
)

type WSDataFeeder struct {
	msgCh  chan *WSFeedPackage
	stopCh chan struct{}
	hub    *Hub
}

func (m *WSDataFeeder) Feed(ctx context.Context, pkg *WSFeedPackage) {
	m.msgCh <- pkg
}

func NewWSDataFeeder(hub *Hub) *WSDataFeeder {
	return &WSDataFeeder{
		hub:    hub,
		msgCh:  make(chan *WSFeedPackage, 1000),
		stopCh: make(chan struct{}),
	}
}

func (m *WSDataFeeder) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case pkg := <-m.msgCh:
				m.hub.BroadcastToSubscribers(pkg.Key, pkg.Data)
			case <-m.stopCh:
				return
			}
		}
	}()
}
