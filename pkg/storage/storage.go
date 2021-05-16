package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Dacode45/redis-store/pkg/models"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-redis/redis/v8"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/nitishm/go-rejson/v4"
)

type Storage struct {
	mngr    *models.DataManager
	goredis *redis.Client
	json    *rejson.Handler
}

func NewStorage(mngr *models.DataManager, goredis *redis.Client, json *rejson.Handler) *Storage {
	return &Storage{mngr: mngr, goredis: goredis, json: json}
}

func (s *Storage) GetCollection(collectionID string) *CollectionClient {
	return &CollectionClient{CollectionID: collectionID, storage: s}
}

type CollectionClient struct {
	CollectionID string
	storage      *Storage
}

func (cc *CollectionClient) Add(ctx context.Context, id string, d map[string]interface{}) (*models.Document, error) {
	if id == "" {
		id = uuid.NewString()
	}
	doc := models.NewDocument(cc.CollectionID, id, d)
	doc.UpdatedAt = time.Now().Unix()

	key := cc.storage.mngr.DocKey(cc.CollectionID, doc.DocumentID)
	colKey := cc.storage.mngr.CollectionKey(cc.CollectionID)

	set, err := cc.storage.json.JSONSet(key, ".", doc.Data)
	if err != nil {
		return nil, err
	}
	if err := IsOk(set); err != nil {
		return nil, err
	}

	zadd := cc.storage.goredis.ZAdd(ctx, colKey, &redis.Z{Score: float64(doc.UpdatedAt), Member: key})
	if err := zadd.Err(); err != nil {
		return nil, err
	}

	pub := cc.storage.goredis.Publish(ctx, colKey, fmt.Sprintf("add:%s", key))
	if err := pub.Err(); err != nil {
		return nil, err
	}

	return doc, nil
}

func (cc *CollectionClient) GetAll(ctx context.Context) ([]*models.Document, error) {
	colKey := cc.storage.mngr.CollectionKey(cc.CollectionID)

	zrev := cc.storage.goredis.ZRevRangeWithScores(ctx, colKey, 0, -1)
	if err := zrev.Err(); err != nil {
		return nil, err
	}

	docs := make([]*models.Document, 0, len(zrev.Val()))
	for _, val := range zrev.Val() {
		key, ok := val.Member.(string)
		if !ok {
			return nil, errors.New("zscore key is not a string")
		}
		d, err := cc.storage.json.JSONGet(key, ".")
		if err != nil {
			return nil, err
		}
		var asMap map[string]interface{}
		err = json.Unmarshal(d.([]byte), &asMap)
		if err != nil {
			return nil, err
		}

		doc := models.NewDocument(cc.CollectionID, key, asMap)
		doc.UpdatedAt = int64(val.Score)
		docs = append(docs, doc)
	}

	return docs, nil
}

func (cc *CollectionClient) Listen(ctx context.Context) (<-chan *redis.Message, func() error) {
	colKey := cc.storage.mngr.CollectionKey(cc.CollectionID)

	sub := cc.storage.goredis.Subscribe(ctx, colKey)

	return sub.Channel(), sub.Close
}

func (cc *CollectionClient) Get(ctx context.Context, id string) (*models.Document, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	key := cc.storage.mngr.DocKey(cc.CollectionID, id)
	d, err := redigo.Bytes(cc.storage.json.JSONGet(key, "."))
	if err != nil {
		return nil, err
	}
	spew.Dump(string(d))
	var thing map[string]interface{}
	err = json.Unmarshal(d, &thing)
	if err != nil {
		return nil, err
	}
	_ = string(d)
	doc := models.NewDocument(cc.CollectionID, id, thing)
	return doc, nil
}
