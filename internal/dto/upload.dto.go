package dto

type InitiateUploadReq struct {
	ObjectName  string `json:"object_name" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	Folder      string `json:"folder"`
	Size        int64  `json:"size" binding:"required,gt=0"`
}

type InitiateSingleUploadRes struct {
	ObjectKey  string `json:"object_key"`
	PresignURL string `json:"presign_url"`
}

type InitiateUploadPartRes struct {
	UploadId   string `json:"upload_id"`
	ObjectKey  string `json:"object_key"`
	PartSize   int64  `json:"part_size"`   // Chunk size in bytes (e.g., 5242880)
	TotalParts int    `json:"total_parts"` // Total number of chunks
}

type PresignPartReq struct {
	UploadId   string `json:"upload_id" binding:"required"`
	ObjectKey  string `json:"object_key" binding:"required"`
	TotalParts int    `json:"total_parts" binding:"required"`
}

type PartPresignURL struct {
	PartNumber int32  `json:"part_number"`
	PresignURL string `json:"presign_url"`
}

type CompletePartReq struct {
	PartNumber int32  `json:"partNumber"`
	ETag       string `json:"etag"`
}

type CompleteUploadReq struct {
	UploadId  string            `json:"upload_id" binding:"required"`
	ObjectKey string            `json:"object_key" binding:"required"`
	Parts     []CompletePartReq `json:"parts" binding:"required"`
}

type CompleteUploadRes struct {
	ObjectKey   string `json:"object_key"`
	DownloadURL string `json:"download_url"`
}

type AbortUpload struct {
	UploadId  string `json:"upload_id" binding:"required"`
	ObjectKey string `json:"object_key" binding:"required"`
}
