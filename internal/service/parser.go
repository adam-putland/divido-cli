package service

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/mitchellh/mapstructure"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Parser struct {
	loadedYaml yaml.Node
	rawData    []byte
}

func NewParser(rawData []byte) *Parser {
	return &Parser{
		rawData: rawData,
	}
}

func (p *Parser) Load() (models.Services, error) {
	err := yaml.Unmarshal(p.rawData, &p.loadedYaml)
	if err != nil {
		return nil, err
	}

	if len(p.loadedYaml.Content) == 0 {
		return nil, errors.New("invalid file")
	}

	// check if loading from document file
	if p.loadedYaml.Kind == yaml.DocumentNode {
		p.loadedYaml = *p.loadedYaml.Content[0]
	}

	services := make(map[string]*models.Service)

	for keyIndex := 0; keyIndex < len(p.loadedYaml.Content); keyIndex++ {
		if p.loadedYaml.Content[keyIndex].Kind == yaml.MappingNode {
			for i := 0; i < len(p.loadedYaml.Content[keyIndex].Content); i++ {
				key := strings.ReplaceAll(p.loadedYaml.Content[keyIndex].Content[i].Value, " ", "")
				if key != "" {
					content := p.loadedYaml.Content[keyIndex].Content[i+1]
					repo, err := p.GetRepo(content)
					if err != nil {
						return nil, err
					}
					services[key] = &models.Service{HLMName: key, Release: models.Release{Version: repo.GetVersion()}}
				}
			}
		}
	}

	return services, nil
}

func (p *Parser) Replace(services models.Services) error {

	for keyIndex := 0; keyIndex < len(p.loadedYaml.Content); keyIndex++ {
		if p.loadedYaml.Content[keyIndex].Kind == yaml.MappingNode {
			for i := 0; i < len(p.loadedYaml.Content[keyIndex].Content); i++ {
				key := strings.ReplaceAll(p.loadedYaml.Content[keyIndex].Content[i].Value, " ", "")
				if service, ok := services[key]; ok {

					content := p.loadedYaml.Content[keyIndex].Content[i+1]
					repo, err := p.GetRepo(content)
					if err != nil {
						return err
					}

					repo.UpdateVersion(service.Version)
					if err = content.Encode(&repo); err != nil {
						return err
					}
					fmt.Printf("updated service %s from %s to %s\n", service.Name, repo.GetVersion(), service.Version)

					delete(services, key)
					break
				}
			}

			// if the service is not in the document it will be created
			if len(services) > 0 {
				for _, service := range services {
					nodes, err := p.CreateServiceNodes(service)
					if err != nil {
						return err
					}
					p.loadedYaml.Content[keyIndex].Content = append(p.loadedYaml.Content[keyIndex].Content, nodes...)
				}

			}
		}
	}

	return nil
}

func (p *Parser) GetRepo(content *yaml.Node) (Repo, error) {
	var externalRepo ExternalRepo
	err := content.Decode(&externalRepo)

	if err == nil && externalRepo.GetVersion() != "" {
		return &externalRepo, nil
	}

	var internalRepo InternalRepo
	err = content.Decode(&internalRepo)
	if err != nil {
		return nil, err
	}
	if internalRepo.GetVersion() == "" {
		return nil, errors.New("no version tag found")
	}

	return &internalRepo, err
}

func (p *Parser) CreateServiceNodes(s *models.Service) ([]*yaml.Node, error) {
	repo := ExternalRepo{Version: s.Version}
	node := yaml.Node{Kind: yaml.MappingNode}
	err := node.Encode(repo)
	if err != nil {
		return nil, err
	}
	nodes := []*yaml.Node{{Value: s.HLMName, Kind: yaml.ScalarNode}, &node}
	return nodes, nil
}

type Repo interface {
	UpdateVersion(string)
	GetVersion() string
}

type InternaLService struct {
	Tag string
}

type InternalRepo struct {
	Podspec struct {
		Services map[string]interface{} `yaml:"containers"`
	} `yaml:"podspec"`
}

func (i *InternalRepo) UpdateVersion(version string) {
	for index, service := range i.Podspec.Services {

		var parsedService InternaLService
		err := mapstructure.Decode(service, &parsedService)
		if err == nil && parsedService.Tag != "" {
			parsedService.Tag = version
			i.Podspec.Services[index] = parsedService
		}
	}
}

func (i *InternalRepo) GetVersion() string {
	for _, service := range i.Podspec.Services {

		var parsedService InternaLService
		err := mapstructure.Decode(service, &parsedService)
		if err == nil && parsedService.Tag != "" {
			return parsedService.Tag
		}
	}
	return ""
}

type ExternalRepo struct {
	Version string `yaml:"serviceVersion"`
}

func (e *ExternalRepo) UpdateVersion(version string) {
	e.Version = version
}

func (e *ExternalRepo) GetVersion() string {
	return e.Version
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
