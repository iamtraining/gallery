package models

import "github.com/jinzhu/gorm"

var _ GalleryDB = &galleryGorm{}

const (
	ErrTitleReq  modelError = "models: title is required"
	ErrUserIDReq modelError = "models: user ID is required"
)

type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not_null;index"`
	Title  string `gorm:"not_null"`
	Img    []Img  `gorm:"-"`
}

type GalleryService interface {
	GalleryDB
}

type GalleryDB interface {
	ByID(id uint) (*Gallery, error)
	ByUserID(userID uint) ([]Gallery, error)
	Create(gallery *Gallery) error
	Update(gallery *Gallery) error
	Delete(id uint) error
}

type galleryGorm struct {
	db *gorm.DB
}

type galleryService struct {
	GalleryDB
}

type galleryValidator struct {
	GalleryDB
}

type galValFunc func(*Gallery) error

func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{
			GalleryDB: &galleryGorm{
				db: db,
			},
		},
	}
}

func (g *galleryGorm) Create(gallery *Gallery) error {
	return g.db.Create(gallery).Error
}

func (g *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gal Gallery
	db := g.db.Where("id = ?", id)
	err := first(db, &gal)
	if err != nil {
		return nil, err
	}

	return &gal, nil
}

func runGalValFuncs(g *Gallery, funcs ...galValFunc) error {
	for _, fn := range funcs {
		if err := fn(g); err != nil {
			return err
		}
	}

	return nil
}

func (g *galleryValidator) titleCheck(gal *Gallery) error {
	if gal.Title == "" {
		return ErrTitleReq
	}

	return nil
}

func (g *galleryValidator) userIDCheck(gal *Gallery) error {
	if gal.UserID <= 0 {
		return ErrUserIDReq
	}

	return nil
}

func (g *galleryValidator) Create(gal *Gallery) error {
	err := runGalValFuncs(gal,
		g.userIDCheck,
		g.titleCheck,
	)
	if err != nil {
		return err
	}

	return g.GalleryDB.Create(gal)
}

func (g *galleryGorm) Update(gallery *Gallery) error {
	return g.db.Save(gallery).Error
}

func (g *galleryValidator) Update(gallery *Gallery) error {
	err := runGalValFuncs(gallery,
		g.userIDCheck,
		g.titleCheck,
	)
	if err != nil {
		return err
	}

	return g.GalleryDB.Update(gallery)
}

func (g *galleryGorm) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID: id}}

	return g.db.Delete(&gallery).Error
}

func (g *galleryValidator) notEqToZero(gal *Gallery) error {
	if gal.ID <= 0 {
		return ErrIDInvalid
	}

	return nil
}

func (g *galleryValidator) Delete(id uint) error {
	gallery := Gallery{}
	gallery.ID = id

	if err := runGalValFuncs(&gallery,
		g.notEqToZero,
	); err != nil {
		return err
	}

	return g.GalleryDB.Delete(gallery.ID)
}

func (g *galleryGorm) ByUserID(userID uint) ([]Gallery, error) {
	var galleries []Gallery

	db := g.db.Where("user_id = ?", userID)
	if err := db.Find(&galleries).Error; err != nil {
		return nil, err
	}

	return galleries, nil
}

func (g *Gallery) Split(n int) [][]Img {
	dd := make([][]Img, n)

	for i := 0; i < n; i++ {
		dd[i] = make([]Img, 0)
	}

	for i, img := range g.Img {
		buck := i % n
		dd[buck] = append(dd[buck], img)
	}

	return dd
}
