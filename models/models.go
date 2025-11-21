package models

import (
	"errors"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id            primitive.ObjectID `bson:"_id" json:"id"`
	UserId        string             `bson:"user_id" json:"user_id"`
	GroupCode     string             `bson:"group_code" json:"group_code"`
	CompanyCode   []string           `bson:"company_code" json:"company_code"`
	Email         string             `bson:"email" json:"email"`
	Role          string             `bson:"role" json:"role"`
	PasswordReset bool               `bson:"password_reset" json:"password_reset"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

func NewUser() *User {
	return &User{}
}

func (user *User) Bind(
	userId string,
	groupCode string,
	companyCode string,
	email string,
	role string,
) error {
	user.Id = primitive.NewObjectIDFromTimestamp(time.Now())
	user.UserId = userId
	user.GroupCode = groupCode
	user.CompanyCode = append(user.CompanyCode, companyCode)
	user.Email = email
	user.Role = role
	user.PasswordReset = true

	return nil
}

func (existingUser *User) BindUpdate(
	companyCode string,
) error {
	if !slices.Contains(existingUser.CompanyCode, companyCode) {
		return errors.New("no update required")
	}

	existingUser.CompanyCode = append(existingUser.CompanyCode, companyCode)
	existingUser.UpdatedAt = time.Now()

	return nil
}
