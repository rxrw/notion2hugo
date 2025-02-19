package pkg

type Config struct {
	DatabaseID string        `json:"database_id"`
	Content    ContentConfig `json:"content"`
	Storage    StorageConfig `json:"storage"`
	Notion     NotionConfig  `json:"notion"`
	Image      ImageConfig   `json:"image"`
}

type ContentConfig struct {
	Folder       string `json:"folder"`
	Archetype    string `json:"archetype"`
	DateFilename bool   `json:"date_filename"`
}

type StorageConfig struct {
	Type  string       `json:"type"`
	Local LocalStorage `json:"local"`
	S3    S3Storage    `json:"s3"`
}

type LocalStorage struct {
	Path      string `json:"path"`
	URLPrefix string `json:"url_prefix"`
}

type S3Storage struct {
	Bucket     string `json:"bucket"`
	Region     string `json:"region"`
	PathPrefix string `json:"path_prefix"`
	URLPrefix  string `json:"url_prefix"`
}

type NotionConfig struct {
	Status struct {
		Draft     string `json:"draft"`
		Ready     string `json:"ready"`
		Published string `json:"published"`
	} `json:"status"`
	CategoryMap map[string]string `json:"category_map"`
}

type ImageConfig struct {
	MaxWidth int      `json:"max_width"`
	Quality  int      `json:"quality"`
	Formats  []string `json:"formats"`
}
