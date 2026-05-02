package request

// LinkGroupAddRequest is the request body for creating a link group.
type LinkGroupAddRequest struct {
	Title string `json:"title"`
}

// LinkGroupUpdateRequest is the request body for updating a link group.
type LinkGroupUpdateRequest struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}
