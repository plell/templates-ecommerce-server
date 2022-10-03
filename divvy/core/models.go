package core

import (
	"gorm.io/gorm"
)

type ByTheBy struct {
	DeletedByID uint `json:"deleted_by_id"`
	CreatedByID uint `json:"created_by_id"`
	UpdatedByID uint `json:"updated_by_id"`
}

// gorm.Model injects id, deleted_at, created_at, and updated_at
var LOGIN_HISTORY_TABLE = "login_histories"

type LoginHistory struct {
	Username string `json:"username"`
	IP       string `json:"ip"`
	Success  bool   `json:"success"`
	ByTheBy
	gorm.Model
}

var USER_TABLE = "users"

type User struct {
	Username     string `gorm:"type:varchar(100);unique_index;unique;not null" json:"username"`
	Selector     string `json:"selector"`
	Locked       bool   `json:"locked"`
	LockedReason string `json:"lockedReason"`
	Token
	ByTheBy
	gorm.Model
}

var TOKEN_TABLE = "tokens"

type Token struct {
	DisplayName string `json:"displayName"`
	Username    string `gorm:"type:varchar(100);unique_index;unique;not null" json:"username"`
	UserID      uint   `json:"userId"`
	TokenTypeID uint   `json:"tokenTypeId"`
	ByTheBy
	gorm.Model
}

var TOKEN_TYPE_TABLE = "token_types"

type TokenType struct {
	Name string `json:"name"`
	ByTheBy
	gorm.Model
}

var SELECTOR_TABLE = "selectors"

type Selector struct {
	Selector string `json:"selector"`
	Type     string `json:"type"`
	gorm.Model
	ByTheBy
}
