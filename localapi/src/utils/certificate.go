package utils

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
