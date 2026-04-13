package service

import "errors"

var (
	ErrInvalidProduct       = errors.New("eksik urun bilgisi")
	ErrInvalidProductID     = errors.New("gecersiz urun id")
	ErrInvalidOrder         = errors.New("siparis verisi gecersiz")
	ErrUnsupportedImageType = errors.New("desteklenmeyen dosya tipi")
	ErrInvalidFileExtension = errors.New("gecersiz dosya uzantisi")
)
