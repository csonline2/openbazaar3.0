package core

import (
	"errors"
	"fmt"
)

var (
	// ErrListingCoinDivisibilityIncorrect - coin divisibility err
	ErrListingCoinDivisibilityIncorrect = errors.New("incorrect coinDivisibility")

	// ErrUnknownListingVersion is returned when creating an order for a listing version
	// greater than our largest known version.
	ErrUnknownListingVersion = errors.New("upgraded needed to purchase listing version")

	// ErrPublishingActive is returned if the node is still publishing during a shutdown.
	ErrPublishingActive = errors.New("publishing active - use force to shutdown")

	// ErrIPFSDelayedShutdown is returned if the IPFS is lagging during shutdown.
	ErrIPFSDelayedShutdown = errors.New("ipfs node delayed shutdown")

	// ErrPeerUnreachable is returned due to an inability to connect to a peer.
	ErrPeerUnreachable = errors.New("peer unreachable")
)

type ErrTooManyItems []string

func (e ErrTooManyItems) Error() string {
	return fmt.Sprintf("field: %s has a size greater than the max of %s", e[0], e[1])
}

type ErrTooManyCharacters []string

func (e ErrTooManyCharacters) Error() string {
	return fmt.Sprintf("field: %s has a length greater than the max of %s", e[0], e[1])
}

type ErrMissingField string

func (e ErrMissingField) Error() string {
	return fmt.Sprintf("missing field: %s", string(e))
}

// ErrPriceModifierOutOfRange - customize limits for price modifier
type ErrPriceModifierOutOfRange struct {
	Min float64
	Max float64
}

func (e ErrPriceModifierOutOfRange) Error() string {
	return fmt.Sprintf("priceModifier out of range: [%.2f, %.2f]", e.Min, e.Max)
}

// ErrCryptocurrencyListingIllegalField - invalid field err
type ErrCryptocurrencyListingIllegalField string

func (e ErrCryptocurrencyListingIllegalField) Error() string {
	return illegalFieldString("cryptocurrency listing", string(e))
}

// ErrMarketPriceListingIllegalField - invalid listing field err
type ErrMarketPriceListingIllegalField string

func (e ErrMarketPriceListingIllegalField) Error() string {
	return illegalFieldString("market price listing", string(e))
}

func illegalFieldString(objectType string, field string) string {
	return fmt.Sprintf("Illegal %s field: %s", objectType, field)
}
