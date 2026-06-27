package entity

type JsonNames struct {
	All      string
	Category string
	Tag      string
}

type R2Config struct {
	Endpoint   string `json:"endpoint"`
	BucketName string `json:"bucketName"`
	BaseUrl    string `json:"baseUrl"`
}

// 設定ファイルの構造
type BriteConfig struct {
	Schema     string   `json:"$schema"`
	ArticleDir string   `json:"articleDir"`
	ImageDir   string   `json:"imageDir"`
	OutputDir  string   `json:"outputDir"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
	R2         R2Config `json:"r2"`
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
	ByCategory map[string][]PostSummary `json:"byCategory"`
	ByTag      map[string][]PostSummary `json:"byTag"`
}

type ImageCache struct {
	FilePath string `json:"filePath"`
	Size     int64  `json:"size"`
}
