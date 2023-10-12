package authentication

import (
	"github.com/golang-jwt/jwt/v5"
)

type RegisterDto struct {
	Email    string `json:"email" validate:"required" example:"johndoe@gmail.com"`
	Password string `json:"password" validate:"required" example:"JohnD0@2123"`
}

type User struct {
	Id       string `json:"id" example:"3604fa26-5ee8-428f-a6dd-c742455e8148"`
	Email    string `json:"email" validate:"required,email" example:"johndoe@gmail.com"`
	Password string `json:"password" validate:"required" example:"JohnD0@2123"`
}

type JwtClaims struct {
	Name   string `json:"name"`
	UserId string `json:"user_id"`
	jwt.RegisteredClaims
}

type LoginResponseDto struct {
	Response string `json:"response" example:"User is logged in successfully"`
	Token    string `json:"token" example:"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiZ3Vlc3RAZ2FtaWwuY29tIiwidXNlcl9pZCI6ImJjMjg2OTIzLTdjMGItNDkxOS1hOWZjLTIyMTdmNTdiMTFlNSIsImV4cCI6MTY5NTY1Nzg0OH0.gKKjicgu53ja0dNntSxlsAVsRN9zWvd98YjkMKYhIe7OyIA6MXfZipBNzxcNuHVcdrWvgw4VPNdYXsI3Aa37Mw"`
}
