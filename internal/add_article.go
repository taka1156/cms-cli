package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/taka1156/brite/internal/entity"
	"gopkg.in/yaml.v3"
)

type AddArticleCommand struct{}

func NewAddArticleCommand() *AddArticleCommand {
	return &AddArticleCommand{}
}

func (c *AddArticleCommand) Add(clientConfig entity.ClientConfig) {
	config, err := loadJson[entity.BriteConfig](clientConfig.ConfigPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	// input title
	fmt.Print("Title: ")
	titleInput, _ := reader.ReadString('\n')
	title := strings.TrimSpace(titleInput)
	if title == "" {
		fmt.Println("Error: title must not be empty.")
		return
	}

	// select category from config.Categories
	category, err := promptSingleSelect(reader, "Category", config.Categories)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// select tags from config.Tags
	tags, err := promptMultiSelect(reader, "Tags", config.Tags)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// auto-generate slug (UUID) and ensure uniqueness
	slug, err := generateUniqueSlug(config.ArticleDir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// set created_at timestamp
	date := time.Now().Format(time.RFC3339)

	post := entity.PostSummary{
		Slug:      slug,
		Title:     title,
		Thumbnail: "",
		Category:  category,
		Tags:      tags,
		CreatedAt: date,
		UpdatedAt: "",
	}

	// write the post file with front matter
	if err := writePostFile(config.ArticleDir, post); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Success! Created %s\n", filepath.Join(config.ArticleDir, post.Slug+".md"))
}

// writePostFile creates a Markdown file with front matter in the specified article directory.
func writePostFile(articleDir string, post entity.PostSummary) error {
	if err := os.MkdirAll(articleDir, 0755); err != nil {
		return fmt.Errorf("failed to create article_dir: %w", err)
	}

	frontMatter := entity.PostSummary{
		Title:     post.Title,
		Thumbnail: post.Thumbnail,
		Category:  post.Category,
		Tags:      post.Tags,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	yamlBytes, err := yaml.Marshal(frontMatter)
	if err != nil {
		return fmt.Errorf("failed to generate front matter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlBytes)
	buf.WriteString("---\n\n")

	path := filepath.Join(articleDir, post.Slug+".md")

	// Check if the file already exists to avoid overwriting
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists unexpectedly: %s", path)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write post file: %w", err)
	}

	return nil
}

// promptSingleSelect prompts the user to select a single option from a list. It returns the selected option or an empty string if skipped.
func promptSingleSelect(reader *bufio.Reader, label string, options []string) (string, error) {
	if len(options) == 0 {
		fmt.Printf("Notice: no %s registered in brite.schema.json. Skipping.\n", strings.ToLower(label))
		return "", nil
	}

	fmt.Printf("%s:\n", label)
	for i, opt := range options {
		fmt.Printf("  [%d] %s\n", i+1, opt)
	}
	fmt.Printf("Select a number (leave empty to skip): ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil
	}

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(options) {
		return "", fmt.Errorf("invalid selection: %s", input)
	}

	return options[idx-1], nil
}

// promptMultiSelect prompts the user to select multiple options from a list. It returns the selected options or an empty slice if skipped.
func promptMultiSelect(reader *bufio.Reader, label string, options []string) ([]string, error) {
	if len(options) == 0 {
		fmt.Printf("Notice: no %s registered in brite.schema.json. Skipping.\n", strings.ToLower(label))
		return []string{}, nil
	}

	fmt.Printf("%s:\n", label)
	for i, opt := range options {
		fmt.Printf("  [%d] %s\n", i+1, opt)
	}
	fmt.Printf("Select numbers, comma-separated (leave empty to skip): ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return []string{}, nil
	}

	parts := strings.Split(input, ",")
	selected := make([]string, 0, len(parts))
	seen := make(map[string]bool)

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		idx, err := strconv.Atoi(p)
		if err != nil || idx < 1 || idx > len(options) {
			return nil, fmt.Errorf("invalid selection: %s", p)
		}
		tag := options[idx-1]
		if !seen[tag] {
			selected = append(selected, tag)
			seen[tag] = true
		}
	}

	return selected, nil
}

// generateUniqueSlug generates a unique slug based on UUID. It retries if a collision occurs.
func generateUniqueSlug(articleDir string) (string, error) {
	for i := 0; i < 10; i++ {
		candidate := uuid.New().String()
		path := filepath.Join(articleDir, candidate+".md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("failed to generate a unique slug after multiple attempts")
}
