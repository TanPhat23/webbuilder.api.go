package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	validSubdomainRe   = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{1,61}[a-z0-9])$`)
	validSupabaseHost  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}\.supabase\.(?:co|in)$`)
	reservedSubdomains = map[string]bool{
		"www": true, "api": true, "app": true, "admin": true, "ws": true,
		"mcp": true, "preview": true, "editor": true, "dashboard": true,
		"sign-in": true, "sign-up": true, "help": true, "helppage": true,
		"static": true, "assets": true, "cdn": true, "mail": true, "webhook": true,
	}
)

type ProjectAuthConfig struct {
	Provider        string `json:"provider"`
	SupabaseURL     string `json:"supabaseUrl"`
	SupabaseAnonKey string `json:"supabaseAnonKey"`
	LoginPath       string `json:"loginPath,omitempty"`
	AfterLoginPath  string `json:"afterLoginPath,omitempty"`
}

type Project struct {
	Pages         []Page         `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"pages,omitempty"`
	Snapshots     []Snapshot     `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"snapshots,omitempty"`
	Collaborators []Collaborator `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"collaborators,omitempty"`
	Invitations   []Invitation   `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"invitations,omitempty"`

	ID      string `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name    string `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	OwnerId string `gorm:"column:OwnerId;type:varchar(255);not null" json:"ownerId"`

	Owner           User             `gorm:"foreignKey:OwnerId" json:"owner,omitempty"`
	Logic           json.RawMessage  `gorm:"column:Logic;type:jsonb" json:"logic,omitempty"`
	Styles          *json.RawMessage `gorm:"column:Styles;type:jsonb" json:"styles,omitempty"`
	Header          *json.RawMessage `gorm:"column:Header;type:jsonb" json:"header,omitempty"`
	AuthConfig      *json.RawMessage `gorm:"column:AuthConfig;type:jsonb" json:"authConfig,omitempty"`
	Description     *string          `gorm:"column:Description;type:text" json:"description,omitempty"`
	Subdomain       *string          `gorm:"column:Subdomain;type:varchar(255)" json:"subdomain,omitempty"`

	Published bool `gorm:"column:Published;not null;default:false" json:"published"`

	*AuditFields

	// Owned by realtime: incremented on every committed logic update.
	LogicRevision int64 `gorm:"column:LogicRevision;not null;default:0;<-:false" json:"-"`
}

func (Project) TableName() string {
	return `public."Project"`
}

// NormalizeSubdomain validates and normalizes a subdomain label. Empty input
// returns ("", false, nil) so callers can clear an existing subdomain.
func NormalizeSubdomain(raw string) (string, bool, error) {
	cleaned := strings.ToLower(strings.TrimSpace(raw))
	if cleaned == "" {
		return "", false, nil
	}
	if !validSubdomainRe.MatchString(cleaned) || strings.Contains(cleaned, "--") {
		return "", false, errors.New("subdomain must be 3-63 characters using lowercase letters, numbers, and single hyphens")
	}
	if reservedSubdomains[cleaned] {
		return "", false, errors.New("this subdomain is reserved")
	}
	return cleaned, true, nil
}

func validateSafeRelativePath(raw, fallback string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback, nil
	}
	if len(trimmed) > 200 || !strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "//") || strings.ContainsAny(trimmed, "\\\r\n\t") {
		return "", errors.New("auth redirect paths must be relative paths starting with /")
	}
	return trimmed, nil
}

func isSupabaseServiceRoleJWT(key string) bool {
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false
	}
	role, _ := claims["role"].(string)
	return strings.EqualFold(role, "service_role")
}

// ValidateProjectAuthConfig validates and normalizes a project's Supabase Auth
// configuration. It enforces strict HTTPS Supabase hostnames (SSRF protection)
// and explicitly rejects service_role / secret keys.
func ValidateProjectAuthConfig(raw json.RawMessage) (*json.RawMessage, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return nil, nil
	}
	var cfg ProjectAuthConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, errors.New("invalid authConfig JSON")
	}
	cfg.Provider = strings.ToLower(strings.TrimSpace(cfg.Provider))
	if cfg.Provider != "supabase" {
		return nil, errors.New("authConfig.provider must be \"supabase\"")
	}
	parsedURL, err := url.Parse(strings.TrimSpace(cfg.SupabaseURL))
	if err != nil || parsedURL.Scheme != "https" || parsedURL.User != nil || parsedURL.Port() != "" || parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return nil, errors.New("supabaseUrl must be a valid https://<project>.supabase.co URL without paths, ports, or credentials")
	}
	if parsedURL.Path != "" && parsedURL.Path != "/" {
		return nil, errors.New("supabaseUrl must not include a path")
	}
	host := strings.ToLower(parsedURL.Hostname())
	if !validSupabaseHost.MatchString(host) {
		return nil, errors.New("supabaseUrl must point to a *.supabase.co or *.supabase.in project")
	}
	cfg.SupabaseURL = "https://" + host

	cfg.SupabaseAnonKey = strings.TrimSpace(cfg.SupabaseAnonKey)
	if cfg.SupabaseAnonKey == "" || len(cfg.SupabaseAnonKey) > 1024 || strings.ContainsAny(cfg.SupabaseAnonKey, " \t\r\n") {
		return nil, errors.New("supabaseAnonKey is required and must not contain whitespace")
	}
	if strings.HasPrefix(cfg.SupabaseAnonKey, "sb_secret_") || isSupabaseServiceRoleJWT(cfg.SupabaseAnonKey) {
		return nil, errors.New("do not use a Supabase secret or service_role key; use the public anon or publishable key")
	}

	loginPath, err := validateSafeRelativePath(cfg.LoginPath, "/login")
	if err != nil {
		return nil, err
	}
	afterLoginPath, err := validateSafeRelativePath(cfg.AfterLoginPath, "/")
	if err != nil {
		return nil, err
	}
	cfg.LoginPath = loginPath
	cfg.AfterLoginPath = afterLoginPath

	encoded, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	msg := json.RawMessage(encoded)
	return &msg, nil
}
