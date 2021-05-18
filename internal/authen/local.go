package authen

import (
	"context"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/nvthongswansea/xtreme/internal/models"
	"github.com/nvthongswansea/xtreme/internal/repository/user"
	"github.com/nvthongswansea/xtreme/pkg/pwd"
)

const (
	usernameAlreadyExistErrorMessage = "username already exists"
	incorrectUsernamePwdErrorMessage = "incorrect username and password"
	jwtAssertionErrorMessage         = "jwt type assertion failed"
	jwtClaimsAssertionErrorMessage   = "jwt claim type assertion failed"
)

type LocalJWTAuthenticator struct {
	userRepo  user.Repository
	pwdUtils  pwd.BCryptHashComparer
	jwtSecret string
}

func (l *LocalJWTAuthenticator) Register(ctx context.Context, username, password string) error {
	isUsernameExist, err := l.userRepo.IsUsernameExist(ctx, username)
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

	err = l.userRepo.InsertNewUser(ctx, username, hashPwd)
	if err != nil {
		return models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return nil
}

func (l *LocalJWTAuthenticator) Login(ctx context.Context, username, password string) (string, error) {
	isUsernameExist, err := l.userRepo.IsUsernameExist(ctx, username)
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
	user, err := l.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !l.pwdUtils.CompareHashAndPwd(password, user.HashPwd) {
		if !isUsernameExist {
			return "", models.XtremeError{
				Code:    models.BadInputErrorCode,
				Message: incorrectUsernamePwdErrorMessage,
			}
		}
	}
	claims := XtremeTokenClaims{
		UserUUID: user.UUID,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: 15000,
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
	return ss, nil
}

func (l *LocalJWTAuthenticator) GetDataViaToken(ctx context.Context, token interface{}) (XtremeTokenClaims, error) {
	jwtToken, ok := token.(*jwt.Token)
	if !ok {
		return XtremeTokenClaims{}, models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: jwtAssertionErrorMessage,
		}
	}
	claims, ok := jwtToken.Claims.(XtremeTokenClaims)
	if !ok {
		return XtremeTokenClaims{}, models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: jwtClaimsAssertionErrorMessage,
		}
	}
	return claims, nil
}

func NewLocalAuthenticator(userRepo user.Repository, pwdUtils pwd.BCryptHashComparer, jwtSecret string) *LocalJWTAuthenticator {
	return &LocalJWTAuthenticator{userRepo: userRepo, pwdUtils: pwdUtils, jwtSecret: jwtSecret}
}
