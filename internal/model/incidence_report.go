package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type IncidenceReport struct {
	ID                primitive.ObjectID       `json:"id" bson:"_id,omitempty"`
	UserID            string                   `json:"user_id" bson:"uid" validate:"required"`
	PharmacyName      string                   `json:"pharmacy_name" bson:"pharmacyName" validate:"required"`
	Description       string                   `json:"description" bson:"description" validate:"required"`
	PharmacyLocation  string                   `json:"pharmacy_location" bson:"pharmacyLocation" validate:"required"`
	EvidenceImagesUrl []string                 `json:"evidence_images_url" bson:"evidenceImagesUrl" validate:"required"`
	ReceiptImageUrl   string                   `json:"receipt_image_url" bson:"receiptImageUrl" validate:"required"`
	SubmittedOn       time.Time                `json:"submitted_on" bson:"submittedOn"`
	Updates           *[]IncidenceReportUpdate `json:"updates" bson:"updates,omitempty"`

	// update this field with something similar to
	// primitive.Timestamp{T:uint32(time.Now().Unix())}
	UpdatedAt primitive.Timestamp `json:"updated_at" bson:"updatedAt"`
}

type IncidenceReportUpdate struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	IncidenceReportID primitive.ObjectID `json:"parent_id" bson:"parent_id" validate:"required"`
	Images            []string           `json:"images" bson:"images"`
	Message           string             `json:"message" validate:"required"`

	// SentBy any official HeartNet partner e.g., NAFDAC
	SentBy string `json:"sent_by" validate:"required"`
}
