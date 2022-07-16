package service

import (
	"github.com/adam-putland/divido-cli/internal/models"
	"reflect"
	"testing"
)

func TestParser_Replace(t *testing.T) {

	tests := []struct {
		services map[string]*models.Service
		yaml     string
		name     string
		wantErr  bool
		want     string
	}{
		{
			name:     "new_version",
			services: map[string]*models.Service{"applicantCommunicationApi": {HLMName: "applicantCommunicationApi", Release: models.Release{Version: "v1.0.7"}}},
			yaml: `services:
  applicantCommunicationApi:
    serviceVersion: v1.0.6
`,
			wantErr: false,
			want: `services:
  applicantCommunicationApi:
    serviceVersion: v1.0.7
`},
		{name: "keep_comments", services: map[string]*models.Service{"applicantCommunicationApi": {HLMName: "applicantCommunicationApi", Release: models.Release{Version: "v1.0.7"}}}, yaml: `services:	
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.6
`, wantErr: false, want: `services:
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.7
`},
		{name: "lower_version_does_update", services: map[string]*models.Service{"applicantCommunicationApi": {HLMName: "applicantCommunicationApi", Release: models.Release{Version: "v1.0.4"}}},
			yaml: `services:	
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.6
`, wantErr: false, want: `services:
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.4
`}, {
			name:     "service_not_found",
			services: map[string]*models.Service{"test": {HLMName: "test", Release: models.Release{Version: "v1.0.4"}}},
			yaml: `services:	
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.6
  # api
  api:
    serviceVersion: v1.0.4
`, wantErr: false, want: `services:
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.6
  # api
  api:
    serviceVersion: v1.0.4
  test:
    serviceVersion: v1.0.4
`},
		{
			name:     "string_version",
			services: map[string]*models.Service{"applicantCommunicationApi": {HLMName: "applicantCommunicationApi", Release: models.Release{Version: "1234"}}},
			yaml: `services:	
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.6
`,
			wantErr: false,
			want: `services:
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: "1234"
`},
		{name: "internal_repo", services: map[string]*models.Service{"application-api": {HLMName: "application-api", Release: models.Release{Version: "1234"}}},
			yaml: `services:	
  application-api:
    podspec:
      containers:
        fpm:
          tag: v1.17.3
        nginx:
          env:
            DIVIDO_NGINX_CLIENT_MAX_BODY_SIZE: 10m
  x:
    serviceVersion: v1.0.9
`,
			wantErr: false,
			want: `services:
  application-api:
    podspec:
      containers:
        fpm:
          tag: "1234"
        nginx:
          env:
            DIVIDO_NGINX_CLIENT_MAX_BODY_SIZE: 10m
  x:
    serviceVersion: v1.0.9
`},
		{
			name:     "new_version_no_top_service_elem",
			services: map[string]*models.Service{"applicantCommunicationApi": {HLMName: "applicantCommunicationApi", Release: models.Release{Version: "v1.0.7"}}},
			yaml: `applicantCommunicationApi:
  serviceVersion: v1.0.6
`,
			wantErr: false,
			want: `applicantCommunicationApi:
  serviceVersion: v1.0.7
`},
		{
			name: "multiple_services_update",
			services: map[string]*models.Service{"applicantCommunicationApi": {HLMName: "applicantCommunicationApi", Release: models.Release{Version: "v1.0.5"}},
				"test": {HLMName: "test", Release: models.Release{Version: "v1.0.7"}}},
			yaml: `applicantCommunicationApi:
  serviceVersion: v1.0.4
test:
  serviceVersion: v1.0.6
x:
  serviceVersion: v1.0.9
`,
			wantErr: false,
			want: `applicantCommunicationApi:
  serviceVersion: v1.0.5
test:
  serviceVersion: v1.0.7
x:
  serviceVersion: v1.0.9
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p := NewParser([]byte(tt.yaml))

			_, err := p.Load()
			if err != nil {
				t.Error(err)
			}

			if err = p.Replace(tt.services); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

			}
			got, err := p.GetContent()
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("GetContent() got = %v, want = %v", string(got), tt.want)
			}
		})
	}
}
