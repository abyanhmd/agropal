package auth

import (
	"agropal/config"
	aAuth "agropal/internal/app/auth"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"agropal/internal/infra/identity"
	"agropal/internal/infra/queue"
	"agropal/internal/utils"
	"agropal/pkg"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

/* ==================== Fakes ====================== */

// fakeTransactor runs WithTx's callback directly against a nil Querier; the
// fake repositories below never touch the Querier they're given.
type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier { return nil }

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

// fakeUserRepository implements dUser.Repository. Only the methods the
// handler tests exercise are wired to a *Func field; everything else panics
// if called.
type fakeUserRepository struct {
	createUserFunc         func(ctx context.Context, q database.Querier, user dUser.User) error
	findUserByIDFunc       func(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error)
	findUserByEmailFunc    func(ctx context.Context, q database.Querier, email string) (dUser.User, error)
	findUserByUsernameFunc func(ctx context.Context, q database.Querier, username string) (dUser.User, error)
	updateUserPasswordFunc func(ctx context.Context, q database.Querier, userID uuid.UUID, password string) error
	verifyUserEmailFunc    func(ctx context.Context, q database.Querier, userID uuid.UUID, verifiedAt time.Time) error
}

func (f *fakeUserRepository) CreateUser(ctx context.Context, q database.Querier, user dUser.User) error {
	if f.createUserFunc == nil {
		panic("fakeUserRepository: createUserFunc not stubbed")
	}
	return f.createUserFunc(ctx, q, user)
}

func (f *fakeUserRepository) FindUserByID(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error) {
	if f.findUserByIDFunc == nil {
		panic("fakeUserRepository: findUserByIDFunc not stubbed")
	}
	return f.findUserByIDFunc(ctx, q, userID)
}

func (f *fakeUserRepository) FindUserByEmail(ctx context.Context, q database.Querier, email string) (dUser.User, error) {
	if f.findUserByEmailFunc == nil {
		panic("fakeUserRepository: findUserByEmailFunc not stubbed")
	}
	return f.findUserByEmailFunc(ctx, q, email)
}

func (f *fakeUserRepository) FindUserByUsername(ctx context.Context, q database.Querier, username string) (dUser.User, error) {
	if f.findUserByUsernameFunc == nil {
		panic("fakeUserRepository: findUserByUsernameFunc not stubbed")
	}
	return f.findUserByUsernameFunc(ctx, q, username)
}

func (f *fakeUserRepository) ListUsers(context.Context, database.Querier) ([]dUser.User, error) {
	panic("fakeUserRepository: ListUsers not stubbed")
}

func (f *fakeUserRepository) UpdateUser(context.Context, database.Querier, dUser.User) error {
	panic("fakeUserRepository: UpdateUser not stubbed")
}

func (f *fakeUserRepository) UpdateUserPassword(ctx context.Context, q database.Querier, userID uuid.UUID, password string) error {
	if f.updateUserPasswordFunc == nil {
		panic("fakeUserRepository: updateUserPasswordFunc not stubbed")
	}
	return f.updateUserPasswordFunc(ctx, q, userID, password)
}

func (f *fakeUserRepository) VerifyUserEmail(ctx context.Context, q database.Querier, userID uuid.UUID, verifiedAt time.Time) error {
	if f.verifyUserEmailFunc == nil {
		panic("fakeUserRepository: verifyUserEmailFunc not stubbed")
	}
	return f.verifyUserEmailFunc(ctx, q, userID, verifiedAt)
}

func (f *fakeUserRepository) SoftDeleteUser(context.Context, database.Querier, uuid.UUID, time.Time) error {
	panic("fakeUserRepository: SoftDeleteUser not stubbed")
}

func (f *fakeUserRepository) DeleteUser(context.Context, database.Querier, uuid.UUID) error {
	panic("fakeUserRepository: DeleteUser not stubbed")
}

func (f *fakeUserRepository) ListUsersByAsset(context.Context, database.Querier, shared.AssetType, uuid.UUID) ([]dUser.User, error) {
	panic("fakeUserRepository: ListUsersByAsset not stubbed")
}

// fakeAuthRepository implements dAuth.Repository.
type fakeAuthRepository struct {
	createRefreshTokenFunc func(ctx context.Context, q database.Querier, rt dAuth.RefreshToken) error
	getRefreshTokenFunc    func(ctx context.Context, q database.Querier, refreshTokenID string) (dAuth.RefreshToken, error)
	deleteRefreshTokenFunc func(ctx context.Context, q database.Querier, refreshTokenID string) error
}

func (f *fakeAuthRepository) InvalidateSessions(context.Context, database.Querier, uuid.UUID) error {
	panic("fakeAuthRepository: InvalidateSessions not stubbed")
}

func (f *fakeAuthRepository) CreateRefreshToken(ctx context.Context, q database.Querier, rt dAuth.RefreshToken) error {
	if f.createRefreshTokenFunc == nil {
		panic("fakeAuthRepository: createRefreshTokenFunc not stubbed")
	}
	return f.createRefreshTokenFunc(ctx, q, rt)
}

func (f *fakeAuthRepository) GetRefreshToken(ctx context.Context, q database.Querier, refreshTokenID string) (dAuth.RefreshToken, error) {
	if f.getRefreshTokenFunc == nil {
		panic("fakeAuthRepository: getRefreshTokenFunc not stubbed")
	}
	return f.getRefreshTokenFunc(ctx, q, refreshTokenID)
}

func (f *fakeAuthRepository) DeleteRefreshToken(ctx context.Context, q database.Querier, refreshTokenID string) error {
	if f.deleteRefreshTokenFunc == nil {
		panic("fakeAuthRepository: deleteRefreshTokenFunc not stubbed")
	}
	return f.deleteRefreshTokenFunc(ctx, q, refreshTokenID)
}

// fakePublisher implements queue.Publisher. A nil publishMessageFunc
// succeeds silently, which is fine for tests that don't care about the
// queued verification email.
type fakePublisher struct {
	publishMessageFunc func(ctx context.Context, data []byte) error
}

func (f *fakePublisher) PublishMessage(ctx context.Context, data []byte) error {
	if f.publishMessageFunc == nil {
		return nil
	}
	return f.publishMessageFunc(ctx, data)
}

// fakeTokenProvider implements dAuth.TokenProvider.
type fakeTokenProvider struct {
	generateTokenPairFunc func(ctx context.Context, info shared.ClaimInfo) (dAuth.TokenPair, error)
	validateTokenFunc     func(ctx context.Context, tokenType shared.TokenType, token string) (shared.ClaimInfo, error)
}

func (f *fakeTokenProvider) GenerateTokenPair(ctx context.Context, info shared.ClaimInfo) (dAuth.TokenPair, error) {
	if f.generateTokenPairFunc == nil {
		panic("fakeTokenProvider: generateTokenPairFunc not stubbed")
	}
	return f.generateTokenPairFunc(ctx, info)
}

func (f *fakeTokenProvider) ValidateToken(ctx context.Context, tokenType shared.TokenType, token string) (shared.ClaimInfo, error) {
	if f.validateTokenFunc == nil {
		panic("fakeTokenProvider: validateTokenFunc not stubbed")
	}
	return f.validateTokenFunc(ctx, tokenType, token)
}

/* ==================== Helpers ====================== */

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

// newTestHandler wires a real *aAuth.Service (backed by fakes and an
// in-process Redis) into a real *Handler, so these tests exercise the
// handler's actual request parsing, status codes, cookies, and response
// bodies rather than a mocked-out service.
func newTestHandler(t *testing.T, userRepo dUser.Repository, authRepo dAuth.Repository, pub *fakePublisher, jwt dAuth.TokenProvider) *Handler {
	t.Helper()

	if pub == nil {
		pub = &fakePublisher{}
	}

	logger := pkg.NewLogger(config.Config{})
	svc := aAuth.NewService(logger, fakeTransactor{}, identity.NewGenerator(), jwt, newTestRedis(t), pub, authRepo, userRepo)
	ioHandler := utils.NewIOHandler(logger, nil, utils.NewValidator())

	return NewHandler(ioHandler, svc)
}

// capturingPublisher records the plaintext verification code from the last
// published queue.EmailVerificationTask, so a test can drive a real
// Register/Login call and then use the code it generated to verify.
func capturingPublisher() (*fakePublisher, *string) {
	var lastCode string
	pub := &fakePublisher{
		publishMessageFunc: func(_ context.Context, data []byte) error {
			var task queue.EmailVerificationTask
			if err := json.Unmarshal(data, &task); err != nil {
				return err
			}
			lastCode = task.Code
			return nil
		},
	}
	return pub, &lastCode
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) utils.StandardEnvelope {
	t.Helper()

	var envelope utils.StandardEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
	}
	return envelope
}

func cookieFromResponse(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func jsonRequest(method, path, body string) *http.Request {
	return httptest.NewRequest(method, path, bytes.NewBufferString(body))
}

/* ==================== Tests: Register ====================== */

func TestHandler_Register(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return dUser.User{}, nil
			},
			createUserFunc: func(context.Context, database.Querier, dUser.User) error { return nil },
		}
		h := newTestHandler(t, userRepo, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/register", `{"name":"Jane","email":"jane@example.com","username":"jane","password":"Str0ng!Pass1"}`)
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dAuth.SuccessRegister.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dAuth.SuccessRegister.String())
		}
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/register", "not-json")
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrInvalidRequest.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrInvalidRequest.Error())
		}
	})

	t.Run("Missing Required Field Fails Mapping Before Hitting The Service", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				t.Fatal("service should not be called when the request fails mapping")
				return dUser.User{}, nil
			},
		}
		h := newTestHandler(t, userRepo, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/register", `{"name":"Jane","username":"jane","password":"Str0ng!Pass1"}`)
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dAuth.ErrInvalidEmail.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dAuth.ErrInvalidEmail.Error())
		}
	})

	t.Run("Already Registered Email Maps To 409", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return dUser.User{ID: uuid.New()}, nil
			},
		}
		h := newTestHandler(t, userRepo, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/register", `{"name":"Jane","email":"jane@example.com","username":"jane","password":"Str0ng!Pass1"}`)
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dAuth.ErrAlreadyRegistered.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dAuth.ErrAlreadyRegistered.Error())
		}
	})
}

/* ==================== Tests: VerifyRegister ====================== */

func TestHandler_VerifyRegister(t *testing.T) {
	email := "jane@example.com"
	userID := uuid.New()

	t.Run("Success Sets The Refresh Cookie", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return dUser.User{}, nil
			},
			createUserFunc:      func(context.Context, database.Querier, dUser.User) error { return nil },
			verifyUserEmailFunc: func(context.Context, database.Querier, uuid.UUID, time.Time) error { return nil },
		}
		authRepo := &fakeAuthRepository{
			createRefreshTokenFunc: func(context.Context, database.Querier, dAuth.RefreshToken) error { return nil },
		}
		jwt := &fakeTokenProvider{
			generateTokenPairFunc: func(context.Context, shared.ClaimInfo) (dAuth.TokenPair, error) {
				return dAuth.TokenPair{
					AccessTokenStr: "access-token",
					RefreshToken:   dAuth.RefreshToken{ID: "refresh-id", UserID: userID, ExpiresAt: time.Now().Add(time.Hour).Unix()},
				}, nil
			},
		}
		pub, code := capturingPublisher()
		h := newTestHandler(t, userRepo, authRepo, pub, jwt)

		// Drive a real Register call first so a verification code actually
		// exists in Redis; VerifyRegister has nothing to confirm otherwise.
		registerRec := httptest.NewRecorder()
		h.Register(registerRec, jsonRequest(http.MethodPost, "/auth/register", `{"name":"Jane","email":"`+email+`","username":"jane","password":"Str0ng!Pass1"}`))
		if registerRec.Code != http.StatusCreated {
			t.Fatalf("setup Register() status = %d, want %d (body=%s)", registerRec.Code, http.StatusCreated, registerRec.Body.String())
		}
		if *code == "" {
			t.Fatal("setup Register() did not publish a verification code")
		}

		// FindUserByEmail is called again inside VerifyRegistration; return
		// the now-existing user with a real ID this time.
		userRepo.findUserByEmailFunc = func(context.Context, database.Querier, string) (dUser.User, error) {
			return dUser.User{ID: userID, Account: dUser.AccountInfo{Email: email}}, nil
		}

		req := jsonRequest(http.MethodPost, "/auth/register/verify", `{"email":"`+email+`","code":"`+*code+`"}`)
		rec := httptest.NewRecorder()

		h.VerifyRegister(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}

		cookie := cookieFromResponse(rec, "rid")
		if cookie == nil {
			t.Fatal("VerifyRegister() did not set the rid cookie")
		}
		if cookie.Value != "refresh-id" {
			t.Errorf("rid cookie value = %q, want %q", cookie.Value, "refresh-id")
		}
		if !cookie.HttpOnly {
			t.Error("rid cookie is not HttpOnly")
		}
	})

	t.Run("Unknown Code Does Not Set A Cookie", func(t *testing.T) {
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/register/verify", `{"email":"`+email+`","code":"wrong-code"}`)
		rec := httptest.NewRecorder()

		h.VerifyRegister(rec, req)

		if rec.Code == http.StatusOK {
			t.Fatalf("status = %d, want a non-200 status for an unknown code", rec.Code)
		}
		if cookieFromResponse(rec, "rid") != nil {
			t.Error("VerifyRegister() set the rid cookie despite a failed verification")
		}
	})
}

/* ==================== Tests: Login / VerifyLogin ====================== */

func TestHandler_Login(t *testing.T) {
	email := "jane@example.com"
	hashed, err := bcrypt.GenerateFromPassword([]byte("Str0ng!Pass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword() error = %v", err)
	}
	verifiedUser := dUser.User{
		ID:       uuid.New(),
		Account:  dUser.AccountInfo{Email: email},
		Security: dUser.SecurityInfo{Password: string(hashed)},
		Metadata: dUser.UserMetadata{EmailVerifiedAt: shared.Ptr(time.Now())},
	}

	t.Run("Success Sends A Verification Code", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return verifiedUser, nil
			},
		}
		h := newTestHandler(t, userRepo, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/login", `{"email":"`+email+`","password":"Str0ng!Pass1"}`)
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dAuth.SuccessVerificationSent.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dAuth.SuccessVerificationSent.String())
		}
	})

	t.Run("Wrong Password Maps To 401", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return verifiedUser, nil
			},
		}
		h := newTestHandler(t, userRepo, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/login", `{"email":"`+email+`","password":"wrong-password"}`)
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dAuth.ErrInvalidPassword.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dAuth.ErrInvalidPassword.Error())
		}
	})

	t.Run("Missing Email And Username Fails Mapping", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				t.Fatal("service should not be called when the request fails mapping")
				return dUser.User{}, nil
			},
		}
		h := newTestHandler(t, userRepo, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/login", `{"password":"Str0ng!Pass1"}`)
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dAuth.ErrInvalidCredentials.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dAuth.ErrInvalidCredentials.Error())
		}
	})
}

func TestHandler_VerifyLogin(t *testing.T) {
	email := "jane@example.com"
	hashed, err := bcrypt.GenerateFromPassword([]byte("Str0ng!Pass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword() error = %v", err)
	}
	userID := uuid.New()
	verifiedUser := dUser.User{
		ID:       userID,
		Account:  dUser.AccountInfo{Email: email},
		Security: dUser.SecurityInfo{Password: string(hashed)},
		Metadata: dUser.UserMetadata{EmailVerifiedAt: shared.Ptr(time.Now())},
	}

	t.Run("Success Sets The Refresh Cookie", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return verifiedUser, nil
			},
		}
		authRepo := &fakeAuthRepository{
			createRefreshTokenFunc: func(context.Context, database.Querier, dAuth.RefreshToken) error { return nil },
		}
		jwt := &fakeTokenProvider{
			generateTokenPairFunc: func(context.Context, shared.ClaimInfo) (dAuth.TokenPair, error) {
				return dAuth.TokenPair{
					AccessTokenStr: "access-token",
					RefreshToken:   dAuth.RefreshToken{ID: "refresh-id", UserID: userID, ExpiresAt: time.Now().Add(time.Hour).Unix()},
				}, nil
			},
		}
		pub, code := capturingPublisher()
		h := newTestHandler(t, userRepo, authRepo, pub, jwt)

		loginRec := httptest.NewRecorder()
		h.Login(loginRec, jsonRequest(http.MethodPost, "/auth/login", `{"email":"`+email+`","password":"Str0ng!Pass1"}`))
		if loginRec.Code != http.StatusOK {
			t.Fatalf("setup Login() status = %d, want %d (body=%s)", loginRec.Code, http.StatusOK, loginRec.Body.String())
		}
		if *code == "" {
			t.Fatal("setup Login() did not publish a verification code")
		}

		req := jsonRequest(http.MethodPost, "/auth/login/verify", `{"email":"`+email+`","code":"`+*code+`"}`)
		rec := httptest.NewRecorder()

		h.VerifyLogin(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}

		cookie := cookieFromResponse(rec, "rid")
		if cookie == nil {
			t.Fatal("VerifyLogin() did not set the rid cookie")
		}
		if cookie.Value != "refresh-id" {
			t.Errorf("rid cookie value = %q, want %q", cookie.Value, "refresh-id")
		}

		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dAuth.SuccessLogin.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dAuth.SuccessLogin.String())
		}
	})

	t.Run("Invalid Body Does Not Set A Cookie", func(t *testing.T) {
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/login/verify", "not-json")
		rec := httptest.NewRecorder()

		h.VerifyLogin(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		if cookieFromResponse(rec, "rid") != nil {
			t.Error("VerifyLogin() set the rid cookie despite an invalid request body")
		}
	})
}

/* ==================== Tests: Refresh ====================== */

func TestHandler_Refresh(t *testing.T) {
	t.Run("Missing Cookie Maps To 400", func(t *testing.T) {
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		rec := httptest.NewRecorder()

		h.Refresh(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrInvalidRequest.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrInvalidRequest.Error())
		}
	})

	t.Run("Unknown Refresh Token Maps To 401", func(t *testing.T) {
		jwt := &fakeTokenProvider{
			validateTokenFunc: func(context.Context, shared.TokenType, string) (shared.ClaimInfo, error) {
				return shared.ClaimInfo{}, dAuth.ErrTokenInvalid
			},
		}
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, jwt)

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "rid", Value: "some-refresh-id"})
		rec := httptest.NewRecorder()

		h.Refresh(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
		}
	})
}

/* ==================== Tests: Logout ====================== */

func TestHandler_Logout(t *testing.T) {
	t.Run("Missing Cookie Maps To 400", func(t *testing.T) {
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		rec := httptest.NewRecorder()

		h.Logout(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("Success Clears The Refresh Cookie", func(t *testing.T) {
		userID := uuid.New()
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		authRepo := &fakeAuthRepository{
			getRefreshTokenFunc: func(context.Context, database.Querier, string) (dAuth.RefreshToken, error) {
				return dAuth.RefreshToken{ID: "some-refresh-id", UserID: userID}, nil
			},
			deleteRefreshTokenFunc: func(context.Context, database.Querier, string) error { return nil },
		}
		jwt := &fakeTokenProvider{
			validateTokenFunc: func(context.Context, shared.TokenType, string) (shared.ClaimInfo, error) {
				return shared.ClaimInfo{UserID: userID}, nil
			},
		}
		h := newTestHandler(t, userRepo, authRepo, nil, jwt)

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "rid", Value: "some-refresh-id"})
		rec := httptest.NewRecorder()

		h.Logout(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}

		cookie := cookieFromResponse(rec, "rid")
		if cookie == nil {
			t.Fatal("Logout() did not set a clearing rid cookie")
		}
		if cookie.Value != "" {
			t.Errorf("rid cookie value = %q, want empty", cookie.Value)
		}
		if !cookie.Expires.Before(time.Now()) {
			t.Errorf("rid cookie Expires = %v, want a time in the past", cookie.Expires)
		}
	})
}

/* ==================== Tests: Password Reset ====================== */

func TestHandler_RequestPasswordReset(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return dUser.User{ID: uuid.New(), Account: dUser.AccountInfo{Email: "jane@example.com"}}, nil
			},
		}
		h := newTestHandler(t, userRepo, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/password-reset/request", `{"email":"jane@example.com"}`)
		rec := httptest.NewRecorder()

		h.RequestPasswordReset(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dAuth.SuccessPasswordResetEmail.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dAuth.SuccessPasswordResetEmail.String())
		}
	})

	t.Run("Invalid Email Fails Mapping", func(t *testing.T) {
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/password-reset/request", `{"email":"not-an-email"}`)
		rec := httptest.NewRecorder()

		h.RequestPasswordReset(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestHandler_ConfirmPasswordReset(t *testing.T) {
	t.Run("Invalid New Password Fails Mapping Before Hitting The Service", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				t.Fatal("service should not be called when the request fails mapping")
				return dUser.User{}, nil
			},
		}
		h := newTestHandler(t, userRepo, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/password-reset/confirm", `{"email":"jane@example.com","code":"abc123","new_password":"short"}`)
		rec := httptest.NewRecorder()

		h.ConfirmPasswordReset(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dAuth.ErrPasswordTooShort.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dAuth.ErrPasswordTooShort.Error())
		}
	})
}

/* ==================== Tests: ResendVerificationCode ====================== */

func TestHandler_ResendVerificationCode(t *testing.T) {
	t.Run("Success For Register Purpose", func(t *testing.T) {
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/register/resend-code", `{"email":"jane@example.com"}`)
		req.SetPathValue("purpose", "register")
		rec := httptest.NewRecorder()

		h.ResendVerificationCode(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dAuth.SuccessVerificationSent.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dAuth.SuccessVerificationSent.String())
		}
	})

	t.Run("Success For Reset Password Purpose", func(t *testing.T) {
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/reset_password/resend-code", `{"email":"jane@example.com"}`)
		req.SetPathValue("purpose", "reset_password")
		rec := httptest.NewRecorder()

		h.ResendVerificationCode(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	// Unlike Login, this endpoint has no password check in front of it, so
	// a login code must not be resendable through it - that would let
	// anyone trigger a login OTP email for any address with no
	// credentials, bypassing the two-factor design.
	t.Run("Rejects Login Purpose", func(t *testing.T) {
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/login/resend-code", `{"email":"jane@example.com"}`)
		req.SetPathValue("purpose", "login")
		rec := httptest.NewRecorder()

		h.ResendVerificationCode(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})

	t.Run("Invalid Email Fails Mapping", func(t *testing.T) {
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/register/resend-code", `{"email":"not-an-email"}`)
		req.SetPathValue("purpose", "register")
		rec := httptest.NewRecorder()

		h.ResendVerificationCode(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})

	t.Run("Unknown Purpose Rejected", func(t *testing.T) {
		h := newTestHandler(t, &fakeUserRepository{}, &fakeAuthRepository{}, nil, &fakeTokenProvider{})

		req := jsonRequest(http.MethodPost, "/auth/bogus/resend-code", `{"email":"jane@example.com"}`)
		req.SetPathValue("purpose", "bogus")
		rec := httptest.NewRecorder()

		h.ResendVerificationCode(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}
