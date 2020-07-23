package kinesis

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"os"
	"reflect"
	"testing"
	"time"
)

const testsFilesDirectory = "./tests"

func TestCredentials_IsExpired(t *testing.T) {
	type fields struct {
		Value      credentials.Value
		expiration time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"default", fields{
			Value:      credentials.Value{},
			expiration: time.Time{},
		}, false},
		{"after", fields{
			Value:      credentials.Value{},
			expiration: time.Now().Add(1 * time.Minute),
		}, false},
		{"before", fields{
			Value:      credentials.Value{},
			expiration: time.Now().Add(-1 * time.Minute),
		}, true},
		{"now", fields{
			Value:      credentials.Value{},
			expiration: time.Now(),
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Credentials{
				Value:      tt.fields.Value,
				expiration: tt.fields.expiration,
			}
			if got := c.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentials_Retrieve(t *testing.T) {
	type fields struct {
		Value      credentials.Value
		expiration time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		want    credentials.Value
		wantErr bool
	}{
		{"default", fields{
			Value: credentials.Value{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
				ProviderName:    "ProviderName",
			},
			expiration: time.Time{},
		}, credentials.Value{
			AccessKeyID:     "AccessKeyID",
			SecretAccessKey: "SecretAccessKey",
			SessionToken:    "SessionToken",
			ProviderName:    "ProviderName",
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Credentials{
				Value:      tt.fields.Value,
				expiration: tt.fields.expiration,
			}
			got, err := c.Retrieve()
			if (err != nil) != tt.wantErr {
				t.Errorf("Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Retrieve() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentials_SetExpiration(t *testing.T) {
	type fields struct {
		Value      credentials.Value
		expiration time.Time
	}
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"default", fields{},
			args{t: time.Now().Add(1 * time.Minute)}, false},
		{"timeIsBeforeNow", fields{},
			args{t: time.Now().Add(-1 * time.Minute)}, true},
		{"timeIsNow", fields{},
			args{t: time.Now()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Credentials{
				Value:      tt.fields.Value,
				expiration: tt.fields.expiration,
			}
			if err := c.SetExpiration(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("SetExpiration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *Credentials
	}{
		{"default", map[string]string{
			"AWS_ACCESS_KEY_ID":     "AWS_ACCESS_KEY_ID",
			"AWS_SECRET_ACCESS_KEY": "AWS_SECRET_ACCESS_KEY",
			"AWS_SESSION_TOKEN":     "AWS_SESSION_TOKEN",
			"AWS_PROVIDER_NAME":     "AWS_PROVIDER_NAME",
		}, &Credentials{
			Value: credentials.Value{
				AccessKeyID:     "AWS_ACCESS_KEY_ID",
				SecretAccessKey: "AWS_SECRET_ACCESS_KEY",
				SessionToken:    "AWS_SESSION_TOKEN",
				ProviderName:    "AWS_PROVIDER_NAME",
			},
		}},
		{"missingEnvVars", map[string]string{
			"AWS_ACCESS_KEY_ID":     "AWS_ACCESS_KEY_ID",
			"AWS_SECRET_ACCESS_KEY": "AWS_SECRET_ACCESS_KEY",
		}, &Credentials{
			Value: credentials.Value{
				AccessKeyID:     "AWS_ACCESS_KEY_ID",
				SecretAccessKey: "AWS_SECRET_ACCESS_KEY",
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				if err := os.Setenv(k, v); err != nil {
					t.Errorf("error setting env vars: %v", err)
				}
			}
			if got := WithEnvironmentVariables(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithEnvironmentVariables() = %+v, want %+v", got, tt.want)
			}
			os.Clearenv()
		})
	}
}

func TestWithJSONFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Credentials
		wantErr bool
	}{
		{"default",
			args{path: fmt.Sprintf("%s/aws_credentials_test.json",
				testsFilesDirectory)},
			&Credentials{
				Value: credentials.Value{
					AccessKeyID:     "access_key_id",
					SecretAccessKey: "secret_access_key",
					SessionToken:    "session_token",
					ProviderName:    "provider_name",
				},
			}, false},
		{"fileNotFound",
			args{path: fmt.Sprintf("%s/404.json",
				testsFilesDirectory)},
			nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WithJSONFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithJSONFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithJSONFile() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWithParameters(t *testing.T) {
	type args struct {
		AccessKeyID     string
		SecretAccessKey string
		SessionToken    string
		ProviderName    string
	}
	tests := []struct {
		name string
		args args
		want *Credentials
	}{
		{"default", args{
			AccessKeyID:     "AccessKeyID",
			SecretAccessKey: "SecretAccessKey",
			SessionToken:    "SessionToken",
			ProviderName:    "ProviderName",
		}, &Credentials{
			Value: credentials.Value{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
				ProviderName:    "ProviderName",
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithParameters(tt.args.AccessKeyID, tt.args.SecretAccessKey, tt.args.SessionToken, tt.args.ProviderName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}
