package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	ApiKey      string   `json:"api_key"`
	ApiEmail    string   `json:"api_email"`
	ZoneID      string   `json:"zone_id"`
	Domain      []string `json:"domain"`
	ApiEndpoint string   `json:"api_endpoint"`
	RecordType  string   `json:"record_type"`
}

var config Config
var logger *zap.Logger
var homePath string

func main() {
	_homePath, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	homePath = _homePath

	logFile := openLogFile()
	defer logFile.Close()

	fileCore := initLogger(logFile)
	logger = zap.New(fileCore)

	defer logger.Sync()
	logger.Info("Initializing logger done successfully.")

	_config, err := loadConfig()
	config = _config

	if err != nil {
		logger.Error(fmt.Sprintf("Error loading config.json : %v", err.Error()))
		return
	} else {
		logger.Info("Loading config done successfully.")
	}

	// Get current public IP address
	currentIP, err := getCurrentIP()
	if err != nil {
		logger.Error(fmt.Sprintf("Error fetching current IP: %v", err.Error()))
		return
	} else {
		logger.Info(fmt.Sprintf("Fetching current IP done successfully: %v", currentIP))
	}

	for _, domain := range config.Domain {
		// Get the DNS record ID for the specified domain
		recordID, err := getRecordID(domain, config.RecordType)

		if err != nil {
			logger.Error(fmt.Sprintf("Error fetching DNS record ID: %v", err.Error()))
			break
		} else {
			logger.Info(fmt.Sprintf("Fetching DNS record ID for %v done successfully: %v", domain, recordID))
		}

		// Update the A name record with the current IP address
		err = updateDNSRecord(config.ZoneID, recordID, domain, currentIP)
		if err != nil {
			logger.Error(fmt.Sprintf("Error updating DNS record ID: %v", err.Error()))
			break
		} else {
			logger.Info(fmt.Sprintf("Updating DNS record ID done successfully: %v", currentIP))
		}

		logger.Info(fmt.Sprintf("DNS record updated successfully: %v(%v) as %v", domain, recordID, currentIP))
	}
}

func loadConfig() (Config, error) {
	file, err := os.Open(homePath + "/CF-DDNS/config.json")

	if err != nil {
		return Config{}, err
	}

	defer file.Close()
	decoder := json.NewDecoder(file)
	var config Config

	if err := decoder.Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func initLogger(logFile *os.File) zapcore.Core {
	log.SetOutput(logFile)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	fileCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(logFile),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)

	return fileCore
}

func openLogFile() *os.File {
	now := time.Now()
	// millis := now.UnixNano() / 1000000
	if err := os.MkdirAll(homePath+"/CF-DDNS/logs", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(fmt.Sprintf(homePath+"/CF-DDNS/logs/log_%v.log", now.Format(time.RFC3339)))

	if err != nil {
		panic(err)
	}

	return file
}

func getCurrentIP() (string, error) {
	resp, err := http.Get("https://api64.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(ip)), nil
}

func getRecordID(domain, recordType string) (string, error) {
	url := fmt.Sprintf(config.ApiEndpoint, config.ZoneID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Auth-Email", config.ApiEmail)
	req.Header.Set("X-Auth-Key", config.ApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = parseJSONResponse(resp, &result)
	if err != nil {
		return "", err
	}

	// Find the DNS record ID for the specified domain and type
	records, ok := result["result"].([]interface{})
	if !ok {
		return "", fmt.Errorf("failed to parse DNS records")
	}

	for _, r := range records {
		record := r.(map[string]interface{})
		if record["name"].(string) == domain && record["type"].(string) == recordType {
			return record["id"].(string), nil
		}
	}

	return "", fmt.Errorf("DNS record not found")
}

func updateDNSRecord(zoneID, recordID, domain, ip string) error {
	url := fmt.Sprintf(config.ApiEndpoint, zoneID) + "/" + recordID
	payload := map[string]interface{}{
		"type":    config.RecordType,
		"name":    domain,
		"content": ip,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("X-Auth-Email", config.ApiEmail)
	req.Header.Set("X-Auth-Key", config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update DNS record, status code: %d", resp.StatusCode)
	}

	return nil
}

func parseJSONResponse(resp *http.Response, target interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		return err
	}

	return nil
}
