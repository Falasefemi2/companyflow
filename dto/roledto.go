package dto

type RoleListRequest struct {
	Page     int    `query:"page" default:"1"`
	PageSize int    `query:"page_size" default:"10"`
	Search   string `query:"search"`
}
