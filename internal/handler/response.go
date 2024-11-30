package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type Response struct {
	Error    *ErrorResponse `json:"error,omitempty"`
	Response interface{}    `json:"response,omitempty"`
	Data     interface{}    `json:"data,omitempty"`
}

func newResponse(c *gin.Context, httpStatus int, err *ErrorResponse, response interface{}, data interface{}) {
	resp := Response{
		Error:    err,
		Response: response,
		Data:     data,
	}

	if err != nil {
		logrus.Errorf("Error: %d %s", err.Code, err.Text)
	} else {
		logrus.Infof("Successful Response: %d", httpStatus)
	}

	if c.Request.Method == http.MethodHead {
		c.Status(httpStatus)
		return
	}
	c.JSON(httpStatus, resp)
}
