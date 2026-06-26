package model

// import (
// 	"go-app/global"
// 	"time"
// )

// const TableNameGoDbRole = "db_roles"

// // GoDbRole mapped from table <db_roles>
// type DbRole struct {
// 	Id        string    `gorm:"column:id;primaryKey;default:gen_random_uuid()" json:"id"`
// 	RoleName  string    `gorm:"column:role_name;not null" json:"role_name"`
// 	RoleNote  string    `gorm:"column:role_note" json:"role_note"`
// 	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
// 	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
// }

// // TableName GoDbRole's table name
// func (*DbRole) TableName() string {
// 	return TableNameGoDbRole
// }

// func init() {
// 	global.RegisterModel(&DbRole{})
// }
