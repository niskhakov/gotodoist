package todoist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	restProjects = "https://api.todoist.com/rest/v1/projects"
	restTasks    = "https://api.todoist.com/rest/v1/tasks"
	restAuth     = "https://todoist.com/oauth/authorize"
)

type Client struct {
	client       *http.Client
	clientID     string
	clientSecret string
}

type Project struct {
	ID           int    `json:"id"`
	Color        int    `json:"color"`
	Order        int    `json:"order"`
	Name         string `json:"name"`
	CommentCount int    `json:"comment_count"`
	Shared       bool   `json:"shared"`
	Favorite     bool   `json:"favourite"`
	SyncId       int    `json:"sync_id"`
	InboxProject bool   `json:"inbox_project"`
	Url          string `json:"url"`
}

type Task struct {
	ID           int       `json:"id"`
	ProjectID    int       `json:"project_id"`
	SectionID    int       `json:"section_id"`
	Content      string    `json:"content"`
	Description  string    `json:"description"`
	Completed    bool      `json:"completed"`
	LabelIDs     []int     `json:"label_ids"`
	ParentID     int       `json:"parent_id"`
	Order        int       `json:"order"`
	Priority     int       `json:"priority"`
	Due          DueObject `json:"due"`
	Url          string    `json:"url"`
	CommentCount int       `json:"comment_count"`
	Assignee     int       `json:"assignee"`
}

type DueObject struct {
	String    string `json:"string"`
	Date      string `json:"date"`
	Recurring bool   `json:"recurring"`
	Datetime  string `json:"datetime"`
	Timezone  string `json:"timezone"`
}

func NewClient(clientID string, clientSecret string) (*Client, error) {
	if clientID == "" {
		return nil, errors.New("consumer key is empty")
	}

	if clientSecret == "" {
		return nil, errors.New("consumer key is empty")
	}

	return &Client{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		clientID:     clientID,
		clientSecret: clientSecret,
	}, nil
}

func (c *Client) doHTTP(ctx context.Context, accessToken string, method string, endpoint string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("DoHTTP invalid: %w", err)
	}

	bearer := "Bearer " + accessToken
	req.Header.Add("Authorization", bearer)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("API Error: got %d", resp.StatusCode)
		return nil, err
	}

	respB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	return respB, nil
}

func (c *Client) GetAuthorizationRequestURL(ctx context.Context) string {
	return fmt.Sprintf("%s?client_id=%s&scope=%s&state=%s", restAuth, c.clientID, "data:read", "state")
}

func (c *Client) GetProjects(ctx context.Context, accessToken string) ([]Project, error) {
	body, err := c.doHTTP(ctx, accessToken, http.MethodGet, restProjects, nil)
	if err != nil {
		return nil, err
	}
	var projects []Project
	if err = json.Unmarshal(body, &projects); err != nil {
		return nil, fmt.Errorf("unable to decode json response: %w", err)
	}

	return projects, nil
}

func (c *Client) GetTasksWithParams(ctx context.Context, accessToken string, paramStr string) ([]Task, error) {
	body, err := c.doHTTP(ctx, accessToken, http.MethodGet, restTasks+paramStr, nil)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	if err = json.Unmarshal(body, &tasks); err != nil {
		return nil, fmt.Errorf("unable to decode json response: %w", err)
	}

	return tasks, nil
}

func (c *Client) GetTasks(ctx context.Context, accessToken string) ([]Task, error) {
	return c.GetTasksWithParams(ctx, accessToken, "")
}

func (c *Client) GetTasksByProject(ctx context.Context, accessToken string, projectID int) ([]Task, error) {
	params := fmt.Sprintf("?project_id=%d", projectID)
	return c.GetTasksWithParams(ctx, accessToken, params)
}
