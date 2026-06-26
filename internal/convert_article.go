package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/taka1156/cms-cli/internal/entity"
	"gopkg.in/yaml.v3"
)

type ConvertArticleCommand struct{}

func NewConvertArticleCommand() *ConvertArticleCommand {
	return &ConvertArticleCommand{}
}

// 記事変換（convert）コマンドの処理
func (c *ConvertArticleCommand) Convert(jsonNames entity.JsonNames) {

	// 1. cmsc.json の読み込み（通常のビルド処理）
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 出力データの初期化
	data := &entity.ResponseData{
		All:        []entity.Post{},
		ByCategory: make(map[string]entity.TaxonomyEntry),
		ByTag:      make(map[string]entity.TaxonomyEntry),
	}

	// カテゴリごとの記事一覧を初期化
	// 画象パスも合わせて保持するため、TaxonomyEntry構造体を使う
	categoryImages := make(map[string]string, len(config.Categories))
	data, categoryImages = initializedTaxonomyDefinition(config.Categories, categoryImages, data)

	// タグごとの記事一覧を初期化
	// 画象パスも合わせて保持するため、TaxonomyEntry構造体を使う
	tagImages := make(map[string]string, len(config.Tags))
	data, tagImages = initializedTaxonomyDefinition(config.Tags, tagImages, data)

	categoryNames := taxonomyNames(config.Categories)
	tagNames := taxonomyNames(config.Tags)

	// article_dir配下のMarkdownファイルを再帰的に探索し、記事データを読み込む
	data, err = walkMarkdownFiles(config.ArticleDir, data, config, categoryNames, tagNames)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 基準となる日付を保持するため、slug -> created_at のマップを作成
	slugToCreatedAt := make(map[string]string, len(data.All))
	for _, p := range data.All {
		slugToCreatedAt[p.Summary.Slug] = p.Summary.CreatedAt
	}

	for name := range data.ByCategory {
		sortSlugsByDateDesc(data.ByCategory[name].Summaries, slugToCreatedAt)
	}
	for name := range data.ByTag {
		sortSlugsByDateDesc(data.ByTag[name].Summaries, slugToCreatedAt)
	}

	// JSONへの変換と書き出し（all.json / category.json / tag.json の3ファイルに分割）
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		fmt.Printf("Error creating output_dir: %v\n", err)
		return
	}

	if err := writeJSONFile(filepath.Join(config.OutputDir, jsonNames.All), data); err != nil {
		fmt.Printf("Error writing %s: %v\n", jsonNames.All, err)
		return
	}

	categoryOutput := buildTaxonomyOutput(data.ByCategory, categoryImages)
	if err := writeJSONFile(filepath.Join(config.OutputDir, jsonNames.Category), categoryOutput); err != nil {
		fmt.Printf("Error writing %s: %v\n", jsonNames.Category, err)
		return
	}

	tagOutput := buildTaxonomyOutput(data.ByTag, tagImages)
	if err := writeJSONFile(filepath.Join(config.OutputDir, jsonNames.Tag), tagOutput); err != nil {
		fmt.Printf("Error writing %s: %v\n", jsonNames.Tag, err)
		return
	}

	fmt.Printf("Success! Exported %s, %s, %s to %s\n", jsonNames.All, jsonNames.Category, jsonNames.Tag, config.OutputDir)
}

// カテゴリ/タグの初期化（画像パスの保持と、空の記事一覧を作成）
func initializedTaxonomyDefinition(taxonomies []entity.TaxonomyDefinition, images map[string]string, data *entity.ResponseData) (*entity.ResponseData, map[string]string) {
	for _, t := range taxonomies {
		images[t.Name] = t.Image
		data.ByTag[t.Name] = entity.TaxonomyEntry{
			Image:     "",
			Summaries: []entity.PostSummary{},
		}
	}
	return data, images
}

// 全記事を走査し、summaryとcontentを読み込む。カテゴリ/タグごとの記事一覧も作成する。
func walkMarkdownFiles(contentDir string, data *entity.ResponseData, config entity.CMSConfig, categoryNames, tagNames []string) (*entity.ResponseData, error) {
	contains := func(list []string, item string) bool {
		for _, x := range list {
			if x == item {
				return true
			}
		}
		return false
	}

	// 3. Markdownディレクトリの巡回
	err := filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
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

		var post entity.Post
		if err := yaml.Unmarshal(parts[1], &post.Summary); err != nil {
			fmt.Printf("Warning: Failed to parse YAML (%s): %v\n", path, err)
			return nil
		}

		relPath, _ := filepath.Rel(contentDir, path)
		post.Summary.Slug = strings.TrimSuffix(relPath, filepath.Ext(relPath))

		// 本文（フロントマター以降の部分）をそのままcontentとして保持
		post.Content = strings.TrimSpace(string(parts[2]))

		data.All = append(data.All, post)

		if post.Summary.Category != "" && contains(categoryNames, post.Summary.Category) {
			for _, c := range config.Categories {
				if c.Name == post.Summary.Category {
					entry := data.ByCategory[c.Name]
					entry.Image = c.Image
					entry.Summaries = append(entry.Summaries, post.Summary)
					data.ByCategory[c.Name] = entry
					break
				}
			}
		} else if post.Summary.Category != "" {
			fmt.Printf("Notice: Skipped unregistered category -> %s (%s)\n", post.Summary.Category, path)
		}

		for _, tag := range post.Summary.Tags {
			if tag != "" && contains(tagNames, tag) {
				for _, t := range config.Tags {
					if t.Name == tag {
						entry := data.ByTag[t.Name]
						entry.Image = t.Image
						entry.Summaries = append(entry.Summaries, post.Summary)
						data.ByTag[t.Name] = entry
						break
					}
				}
			} else if tag != "" {
				fmt.Printf("Notice: Skipped unregistered tag -> %s (%s)\n", tag, path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking paths: %v\n", err)
		return nil, &json.InvalidUnmarshalError{}
	}

	// 3.5. 作成日(date)の降順で全カテゴリのソート
	sortPostsByDateDesc(data.All)

	return data, nil
}

// {名前: [articles,...]} と {名前: image} を合成して、
// category.json / tag.json 用の {名前: {image, articles}} 構造を組み立てる
func buildTaxonomyOutput(taxonomyEntries map[string]entity.TaxonomyEntry, imagesByName map[string]string) map[string]entity.TaxonomyEntry {
	output := make(map[string]entity.TaxonomyEntry, len(taxonomyEntries))
	for name, entries := range taxonomyEntries {
		output[name] = entity.TaxonomyEntry{
			Image:     imagesByName[name],
			Summaries: entries.Summaries,
		}
	}
	return output
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

// slug配列を、対応するcreated_atを基準に降順（新しい記事が先頭）でソートする。
// slugToCreatedAtに存在しない/パース不能なslugは最も古い扱いとして末尾に回す。
func sortSlugsByDateDesc(articles []entity.PostSummary, slugToCreatedAt map[string]string) {
	sort.SliceStable(articles, func(i, j int) bool {
		ti, errI := time.Parse(time.RFC3339, slugToCreatedAt[articles[i].Slug])
		tj, errJ := time.Parse(time.RFC3339, slugToCreatedAt[articles[j].Slug])

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

// dateを基準に降順（新しい記事が先頭）でソートする。
// パース不能なdateは最も古い扱いとして末尾に回す。
func sortPostsByDateDesc(posts []entity.Post) {
	sort.SliceStable(posts, func(i, j int) bool {
		ti, errI := time.Parse(time.RFC3339, posts[i].Summary.CreatedAt)
		tj, errJ := time.Parse(time.RFC3339, posts[j].Summary.CreatedAt)

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
