package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/joho/godotenv"
)

type ListContentResponse struct {
	Links   ListLinks   `json:"_links"`
	Limit   int        `json:"limit"`
	Size    int        `json:"size"`
	Start   int        `json:"start"`
	Results []ListResult `json:"results"`
}

type ListLinks struct {
	Base    string `json:"base"`
	Context string `json:"context"`
	Next    string `json:"next"`
	Self    string `json:"self"`
}

type ResultLinks struct {
	Self   string `json:"self"`
	Tinyui string `json:"tinyui"`
	Editui string `json:"editui"`
	Webui  string `json:"webui"`
}

type ListResult struct {
	ID     string      `json:"id"`
	Type   string      `json:"type"`
	Status string      `json:"status"`
	Title  string      `json:"title"`
	Links  ResultLinks `json:"_links"`
}

type ApiResponse struct {
	Page struct {
		Results []struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Status string `json:"status"`
			Title  string `json:"title"`
			Links  struct {
				Self   string `json:"self"`
				Tinyui string `json:"tinyui"`
				Editui string `json:"editui"`
				Webui  string `json:"webui"`
			} `json:"_links"`
		} `json:"results"`
	} `json:"page"`
}

type FetchContentResponse struct {
	Body struct {
		Storage struct {
			Value string `json:"value"`
		} `json:"storage"`
	} `json:"body"`
}

func main() {
	version := "0.1.0"
	fmt.Printf("Confuddlement %s\n", version)

	loadEnvVars()

	gitDir := getEnv("CONFLUENCE_DUMP_DIR")
	if gitDir == "" {
		panic("CONFLUENCE_DUMP_DIR not set (e.g. ./confluence_dump)")
	}

	debug := getEnv("DEBUG") == "true"

	limit := getEnv("CONFLUENCE_LIMIT")
	if limit == "" {
		fmt.Println("CONFLUENCE_LIMIT not set, defaulting to 25")
		limit = "25"
	}

	base := getEnv("CONFLUENCE_BASE_URL")
	if base == "" {
		panic("CONFLUENCE_BASE_URL not set (e.g. https://mycompany.atlassian.net/wiki)")
	}
	next := getEnv("CONFLUENCE_PATH_URL")
	if next == "" {
		panic("CONFLUENCE_PATH_URL not set (e.g. /rest/api/space/myteam/content)")
	}

	limitedUrl := fmt.Sprintf("%s%s?limit=%s", base, next, limit)

	state, err := loadState()
	if err == nil {
		if state.Links.Next != "" {
			limitedUrl = state.Links.Next
		}
		if state.Size > 0 {
			for _, result := range state.Results {
				if debug {
					fmt.Println(result.Title)
				}
			}
		}
	}

	apiResponse, err := listContent(limitedUrl)
	if err != nil {
		panic(err)
	}

	for _, result := range apiResponse.Page.Results {
		if debug {
			fmt.Println(result.Title)
		}
	}

	savePagesToLocalFS := getEnv("SAVE_PAGES_TO_LOCAL_FS")
	if savePagesToLocalFS == "true" {
		savePagesToLocalFSDir := getEnv("CONFLUENCE_DUMP_DIR")
		if savePagesToLocalFSDir == "" {
			savePagesToLocalFSDir = "confluence_dump"
		}
		err := os.MkdirAll(savePagesToLocalFSDir, fs.FileMode(0755))
		if err != nil {
			panic(err)
		}
		for _, result := range apiResponse.Page.Results {
			pageContent, err := fetchPageContent(result.ID, debug)
			if err != nil {
				if err.Error() == "empty ExportView URL for page" {
					fmt.Printf("Skipping empty page %s\n", result.Title)
					continue
				}

				continue
			}
			if pageContent == "" {
				fmt.Printf("Skipping empty page %s\n", result.Title)
				continue
			}

			minPageLength, _ := strconv.Atoi(getEnv("MIN_PAGE_LENGTH"))
			if len(pageContent) < minPageLength {
				fmt.Printf("Skipping page %s, less than %d characters\n", result.Title, minPageLength)
				continue
			}

			markdown, err := convertHtmlToMarkdown(pageContent)
			if err != nil {
				fmt.Printf("Error converting page %s to markdown: %v\n", result.Title, err)
				continue
			}
			markdownFile := fmt.Sprintf("%s/%s.md", savePagesToLocalFSDir, result.ID)
			err = os.WriteFile(markdownFile, []byte(markdown), fs.FileMode(0644))
			if err != nil {
				fmt.Printf("Error writing page %s to file: %v\n", result.Title, err)
				continue
			}
			fmt.Printf("Saved page %s to %s\n", result.Title, markdownFile)
		}
	}
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func loadEnvVars() {
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			panic(err)
		}
	}
}

func fetchPageContent(id string, debug bool) (string, error) {
	// check the state and resume from where we left off if possible
	state, err := loadState()
	if err == nil {
		for _, result := range state.Results {
			if result.ID == id && os.Getenv("SKIP_FETCHED_PAGES") == "true" {
				if debug {
					fmt.Printf("Skipping page %s, already fetched\n", result.Title)
				}
				return "", nil
			}
		}
	}

	url := fmt.Sprintf("%s/rest/api/content/%s?expand=body.storage,metadata,extensions.inlineProperties,metadata.labels,metadata.properties,extensions.tableCells", os.Getenv("CONFLUENCE_BASE_URL"), id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(os.Getenv("CONFLUENCE_USER"), os.Getenv("CONFLUENCE_API_TOKEN"))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	var pageContent struct {
		Body struct {
			Storage struct {
				Value string `json:"value"`
			} `json:"storage"`
		} `json:"body"`
		Metadata struct {
			InlineProperties struct {
				Extensions struct {
					InlineProperties []struct {
						Body struct {
							Storage struct {
								Value string `json:"value"`
							} `json:"storage"`
						} `json:"body"`
					} `json:"inlineProperties"`
				} `json:"extensions"`
			} `json:"inlineProperties"`
			Labels struct {
				Results []struct {
					Prefix string `json:"prefix"`
					Name   string `json:"name"`
				} `json:"results"`
			} `json:"labels"`
			Properties struct {
				Results []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"results"`
			} `json:"properties"`
		} `json:"metadata"`
		Extensions struct {
			TableCells struct {
				Results []struct {
					Id   string `json:"id"`
					Type string `json:"type"`
					Body struct {
						Storage struct {
							Value string `json:"value"`
						} `json:"storage"`
					} `json:"body"`
				} `json:"results"`
			} `json:"tableCells"`
		} `json:"extensions"`
	}

	err = json.NewDecoder(resp.Body).Decode(&pageContent)
	if err != nil {
		return "", err
	}

	content := pageContent.Body.Storage.Value
	for _, prop := range pageContent.Metadata.InlineProperties.Extensions.InlineProperties {
		content += prop.Body.Storage.Value
	}

	for _, tableCell := range pageContent.Extensions.TableCells.Results {
		content += tableCell.Body.Storage.Value
	}

	// add the state to the list of fetched pages
	state, err = loadState()
	if err != nil {
		state = &ListContentResponse{}
	}
	state.Results = append(state.Results, ListResult{
		ID:     id,
		Type:   "page",
		Status: "current",
		Links:  ResultLinks{},
	})
	err = saveState(state.Results, "")
	if err != nil {
		return "", err
	}

	return content, nil
}

func convertHtmlToMarkdown(html string) (string, error) {
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(html)
	if err != nil {
		return "", err
	}

	return markdown, nil
}

func listContent(url string) (*ApiResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(os.Getenv("CONFLUENCE_USER"), os.Getenv("CONFLUENCE_API_TOKEN"))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, errors.New("API request failed with status code 403")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	var apiResponse ApiResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func saveState(pages []ListResult, next string) error {
	currentState := ListContentResponse{
		Links: ListLinks{
			Base:    os.Getenv("CONFLUENCE_BASE_URL"),
			Context: os.Getenv("CONFLUENCE_PATH_URL"),
			Next:    next,
			Self:    "",
		},
		Size:    len(pages),
		Start:   0,
		Results: pages,
	}

	currentStateJson, err := json.Marshal(currentState)
	if err != nil {
		return err
	}

	err = os.WriteFile("state.json", currentStateJson, fs.FileMode(0644))
	if err != nil {
		return err
	}

	return nil
}

func loadState() (*ListContentResponse, error) {
	stateJson, err := os.ReadFile("state.json")
	if err != nil {
		return nil, err
	}

	var state ListContentResponse
	err = json.Unmarshal(stateJson, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}
