package models

import (
	"fmt"
	"regexp"
	"strconv"
)

func RefFloat64(value float64) *float64 {
	return &value
}
func RefInt(value int) *int {
	return &value
}
func RefString(value string) *string {
	return &value
}

func RefStringAsInt(value string) *int {

	if value == "" {
		return nil
	}

	v, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		return nil
	}
	vv := int(v)

	return &vv
}

func RefStringAsFloat(value string) *float64 {

	if value == "" {
		return nil
	}

	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}

	return &v
}

func regexToMap(value, pattern string) (result map[string]string, err error) {

	result = map[string]string{}
	reg := regexp.MustCompile(pattern)

	if !reg.MatchString(value) {
		err = fmt.Errorf("value should match %s", pattern)
		return
	}

	finds := reg.FindStringSubmatch(value)

	for i, k := range reg.SubexpNames() {
		if k != "" {
			result[k] = finds[i]
		}
	}

	return

}
