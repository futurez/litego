package httplib

import (
	"fmt"
)

//All API respone must inherit this struct.
type ApiResponse struct {
	Access string `json:"access"`
	Code   int    `json:"code"`
	Desc   string `json:"desc"`
}

func (r *ApiResponse) SetAccess(access string) {
	r.Access = access
}

func (r *ApiResponse) SetCode(code int, val string) {
	r.Code = code
	r.Desc = fmt.Sprintf("%s", val)
}
