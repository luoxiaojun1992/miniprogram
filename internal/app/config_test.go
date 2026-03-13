package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitConfig_EnvVarsOnly verifies that nested config keys (e.g.
// database.host) are correctly read from APP_* environment variables even
// when no config file is provided.  This guards against the Viper
// AutomaticEnv bug where dot-separated keys produce invalid env var names
// (e.g. APP_DATABASE.HOST instead of APP_DATABASE_HOST).
func TestInitConfig_EnvVarsOnly(t *testing.T) {
	t.Setenv("APP_SERVER_PORT", "9090")
	t.Setenv("APP_SERVER_MODE", "release")
	t.Setenv("APP_DATABASE_HOST", "db-host")
	t.Setenv("APP_DATABASE_PORT", "5432")
	t.Setenv("APP_DATABASE_USER", "testuser")
	t.Setenv("APP_DATABASE_PASSWORD", "testpass")
	t.Setenv("APP_DATABASE_NAME", "testdb")
	t.Setenv("APP_JWT_SECRET", "supersecret")
	t.Setenv("APP_JWT_EXPIRY", "3600")
	t.Setenv("APP_LOG_LEVEL", "warn")
	t.Setenv("APP_WECHAT_APP_ID", "wx123")
	t.Setenv("APP_WECHAT_APP_SECRET", "wxsecret")
	t.Setenv("APP_UPLOAD_DIR", "/tmp/uploads")
	t.Setenv("APP_UPLOAD_BASE_URL", "http://example.com/static")
	t.Setenv("APP_DEBUG_ENABLE_TEST_TOKEN", "true")

	cfg, err := InitConfig("")
	require.NoError(t, err)

	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "release", cfg.Server.Mode)
	assert.Equal(t, "db-host", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "supersecret", cfg.JWT.Secret)
	assert.Equal(t, 3600, cfg.JWT.Expiry)
	assert.Equal(t, "warn", cfg.Log.Level)
	assert.Equal(t, "wx123", cfg.Wechat.AppID)
	assert.Equal(t, "wxsecret", cfg.Wechat.AppSecret)
	assert.Equal(t, "/tmp/uploads", cfg.Upload.Dir)
	assert.Equal(t, "http://example.com/static", cfg.Upload.BaseURL)
	assert.True(t, cfg.Debug.EnableTestToken)
}

// TestInitConfig_Defaults verifies that sensible defaults are applied when
// no config file or environment variables are provided.
func TestInitConfig_Defaults(t *testing.T) {
	// Unset any vars that might be set in the test environment.
	vars := []string{
		"APP_SERVER_PORT", "APP_SERVER_MODE",
		"APP_DATABASE_HOST", "APP_DATABASE_PORT",
		"APP_DATABASE_USER", "APP_DATABASE_PASSWORD", "APP_DATABASE_NAME",
		"APP_JWT_SECRET", "APP_JWT_EXPIRY",
		"APP_LOG_LEVEL",
		"APP_WECHAT_APP_ID", "APP_WECHAT_APP_SECRET",
		"APP_UPLOAD_DIR", "APP_UPLOAD_BASE_URL",
		"APP_DEBUG_ENABLE_TEST_TOKEN",
	}
	for _, v := range vars {
		t.Setenv(v, "")
	}

	cfg, err := InitConfig("")
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "release", cfg.Server.Mode)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 3306, cfg.Database.Port)
	assert.Equal(t, "root", cfg.Database.User)
	assert.Equal(t, "miniapp", cfg.Database.Name)
	assert.Equal(t, 7200, cfg.JWT.Expiry)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "storage/uploads", cfg.Upload.Dir)
	assert.False(t, cfg.Debug.EnableTestToken)
}
