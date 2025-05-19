package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type RawRecord struct {
	ReturnToLaunchPlan     string `json:"RETURN_TO_LAUNCH_PLAN"`
	OperateWithinWindspeed string `json:"OPERATE_WITHIN_WINDSPEED"`
	MonitorWindConditions  string `json:"MONITOR_WIND_CONDITIONS"`
	SeenWeatherForecast    string `json:"SEEN_WEATHER_FORECAST"`
	SustainedWinds         string `json:"SUSTAINED_WINDS"`
	SustainedWindsPercent  string `json:"SUSTAINED_WINDS_PERCENT"`
	MaxWindspeed           string `json:"MAX_WINDSPEED"`
	ExpirationDate         string `json:"expirationDate"`
	CertificateName        string `json:"certificate_name"`
	SubmittedBy            string `json:"submitted_by"`
	SubmitterName          string `json:"submitter_name"`
	CurrentDate            string `json:"current_date"`
	SvgFile                string `json:"svg_file"`
	Passed                 string `json:"passed"`
	Archived               string `json:"archived"`
}

type Record struct {
	DroneID   string `json:"droneID"`
	Zip       string `json:"zip"`
	FlyTime   string `json:"flyTime"`
	FlyRecord string `json:"flyRecord"`
	Reserved  string `json:"reserved"`
}

func ConvertToRFC3339(unixtime string) string {
	unixtimeInt, _ := strconv.ParseInt(unixtime, 10, 64)
	t := time.Unix(unixtimeInt, 0)
	return t.Format(time.RFC3339)
}

func ConvertToUnixTime(datetime string) string {
	t, _ := time.Parse(time.RFC3339, datetime)
	return strconv.FormatInt(t.Unix(), 10)
}

func DecompressRecord(returnedRecord string) string {
	// return returnedRecord
	// Parse the JSON string to
	var records []Record
	var rawRecord RawRecord
	var rawRecords []RawRecord
	json.Unmarshal([]byte(returnedRecord), &records)

	for _, record := range records {
		// skip if record.FlyTime == "-1"
		if record.DroneID == "" {
			continue
		}

		// use default id for testing
		id := "user1"

		// fmt.Println(record)
		decryptedFlyRecord, _ := Decrypt(record.FlyRecord, id)
		// fmt.Println(decryptedFlyRecord)
		// split the decryptedFlyRecord by ","
		err := json.Unmarshal([]byte(decryptedFlyRecord), &rawRecord)
		if err != nil {
			fmt.Println("Error unmarshalling decryptedFlyRecord:", err)
			continue
		}

		rawRecords = append(rawRecords, rawRecord)
	}

	// fmt.Println(rawRecords)

	rawRecordsJSON, _ := json.Marshal(rawRecords)
	return string(rawRecordsJSON)
}

func CompressRecord(rawRecord *RawRecord) *Record {
	// Convert the rawRecord to a string
	jsonBytes, err := json.Marshal(rawRecord)

	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	recordString := string(jsonBytes)

	// use default id for testing
	id := "user1"

	encryptedFlyRecord, _ := Encrypt(recordString, id)
	droneRecord := Record{
		DroneID:   rawRecord.SubmittedBy,
		Zip:       "",
		FlyTime:   ConvertToUnixTime(rawRecord.ExpirationDate),
		FlyRecord: encryptedFlyRecord,
		Reserved:  "",
	}
	return &droneRecord
}

// func ImportFromFile(filePath string) {
// 	// Open the CSV file
// 	csvFile, err := os.Open("./ds1.csv")
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}
// 	defer csvFile.Close()

// 	// Parse the CSV file
// 	reader := csv.NewReader(csvFile)
// 	reader.TrimLeadingSpace = true
// 	records, err := reader.ReadAll()
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	// Convert the CSV records
// 	for i, record := range records {
// 		if i == 0 {
// 			continue
// 		}

// 		flyRecord := strings.Join(record, ",")
// 		encryptedFlyRecord, _ := Encrypt(flyRecord)

// 		droneRecord := Record{
// 			DroneID:   record[0],
// 			Zip:       record[2],
// 			FlyTime:   ConvertToUnixTime(record[3]),
// 			FlyRecord: encryptedFlyRecord,
// 			Reserved:  "",
// 		}
// 		fmt.Println(i, droneRecord)
// 		createDroneRecord(droneRecord.DroneID, droneRecord.Zip, droneRecord.FlyTime, droneRecord.FlyRecord, droneRecord.Reserved)
// 	}
// }
