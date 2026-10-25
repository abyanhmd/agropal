package auth

import (
	"agropal/config"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"agropal/internal/infra/identity"
	"agropal/internal/infra/queue"
	"agropal/pkg"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

/* ==================== Fakes ====================== */

// fakeTransactor runs WithTx's callback directly against a nil Querier
// instead of opening a real transaction. Safe here because the fake
// repositories below never touch the Querier they're given - they stand in
// for the whole repository, not just the SQL layer underneath it.
type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier {
	return nil
}

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

// fakeUserRepository implements dUser.Repository. Only the methods
// auth.Service actually calls are wired to a *Func field; everything else
// panics if called.
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

// fakePublisher implements queue.Publisher.
type fakePublisher struct {
	publishMessageFunc func(ctx context.Context, data []byte) error
}

func (f *fakePublisher) PublishMessage(ctx context.Context, data []byte) error {
	if f.publishMessageFunc == nil {
		panic("fakePublisher: publishMessageFunc not stubbed")
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

// newTestRedis returns a *redis.Client backed by an in-process miniredis
// server instead of a hand-rolled fake: SendVerificationCode/VerifyCode
// rely on real TTL/expiry semantics that are easier to trust from an actual
// (if in-memory) Redis protocol implementation than to hand-simulate.
func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func newLogger() *pkg.Logger {
	return pkg.NewLogger(config.Config{})
}

func newService(userRepo dUser.Repository, authRepo dAuth.Repository, pub *fakePublisher, jwt dAuth.TokenProvider, redisClient *redis.Client) *Service {
	return NewService(newLogger(), fakeTransactor{}, identity.NewGenerator(), jwt, redisClient, pub, authRepo, userRepo)
}

/* ==================== Tests: SendVerificationCode / VerifyCode ==================== */

func TestService_SendVerificationCode(t *testing.T) {
	email := "user@example.com"

	t.Run("publishes the code and stores its hash in redis", func(t *testing.T) {
		redisClient := newTestRedis(t)
		var published queue.EmailVerificationTask
		pub := &fakePublisher{
			publishMessageFunc: func(_ context.Context, data []byte) error {
				return json.Unmarshal(data, &published)
			},
		}
		svc := newService(&fakeUserRepository{}, &fakeAuthRepository{}, pub, &fakeTokenProvider{}, redisClient)

		if err := svc.SendVerificationCode(context.Background(), VerificationPurposeRegister, email); err != nil {
			t.Fatalf("SendVerificationCode() error = %v, want nil", err)
		}

		if published.To != email || published.Purpose != string(VerificationPurposeRegister) || published.Code == "" {
			t.Errorf("published task = %+v, want To=%q Purpose=%q non-empty Code", published, email, VerificationPurposeRegister)
		}

		raw, err := redisClient.Get(context.Background(), "register:"+email).Result()
		if err != nil {
			t.Fatalf("redis.Get() error = %v, want nil", err)
		}
		var record verificationRecord
		if err := json.Unmarshal([]byte(raw), &record); err != nil {
			t.Fatalf("unmarshal record: %v", err)
		}
		if record.CodeHash == published.Code {
			t.Error("redis stores the plaintext code instead of its hash")
		}
		if record.CodeHash != hashVerificationCode(published.Code) {
			t.Error("redis-stored hash does not match the published code")
		}
	})

	t.Run("rate-limits a resend while a fresh code is still active", func(t *testing.T) {
		redisClient := newTestRedis(t)
		key := "register:" + email
		record := verificationRecord{CodeHash: "x", Status: verificationStatusPending, ExpiresAt: time.Now().UTC().Add(verificationCodeTTL)}
		payload, _ := json.Marshal(record)
		if err := redisClient.Set(context.Background(), key, payload, verificationCodeTTL).Err(); err != nil {
			t.Fatalf("seed redis: %v", err)
		}

		pub := &fakePublisher{
			publishMessageFunc: func(context.Context, []byte) error {
				t.Fatal("PublishMessage should not be called while rate-limited")
				return nil
			},
		}
		svc := newService(&fakeUserRepository{}, &fakeAuthRepository{}, pub, &fakeTokenProvider{}, redisClient)

		err := svc.SendVerificationCode(context.Background(), VerificationPurposeRegister, email)
		if !errors.Is(err, dAuth.ErrRateLimited) {
			t.Errorf("SendVerificationCode() error = %v, want %v", err, dAuth.ErrRateLimited)
		}
	})

	t.Run("cleans up the redis key when publishing fails", func(t *testing.T) {
		redisClient := newTestRedis(t)
		wantErr := errors.New("broker unavailable")
		pub := &fakePublisher{
			publishMessageFunc: func(context.Context, []byte) error {
				return wantErr
			},
		}
		svc := newService(&fakeUserRepository{}, &fakeAuthRepository{}, pub, &fakeTokenProvider{}, redisClient)

		err := svc.SendVerificationCode(context.Background(), VerificationPurposeRegister, email)
		if !errors.Is(err, wantErr) {
			t.Errorf("SendVerificationCode() error = %v, want %v", err, wantErr)
		}

		if _, err := redisClient.Get(context.Background(), "register:"+email).Result(); !errors.Is(err, redis.Nil) {
			t.Errorf("redis key still present after publish failure (err = %v, want redis.Nil)", err)
		}
	})
}

func TestService_VerifyCode(t *testing.T) {
	email := "user@example.com"
	code := "abc123"
	key := "register:" + email

	seedRecord := func(t *testing.T, redisClient *redis.Client, record verificationRecord, ttl time.Duration) {
		t.Helper()
		payload, err := json.Marshal(record)
		if err != nil {
			t.Fatalf("marshal record: %v", err)
		}
		if err := redisClient.Set(context.Background(), key, payload, ttl).Err(); err != nil {
			t.Fatalf("seed redis: %v", err)
		}
	}

	t.Run("accepts the correct code and marks it verified", func(t *testing.T) {
		redisClient := newTestRedis(t)
		seedRecord(t, redisClient, verificationRecord{
			CodeHash:  hashVerificationCode(code),
			Status:    verificationStatusPending,
			ExpiresAt: time.Now().UTC().Add(verificationCodeTTL),
		}, verificationCodeTTL)
		svc := newService(&fakeUserRepository{}, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		if err := svc.VerifyCode(context.Background(), VerificationPurposeRegister, email, code); err != nil {
			t.Fatalf("VerifyCode() error = %v, want nil", err)
		}

		raw, _ := redisClient.Get(context.Background(), key).Result()
		var record verificationRecord
		_ = json.Unmarshal([]byte(raw), &record)
		if record.Status != verificationStatusVerified {
			t.Errorf("record.Status = %q, want %q", record.Status, verificationStatusVerified)
		}
	})

	t.Run("rejects a wrong code and increments attempts", func(t *testing.T) {
		redisClient := newTestRedis(t)
		seedRecord(t, redisClient, verificationRecord{
			CodeHash:  hashVerificationCode(code),
			Status:    verificationStatusPending,
			ExpiresAt: time.Now().UTC().Add(verificationCodeTTL),
		}, verificationCodeTTL)
		svc := newService(&fakeUserRepository{}, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		err := svc.VerifyCode(context.Background(), VerificationPurposeRegister, email, "wrong-code")
		if !errors.Is(err, dAuth.ErrVerificationCodeInvalid) {
			t.Errorf("VerifyCode() error = %v, want %v", err, dAuth.ErrVerificationCodeInvalid)
		}

		raw, _ := redisClient.Get(context.Background(), key).Result()
		var record verificationRecord
		_ = json.Unmarshal([]byte(raw), &record)
		if record.Attempts != 1 {
			t.Errorf("record.Attempts = %d, want 1", record.Attempts)
		}
	})

	t.Run("rejects a missing code", func(t *testing.T) {
		redisClient := newTestRedis(t)
		svc := newService(&fakeUserRepository{}, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		err := svc.VerifyCode(context.Background(), VerificationPurposeRegister, email, code)
		if !errors.Is(err, dAuth.ErrVerificationCodeNotFound) {
			t.Errorf("VerifyCode() error = %v, want %v", err, dAuth.ErrVerificationCodeNotFound)
		}
	})

	t.Run("rejects an already-used code", func(t *testing.T) {
		redisClient := newTestRedis(t)
		seedRecord(t, redisClient, verificationRecord{
			CodeHash:  hashVerificationCode(code),
			Status:    verificationStatusVerified,
			ExpiresAt: time.Now().UTC().Add(verificationCodeTTL),
		}, verificationCodeTTL)
		svc := newService(&fakeUserRepository{}, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		err := svc.VerifyCode(context.Background(), VerificationPurposeRegister, email, code)
		if !errors.Is(err, dAuth.ErrVerificationCodeUsed) {
			t.Errorf("VerifyCode() error = %v, want %v", err, dAuth.ErrVerificationCodeUsed)
		}
	})

	t.Run("rejects an expired code", func(t *testing.T) {
		redisClient := newTestRedis(t)
		seedRecord(t, redisClient, verificationRecord{
			CodeHash:  hashVerificationCode(code),
			Status:    verificationStatusPending,
			ExpiresAt: time.Now().UTC().Add(-time.Minute),
		}, verificationCodeTTL)
		svc := newService(&fakeUserRepository{}, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		err := svc.VerifyCode(context.Background(), VerificationPurposeRegister, email, code)
		if !errors.Is(err, dAuth.ErrVerificationCodeExpired) {
			t.Errorf("VerifyCode() error = %v, want %v", err, dAuth.ErrVerificationCodeExpired)
		}
	})

	t.Run("rejects after too many attempts", func(t *testing.T) {
		redisClient := newTestRedis(t)
		seedRecord(t, redisClient, verificationRecord{
			CodeHash:  hashVerificationCode(code),
			Status:    verificationStatusPending,
			ExpiresAt: time.Now().UTC().Add(verificationCodeTTL),
			Attempts:  verificationMaxAttempts,
		}, verificationCodeTTL)
		svc := newService(&fakeUserRepository{}, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		err := svc.VerifyCode(context.Background(), VerificationPurposeRegister, email, code)
		if !errors.Is(err, dAuth.ErrTooManyAttempts) {
			t.Errorf("VerifyCode() error = %v, want %v", err, dAuth.ErrTooManyAttempts)
		}
	})
}

/* ==================== Tests: Register / VerifyRegistration ==================== */

func TestService_Register(t *testing.T) {
	t.Run("stages the hashed user in redis and sends a verification code", func(t *testing.T) {
		redisClient := newTestRedis(t)
		// createUserFunc is deliberately left unstubbed: Register must not
		// touch Postgres at all - the account is only created once
		// VerifyRegistration confirms the code.
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return dUser.User{}, nil
			},
		}
		pub := &fakePublisher{publishMessageFunc: func(context.Context, []byte) error { return nil }}
		svc := newService(userRepo, &fakeAuthRepository{}, pub, &fakeTokenProvider{}, redisClient)

		cmd := CreateUserCommand{Name: "Jane", Email: "jane@example.com", Username: "jane", Password: "Str0ng!Pass"}
		if err := svc.Register(context.Background(), cmd); err != nil {
			t.Fatalf("Register() error = %v, want nil", err)
		}

		staged, err := redisClient.Get(context.Background(), registerPendingKey(cmd.Email)).Bytes()
		if err != nil {
			t.Fatalf("staged user not found in redis under %q: %v", registerPendingKey(cmd.Email), err)
		}
		var stagedUser dUser.User
		if err := json.Unmarshal(staged, &stagedUser); err != nil {
			t.Fatalf("unmarshal staged user: %v", err)
		}

		if stagedUser.Account.Email != "jane@example.com" {
			t.Errorf("stagedUser.Account.Email = %q, want %q", stagedUser.Account.Email, "jane@example.com")
		}
		if stagedUser.Security.Password == "Str0ng!Pass" {
			t.Error("stagedUser.Security.Password is the plaintext password, want a bcrypt hash")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(stagedUser.Security.Password), []byte("Str0ng!Pass")); err != nil {
			t.Errorf("stored password does not verify against the original: %v", err)
		}
	})

	t.Run("leaves no usable registration when sending the verification code fails", func(t *testing.T) {
		redisClient := newTestRedis(t)
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return dUser.User{}, nil
			},
		}
		pub := &fakePublisher{publishMessageFunc: func(context.Context, []byte) error { return errors.New("broker unavailable") }}
		svc := newService(userRepo, &fakeAuthRepository{}, pub, &fakeTokenProvider{}, redisClient)

		cmd := CreateUserCommand{Name: "Jane", Email: "jane@example.com", Username: "jane", Password: "Str0ng!Pass"}
		err := svc.Register(context.Background(), cmd)
		if err == nil {
			t.Fatal("Register() error = nil, want error when SendVerificationCode fails")
		}

		// SendVerificationCode cleans up its own code key on a failed
		// publish, so VerifyRegistration has nothing to verify against even
		// though the staged user blob may still be sitting in redis.
		if _, err := redisClient.Get(context.Background(), "register:"+cmd.Email).Result(); !errors.Is(err, redis.Nil) {
			t.Errorf("verification code key still present after a failed send: err = %v", err)
		}
	})

	t.Run("rejects an already-registered email", func(t *testing.T) {
		redisClient := newTestRedis(t)
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return dUser.User{ID: uuid.New()}, nil
			},
		}
		svc := newService(userRepo, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		cmd := CreateUserCommand{Name: "Jane", Email: "jane@example.com", Username: "jane", Password: "Str0ng!Pass"}
		err := svc.Register(context.Background(), cmd)
		if !errors.Is(err, dAuth.ErrAlreadyRegistered) {
			t.Errorf("Register() error = %v, want %v", err, dAuth.ErrAlreadyRegistered)
		}
	})
}

func TestService_VerifyRegistration(t *testing.T) {
	email := "jane@example.com"
	code := "abc123"
	userID := uuid.New()

	redisClient := newTestRedis(t)
	payload, _ := json.Marshal(verificationRecord{
		CodeHash:  hashVerificationCode(code),
		Status:    verificationStatusPending,
		ExpiresAt: time.Now().UTC().Add(verificationCodeTTL),
	})
	if err := redisClient.Set(context.Background(), "register:"+email, payload, verificationCodeTTL).Err(); err != nil {
		t.Fatalf("seed redis: %v", err)
	}
	stagedUser, _ := json.Marshal(dUser.User{ID: userID, Account: dUser.AccountInfo{Email: email}})
	if err := redisClient.Set(context.Background(), registerPendingKey(email), stagedUser, verificationCodeTTL).Err(); err != nil {
		t.Fatalf("seed redis: %v", err)
	}

	var createdUser, verifiedID uuid.UUID
	// findUserByEmailFunc is deliberately left unstubbed: VerifyRegistration
	// must build the account from the staged redis payload, not look it up.
	userRepo := &fakeUserRepository{
		createUserFunc: func(_ context.Context, _ database.Querier, u dUser.User) error {
			createdUser = u.ID
			return nil
		},
		verifyUserEmailFunc: func(_ context.Context, _ database.Querier, id uuid.UUID, _ time.Time) error {
			verifiedID = id
			return nil
		},
	}
	wantTokenPair := dAuth.TokenPair{
		AccessTokenStr: "access-token",
		RefreshToken:   dAuth.RefreshToken{ID: "refresh-id", UserID: userID},
	}
	tokenProvider := &fakeTokenProvider{
		generateTokenPairFunc: func(context.Context, shared.ClaimInfo) (dAuth.TokenPair, error) {
			return wantTokenPair, nil
		},
	}
	var storedRefreshToken dAuth.RefreshToken
	authRepo := &fakeAuthRepository{
		createRefreshTokenFunc: func(_ context.Context, _ database.Querier, rt dAuth.RefreshToken) error {
			storedRefreshToken = rt
			return nil
		},
	}
	svc := newService(userRepo, authRepo, &fakePublisher{}, tokenProvider, redisClient)

	tokenPair, err := svc.VerifyRegistration(context.Background(), email, code)
	if err != nil {
		t.Fatalf("VerifyRegistration() error = %v, want nil", err)
	}
	if createdUser != userID {
		t.Errorf("CreateUser called with ID = %v, want %v", createdUser, userID)
	}
	if verifiedID != userID {
		t.Errorf("VerifyUserEmail called with %v, want %v", verifiedID, userID)
	}
	if tokenPair.AccessTokenStr != wantTokenPair.AccessTokenStr {
		t.Errorf("VerifyRegistration() token pair = %+v, want %+v", tokenPair, wantTokenPair)
	}
	if storedRefreshToken.UserID != userID {
		t.Errorf("CreateRefreshToken called with UserID = %v, want %v", storedRefreshToken.UserID, userID)
	}
}

/* ==================== Tests: Login / VerifyLogin ==================== */

func TestService_Login(t *testing.T) {
	email := "jane@example.com"
	hashed, _ := bcrypt.GenerateFromPassword([]byte("Str0ng!Pass"), bcrypt.DefaultCost)
	verifiedUser := dUser.User{
		ID:      uuid.New(),
		Account: dUser.AccountInfo{Email: email},
		Security: dUser.SecurityInfo{
			Password: string(hashed),
		},
		Metadata: dUser.UserMetadata{EmailVerifiedAt: shared.Ptr(time.Now().UTC())},
	}

	t.Run("sends a verification code for correct credentials", func(t *testing.T) {
		redisClient := newTestRedis(t)
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return verifiedUser, nil
			},
		}
		var published bool
		pub := &fakePublisher{publishMessageFunc: func(context.Context, []byte) error { published = true; return nil }}
		svc := newService(userRepo, &fakeAuthRepository{}, pub, &fakeTokenProvider{}, redisClient)

		cmd := LoginUserCommand{Email: shared.Ptr(email), Password: "Str0ng!Pass"}
		if err := svc.Login(context.Background(), cmd); err != nil {
			t.Fatalf("Login() error = %v, want nil", err)
		}
		if !published {
			t.Error("Login() did not publish a verification code")
		}
	})

	t.Run("rejects a wrong password", func(t *testing.T) {
		redisClient := newTestRedis(t)
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return verifiedUser, nil
			},
		}
		svc := newService(userRepo, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		cmd := LoginUserCommand{Email: shared.Ptr(email), Password: "wrong-password"}
		err := svc.Login(context.Background(), cmd)
		if !errors.Is(err, dAuth.ErrInvalidPassword) {
			t.Errorf("Login() error = %v, want %v", err, dAuth.ErrInvalidPassword)
		}
	})

	t.Run("rejects an unverified email", func(t *testing.T) {
		redisClient := newTestRedis(t)
		unverified := verifiedUser
		unverified.Metadata.EmailVerifiedAt = nil
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return unverified, nil
			},
		}
		svc := newService(userRepo, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		cmd := LoginUserCommand{Email: shared.Ptr(email), Password: "Str0ng!Pass"}
		err := svc.Login(context.Background(), cmd)
		if !errors.Is(err, dAuth.ErrEmailNotVerified) {
			t.Errorf("Login() error = %v, want %v", err, dAuth.ErrEmailNotVerified)
		}
	})
}

func TestService_VerifyLogin(t *testing.T) {
	email := "jane@example.com"
	code := "abc123"
	userID := uuid.New()

	redisClient := newTestRedis(t)
	payload, _ := json.Marshal(verificationRecord{
		CodeHash:  hashVerificationCode(code),
		Status:    verificationStatusPending,
		ExpiresAt: time.Now().UTC().Add(verificationCodeTTL),
	})
	if err := redisClient.Set(context.Background(), "login:"+email, payload, verificationCodeTTL).Err(); err != nil {
		t.Fatalf("seed redis: %v", err)
	}

	userRepo := &fakeUserRepository{
		findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
			return dUser.User{ID: userID, Account: dUser.AccountInfo{Email: email}}, nil
		},
	}
	wantPair := dAuth.TokenPair{AccessTokenStr: "access-token", RefreshToken: dAuth.RefreshToken{ID: "refresh-id", UserID: userID}}
	jwt := &fakeTokenProvider{
		generateTokenPairFunc: func(_ context.Context, claims shared.ClaimInfo) (dAuth.TokenPair, error) {
			if claims.UserID != userID {
				t.Fatalf("GenerateTokenPair called with UserID=%v, want %v", claims.UserID, userID)
			}
			return wantPair, nil
		},
	}
	var storedToken dAuth.RefreshToken
	authRepo := &fakeAuthRepository{
		createRefreshTokenFunc: func(_ context.Context, _ database.Querier, rt dAuth.RefreshToken) error {
			storedToken = rt
			return nil
		},
	}
	svc := newService(userRepo, authRepo, &fakePublisher{}, jwt, redisClient)

	got, err := svc.VerifyLogin(context.Background(), email, code)
	if err != nil {
		t.Fatalf("VerifyLogin() error = %v, want nil", err)
	}
	if got != wantPair {
		t.Errorf("VerifyLogin() = %+v, want %+v", got, wantPair)
	}
	if storedToken.ID != wantPair.RefreshToken.ID {
		t.Errorf("CreateRefreshToken called with %+v, want %+v", storedToken, wantPair.RefreshToken)
	}
}

/* ==================== Tests: Refresh / Invalidate ==================== */

func TestService_Refresh(t *testing.T) {
	userID := uuid.New()
	refreshID := "refresh-id"

	t.Run("rotates the refresh token", func(t *testing.T) {
		redisClient := newTestRedis(t)
		jwt := &fakeTokenProvider{
			validateTokenFunc: func(_ context.Context, tokenType shared.TokenType, token string) (shared.ClaimInfo, error) {
				if tokenType != shared.TypeRefresh || token != refreshID {
					t.Fatalf("ValidateToken called with (%v, %q), want (%v, %q)", tokenType, token, shared.TypeRefresh, refreshID)
				}
				return shared.ClaimInfo{UserID: userID}, nil
			},
			generateTokenPairFunc: func(context.Context, shared.ClaimInfo) (dAuth.TokenPair, error) {
				return dAuth.TokenPair{AccessTokenStr: "new-access", RefreshToken: dAuth.RefreshToken{ID: "new-refresh", UserID: userID}}, nil
			},
		}
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		var deletedID string
		var created dAuth.RefreshToken
		authRepo := &fakeAuthRepository{
			getRefreshTokenFunc: func(context.Context, database.Querier, string) (dAuth.RefreshToken, error) {
				return dAuth.RefreshToken{ID: refreshID, UserID: userID}, nil
			},
			deleteRefreshTokenFunc: func(_ context.Context, _ database.Querier, id string) error {
				deletedID = id
				return nil
			},
			createRefreshTokenFunc: func(_ context.Context, _ database.Querier, rt dAuth.RefreshToken) error {
				created = rt
				return nil
			},
		}
		svc := newService(userRepo, authRepo, &fakePublisher{}, jwt, redisClient)

		got, err := svc.Refresh(context.Background(), refreshID)
		if err != nil {
			t.Fatalf("Refresh() error = %v, want nil", err)
		}
		if got.RefreshToken.ID != "new-refresh" {
			t.Errorf("Refresh() returned RefreshToken.ID = %q, want %q", got.RefreshToken.ID, "new-refresh")
		}
		if deletedID != refreshID {
			t.Errorf("DeleteRefreshToken called with %q, want %q", deletedID, refreshID)
		}
		if created.ID != "new-refresh" {
			t.Errorf("CreateRefreshToken called with %+v, want ID=new-refresh", created)
		}
	})

	t.Run("treats an unknown refresh token as revoked", func(t *testing.T) {
		redisClient := newTestRedis(t)
		jwt := &fakeTokenProvider{
			validateTokenFunc: func(context.Context, shared.TokenType, string) (shared.ClaimInfo, error) {
				return shared.ClaimInfo{UserID: userID}, nil
			},
		}
		authRepo := &fakeAuthRepository{
			getRefreshTokenFunc: func(context.Context, database.Querier, string) (dAuth.RefreshToken, error) {
				return dAuth.RefreshToken{}, errors.New("not found")
			},
		}
		svc := newService(&fakeUserRepository{}, authRepo, &fakePublisher{}, jwt, redisClient)

		_, err := svc.Refresh(context.Background(), refreshID)
		if !errors.Is(err, dAuth.ErrTokenRevoked) {
			t.Errorf("Refresh() error = %v, want %v", err, dAuth.ErrTokenRevoked)
		}
	})

	t.Run("rejects a claims/stored-token user mismatch", func(t *testing.T) {
		redisClient := newTestRedis(t)
		jwt := &fakeTokenProvider{
			validateTokenFunc: func(context.Context, shared.TokenType, string) (shared.ClaimInfo, error) {
				return shared.ClaimInfo{UserID: userID}, nil
			},
		}
		authRepo := &fakeAuthRepository{
			getRefreshTokenFunc: func(context.Context, database.Querier, string) (dAuth.RefreshToken, error) {
				return dAuth.RefreshToken{ID: refreshID, UserID: uuid.New()}, nil
			},
		}
		svc := newService(&fakeUserRepository{}, authRepo, &fakePublisher{}, jwt, redisClient)

		_, err := svc.Refresh(context.Background(), refreshID)
		if !errors.Is(err, dAuth.ErrTokenInvalid) {
			t.Errorf("Refresh() error = %v, want %v", err, dAuth.ErrTokenInvalid)
		}
	})
}

func TestService_Invalidate(t *testing.T) {
	userID := uuid.New()
	refreshID := "refresh-id"

	redisClient := newTestRedis(t)
	jwt := &fakeTokenProvider{
		validateTokenFunc: func(context.Context, shared.TokenType, string) (shared.ClaimInfo, error) {
			return shared.ClaimInfo{UserID: userID}, nil
		},
	}
	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: userID}, nil
		},
	}
	var deletedID string
	authRepo := &fakeAuthRepository{
		getRefreshTokenFunc: func(context.Context, database.Querier, string) (dAuth.RefreshToken, error) {
			return dAuth.RefreshToken{ID: refreshID, UserID: userID}, nil
		},
		deleteRefreshTokenFunc: func(_ context.Context, _ database.Querier, id string) error {
			deletedID = id
			return nil
		},
	}
	svc := newService(userRepo, authRepo, &fakePublisher{}, jwt, redisClient)

	if err := svc.Invalidate(context.Background(), refreshID); err != nil {
		t.Fatalf("Invalidate() error = %v, want nil", err)
	}
	if deletedID != refreshID {
		t.Errorf("DeleteRefreshToken called with %q, want %q", deletedID, refreshID)
	}
}

/* ==================== Tests: RequestPasswordReset / ConfirmPasswordReset ==================== */

func TestService_RequestPasswordReset(t *testing.T) {
	t.Run("silently succeeds for an unknown email", func(t *testing.T) {
		redisClient := newTestRedis(t)
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return dUser.User{}, dUser.ErrNotFound
			},
		}
		pub := &fakePublisher{publishMessageFunc: func(context.Context, []byte) error {
			t.Fatal("PublishMessage should not be called for an unknown email")
			return nil
		}}
		svc := newService(userRepo, &fakeAuthRepository{}, pub, &fakeTokenProvider{}, redisClient)

		if err := svc.RequestPasswordReset(context.Background(), "nobody@example.com"); err != nil {
			t.Fatalf("RequestPasswordReset() error = %v, want nil", err)
		}
	})

	t.Run("sends a code for a known email", func(t *testing.T) {
		redisClient := newTestRedis(t)
		userRepo := &fakeUserRepository{
			findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
				return dUser.User{ID: uuid.New()}, nil
			},
		}
		var published bool
		pub := &fakePublisher{publishMessageFunc: func(context.Context, []byte) error { published = true; return nil }}
		svc := newService(userRepo, &fakeAuthRepository{}, pub, &fakeTokenProvider{}, redisClient)

		if err := svc.RequestPasswordReset(context.Background(), "jane@example.com"); err != nil {
			t.Fatalf("RequestPasswordReset() error = %v, want nil", err)
		}
		if !published {
			t.Error("RequestPasswordReset() did not publish a verification code")
		}
	})
}

func TestService_ConfirmPasswordReset(t *testing.T) {
	email := "jane@example.com"
	code := "abc123"
	userID := uuid.New()

	redisClient := newTestRedis(t)
	payload, _ := json.Marshal(verificationRecord{
		CodeHash:  hashVerificationCode(code),
		Status:    verificationStatusPending,
		ExpiresAt: time.Now().UTC().Add(verificationCodeTTL),
	})
	if err := redisClient.Set(context.Background(), "reset_password:"+email, payload, verificationCodeTTL).Err(); err != nil {
		t.Fatalf("seed redis: %v", err)
	}

	userRepo := &fakeUserRepository{
		findUserByEmailFunc: func(context.Context, database.Querier, string) (dUser.User, error) {
			return dUser.User{ID: userID}, nil
		},
	}
	var savedUserID uuid.UUID
	var savedPassword string
	userRepo.updateUserPasswordFunc = func(_ context.Context, _ database.Querier, id uuid.UUID, password string) error {
		savedUserID = id
		savedPassword = password
		return nil
	}
	svc := newService(userRepo, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

	if err := svc.ConfirmPasswordReset(context.Background(), email, code, "NewStr0ng!Pass"); err != nil {
		t.Fatalf("ConfirmPasswordReset() error = %v, want nil", err)
	}
	if savedUserID != userID {
		t.Errorf("UpdateUserPassword called with %v, want %v", savedUserID, userID)
	}
	if savedPassword == "NewStr0ng!Pass" || savedPassword == "" {
		t.Errorf("savedPassword = %q, want a bcrypt hash", savedPassword)
	}
}

/* ==================== Tests: Authorize ==================== */

func TestService_Authorize(t *testing.T) {
	redisClient := newTestRedis(t)

	t.Run("admin bypasses the required role", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{Security: dUser.SecurityInfo{Role: "ADMIN"}}, nil
			},
		}
		svc := newService(userRepo, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		if err := svc.Authorize(context.Background(), uuid.New(), "USER"); err != nil {
			t.Errorf("Authorize() error = %v, want nil", err)
		}
	})

	t.Run("allows a matching role", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{Security: dUser.SecurityInfo{Role: "USER"}}, nil
			},
		}
		svc := newService(userRepo, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		if err := svc.Authorize(context.Background(), uuid.New(), "USER"); err != nil {
			t.Errorf("Authorize() error = %v, want nil", err)
		}
	})

	t.Run("forbids a mismatched role", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{Security: dUser.SecurityInfo{Role: "USER"}}, nil
			},
		}
		svc := newService(userRepo, &fakeAuthRepository{}, &fakePublisher{}, &fakeTokenProvider{}, redisClient)

		err := svc.Authorize(context.Background(), uuid.New(), "ADMIN")
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("Authorize() error = %v, want %v", err, shared.ErrForbidden)
		}
	})
}
