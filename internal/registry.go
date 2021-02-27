package internal

import (
	"os"
	"path/filepath"
	"strings"
)

type MetaFile struct {
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
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

		relFilename := strings.TrimPrefix(path, directory+"/")

		reg = append(reg, MetaFile{
			Filename:  relFilename,
			Size:      info.Size(),
			UpdatedAt: info.ModTime().Unix(),
			IsDir:     info.IsDir(),
		})
		return nil
	})
	return reg, err
}

func DiffDirs(local, remote Registry) []string {
	result := []string{}

	for _, rf := range remote {
		if !rf.IsDir {
			continue
		}

		notFound := true
		for _, lf := range local {
			if rf.Filename != lf.Filename {
				continue
			}
			notFound = false

			if rf.UpdatedAt > lf.UpdatedAt {
				result = append(result, rf.Filename)
			}
		}
		if notFound {
			result = append(result, rf.Filename)
		}
	}

	return result
}

func DiffFiles(local, remote Registry) []string {
	result := []string{}

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
				result = append(result, rf.Filename)
			}
		}
		if notFound {
			result = append(result, rf.Filename)
		}
	}

	return result
}
