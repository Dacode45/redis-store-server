package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Dacode45/redis-store/pkg/storage"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Server is a http server
type Server struct {
	storage *storage.Storage
}

func NewServer(s *storage.Storage) *Server {
	return &Server{storage: s}
}

func (s *Server) ListenAndServe(addr string) error {
	r := mux.NewRouter()
	r.HandleFunc("/collections/{collectionID}", s.CollectionHandler)

	http.Handle("/", r)
	log.Printf("server listening on %q", addr)
	err := http.ListenAndServe(addr, nil)
	return err
}

type CollectionHandler struct {
	collectionID string
	send         chan *RawResponse
	errs         chan error
	readClose    chan bool
	writeClose   chan bool
	store        *storage.Storage
	conn         *websocket.Conn
}

func (s *Server) CollectionHandler(w http.ResponseWriter, r *http.Request) {
	(w).Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	collectionID := vars["collectionID"]

	if collectionID == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Need categoryID")
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(w, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	handler := NewCollectionHandler(collectionID, conn, s.storage)
	handler.Run()
}

func NewCollectionHandler(collectionID string, conn *websocket.Conn, s *storage.Storage) *CollectionHandler {
	return &CollectionHandler{collectionID: collectionID, store: s, conn: conn, send: make(chan *RawResponse), errs: make(chan error), readClose: make(chan bool), writeClose: make(chan bool)}
}

func (h *CollectionHandler) errorPump() {
	var (
		rclose bool
		wclose bool
	)
	go func() {
		for !rclose && !wclose {
			select {
			case <-h.readClose:
				rclose = true
			case <-h.writeClose:
				wclose = true
			}
		}
		fmt.Println("should close errors channel")
		// close(h.errs)
	}()
	for err := range h.errs {
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func (h *CollectionHandler) readPump() {
	defer func() {
		close(h.readClose)
		// h.conn.Close()
	}()
	h.conn.SetReadLimit(maxMessageSize)
	h.conn.SetReadDeadline(time.Now().Add(pongWait))
	h.conn.SetPongHandler(func(string) error { h.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := h.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(message)
		var rawReq RawRequest
		err = json.Unmarshal(message, &rawReq)
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}
		rawReq.CollectionID = h.collectionID
		doer, err := rawReq.Parse()
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			doer.Do(ctx, *h.store, h.send, h.errs)
		}()
		defer cancel()
	}
}

func (h *CollectionHandler) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		close(h.writeClose)
		// h.conn.Close()
	}()
	for {
		select {
		case res, ok := <-h.send:
			h.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				h.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			message, err := json.Marshal(res)
			if err != nil {
				log.Printf("can't marshall: %v", err)
				continue
			}
			w, err := h.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("can't get nextwriter: %v", err)
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				log.Printf("can't close writer: %v", err)
				return
			}
		case <-ticker.C:
			h.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := h.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("can't write ping: %v", err)
				return
			}
		}
	}
}

func (h *CollectionHandler) Run() {
	go h.errorPump()
	go h.writePump()
	go h.readPump()
}
