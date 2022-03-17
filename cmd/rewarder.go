package main

import (
	"github.com/Hrtnet/social-activities/internal/model"
	"time"
)

type verificationResult struct {
	Drug                   *model.Drug `json:"drug"`
	CheckedOn              time.Time
	CheckedByWalletAddress string
	CurrentLocation        string `json:"current_location"`

	// specifies that the value for the tracking
	// option is found in the database/ on the blockchain
	TrackingOptionValueFound bool
}

func (v *verificationResult) rewardChecker() {
	// case 1: Random TrackingOptionValue
	// case 2: Expired
	// case 3: Valid
	// case 4: Counterfeit (containing copied/photocopied QR code)
	// In case 4, we use the CurrentLocation. The CurrentLocation is
	// echoed by the location of the NFT
	// If code is random code will include if drug manufacturer
	// isn't using our product yet
	switch {

	}
}
