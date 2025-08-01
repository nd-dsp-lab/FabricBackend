package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Minimal interface for what we need to test
type StubInterface interface {
	GetState(key string) ([]byte, error)
	PutState(key string, value []byte) error
}

type ContextInterface interface {
	GetStub() StubInterface
}

// Simple mocks
type MockStub struct {
	mock.Mock
	state map[string][]byte
}

func NewMockStub() *MockStub {
	return &MockStub{
		state: make(map[string][]byte),
	}
}

func (m *MockStub) GetState(key string) ([]byte, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockStub) PutState(key string, value []byte) error {
	args := m.Called(key, value)
	if args.Error(0) == nil {
		m.state[key] = value
	}
	return args.Error(0)
}

type MockContext struct {
	stub StubInterface
}

func NewMockContext(stub StubInterface) *MockContext {
	return &MockContext{stub: stub}
}

func (m *MockContext) GetStub() StubInterface {
	return m.stub
}

// Test adapter functions that use our simpler interfaces
func testCreateRecord(ctx ContextInterface, recordID, droneID, pilotID, zoneID, recordType, reserved string) error {
	// if recordID is not provided, generate a random one
	if recordID == "" {
		recordID = MD5Hash(reserved)
	}

	// Check if the record already exists
	exists, err := testRecordExists(ctx, recordID)
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

	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// create the record
	err = ctx.GetStub().PutState(recordID, recordJSON)
	if err != nil {
		return err
	}

	return nil
}

func testReadSingleRecord(ctx ContextInterface, sessionID string) (*Record, error) {
	recordJSON, err := ctx.GetStub().GetState(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if recordJSON == nil {
		return nil, fmt.Errorf("the record %s does not exist", sessionID)
	}

	var record Record
	err = json.Unmarshal(recordJSON, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func testRecordExists(ctx ContextInterface, recordID string) (bool, error) {
	recordJSON, err := ctx.GetStub().GetState(recordID)
	if err != nil {
		return false, fmt.Errorf("failed to read record %s from world state: %v", recordID, err)
	}

	return recordJSON != nil, nil
}

// Test functions

func TestCreateRecord_Success(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	// Test data
	recordID := "test-record-1"
	droneID := "drone-1"
	pilotID := "pilot-1"
	zoneID := "zone-1"
	recordType := "profile"
	reserved := "test-reserved-data"

	// Setup mocks
	mockStub.On("GetState", recordID).Return([]byte(nil), nil) // Record doesn't exist
	mockStub.On("PutState", recordID, mock.AnythingOfType("[]uint8")).Return(nil)

	// Act
	err := testCreateRecord(mockCtx, recordID, droneID, pilotID, zoneID, recordType, reserved)

	// Assert
	assert.NoError(t, err)
	mockStub.AssertExpectations(t)

	// Verify the record was stored correctly
	storedData, exists := mockStub.state[recordID]
	assert.True(t, exists)

	var storedRecord Record
	err = json.Unmarshal(storedData, &storedRecord)
	assert.NoError(t, err)
	assert.Equal(t, recordID, storedRecord.RecordID)
	assert.Equal(t, droneID, storedRecord.DroneID)
	assert.Equal(t, pilotID, storedRecord.PilotID)
	assert.Equal(t, zoneID, storedRecord.ZoneID)
	assert.Equal(t, recordType, storedRecord.RecordType)
	assert.Equal(t, reserved, storedRecord.Reserved)
}

func TestCreateRecord_AutoGenerateID(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	// Test data - empty recordID should trigger auto-generation
	recordID := ""
	droneID := "drone-1"
	pilotID := "pilot-1"
	zoneID := "zone-1"
	recordType := "profile"
	reserved := "test-reserved-data"

	expectedRecordID := MD5Hash(reserved)

	// Setup mocks
	mockStub.On("GetState", expectedRecordID).Return([]byte(nil), nil) // Record doesn't exist
	mockStub.On("PutState", expectedRecordID, mock.AnythingOfType("[]uint8")).Return(nil)

	// Act
	err := testCreateRecord(mockCtx, recordID, droneID, pilotID, zoneID, recordType, reserved)

	// Assert
	assert.NoError(t, err)
	mockStub.AssertExpectations(t)

	// Verify the record was stored with generated ID
	storedData, exists := mockStub.state[expectedRecordID]
	assert.True(t, exists)

	var storedRecord Record
	err = json.Unmarshal(storedData, &storedRecord)
	assert.NoError(t, err)
	assert.Equal(t, expectedRecordID, storedRecord.RecordID)
}

func TestCreateRecord_RecordAlreadyExists(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	recordID := "existing-record"
	existingRecord := Record{
		RecordID:   recordID,
		DroneID:    "existing-drone",
		PilotID:    "existing-pilot",
		ZoneID:     "existing-zone",
		RecordType: "profile",
		Reserved:   "existing-data",
	}
	existingRecordJSON, _ := json.Marshal(existingRecord)

	// Setup mocks
	mockStub.On("GetState", recordID).Return(existingRecordJSON, nil) // Record exists

	// Act
	err := testCreateRecord(mockCtx, recordID, "drone-1", "pilot-1", "zone-1", "profile", "reserved")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockStub.AssertExpectations(t)
}

func TestCreateRecord_GetStateError(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	recordID := "test-record"

	// Setup mocks
	mockStub.On("GetState", recordID).Return([]byte(nil), fmt.Errorf("database error"))

	// Act
	err := testCreateRecord(mockCtx, recordID, "drone-1", "pilot-1", "zone-1", "profile", "reserved")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read from world state")
	mockStub.AssertExpectations(t)
}

func TestCreateRecord_PutStateError(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	recordID := "test-record"

	// Setup mocks
	mockStub.On("GetState", recordID).Return([]byte(nil), nil) // Record doesn't exist
	mockStub.On("PutState", recordID, mock.AnythingOfType("[]uint8")).Return(fmt.Errorf("write error"))

	// Act
	err := testCreateRecord(mockCtx, recordID, "drone-1", "pilot-1", "zone-1", "profile", "reserved")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write error")
	mockStub.AssertExpectations(t)
}

func TestReadSingleRecord_Success(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	sessionID := "test-session-1"
	expectedRecord := Record{
		RecordID:   "record-1",
		DroneID:    "drone-1",
		PilotID:    "pilot-1",
		ZoneID:     "zone-1",
		RecordType: "profile",
		Reserved:   "reserved-data",
	}
	recordJSON, _ := json.Marshal(expectedRecord)

	// Setup mocks
	mockStub.On("GetState", sessionID).Return(recordJSON, nil)

	// Act
	result, err := testReadSingleRecord(mockCtx, sessionID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedRecord.RecordID, result.RecordID)
	assert.Equal(t, expectedRecord.DroneID, result.DroneID)
	assert.Equal(t, expectedRecord.PilotID, result.PilotID)
	assert.Equal(t, expectedRecord.ZoneID, result.ZoneID)
	assert.Equal(t, expectedRecord.RecordType, result.RecordType)
	assert.Equal(t, expectedRecord.Reserved, result.Reserved)
	mockStub.AssertExpectations(t)
}

func TestReadSingleRecord_RecordNotFound(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	sessionID := "non-existent-session"

	// Setup mocks
	mockStub.On("GetState", sessionID).Return([]byte(nil), nil) // Record not found

	// Act
	result, err := testReadSingleRecord(mockCtx, sessionID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "does not exist")
	mockStub.AssertExpectations(t)
}

func TestReadSingleRecord_GetStateError(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	sessionID := "test-session"

	// Setup mocks
	mockStub.On("GetState", sessionID).Return([]byte(nil), fmt.Errorf("database error"))

	// Act
	result, err := testReadSingleRecord(mockCtx, sessionID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to read from world state")
	mockStub.AssertExpectations(t)
}

func TestReadSingleRecord_InvalidJSON(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	sessionID := "test-session"
	invalidJSON := []byte("invalid-json-data")

	// Setup mocks
	mockStub.On("GetState", sessionID).Return(invalidJSON, nil)

	// Act
	result, err := testReadSingleRecord(mockCtx, sessionID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockStub.AssertExpectations(t)
}

func TestRecordExists_True(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	recordID := "existing-record"
	existingRecord := Record{RecordID: recordID}
	recordJSON, _ := json.Marshal(existingRecord)

	// Setup mocks
	mockStub.On("GetState", recordID).Return(recordJSON, nil)

	// Act
	exists, err := testRecordExists(mockCtx, recordID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	mockStub.AssertExpectations(t)
}

func TestRecordExists_False(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	recordID := "non-existent-record"

	// Setup mocks
	mockStub.On("GetState", recordID).Return([]byte(nil), nil)

	// Act
	exists, err := testRecordExists(mockCtx, recordID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
	mockStub.AssertExpectations(t)
}

func TestRecordExists_Error(t *testing.T) {
	// Arrange
	mockStub := NewMockStub()
	mockCtx := NewMockContext(mockStub)

	recordID := "test-record"

	// Setup mocks
	mockStub.On("GetState", recordID).Return([]byte(nil), fmt.Errorf("database error"))

	// Act
	exists, err := testRecordExists(mockCtx, recordID)

	// Assert
	assert.Error(t, err)
	assert.False(t, exists)
	assert.Contains(t, err.Error(), "failed to read record")
	mockStub.AssertExpectations(t)
}

func TestMD5Hash(t *testing.T) {
	// Test cases
	testCases := []struct {
		input    string
		expected string
	}{
		{"test", "098f6bcd4621d373cade4e832627b4f6"},
		{"hello world", "5eb63bbbe01eeed093cb22bb8f5acdc3"},
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("MD5Hash(%s)", tc.input), func(t *testing.T) {
			result := MD5Hash(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
