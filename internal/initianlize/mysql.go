package initianlize

// import (
// 	"fmt"
// 	"go-app/global"
// 	_ "go-app/internal/model"
// 	"time"

// 	"go.uber.org/zap"
// 	"gorm.io/driver/mysql"
// 	"gorm.io/gen"
// 	"gorm.io/gorm"
// )

// func checkErrorPanic(err error, errString string) {
// 	if err != nil {
// 		global.Logger.Error(errString, zap.Error(err))
// 		panic(err)
// 	}
// }

// func InitMySql() {
// 	m := global.Config.MySQL
// 	dsn := "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local"
// 	var s = fmt.Sprintf(dsn,
// 		m.User,
// 		m.Password,
// 		m.Host,
// 		m.Port,
// 		m.Database,
// 	)
// 	db, err := gorm.Open(mysql.Open(s), &gorm.Config{
// 		SkipDefaultTransaction: true,
// 	})
// 	checkErrorPanic(err, "Init MySql connection failed")

// 	global.Mdb = db
// 	global.Logger.Info("Init MySql connection success")

// 	SetPool()
// 	migrateTables()
// 	// genTableDAO()
// }

// // mở nhóm kết nối database
// func SetPool() {
// 	m := global.Config.MySQL
// 	sqlDb, err := global.Mdb.DB()
// 	if err != nil {
// 		fmt.Println("MySQL error: " + err.Error())
// 		return
// 	}
// 	sqlDb.SetMaxIdleConns(m.MaxIdleConns)                      // số kết nối tối đa trong pool
// 	sqlDb.SetMaxOpenConns(m.MaxOpenConns)                      // số kết nối tối đa với database
// 	sqlDb.SetConnMaxLifetime(time.Duration(m.ConnMaxLifetime)) // thời gian sống của kết nối
// }

// // dùng để tự động tạo table từ struct vô db
// func migrateTables() {
// 	err := global.Mdb.AutoMigrate(
// 		// &model.GoCrmUser{},
// 		// &model.GoDbUser{},
// 		// &model.GoDbRole{},
// 		global.ModelsToMigrate...,
// 	)

// 	if err != nil {
// 		global.Logger.Error("AutoMigrate tables failed", zap.Error(err))
// 		panic(err)
// 	}

// 	global.Logger.Info("AutoMigrate tables success")
// }

// // dùng để gen struct của table trong db
// func genTableDAO() {
// 	g := gen.NewGenerator(gen.Config{
// 		OutPath: "./internal/model",
// 		Mode: gen.WithoutContext |
// 			gen.WithDefaultQuery |
// 			gen.WithQueryInterface, // generate mode
// 	})

// 	g.UseDB(global.Mdb) // reuse your gorm db
// 	g.GenerateAllTable()
// 	g.Execute()
// }
