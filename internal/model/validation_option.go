package model

type ValidationOption int

const (
	RFID ValidationOption = iota

	// QrCode texts will be generated as JWT encapsulating
	// drug data
	QrCode
	Code
)
