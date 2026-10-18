package storage

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"example.com/inga/internal/config"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

type Storage interface {
	UploadImage(ctx *gin.Context, file *multipart.FileHeader) (string, error)
}
type Service struct {
	cfg config.CloudinaryConfig
	cld *cloudinary.Cloudinary
}

func New(cfg config.CloudinaryConfig) (*Service, error) {
	cld, err := buildCloudinaryConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Service{cfg: cfg, cld: cld}, nil
}

func buildCloudinaryConfig(cf config.CloudinaryConfig) (*cloudinary.Cloudinary, error) {
	cld, _ := cloudinary.NewFromParams(cf.CloudName, cf.APIKey, cf.APISecret)
	cld.Config.URL.Secure = true
	return cld, nil
}

func (s *Service) UploadImage(ctx *gin.Context, file *multipart.FileHeader) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("nil request context")
	}
	if file == nil {
		return "", fmt.Errorf("nil file header")
	}

	opened, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open upload file: %w", err)
	}
	defer func(opened multipart.File) {
		err := opened.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(opened)

	baseName := strings.TrimSuffix(filepath.Base(file.Filename), filepath.Ext(file.Filename))
	if baseName == "" || baseName == "." {
		baseName = "upload"
	}

	resp, err := s.cld.Upload.Upload(ctx.Request.Context(), opened, uploader.UploadParams{
		Folder:         "assets/inga",
		PublicID:       baseName,
		UniqueFilename: api.Bool(false),
		Overwrite:      api.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("cloudinary upload: %w", err)
	}

	return resp.SecureURL, nil
}
