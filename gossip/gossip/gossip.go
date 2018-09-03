/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gossip

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric/gossip/api"
	"github.com/hyperledger/fabric/gossip/comm"
	"github.com/hyperledger/fabric/gossip/common"
	"github.com/hyperledger/fabric/gossip/discovery"
	"github.com/hyperledger/fabric/gossip/filter"
	proto "github.com/hyperledger/fabric/protos/gossip"
)

// Gossip is the interface of the gossip component
//gossip组件接口定义
type Gossip interface {

	// Send sends a message to remote peers
	//向远程节点发送消息
	Send(msg *proto.GossipMessage, peers ...*comm.RemotePeer)

	// SendByCriteria sends a given message to all peers that match the given SendCriteria

	SendByCriteria(*proto.SignedGossipMessage, SendCriteria) error

	// GetPeers returns the NetworkMembers considered alive
	//获取活跃节点
	Peers() []discovery.NetworkMember

	// PeersOfChannel returns the NetworkMembers considered alive
	// and also subscribed to the channel given
	//返回指定通道的活跃节点
	PeersOfChannel(common.ChainID) []discovery.NetworkMember

	// UpdateMetadata updates the self metadata of the discovery layer
	// the peer publishes to other peers
	//更新元数据
	UpdateMetadata(metadata []byte)

	// UpdateChannelMetadata updates the self metadata the peer
	// publishes to other peers about its channel-related state
	//更新通道元数据
	UpdateChannelMetadata(metadata []byte, chainID common.ChainID)

	// Gossip sends a message to other peers to the network
	//向网络内其他节点发送消息
	Gossip(msg *proto.GossipMessage)

	// PeerFilter receives a SubChannelSelectionCriteria and returns a RoutingFilter that selects
	// only peer identities that match the given criteria, and that they published their channel participation
	//接收channel订阅，节点发布订阅通道
	PeerFilter(channel common.ChainID, messagePredicate api.SubChannelSelectionCriteria) (filter.RoutingFilter, error)

	// Accept returns a dedicated read-only channel for messages sent by other nodes that match a certain predicate.
	// If passThrough is false, the messages are processed by the gossip layer beforehand.
	// If passThrough is true, the gossip layer doesn't intervene and the messages
	// can be used to send a reply back to the sender
	//接收通道只读信息
	Accept(acceptor common.MessageAcceptor, passThrough bool) (<-chan *proto.GossipMessage, <-chan proto.ReceivedMessage)

	// JoinChan makes the Gossip instance join a channel
	//加入通道实例
	JoinChan(joinMsg api.JoinChannelMessage, chainID common.ChainID)

	// LeaveChan makes the Gossip instance leave a channel.
	// It still disseminates stateInfo message, but doesn't participate
	// in block pulling anymore, and can't return anymore a list of peers
	// in the channel.
	//离开通道
	LeaveChan(chainID common.ChainID)

	// SuspectPeers makes the gossip instance validate identities of suspected peers, and close
	// any connections to peers with identities that are found invalid
	//验证可疑节点身份，并关闭无效链接
	SuspectPeers(s api.PeerSuspector)

	// Stop stops the gossip component
	//停止gossip服务
	Stop()
}

// emittedGossipMessage encapsulates signed gossip message to compose
// with routing filter to be used while message is forwarded
type emittedGossipMessage struct {
	*proto.SignedGossipMessage
	filter func(id common.PKIidType) bool
}

// SendCriteria defines how to send a specific message
type SendCriteria struct {
	Timeout    time.Duration        // Timeout defines the time to wait for acknowledgements
	MinAck     int                  // MinAck defines the amount of peers to collect acknowledgements from
	MaxPeers   int                  // MaxPeers defines the maximum number of peers to send the message to
	IsEligible filter.RoutingFilter // IsEligible defines whether a specific peer is eligible of receiving the message
	Channel    common.ChainID       // Channel specifies a channel to send this message on. \
	// Only peers that joined the channel would receive this message
}

// String returns a string representation of this SendCriteria
func (sc SendCriteria) String() string {
	return fmt.Sprintf("channel: %s, tout: %v, minAck: %d, maxPeers: %d", sc.Channel, sc.Timeout, sc.MinAck, sc.MaxPeers)
}

// Config is the configuration of the gossip component
type Config struct {
	BindPort            int      // Port we bind to, used only for tests绑定测试端口
	ID                  string   // ID of this instance 本实例id，即对外开放的节点地址
	BootstrapPeers      []string // Peers we connect to at startup 启动时要链接的节点
	PropagateIterations int      // Number of times a message is pushed to remote peers 取自peer.gossip.propagateIterations，消息转发的次数，默认为1
	PropagatePeerNum    int      // Number of peers selected to push messages to 取自peer.gossip.propagatePeerNum，消息推送给节点的个数，默认为3

	MaxBlockCountToStore int // Maximum count of blocks we store in memory 取自peer.gossip.maxBlockCountToStore，保存到内存中的区块个数上限，默认100

	//取自peer.gossip.maxPropagationBurstSize，保存的最大消息个数，超过则转发给其他节点，默认为10
	MaxPropagationBurstSize    int           // Max number of messages stored until it triggers a push to remote peers

	//取自peer.gossip.maxPropagationBurstLatency，保存消息的最大时间，超过则转发给其他节点，默认为10毫秒
	MaxPropagationBurstLatency time.Duration // Max time between consecutive message pushes

	//取自peer.gossip.pullInterval，拉取消息的时间间隔，默认为4秒
	PullInterval time.Duration // Determines frequency of pull phases

	//取自peer.gossip.pullPeerNum，从指定个数的节点拉取信息，默认3个
	PullPeerNum  int           // Number of peers to pull from

	//取自peer.gossip.skipBlockVerification，是否不对区块消息进行校验，默认false即需校验
	SkipBlockVerification bool // Should we skip verifying block messages or not

	//取自peer.gossip.publishCertPeriod，包括在消息中的启动证书的时间，默认10s
	PublishCertPeriod        time.Duration // Time from startup certificates are included in Alive messages
	//取自peer.gossip.publishStateInfoInterval，向其他节点推送状态信息的时间间隔，默认4s
	PublishStateInfoInterval time.Duration // Determines frequency of pushing state info messages to peers
	//取自peer.gossip.requestStateInfoInterval，从其他节点拉取状态信息的时间间隔，默认4s
	RequestStateInfoInterval time.Duration // Determines frequency of pulling state info messages from peers

	//本节点TLS 证书
	TLSCerts *common.TLSCertificates // TLS certificates of the peer

	//本节点在组织内使用的地址
	InternalEndpoint string // Endpoint we publish to peers in our organization
	//本节点对外部组织公布的地址
	ExternalEndpoint string // Peer publishes this endpoint instead of SelfEndpoint to foreign organizations
}
