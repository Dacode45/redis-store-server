package server

import (
	"context"
	"fmt"
	"log"

	"github.com/Dacode45/redis-store/pkg/models"
	"github.com/Dacode45/redis-store/pkg/storage"
	"github.com/davecgh/go-spew/spew"
)

type RequestType string

const (
	RTGotCollection   = "got:collection"
	RTGetCollection   = "get:collection"
	RTPostCollection  = "post:collection"
	RTUpdatedDocument = "updated:document"
)

type RawRequest struct {
	Type         RequestType            `json:"type"`
	CollectionID string                 `json:"collectionID"`
	DocumentID   string                 `json:"documentID"`
	RequestID    string                 `json:"requestID"`
	Data         map[string]interface{} `json:"data"`
	MultiData    []map[string]interface{}
}

type Request interface {
	Do(ctx context.Context, store storage.Storage, out chan<- *RawResponse, errs chan<- error)
}

type GetCollectionRequest struct {
	*RawRequest
}

func (req *RawRequest) Parse() (Request, error) {
	switch req.Type {
	case RTGetCollection:
		return &GetCollectionRequest{req}, nil
	case RTPostCollection:
		return &PostCollectionRequest{req}, nil
	default:
		return nil, fmt.Errorf("cannot parse [%s]%+v", req.Type, req)
	}
}

func (req *GetCollectionRequest) Do(ctx context.Context, store storage.Storage, out chan<- *RawResponse, errs chan<- error) {
	col := store.GetCollection(req.CollectionID)
	docs, err := col.GetAll(ctx)
	if req.OnErr(ctx, errs, err) != nil {
		return
	}

	req.SendReturnGotCollection(ctx, out, docs)

	subs, close := col.Listen(ctx)
	defer func() {
		if err := close(); err != nil {
			log.Println(err.Error())
		}
	}()
	for {
		select {
		case msg := <-subs:
			log.Print("Got update to collection")
			spew.Dump(msg)
			next, err := col.GetAll(ctx)
			if req.OnErr(ctx, errs, err) != nil {
				return
			}
			req.SendReturnGotCollection(ctx, out, next)
		case <-ctx.Done():
		}
	}
}

type PostCollectionRequest struct {
	*RawRequest
}

func (req *PostCollectionRequest) Do(ctx context.Context, store storage.Storage, out chan<- *RawResponse, errs chan<- error) {
	col := store.GetCollection(req.CollectionID)
	doc, err := col.Add(ctx, req.DocumentID, req.Data)
	if req.OnErr(ctx, errs, err) != nil {
		return
	}
	log.Println("updated document")
	spew.Dump(doc)
	req.SendUpdatedDocument(ctx, out, doc)
}

func (req *RawRequest) OnErr(ctx context.Context, errs chan<- error, err error) error {
	if err != nil {
		select {
		case <-ctx.Done():
			fmt.Println("closing on error")
		case errs <- err:
		}
		return err
	}
	return nil
}

func (req *RawRequest) SendReturnGotCollection(ctx context.Context, out chan<- *RawResponse, docs []*models.Document) {
	res := ReturnGotCollection(req.RequestID, req.CollectionID, docs)

	select {
	case out <- res:
	case <-ctx.Done():
		fmt.Println("closing send return ")

		return
	}
}

func (req *RawRequest) SendUpdatedDocument(ctx context.Context, out chan<- *RawResponse, doc *models.Document) {
	res := ReturnUpdatedDocument(req.RequestID, doc)

	select {
	case out <- res:
		log.Println("sent add message update")
	case <-ctx.Done():
		fmt.Println("closing send update")

		return
	}
}
