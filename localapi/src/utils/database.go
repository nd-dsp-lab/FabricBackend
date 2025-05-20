package utils

import (
	"context"
	"fmt"
	"log"

	kivik "github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb" // The CouchDB driver
)

// PutCertificateToDB stores the certificate in the database
// makes sure that the CertificateID is unique
func PutCertificateToDB(db *kivik.DB, certDBO *CertificateDBObject) error {
	// Create a new document in the database
	_, err := db.Put(context.Background(), certDBO.CertificateID, certDBO)
	if err != nil {
		return fmt.Errorf("failed to store certificate: %w", err)
	}
	return nil
}

// GetCertificateWithCertID retrieves the certificate from the database
func GetCertificateWithCertID(db *kivik.DB, certID string) (*CertificateDBObject, error) {
	// Get the document from the database
	var certDBO CertificateDBObject
	err := db.Get(context.Background(), certID).ScanDoc(&certDBO)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve certificate: %w", err)
	}
	return &certDBO, nil
}

// GetCertificatesWithSelector retrieves the certificate from the database
func GetCertificatesWithSelector(db *kivik.DB, rawSelector map[string]interface{}) ([]CertificateDBObject, error) {
	selector := map[string]interface{}{
		"selector": rawSelector,
	}
	log.Printf("Selector: %v", selector)
	rows := db.Find(context.Background(), selector)
	defer rows.Close()

	var certDBO []CertificateDBObject
	for rows.Next() {
		var cert CertificateDBObject
		if err := rows.ScanDoc(&cert); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		certDBO = append(certDBO, cert)
	}
	return certDBO, nil
}
