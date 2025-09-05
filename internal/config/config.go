package config

import (
	"fmt"
	"os"
)

type GCSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

type Config struct {
	SmallImgDir string
	MedImgDir   string
	OrgImgDir   string
	FileDir     string // Directory where original files are stored before processing
}

func NewConfig() *Config {

	smallDir, err := getEnv("SMALL_IMG_DIR")
	if err != nil {
		panic(err)
	}
	medDir, err := getEnv("MED_IMG_DIR")
	if err != nil {
		panic(err)
	}
	orgDir, err := getEnv("ORG_IMG_DIR")
	if err != nil {
		panic(err)
	}
	fileDir, err := getEnv("FILE_DIR")
	if err != nil {
		panic(err)
	}

	return &Config{
		SmallImgDir: smallDir,
		MedImgDir:   medDir,
		OrgImgDir:   orgDir,
		FileDir:     fileDir,
	}
}

type ImageTarget struct {
	Path    string
	Width   int
	Quality int
	IsOrg   bool
}

func getEnv(envVar string) (string, error) {
	value := os.Getenv(envVar)
	if value == "" {
		return "", fmt.Errorf("environment variable %s is not set", envVar)
	}
	return value, nil
}
