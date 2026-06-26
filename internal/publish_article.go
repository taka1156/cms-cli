package internal

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/taka1156/cms-cli/internal/entity"
)

type PublishArticleCommand struct{}

func NewPublishArticleCommand() *PublishArticleCommand {
	return &PublishArticleCommand{}
}

func (c *PublishArticleCommand) Publish() {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	ctx := context.Background()

	clint, err := newS3Client(config)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	changedImages, err := getChangeImages(config.ImageDir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, imagePath := range changedImages {
		key := strings.TrimPrefix(imagePath, config.ImageDir+"/")
		if err := uploadFileToR2(ctx, clint, config.R2.BucketName, imagePath, key); err != nil {
			fmt.Println("Error:", err)
			return
		}
	}

	deletedImages, err := getDeletedImages(config.ImageDir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, imagePath := range deletedImages {
		key := strings.TrimPrefix(imagePath, config.ImageDir+"/")
		if err := deleteFileFromR2(ctx, clint, config.R2.BucketName, key); err != nil {
			fmt.Println("Error:", err)
			return
		}
	}

	if err := postOutput(config); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Successfully posted output files and images to R2.")
}

func contentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

func getChangeImages(imageDir string) ([]string, error) {
	cmd := exec.Command("git", "diff", "HEAD~1", "HEAD", "--name-only", "--cached")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return getAllImages(imageDir)
	}

	var images []string
	for _, f := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(f, imageDir) {
			images = append(images, f)
		}
	}

	return images, nil
}

func getDeletedImages(imageDir string) ([]string, error) {
	cmd := exec.Command("git", "diff", "HEAD~1", "HEAD", "--name-only", "--diff-filter=D")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, nil
	}

	var images []string
	for _, f := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(f, imageDir) {
			images = append(images, f)
		}
	}

	return images, nil
}

func getAllImages(imageDir string) ([]string, error) {
	var images []string
	err := filepath.Walk(imageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			images = append(images, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return images, nil
}

func newS3Client(cmsConfig entity.CMSConfig) (*s3.Client, error) {
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")

	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("R2_ACCESS_KEY_ID and R2_SECRET_ACCESS_KEY environment variables must be set")
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cmsConfig.R2.Endpoint)
	})

	return client, nil
}

func uploadFileToR2(ctx context.Context, client *s3.Client, bucketName, filePath, key string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer f.Close()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(bucketName),
		Key:          aws.String(key),
		Body:         f,
		ContentType:  aws.String(contentType(filePath)),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to R2: %w", err)
	}

	fmt.Printf("Uploaded %s to R2 bucket %s\n", filePath, bucketName)

	return nil
}

func deleteFileFromR2(ctx context.Context, client *s3.Client, bucketName, key string) error {
	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}

	fmt.Printf("Deleted %s from R2 bucket %s\n", key, bucketName)

	return nil
}

func postOutput(cmsConfig entity.CMSConfig) error {
	outputDir := []string{
		entity.ALL_JSON_FILE_NAME,
		entity.CATEGORY_JSON_FILE_NAME,
		entity.TAG_JSON_FILE_NAME,
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for _, filePath := range outputDir {
		path := filepath.Join(cmsConfig.OutputDir, filePath)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		part, err := writer.CreateFormFile("files", filepath.Base(path))
		if err != nil {
			return fmt.Errorf("failed to create form file for %s: %w", path, err)
		}

		if _, err := part.Write(data); err != nil {
			return fmt.Errorf("failed to write file %s to form: %w", path, err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", cmsConfig.R2.Endpoint, &body)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+os.Getenv("R2_AUTH_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload files, status code: %d", resp.StatusCode)
	}

	fmt.Println("Successfully uploaded output files.")

	return nil
}
