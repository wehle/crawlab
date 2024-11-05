package controllers

import (
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PostUser(c *gin.Context) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}
	u := GetUserFromContext(c)
	model := models.User{
		Username: payload.Username,
		Password: utils.EncryptMd5(payload.Password),
		Role:     payload.Role,
		Email:    payload.Email,
	}
	model.SetCreated(u.Id)
	model.SetUpdated(u.Id)
	id, err := service.NewModelService[models.User]().InsertOne(model)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	result, err := service.NewModelService[models.User]().GetById(id)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccessWithData(c, result)

}

func PostUserChangePassword(c *gin.Context) {
	// get id
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// get payload
	var payload struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// get user
	u := GetUserFromContext(c)
	modelSvc := service.NewModelService[models.User]()

	// update password
	user, err := modelSvc.GetById(id)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	user.SetUpdated(u.Id)
	user.Password = utils.EncryptMd5(payload.Password)
	if err := modelSvc.ReplaceById(user.Id, *user); err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// handle success
	HandleSuccess(c)
}

func GetUserMe(c *gin.Context) {
	u := GetUserFromContext(c)
	_u, err := service.NewModelService[models.User]().GetById(u.Id)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	HandleSuccessWithData(c, _u)
}

func PutUserById(c *gin.Context) {
	// get payload
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// get user
	u := GetUserFromContext(c)

	modelSvc := service.NewModelService[models.User]()

	// update user
	userDb, err := modelSvc.GetById(u.Id)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	user.Password = userDb.Password
	user.SetUpdated(u.Id)
	if user.Id.IsZero() {
		user.Id = u.Id
	}
	if err := modelSvc.ReplaceById(u.Id, user); err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// handle success
	HandleSuccess(c)
}