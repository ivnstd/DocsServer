package handler

var (
	ErrBadRequest = &ErrorResponse{
		Code: 400,
		Text: "Invalid request body",
	}

	ErrUnauthorized = &ErrorResponse{
		Code: 401,
		Text: "Unauthorized",
	}

	ErrForbidden = &ErrorResponse{
		Code: 403,
		Text: "Forbidden",
	}

	ErrNotFound = &ErrorResponse{
		Code: 404,
		Text: "Not Found",
	}

	ErrMethodNotAllowed = &ErrorResponse{
		Code: 405,
		Text: "Method Not Allowed",
	}

	ErrInternalServer = &ErrorResponse{
		Code: 500,
		Text: "Internal Server Error",
	}

	ErrNotImplemented = &ErrorResponse{
		Code: 501,
		Text: "Not Implemented",
	}
)
