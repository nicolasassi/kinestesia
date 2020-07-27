package receivers

import (
	"cloud.google.com/go/pubsub"
	"encoding/json"
	translator2 "github.com/nicolasassi/kinestesia/translator"
	"log"
	"reflect"
	"testing"
)

func TestClient_Translate(t *testing.T) {
	type fields struct {
		client     *pubsub.Client
		topics     []string
		translator *translator2.Translator
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
			client:     nil,
			topics:     nil,
			translator: translator2.NewTranslator("."),
			stream:     nil,
			errors:     nil,
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
			[]byte(`{"aa":"a","cd":1}`), false},
		{"keyInSliceOfMap", fields{
			client:     nil,
			topics:     nil,
			translator: translator2.NewTranslator("."),
			stream:     nil,
			errors:     nil,
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
			[]byte(`{"aa":"a","chi":1}`), false},
		{"valueInSlice", fields{
			client:     nil,
			topics:     nil,
			translator: translator2.NewTranslator("."),
			stream:     nil,
			errors:     nil,
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
							"j": []int{1,2,3,4,5},
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
				"c.h.j.[1]": "j",
			},
			sep: ".",
		},
			[]byte(`{"aa":"a","chi":1,"j":2}`), false},
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
			c.SetTranslation(tt.reference.ref, tt.reference.sep)
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