package internal

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type MetaFile struct {
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	Hash      string `json:"hash"`
	UpdatedAt int64  `json:"updated_at"`
	IsDir     bool   `json:"is_dir"`
}

type Registry []MetaFile

func GetLocal(directory string) (Registry, error) {
	reg := Registry{}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == directory {
			return nil
		}

		relFilename := strings.TrimPrefix(path, directory+"/")

		hash := ""
		if !info.IsDir() {
			hash = GetFileHash(path)
		}

		reg = append(reg, MetaFile{
			Filename:  relFilename,
			Size:      info.Size(),
			Hash:      hash,
			UpdatedAt: info.ModTime().Unix(),
			IsDir:     info.IsDir(),
		})
		return nil
	})
	return reg, err
}

func DiffDirs(remote Registry) []string {
	result := []string{}

	for _, rf := range remote {
		if !rf.IsDir {
			continue
		}

		result = append(result, rf.Filename)
	}

	return result
}

func DiffFiles(local, remote Registry) Registry {
	result := Registry{}

	for _, rf := range remote {
		if rf.IsDir {
			continue
		}
		notFound := true
		for _, lf := range local {
			if rf.Filename != lf.Filename {
				continue
			}
			notFound = false

			if rf.UpdatedAt > lf.UpdatedAt {
				if rf.Hash != lf.Hash {
					result = append(result, rf)
					continue
				}

				if rf.Size != lf.Size {
					result = append(result, rf)
					continue
				}
			}
		}
		if notFound {
			result = append(result, rf)
		}
	}

	return result
}

func GetFileHash(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
