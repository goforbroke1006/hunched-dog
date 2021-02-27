package internal

import (
	"os"
	"path/filepath"
)

type MetaFile struct {
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	UpdatedAt int64  `json:"updated_at"`
}

type Registry []MetaFile

func GetLocal(directory string) (Registry, error) {
	reg := Registry{}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		reg = append(reg, MetaFile{
			Filename:  path,
			Size:      info.Size(),
			UpdatedAt: info.ModTime().Unix(),
		})
		return nil
	})
	return reg, err
}

func Diff(local, remote Registry) []string {
	result := []string{}

	for _, rf := range remote {
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
