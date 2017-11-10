package util

import "github.com/aws/aws-sdk-go/aws"

// ConvertStringSlice []string -> []*string
func ConvertStringSlice(s []string) []*string {
	var ret []*string
	for _, t := range s {
		ret = append(ret, aws.String(t))
	}
	return ret
}

// ConvertStringMap map[string]string -> map[string]*string
func ConvertStringMap(s map[string]string) map[string]*string {
	m := map[string]*string{}
	for k, v := range s {
		m[k] = aws.String(v)
	}
	return m
}
