package user

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/ent"
	"github.com/nvthongswansea/xtreme/internal/ent/user"
	"github.com/nvthongswansea/xtreme/internal/models"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuidUtils"
)

type EntSQLUserRepo struct {
	client   *ent.Client
	uuidTool uuidUtils.UUIDGenerator
}

func NewEntSQLUserRepo(client *ent.Client, uuidTool uuidUtils.UUIDGenerator) EntSQLUserRepo {
	return EntSQLUserRepo{
		client:   client,
		uuidTool: uuidTool,
	}
}

func (e EntSQLUserRepo) InsertNewUser(ctx context.Context, tx transaction.RollbackCommitter, pswHash, username string) (string, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	u, err := client.User.
		Create().
		SetID(e.uuidTool.NewUUID()).
		SetUsername(username).
		SetPasswordHash(pswHash).
		Save(ctx)
	if err != nil {
		return "", err
	}
	return u.ID, nil
}

func (e EntSQLUserRepo) GetUserByUsername(ctx context.Context, tx transaction.RollbackCommitter, username string) (models.User, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	u, err := client.User.Query().Where(user.Username(username)).First(ctx)
	if err != nil {
		return models.User{}, err
	}
	return models.User{
		UUID:      u.ID,
		Username:  u.Username,
		HashPwd:   u.PasswordHash,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (e EntSQLUserRepo) IsUsernameExist(ctx context.Context, tx transaction.RollbackCommitter, username string) (bool, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	count, err := client.User.Query().Where(user.Username(username)).Count(ctx)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}
