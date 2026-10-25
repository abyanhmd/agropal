package response

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ResponseDetail struct {
	ResCode  string `json:"code"`
	HTTPCode int    `json:"http_code"`
	Message  string `json:"message"`
}

type Registry map[string]map[string]ResponseDetail

type ResponseRegistry struct {
	Errors    Registry
	Successes Registry
}

func NewResponseRegistry(
	errors Registry,
	successes Registry,
) *ResponseRegistry {
	return &ResponseRegistry{
		Errors:    errors,
		Successes: successes,
	}
}

func LoadRegistries(errorPath, successPath string) (*ResponseRegistry, error) {
	errs, err := loadFile(errorPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load error registry: %w", err)
	}

	succs, err := loadFile(successPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load success registry: %w", err)
	}

	return &ResponseRegistry{
		Errors:    errs,
		Successes: succs,
	}, nil
}

func loadFile(path string) (Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, err
	}

	return reg, nil
}

func (r *ResponseRegistry) GetErrorRegistry(key error) (ResponseDetail, bool) {
	return r.lookup(r.Errors, key.Error())
}

func (r *ResponseRegistry) GetSuccessRegistry(key string) (ResponseDetail, bool) {
	return r.lookup(r.Successes, key)
}

func (r *ResponseRegistry) lookup(reg Registry, key string) (ResponseDetail, bool) {
	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		return ResponseDetail{}, false
	}

	category := strings.ToUpper(parts[0])
	name := strings.ToUpper(parts[1])

	if group, ok := reg[category]; ok {
		if detail, ok := group[name]; ok {
			return detail, true
		}
	}
	return ResponseDetail{}, false
}
