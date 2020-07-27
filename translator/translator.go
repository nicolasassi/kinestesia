package translator

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var indexPattern = regexp.MustCompile(`^\[(\d+)]$`)

type rule struct {
	arg1     string
	modifier string
	arg2     string
}

type Translator struct {
	translationKeys  map[int][]string
	translationValue map[int]string
	sep              string
	rules            []rule
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

func (t *Translator) AddRule(arg1, modifier, arg2 string) {
	t.rules = append(t.rules, rule{
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
	for k, v := range obj {
		for tk, tv := range t.translationKeys {
			if tv[0] == k {
				if len(tv) == 1 {
					resp[t.translationValue[tk]] = v
					break
				}
				resp[t.translationValue[tk]] = t.translate(v, tv[1:])
			}
		}
	}
	return &resp
}
