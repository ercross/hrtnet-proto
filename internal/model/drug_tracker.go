package model

// Tracker maps drugs to HeartNet tracking options.
// Since the 3 tracking options rely on an encrypted byte of data
// that contains some metadata about the drug,
type Tracker struct {
	ID      string
	Drug    *Drug            `json:"drug_id"`
	NFT     *NFT             `json:"nft_id"`
	Option  ValidationOption `json:"option"`
	IsValid bool             `json:"is_valid" bson:"is_valid"`

	// OptionValue is the string embedded in
	// the physical representation of the ValidationOption
	OptionValue []byte `json:"option_value"`
}
