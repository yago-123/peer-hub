package main

import (
	"context"
	"encoding/base64"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/yago-123/peer-hub/pkg/common"
	"github.com/yago-123/peer-hub/pkg/types"

	"github.com/yago-123/peer-hub/pkg/client"

	"github.com/yago-123/peer-hub/pkg/util"

	// todo(): make it WireGuard agnostic in the future
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	RendezvousServerAddr = "http://rendezvous.yago.ninja:7777"
	registerTimeout      = 10 * time.Second
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Generate WireGuard keypair
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Fatalf("failed to generate private key: %v", err)
	}
	pubKey := privKey.PublicKey()

	logger.Info("Generated keys",
		"private key", base64.StdEncoding.EncodeToString(privKey[:]),
		"public key", base64.StdEncoding.EncodeToString(pubKey[:]))
	// Create rendezvous client
	client := client.New(RendezvousServerAddr, 1*time.Second)

	// Register this peer
	ctx, cancel := context.WithTimeout(context.Background(), registerTimeout)
	defer cancel()

	stunServers := []string{
		"stun.l.google.com:19302",
		"stun1.l.google.com:19302",
	}

	udpConn, err := net.ListenUDP(common.UDPProtocol, nil)
	if err != nil {
		logger.Error("failed to create UDP socket", "err", err)
		return
	}
	defer udpConn.Close()

	// Get public endpoint using STUN servers
	endpoint, err := util.GetPublicEndpoint(ctx, udpConn, stunServers)
	if err != nil {
		logger.Error("failed to get public endpoint", "err", err)
		return
	}

	err = client.Register(ctx, types.RegisterRequest{
		PeerID:     "peer-a",
		PublicKey:  base64.StdEncoding.EncodeToString(pubKey[:]),
		AllowedIPs: []string{"10.0.0.2/32"},
		Endpoint:   endpoint.String(),
	})
	if err != nil {
		logger.Error("registration of peer failed", "err", err, "peer-id", "peer-a")
		return
	}

	logger.Info("Registered successfully")

	// Discover remote peer
	resp, udpAddr, err := client.Discover(ctx, "peer-a")
	if err != nil {
		logger.Error("discovery failed", "err", err)
		return
	}

	logger.Info("Discovered peer", "peer-id", resp.PeerID, "public-key", resp.PublicKey, "endpoint", udpAddr, "allowed-ips", resp.AllowedIPs)
}
