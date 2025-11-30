package ldap

import "testing"

func TestValidateLDAPUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "Valid username",
			username: "john.doe",
			wantErr:  false,
		},
		{
			name:     "Valid username with numbers",
			username: "user123",
			wantErr:  false,
		},
		{
			name:     "LDAP injection attempt - wildcard",
			username: "admin*",
			wantErr:  true,
		},
		{
			name:     "LDAP injection attempt - filter",
			username: "admin)(cn=*",
			wantErr:  true,
		},
		{
			name:     "Empty username",
			username: "",
			wantErr:  true,
		},
		{
			name:     "Too long username",
			username: string(make([]byte, 300)),
			wantErr:  true,
		},
		{
			name:     "Null byte injection",
			username: "admin\x00",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLDAPUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLDAPUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidLDAPDN(t *testing.T) {
	tests := []struct {
		name string
		dn   string
		want bool
	}{
		{
			name: "Valid DN",
			dn:   "cn=admin,dc=example,dc=com",
			want: true,
		},
		{
			name: "Valid DN with OU",
			dn:   "cn=user,ou=users,dc=example,dc=com",
			want: true,
		},
		{
			name: "Invalid DN - no comma",
			dn:   "cn=admin",
			want: false,
		},
		{
			name: "Invalid DN - no equals",
			dn:   "adminexample",
			want: false,
		},
		{
			name: "Empty DN",
			dn:   "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidLDAPDN(tt.dn); got != tt.want {
				t.Errorf("isValidLDAPDN() = %v, want %v", got, tt.want)
			}
		})
	}
}
