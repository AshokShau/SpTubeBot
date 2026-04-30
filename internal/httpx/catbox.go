package httpx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	CatboxAPI    = "https://catbox.moe/user/api.php"
	LitterboxAPI = "https://litterbox.catbox.moe/resources/internals/api.php"
	UserAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// UploadToCatbox uploads a file to Catbox.
func UploadToCatbox(filePath, userhash string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	r, w := io.Pipe()
	writer := multipart.NewWriter(w)

	go func() {
		defer w.Close()
		defer writer.Close()

		_ = writer.WriteField("reqtype", "fileupload")
		if userhash != "" {
			_ = writer.WriteField("userhash", userhash)
		}

		part, err := writer.CreateFormFile("fileToUpload", filepath.Base(filePath))
		if err != nil {
			return
		}
		_, _ = io.Copy(part, file)
	}()

	req, err := http.NewRequest("POST", CatboxAPI, r)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", UserAgent)

	resp, err := catClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	resStr := strings.TrimSpace(string(respBody))
	if resStr == "" {
		return "", fmt.Errorf("catbox returned an empty response")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("catbox error (%d): %s", resp.StatusCode, resStr)
	}

	if strings.Contains(resStr, "Hashes do not match") || strings.Contains(resStr, "Error") {
		return "", fmt.Errorf("catbox error: %s", resStr)
	}

	return resStr, nil
}

// UploadToLitterbox uploads a file to Litterbox with a specific expiration time.
func UploadToLitterbox(filePath, timeStr string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	r, w := io.Pipe()
	writer := multipart.NewWriter(w)

	go func() {
		defer w.Close()
		defer writer.Close()

		_ = writer.WriteField("reqtype", "fileupload")
		if timeStr == "" {
			timeStr = "24h"
		}
		_ = writer.WriteField("time", timeStr)

		part, err := writer.CreateFormFile("fileToUpload", filepath.Base(filePath))
		if err != nil {
			return
		}
		_, _ = io.Copy(part, file)
	}()

	req, err := http.NewRequest("POST", LitterboxAPI, r)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", UserAgent)

	resp, err := catClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	resStr := strings.TrimSpace(string(respBody))
	if resStr == "" {
		return "", fmt.Errorf("litterbox returned an empty response")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("litterbox error (%d): %s", resp.StatusCode, resStr)
	}

	if strings.Contains(resStr, "Error") {
		return "", fmt.Errorf("litterbox error: %s", resStr)
	}

	return resStr, nil
}

// DeleteCatboxFiles deletes files from Catbox.
func DeleteCatboxFiles(files []string, userhash string) error {
	if userhash == "" {
		return errors.New("userhash is required")
	}

	data := fmt.Sprintf("reqtype=deletefiles&userhash=%s&files=%s", userhash, strings.Join(files, " "))
	resp, err := client.Post(CatboxAPI, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	resStr := string(respBody)
	if !strings.Contains(resStr, "Files successfully deleted") && resStr != "" {
		return fmt.Errorf("catbox error: %s", resStr)
	}

	return nil
}

// CreateCatboxAlbum creates an album on Catbox.
func CreateCatboxAlbum(title, desc string, files []string, userhash string) (string, error) {
	if userhash == "" {
		return "", errors.New("userhash is required")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("reqtype", "createalbum")
	_ = writer.WriteField("userhash", userhash)
	_ = writer.WriteField("title", title)
	_ = writer.WriteField("desc", desc)
	_ = writer.WriteField("files", strings.Join(files, " "))

	err := writer.Close()
	if err != nil {
		return "", err
	}

	resp, err := client.Post(CatboxAPI, writer.FormDataContentType(), body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

// EditCatboxAlbum edits an album on Catbox.
func EditCatboxAlbum(short, title, desc string, files []string, userhash string) error {
	if userhash == "" {
		return errors.New("userhash is required")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("reqtype", "editalbum")
	_ = writer.WriteField("userhash", userhash)
	_ = writer.WriteField("short", short)
	_ = writer.WriteField("title", title)
	_ = writer.WriteField("desc", desc)
	_ = writer.WriteField("files", strings.Join(files, " "))

	err := writer.Close()
	if err != nil {
		return err
	}

	resp, err := client.Post(CatboxAPI, writer.FormDataContentType(), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// AddToCatboxAlbum adds files to an album on Catbox.
func AddToCatboxAlbum(short string, files []string, userhash string) error {
	if userhash == "" {
		return errors.New("userhash is required")
	}

	data := fmt.Sprintf("reqtype=addtoalbum&userhash=%s&short=%s&files=%s", userhash, short, strings.Join(files, " "))
	resp, err := client.Post(CatboxAPI, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// RemoveFromCatboxAlbum removes files from an album on Catbox.
func RemoveFromCatboxAlbum(short string, files []string, userhash string) error {
	if userhash == "" {
		return errors.New("userhash is required")
	}

	data := fmt.Sprintf("reqtype=removefromalbum&userhash=%s&short=%s&files=%s", userhash, short, strings.Join(files, " "))
	resp, err := client.Post(CatboxAPI, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// DeleteCatboxAlbum deletes an album on Catbox.
func DeleteCatboxAlbum(short string, userhash string) error {
	if userhash == "" {
		return errors.New("userhash is required")
	}
	data := fmt.Sprintf("reqtype=deletealbum&userhash=%s&short=%s", userhash, short)
	resp, err := client.Post(CatboxAPI, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
