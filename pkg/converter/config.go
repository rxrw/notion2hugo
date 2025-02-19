package converter

type Config struct {
	DatabaseID string `json:"databaseID"`
	Content    struct {
		Folder    string `json:"folder"`
		Archetype string `json:"archetype"`
	} `json:"content"`
	Storage struct {
		Type  string `json:"type"`
		Local struct {
			Path      string `json:"path"`
			URLPrefix string `json:"urlPrefix"`
		} `json:"local"`
		S3 struct {
			Bucket     string `json:"bucket"`
			Region     string `json:"region"`
			PathPrefix string `json:"pathPrefix"`
			URLPrefix  string `json:"urlPrefix"`
		} `json:"s3"`
	} `json:"storage"`
	Notion struct {
		Status struct {
			Draft     string `json:"draft"`
			Ready     string `json:"ready"`
			Published string `json:"published"`
			ToDelete  string `json:"toDelete"`
			Deleted   string `json:"deleted"`
		} `json:"status"`
		CategoryMap map[string]string `json:"categoryMap"`
		Properties  struct {
			Title       string `json:"title"`
			Categories  string `json:"categories"`
			Tags        string `json:"tags"`
			Status      string `json:"status"`
			Description string `json:"description"`
			Author      string `json:"author"`
			MetaTitle   string `json:"metaTitle"`
			Slug        string `json:"slug"`
			Toc         string `json:"toc"`
			Comments    string `json:"comments"`
			Weight      string `json:"weight"`
		} `json:"properties"`
	} `json:"notion"`
	Image struct {
		MaxWidth int      `json:"max_width"`
		Quality  int      `json:"quality"`
		Formats  []string `json:"formats"`
	} `json:"image"`
}
