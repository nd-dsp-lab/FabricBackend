package utils

import (
	"encoding/json"
	"strconv"
	"time"
)

// The Certificate struct represents the structure of a certificate
// that is used in the application. It contains various fields
// related to the drone, pilot, and certificate details.
// This structure can be changed to suit the needs of the application.
// 
// PilotID, DroneID, and ExpirationDate are mandatory fields, because
// they are used to identify the certificate and its validity.
// Eventually the certificate will be stored as follow:
// type CertificateDBObject struct {
// 	PilotID        string `json:"pilot_id"`
// 	CertificateID  string `json:"certificate_id"`
// 	DroneID        string `json:"drone_id"`
// 	ExpirationDate string `json:"expiration_date"`
// 	SerializedCertificate string `json:"serialized_certificate"`
// }
// This structure can be changed in handlers.go, but it requires more modifications
// in the code, so it is better to keep it as it is for now.

// The exact format of the Certificate struct can be anything.
// The POST request needs to include the DroneID, PilotID, and ExpirationDate fields.
// The ExpirationDate field is a string in ISO8601 format.
// The SerializedCertificate field is a string that contains the serialized certificate.
type Certificate struct {
	DroneID                string `json:"drone_id" default:"Drone01"`
	PilotID                string `json:"pilot_id" default:"Pilot01"`
	ExpirationDate         string `json:"expirationDate" default:"2025-07-11T15:26:19.316Z"`
	SerializedCertificate  string `json:"serialized_certificate" default:"<Serialized Certificate>"`
}

// SerializedCertificate example:
// 
// The original certificate is a JSON object that contains the following fields:
// type RealCertificate struct {
// 	DroneID                string `json:"drone_id" default:"Drone01"`
// 	PilotID                string `json:"pilot_id" default:"Pilot01"`
// 	ReturnToLaunchPlan     string `json:"RETURN_TO_LAUNCH_PLAN" default:"true"`
// 	OperateWithinWindspeed string `json:"OPERATE_WITHIN_WINDSPEED" default:"true"`
// 	MonitorWindConditions  string `json:"MONITOR_WIND_CONDITIONS" default:"true"`
// 	SeenWeatherForecast    string `json:"SEEN_WEATHER_FORECAST" default:"true"`
// 	SustainedWinds         string `json:"SUSTAINED_WINDS" default:"1"`
// 	SustainedWindsPercent  string `json:"SUSTAINED_WINDS_PERCENT" default:"1"`
// 	MaxWindspeed           string `json:"MAX_WINDSPEED" default:"1"`
// 	ExpirationDate         string `json:"expirationDate" default:"2025-07-11T15:26:19.316Z"`
// 	CertificateName        string `json:"certificate_name" default:"WIND_FEATURE_MODEL"`
// 	SubmittedBy            string `json:"submitted_by" default:"jakolker"`
// 	SubmitterName          string `json:"submitter_name" default:"Jack Kolker"`
// 	CurrentDate            string `json:"current_date" default:"2025-05-02"`
// 	SvgFile                string `json:"svg_file" default:"<SVG Content>"`
// 	Passed                 string `json:"passed" default:"false"`
// 	Archived               string `json:"archived" default:"false"`
// }
// 
// The serialized certificate is a string that looks like this:
// "{\"drone_id\":\"Drone02\",\"pilot_id\":\"Pilot02\",\"RETURN_TO_LAUNCH_PLAN\":true,\"OPERATE_WITHIN_WINDSPEED\":true,\"MONITOR_WIND_CONDITIONS\":true,\"SEEN_WEATHER_FORECAST\":true,\"SUSTAINED_WINDS\":1,\"SUSTAINED_WINDS_PERCENT\":1,\"MAX_WINDSPEED\":1,\"expirationDate\":\"2025-07-11T15:26:19.316Z\",\"certificate_name\":\"WIND_FEATURE_MODEL\",\"submitted_by\":\"jakolker\",\"submitter_name\":\"Jack Kolker\",\"current_date\":\"2025-05-02\",\"svg_file\":\"\\u003cSVG Content\\u003e\",\"passed\":false,\"archived\":false}"




// GetCertificateDBObject converts a Certificate struct to a CertificateDBObject struct.
// CertificateID can be modified here to ensure uniqueness.
func GetCertificateDBObject(cert *Certificate) (*CertificateDBObject, error) {
	serializedCert, err := json.Marshal(cert)
	if err != nil {
		return nil, err
	}
	// Parse the ISO8601 expiration date to time.Time
	expTime, err := time.Parse(time.RFC3339, cert.ExpirationDate)
	if err != nil {
		return nil, err
	}

	cert_db_obj := &CertificateDBObject{
		PilotID:               cert.PilotID,
		CertificateID:         GetRandomString(10),
		DroneID:               cert.DroneID,
		ExpirationDate:        strconv.FormatInt(expTime.Unix(), 10),
		SerializedCertificate: string(serializedCert),
	}
	return cert_db_obj, nil
}