package translator

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
)

func Test_Translate(t1 *testing.T) {
	type fields struct {
		translationKeys  map[int][]string
		translationValue map[int]string
		sep              string
	}
	type args struct {
		m    map[string]interface{}
		keys []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{"default",
			fields{},
			args{
				m: map[string]interface{}{
					"a": map[string]interface{}{
						"b": map[string]interface{}{
							"c": "ok",
						},
					},
				},
				keys: []string{"a", "b", "c"},
			},
			"ok",
		},
		{"slice",
			fields{},
			args{
				m: map[string]interface{}{
					"a": map[string]interface{}{
						"b": map[string]interface{}{
							"c": []interface{}{1, 2, 3},
						},
					},
				},
				keys: []string{"a", "b", "c"},
			},
			[]interface{}{1, 2, 3},
		},
		{"sliceOfString",
			fields{},
			args{
				m: map[string]interface{}{
					"a": map[string]interface{}{
						"b": map[string]interface{}{
							"c": []interface{}{"d", "e", "f"},
						},
					},
				},
				keys: []string{"a", "b", "c"},
			},
			[]interface{}{"d", "e", "f"},
		},
		{"sliceIndex",
			fields{},
			args{
				m: map[string]interface{}{
					"a": map[string]interface{}{
						"b": map[string]interface{}{
							"c": []interface{}{1, 2, 3},
						},
					},
				},
				keys: []string{"a", "b", "c", "[2]"},
			},
			3,
		},
		{"sliceIndexString",
			fields{},
			args{
				m: map[string]interface{}{
					"a": map[string]interface{}{
						"b": map[string]interface{}{
							"c": []interface{}{"d", "e", "f"},
						},
					},
				},
				keys: []string{"a", "b", "c", "[2]"},
			},
			"f",
		},
		{"sliceOfMap",
			fields{},
			args{
				m: map[string]interface{}{
					"a": map[string]interface{}{
						"b": map[string]interface{}{
							"c": []interface{}{
								map[string]interface{}{
									"a1": 1,
									"b1": 2,
									"c1": 3,
								},
							},
						},
					},
				},
				keys: []string{"a", "b", "c", "b1"},
			},
			2,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Translator{
				translationKeys:  tt.fields.translationKeys,
				translationValue: tt.fields.translationValue,
				sep:              tt.fields.sep,
			}
			if got := t.translate(tt.args.m, tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslator_filter(t1 *testing.T) {
	type fields struct {
		translationKeys  map[int][]string
		translationValue map[int]string
		sep              string
		rules            []filterRule
	}
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"equalString", fields{
			rules: []filterRule{
				{
					arg1:     "a",
					modifier: "==",
					arg2:     "b",
				},
			},
		}, args{
			key:   "a",
			value: "b",
		}, true},
		{"equalListString", fields{
			rules: []filterRule{
				{
					arg1:     "a",
					modifier: "==",
					arg2:     []interface{}{"b", "c"},
				},
			},
		}, args{
			key:   "a",
			value: []interface{}{"b", "c"},
		}, true},
		{"diffString", fields{
			rules: []filterRule{
				{
					arg1:     "a",
					modifier: "!=",
					arg2:     "b",
				},
			},
		}, args{
			key:   "a",
			value: "c",
		}, true},
		{"typeIsString", fields{
			rules: []filterRule{
				{
					arg1:     "a",
					modifier: "type_is",
					arg2:     "string",
				},
			},
		}, args{
			key:   "a",
			value: "c",
		}, true},
		{"typeIsList", fields{
			rules: []filterRule{
				{
					arg1:     "a",
					modifier: "type_is",
					arg2:     "slice",
				},
			},
		}, args{
			key:   "a",
			value: []interface{}{1, 2, 3},
		}, true},
		{"typeIsNotString", fields{
			rules: []filterRule{
				{
					arg1:     "a",
					modifier: "type_is_not",
					arg2:     "string",
				},
			},
		}, args{
			key:   "a",
			value: 1,
		}, true},
		{"multiRules_no_match", fields{
			rules: []filterRule{
				{
					arg1:     "a",
					modifier: "==",
					arg2:     "b",
				},
				{
					arg1:     "a",
					modifier: "==",
					arg2:     "c",
				},
			},
		}, args{
			key:   "a",
			value: 1,
		}, false},
		{"multiRules_one_match", fields{
			rules: []filterRule{
				{
					arg1:     "a",
					modifier: "==",
					arg2:     "b",
				},
				{
					arg1:     "a",
					modifier: "==",
					arg2:     "c",
				},
			},
		}, args{
			key:   "a",
			value: "c",
		}, true},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Translator{
				rules: tt.fields.rules,
			}
			if got := t.filter(tt.args.key, tt.args.value); got != tt.want {
				t1.Errorf("filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslator_AddFilterRule(t1 *testing.T) {
	type fields struct {
		translationKeys  map[int][]string
		translationValue map[int]string
		sep              string
		rules            []filterRule
	}
	type args struct {
		arg1     string
		modifier string
		arg2     string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   filterRule
	}{
		{"default", fields{}, args{
			arg1:     "a",
			modifier: "type_is",
			arg2:     "list",
		}, filterRule{
			arg1:     "a",
			modifier: "type_is",
			arg2:     "slice",
		}},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Translator{
				translationKeys:  tt.fields.translationKeys,
				translationValue: tt.fields.translationValue,
				sep:              tt.fields.sep,
				rules:            tt.fields.rules,
			}
			t.AddFilterRule(tt.args.arg1, tt.args.modifier, tt.args.arg2)
			if !reflect.DeepEqual(t.rules[0], tt.want) {
				t1.Errorf("AddFilterRule() = %v, want %v", t.rules[0], tt.want)
			}
		})
	}
}

func TestTranslator_Translate(t1 *testing.T) {
	type fields struct {
		reference map[string]string
		sep string
		rules []filterRule
	}
	type args struct {
		obj ObjectJSON
	}
	tests := []struct {
		name   string

		fields fields
		args   args
		want   *ObjectJSON
	}{
		{"noTranslationOrFilter", fields{
			reference: nil,
			rules:     nil,
		}, args{obj: func() map[string]interface{} {
			var m map[string]interface{}
			raw := `{"operation":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
			if err := json.Unmarshal([]byte(raw), &m); err != nil {
				log.Fatal(err)
			}
			return m
		}()},
			func() *ObjectJSON {
				obj := new(ObjectJSON)
				raw := `{"operation":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
				if err := json.Unmarshal([]byte(raw), obj); err != nil {
					log.Fatal(err)
				}
				return obj
			}(),
		},
		{"translateOneField", fields{
			reference: map[string]string{
				"operation": "op",
			},
			rules:     nil,
		}, args{obj: func() map[string]interface{} {
			var m map[string]interface{}
			raw := `{"operation":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
			if err := json.Unmarshal([]byte(raw), &m); err != nil {
				log.Fatal(err)
			}
			return m
		}()},
			func() *ObjectJSON {
				obj := new(ObjectJSON)
				raw := `{"op":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
				if err := json.Unmarshal([]byte(raw), obj); err != nil {
					log.Fatal(err)
				}
				return obj
			}(),
		},
		{"translateOneFieldAndOneInnerField", fields{
			reference: map[string]string{
				"operation": "op",
				"payload.structure.dep": "department",
			},
			rules:     nil,
		}, args{obj: func() map[string]interface{} {
			var m map[string]interface{}
			raw := `{"operation":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
			if err := json.Unmarshal([]byte(raw), &m); err != nil {
				log.Fatal(err)
			}
			return m
		}()},
			func() *ObjectJSON {
				obj := new(ObjectJSON)
				raw := `{"department":"09090","op":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00"}`
				if err := json.Unmarshal([]byte(raw), obj); err != nil {
					log.Fatal(err)
				}
				return obj
			}(),
		},
		{"translateOneFieldAndInnerField", fields{
			reference: map[string]string{
				"operation": "op",
				"payload.structure.dep": "department",
				"payload.structure.class": "class",
			},
			rules:     nil,
		}, args{obj: func() map[string]interface{} {
			var m map[string]interface{}
			raw := `{"operation":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
			if err := json.Unmarshal([]byte(raw), &m); err != nil {
				log.Fatal(err)
			}
			return m
		}()},
			func() *ObjectJSON {
				obj := new(ObjectJSON)
				raw := `{"department":"09090","class":"6565","op":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00"}`
				if err := json.Unmarshal([]byte(raw), obj); err != nil {
					log.Fatal(err)
				}
				return obj
			}(),
		},
		{"simpleFilterByFieldValueEquality", fields{
			rules:     []filterRule{
				{arg1: "operation", modifier: "==", arg2: "UPDATE"},
			},
		}, args{obj: func() map[string]interface{} {
			var m map[string]interface{}
			raw := `{"operation":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
			if err := json.Unmarshal([]byte(raw), &m); err != nil {
				log.Fatal(err)
			}
			return m
		}()},
			func() *ObjectJSON {
				obj := new(ObjectJSON)
				raw := `{"operation":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
				if err := json.Unmarshal([]byte(raw), obj); err != nil {
					log.Fatal(err)
				}
				return obj
			}(),
		},
		{"simpleFilterByFieldValueEquality", fields{
			rules:     []filterRule{
				{arg1: "operation", modifier: "==", arg2: "INSERT"},
			},
		}, args{obj: func() map[string]interface{} {
			var m map[string]interface{}
			raw := `{"operation":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
			if err := json.Unmarshal([]byte(raw), &m); err != nil {
				log.Fatal(err)
			}
			return m
		}()},
			nil,
		},
		{"multiFilterByFieldValueEquality", fields{
			rules:     []filterRule{
				{arg1: "operation", modifier: "==", arg2: "INSERT"},
				{arg1: "operation", modifier: "==", arg2: "UPDATE"},
				{arg1: "operation", modifier: "==", arg2: "DELETE"},
			},
		}, args{obj: func() map[string]interface{} {
			var m map[string]interface{}
			raw := `{"operation":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
			if err := json.Unmarshal([]byte(raw), &m); err != nil {
				log.Fatal(err)
			}
			return m
		}()},
			func() *ObjectJSON {
				obj := new(ObjectJSON)
				raw := `{"operation":"UPDATE","type":"FHNS","id":"2983736282937","version":"1","sentAt":"2020-07-27T12:01:01.833-03:00","payload":{"id":"9832987392873","condition":"NEW","entity":"9809809","structure":{"dep":"09090","sec":"0909565","subclass":"34434","class":"6565"},"quantity":99,"updatedAt":"2020-07-27T12:01:01.832-03:00"}}`
				if err := json.Unmarshal([]byte(raw), obj); err != nil {
					log.Fatal(err)
				}
				return obj
			}(),
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := NewTranslator(tt.fields.reference, tt.fields.sep)
			for _, rule := range tt.fields.rules {
				t.AddFilterRule(rule.arg1, rule.modifier, rule.arg2)
			}
			if got := t.Translate(tt.args.obj); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}
