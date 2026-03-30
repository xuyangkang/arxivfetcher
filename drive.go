package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config, tokFile string) (*http.Client, error) {
	// Check if we have a token cached.
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\n--- GOOGLE DRIVE AUTHENTICATION ---\n")
	fmt.Printf("Go to the following link in your browser:\n\n%v\n\n", authURL)
	fmt.Printf("Type the authorization code you receive here and press enter: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	return tok, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Unable to cache oauth token: %v\n", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

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

// uploadFileToDrive uploads a file to Google Drive under the specified folderName.
func uploadFileToDrive(ctx context.Context, credentialsFile, folderName, filePath, fileName string) (string, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return "", fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// Because we might upload dynamically and search for the folder, we need DriveFileScope minimum.
	// We'll use DriveFileScope to keep it isolated to the app's files.
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		return "", fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}

	client, err := getClient(config, "token.json")
	if err != nil {
		return "", fmt.Errorf("Unable to get OAuth client: %v", err)
	}

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
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
