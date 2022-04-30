package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type IncidenceReport struct {
	ID                primitive.ObjectID `json:"id" bson:"id"`
	UserID            string             `json:"user_id" bson:"uid" validate:"required"`
	PharmacyName      string             `json:"pharmacy_name" bson:"pharmacyName" validate:"required"`
	Description       string             `json:"description" bson:"description" validate:"required"`
	PharmacyLocation  string             `json:"pharmacy_location" bson:"pharmacyLocation" validate:"required"`
	EvidenceImagesUrl []string           `json:"evidence_images_url" bson:"evidenceImagesUrl" validate:"required"`
	ReceiptImageUrl   string             `json:"receipt_image_url" bson:"receiptImageUrl" validate:"required"`
	SubmittedOn       time.Time          `json:"submitted_on" bson:"submittedOn"`

	// update this field with something similar to
	// primitive.Timestamp{T:uint32(time.Now().Unix())}
	UpdatedAt primitive.Timestamp `json:"updated_at" bson:"updatedAt"`
}
