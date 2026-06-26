package model

// import (
// 	"go-app/global"
// 	"time"
// )

// const TableNameGoDbUser = "db_users"

// // GoDbUser mapped from table <db_users>
// type DbUser struct {
// 	Id        string           `gorm:"column:id;primaryKey;default:gen_random_uuid()" json:"id"`
// 	UserName  string           `gorm:"column:user_name;not_null" json:"user_name"`
// 	Password  string           `gorm:"column:password;not_null" json:"password"`
// 	Email     string           `gorm:"column:email;not_null" json:"email	"`
// 	IsActive  bool             `gorm:"column:is_active;default:true" json:"is_active"`
// 	Role      []DbRole         `gorm:"many2many:go_user_roles;"` // gắn relation với db_role
// 	RfTokens  []DbRefreshToken `gorm:"foreignKey:UserId;references:Id"`
// 	CreatedAt time.Time        `gorm:"column:created_at" json:"created_at"`
// 	UpdatedAt time.Time        `gorm:"column:updated_at" json:"updated_at"`
// }

// // TableName GoDbUser's table name
// // prevent Go parse table name of struct
// func (*DbUser) TableName() string {
// 	return TableNameGoDbUser
// }

// // hàm init tự động đưa
// func init() {
// 	global.RegisterModel(&DbUser{})
// }
