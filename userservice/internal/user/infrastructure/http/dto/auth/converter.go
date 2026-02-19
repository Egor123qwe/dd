package auth

import (
	refreshTokenUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/auth/refresh_token"
	registerUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/auth/register"
	signInUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/auth/sign_in"
)

func (dto *SignInReq) Convert() signInUsecase.Req {
	return signInUsecase.Req{
		Email:    dto.Email,
		Password: dto.Password,
	}
}

func ToSignInResp(resp signInUsecase.Resp) SignInResp {
	return SignInResp{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}
}

func (dto *SignInResp) Convert() refreshTokenUsecase.Req {
	return refreshTokenUsecase.Req{
		RefreshToken: dto.RefreshToken,
	}
}

func ToRefreshResp(resp refreshTokenUsecase.Resp) RefreshResp {
	return RefreshResp{
		RefreshToken: resp.RefreshToken,
		AccessToken:  resp.AccessToken,
	}
}

func (dto *RefreshReq) Convert() refreshTokenUsecase.Req {
	return refreshTokenUsecase.Req{
		RefreshToken: dto.RefreshToken,
	}
}

func (dto *RegisterReq) Convert() registerUsecase.Req {
	return registerUsecase.Req{
		Email:    dto.Email,
		Password: dto.Password,
		Username: dto.Username,
		Name:     dto.Name,
		ZipCode:  dto.ZipCode,
		Phone:    dto.Phone,
	}
}

func ToRegisterResp(resp registerUsecase.Resp) SignInResp {
	return SignInResp{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}
}
