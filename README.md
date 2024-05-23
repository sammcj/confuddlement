# Confuddlement

Confuddlement is a command-line tool that downloads Confluence pages and saves them as Markdown files.
It uses the Confluence REST API to fetch page content and convert it to Markdown.

```plain
$ go run ./main.go

Confuddlement 0.2.0
Spaces: [COOLTEAM, MANAGEMENT]
Fetching content from space COOLTEAM

COOLTEAM (Totally Cool Team Homepage)
Retrospectives
Decision log
Development Onboarding
Saved page COOLTEAM - Feature List to ./confluence_dump/COOLTEAM - Feature List.md
Skipping page 7. Support, less than 300 characters

MANAGEMENT (Department of Overhead and Bureaucracy)
Painful Change Management
Illogical Diagrams
Saved page ./confluence_dump/Painful Change Management.md
Saved page Illogical Diagrams to ./confluence_dump/Ilogical Diagrams.md
```

## Usage

### Running the Program

1. Copy [.env.template](.env.template) to `.env` and update the environment variables.
2. Run the program using the command `go run main.go` or build the program using the command `go build` and run the resulting executable.
3. The program will fetch Confluence pages and save them as Markdown files in the specified directory.

### Environment Variables

The following environment must be set:

* `CONFLUENCE_DUMP_DIR`: The directory where the Markdown files will be saved.
* `CONFLUENCE_LIMIT`: The number of pages to fetch per API request.
* `CONFLUENCE_BASE_URL`: The base URL of the Confluence instance.
* `CONFLUENCE_USER`: The username to use for API authentication.
* `CONFLUENCE_SPACES`: The space keys to fetch pages from, separated by commas.
* `CONFLUENCE_API_TOKEN`: The API token to use for authentication.
* `DELETE_PREVIOUS_DUMP`: Set to `true` to delete the previous dump directory (and state) before fetching pages.
* `MIN_PAGE_LENGTH`: The minimum length of a page to be considered valid.
* `SKIP_FETCHED_PAGES`: Set to `true` to skip pages that have already been fetched.
* `DEBUG`: Set to `true` to enable debug logging.

## License

This program is licensed under the MIT License.

Copyright (c) 2024, Sam McLeod
