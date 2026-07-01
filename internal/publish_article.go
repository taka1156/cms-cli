package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/taka1156/brite/internal/entity"
)

type ChangeType int

const (
	Added ChangeType = iota
	Modified
	NoChange
	Deleted
)

type ImageDiff struct {
	FilePath   string
	Size       int64
	ChangeType ChangeType
}

type PublishArticleCommand struct{}

func NewPublishArticleCommand() *PublishArticleCommand {
	return &PublishArticleCommand{}
}

func (c *PublishArticleCommand) Publish(clientConfig entity.ClientConfig) {
	briteConfig, err := loadJson[entity.BriteConfig](clientConfig.ConfigPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	caches := []entity.ImageCache{}
	cacheFilePath := filepath.Join(briteConfig.CacheDir, entity.CACHE_FILE_NAME)
	if _, err := os.Stat(cacheFilePath); os.IsNotExist(err) {
		if err := saveCache(cacheFilePath, caches); err != nil {
			fmt.Printf("Error creating .caches.json: %v\n", err)
			return
		}
	} else {
		caches, err = loadJson[[]entity.ImageCache](cacheFilePath)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}

	// Create a map for quick lookup of existing caches
	cacheByPath := make(map[string]entity.ImageCache)
	for _, cache := range caches {
		cacheByPath[cache.FilePath] = cache
	}

	diffs, err := detectDiff(briteConfig.ImageDir, cacheByPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	ctx := context.Background()

	client, err := newS3Client()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = applyDiffs(ctx, client, briteConfig.R2.BucketName, diffs)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if err := postOutput(briteConfig, client); err != nil {
		fmt.Println("Error:", err)
		return
	}

	newCaches := []entity.ImageCache{}
	for _, diff := range diffs {
		switch diff.ChangeType {
		case Added, Modified, NoChange:
			newCaches = append(newCaches, entity.ImageCache{
				FilePath: diff.FilePath,
				Size:     diff.Size,
			})
		case Deleted:
			// skip deleted images
		}
	}

	err = saveCache(cacheFilePath, newCaches)
	if err != nil {
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
	case ".webp":
		return "image/webp"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

func detectDiff(imageDir string, caches map[string]entity.ImageCache) ([]ImageDiff, error) {
	current := map[string]entity.ImageCache{}

	err := filepath.Walk(imageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		current[path] = entity.ImageCache{
			FilePath: path,
			Size:     info.Size(),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var diffs []ImageDiff

	// Detect added, modified, no change images
	for path, img := range current {
		if prev, ok := caches[path]; !ok {
			diffs = append(diffs, ImageDiff{
				FilePath:   path,
				Size:       img.Size,
				ChangeType: Added,
			})
		} else if prev.Size != img.Size {
			diffs = append(diffs, ImageDiff{
				FilePath:   path,
				Size:       img.Size,
				ChangeType: Modified,
			})
		} else if prev.Size == img.Size {
			diffs = append(diffs, ImageDiff{
				FilePath:   path,
				Size:       img.Size,
				ChangeType: NoChange,
			})
		}
	}

	// Detect deleted images
	for path, cache := range caches {
		if _, ok := current[path]; !ok {
			diffs = append(diffs, ImageDiff{
				FilePath:   cache.FilePath,
				Size:       0,
				ChangeType: Deleted,
			})
		}
	}

	return diffs, nil
}

func saveCache(path string, caches []entity.ImageCache) error {
	data, err := json.MarshalIndent(caches, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func applyDiffs(ctx context.Context, client *s3.Client, bucketName string, diffs []ImageDiff) error {
	for _, diff := range diffs {
		switch diff.ChangeType {
		case Added, Modified:
			if err := uploadFileToR2(ctx, client, bucketName, diff.FilePath, diff.FilePath); err != nil {
				return err
			}
		case Deleted:
			if err := deleteFileFromR2(ctx, client, bucketName, diff.FilePath); err != nil {
				return err
			}
		}
	}
	return nil
}

func postOutput(briteConfig entity.BriteConfig, client *s3.Client) error {
	jsonDir := []string{
		entity.ALL_JSON_FILE_NAME,
		entity.CATEGORY_JSON_FILE_NAME,
		entity.TAG_JSON_FILE_NAME,
	}
	for _, jsonFile := range jsonDir {
		filePath := filepath.Join(briteConfig.OutputDir, jsonFile)

		if err := uploadFileToR2(context.TODO(), client, briteConfig.R2.BucketName, filePath, filePath); err != nil {
			return fmt.Errorf("failed to add file %s to R2: %w", filePath, err)
		}
	}

	fmt.Println("Successfully uploaded output files.")

	return nil
}

func newS3Client() (*s3.Client, error) {
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	endpoint := os.Getenv("R2_ENDPOINT")
	usePathStyle, err := strconv.ParseBool(os.Getenv("ENABLE_PATH_STYLE_ENDPOINTS"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ENABLE_PATH_STYLE_ENDPOINTS: %w", err)
	}

	if accessKey == "" || secretKey == "" || endpoint == "" {
		return nil, fmt.Errorf("R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, and R2_ENDPOINT environment variables must be set")
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
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = usePathStyle
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
