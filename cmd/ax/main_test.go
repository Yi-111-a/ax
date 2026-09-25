package main

import (
	"context"
	"errors"
	"net"
	"slices"
	"testing"

	"github.com/google/ax/pkg/apis/v1alpha1"
	"google.golang.org/grpc"
)

var errGetDispatch = errors.New("get dispatch")

type getTestServer struct {
	v1alpha1.UnimplementedAXServer
	calls []string
}

func (s *getTestServer) ListTasks(context.Context, *v1alpha1.ListTasksRequest) (*v1alpha1.ListTasksResponse, error) {
	s.calls = append(s.calls, "ListTasks")
	return nil, errGetDispatch
}

func (s *getTestServer) GetTask(_ context.Context, req *v1alpha1.GetTaskRequest) (*v1alpha1.Task, error) {
	s.calls = append(s.calls, "GetTask:"+req.GetName())
	return nil, errGetDispatch
}

func (s *getTestServer) ListWorkspaces(context.Context, *v1alpha1.ListWorkspacesRequest) (*v1alpha1.ListWorkspacesResponse, error) {
	s.calls = append(s.calls, "ListWorkspaces")
	return nil, errGetDispatch
}

func (s *getTestServer) GetWorkspace(_ context.Context, req *v1alpha1.GetWorkspaceRequest) (*v1alpha1.Workspace, error) {
	s.calls = append(s.calls, "GetWorkspace:"+req.GetName())
	return nil, errGetDispatch
}

func (s *getTestServer) ListModels(context.Context, *v1alpha1.ListModelsRequest) (*v1alpha1.ListModelsResponse, error) {
	s.calls = append(s.calls, "ListModels")
	return nil, errGetDispatch
}

func (s *getTestServer) GetModel(_ context.Context, req *v1alpha1.GetModelRequest) (*v1alpha1.Model, error) {
	s.calls = append(s.calls, "GetModel:"+req.GetName())
	return nil, errGetDispatch
}

func TestRunGetDispatchesSingularAndPluralResources(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCall []string
	}{
		{name: "singular task list", args: []string{"task"}, wantCall: []string{"ListTasks"}},
		{name: "plural task list", args: []string{"tasks"}, wantCall: []string{"ListTasks"}},
		{name: "singular task get", args: []string{"task", "example"}, wantCall: []string{"GetTask:example"}},
		{name: "plural task get", args: []string{"tasks", "example"}, wantCall: []string{"GetTask:example"}},
		{name: "singular workspace list", args: []string{"workspace"}, wantCall: []string{"ListWorkspaces"}},
		{name: "plural workspace list", args: []string{"workspaces"}, wantCall: []string{"ListWorkspaces"}},
		{name: "singular workspace get", args: []string{"workspace", "example"}, wantCall: []string{"GetWorkspace:example"}},
		{name: "plural workspace get", args: []string{"workspaces", "example"}, wantCall: []string{"GetWorkspace:example"}},
		{name: "singular model list", args: []string{"model"}, wantCall: []string{"ListModels"}},
		{name: "plural model list", args: []string{"models"}, wantCall: []string{"ListModels"}},
		{name: "singular model get", args: []string{"model", "example"}, wantCall: []string{"GetModel:example"}},
		{name: "plural model get", args: []string{"models", "example"}, wantCall: []string{"GetModel:example"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("listen: %v", err)
			}

			grpcServer := grpc.NewServer()
			service := &getTestServer{}
			v1alpha1.RegisterAXServer(grpcServer, service)
			go func() {
				_ = grpcServer.Serve(listener)
			}()
			t.Cleanup(grpcServer.Stop)

			if err := runGet(listener.Addr().String(), "default", tt.args); err == nil {
				t.Fatal("runGet error = nil, want dispatch error")
			}
			if !slices.Equal(service.calls, tt.wantCall) {
				t.Fatalf("calls = %v, want %v", service.calls, tt.wantCall)
			}
		})
	}
}
