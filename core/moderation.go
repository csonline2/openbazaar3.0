package core

import (
	"context"
	"errors"
	"github.com/cpacia/openbazaar3.0/database"
	"github.com/cpacia/openbazaar3.0/models"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"strings"
)

const (
	// moderatorTopic is the DHT key at which moderator "providers" are stored.
	moderatorTopic = "openbazaar:moderators"

	// moderatorCid is the cid path of the provider block.
	moderatorCid = "/ipld/QmV9mSiAvEMvv6JyVYFaojPb4Se3XSpb4tW35AcjGfVqxb"
)

// SetSelfAsModerator sets this node as a node that is offering moderation services.
// It will update the profile with the moderation info, set itsef as a moderator
// in the DHT so it can be discovered by other peers, and publish.
func (n *OpenBazaarNode) SetSelfAsModerator(ctx context.Context, modInfo *models.ModeratorInfo, done chan struct{}) error {
	if (int(modInfo.Fee.FeeType) == 0 || int(modInfo.Fee.FeeType) == 2) && modInfo.Fee.FixedFee == nil {
		maybeCloseDone(done)
		return errors.New("fixed fee must be set when using a fixed fee type")
	}

	var currencies []string
	// TODO: check preferred currencies in settings and use them here if they exist.
	for ct := range n.multiwallet {
		currencies = append(currencies, ct.CurrencyCode())
	}
	for _, cc := range currencies {
		modInfo.AcceptedCurrencies = append(modInfo.AcceptedCurrencies, normalizeCurrencyCode(cc))
	}

	err := n.repo.DB().Update(func(tx database.Tx)error {
		profile, err := tx.GetProfile()
		if err != nil {
			return err
		}
		profile.ModeratorInfo = modInfo
		profile.Moderator = true

		if err := tx.SetProfile(profile); err != nil {
			return err
		}

		api, err := coreapi.NewCoreAPI(n.ipfsNode)
		if err != nil {
			return err
		}
		// This sets us as a "provider" in the DHT for the moderator key.
		// Other peers can find us by doing a DHT GetProviders query for
		// the same key.
		_, err = api.Block().Put(ctx, strings.NewReader(moderatorTopic))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		maybeCloseDone(done)
		return err
	}
	n.Publish(done)
	return nil
}

// RemoveSelfAsModerator removes this node as a moderator in the DHT and updates
// the profile and publishes.
func (n *OpenBazaarNode) RemoveSelfAsModerator(ctx context.Context, done chan<- struct{}) error {
	err := n.repo.DB().Update(func(tx database.Tx)error {
		profile, err := tx.GetProfile()
		if err != nil {
			return err
		}
		profile.Moderator = true

		if err := tx.SetProfile(profile); err != nil {
			return err
		}

		api, err := coreapi.NewCoreAPI(n.ipfsNode)
		if err != nil {
			return err
		}
		if err = api.Block().Rm(ctx, path.New(moderatorCid)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		maybeCloseDone(done)
		return err
	}
	n.Publish(done)
	return nil
}