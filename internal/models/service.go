package models

type Services map[string]*Service

type Service struct {
	Release
	HLMName string
}
