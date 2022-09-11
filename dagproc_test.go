package dagproc

import (
	"errors"
	"testing"
)

func Test_isErrIgnore(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"ignore",
			args{NewErrIgnore("ignore this")},
			true,
		},
		{
			"not an ignore error",
			args{errors.New("random error that you don't know about")},
			false,
		},
		{
			"nil",
			args{nil},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIgnoreError(tt.args.err); got != tt.want {
				t.Errorf("isErrIgnore() = %v, want %v", got, tt.want)
			}
		})
	}
}
