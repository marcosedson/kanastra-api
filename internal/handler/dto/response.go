package dto

type ProcessFilesResponse struct {
	Message string `json:"message"`
}
type ProcessStatus struct {
	FileName        string `json:"file_name"`
	TotalLines      int    `json:"total_lines"`
	ProcessedLines  int    `json:"processed_lines"`
	Status          string `json:"status"`
	LastUpdatedTime string `json:"last_updated_time"`
}
