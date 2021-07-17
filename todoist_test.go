package todoist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}

func newClient(t *testing.T, statusCode int, response []byte) *Client {
	return &Client{
		client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewReader(response)),
				}, nil
			}),
		},
		clientID:     "clientID",
		clientSecret: "clientSecret",
	}
}

func TestNewClient(t *testing.T) {
	type args struct {
		clientID     string
		clientSecret string
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{
			name: "Client must be with key",
			args: args{
				clientID: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.clientID, tt.args.clientSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_doHTTP(t *testing.T) {
	type fields struct {
		statusCode int
		response   []byte
	}
	type args struct {
		ctx         context.Context
		accessToken string
		method      string
		endpoint    string
		body        io.Reader
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		nonEmpty bool
		wantErr  bool
	}{
		{
			name: "Standard scenario",
			fields: fields{
				statusCode: 200,
				response:   []byte("response"),
			},
			args: args{
				ctx:         context.Background(),
				accessToken: "",
				method:      http.MethodGet,
				endpoint:    restProjects,
				body:        nil,
			},
			nonEmpty: true,
		},
		{
			name: "Empty url",
			fields: fields{
				statusCode: 400,
				response:   []byte(""),
			},
			args: args{
				ctx:      context.Background(),
				method:   http.MethodGet,
				endpoint: "",
				body:     nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, tt.fields.statusCode, tt.fields.response)
			got, err := c.doHTTP(tt.args.ctx, tt.args.accessToken, tt.args.method, tt.args.endpoint, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.doHTTP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.nonEmpty && len(got) == 0 {
				t.Errorf("Client.doHTTP() empty return, want non empty")
			}
		})
	}
}

func TestClient_GetProjects(t *testing.T) {
	type fields struct {
		statusCode int
		response   []byte
	}
	type args struct {
		ctx         context.Context
		accessToken string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		isEmpty bool
		isOne   bool
		wantErr bool
	}{
		{
			name: "Get Projects",
			fields: fields{
				statusCode: 200,
				response:   createMockProjects(2),
			},
			args: args{
				ctx:         context.Background(),
				accessToken: "",
			},
			isEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, tt.fields.statusCode, tt.fields.response)

			got, err := c.GetProjects(tt.args.ctx, tt.args.accessToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetProjects() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.isEmpty && len(got) == 0 {
				t.Errorf("Client.GetProjects() = %v, want not empty, got %v", got, len(got))
			}

			if tt.isOne && len(got) != 1 {
				t.Errorf("Client.GetProjects() = %v, want one object, got %v", got, len(got))
			}
		})
	}
}

func TestClient_GetTasks(t *testing.T) {
	type fields struct {
		statusCode int
		response   []byte
	}
	type args struct {
		ctx         context.Context
		accessToken string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		nonEmpty bool
		wantErr  bool
	}{
		{
			name: "Got active tasks",
			fields: fields{
				statusCode: 200,
				response:   createMockTasks(2),
			},
			args: args{
				ctx:         context.Background(),
				accessToken: "",
			},
			nonEmpty: true,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, tt.fields.statusCode, tt.fields.response)
			got, err := c.GetTasks(tt.args.ctx, tt.args.accessToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.nonEmpty && len(got) == 0 {
				t.Errorf("Client.GetTasks() = %v, want not empty slice, got %v objects", got, len(got))
				return
			}
		})
	}
}

func TestClient_GetTasksByProject(t *testing.T) {
	type fields struct {
		statusCode int
		response   []byte
	}
	type args struct {
		ctx         context.Context
		accessToken string
		projectID   int
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		nonEmpty bool
		wantErr  bool
	}{
		{
			name: "Got active tasks",
			fields: fields{
				statusCode: 200,
				response:   createMockTasks(2),
			},
			args: args{
				ctx:         context.Background(),
				accessToken: "",
				projectID:   2187255141,
			},
			nonEmpty: true,
			wantErr:  false,
		},
		{
			name: "Got error with unauthorized project",
			fields: fields{
				statusCode: 200,
				response:   createMockProjects(0),
			},
			args: args{
				ctx:       context.Background(),
				projectID: 1,
			},
			nonEmpty: false,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, tt.fields.statusCode, tt.fields.response)
			got, err := c.GetTasksByProject(tt.args.ctx, tt.args.accessToken, tt.args.projectID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.nonEmpty && len(got) == 0 {
				t.Errorf("Client.GetTasks() = %v, want not empty slice, got %v objects", got, len(got))
				return
			}
		})
	}
}

func createMockTasks(num int) []byte {
	tasks := make([]Task, 0)
	for i := 0; i < num; i++ {
		tasks = append(tasks, Task{
			ID:        i,
			ProjectID: i,
			Content:   fmt.Sprintf("SomeContent%d", i),
		})
	}

	tasksBody, _ := json.Marshal(tasks)
	return tasksBody
}

func createMockProjects(num int) []byte {
	projects := make([]Project, 0)
	for i := 0; i < num; i++ {
		projects = append(projects, Project{
			ID:   i,
			Name: fmt.Sprintf("SomeContent%d", i),
		})
	}

	projectsBody, _ := json.Marshal(projects)

	return projectsBody
}
