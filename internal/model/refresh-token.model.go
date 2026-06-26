package model

// import (
// 	"go-app/global"
// 	"time"
// )

// const TableNameGoDbRefreshToken = "db_refresh_tokens"

// // DbRefreshToken mapped from table <db_refesh_tokens>
// type DbRefreshToken struct {
// 	Id        string    `gorm:"column:id;primaryKey;default:gen_random_uuid()" json:"id"`
// 	UserId    string    `gorm:"column:user_id;not_null" json:"user_id"`
// 	Token     string    `gorm:"column:token;not_null" json:"token"`
// 	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
// 	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
// }

// // TableName GoDbUser's table name
// // prevent Go parse table name of struct
// func (*DbRefreshToken) TableName() string {
// 	return TableNameGoDbRefreshToken
// }

// // hàm init tự động đưa
// func init() {
// 	global.RegisterModel(&DbRefreshToken{})
// }
