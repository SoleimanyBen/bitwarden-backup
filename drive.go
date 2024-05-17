package main

import (
	"context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io"
)

type googleDriveUpload struct {
	srv    *drive.Service
	parent string
}

func newGoogleDriveDriver(parent string, key io.Reader) (*googleDriveUpload, error) {
	client, err := getDriveClient(key)
	if err != nil {
		return nil, err
	}

	return &googleDriveUpload{
		srv:    client,
		parent: parent,
	}, nil
}

func (gdp *googleDriveUpload) Upload(r io.Reader) error {
	_, err := gdp.srv.Files.Create(&drive.File{
		Name:    "vaultwarden-backup.json",
		Parents: []string{gdp.parent},
	}).Media(r).Do()
	return err
}

func getDriveClient(key io.Reader) (*drive.Service, error) {
	buf, err := io.ReadAll(key)
	if err != nil {
		return nil, err
	}

	creds, err := google.CredentialsFromJSON(context.Background(), buf, drive.DriveFileScope)
	if err != nil {
		return nil, err
	}

	srv, err := drive.NewService(context.Background(), option.WithCredentials(creds))
	if err != nil {
		return nil, err
	}

	return srv, nil
}
