package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ImgService interface {
	Create(galleryID uint, r io.Reader, filename string) error
	ByGalleryID(galleryID uint) ([]Img, error)
	Delete(i *Img) error
}

type imgService struct{}

type Img struct {
	GalleryID uint
	Filename  string
}

func NewImgService() ImgService {
	return &imgService{}
}

func (s *imgService) Create(galleryID uint, r io.Reader, filename string) error {
	path, err := s.mkImgDir(galleryID)
	if err != nil {
		return err
	}

	name, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return err
	}

	defer name.Close()

	_, err = io.Copy(name, r)
	if err != nil {
		return err
	}

	return nil
}

func (s *imgService) mkImgDir(galleryID uint) (string, error) {
	path := s.imgDir(galleryID)

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (i *imgService) imgDir(galleryID uint) string {
	return filepath.Join("images", "galleries", fmt.Sprintf("%v", galleryID))
}

func (i *imgService) ByGalleryID(galleryID uint) ([]Img, error) {
	path := i.imgDir(galleryID)

	s, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil, err
	}

	img := make([]Img, len(s))
	for i, str := range s {
		img[i] = Img{
			Filename:  filepath.Base(str),
			GalleryID: galleryID,
		}
	}

	return img, nil
}

func (i *Img) Path() string {
	return "/" + i.RelativePath()
}

func (i *Img) RelativePath() string {
	galleryID := fmt.Sprintf("%v", i.GalleryID)
	return filepath.Join("images", "galleries", galleryID, i.Filename)
}

func (is *imgService) Delete(i *Img) error {
	return os.Remove(i.RelativePath())
}
