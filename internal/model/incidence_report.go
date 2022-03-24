package model

import "time"

type IncidenceReport struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	PharmacyName      string    `json:"pharmacy_name"`
	Description       string    `json:"description"`
	PharmacyLocation  string    `json:"pharmacy_location"`
	EvidenceImagesUrl string    `json:"evidence_images_url"`
	ReceiptImageUrl   string    `json:"receipt_image_url"`
	Submitted         time.Time `json:"submitted"`
}
