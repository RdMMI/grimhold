package open5e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const BaseURL = "https://api.open5e.com/v1"

type Action struct {
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	AttackBonus int    `json:"attack_bonus"`
	DamageDice  string `json:"damage_dice"`
}

type Monster struct {
	Slug            string   `json:"slug"`
	Name            string   `json:"name"`
	Size            string   `json:"size"`
	Type            string   `json:"type"`
	Alignment       string   `json:"alignment"`
	ArmorClass      int      `json:"armor_class"`
	HitPoints       int      `json:"hit_points"`
	HitDice         string   `json:"hit_dice"`
	ChallengeRating string   `json:"challenge_rating"`
	CrFloat         float64  `json:"cr"`
	Actions         []Action `json:"actions"`
}

type MonsterListResponse struct {
	Count    int       `json:"count"`
	Next     *string   `json:"next"`
	Previous *string   `json:"previous"`
	Results  []Monster `json:"results"`
}

type Client struct {
	baseURL    string
	httpClient http.Client
	cache      *Cache
}

func NewClient(timeout, cacheTTL time.Duration) *Client {
	return NewClientWithBaseURL(BaseURL, timeout, cacheTTL)
}

func NewClientWithBaseURL(baseURL string, timeout, cacheTTL time.Duration) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: http.Client{Timeout: timeout},
		cache:      NewCache(cacheTTL),
	}
}

func (c *Client) getBytes(url string) ([]byte, error) {
	if data, ok := c.cache.Get(url); ok {
		return data, nil
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	c.cache.Add(url, body)
	return body, nil
}

func (c *Client) GetMonster(slug string) (*Monster, error) {
	url := fmt.Sprintf("%s/monsters/%s/", c.baseURL, slug)
	data, err := c.getBytes(url)
	if err != nil {
		return nil, fmt.Errorf("fetching monster %q: %w", slug, err)
	}

	var m Monster
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("decoding monster %q: %w", slug, err)
	}
	return &m, nil
}

func (c *Client) ListMonsters(pageURL string) (*MonsterListResponse, error) {
	if pageURL == "" {
		pageURL = c.baseURL + "/monsters/"
	}
	data, err := c.getBytes(pageURL)
	if err != nil {
		return nil, fmt.Errorf("fetching monster list: %w", err)
	}

	var list MonsterListResponse
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("decoding monster list: %w", err)
	}
	return &list, nil
}

func (c *Client) PrefetchMonsters(slugs []string) {
	var wg sync.WaitGroup
	for _, slug := range slugs {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			_, _ = c.GetMonster(s)
		}(slug)
	}
	wg.Wait()
}
