package kinesis

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"io/ioutil"
	"os"
	"reflect"
	"time"
)

// Credentials provides de credentials to access AWS Kinesis service.
type Credentials struct {
	Value      credentials.Value
	// expiration sets the max time the credentials are valid.
	// The value should be after the current time.
	// If expiration value is nil the credentials have no time limit.
	expiration time.Time
}

// WithEnvironmentVariables creates a new Credentials using environment variables.
// The namespaces searched are:
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY
// AWS_SESSION_TOKEN
// AWS_PROVIDER_NAME
// If any of the previous fields are not present as environment variable its value
// will be an empty string.
func WithEnvironmentVariables() *Credentials {
	return &Credentials{
		Value: credentials.Value{
			AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
			ProviderName:    os.Getenv("AWS_PROVIDER_NAME"),
		},
	}
}

// WithParameters creates a new Credentials using given parameters.
// If any of the parameters are not given its value will be an empty string.
func WithParameters(AccessKeyID, SecretAccessKey, SessionToken, ProviderName string) *Credentials {
	return &Credentials{
		Value: credentials.Value{
			AccessKeyID:     AccessKeyID,
			SecretAccessKey: SecretAccessKey,
			SessionToken:    SessionToken,
			ProviderName:    ProviderName,
		},
	}
}

// WithJSONFile creates a new Credentials using the values present in a JSON file.
// The namespaces searched are:
// access_key_id
// secret_access_key
// session_token
// provider_name
// If any of the parameters are not given its value will be an empty string.
func WithJSONFile(path string) (*Credentials, error) {
	var valuesMap map[string]interface{}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("[CREDENTIALS]: %v", err)
	}
	b, err := ioutil.ReadAll(f)
	if err := json.Unmarshal(b, &valuesMap); err != nil {
		return nil, fmt.Errorf("[CREDENTIALS]: %v", err)
	}
	value := credentials.Value{}
	for k, v := range valuesMap {
		v = v.(string)
		switch k {
		case "access_key_id":
			value.AccessKeyID = v.(string)
		case "secret_access_key":
			value.SecretAccessKey = v.(string)
		case "session_token":
			value.SessionToken = v.(string)
		case "provider_name":
			value.ProviderName = v.(string)
		}
	}
	return &Credentials{Value: value}, nil
}

// SetExpiration is a setter for the expiration field on Credentials.
// t cannot be before the current time.
func (c *Credentials) SetExpiration(t time.Time) error {
	if t.Before(time.Now()) {
		return fmt.Errorf("time %v is before now", t.String())
	}
	c.expiration = t
	return nil
}

// Retrieve is a getter for c.Value required by the Provider interface.
func (c Credentials) Retrieve() (credentials.Value, error) {
	return c.Value, nil
}

// IsExpired checks if the expiration time is before the current time returning
// true if so. If the expiration time is nil it will always return false.
func (c Credentials) IsExpired() bool {
	if !reflect.DeepEqual(c.expiration, time.Time{}) {
		if time.Now().After(c.expiration) {
			return true
		}
	}
	return false
}
