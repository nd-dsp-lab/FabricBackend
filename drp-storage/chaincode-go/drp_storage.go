package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	// Add this import statement
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SimpleChaincode struct {
	contractapi.Contract
}

// Structure of the record, only the droneID, pilotID, zoneID, recordType and reserved are stored in the ledger for query
// The rest are serialized and encrypted in the reserved field
type Record struct {
	RecordID   string `json:"recordID"`
	DroneID    string `json:"droneID"`
	PilotID    string `json:"pilotID"`
	ZoneID     string `json:"zoneID"`
	RecordType string `json:"recordType"` // "profile" or "certificate"
	Reserved   string `json:"reserved"`
}

type HistoryQueryResult struct {
	Record    *Record   `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

type PaginatedQueryResult struct {
	Records             []Record `json:"records"`
	FetchedRecordsCount int32    `json:"fetchedRecordsCount"`
	Bookmark            string   `json:"bookmark"`
}

// Hello returns a greeting message to check if the chaincode is alive
func (s *SimpleChaincode) Hello(ctx contractapi.TransactionContextInterface) string {
	return "Hello from fabric, the service is running!"
}

// ReadRecord returns the record with the given recordID
func (s *SimpleChaincode) ReadRecord(ctx contractapi.TransactionContextInterface, recordID string) (*Record, error) {
	recordJSON, err := ctx.GetStub().GetState(recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if recordJSON == nil {
		return nil, fmt.Errorf("the record %s does not exist", recordID)
	}

	var record Record
	err = json.Unmarshal(recordJSON, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func MD5Hash(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	return hex.EncodeToString(hash.Sum(nil))
}

// CreateRecord adds a new record to the world state with given details
func (s *SimpleChaincode) CreateRecord(ctx contractapi.TransactionContextInterface, recordID string, droneID string, pilotID string, zoneID string, recordType string, reserved string) error {

	// if recordID is not provided, generate a random one
	if recordID == "" {
		recordID = MD5Hash(reserved)
	}

	// Check if the record already exists
	exists, err := s.RecordExists(ctx, recordID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if exists {
		return fmt.Errorf("the record %s already exists", recordID)
	}

	record := Record{
		RecordID:   recordID,
		DroneID:    droneID,
		PilotID:    pilotID,
		ZoneID:     zoneID,
		RecordType: recordType,
		Reserved:   reserved,
	}
	// print the record to be created
	fmt.Println("Record to be created", record)

	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// create the record
	err = ctx.GetStub().PutState(recordID, recordJSON)
	if err != nil {
		fmt.Println("Error in creating record for", recordID, err)
		return err
	}

	return nil
}

// InitLedger adds a base set of records to the ledger, not used in the current implementation
func (s *SimpleChaincode) InitLedgerWithExampleRecords(ctx contractapi.TransactionContextInterface) error {
	records := []Record{
		{RecordID: "record1", DroneID: "drone1", PilotID: "pilot1", ZoneID: "zone1", RecordType: "profile", Reserved: "reserved1"},
		{RecordID: "record2", DroneID: "drone2", PilotID: "pilot2", ZoneID: "zone2", RecordType: "profile", Reserved: "reserved2"},
		{RecordID: "record3", DroneID: "drone3", PilotID: "pilot3", ZoneID: "zone3", RecordType: "profile", Reserved: "reserved3"},
		{RecordID: "record4", DroneID: "drone4", PilotID: "pilot4", ZoneID: "zone4", RecordType: "profile", Reserved: "reserved4"},
		{RecordID: "record5", DroneID: "drone5", PilotID: "pilot5", ZoneID: "zone5", RecordType: "profile", Reserved: "reserved5"},
	}

	// records, _ := importFromFile()

	for _, record := range records {
		err := s.CreateRecord(ctx, record.RecordID, record.DroneID, record.PilotID, record.ZoneID, record.RecordType, record.Reserved)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAllRecords returns all records found in world state
func (s *SimpleChaincode) GetAllRecords(ctx contractapi.TransactionContextInterface) ([]Record, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []Record
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var record Record
		err = json.Unmarshal(queryResponse.Value, &record)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// RecordExists returns true when record with given ID exists in world state
func (s *SimpleChaincode) RecordExists(ctx contractapi.TransactionContextInterface, recordID string) (bool, error) {
	recordJSON, err := ctx.GetStub().GetState(recordID)
	if err != nil {
		return false, fmt.Errorf("failed to read record %s from world state: %v", recordID, err)
	}

	return recordJSON != nil, nil
}

// constructQueryResponseFromIterator constructs a slices of Records from QueryResultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Record, error) {
	var records []*Record
	for resultsIterator.HasNext() {
		recordResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var record Record
		err = json.Unmarshal(recordResponse.Value, &record)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	return records, nil
}

// GetRecordByRange performs a range query based on the start and end keys provided.
func (s *SimpleChaincode) GetRecordByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]*Record, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// getQueryResultForQueryString queries for records based on a passed in query string.
// This is only supported for couchdb
func (s *SimpleChaincode) getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Record, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// QueryRecords uses a query string to perform a query for records.
func (s *SimpleChaincode) QueryRecords(ctx contractapi.TransactionContextInterface, queryString string) ([]*Record, error) {
	return s.getQueryResultForQueryString(ctx, queryString)
}

// QueryRecordsByDroneID queries for records based on a passed in droneID.
func (s *SimpleChaincode) QueryRecordsByDroneID(ctx contractapi.TransactionContextInterface, droneID string) ([]*Record, error) {
	queryString := fmt.Sprintf(`{"selector":{"droneID":"%s"}}`, droneID)
	return s.getQueryResultForQueryString(ctx, queryString)
}

// QueryRecordsByPilotID queries for records based on a passed in pilotID.
func (s *SimpleChaincode) QueryRecordsByPilotID(ctx contractapi.TransactionContextInterface, pilotID string) ([]*Record, error) {
	queryString := fmt.Sprintf(`{"selector":{"pilotID":"%s"}}`, pilotID)
	return s.getQueryResultForQueryString(ctx, queryString)
}

// GetRecordHistory returns the history of records for a given recordID.
func (s *SimpleChaincode) GetRecordHistory(ctx contractapi.TransactionContextInterface, recordID string) ([]HistoryQueryResult, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(recordID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var historyRecords []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var record Record
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &record)
			if err != nil {
				return nil, err
			}
		} else {
			record = Record{
				RecordID: recordID,
			}
		}

		historyRecord := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: response.Timestamp.AsTime(),
			Record:    &record,
			IsDelete:  response.IsDelete,
		}
		historyRecords = append(historyRecords, historyRecord)
	}
	return historyRecords, nil
}

// update record
func (s *SimpleChaincode) UpdateRecord(ctx contractapi.TransactionContextInterface, recordID string, droneID string, pilotID string, zoneID string, recordType string, reserved string) error {

	record, err := s.ReadRecord(ctx, recordID)
	if err != nil {
		return fmt.Errorf("failed to read record %s from world state: %v", recordID, err)
	}

	// if the field is not provided, do not update it
	if droneID != "" {
		record.DroneID = droneID
	}
	if pilotID != "" {
		record.PilotID = pilotID
	}
	if zoneID != "" {
		record.ZoneID = zoneID
	}
	if recordType != "" {
		record.RecordType = recordType
	}
	if reserved != "" {
		record.Reserved = reserved
	}

	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(recordID, recordJSON)
	if err != nil {
		return fmt.Errorf("failed to update record %s in world state: %v", recordID, err)
	}

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SimpleChaincode{})
	if err != nil {
		log.Panicf("Error creating asset chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting asset chaincode: %v", err)
	}
}
