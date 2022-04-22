package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"
)

// DBDrug represent a drug entry in the DB
type DBDrug struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	ValidationOption string             `bson:"validationOption"`
	ValidationData   string             `bson:"data"`
	Drug             Drug               `bson:"drug"`
}

type Drug struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	Manufacturer   string             `json:"manufacturer" bson:"manufacturer" validate:"required"`
	Name           string             `json:"name" bson:"name" validate:"required""`
	ManufacturedOn time.Time          `json:"manufactured_on" bson:"manufactureDate" validate:"required""`
	Expiry         time.Time          `json:"expiry" bson:"expiry" validate:"required"`
	BatchNumber    string             `json:"batch_number" bson:"batchNumber"`
	CreatedAt      time.Time          `json:"created_at" bson:"createdAt"`
}

const separator string = "-*-"

func (drug *Drug) String() string {
	return fmt.Sprintf("Manufacturer:%s%sName:%s%sManufacturedOn:%s%sExpiry:%s%sBatchNumber:%s",
		drug.Manufacturer, separator, drug.Name, separator, drug.ManufacturedOn,
		separator, drug.Expiry, separator, drug.BatchNumber)
}

// DrugFromString returns nil if @value isn't a Drug
func DrugFromString(value string) *Drug {
	parts := strings.Split(value, separator)

	// 5 corresponding to the number of fields embedded in String
	if len(parts) != 5 {
		return nil
	}
	manufactured, err := time.Parse(time.RFC3339, parts[2])
	if err != nil {
		return nil
	}
	expiry, err := time.Parse(time.RFC3339, parts[3])
	if err != nil {
		return nil
	}
	return &Drug{
		Manufacturer:   parts[0],
		Name:           parts[1],
		ManufacturedOn: manufactured,
		Expiry:         expiry,
		BatchNumber:    parts[4],
		CreatedAt:      time.Now(),
	}
}

var SampleDrug1 = Drug{
	Manufacturer:   "Heart Pharmaceutical",
	Name:           "Chloramphenicol",
	ManufacturedOn: time.Now(),
	Expiry:         time.Now().Add(time.Hour * 72),
	BatchNumber:    "1234BRQ",
	CreatedAt:      time.Now(),
}

var SampleDrug2 = Drug{
	Manufacturer:   "Heart Pharmaceutical",
	Name:           "Paracetamol",
	ManufacturedOn: time.Now(),
	Expiry:         time.Now().Add(time.Hour * 72),
	BatchNumber:    "0454BcQ",
	CreatedAt:      time.Now(),
}

var SampleDrug3 = Drug{
	Manufacturer:   "Heart Pharmaceutical",
	Name:           "Coalescere",
	ManufacturedOn: time.Now(),
	Expiry:         time.Now().Add(time.Hour * 72),
	BatchNumber:    "12qzKL6",
	CreatedAt:      time.Now(),
}

var SampleDrug4 = Drug{
	Manufacturer:   "Heart Pharmaceutical",
	Name:           "Cacoonamide",
	ManufacturedOn: time.Now(),
	Expiry:         time.Now().Add(time.Hour * 72),
	BatchNumber:    "SD198HJ",
	CreatedAt:      time.Now(),
}

var SampleDrug5 = Drug{
	Manufacturer:   "Heart Pharmaceutical",
	Name:           "Thanosavengers",
	ManufacturedOn: time.Now(),
	Expiry:         time.Now().Add(time.Hour * 72),
	BatchNumber:    "09ERWKV",
	CreatedAt:      time.Now(),
}
