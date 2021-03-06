/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
//
package comm

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric/gossip/api"
	"github.com/hyperledger/fabric/gossip/common"
	proto "github.com/hyperledger/fabric/protos/gossip"
)

// Comm is an object that enables to communicate with other peers
// that also embed a CommModule.
type Comm interface {

	// GetPKIid returns this instance's PKI id
	//返回PKI_ID,即PeerID
	GetPKIid() common.PKIidType

	// Send sends a message to remote peers
	//向节点发送消息
	Send(msg *proto.SignedGossipMessage, peers ...*RemotePeer)

	// SendWithAck sends a message to remote peers, waiting for acknowledgement from minAck of them, or until a certain timeout expires
	//向节点发送信息，有延迟等待远程节点响应
	SendWithAck(msg *proto.SignedGossipMessage, timeout time.Duration, minAck int, peers ...*RemotePeer) AggregatedSendResult

	// Probe probes a remote node and returns nil if its responsive,
	// and an error if it's not.
	//探测远程节点是否有响应
	Probe(peer *RemotePeer) error

	// Handshake authenticates a remote peer and returns
	// (its identity, nil) on success and (nil, error)
	//握手远程节点是否成功连接
	Handshake(peer *RemotePeer) (api.PeerIdentityType, error)

	// Accept returns a dedicated read-only channel for messages sent by other nodes that match a certain predicate.
	// Each message from the channel can be used to send a reply back to the sender
	//接收消息
	Accept(common.MessageAcceptor) <-chan proto.ReceivedMessage

	// PresumedDead returns a read-only channel for node endpoints that are suspected to be offline
	//获取可疑离线节点的只读通道
	PresumedDead() <-chan common.PKIidType

	// CloseConn closes a connection to a certain endpoint
	//关闭连接
	CloseConn(peer *RemotePeer)

	// Stop stops the module
	Stop()
}

// RemotePeer defines a peer's endpoint and its PKIid
//远程节点结构体
type RemotePeer struct {
	Endpoint string
	PKIID    common.PKIidType
}

// SendResult defines a result of a send to a remote peer
//向远程节点的发送结果
type SendResult struct {
	error
	RemotePeer
}

// Error returns the error of the SendResult, or an empty string
// if an error hasn't occurred
func (sr SendResult) Error() string {
	if sr.error != nil {
		return sr.error.Error()
	}
	return ""
}

// AggregatedSendResult represents a slice of SendResults
//提供切片数据封装
type AggregatedSendResult []SendResult

// AckCount returns the number of successful acknowledgements
func (ar AggregatedSendResult) AckCount() int {
	c := 0
	for _, ack := range ar {
		if ack.error == nil {
			c++
		}
	}
	return c
}

// NackCount returns the number of unsuccessful acknowledgements
func (ar AggregatedSendResult) NackCount() int {
	return len(ar) - ar.AckCount()
}

// String returns a JSONed string representation
// of the AggregatedSendResult
func (ar AggregatedSendResult) String() string {
	errMap := map[string]int{}
	for _, ack := range ar {
		if ack.error == nil {
			continue
		}
		errMap[ack.Error()]++
	}

	ackCount := ar.AckCount()
	output := map[string]interface{}{}
	if ackCount > 0 {
		output["successes"] = ackCount
	}
	if ackCount < len(ar) {
		output["failures"] = errMap
	}
	b, _ := json.Marshal(output)
	return string(b)
}

// String converts a RemotePeer to a string
//数据类型转换
func (p *RemotePeer) String() string {
	return fmt.Sprintf("%s, PKIid:%v", p.Endpoint, p.PKIID)
}
