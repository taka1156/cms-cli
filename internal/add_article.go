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
	"github.com/taka1156/cms-cli/internal/entity"
	"gopkg.in/yaml.v3"
)

type AddArticleCommand struct{}

func NewAddArticleCommand() *AddArticleCommand {
	return &AddArticleCommand{}
}

// 記事登録（new）コマンドの処理
func (c *AddArticleCommand) Add() {
	config, err := loadJson[entity.CMSConfig](entity.CONFIG_FILE_NAME)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	// 1. タイトルの対話入力
	fmt.Print("Title: ")
	titleInput, _ := reader.ReadString('\n')
	title := strings.TrimSpace(titleInput)
	if title == "" {
		fmt.Println("Error: title must not be empty.")
		return
	}

	// 2. categoryの選択（一覧から番号選択）
	category, err := promptSingleSelect(reader, "Category", config.Categories)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 3. tagの選択（一覧から複数選択、カンマ区切り番号）
	tags, err := promptMultiSelect(reader, "Tags", config.Tags)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 4. slugの自動採番（UUIDベース、人間に採番させない）
	slug, err := generateUniqueSlug(config.ArticleDir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 5. dateの自動埋め込み
	date := time.Now().Format(time.RFC3339)

	post := entity.PostSummary{
		Slug:      slug,
		Title:     title,
		Category:  category,
		Tags:      tags,
		CreatedAt: date,
		UpdatedAt: "",
	}

	// 6. フロントマター付きMarkdownの書き出し
	if err := writePostFile(config.ArticleDir, post); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Success! Created %s\n", filepath.Join(config.ArticleDir, post.Slug+".md"))
}

// フロントマター付きMarkdownファイルを書き出す
func writePostFile(articleDir string, post entity.PostSummary) error {
	if err := os.MkdirAll(articleDir, 0755); err != nil {
		return fmt.Errorf("failed to create article_dir: %w", err)
	}

	frontMatter := struct {
		Title     string   `yaml:"title"`
		Category  string   `yaml:"category,omitempty"`
		Tags      []string `yaml:"tags,omitempty"`
		CreatedAt string   `yaml:"created_at"`
		UpdatedAt string   `yaml:"updated_at"`
	}{
		Title:     post.Title,
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

	// 念のため上書き防止（UUID採番後の最終チェック）
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists unexpectedly: %s", path)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write post file: %w", err)
	}

	return nil
}

// 単一選択のプロンプト（category用）
func promptSingleSelect(reader *bufio.Reader, label string, options []string) (string, error) {
	if len(options) == 0 {
		fmt.Printf("Notice: no %s registered in cmsc.json. Skipping.\n", strings.ToLower(label))
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

// 複数選択のプロンプト（tag用、カンマ区切りで番号を受け取る）
func promptMultiSelect(reader *bufio.Reader, label string, options []string) ([]string, error) {
	if len(options) == 0 {
		fmt.Printf("Notice: no %s registered in cmsc.json. Skipping.\n", strings.ToLower(label))
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

// UUIDベースでslugを自動採番する。衝突した場合は再採番する。
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
