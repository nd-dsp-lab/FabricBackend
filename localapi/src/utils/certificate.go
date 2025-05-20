package utils


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
// 	CertificateContent    *Certificate `json:"certificate_content"`
// 	SerializedCertificate string `json:"serialized_certificate"`
// }
// This structure can be changed in handlers.go, but it requires more modifications
// in the code, so it is better to keep it as it is for now.
type Certificate struct {
	DroneID                string `json:"drone_id" default:"Drone01"`
	PilotID                string `json:"pilot_id" default:"Pilot01"`
	ReturnToLaunchPlan     string `json:"RETURN_TO_LAUNCH_PLAN" default:"true"`
	OperateWithinWindspeed string `json:"OPERATE_WITHIN_WINDSPEED" default:"true"`
	MonitorWindConditions  string `json:"MONITOR_WIND_CONDITIONS" default:"true"`
	SeenWeatherForecast    string `json:"SEEN_WEATHER_FORECAST" default:"true"`
	SustainedWinds         string `json:"SUSTAINED_WINDS" default:"1"`
	SustainedWindsPercent  string `json:"SUSTAINED_WINDS_PERCENT" default:"1"`
	MaxWindspeed           string `json:"MAX_WINDSPEED" default:"1"`
	ExpirationDate         string `json:"expirationDate" default:"2025-07-11T15:26:19.316Z"`
	CertificateName        string `json:"certificate_name" default:"WIND_FEATURE_MODEL"`
	SubmittedBy            string `json:"submitted_by" default:"jakolker"`
	SubmitterName          string `json:"submitter_name" default:"Jack Kolker"`
	CurrentDate            string `json:"current_date" default:"2025-05-02"`
	SvgFile                string `json:"svg_file" default:"<SVG Content>"`
	Passed                 string `json:"passed" default:"false"`
	Archived               string `json:"archived" default:"false"`
}
