package response

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type ResponseDetail struct {
	ID       string `json:"code"`
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
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var reg Registry
	if err := v.Unmarshal(&reg); err != nil {
		return nil, err
	}

	return reg, nil
}

func (r *ResponseRegistry) GetErrorRegistry(key string) (ResponseDetail, bool) {
	return r.lookup(r.Errors, key)
}

func (r *ResponseRegistry) GetSuccessRegistry(key string) (ResponseDetail, bool) {
	return r.lookup(r.Successes, key)
}

func (r *ResponseRegistry) lookup(reg Registry, key string) (ResponseDetail, bool) {
	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		return ResponseDetail{}, false
	}

	category := strings.ToLower(parts[0])
	name := strings.ToUpper(parts[1])

	if group, ok := reg[category]; ok {
		if detail, ok := group[name]; ok {
			return detail, true
		}
	}
	return ResponseDetail{}, false
}
