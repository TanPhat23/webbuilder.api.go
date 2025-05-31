package database

import "my-go-app/internal/repositories"

type Repositories struct {
	*repositories.ElementRepository
}

func  (d *Repositories) DatabaseConn() (*Repositories, error) {
	return nil, nil
}