package user

import userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"

func (request CreateUserRequest) ToProto() *userv1.CreateUserRequest {
	protoRequest := &userv1.CreateUserRequest{}
	if request.FirstName != nil {
		protoRequest.SetFirstName(*request.FirstName)
	}
	if request.LastName != nil {
		protoRequest.SetLastName(*request.LastName)
	}
	protoRequest.SetEmail(request.Email)
	return protoRequest
}

func newUserResponse(user *userv1.User) UserResponse {
	response := UserResponse{
		ID:    user.GetId(),
		Email: user.GetEmail(),
	}
	if user.HasFirstName() {
		value := user.GetFirstName()
		response.FirstName = &value
	}
	if user.HasLastName() {
		value := user.GetLastName()
		response.LastName = &value
	}
	return response
}
