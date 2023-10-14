package main

import (
	"context"
	"fmt"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
	"log"
	"os"
	"os/user"
	"time"
)

func clientGRPCAuth() grpc.DialOption {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	homeDir := usr.HomeDir
	lndDir := fmt.Sprintf("%s/Library/Application Support/Lnd", homeDir)
	fmt.Println(lndDir)
	macaroonFileLocation := fmt.Sprintf("%s/data/chain/bitcoin/regtest/admin.macaroon", lndDir)
	macBytes, err := os.ReadFile(macaroonFileLocation)
	if err != nil {
		log.Fatalf("unable to read macaroon path : %v", err)
	}

	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macBytes); err != nil {
		log.Fatalf("unable to decode macaroon: %v", err)
	}

	cred, err := macaroons.NewMacaroonCredential(mac)
	if err != nil {
		log.Fatalf("error creating macaroon credential: %v", err)
	}

	return grpc.WithPerRPCCredentials(cred)
}

func main() {
	clientConn := connectGRPC(clientGRPCAuth())
	defer clientConn.Close()

	lncli := lnrpc.NewLightningClient(clientConn)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	walletBalanceReq := lnrpc.WalletBalanceRequest{}
	walletRes, err := lncli.WalletBalance(ctx, &walletBalanceReq)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(walletRes.TotalBalance)
}

func connectGRPC(authOption grpc.DialOption) *grpc.ClientConn {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	homeDir := usr.HomeDir
	lndDir := fmt.Sprintf("%s/Library/Application Support/Lnd", homeDir)
	fmt.Println(lndDir)
	// SSL credentials setup
	var serverName string
	certFileLocation := fmt.Sprintf("%s/tls.cert", lndDir)
	creds, err := credentials.NewClientTLSFromFile(certFileLocation, serverName)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(creds)

	conn, err := grpc.Dial("localhost:10009", authOption, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}
	return conn
}
