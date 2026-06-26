package entity

type JsonNames struct {
	All      string
	Category string
	Tag      string
}

// categories/tagsの1要素（名前+紐づく画像パス）
type TaxonomyDefinition struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// 設定ファイルの構造
type CMSConfig struct {
	Schema     string               `json:"$schema"`
	ArticleDir string               `json:"articleDir"`
	ImagesDir  string               `json:"imagesDir"`
	OutputDir  string               `json:"outputDir"`
	Categories []TaxonomyDefinition `json:"categories"`
	Tags       []TaxonomyDefinition `json:"tags"`
}

type PostSummary struct {
	Slug      string   `json:"slug" yaml:"-"`
	Title     string   `json:"title" yaml:"title"`
	Category  string   `json:"category" yaml:"category"`
	Tags      []string `json:"tags" yaml:"tags"`
	CreatedAt string   `json:"created_at" yaml:"created_at"`
	UpdatedAt string   `json:"updated_at" yaml:"updated_at"`
}

// 各記事のデータ構造
type Post struct {
	Summary PostSummary `json:"summary" yaml:"summary"`
	Content string      `json:"content" yaml:"-"`
}

// 最終出力のデータ構造（byCategory/byTagはslug参照のみで本文の重複を避ける）
type ResponseData struct {
	All        []Post                   `json:"all"`
	ByCategory map[string]TaxonomyEntry `json:"byCategory"`
	ByTag      map[string]TaxonomyEntry `json:"byTag"`
}

// category.json / tag.json の1エントリ（画像情報 + 紐づく記事slug一覧）
type TaxonomyEntry struct {
	Image     string        `json:"image"`
	Summaries []PostSummary `json:"summaries"`
}
