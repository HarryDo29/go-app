package dto

type GeneratePresignedURLReq struct {
	ObjectName  string `json:"object_name" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	Folder      string `json:"folder"`
}

type GeneratePresignedURLRes struct {
	URL string `json:"url"`
}
