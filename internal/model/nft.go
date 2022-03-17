package model

type SupplyChainPosition int

const (
	WithManufacturer SupplyChainPosition = iota
	WithDistributor
	EndUser
)

type NFT struct {
	ArtUrl              string              `json:"art_url"`
	SupplyChainPosition SupplyChainPosition `json:"supply_chain_position"`
	Twin                Drug                `json:"twin"`
}
