package models

type Services map[string]*Service

func (services Services) ToArray() []*Service {
	arr := make([]*Service, 0, len(services))
	for _, ser := range services {
		arr = append(arr, ser)
	}
	return arr
}

type Service struct {
	Release
	HLMName string
}
