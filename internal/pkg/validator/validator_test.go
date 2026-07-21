package validator

import (
	"errors"
	"testing"

	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
)

type sampleInput struct {
	Username string `validate:"required,min=3,max=64"`
	Password string `validate:"required,min=6,max=72"`
}

func TestStruct(t *testing.T) {
	tests := []struct {
		name    string
		in      sampleInput
		wantErr bool
	}{
		{
			name: "valid",
			in:   sampleInput{Username: "alice", Password: "secret123"},
		},
		{
			name:    "missing username",
			in:      sampleInput{Password: "secret123"},
			wantErr: true,
		},
		{
			name:    "username too short",
			in:      sampleInput{Username: "ab", Password: "secret123"},
			wantErr: true,
		},
		{
			name:    "password too long",
			in:      sampleInput{Username: "alice", Password: string(make([]byte, 73))},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Struct(&tt.in)
			if tt.wantErr {
				if !errors.Is(err, errcode.ErrInvalidArgument) {
					t.Fatalf("error = %v, want ErrInvalidArgument", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
