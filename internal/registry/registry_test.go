package registry

import (
	"path/filepath"
	"testing"
)

func TestGetLocal(t *testing.T) {
	type args struct {
		directory string
	}
	tests := []struct {
		name    string
		args    args
		want    Registry
		wantErr bool
	}{
		{
			name: "project structure",
			args: args{directory: "./testdata"},
			want: Registry{
				{Filename: "registry", Size: 0, Hash: "", UpdatedAt: 1614938297, IsDir: true},
				{Filename: "registry/test.txt", Size: 6, Hash: "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92", UpdatedAt: 1614938297, IsDir: false},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			absPath, _ := filepath.Abs(tt.args.directory)

			got, err := GetLocal(absPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for index := range got {
				// exclude UpdatedAt because testdata can be related to machine
				if got[index].Filename != tt.want[index].Filename ||
					got[index].Size != tt.want[index].Size ||
					got[index].Hash != tt.want[index].Hash ||
					got[index].IsDir != tt.want[index].IsDir {
					t.Errorf("GetLocal() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
