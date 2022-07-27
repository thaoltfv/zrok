// Code generated by go-swagger; DO NOT EDIT.

package tunnel

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// UntunnelOKCode is the HTTP code returned for type UntunnelOK
const UntunnelOKCode int = 200

/*UntunnelOK tunnel removed

swagger:response untunnelOK
*/
type UntunnelOK struct {
}

// NewUntunnelOK creates UntunnelOK with default headers values
func NewUntunnelOK() *UntunnelOK {

	return &UntunnelOK{}
}

// WriteResponse to the client
func (o *UntunnelOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// UntunnelInternalServerErrorCode is the HTTP code returned for type UntunnelInternalServerError
const UntunnelInternalServerErrorCode int = 500

/*UntunnelInternalServerError internal server error

swagger:response untunnelInternalServerError
*/
type UntunnelInternalServerError struct {
}

// NewUntunnelInternalServerError creates UntunnelInternalServerError with default headers values
func NewUntunnelInternalServerError() *UntunnelInternalServerError {

	return &UntunnelInternalServerError{}
}

// WriteResponse to the client
func (o *UntunnelInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(500)
}
