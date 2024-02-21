package chord

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"reflect"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
Explanations on Chord Configurations:
- NodeID: the identifier of the node. This is mostly used
  for the client, and is not used in the protocol.
- Address: IP address and port (as a single string) to denote
  the address to which the Chord service will listen to.
- Hash: Chord is a consistent hashing protocol. The hash
  function used must be consistent throughout the cluster.
- JoinAddr: To allow nodes to join to an exisiting Chord cluster,
  a join address must be specified.
- NumSuccessor: To support fault tolerance on Chord, successors
  must be tracked. The higher the number, the more durable the system.
- MaxIdle: Timer to garbage collect idle gRPC connections.
- Stabilization: Timer for Chord's stabilization logic.
- FingerFix: Timer for Chord's finger fixing logic.
- Server: gRPC server. This can be provided by the user if there is
  a need to multiplex the gRPC server on multiple services.
- Timeout: Timeouts for gRPC client responses.
- ServerOptions: gRPC server options.
- DialOptions: gRPC dial options (for client).
*/

// Configuration for Chord node. The hash function
// should be homogenous for a single cluster.
// Idealy, the entire configuration should be homogenous.
// JoinAddr may be an empty string when initially
// creating the cluster.
type Config struct {
	// Chord Settings
	NodeID       string
	Address      string
	Hash         func() hash.Hash
	JoinAddr     string
	NumSuccessor int
	// Chord Background Process
	MaxIdle       time.Duration
	Stabilization time.Duration
	FingerFix     time.Duration
	// Chord gRPC Settings
	Server        *grpc.Server
	Timeout       time.Duration
	ServerOptions []grpc.ServerOption
	DialOptions   []grpc.DialOption
}

// Default Configurations. We default the NodeID to the IP address.
func DefaultConfig(address string, joinaddr string, server *grpc.Server) *Config {
	return &Config{
		NodeID:        address,
		Address:       address,
		Hash:          sha1.New,
		JoinAddr:      joinaddr,
		NumSuccessor:  3,
		MaxIdle:       8 * time.Second,
		Stabilization: time.Second,
		FingerFix:     400 * time.Millisecond,
		Server:        server,
		Timeout:       3 * time.Second,
		ServerOptions: nil,
		DialOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		},
	}
}

// Used to compare hash functions of Chord instances.
// A more robust method might be useful in the future,
// but for now we simply return the name of the hash
// function and compare string values.
func (c *Config) HashName() string {
	// for instance, for outputs string "crypto/sha1.New"
	return runtime.FuncForPC(reflect.ValueOf(c.Hash).Pointer()).Name()
}

// Validate Configurations.
func ValidateConfig(config *Config) error {
	// TODO: need a better way to identify ip address and port
	if len(config.Address) == 0 {
		return fmt.Errorf("chord instance ip address should not be empty")
	}
	if config.NumSuccessor < 1 {
		return fmt.Errorf("number of successors kept track should be positive")
	}
	if config.Stabilization > config.MaxIdle || config.FingerFix > config.MaxIdle {
		return fmt.Errorf("maxidle parameter should be the larger than stabilization and fingerfix routines")
	}
	if config.Server != nil {
		return fmt.Errorf("chord grpc multiplexing is not supported yet")
	}
	return nil
}
