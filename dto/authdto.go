package dto

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token    string            `json:"token"`
	Role     string            `json:"role"`
	Employee *EmployeeResponse `json:"employee"`
	Company  *CompanyResponse  `json:"company"`
}
