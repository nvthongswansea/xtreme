package local

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/database"
	fileUtils "github.com/nvthongswansea/xtreme/pkg/file-utils"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuid-utils"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestMultiOSLocalFManServiceHandler_validateSelectedInputArgs(t *testing.T) {
	type fields struct {
		localFManDBRepo database.LocalFManRepository
		uuidTool        uuidUtils.UUIDGenerateValidator
		fileOps         fileUtils.FileSaveReadRemover
		fileCompress    fileUtils.FileCompressor
	}
	type args struct {
		ctx       context.Context
		logger    *log.Entry
		checkType int
		userUUID  string
		dirUUID   string
		fileUUID  string
		filename  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MultiOSLocalFManServiceHandler{
				localFManDBRepo: tt.fields.localFManDBRepo,
				uuidTool:        tt.fields.uuidTool,
				fileOps:         tt.fields.fileOps,
				fileCompress:    tt.fields.fileCompress,
			}
			if err := m.validateSelectedInputArgs(tt.args.ctx, tt.args.logger, tt.args.checkType, tt.args.userUUID, tt.args.dirUUID, tt.args.fileUUID, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("validateSelectedInputArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
