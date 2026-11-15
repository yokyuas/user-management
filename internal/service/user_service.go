package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"user-management/internal/domain"
	"user-management/internal/repository"
	"user-management/internal/validation"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidInput = errors.New("invalid input")
var ErrEmailExists = errors.New("email already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")

type UserService struct {
	repo      repository.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewUserService(r repository.UserRepository, secret string, ttl time.Duration) *UserService {
	return &UserService{repo: r, jwtSecret: []byte(secret), tokenTTL: ttl}
}

func (s *UserService) Register(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if !validation.Name(input.Name) || !validation.Email(input.Email) || !validation.PasswordRegister(input.Password) {
		return nil, ErrInvalidInput
	}
	if _, err := s.repo.GetByEmail(ctx, input.Email); err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &domain.User{Name: input.Name, Email: input.Email, Password: string(hash), CreatedAt: time.Now().UTC()}
	if err = s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !validation.Email(email) || !validation.PasswordLogin(password) {
		return "", ErrInvalidCredentials
	}
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   u.ID.Hex(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func (s *UserService) Get(ctx context.Context, id string) (*domain.User, error) {
	oid, e := primitive.ObjectIDFromHex(id)
	if e != nil {
		return nil, ErrInvalidInput
	}
	return s.repo.GetByID(ctx, oid)
}

func (s *UserService) List(ctx context.Context) ([]domain.User, error) { return s.repo.List(ctx) }

func (s *UserService) Update(ctx context.Context, id string, in domain.UpdateUserInput) (*domain.User, error) {
	oid, e := primitive.ObjectIDFromHex(id)
	if e != nil {
		return nil, ErrInvalidInput
	}
	if in.Name != nil {
		*in.Name = strings.TrimSpace(*in.Name)
		if !validation.Name(*in.Name) {
			return nil, ErrInvalidInput
		}
	}
	if in.Email != nil {
		*in.Email = strings.ToLower(strings.TrimSpace(*in.Email))
		if !validation.Email(*in.Email) {
			return nil, ErrInvalidInput
		}
		existing, e := s.repo.GetByEmail(ctx, *in.Email)
		if e == nil && existing.ID != oid {
			return nil, ErrEmailExists
		}
		if e != nil && !errors.Is(e, repository.ErrNotFound) {
			return nil, e
		}
	}
	return s.repo.Update(ctx, oid, in)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	oid, e := primitive.ObjectIDFromHex(id)
	if e != nil {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, oid)
}

func (s *UserService) Count(ctx context.Context) (int64, error) { return s.repo.Count(ctx) }
