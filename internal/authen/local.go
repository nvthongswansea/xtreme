package authen

import (
	"context"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/nvthongswansea/xtreme/internal/models"
	"github.com/nvthongswansea/xtreme/internal/repository/directory"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
	"github.com/nvthongswansea/xtreme/internal/repository/user"
	"github.com/nvthongswansea/xtreme/pkg/pwd"
	"time"
)

const (
	usernameAlreadyExistErrorMessage = "username already exists"
	incorrectUsernamePwdErrorMessage = "incorrect username and password"
	jwtAssertionErrorMessage         = "jwt type assertion failed"
	jwtClaimsAssertionErrorMessage   = "jwt claim type assertion failed"
)

type LocalJWTAuthenticator struct {
	userRepo  user.Repository
	txRepo    transaction.TxRepository
	rootDir   directory.Inserter
	pwdUtils  pwd.BCryptHashComparer
	jwtSecret string
}

func (l *LocalJWTAuthenticator) Register(ctx context.Context, username, password string) error {
	var err error
	tx, err := l.txRepo.StartTransaction(ctx)
	if err != nil {
		return models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	defer func(err error) {
		l.txRepo.FinishTransaction(tx, err)
	}(err)
	isUsernameExist, err := l.userRepo.IsUsernameExist(ctx, tx, username)
	if err != nil {
		return models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if isUsernameExist {
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: usernameAlreadyExistErrorMessage,
		}
	}
	hashPwd, err := l.pwdUtils.GetPwdHash(password)
	if err != nil {
		return models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}

	userUUID, err := l.userRepo.InsertNewUser(ctx, tx, hashPwd, username)
	if err != nil {
		return models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}

	// Create new user's root directory
	err = l.rootDir.InsertRootDirectory(ctx, tx, userUUID)
	if err != nil {
		return models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return nil
}

func (l *LocalJWTAuthenticator) Login(ctx context.Context, username, password string) (string, error) {
	isUsernameExist, err := l.userRepo.IsUsernameExist(ctx, nil, username)
	if err != nil {
		return "", models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isUsernameExist {
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: incorrectUsernamePwdErrorMessage,
		}
	}
	user, err := l.userRepo.GetUserByUsername(ctx, nil, username)
	if err != nil {
		return "", models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !l.pwdUtils.CompareHashAndPwd(password, user.HashPwd) {
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: incorrectUsernamePwdErrorMessage,
		}
	}
	claims := XtremeTokenClaims{
		UserUUID: user.UUID,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(l.jwtSecret))
	if err != nil {
		return "", models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	fmt.Println(token.Claims)
	return ss, nil
}

func (l *LocalJWTAuthenticator) GetDataViaToken(ctx context.Context, token interface{}) (XtremeTokenClaims, error) {
	jwtToken, ok := token.(*jwt.Token)
	if !ok {
		fmt.Println("not ok")
		return XtremeTokenClaims{}, models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: jwtAssertionErrorMessage,
		}
	}
	claims, ok := jwtToken.Claims.(*XtremeTokenClaims)
	if !ok {
		fmt.Println("not ok 2")
		return XtremeTokenClaims{}, models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: jwtClaimsAssertionErrorMessage,
		}
	}
	return *claims, nil
}

func NewLocalAuthenticator(userRepo user.Repository, txRepo transaction.TxRepository, rootDir directory.Inserter, pwdUtils pwd.BCryptHashComparer, jwtSecret string) *LocalJWTAuthenticator {
	return &LocalJWTAuthenticator{
		userRepo:  userRepo,
		rootDir:   rootDir,
		txRepo:    txRepo,
		pwdUtils:  pwdUtils,
		jwtSecret: jwtSecret,
	}
}
