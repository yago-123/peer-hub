package main

import (
	"context"
	"encoding/base64"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/yago-123/peer-hub/pkg/types"

	"github.com/yago-123/peer-hub/pkg/client"

	"github.com/yago-123/peer-hub/pkg/util"

	// todo(): make it WireGuard agnostic in the future
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	RendezvousServerAddr = "http://rendezvous.yago.ninja:7777"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stunServers := []string{
		"stun.l.google.com:19302",
		"stun1.l.google.com:19302",
	}

	// Get public endpoint using STUN servers
	endpoint, err := util.GetPublicEndpoint(ctx, stunServers)
	if err != nil {
		logger.Error("failed to get public endpoint", err)
		return
	}

	err = client.Register(ctx, types.RegisterRequest{
		PeerID:     "peer-a",
		PublicKey:  base64.StdEncoding.EncodeToString(pubKey[:]),
		AllowedIPs: []string{"10.0.0.2/32"},
		Endpoint:   endpoint.String(),
	})
	if err != nil {
		logger.Error("registration of peer failed", err, "peer-id", "peer-a")
		return
	}

	logger.Info("Registered successfully")

	// Discover remote peer
	resp, udpAddr, err := client.Discover(ctx, "peer-a")
	if err != nil {
		logger.Error("discovery failed", err)
		return
	}

	logger.Info("Discovered peer", "peer-id", resp.PeerID, "public-key", resp.PublicKey, "endpoint", udpAddr, "allowed-ips", resp.AllowedIPs)
}
