# notion-blog

`notion-blog` allows you to use Notion as a CMS for pages built with hugo. You can use it as a cli, Docker container, or even automate your blog repo to update itself with the Github Action.


## Requisites

- Notion database for your articles.
- Notion API secret token.
- Hugo powered blog.

## Usage

### CLI

The cli shows the executable flags when using flag `—help`.

```bash
$> notion-blog —help
```

### Binary

The binary looks for a config file called `notionblog.config.json` in the directory where it is executed. You can see the example config in [notionblog.config.json](notionblog.config.json).


### Github Action

To use it as a Github Action, you can follow the example of the repository in [.github/worflows/notion.yml](.github/workflows/notion.yml).


### As a GitHub Action

```yaml
- uses: rxrw/notion2hugo@v1
  with:
    notion_secret: ${{ secrets.NOTION_SECRET }}
    database_id: 'your-database-id'
    storage_type: 's3'  # optional
    s3_bucket: 'your-bucket'  # if using S3
    s3_endpoint: 'https://your-endpoint'  # if using S3
```

### Using Docker

```bash
docker run -v $(pwd):/workspace \
  -e NOTION_SECRET=your-secret \
  -e DATABASE_ID=your-database-id \
  ghcr.io/rxrw/notion2hugo:latest
```

With S3 storage:
```bash
docker run -v $(pwd):/workspace \
  -e NOTION_SECRET=your-secret \
  -e DATABASE_ID=your-database-id \
  -e STORAGE_TYPE=s3 \
  -e S3_BUCKET=your-bucket \
  -e S3_ENDPOINT=your-endpoint \
  -e S3_REGION=your-region \
  ghcr.io/rxrw/notion2hugo:latest
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| NOTION_SECRET | Notion API token | Required |
| DATABASE_ID | Notion database ID | Required |
| CONTENT_FOLDER | Hugo content folder | content/posts |
| IMAGES_FOLDER | Images storage path | static/images |
| IMAGES_PREFIX | Images URL prefix | /images |
| STORAGE_TYPE | Storage type (local/s3) | local |
| S3_BUCKET | S3 bucket name | Required for S3 |
| S3_REGION | S3 region | Required for S3 |
| S3_ENDPOINT | S3 endpoint URL | Optional |
| S3_PATH_PREFIX | S3 path prefix | images |
| S3_URL_PREFIX | S3 URL prefix | Required for S3 |

## Compilation

This is only required if you are not going to use the repo as a Github Action. The compilation is simple as Golang installs everything for you.

```bash
go build -o ./bin/main cmd/main/main.go
```

You can compile any form of the app (cli or binary) by compiling the main file in any of the packages in `cmd/`.

