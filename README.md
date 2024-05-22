# Confuddlement

Confuddlement is a command-line tool that downloads Confluence pages and saves them as Markdown files. It uses the Confluence REST API to fetch page content and convert it to Markdown.

## Usage

### Environment Variables

The following environment must be set:

* `CONFLUENCE_DUMP_DIR`: The directory where the Markdown files will be saved.
* `CONFLUENCE_LIMIT`: The number of pages to fetch per API request.
* `CONFLUENCE_BASE_URL`: The base URL of the Confluence instance.
* `CONFLUENCE_PATH_URL`: The path URL of the Confluence space.
* `CONFLUENCE_USER`: The username to use for API authentication.
* `CONFLUENCE_API_TOKEN`: The API token to use for authentication.
* `DEBUG`: Set to `true` to enable debug logging.
* `MIN_PAGE_LENGTH`: The minimum length of a page to be considered valid.
* `SKIP_FETCHED_PAGES`: Set to `true` to skip pages that have already been fetched.

You can copy the .env.template file to .env and set the environment variables there.

### Running the Program

1. Set the environment variables as described above.
2. Run the program using the command `go run main.go` or build the program using the command `go build` and run the resulting executable.
3. The program will fetch Confluence pages and save them as Markdown files in the specified directory.

## License

This program is licensed under the MIT License.

Copyright (c) 2024, Sam McLeod
