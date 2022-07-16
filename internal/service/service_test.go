package service

import (
	"context"
	"github.com/adam-putland/divido-cli/internal/models"
	util "github.com/adam-putland/divido-cli/internal/util/github"
	"github.com/adam-putland/divido-cli/internal/util/github/mock"
	"github.com/google/go-github/v45/github"
	"net/http"
	"reflect"
	"testing"
)

func TestService_GetServiceLatest(t *testing.T) {

	config := models.Config{Github: models.GithubConfig{
		Org: "test",
	}}

	tests := []struct {
		name        string
		s           Service
		serviceName string
		want        *models.Release
		wantErr     bool
	}{
		{

			name: "service_found",
			s: Service{
				gh: &util.GithubClient{
					Client: github.NewClient(mock.NewMockedHTTPClient(
						mock.WithRequestMatch(
							mock.GetReposReleasesLatestByOwnerByRepo,
							github.RepositoryRelease{
								Name:    github.String("foobar"),
								TagName: github.String("v1.0.0"),
								HTMLURL: github.String("url"),
								Body:    github.String("body"),
							},
						))),
				},
				config: &config,
			},
			serviceName: "foobar",
			want: &models.Release{
				Name:      "foobar",
				Version:   "v1.0.0",
				Changelog: "body",
				URL:       "url",
			},
			wantErr: false,
		},
		{
			name: "service_not_found",
			s: Service{
				gh: &util.GithubClient{
					Client: github.NewClient(mock.NewMockedHTTPClient(
						mock.WithRequestMatchHandler(
							mock.GetReposReleasesLatestByOwnerByRepo,
							http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								mock.WriteError(
									w,
									http.StatusInternalServerError,
									"github went wrong",
								)
							}),
						))),
				},
				config: &config,
			},
			serviceName: "foobar",
			want:        nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.s.GetLatest(context.Background(), tt.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatest() error = %s, wantErr %v", err.Error(), tt.wantErr)

			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLatest() got = %v, want %v", got, tt.want)
			}
		})
	}
}
