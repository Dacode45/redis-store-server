package storage

import (
	"context"
	"testing"

	"github.com/Dacode45/redis-store/pkg/models"
	"github.com/stretchr/testify/assert"
)

var testNamespace = "rebse-test"
var testCollection = "test-col"

func TestCollectionAdd(t *testing.T) {
	ctx := context.Background()
	mngr := models.NewDataManager(testNamespace)
	cli, rh := DefaultRedisConn()
	s := NewStorage(mngr, cli, rh)

	col := s.GetCollection(testCollection)
	doc, err := col.Add(ctx, "", map[string]interface{}{"hello": "hello"})
	onErr(t, err)

	data, err := col.Get(ctx, doc.DocumentID)
	onErr(t, err)

	s1, e1 := data.ToJSON()
	s2, e2 := doc.ToJSON()
	onErr(t, e1)
	onErr(t, e2)
	assert.Equal(t, s1, s2)
}

func onErr(t *testing.T, err error) {
	if !assert.NoError(t, err) {
		panic(err.Error())
	}
}
