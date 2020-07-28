package translator

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var indexPattern = regexp.MustCompile(`^\[(\d+)]$`)

type filterRule struct {
	arg1     string
	modifier string
	arg2     interface{}
}

type Translator struct {
	translationKeys  map[int][]string
	translationValue map[int]string
	sep              string
	rules            []filterRule
}

type ObjectJSON map[string]interface{}

func NewTranslator(reference map[string]string, sep string) *Translator {
	if sep == "" {
		sep = "."
	}
	t := &Translator{
		translationKeys:  map[int][]string{},
		translationValue: map[int]string{},
		sep:              sep,
	}
	var cntr int
	for k, v := range reference {
		t.translationKeys[cntr] = strings.Split(k, t.sep)
		t.translationValue[cntr] = v
		cntr++
	}
	return t
}

func (t *Translator) AddFilterRule(arg1, modifier string, arg2 interface{}) {
	switch modifier {
	case "type_is", "type_is_not":
		if arg2 == "list" {
			arg2 = "slice"
		}
	}
	t.rules = append(t.rules, filterRule{
		arg1:     arg1,
		modifier: modifier,
		arg2:     arg2,
	})
}

func (t *Translator) translate(m interface{}, keys []string) interface{} {
	key := keys[0]
	rt := reflect.TypeOf(m)
	switch rt.Kind() {
	case reflect.Slice:
		if indexPattern.MatchString(key) {
			var value interface{}
			finds := indexPattern.FindStringSubmatch(key)
			index, _ := strconv.Atoi(finds[1])
			value = m.([]interface{})[index]
			if len(keys) == 1 {
				return value
			}
			if len(keys) > 1 {
				keys = keys[1:]
			}
			return t.translate(value, keys)
		}
		for _, values := range m.([]interface{}) {
			switch values.(type) {
			case map[string]interface{}:
				for k, v := range values.(map[string]interface{}) {
					if k == key {
						if len(keys) == 1 {
							return v
						}
						if len(keys) > 1 {
							keys = keys[1:]
						}
						return t.translate(v, keys)
					}
				}
			}
		}
	case reflect.Map:
		value := m.(map[string]interface{})[key]
		if len(keys) == 1 {
			return value
		}
		if len(keys) > 1 {
			keys = keys[1:]
		}
		return t.translate(value, keys)
	}
	return nil
}

func (t Translator) Translate(obj ObjectJSON) *ObjectJSON {
	resp := ObjectJSON{}
	var keyTranslated bool
	for k, v := range obj {
		keyTranslated = false
		for tk, tv := range t.translationKeys {
			if tv[0] == k {
				keyTranslated = true
				if len(tv) == 1 {
					resp[t.translationValue[tk]] = v
					continue
				}
				key := t.translationValue[tk]
				translated := t.translate(v, tv[1:])
				resp[key] = translated
				continue
			}
		}
		if !keyTranslated{
			resp[k] = v
		}
	}
	for k, v := range obj {
		if !t.filter(k, v) {
			return nil
		}
	}
	return &resp
}

func (t Translator) filter(key string, value interface{}) bool {
	resp := true
	for _, rule := range t.rules {
		if rule.arg1 == key {
			resp = false
			switch rule.modifier {
			case "==":
				if reflect.DeepEqual(rule.arg2, value) {
					return true
				}
			case "!=":
				if !reflect.DeepEqual(rule.arg2, value) {
					return true
				}
			case "type_is":
				if reflect.TypeOf(rule.arg2).Kind() != reflect.String {
					return false
				}
				if reflect.TypeOf(value).Kind().String() == rule.arg2.(string) {
					return true
				}
			case "type_is_not":
				if reflect.TypeOf(rule.arg2).Kind() != reflect.String {
					return false
				}
				if reflect.TypeOf(value).Kind().String() != rule.arg2.(string)  {
					return true
				}
			}
		}
	}
	return resp
}
