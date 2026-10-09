package user

import userv1 "github.com/crylabsorg/prostack-go-showcase/pkg/proto/user/v1"

type createUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

func (request createUserRequest) ToProto() *userv1.CreateUserRequest {
	protoRequest := &userv1.CreateUserRequest{}
	protoRequest.SetFirstName(request.FirstName)
	protoRequest.SetLastName(request.LastName)
	protoRequest.SetEmail(request.Email)
	return protoRequest
}

type userResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

func newUserResponse(user *userv1.User) userResponse {
	return userResponse{user.GetId(), user.GetFirstName(), user.GetLastName(), user.GetEmail()}
}
