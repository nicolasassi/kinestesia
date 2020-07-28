package receivers

import (
	"cloud.google.com/go/pubsub"
	"encoding/json"
	"github.com/nicolasassi/kinestesia/translator"
	"log"
	"reflect"
	"testing"
)

func TestClient_Translate(t *testing.T) {
	type fields struct {
		client     *pubsub.Client
		topics     []string
		translator *translator.Translator
		stream     chan []byte
		errors     chan error
	}
	type args struct {
		b []byte
	}
	type reference struct {
		ref map[string]string
		sep string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		reference reference
		want      []byte
		wantErr   bool
	}{
		{"default", fields{
			client: nil,
			topics: nil,
			stream: nil,
			errors: nil,
		},
			args{b: func() []byte {
				values := map[string]interface{}{
					"a": "a",
					"b": "b",
					"c": map[string]interface{}{
						"d": 1,
					},
				}
				b, err := json.Marshal(values)
				if err != nil {
					log.Fatal(err)
				}
				return b
			}(),
			}, reference{
				ref: map[string]string{
					"a":   "aa",
					"c.d": "cd",
				},
				sep: ".",
			},
			[]byte(`{"aa":"a","b":"b","cd":1}`), false},
		{"keyInSliceOfMap", fields{
			client: nil,
			topics: nil,
			stream: nil,
			errors: nil,
		},
			args{b: func() []byte {
				values := map[string]interface{}{
					"a": "a",
					"b": "b",
					"c": []map[string]interface{}{
						{"d": 1},
						{"e": 2},
						{"f": 3},
						{"h": map[string]interface{}{
							"i": 1,
						}},
					},
				}
				b, err := json.Marshal(values)
				if err != nil {
					log.Fatal(err)
				}
				return b
			}(),
			}, reference{
				ref: map[string]string{
					"a":     "aa",
					"c.h.i": "chi",
				},
				sep: ".",
			},
			[]byte(`{"aa":"a","b":"b","chi":1}`), false},
		{"valueInSlice", fields{
			client: nil,
			topics: nil,
			stream: nil,
			errors: nil,
		},
			args{b: func() []byte {
				values := map[string]interface{}{
					"a": "a",
					"b": "b",
					"c": []map[string]interface{}{
						{"d": 1},
						{"e": 2},
						{"f": 3},
						{"h": map[string]interface{}{
							"i": 1,
							"j": []int{1, 2, 3, 4, 5},
						}},
					},
				}
				b, err := json.Marshal(values)
				if err != nil {
					log.Fatal(err)
				}
				return b
			}(),
			}, reference{
				ref: map[string]string{
					"a":         "aa",
					"c.h.i":     "chi",
					"c.h.j.[1]": "j",
				},
				sep: ".",
			},
			[]byte(`{"aa":"a","b":"b","chi":1,"j":2}`), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				client:     tt.fields.client,
				topics:     tt.fields.topics,
				translator: tt.fields.translator,
				stream:     tt.fields.stream,
				errors:     tt.fields.errors,
			}
			c.SetTranslation(translator.NewTranslator(tt.reference.ref, tt.reference.sep))
			got, err := c.Translate(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Translate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Translate() got = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestClient_TranslateWithFilter(t *testing.T) {
	type fields struct {
		client     *pubsub.Client
		name       string
		topics     []string
		translator *translator.Translator
		stream     chan []byte
		sent       chan struct{}
		errors     chan error
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{"default", fields{
			translator: func() *translator.Translator {
				t := new(translator.Translator)
				t.AddFilterRule("operation", "==", "INSERT")
				return t
			}(),
		}, args{
			func() []byte {
				raw := `{"operation":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
				return []byte(raw)
			}()},
			nil,
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				client:     tt.fields.client,
				name:       tt.fields.name,
				topics:     tt.fields.topics,
				translator: tt.fields.translator,
				stream:     tt.fields.stream,
				sent:       tt.fields.sent,
				errors:     tt.fields.errors,
			}
			got, err := c.Translate(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Translate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Translate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
