package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/lillilli/logger"

	"github.com/lillilli/graphex/config"
	"github.com/lillilli/graphex/server/handler"
	"github.com/lillilli/graphex/server/hub"
	"github.com/lillilli/graphex/watcher"
)

var upgrader = websocket.Upgrader{}

// Server - ws server interface
type Server interface {
	Start() error
	Stop()
}

type server struct {
	hub     *hub.Hub
	cfg     *config.Config
	manager handler.Manager

	log logger.Logger

	ctx    context.Context
	cancel context.CancelFunc
}

// NewServer - return a new ws server instance
func NewServer(cfg *config.Config, watcher watcher.Watcher) Server {
	eventEmitter := hub.NewEventEmitter(watcher)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &server{
		ctx: ctx,
		cfg: cfg,

		hub:     hub.New(ctx, eventEmitter),
		manager: handler.NewManager(eventEmitter, watcher),

		cancel: cancel,
		log:    logger.NewLogger("ws server"),
	}
}

// Start - start receive and handling messages
func (s *server) Start() error {
	s.log.Info("Starting ...")
	addr := fmt.Sprintf("%s:%d", s.cfg.WS.Host, s.cfg.WS.Port)

	s.hub.Start()
	http.Handle("/", http.FileServer(http.Dir(s.cfg.FrontendDistPath)))
	http.HandleFunc("/ws", s.handleWS)

	go func() {
		s.log.Errorf("Serving error: %v", http.ListenAndServe(addr, nil))
	}()

	s.log.Infof("Start listen on ws://%s/", addr)
	return nil
}

func (s server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Errorf("Upgrading ws connection failed: %v", err)
		return
	}

	client := s.hub.NewClient(conn)
	go s.manager.HandleClientEvents(client)
}

// Stop - stop server work
func (s server) Stop() {
	s.cancel()
}
