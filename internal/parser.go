package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type YamlFile struct {
	Services map[string]interface{} `yaml:"services"`
}

type Parser struct {
	version    string
	service    string
	loadedYaml yaml.Node
}

func NewParser(version string, service string) *Parser {
	return &Parser{version: version, service: service}
}

func (p *Parser) Load(data []byte) error {
	return yaml.Unmarshal(data, &p.loadedYaml)
}

func (p Parser) Run() error {
	if len(p.loadedYaml.Content) == 0 {
		return errors.New("invalid file")
	}

	// check if loading from document file
	if p.loadedYaml.Kind == yaml.DocumentNode {
		p.loadedYaml = *p.loadedYaml.Content[0]
	}

	foundService := false
	for keyIndex := 0; keyIndex < len(p.loadedYaml.Content); keyIndex++ {
		if p.loadedYaml.Content[keyIndex].Kind == yaml.MappingNode {
			for i := 0; i < len(p.loadedYaml.Content[keyIndex].Content); i++ {
				if strings.ReplaceAll(p.loadedYaml.Content[keyIndex].Content[i].Value, " ", "") == p.service {
					foundService = true
					updatedRepo, err := p.UpdateExternalRepo(keyIndex, i)
					if err != nil {
						return err
					}
					if !updatedRepo {
						err = p.UpdateInternalRepo(keyIndex, i)
						if err != nil {
							return err
						}
					}
					break
				}
			}

			// if the service is not in the document it will be created
			if !foundService {
				nodes, err := p.CreateServiceNodes()
				if err != nil {
					return err
				}
				p.loadedYaml.Content[keyIndex].Content = append(p.loadedYaml.Content[keyIndex].Content, nodes...)
			}
		}
	}

	return nil
}

func (p *Parser) UpdateInternalRepo(keyIndex int, i int) error {
	var internalRepo Repo
	err := p.loadedYaml.Content[keyIndex].Content[i+1].Decode(&internalRepo)
	if err != nil {
		return err
	}
	if len(internalRepo.Podspec.Services) > 0 {
		for index, service := range internalRepo.Podspec.Services {
			jsonb, err := json.Marshal(service)
			if err != nil {
				fmt.Println(err)
				return err
			}
			var parsedService Service
			if err = json.Unmarshal(jsonb, &parsedService); err != nil {
				fmt.Println(err)
				return err
			}
			if parsedService.Tag != "" {
				fmt.Printf("updating service %s from %s to %s\n", p.service, parsedService.Tag, p.version)
				parsedService.Tag = p.version
				internalRepo.Podspec.Services[index] = parsedService
				break
			}

		}
		return p.loadedYaml.Content[keyIndex].Content[i+1].Encode(&internalRepo)
	}
	return nil
}

func (p *Parser) UpdateExternalRepo(keyIndex int, i int) (bool, error) {
	var externalRepo ExternalRepo
	content := p.loadedYaml.Content[keyIndex].Content[i+1]
	err := content.Decode(&externalRepo)
	if err == nil && externalRepo.Version != "" {
		fmt.Printf("updating service %s from %s to %s\n", p.service, externalRepo.Version, p.version)
		externalRepo.Version = p.version
		err = p.loadedYaml.Content[keyIndex].Content[i+1].Encode(&externalRepo)
		return err != nil, err
	}
	return false, nil
}

func (p *Parser) CreateServiceNodes() ([]*yaml.Node, error) {
	repo := ExternalRepo{Version: p.version}
	node := yaml.Node{Kind: yaml.MappingNode}
	err := node.Encode(repo)
	if err != nil {
		return nil, err
	}
	nodes := []*yaml.Node{{Value: p.service, Kind: yaml.ScalarNode}, &node}
	return nodes, nil
}

type ExternalRepo struct {
	Version string `yaml:"serviceVersion"`
}

type Repo struct {
	Podspec struct {
		Services map[string]interface{} `yaml:"containers"`
	} `yaml:"podspec"`
}

type Service struct {
	Tag string
}

func (p Parser) GetContent() ([]byte, error) {
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(&p.loadedYaml)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (p *Parser) SaveFile(filename string) error {
	data, err := yaml.Marshal(&p.loadedYaml)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
