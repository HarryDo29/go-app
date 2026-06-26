package initianlize

import (
	"go-app/global"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitMinio() {
	mConfig := global.Config.Minio

	client, err := minio.New(mConfig.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			mConfig.AccessKey,
			mConfig.SecretKey,
			"",
		),
		Secure: mConfig.UseSSL, // ssl la su dung http hay https
	})
	if err != nil {
		global.Logger.Fatal("Failed to initialize Minio: " + err.Error())
	}

	global.Minio = client
}
