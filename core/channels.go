package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/cpacia/openbazaar3.0/channels"
	"github.com/cpacia/openbazaar3.0/core/coreiface"
	"github.com/cpacia/openbazaar3.0/database"
	"github.com/cpacia/openbazaar3.0/events"
	"github.com/cpacia/openbazaar3.0/models"
	"github.com/cpacia/openbazaar3.0/net/pb"
	"github.com/golang/protobuf/ptypes"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
)

// OpenChannel opens a new chat channel and subscribes.
func (n *OpenBazaarNode) OpenChannel(topic string) error {
	_, ok := n.channels[topic]
	if ok {
		return fmt.Errorf("%w: channel already open", coreiface.ErrBadRequest)
	}

	ch, err := channels.NewChannel(topic, n.ipfsNode, n.networkService, n.eventBus, n.repo.DB())
	if err != nil {
		return fmt.Errorf("%w: %s", coreiface.ErrInternalServer, err)
	}

	n.channels[topic] = ch
	return nil
}

// CloseChannel closes the chat channel and unsubscribes.
func (n *OpenBazaarNode) CloseChannel(topic string) error {
	ch, ok := n.channels[topic]
	if !ok {
		return fmt.Errorf("%w: channel not open", coreiface.ErrBadRequest)
	}
	ch.Close()
	return nil
}

// PublishChannelMessage publishes a message to the given channel.
func (n *OpenBazaarNode) PublishChannelMessage(ctx context.Context, topic, message string) error {
	ch, ok := n.channels[topic]
	if !ok {
		return fmt.Errorf("%w: channel not open", coreiface.ErrBadRequest)
	}
	return ch.Publish(ctx, message)
}

// GetChannelMessages returns the messages in the channel.
func (n *OpenBazaarNode) GetChannelMessages(ctx context.Context, topic string, from *cid.Cid, limit int) ([]models.ChannelMessage, error) {
	ch, ok := n.channels[topic]
	if !ok {
		return nil, fmt.Errorf("%w: channel not open", coreiface.ErrBadRequest)
	}
	return ch.Messages(ctx, from, limit)
}

// handleChannelRequest is the handler for the CHANNEL_REQUEST message. It responds to
// request with an CHANNEL_RESPONSE message using an online message.
func (n *OpenBazaarNode) handleChannelRequest(from peer.ID, message *pb.Message) error {
	if message.MessageType != pb.Message_CHANNEL_REQUEST {
		return errors.New("message is not type CHANNEL_REQUEST")
	}

	req := new(pb.ChannelRequestMessage)
	if err := ptypes.UnmarshalAny(message.Payload, req); err != nil {
		return err
	}

	var channelRec models.Channel
	err := n.repo.DB().View(func(tx database.Tx) error {
		return tx.Read().Where("topic=?", req.Topic).Error
	})
	if err != nil {
		return err
	}

	ids, err := channelRec.GetHead()
	if err != nil {
		return err
	}

	cidBytes := make([][]byte, 0, len(ids))
	for _, id := range ids {
		cidBytes = append(cidBytes, id.Bytes())
	}

	channelResp := pb.ChannelResponseMessage{
		Topic: req.Topic,
		Cids:  cidBytes,
	}

	payload, err := ptypes.MarshalAny(&channelResp)
	if err != nil {
		return err
	}

	resp := newMessageWithID()
	resp.MessageType = pb.Message_CHANNEL_RESPONSE
	resp.Payload = payload

	return n.networkService.SendMessage(context.Background(), from, resp)
}

// handleChannelResponse is the handler for the CHANNEL_RESPONSE message. It pushes
// the response to the event bus for any listening subscribers.
func (n *OpenBazaarNode) handleChannelResponse(from peer.ID, message *pb.Message) error {
	if message.MessageType != pb.Message_CHANNEL_RESPONSE {
		return errors.New("message is not type CHANNEL_RESPONSE")
	}

	resp := new(pb.ChannelResponseMessage)
	if err := ptypes.UnmarshalAny(message.Payload, resp); err != nil {
		return err
	}

	ids := make([]cid.Cid, 0, len(resp.Cids))
	for _, idBytes := range resp.Cids {
		_, id, err := cid.CidFromBytes(idBytes)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	n.eventBus.Emit(&events.ChannelRequestResponse{
		PeerID: from.Pretty(),
		Topic:  resp.Topic,
		Cids:   ids,
	})
	return nil
}
