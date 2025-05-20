package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
)

type CertificateDBObject struct {
	PilotID        string `json:"pilot_id"`
	CertificateID  string `json:"certificate_id"`
	DroneID        string `json:"drone_id"`
	ExpirationDate string `json:"expiration_date"`
	CertificateContent    *Certificate `json:"certificate_content"`
	SerializedCertificate string `json:"serialized_certificate"`
}

// 
// get a random base64 string
func GetRandomString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
