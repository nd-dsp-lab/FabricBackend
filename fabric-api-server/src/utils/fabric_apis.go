package utils

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID       = "Org1MSP"
	gatewayPeer = "peer0.org1.example.com"
)

var (
	cryptoPath   string
	certPath     string
	keyPath      string
	tlsCertPath  string
	peerEndpoint string
)

func init() {
	// Get crypto path from environment or use default
	cryptoPath = os.Getenv("CRYPTO_PATH")
	if cryptoPath == "" {
		cryptoPath = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com"
	}

	// Build paths based on crypto path
	certPath = cryptoPath + "/users/User1@org1.example.com/msp/signcerts"
	keyPath = cryptoPath + "/users/User1@org1.example.com/msp/keystore"
	tlsCertPath = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"

	// Get peer endpoint from environment or use default
	peerEndpoint = os.Getenv("PEER_ENDPOINT")
	if peerEndpoint == "" {
		peerEndpoint = "localhost:7051"
	}
}

var (
	GatewayConn    *client.Gateway
	ClientConn     *grpc.ClientConn
	ClientContract *client.Contract
)

// InitGateway initializes the Gateway connection.
func InitGateway() {
	ClientConn = newGrpcConnection()

	id := newIdentity()
	sign := newSign()
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(ClientConn),
		client.WithEvaluateTimeout(10*time.Second),
		client.WithEndorseTimeout(30*time.Second),
		client.WithSubmitTimeout(10*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	GatewayConn = gw

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "basic"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)
	ClientContract = network.GetContract(chaincodeName)
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificatePEM, err := os.ReadFile(tlsCertPath)
	if err != nil {
		panic(fmt.Errorf("failed to read TLS certifcate file: %w", err))
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.NewClient(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificatePEM, err := readFirstFile(certPath)
	if err != nil {
		panic(fmt.Errorf("failed to read certificate file: %w", err))
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	privateKeyPEM, err := readFirstFile(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

func readFirstFile(dirPath string) ([]byte, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}

	fileNames, err := dir.Readdirnames(1)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(path.Join(dirPath, fileNames[0]))
}

// Format JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}

// =====================================================================================================================
// Drone Record APIs
// =====================================================================================================================

func InitLedgerWithExampleRecords() error {
	_, err := ClientContract.SubmitTransaction("InitLedgerWithExampleRecords")
	if err != nil {
		fmt.Printf("failed to submit InitLedgerWithExampleRecords transaction: %v\n", err)
		return err
	}
	return nil
}

func CreateRecord(recordID string, droneID string, pilotID string, zoneID string, recordType string, reserved string) error {
	_, err := ClientContract.SubmitTransaction("CreateRecord", recordID, droneID, pilotID, zoneID, recordType, reserved)
	if err != nil {
		fmt.Printf("failed to submit CreateRecord transaction, possibly the record already exists: %v\n", err)
		return err
	}
	return nil
}

func GetAllRecords() (string, error) {
	evaluateResult, err := ClientContract.EvaluateTransaction("GetAllRecords")
	if err != nil {
		fmt.Printf("failed to evaluate GetAllRecords transaction: %v\n", err)
		return "", err
	}
	// Handle empty results (no records found)
	if len(evaluateResult) == 0 {
		return "[]", nil
	}

	log.Printf("evaluateResult: %s", string(evaluateResult))
	result := formatJSON(evaluateResult)
	return result, nil
}

func UpdateRecord(recordID string, droneID string, pilotID string, zoneID string, recordType string, reserved string) error {
	_, err := ClientContract.SubmitTransaction("UpdateRecord", recordID, droneID, pilotID, zoneID, recordType, reserved)
	if err != nil {
		fmt.Printf("failed to submit UpdateRecord transaction: %v\n", err)
		return err
	}
	return nil
}

func GetRecordWithSelector(rawSelector string) (string, error) {
	// selector := map[string]interface{}{"selector": rawSelector}

	// selectorJSON, err := json.Marshal(rawSelector)
	// if err != nil {
	// 	fmt.Printf("failed to marshal selector: %v\n", err)
	// 	return "", err
	// }

	// log.Printf("selectorJSON: %s", string(selectorJSON))

	evaluateResult, err := ClientContract.EvaluateTransaction("QueryRecords", rawSelector)
	if err != nil {
		fmt.Printf("failed to evaluate QueryRecords transaction: %v\n", err)
		return "", err
	}

	// Handle empty results (no records found)
	if len(evaluateResult) == 0 {
		return "[]", nil
	}

	result := formatJSON(evaluateResult)
	return result, nil
}

func GetRecordHistory(recordID string) (string, error) {
	evaluateResult, err := ClientContract.EvaluateTransaction("GetRecordHistory", recordID)
	if err != nil {
		fmt.Printf("failed to evaluate GetRecordHistory transaction: %v\n", err)
		return "", err
	}

	result := formatJSON(evaluateResult)
	return result, nil
}
