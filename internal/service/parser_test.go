package service

import (
	"github.com/adam-putland/divido-cli/internal/models"
	"reflect"
	"testing"
)

func TestParser_Replace(t *testing.T) {
	type fields struct {
		version string
		service string
		yaml    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		want    string
	}{
		{name: "new_version", fields: struct {
			version string
			service string
			yaml    string
		}{version: "v1.0.7", service: "applicantCommunicationApi", yaml: `services:
  applicantCommunicationApi:
    serviceVersion: v1.0.6
`}, wantErr: false, want: `services:
  applicantCommunicationApi:
    serviceVersion: v1.0.7
`},
		{name: "keep_comments", fields: struct {
			version string
			service string
			yaml    string
		}{version: "v1.0.7", service: "applicantCommunicationApi", yaml: `services:	
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.6
`}, wantErr: false, want: `services:
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.7
`},
		{name: "lower_version_does_update", fields: struct {
			version string
			service string
			yaml    string
		}{version: "v1.0.4", service: "applicantCommunicationApi", yaml: `services:	
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.6
`}, wantErr: false, want: `services:
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.4
`},
		{name: "service_not_found", fields: struct {
			version string
			service string
			yaml    string
		}{version: "v1.0.4", service: "test", yaml: `services:	
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.6
  # api
  api:
    serviceVersion: v1.0.4
`}, wantErr: false, want: `services:
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.6
  # api
  api:
    serviceVersion: v1.0.4
  test:
    serviceVersion: v1.0.4
`},
		{name: "string_version", fields: struct {
			version string
			service string
			yaml    string
		}{version: "1234", service: "applicantCommunicationApi", yaml: `services:	
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: v1.0.6
`}, wantErr: false, want: `services:
  # applicant-communication-api
  applicantCommunicationApi:
    serviceVersion: "1234"
`},
		{name: "internal_repo", fields: struct {
			version string
			service string
			yaml    string
		}{version: "1234", service: "application-api", yaml: `services:	
  application-api:
    podspec:
      containers:
        fpm:
          tag: v1.17.3
        nginx:
          env:
            DIVIDO_NGINX_CLIENT_MAX_BODY_SIZE: 10m
`}, wantErr: false, want: `services:
  application-api:
    podspec:
      containers:
        fpm:
          tag: "1234"
        nginx:
          env:
            DIVIDO_NGINX_CLIENT_MAX_BODY_SIZE: 10m
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p := NewParser([]byte(tt.fields.yaml))

			_, err := p.Load()
			if err != nil {
				t.Error(err)
			}

			services := make(map[string]*models.Service)

			services[tt.fields.service] = &models.Service{HLMName: tt.fields.service, Release: models.Release{Version: tt.fields.version}}

			if err = p.Replace(services); (err != nil) != tt.wantErr {
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
