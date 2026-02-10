package database

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type solrClient struct {
	baseURL string;
	http *http.Client;
}

type SolrResponse struct {
    Response struct {
        Docs []struct {
            ID []string `json:"track_id"`
        } `json:"docs"`
    } `json:"response"`

    NextCursorMark string `json:"nextCursorMark"`
}


func NewSolrClient(baseURL string) *solrClient {
	return &solrClient{
        baseURL: strings.TrimRight(baseURL, "/"),
        http: &http.Client{
            Timeout: 3 * time.Second,
        },
    }
}

func (c *solrClient) Search(
    ctx context.Context,
    query string,
    limit int,
    cursor string,
) ([]string, string, error) {

    if cursor == "" {
        cursor = "*"
    }

    if limit <= 0 || limit > 50 {
        limit = 10
    }

    params := url.Values{}
    params.Set("q", query)
    params.Set("defType", "edismax")
    params.Set("qf", "title^5 artists^2")
    params.Set("pf", "title^15")
    params.Set("mm", "2")
    params.Set("sort", "score desc, id asc")
    params.Set("rows", strconv.Itoa(limit))
    params.Set("cursorMark", cursor)
    params.Set("fl", "track_id")
    params.Set("wt", "json")

    req, err := http.NewRequestWithContext(
        ctx,
        http.MethodPost,
        c.baseURL+"/select",
        strings.NewReader(params.Encode()),
    )
    if err != nil {
        return nil, "", err
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := c.http.Do(req)
    if err != nil {
        return nil, "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, "", fmt.Errorf("solr error: %s", string(body))
    }

    var sr SolrResponse
    if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
        return nil, "", err
    }

    ids := make([]string, 0, len(sr.Response.Docs))
	for _, d := range sr.Response.Docs {
		if len(d.ID) > 0 {
			ids = append(ids, d.ID[0])
		}
	}

    return ids, sr.NextCursorMark, nil
}
