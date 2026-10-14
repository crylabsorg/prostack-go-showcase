package user

import userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"

func (request CreateUserRequest) ToProto() *userv1.CreateUserRequest {
	protoRequest := &userv1.CreateUserRequest{}
	protoRequest.SetFirstName(request.FirstName)
	protoRequest.SetLastName(request.LastName)
	protoRequest.SetEmail(request.Email)
	return protoRequest
}

func newUserResponse(user *userv1.User) UserResponse {
	return UserResponse{
		ID:        user.GetId(),
		FirstName: user.GetFirstName(),
		LastName:  user.GetLastName(),
		Email:     user.GetEmail(),
	}
}
