package translator

import (
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
		name string
		fields fields
		args args
		want interface{}
	}{
		{"default",
			fields{},
			args{
			m:    map[string]interface{}{
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
			m:    map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": []int{1,2,3},
					},
				},
			},
			keys: []string{"a", "b", "c"},
		},
			[]int{1,2,3},
		},
		{"sliceIndex",
			fields{},
			args{
			m:    map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": []int{1,2,3},
					},
				},
			},
			keys: []string{"a", "b", "c", "[2]"},
		},
			3,
		},
		{"sliceOfMap",
			fields{},
			args{
			m:    map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": []map[string]interface{}{
							{"a1": 1},
							{"b1": 2},
							{"c1": 3},
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
			if got := t.Translate(tt.args.m, tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}