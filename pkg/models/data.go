package models

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Document struct {
	CollectionID string                 `json:"collectionID"`
	DocumentID   string                 `json:"documentID"`
	Data         map[string]interface{} `json:"data"`
	UpdatedAt    int64                  `json:"updatedAt"`
}

func NewDocument(collectionID string, documentID string, data map[string]interface{}) *Document {
	return &Document{CollectionID: collectionID, DocumentID: documentID, Data: data}
}

type DataManager struct {
	namespace string
}

func NewDataManager(namespace string) *DataManager {
	return &DataManager{namespace: namespace}
}

func (m *DataManager) DocKey(collection string, document string) string {
	return fmt.Sprintf("%s:collections:%s:%s", m.namespace, collection, document)
}

func (m *DataManager) CollectionKey(collection string) string {
	return fmt.Sprintf("%s:collection:%s", m.namespace, collection)
}

func (doc *Document) ToJSON() (string, error) {
	var prettyJSON bytes.Buffer
	enc := json.NewEncoder(&prettyJSON)
	enc.SetIndent("", "\t")
	err := enc.Encode(&doc.Data)

	if err != nil {
		return "", err
	}

	return prettyJSON.String(), nil
}
