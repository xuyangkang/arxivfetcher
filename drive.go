package main

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// getOrCreateFolder checks if a folder exists by name, returns its ID. If not, creates it and returns ID.
func getOrCreateFolder(srv *drive.Service, folderName string) (string, error) {
	query := fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and name='%s' and trashed=false", folderName)
	r, err := srv.Files.List().Q(query).Spaces("drive").Fields("files(id, name)").Do()
	if err != nil {
		return "", fmt.Errorf("unable to search for folder: %v", err)
	}

	if len(r.Files) > 0 {
		return r.Files[0].Id, nil
	}

	// Create folder
	f := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
	}

	folder, err := srv.Files.Create(f).Fields("id").Do()
	if err != nil {
		return "", fmt.Errorf("unable to create folder: %v", err)
	}

	return folder.Id, nil
}

// uploadFile uploads a file to Google Drive under the specified folderName.
func uploadFileToDrive(ctx context.Context, credentialsFile, folderName, filePath, fileName string) (string, error) {
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		return "", fmt.Errorf("credentials file not found at %s. Please ensure you have it for Google Drive upload", credentialsFile)
	}

	srv, err := drive.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return "", fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	folderId, err := getOrCreateFolder(srv, folderName)
	if err != nil {
		return "", err
	}

	fileContent, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file for upload: %v", err)
	}
	defer fileContent.Close()

	driveFile := &drive.File{
		Name:    fileName,
		Parents: []string{folderId},
	}

	res, err := srv.Files.Create(driveFile).Media(fileContent).Fields("id, webViewLink").Do()
	if err != nil {
		return "", fmt.Errorf("unable to upload file: %v", err)
	}

	return res.WebViewLink, nil
}
