package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"user-management/internal/domain"
	"user-management/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type userRepoMock struct {
	getEmail func(context.Context, string) (*domain.User, error)
	create   func(context.Context, *domain.User) error
	getID    func(context.Context, primitive.ObjectID) (*domain.User, error)
	list     func(context.Context) ([]domain.User, error)
	update   func(context.Context, primitive.ObjectID, domain.UpdateUserInput) (*domain.User, error)
	delete   func(context.Context, primitive.ObjectID) error
	count    func(context.Context) (int64, error)
}

var _ repository.UserRepository = (*userRepoMock)(nil)

func (m *userRepoMock) GetByEmail(c context.Context, e string) (*domain.User, error) {
	if m.getEmail != nil { return m.getEmail(c, e) }; return nil, repository.ErrNotFound
}
func (m *userRepoMock) Create(c context.Context, u *domain.User) error {
	if m.create != nil { return m.create(c, u) }; return nil
}
func (m *userRepoMock) GetByID(c context.Context, id primitive.ObjectID) (*domain.User, error) {
	if m.getID != nil { return m.getID(c, id) }; return nil, repository.ErrNotFound
}
func (m *userRepoMock) List(c context.Context) ([]domain.User, error) {
	if m.list != nil { return m.list(c) }; return nil, nil
}
func (m *userRepoMock) Update(c context.Context, id primitive.ObjectID, in domain.UpdateUserInput) (*domain.User, error) {
	if m.update != nil { return m.update(c, id, in) }; return nil, repository.ErrNotFound
}
func (m *userRepoMock) Delete(c context.Context, id primitive.ObjectID) error {
	if m.delete != nil { return m.delete(c, id) }; return nil
}
func (m *userRepoMock) Count(c context.Context) (int64, error) {
	if m.count != nil { return m.count(c) }; return 0, nil
}

func fixtureUser() *domain.User {
	return &domain.User{ID: primitive.NewObjectID(), Name: "Test User", Email: "test@example.com"}
}
func assertErrIs(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) { t.Fatalf("error = %v, want %v", got, want) }
}

func TestRegister(t *testing.T) {
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		var saved *domain.User
		r := &userRepoMock{
			getEmail: func(_ context.Context, email string) (*domain.User, error) {
				if email != "test@example.com" { t.Errorf("email=%q", email) }
				return nil, repository.ErrNotFound
			},
			create: func(_ context.Context, u *domain.User) error {
				saved = u; u.ID = primitive.NewObjectID(); return nil
			},
		}
		got, err := NewUserService(r, "secret", time.Hour).Register(ctx,
			domain.CreateUserInput{Name:" Test User ", Email:" TEST@EXAMPLE.COM ", Password:"password123"})
		if err != nil { t.Fatal(err) }
		if got != saved || got.Name != "Test User" || got.Email != "test@example.com" { t.Fatalf("got %+v", got) }
		if bcrypt.CompareHashAndPassword([]byte(got.Password), []byte("password123")) != nil { t.Fatal("password not hashed") }
	})
	t.Run("invalid input", func(t *testing.T) {
		_, err := NewUserService(&userRepoMock{}, "secret", time.Hour).Register(ctx, domain.CreateUserInput{Name:"x", Email:"bad", Password:"1"})
		assertErrIs(t, err, ErrInvalidInput)
	})
	t.Run("duplicate email", func(t *testing.T) {
		r := &userRepoMock{getEmail:func(context.Context,string)(*domain.User,error){return fixtureUser(),nil}}
		_, err := NewUserService(r,"secret",time.Hour).Register(ctx,domain.CreateUserInput{Name:"Test User",Email:"test@example.com",Password:"password123"})
		assertErrIs(t,err,ErrEmailExists)
	})
	t.Run("repository lookup error", func(t *testing.T) {
		dbErr:=errors.New("db error"); r:=&userRepoMock{getEmail:func(context.Context,string)(*domain.User,error){return nil,dbErr}}
		_,err:=NewUserService(r,"secret",time.Hour).Register(ctx,domain.CreateUserInput{Name:"Test User",Email:"test@example.com",Password:"password123"})
		assertErrIs(t,err,dbErr)
	})
	t.Run("create error", func(t *testing.T) {
		dbErr:=errors.New("insert error"); r:=&userRepoMock{create:func(context.Context,*domain.User)error{return dbErr}}
		_,err:=NewUserService(r,"secret",time.Hour).Register(ctx,domain.CreateUserInput{Name:"Test User",Email:"test@example.com",Password:"password123"})
		assertErrIs(t,err,dbErr)
	})
}

func TestLogin(t *testing.T) {
	ctx:=context.Background()
	u:=fixtureUser(); hash,_:=bcrypt.GenerateFromPassword([]byte("password123"),bcrypt.DefaultCost); u.Password=string(hash)
	t.Run("success",func(t *testing.T){
		r:=&userRepoMock{getEmail:func(_ context.Context,e string)(*domain.User,error){if e!=u.Email{t.Errorf("email=%s",e)};return u,nil}}
		secret:="test-secret"; tokenString,err:=NewUserService(r,secret,time.Hour).Login(ctx," TEST@EXAMPLE.COM ","password123")
		if err!=nil{t.Fatal(err)}
		token,err:=jwt.ParseWithClaims(tokenString,&jwt.RegisteredClaims{},func(*jwt.Token)(any,error){return []byte(secret),nil})
		if err!=nil||!token.Valid{t.Fatalf("invalid JWT: %v",err)}
		claims:=token.Claims.(*jwt.RegisteredClaims);if claims.Subject!=u.ID.Hex(){t.Fatalf("subject=%s",claims.Subject)}
	})
	t.Run("invalid input",func(t *testing.T){_,err:=NewUserService(&userRepoMock{},"s",time.Hour).Login(ctx,"bad","");assertErrIs(t,err,ErrInvalidCredentials)})
	t.Run("repository error",func(t *testing.T){r:=&userRepoMock{getEmail:func(context.Context,string)(*domain.User,error){return nil,errors.New("db")}};_,err:=NewUserService(r,"s",time.Hour).Login(ctx,u.Email,"password123");assertErrIs(t,err,ErrInvalidCredentials)})
	t.Run("wrong password",func(t *testing.T){r:=&userRepoMock{getEmail:func(context.Context,string)(*domain.User,error){return u,nil}};_,err:=NewUserService(r,"s",time.Hour).Login(ctx,u.Email,"wrongpass");assertErrIs(t,err,ErrInvalidCredentials)})
}

func TestGet(t *testing.T){
	ctx:=context.Background();u:=fixtureUser()
	r:=&userRepoMock{getID:func(_ context.Context,id primitive.ObjectID)(*domain.User,error){if id!=u.ID{t.Errorf("id mismatch")};return u,nil}}
	got,err:=NewUserService(r,"s",time.Hour).Get(ctx,u.ID.Hex());if err!=nil||got!=u{t.Fatalf("got=%v err=%v",got,err)}
	_,err=NewUserService(r,"s",time.Hour).Get(ctx,"bad");assertErrIs(t,err,ErrInvalidInput)
	dbErr:=repository.ErrNotFound;r.getID=func(context.Context,primitive.ObjectID)(*domain.User,error){return nil,dbErr}
	_,err=NewUserService(r,"s",time.Hour).Get(ctx,u.ID.Hex());assertErrIs(t,err,dbErr)
}

func TestListAndCount(t *testing.T){
	ctx:=context.Background();users:=[]domain.User{*fixtureUser()}
	r:=&userRepoMock{list:func(context.Context)([]domain.User,error){return users,nil},count:func(context.Context)(int64,error){return 1,nil}}
	got,err:=NewUserService(r,"s",time.Hour).List(ctx);if err!=nil||len(got)!=1{t.Fatalf("List: %v %v",got,err)}
	n,err:=NewUserService(r,"s",time.Hour).Count(ctx);if err!=nil||n!=1{t.Fatalf("Count: %d %v",n,err)}
	dbErr:=errors.New("db");r.list=func(context.Context)([]domain.User,error){return nil,dbErr};_,err=NewUserService(r,"s",time.Hour).List(ctx);assertErrIs(t,err,dbErr)
	r.count=func(context.Context)(int64,error){return 0,dbErr};_,err=NewUserService(r,"s",time.Hour).Count(ctx);assertErrIs(t,err,dbErr)
}

func TestUpdate(t *testing.T){
	ctx:=context.Background();u:=fixtureUser()
	t.Run("success normalize and same email",func(t *testing.T){
		name:=" Updated User ";email:=" TEST@EXAMPLE.COM "
		r:=&userRepoMock{
			getEmail:func(context.Context,string)(*domain.User,error){return u,nil},
			update:func(_ context.Context,id primitive.ObjectID,in domain.UpdateUserInput)(*domain.User,error){
				if *in.Name!="Updated User"||*in.Email!="test@example.com"{t.Errorf("input not normalized: %+v",in)}
				return u,nil
			},
		}
		got,err:=NewUserService(r,"s",time.Hour).Update(ctx,u.ID.Hex(),domain.UpdateUserInput{Name:&name,Email:&email});if err!=nil||got!=u{t.Fatalf("%v %v",got,err)}
	})
	t.Run("invalid ID and name/email",func(t *testing.T){
		s:=NewUserService(&userRepoMock{},"s",time.Hour)
		_,err:=s.Update(ctx,"bad",domain.UpdateUserInput{});assertErrIs(t,err,ErrInvalidInput)
		name:="x";_,err=s.Update(ctx,u.ID.Hex(),domain.UpdateUserInput{Name:&name});assertErrIs(t,err,ErrInvalidInput)
		email:="bad";_,err=s.Update(ctx,u.ID.Hex(),domain.UpdateUserInput{Email:&email});assertErrIs(t,err,ErrInvalidInput)
	})
	t.Run("email owned by other user",func(t *testing.T){
		email:="other@example.com";other:=fixtureUser();r:=&userRepoMock{getEmail:func(context.Context,string)(*domain.User,error){return other,nil}}
		_,err:=NewUserService(r,"s",time.Hour).Update(ctx,u.ID.Hex(),domain.UpdateUserInput{Email:&email});assertErrIs(t,err,ErrEmailExists)
	})
	t.Run("email lookup error",func(t *testing.T){
		email:="new@example.com";dbErr:=errors.New("db");r:=&userRepoMock{getEmail:func(context.Context,string)(*domain.User,error){return nil,dbErr}}
		_,err:=NewUserService(r,"s",time.Hour).Update(ctx,u.ID.Hex(),domain.UpdateUserInput{Email:&email});assertErrIs(t,err,dbErr)
	})
	t.Run("update repository error",func(t *testing.T){
		dbErr:=errors.New("update");r:=&userRepoMock{update:func(context.Context,primitive.ObjectID,domain.UpdateUserInput)(*domain.User,error){return nil,dbErr}}
		_,err:=NewUserService(r,"s",time.Hour).Update(ctx,u.ID.Hex(),domain.UpdateUserInput{});assertErrIs(t,err,dbErr)
	})
}

func TestDelete(t *testing.T){
	ctx:=context.Background();u:=fixtureUser()
	r:=&userRepoMock{delete:func(_ context.Context,id primitive.ObjectID)error{if id!=u.ID{t.Error("ID mismatch")};return nil}}
	if err:=NewUserService(r,"s",time.Hour).Delete(ctx,u.ID.Hex());err!=nil{t.Fatal(err)}
	assertErrIs(t,NewUserService(r,"s",time.Hour).Delete(ctx,"bad"),ErrInvalidInput)
	dbErr:=repository.ErrNotFound;r.delete=func(context.Context,primitive.ObjectID)error{return dbErr}
	assertErrIs(t,NewUserService(r,"s",time.Hour).Delete(ctx,u.ID.Hex()),dbErr)
}
