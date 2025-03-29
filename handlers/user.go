package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"manage-products/constants"
	"manage-products/models"
	"manage-products/utils"
	"net/http"
)

type UserHandler struct {
	DB *pg.DB
}

// @Summary      SignUp
// @Description  create account for user to use api
// @Param        request  body  models.CreateUserRequest  true  "Create user request"
// @Success      200  {array}  map[string]interface{}
// @Router       /users/sign-up [post]
func (h *UserHandler) SignUp(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"msg":   "invalid request body",
		})
		return
	}

	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "password must be at least 6 characters",
		})
		return
	}

	userExists := &models.User{} // equivalent to userExists := new(model.User)
	err := h.DB.Model(userExists).Where("email = ?", req.Email).Select()
	if err != nil && err.Error() != constants.ErrorNotFound {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when get user",
		})
		return
	}
	if userExists != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "email already exists",
		})
		return
	}

	password, err := utils.HashPassword(req.Password)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when hash password user",
		})
		return
	}

	newUser := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: password,
		Role:     req.Role,
	}
	_, err = h.DB.Model(&newUser).Insert()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when create user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "sign up successfully",
	})
}

// @Summary      SignIn
// @Description  signin to get token to use api
// @Param        request  body  models.LoginRequest  true  "Login"
// @Success      200  {array}  map[string]interface{}
// @Router       /users/sign-in [post]
func (h *UserHandler) SignIn(c *gin.Context) {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"msg":   "invalid request body",
		})
		return
	}

	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "password must be at least 6 characters",
		})
		return
	}

	user := &models.User{} // equivalent to userExists := new(model.User)
	err := h.DB.Model(user).Where("email = ?", req.Email).Select()
	if err != nil {
		if err.Error() == constants.ErrorNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"msg": "user not exists",
			})
			return
		}

		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when get user",
		})
		return
	}

	if !utils.VerifyPassword(user.Password, req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "incorrect password",
		})
		return
	}

	token, err := utils.GenerateToken(*user)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
