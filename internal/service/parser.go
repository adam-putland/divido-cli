package service

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/adam-putland/divido-cli/internal/models"
	"github.com/mitchellh/mapstructure"
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

	startingNode := p.loadedYaml
	if len(p.loadedYaml.Content) > 0 && p.loadedYaml.Content[0].Kind == yaml.ScalarNode && p.loadedYaml.Content[0].Value == "services" {
		startingNode = *p.loadedYaml.Content[1]
	}

	services := make(map[string]*models.Service)

	if startingNode.Kind == yaml.MappingNode {
		for i := 0; i < len(startingNode.Content); i++ {
			key := strings.ReplaceAll(startingNode.Content[i].Value, " ", "")
			if key != "" {
				content := startingNode.Content[i+1]
				repo, err := p.getRepo(content)
				if err != nil {
					return nil, err
				}
				services[key] = &models.Service{HLMName: key, Release: models.Release{Name: key, Version: repo.GetVersion()}}
			}
		}
	}
	return services, nil
}

func (p *Parser) Replace(services models.Services) error {

	startingNode := p.loadedYaml
	if len(p.loadedYaml.Content) > 0 && p.loadedYaml.Content[0].Kind == yaml.ScalarNode && p.loadedYaml.Content[0].Value == "services" {
		startingNode = *p.loadedYaml.Content[1]
	}

	if startingNode.Kind == yaml.MappingNode {
		for i := 0; i < len(startingNode.Content); i++ {
			key := strings.ReplaceAll(startingNode.Content[i].Value, " ", "")
			if service, ok := services[key]; ok {

				content := startingNode.Content[i+1]
				repo, err := p.getRepo(content)
				if err != nil {
					return err
				}

				if repo.GetVersion() != service.Version {
					repo.UpdateVersion(service.Version)
					if err = content.Encode(&repo); err != nil {
						return err
					}
					fmt.Printf("updated service %s from %s to %s\n", service.Name, repo.GetVersion(), service.Version)
				}
				delete(services, key)

			}
		}
	}

	// if the service is not in the document it will be created
	if len(services) > 0 {
		for _, service := range services {
			nodes, err := p.createServiceNodes(service)
			if err != nil {
				return err
			}
			startingNode.Content = append(startingNode.Content, nodes...)
		}

	}

	if len(p.loadedYaml.Content) > 0 && p.loadedYaml.Content[0].Kind == yaml.ScalarNode && p.loadedYaml.Content[0].Value == "services" {
		p.loadedYaml.Content[1] = &startingNode
	} else {
		p.loadedYaml = startingNode
	}

	return nil
}

func (p *Parser) getRepo(content *yaml.Node) (Repo, error) {
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

	return &internalRepo, err
}

func (p *Parser) createServiceNodes(s *models.Service) ([]*yaml.Node, error) {
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

type InternalService struct {
	Tag     string
	Version string
}
type InternalRepo struct {
	Podspec struct {
		Services map[string]interface{} `yaml:"containers"`
	} `yaml:"podspec"`
}

func (i *InternalRepo) UpdateVersion(version string) {
	for index, service := range i.Podspec.Services {

		var parsedService InternalService
		err := mapstructure.Decode(service, &parsedService)
		if err == nil {
			if parsedService.Tag != "" {
				parsedService.Tag = version
				i.Podspec.Services[index] = parsedService
			}
			if parsedService.Tag != "" {
				parsedService.Tag = version
				i.Podspec.Services[index] = parsedService
			}
			i.Podspec.Services[index] = parsedService
		}

	}
}

func (i *InternalRepo) GetVersion() string {
	for _, service := range i.Podspec.Services {

		var parsedService InternalService
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
