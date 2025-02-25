package semscholar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPClient abstracts the Do method so that any client (e.g., http.Client) can be used.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is the base client for interacting with Semantic Scholar APIs.
type Client struct {
	BaseURL    string
	HTTPClient HTTPClient
}

// NewClient creates a new Semantic Scholar API client.
func NewClient(baseURL string, client HTTPClient) *Client {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: client,
	}
}

/***************************************
 *          Graph API Endpoints        *
 ***************************************/

// ----- Author Endpoints -----

// Author represents an author's details returned by the API.
type Author struct {
	AuthorID     string   `json:"authorId"`
	Name         string   `json:"name"`
	URL          string   `json:"url,omitempty"`
	Affiliations []string `json:"affiliations,omitempty"`
	HIndex       int      `json:"hIndex,omitempty"`
	PaperCount   int      `json:"paperCount,omitempty"`
	Papers       []Paper  `json:"papers,omitempty"`
}

// GetAuthor retrieves details for a single author using their author ID.
func (c *Client) GetAuthor(authorID, fields string) (*Author, error) {
	endpoint := fmt.Sprintf("%s/author/%s", c.BaseURL, authorID)
	if fields != "" {
		endpoint = fmt.Sprintf("%s?fields=%s", endpoint, url.QueryEscape(fields))
	}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetAuthor: unexpected status code %d", resp.StatusCode)
	}
	var author Author
	if err := json.NewDecoder(resp.Body).Decode(&author); err != nil {
		return nil, err
	}
	return &author, nil
}

// AuthorBatchRequest represents the payload for batch retrieval of authors.
type AuthorBatchRequest struct {
	IDs []string `json:"ids"`
}

// GetAuthorsBatch retrieves details for multiple authors at once.
func (c *Client) GetAuthorsBatch(ids []string, fields string) ([]Author, error) {
	endpoint := fmt.Sprintf("%s/author/batch", c.BaseURL)
	if fields != "" {
		endpoint = fmt.Sprintf("%s?fields=%s", endpoint, url.QueryEscape(fields))
	}
	reqBody, err := json.Marshal(AuthorBatchRequest{IDs: ids})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GetAuthorsBatch: unexpected status code %d, body: %s", resp.StatusCode, string(body))
	}
	var authors []Author
	if err := json.NewDecoder(resp.Body).Decode(&authors); err != nil {
		return nil, err
	}
	return authors, nil
}

// AuthorSearchResponse represents the response from searching for authors.
type AuthorSearchResponse struct {
	Total  int      `json:"total"`
	Offset int      `json:"offset"`
	Next   int      `json:"next,omitempty"`
	Data   []Author `json:"data"`
}

// SearchAuthors searches for authors by name.
func (c *Client) SearchAuthors(query string, offset, limit int, fields string) (*AuthorSearchResponse, error) {
	endpoint := fmt.Sprintf("%s/author/search?query=%s&offset=%d&limit=%d", c.BaseURL, url.QueryEscape(query), offset, limit)
	if fields != "" {
		endpoint = fmt.Sprintf("%s&fields=%s", endpoint, url.QueryEscape(fields))
	}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SearchAuthors: unexpected status code %d", resp.StatusCode)
	}
	var result AuthorSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AuthorPapersResponse represents the response when fetching an author's papers.
type AuthorPapersResponse struct {
	Total  int     `json:"total"`
	Offset int     `json:"offset"`
	Next   int     `json:"next,omitempty"`
	Data   []Paper `json:"data"`
}

// GetAuthorPapers retrieves papers associated with a specific author.
func (c *Client) GetAuthorPapers(authorID string, offset, limit int, fields string) (*AuthorPapersResponse, error) {
	endpoint := fmt.Sprintf("%s/author/%s/papers?offset=%d&limit=%d", c.BaseURL, authorID, offset, limit)
	if fields != "" {
		endpoint = fmt.Sprintf("%s&fields=%s", endpoint, url.QueryEscape(fields))
	}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetAuthorPapers: unexpected status code %d", resp.StatusCode)
	}
	var result AuthorPapersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ----- Paper Endpoints -----

// Paper represents the details of a research paper.
type Paper struct {
	PaperID         string                 `json:"paperId"`
	CorpusID        int                    `json:"corpusId,omitempty"`
	Title           string                 `json:"title"`
	Abstract        string                 `json:"abstract,omitempty"`
	URL             string                 `json:"url,omitempty"`
	Venue           string                 `json:"venue,omitempty"`
	PublicationDate string                 `json:"publicationDate,omitempty"`
	CitationCount   int                    `json:"citationCount,omitempty"`
	ReferenceCount  int                    `json:"referenceCount,omitempty"`
	Authors         []Author               `json:"authors,omitempty"`
	FieldsOfStudy   []string               `json:"fieldsOfStudy,omitempty"`
	IsOpenAccess    bool                   `json:"isOpenAccess,omitempty"`
	OpenAccessPdf   map[string]interface{} `json:"openAccessPdf,omitempty"`
	// Additional fields can be added as needed.
}

// AutocompletePaper returns minimal paper information for autocomplete purposes.
func (c *Client) AutocompletePaper(query string) ([]Paper, error) {
	endpoint := fmt.Sprintf("%s/paper/autocomplete?query=%s", c.BaseURL, url.QueryEscape(query))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AutocompletePaper: unexpected status code %d", resp.StatusCode)
	}
	var papers []Paper
	if err := json.NewDecoder(resp.Body).Decode(&papers); err != nil {
		return nil, err
	}
	return papers, nil
}

// PaperBatchRequest represents the request payload for batch paper retrieval.
type PaperBatchRequest struct {
	IDs []string `json:"ids"`
}

// GetPapersBatch retrieves details for multiple papers in a single call.
func (c *Client) GetPapersBatch(ids []string, fields string) ([]Paper, error) {
	endpoint := fmt.Sprintf("%s/paper/batch", c.BaseURL)
	if fields != "" {
		endpoint = fmt.Sprintf("%s?fields=%s", endpoint, url.QueryEscape(fields))
	}
	reqBody, err := json.Marshal(PaperBatchRequest{IDs: ids})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetPapersBatch: unexpected status code %d", resp.StatusCode)
	}
	var papers []Paper
	if err := json.NewDecoder(resp.Body).Decode(&papers); err != nil {
		return nil, err
	}
	return papers, nil
}

// PaperSearchResponse represents the response structure for paper search endpoints.
type PaperSearchResponse struct {
	Total  int     `json:"total"`
	Offset int     `json:"offset"`
	Next   int     `json:"next,omitempty"`
	Data   []Paper `json:"data"`
}

// SearchPapers performs a relevance-ranked search for papers.
func (c *Client) SearchPapers(query string, offset, limit int, fields string, filters map[string]string) (*PaperSearchResponse, error) {
	params := url.Values{}
	params.Add("query", query)
	params.Add("offset", fmt.Sprintf("%d", offset))
	params.Add("limit", fmt.Sprintf("%d", limit))
	if fields != "" {
		params.Add("fields", fields)
	}
	for k, v := range filters {
		params.Add(k, v)
	}
	endpoint := fmt.Sprintf("%s/paper/search?%s", c.BaseURL, params.Encode())
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SearchPapers: unexpected status code %d", resp.StatusCode)
	}
	var result PaperSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BulkSearchPapers performs a bulk search for papers without full relevance ranking.
func (c *Client) BulkSearchPapers(query, token, fields, sort, publicationTypes string, additionalFilters map[string]string) (*PaperSearchResponse, error) {
	params := url.Values{}
	if query != "" {
		params.Add("query", query)
	}
	if token != "" {
		params.Add("token", token)
	}
	if fields != "" {
		params.Add("fields", fields)
	}
	if sort != "" {
		params.Add("sort", sort)
	}
	if publicationTypes != "" {
		params.Add("publicationTypes", publicationTypes)
	}
	for k, v := range additionalFilters {
		params.Add(k, v)
	}
	endpoint := fmt.Sprintf("%s/paper/search/bulk?%s", c.BaseURL, params.Encode())
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("BulkSearchPapers: unexpected status code %d", resp.StatusCode)
	}
	var result PaperSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MatchSearchPapers performs a minimal match search for papers.
func (c *Client) MatchSearchPapers(query, fields, publicationTypes string, additionalFilters map[string]string) (*PaperSearchResponse, error) {
	params := url.Values{}
	params.Add("query", query)
	if fields != "" {
		params.Add("fields", fields)
	}
	if publicationTypes != "" {
		params.Add("publicationTypes", publicationTypes)
	}
	for k, v := range additionalFilters {
		params.Add(k, v)
	}
	endpoint := fmt.Sprintf("%s/paper/search/match?%s", c.BaseURL, params.Encode())
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MatchSearchPapers: unexpected status code %d", resp.StatusCode)
	}
	var result PaperSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

/***************************************
 *    Recommendations API Endpoints    *
 ***************************************/

// RecommendationRequest represents the request payload for getting paper recommendations.
type RecommendationRequest struct {
	Positive []string `json:"positive"`
	Negative []string `json:"negative,omitempty"`
}

// RecommendationResponse represents the response for paper recommendations.
type RecommendationResponse struct {
	RecommendedPapers []Paper `json:"recommendedPapers"`
}

// GetRecommendations retrieves recommended papers given positive (and optionally negative) paper IDs.
func (c *Client) GetRecommendations(reqData RecommendationRequest, limit int, fields string) (*RecommendationResponse, error) {
	endpoint := fmt.Sprintf("%s/papers?limit=%d", c.BaseURL, limit)
	if fields != "" {
		endpoint = fmt.Sprintf("%s&fields=%s", endpoint, url.QueryEscape(fields))
	}
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetRecommendations: unexpected status code %d", resp.StatusCode)
	}
	var result RecommendationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRecommendationsForPaper retrieves recommended papers based on a single positive paper.
func (c *Client) GetRecommendationsForPaper(paperID, from string, limit int, fields string) (*RecommendationResponse, error) {
	endpoint := fmt.Sprintf("%s/papers/forpaper/%s?limit=%d", c.BaseURL, paperID, limit)
	if from != "" {
		endpoint = fmt.Sprintf("%s&from=%s", endpoint, url.QueryEscape(from))
	}
	if fields != "" {
		endpoint = fmt.Sprintf("%s&fields=%s", endpoint, url.QueryEscape(fields))
	}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetRecommendationsForPaper: unexpected status code %d", resp.StatusCode)
	}
	var result RecommendationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

/***************************************
 *         Datasets API Endpoints      *
 ***************************************/

// ReleaseMetadata represents metadata describing a particular release.
type ReleaseMetadata struct {
	ReleaseID string           `json:"release_id"`
	README    string           `json:"README"`
	Datasets  []DatasetSummary `json:"datasets"`
}

// DatasetSummary represents a summary of a dataset available in a release.
type DatasetSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	README      string `json:"README"`
}

// DatasetMetadata represents detailed metadata about a specific dataset.
type DatasetMetadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	README      string   `json:"README"`
	Files       []string `json:"files"`
}

// DatasetDiff represents the diff information for a dataset between two releases.
type DatasetDiff struct {
	FromRelease string   `json:"from_release"`
	ToRelease   string   `json:"to_release"`
	UpdateFiles []string `json:"update_files"`
	DeleteFiles []string `json:"delete_files"`
}

// DatasetDiffList represents the list of diffs required to update a dataset.
type DatasetDiffList struct {
	Dataset      string        `json:"dataset"`
	StartRelease string        `json:"start_release"`
	EndRelease   string        `json:"end_release"`
	Diffs        []DatasetDiff `json:"diffs"`
}

// GetDatasetDiffs retrieves the incremental diff links for updating a dataset between releases.
func (c *Client) GetDatasetDiffs(startReleaseID, endReleaseID, datasetName string) (*DatasetDiffList, error) {
	endpoint := fmt.Sprintf("%s/diffs/%s/to/%s/%s", c.BaseURL, url.PathEscape(startReleaseID), url.PathEscape(endReleaseID), url.PathEscape(datasetName))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetDatasetDiffs: unexpected status code %d", resp.StatusCode)
	}
	var diffList DatasetDiffList
	if err := json.NewDecoder(resp.Body).Decode(&diffList); err != nil {
		return nil, err
	}
	return &diffList, nil
}

// GetReleases retrieves a list of available release IDs.
func (c *Client) GetReleases() ([]string, error) {
	endpoint := fmt.Sprintf("%s/release/", c.BaseURL)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetReleases: unexpected status code %d", resp.StatusCode)
	}
	var releases []string
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	return releases, nil
}

// GetRelease retrieves metadata for a specific release.
func (c *Client) GetRelease(releaseID string) (*ReleaseMetadata, error) {
	endpoint := fmt.Sprintf("%s/release/%s", c.BaseURL, url.PathEscape(releaseID))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetRelease: unexpected status code %d", resp.StatusCode)
	}
	var releaseMeta ReleaseMetadata
	if err := json.NewDecoder(resp.Body).Decode(&releaseMeta); err != nil {
		return nil, err
	}
	return &releaseMeta, nil
}

// GetDataset retrieves metadata and download links for a specific dataset within a release.
func (c *Client) GetDataset(releaseID, datasetName string) (*DatasetMetadata, error) {
	endpoint := fmt.Sprintf("%s/release/%s/dataset/%s", c.BaseURL, url.PathEscape(releaseID), url.PathEscape(datasetName))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetDataset: unexpected status code %d", resp.StatusCode)
	}
	var datasetMeta DatasetMetadata
	if err := json.NewDecoder(resp.Body).Decode(&datasetMeta); err != nil {
		return nil, err
	}
	return &datasetMeta, nil
}
