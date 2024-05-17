package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type BackupType string

const (
	GoogleDrive BackupType = "google-drive"
)

type Config struct {
	//Bitwarden config
	BitwardenID             string
	BitwardenSecret         string
	BitwardenMasterPassword string
	BitwardenServer         string
	BitwardenExportFormat   string

	// Google Drive Config
	GoogleDriveCredentialsName string
	GoogleDriveParentID        string

	// BackupDelayMinutes represents how long the program should wait in minutes before backing up the data again
	BackupDelayMinutes int
}

func NewConfigFromEnv() (*Config, error) {
	backupDelayMinutes := os.Getenv("BACKUP_DELAY_MINUTES")
	if backupDelayMinutes == "" {
		return nil, errors.New("environment var 'BACKUP_DELAY_MINUTES' is required to run")
	}

	backupDelayMinutesInt, err := strconv.Atoi(backupDelayMinutes)
	if err != nil {
		return nil, errors.New("BACKUP_DELAY_MINUTES must an integer value")
	}

	bitwardenId := os.Getenv("BITWARDEN_ID")
	if bitwardenId == "" {
		return nil, errors.New("environment var 'BITWARDEN_ID' is required to run")
	}

	bitwardenSecret := os.Getenv("BITWARDEN_SECRET")
	if bitwardenSecret == "" {
		return nil, errors.New("environment var 'BITWARDEN_SECRET' is required to run")
	}

	bitwardenMasterPassword := os.Getenv("BITWARDEN_MASTER_PASSWORD")
	if bitwardenMasterPassword == "" {
		return nil, errors.New("environment var 'BITWARDEN_MASTER_PASSWORD' is required to run")
	}
	bitwardenExportFormat := os.Getenv("BITWARDEN_EXPORT_FORMAT")
	switch ExportFormat(bitwardenExportFormat) {
	case JSON, JSONEncrypted, CSV:
	default:
		return nil, errors.New("value for BITWARDEN_EXPORT_FORMAT must be: 'json', 'encrypted_json', or 'csv'")
	}

	bitwardenServer := os.Getenv("BITWARDEN_SERVER")

	// TODO: build out modular system for pushing to other services
	googleDriveCredName := os.Getenv("GOOGLE_DRIVE_CREDENTIALS")
	googleDriveParentID := os.Getenv("GOOGLE_DRIVE_PARENT_ID")

	return &Config{
		BitwardenID:             bitwardenId,
		BitwardenSecret:         bitwardenSecret,
		BitwardenMasterPassword: bitwardenMasterPassword,
		BitwardenServer:         bitwardenServer,

		GoogleDriveCredentialsName: googleDriveCredName,
		GoogleDriveParentID:        googleDriveParentID,

		BackupDelayMinutes: backupDelayMinutesInt,
	}, nil
}

func main() {
	cfg, err := NewConfigFromEnv()
	if err != nil {
		panic(err)
	}

	firstRun := true
	ticker := time.NewTicker(time.Duration(cfg.BackupDelayMinutes))
	for {
		if firstRun {
			if err := backup(cfg, time.Now()); err != nil {
				panic(err)
			}
			firstRun = false
		}

		fmt.Printf("Waiting %d minutes for next backup...\n", cfg.BackupDelayMinutes)
		t := <-ticker.C
		if err := backup(cfg, t); err != nil {
			panic(err)
		}
	}
}

func backup(cfg *Config, at time.Time) error {
	fmt.Printf("Running backup utility at %s\n", at.String())
	bw, err := NewClient(cfg)
	if err != nil {
		return err
	}
	defer bw.Close()

	res, err := bw.Export(JSON)
	if err != nil {
		return err
	}

	f, err := os.Open(filepath.Join("/config", cfg.GoogleDriveCredentialsName))
	if err != nil {
		return err
	}
	defer f.Close()

	drive, err := newGoogleDriveDriver(cfg.GoogleDriveParentID, f)
	if err != nil {
		return err
	}

	return drive.Upload(res)
}
