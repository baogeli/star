// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Users is the golang structure for table users.
type Users struct {
	Id        uint        `json:"id"        orm:"id"         description:""` //
	Username  string      `json:"username"  orm:"username"   description:""` //
	Password  string      `json:"password"  orm:"password"   description:""` //
	Email     string      `json:"email"     orm:"email"      description:""` //
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""` //
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""` //
}
