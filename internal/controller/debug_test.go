package controller

import (
"bytes"
"context"
"encoding/json"
"fmt"
"net/http"
"net/http/httptest"
"testing"

"github.com/gin-gonic/gin"
"github.com/golang-jwt/jwt/v5"
"github.com/sirupsen/logrus"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"

apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

// apiResponse is used to unmarshal the standard JSON envelope.
type apiResponse struct {
Code    int                    `json:"code"`
Message string                 `json:"message"`
Data    map[string]interface{} `json:"data"`
}

func newDebugCtrl(repo *testutil.MockUserRepository) *DebugController {
return NewDebugController(repo, "test-secret", 3600, logrus.New())
}

func newDebugRouter(ctrl *DebugController) *gin.Engine {
r := newTestRouter()
r.POST("/v1/debug/token", ctrl.GenerateTestToken)
return r
}

func postDebug(r *gin.Engine, body interface{}) *httptest.ResponseRecorder {
raw, _ := json.Marshal(body)
req, _ := http.NewRequest("POST", "/v1/debug/token", bytes.NewReader(raw))
req.Header.Set("Content-Type", "application/json")
w := httptest.NewRecorder()
r.ServeHTTP(w, req)
return w
}

// ── Happy path: user looked up from DB ────────────────────────────────────────

func TestDebugCtrl_GenerateTestToken_LooksUpUserType(t *testing.T) {
repo := &testutil.MockUserRepository{
GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
return &entity.User{ID: id, UserType: 2}, nil // admin
},
}
ctrl := newDebugCtrl(repo)
r := newDebugRouter(ctrl)
w := postDebug(r, map[string]interface{}{"user_id": 1})

assert.Equal(t, 200, w.Code)
var resp apiResponse
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
assert.Equal(t, 0, resp.Code)
assert.NotEmpty(t, resp.Data["access_token"])
assert.EqualValues(t, 2, resp.Data["user_type"])
assert.EqualValues(t, 1, resp.Data["user_id"])
assert.Equal(t, "Bearer", resp.Data["token_type"])
}

func TestDebugCtrl_GenerateTestToken_UserTypeOverride(t *testing.T) {
// user_type_override != 0 → skip DB lookup, use override directly
ctrl := NewDebugController(nil, "test-secret", 3600, logrus.New())
r := newDebugRouter(ctrl)
w := postDebug(r, map[string]interface{}{
"user_id":            5,
"user_type_override": 1,
})
assert.Equal(t, 200, w.Code)
var resp apiResponse
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
assert.EqualValues(t, 1, resp.Data["user_type"])
}

func TestDebugCtrl_GenerateTestToken_UserTypeOverride_BackendTypes(t *testing.T) {
ctrl := NewDebugController(nil, "test-secret", 3600, logrus.New())
r := newDebugRouter(ctrl)
for _, userType := range []int{2, 3} {
w := postDebug(r, map[string]interface{}{
"user_id":            5,
"user_type_override": userType,
})
assert.Equal(t, 200, w.Code)
var resp apiResponse
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
assert.EqualValues(t, userType, resp.Data["user_type"])
}
}

func TestDebugCtrl_GenerateTestToken_CustomExpiry(t *testing.T) {
ctrl := NewDebugController(nil, "test-secret", 3600, logrus.New())
r := newDebugRouter(ctrl)
w := postDebug(r, map[string]interface{}{
"user_id":            42,
"user_type_override": 1,
"expiry_seconds":     600,
})
assert.Equal(t, 200, w.Code)
var resp apiResponse
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
assert.EqualValues(t, 600, resp.Data["expires_in"])
}

func TestDebugCtrl_GenerateTestToken_DefaultExpiry_WhenZero(t *testing.T) {
ctrl := NewDebugController(nil, "test-secret", 1800, logrus.New())
r := newDebugRouter(ctrl)
w := postDebug(r, map[string]interface{}{
"user_id":            7,
"user_type_override": 1,
"expiry_seconds":     0,
})
assert.Equal(t, 200, w.Code)
var resp apiResponse
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
assert.EqualValues(t, 1800, resp.Data["expires_in"])
}

// ── Error paths ───────────────────────────────────────────────────────────────

func TestDebugCtrl_GenerateTestToken_MissingUserID(t *testing.T) {
ctrl := newDebugCtrl(&testutil.MockUserRepository{})
r := newDebugRouter(ctrl)
w := postDebug(r, map[string]interface{}{"user_type_override": 1})
assert.Equal(t, 400, w.Code)
}

func TestDebugCtrl_GenerateTestToken_InvalidJSON(t *testing.T) {
ctrl := newDebugCtrl(&testutil.MockUserRepository{})
r := newDebugRouter(ctrl)
req, _ := http.NewRequest("POST", "/v1/debug/token", bytes.NewBufferString("{bad json"))
req.Header.Set("Content-Type", "application/json")
w := httptest.NewRecorder()
r.ServeHTTP(w, req)
assert.Equal(t, 400, w.Code)
}

func TestDebugCtrl_GenerateTestToken_UserNotFound(t *testing.T) {
repo := &testutil.MockUserRepository{
GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
return nil, nil
},
}
ctrl := newDebugCtrl(repo)
r := newDebugRouter(ctrl)
w := postDebug(r, map[string]interface{}{"user_id": 999})
assert.Equal(t, 404, w.Code)
}

func TestDebugCtrl_GenerateTestToken_UserRepoError(t *testing.T) {
repo := &testutil.MockUserRepository{
GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
return nil, apperrors.NewInternal("db error", nil)
},
}
ctrl := newDebugCtrl(repo)
r := newDebugRouter(ctrl)
w := postDebug(r, map[string]interface{}{"user_id": 1})
assert.Equal(t, 500, w.Code)
}

func TestDebugCtrl_GenerateTestToken_InvalidUserTypeOverride(t *testing.T) {
ctrl := NewDebugController(nil, "test-secret", 3600, logrus.New())
r := newDebugRouter(ctrl)
w := postDebug(r, map[string]interface{}{
"user_id":            5,
"user_type_override": 4,
})
assert.Equal(t, 400, w.Code)
}

// ── Token sign failure ────────────────────────────────────────────────────────

type debugFailingSign struct{}

func (debugFailingSign) Alg() string                                    { return "FAIL" }
func (debugFailingSign) Sign(_ string, _ interface{}) ([]byte, error)   { return nil, fmt.Errorf("forced") }
func (debugFailingSign) Verify(_ string, _ []byte, _ interface{}) error { return nil }

func TestDebugCtrl_GenerateTestToken_SignFail(t *testing.T) {
ctrl := &DebugController{
userRepo:      nil,
jwtSecret:     "secret",
jwtExpiry:     3600,
log:           logrus.New(),
signingMethod: debugFailingSign{},
}
r := newDebugRouter(ctrl)
w := postDebug(r, map[string]interface{}{
"user_id":            1,
"user_type_override": 1,
})
assert.Equal(t, 500, w.Code)
}

// ── Route disabled by default ─────────────────────────────────────────────────

func TestDebugRoute_DisabledByDefault(t *testing.T) {
r := newTestRouter() // no debug route registered
req, _ := http.NewRequest("POST", "/v1/debug/token", nil)
w := httptest.NewRecorder()
r.ServeHTTP(w, req)
assert.Equal(t, 404, w.Code)
}

// ── Router integration test with jwt.SigningMethod ───────────────────────────

func TestDebugCtrl_GenerateTestToken_ValidJWT(t *testing.T) {
repo := &testutil.MockUserRepository{
GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
return &entity.User{ID: id, UserType: 1}, nil
},
}
ctrl := NewDebugController(repo, "my-secret", 3600, logrus.New())
r := newDebugRouter(ctrl)
w := postDebug(r, map[string]interface{}{"user_id": 10})
require.Equal(t, 200, w.Code)

var resp apiResponse
require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
tokenStr, _ := resp.Data["access_token"].(string)
require.NotEmpty(t, tokenStr)

// Verify the token is a real HS256 JWT signed with "my-secret"
parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
return []byte("my-secret"), nil
})
require.NoError(t, err)
assert.True(t, parsed.Valid)
}
