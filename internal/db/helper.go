package db

import (
	"github.com/skip2/go-qrcode"
	"image/color"
)

func GenerateQrCode(from string) *qrcode.QRCode {
	return &qrcode.QRCode{
		Content:         from,
		Level:           1,
		VersionNumber:   1,
		ForegroundColor: color.Black,
		BackgroundColor: color.White,
		DisableBorder:   false,
	}
}
