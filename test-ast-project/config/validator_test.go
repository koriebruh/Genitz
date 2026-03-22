package config

import (
	"testing"
)

func TestValidator(t *testing.T) {
	validate := NewValidator()

	type User struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"gte=0,lte=130"`
	}

	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "Valid User",
			user: User{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   30,
			},
			wantErr: false,
		},
		{
			name: "Invalid Email",
			user: User{
				Name:  "Jane Doe",
				Email: "invalid-email",
				Age:   25,
			},
			wantErr: true,
		},
		{
			name: "Missing Name",
			user: User{
				Name:  "",
				Email: "jane@example.com",
				Age:   25,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate.Struct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
