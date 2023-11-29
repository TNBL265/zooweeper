// Package ztree implements the Replicated Database component for our ZooWeeper.
//
// 1. Instead of using the filesystem, we implemented the data model and hierarchical namespace using sqlite
// - each row is a ZNode storing Metadata
// - the field ParentId will represent the hierarchical relationship
// 2. (Use-case specific) We only support Regular (permanent) ZNode, no Ephemeral ZNode
// 3. Metadata fields in ZNode:
// - NodeId (int): similar to zxid, representing metadata transaction (1st NodeId is a self-identified by design)
// - NodePort (string): the port of the current ZooWeeper server (808x by design)
// - Leader (string): the port  of the current leader in the ensemble (highest NodePort by design)
// - Servers (string): comma-separated list of the ports of all ZooWeeper servers in the ensemble
// - Timestamp (string): timestamp at which this ZNode is created
// - Version (int): keep track of ZNode changes
// - ParentId (int): NodeId of parent ZNode
// - Clients (string): (Use-case specific) comma-separated list of the ports of all clients (Kafka-Server) that use our ZooWeeper service
// - SenderIp (string): (Use-case specific) the port of the client (Kafka-Server) that sent the Write Request
// - ReceiverIp (string): (Use-case specific) the port of the ZooWeeper server that the client (Kafka-Server) chose to send the Write Request to
//
// Reference: https://zookeeper.apache.org/doc/current/zookeeperOver.html

package ztree

import (
	"database/sql"
)

type ZTree struct {
	DB *sql.DB
}

type ZNodeHandlers interface {
	// Utils
	Connection() *sql.DB
	InitializeDB()

	// Getter
	AllMetadata() ([]*Metadata, error)
	ZNodeIdExists(nodeId int) (bool, error)
	GetHighestZNodeId() (int, error)
	GetLocalMetadata() (*Metadata, error)
	GetMetadatasGreaterThanZNodeId(highestZNodeId int) (Metadatas, error)
	GetClients(client string) ([]string, error)

	// Setter
	InsertFirstMetadata(metadata Metadata) error
	InsertMetadata(metadata Metadata) error
	InsertMetadataWithParent(metadata Metadata) error
	UpdateFirstLeader(Leader string) error
}
