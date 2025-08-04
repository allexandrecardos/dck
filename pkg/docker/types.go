package docker

import "time"

type ContainerOption struct {
	ID    string // ID completo
	Label string // Exibição resumida para o usuário (imagem - id curto - status)
}

type Container struct {
	ID      string
	Name    string
	Image   string
	Status  string
	Created time.Time
	// Size    int64
	// Network string
	Port []string
	// Volumes     []string
	State string // running, stopped, paused, etc.
}

type Port struct {
	Private int
	Public  int
	Type    string
}

type Image struct {
	ID      string
	Tag     string
	Size    int64
	Created time.Time
}
