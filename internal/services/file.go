package service

import (
	"context"
	"io"
)

type File struct {
	repo Repositorier
}

func NewFile(repo Repositorier) *File {
	return &File{repo: repo}
}

func (f *File) Get(ctx context.Context, identifier string) (io.Reader, error) {
	return f.repo.Get(ctx, identifier)
}

func (f *File) Upload(ctx context.Context, fileName string, fileData io.Reader) (string, error) {
	return f.repo.Upload(ctx, fileName, fileData)
}

func (f *File) Delete(ctx context.Context, identifier string) error {
	return f.repo.Delete(ctx, identifier)
}
