package restapi

import (
	"github.com/gin-gonic/gin" 
	"github.com/irononet/go-exchange/service" 
	"net/http" 
	"time"
)

// POST /users 
func SignUp(ctx *gin.Context){
	var request SignupRequest 
	err := ctx.BindJSON(&request) 
	if err != nil{
		ctx.JSON(http.StatusInternalServerError, newMessageVo(err)) 
		return 
	}

	_, err = service.CreateUser(request.Email, request.Password)
	if err != nil{
		ctx.JSON(http.StatusBadRequest, newMessageVo(err)) 
		return 
	}
	
	
	ctx.JSON(http.StatusOK, nil)
}

// POST /users/accessToken 
func SignIn(ctx *gin.Context){
	var request SignupRequest 
	err := ctx.BindJSON(&request) 
	if err != nil{
		ctx.JSON(http.StatusInternalServerError, newMessageVo(err))
		return 
	}

	token, err := service.RefreshAccessToken(request.Email, request.Password) 
	if err != nil{
		ctx.JSON(http.StatusBadRequest, newMessageVo(err)) 
		return 
	}

	ctx.SetCookie("accessToken", token, 7*24*60*60, "/", "*", false, false)
	ctx.JSON(http.StatusOK, token)
}

// POST /users/token 
func GetToken(ctx *gin.Context){
	var request SignupRequest 
	err := ctx.BindJSON(&request) 
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newMessageVo(err)) 
		return 
	}

	token, err := service.RefreshAccessToken(request.Email, request.Password) 
	if err != nil{
		ctx.JSON(http.StatusBadRequest, newMessageVo(err)) 
		return 
	}
	ctx.JSON(http.StatusOK, token)
}

// POST /users/password 
func ChangePassword(ctx *gin.Context){
	var request changePasswordRequest 
	err := ctx.BindJSON(&request)
	if err != nil{
		ctx.JSON(http.StatusBadRequest, newMessageVo(err)) 
		return 
	}

	// check old password 
	_, err = service.GetUserByPassword(GetCurrentUser(ctx).Email, request.OldPassword) 
	if err != nil{
		ctx.JSON(http.StatusBadRequest, newMessageVo(err)) 
		return 
	}

	// Change password 
	err = service.ChangePassword(GetCurrentUser(ctx).Email, request.NewPassword) 
	if err != nil{
		ctx.JSON(http.StatusInternalServerError, newMessageVo(err)) 
		return 
	}

	ctx.JSON(http.StatusOK, nil)
}

// DELETE /users/accessToken 
func SignOut(ctx *gin.Context){
	ctx.SetCookie("accessToken", "", -1, "/", "*", false, false) 
	ctx.JSON(http.StatusOK, nil)
}

// GET /users/self 
func GetUserSelf(ctx *gin.Context){
	user := GetCurrentUser(ctx) 
	if user == nil{
		ctx.AbortWithStatus(http.StatusUnauthorized) 
		return 
	}

	userVo := &userVo{
		Id: user.Email, 
		Email: user.Email, 
		Name: user.Email, 
		ProfilePhoto: "https://cdn.onlinewebfonts.com/svg/img_139247.png", 
		IsBand: false, 
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, userVo)
}