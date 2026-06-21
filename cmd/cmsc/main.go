package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// categories/tagsの1要素（名前+紐づく画像パス）
type TaxonomyDefinition struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// 設定ファイルの構造
type CMSConfig struct {
	Schema     string               `json:"$schema"`
	ContentDir string               `json:"content_dir"`
	OutputDir  string               `json:"output_dir"`
	Categories []TaxonomyDefinition `json:"categories"`
	Tags       []TaxonomyDefinition `json:"tags"`
}

// 各記事のデータ構造
type Post struct {
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	Category  string   `json:"category"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at" yaml:"created_at"`
	UpdatedAt string   `json:"updated_at" yaml:"updated_at"`
	Content   string   `json:"content" yaml:"-"`
}

// 最終出力のデータ構造（byCategory/byTagはslug参照のみで本文の重複を避ける）
type ResponseData struct {
	All        []Post              `json:"all"`
	ByCategory map[string][]string `json:"byCategory"`
	ByTag      map[string][]string `json:"byTag"`
}

// category.json / tag.json の1エントリ（画像情報 + 紐づく記事slug一覧）
type TaxonomyEntry struct {
	Image string   `json:"image"`
	Slugs []string `json:"slugs"`
}

// 出力ディレクトリ配下の固定名
const (
	allJSONFileName      = "all.json"
	categoryJSONFileName = "category.json"
	tagJSONFileName      = "tag.json"
)

// 初期化コマンドの処理
func runInit() {
	configName := "cmsc.json"

	// すでにファイルがある場合は上書きを防ぐ
	if _, err := os.Stat(configName); err == nil {
		fmt.Println("Error: cmsc.json already exists in this directory.")
		return
	}

	// デフォルト設定の作成
	defaultConfig := CMSConfig{
		Schema:     "./cms.schema.json",
		ContentDir: "./content",
		OutputDir:  "./dist",
		Categories: []TaxonomyDefinition{
			{Name: "Tech", Image: ""},
			{Name: "Log", Image: ""},
		},
		Tags: []TaxonomyDefinition{
			{Name: "Go", Image: ""},
			{Name: "CLI", Image: ""},
		},
	}

	// JSONに変換
	jsonBytes, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		fmt.Printf("Error generating default config: %v\n", err)
		return
	}

	// ファイル書き出し
	if err := os.WriteFile(configName, jsonBytes, 0644); err != nil {
		fmt.Printf("Error writing cmsc.json: %v\n", err)
		return
	}

	fmt.Println("Success! Created default cmsc.json with schema link.")
}

// cmsc.json を読み込むだけの共通処理
func loadConfig() (CMSConfig, error) {
	var config CMSConfig

	configFile, err := os.ReadFile("cmsc.json")
	if err != nil {
		return config, fmt.Errorf("cmsc.json not found. Run './cmsc init' to create a default configuration")
	}

	if err := json.Unmarshal(configFile, &config); err != nil {
		return config, fmt.Errorf("failed to parse cmsc.json: %w", err)
	}

	return config, nil
}

// dateを基準に降順（新しい記事が先頭）でソートする。
// パース不能なdateは最も古い扱いとして末尾に回す。
func sortPostsByDateDesc(posts []Post) {
	sort.SliceStable(posts, func(i, j int) bool {
		ti, errI := time.Parse(time.RFC3339, posts[i].CreatedAt)
		tj, errJ := time.Parse(time.RFC3339, posts[j].CreatedAt)

		if errI != nil && errJ != nil {
			return false
		}
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}

		return ti.After(tj)
	})
}

// slug配列を、対応するcreated_atを基準に降順（新しい記事が先頭）でソートする。
// slugToCreatedAtに存在しない/パース不能なslugは最も古い扱いとして末尾に回す。
func sortSlugsByDateDesc(slugs []string, slugToCreatedAt map[string]string) {
	sort.SliceStable(slugs, func(i, j int) bool {
		ti, errI := time.Parse(time.RFC3339, slugToCreatedAt[slugs[i]])
		tj, errJ := time.Parse(time.RFC3339, slugToCreatedAt[slugs[j]])

		if errI != nil && errJ != nil {
			return false
		}
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}

		return ti.After(tj)
	})
}

// 任意のデータをインデント付きJSONとしてファイルに書き出す共通処理
func writeJSONFile(path string, v interface{}) error {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	if err := os.WriteFile(path, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// {名前: [slug,...]} と {名前: image} を合成して、
// category.json / tag.json 用の {名前: {image, slugs}} 構造を組み立てる
func buildTaxonomyOutput(slugsByName map[string][]string, imagesByName map[string]string) map[string]TaxonomyEntry {
	output := make(map[string]TaxonomyEntry, len(slugsByName))
	for name, slugs := range slugsByName {
		output[name] = TaxonomyEntry{
			Image: imagesByName[name],
			Slugs: slugs,
		}
	}
	return output
}

// TaxonomyDefinitionのスライスからnameだけを取り出す
func taxonomyNames(defs []TaxonomyDefinition) []string {
	names := make([]string, 0, len(defs))
	for _, d := range defs {
		names = append(names, d.Name)
	}
	return names
}

// 記事登録（new）コマンドの処理
func runNew() {
	config, err := loadConfig()
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
	category, err := promptSingleSelect(reader, "Category", taxonomyNames(config.Categories))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 3. tagの選択（一覧から複数選択、カンマ区切り番号）
	tags, err := promptMultiSelect(reader, "Tags", taxonomyNames(config.Tags))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 4. slugの自動採番（UUIDベース、人間に採番させない）
	slug, err := generateUniqueSlug(config.ContentDir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 5. dateの自動埋め込み
	date := time.Now().Format(time.RFC3339)

	post := Post{
		Slug:      slug,
		Title:     title,
		Category:  category,
		Tags:      tags,
		CreatedAt: date,
		UpdatedAt: "",
	}

	// 6. フロントマター付きMarkdownの書き出し
	if err := writePostFile(config.ContentDir, post); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Success! Created %s\n", filepath.Join(config.ContentDir, post.Slug+".md"))
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
func generateUniqueSlug(contentDir string) (string, error) {
	for i := 0; i < 10; i++ {
		candidate := uuid.New().String()
		path := filepath.Join(contentDir, candidate+".md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("failed to generate a unique slug after multiple attempts")
}

// フロントマター付きMarkdownファイルを書き出す
func writePostFile(contentDir string, post Post) error {
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		return fmt.Errorf("failed to create content_dir: %w", err)
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

	path := filepath.Join(contentDir, post.Slug+".md")

	// 念のため上書き防止（UUID採番後の最終チェック）
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists unexpectedly: %s", path)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write post file: %w", err)
	}

	return nil
}

func main() {
	// コマンド引数のチェック
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			runInit()
			return
		case "new":
			runNew()
			return
		}
	}

	// 1. cmsc.json の読み込み（通常のビルド処理）
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 2. 出力データの初期化
	data := ResponseData{
		All:        []Post{},
		ByCategory: make(map[string][]string),
		ByTag:      make(map[string][]string),
	}

	categoryNames := taxonomyNames(config.Categories)
	tagNames := taxonomyNames(config.Tags)

	categoryImages := make(map[string]string, len(config.Categories))
	for _, c := range config.Categories {
		categoryImages[c.Name] = c.Image
		data.ByCategory[c.Name] = []string{}
	}

	tagImages := make(map[string]string, len(config.Tags))
	for _, t := range config.Tags {
		tagImages[t.Name] = t.Image
		data.ByTag[t.Name] = []string{}
	}

	contains := func(list []string, item string) bool {
		for _, x := range list {
			if x == item {
				return true
			}
		}
		return false
	}

	// 3. Markdownディレクトリの巡回
	err = filepath.WalkDir(config.ContentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		parts := bytes.SplitN(content, []byte("---\n"), 3)
		if len(parts) < 3 {
			parts = bytes.SplitN(content, []byte("---\r\n"), 3)
			if len(parts) < 3 {
				return nil
			}
		}

		var post Post
		if err := yaml.Unmarshal(parts[1], &post); err != nil {
			fmt.Printf("Warning: Failed to parse YAML (%s): %v\n", path, err)
			return nil
		}

		relPath, _ := filepath.Rel(config.ContentDir, path)
		post.Slug = strings.TrimSuffix(relPath, filepath.Ext(relPath))

		// 本文（フロントマター以降の部分）をそのままcontentとして保持
		post.Content = strings.TrimSpace(string(parts[2]))

		data.All = append(data.All, post)

		if post.Category != "" && contains(categoryNames, post.Category) {
			data.ByCategory[post.Category] = append(data.ByCategory[post.Category], post.Slug)
		} else if post.Category != "" {
			fmt.Printf("Notice: Skipped unregistered category -> %s (%s)\n", post.Category, path)
		}

		for _, tag := range post.Tags {
			if tag != "" && contains(tagNames, tag) {
				data.ByTag[tag] = append(data.ByTag[tag], post.Slug)
			} else if tag != "" {
				fmt.Printf("Notice: Skipped unregistered tag -> %s (%s)\n", tag, path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking paths: %v\n", err)
		return
	}

	// 3.5. 作成日(date)の降順で全カテゴリのソート
	sortPostsByDateDesc(data.All)

	slugToCreatedAt := make(map[string]string, len(data.All))
	for _, p := range data.All {
		slugToCreatedAt[p.Slug] = p.CreatedAt
	}

	for cat := range data.ByCategory {
		sortSlugsByDateDesc(data.ByCategory[cat], slugToCreatedAt)
	}
	for tag := range data.ByTag {
		sortSlugsByDateDesc(data.ByTag[tag], slugToCreatedAt)
	}

	// 4. JSONへの変換と書き出し（all.json / category.json / tag.json の3ファイルに分割）
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		fmt.Printf("Error creating output_dir: %v\n", err)
		return
	}

	if err := writeJSONFile(filepath.Join(config.OutputDir, allJSONFileName), data); err != nil {
		fmt.Printf("Error writing %s: %v\n", allJSONFileName, err)
		return
	}

	categoryOutput := buildTaxonomyOutput(data.ByCategory, categoryImages)
	if err := writeJSONFile(filepath.Join(config.OutputDir, categoryJSONFileName), categoryOutput); err != nil {
		fmt.Printf("Error writing %s: %v\n", categoryJSONFileName, err)
		return
	}

	tagOutput := buildTaxonomyOutput(data.ByTag, tagImages)
	if err := writeJSONFile(filepath.Join(config.OutputDir, tagJSONFileName), tagOutput); err != nil {
		fmt.Printf("Error writing %s: %v\n", tagJSONFileName, err)
		return
	}

	fmt.Printf("Success! Exported %s, %s, %s to %s\n", allJSONFileName, categoryJSONFileName, tagJSONFileName, config.OutputDir)
}
