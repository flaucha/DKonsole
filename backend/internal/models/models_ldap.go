package models

// LDAPConfig representa la configuración del servidor LDAP
type LDAPConfig struct {
	Enabled            bool     `json:"enabled"`
	URL                string   `json:"url"`
	BaseDN             string   `json:"baseDN"`
	UserDN             string   `json:"userDN"`
	GroupDN            string   `json:"groupDN"`
	UserFilter         string   `json:"userFilter,omitempty"`
	RequiredGroup      string   `json:"requiredGroup,omitempty"`      // Grupo requerido para acceso (opcional)
	AdminGroups        []string `json:"adminGroups,omitempty"`        // Grupos LDAP que tienen acceso de admin al cluster
	InsecureSkipVerify bool     `json:"insecureSkipVerify,omitempty"` // Skip TLS certificate verification (warning: insecure)
	CACert             string   `json:"caCert,omitempty"`             // CA certificate in PEM format for TLS verification
}

// LDAPGroupPermission representa los permisos de un grupo LDAP para un namespace
type LDAPGroupPermission struct {
	Namespace  string `json:"namespace"`
	Permission string `json:"permission"` // "view", "edit"
}

// LDAPGroup representa un grupo LDAP con sus permisos
type LDAPGroup struct {
	Name        string                `json:"name"`
	Permissions []LDAPGroupPermission `json:"permissions"`
}

// LDAPGroupsConfig representa la configuración de grupos LDAP
type LDAPGroupsConfig struct {
	Groups []LDAPGroup `json:"groups"`
}
