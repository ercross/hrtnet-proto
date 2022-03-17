package model

import (
	"fmt"
	"strings"
	"time"
)

type Drug struct {
	Manufacturer   string    `json:"manufacturer"`
	Name           string    `json:"drug_name"`
	ManufacturedOn time.Time `json:"manufactured_on"`
	Expiry         time.Time `json:"expiry"`
	BatchNumber    string    `json:"batch_number"`
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
	}
}
