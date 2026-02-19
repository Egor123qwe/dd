package user

import (
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	createUserUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/create_user"
	updateUserPasswordUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/update_user_password"
	updateUserProfileUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/user/update_user_profile"
	"github.com/Interpuls/ifc2-service-farm/pkg/util"
)

type Converter struct {
}

func NewConverter() *Converter {
	return &Converter{}
}

func (dto *CreateReq) Convert() createUserUsecase.Req {
	return createUserUsecase.Req{
		Username: dto.Username,
		RoleIDs:  dto.RoleIDs,
		Email:    dto.Email,
		Password: dto.Password,
		Name:     dto.Name,
		ZipCode:  dto.ZipCode,
		Phone:    dto.Phone,
	}
}

func (c Converter) ToCreateResp(id int) CreateResp {
	return CreateResp{
		ID: id,
	}
}

func (c Converter) ToGetResp(requesterID int, user entity.User) (GetResp, error) {
	userDTO := user.ToDTO()

	var createdAt, updatedAt *string

	if userDTO.CreatedAt != nil {
		createdAt = util.Ptr(userDTO.CreatedAt.String())
	}

	if userDTO.UpdatedAt != nil {
		updatedAt = util.Ptr(userDTO.UpdatedAt.String())
	}

	return GetResp{
		ID:         userDTO.ID,
		RoleIDs:    userDTO.RoleIDs,
		Username:   userDTO.Username,
		Email:      userDTO.Email,
		Name:       userDTO.Name,
		ZipCode:    userDTO.ZipCode,
		Phone:      userDTO.Phone,
		LanguageID: userDTO.LanguageID,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func (dto *UpdateProfileReq) Convert(userID int) updateUserProfileUsecase.Req {
	return updateUserProfileUsecase.Req{
		ID:         userID,
		RoleIDs:    dto.RoleIDs,
		Username:   dto.Username,
		Email:      dto.Email,
		Name:       dto.Name,
		ZipCode:    dto.ZipCode.Convert(),
		Phone:      dto.Phone.Convert(),
		LanguageID: dto.LanguageID.Convert(),
	}
}

func (dto *UpdatePasswordReq) Convert(userID int) updateUserPasswordUsecase.Req {
	return updateUserPasswordUsecase.Req{
		ID:              userID,
		CurrentPassword: dto.CurrentPassword,
		NewPassword:     dto.NewPassword,
	}
}

func (c Converter) ToGetListResp(requesterID int, users []entity.User) (GetListResp, error) {
	var userResponses []GetResp

	for _, user := range users {
		userDTO := user.ToDTO()

		var createdAt, updatedAt *string

		if userDTO.CreatedAt != nil {
			createdAt = util.Ptr(userDTO.CreatedAt.String())
		}

		if userDTO.UpdatedAt != nil {
			updatedAt = util.Ptr(userDTO.UpdatedAt.String())
		}

		userResponses = append(userResponses, GetResp{
			ID:         userDTO.ID,
			RoleIDs:    userDTO.RoleIDs,
			Username:   userDTO.Username,
			Email:      userDTO.Email,
			Name:       userDTO.Name,
			ZipCode:    userDTO.ZipCode,
			Phone:      userDTO.Phone,
			LanguageID: userDTO.LanguageID,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		})
	}

	return GetListResp{
		Users: userResponses,
	}, nil
}

func (dto *DeleteReq) Convert() int {
	return dto.ID
}
