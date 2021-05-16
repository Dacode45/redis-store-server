package server

import "github.com/Dacode45/redis-store/pkg/models"

type RawResponse struct {
	Type         RequestType        `json:"kind"`
	CollectionID string             `json:"collectionID"`
	DocumentID   string             `json:"documentID"`
	RequestID    string             `json:"requestID"`
	Data         *models.Document   `json:"data"`
	MultiData    []*models.Document `json:"multiData"`
}

func ReturnGotCollection(reqID, collectionID string, multiData []*models.Document) *RawResponse {
	return &RawResponse{
		Type:         RTGotCollection,
		CollectionID: collectionID,
		RequestID:    reqID,
		MultiData:    multiData,
	}
}

func ReturnUpdatedDocument(reqID string, doc *models.Document) *RawResponse {
	return &RawResponse{
		Type:         RTUpdatedDocument,
		CollectionID: doc.CollectionID,
		RequestID:    reqID,
		Data:         doc,
	}
}
