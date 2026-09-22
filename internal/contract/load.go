package contract

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

func Load(r io.Reader) (Service, error) {
	var service Service
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)
	if err := decoder.Decode(&service); err != nil {
		return Service{}, fmt.Errorf("decode service definition: %w", err)
	}
	return service, nil
}
