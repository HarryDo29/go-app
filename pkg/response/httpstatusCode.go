package response

const (
	// --- 1. NHÓM THÀNH CÔNG (SUCCESS - 200xx) ---
	ErrCodeSuccess = 20000 // Thành công / Success

	// --- 2. NHÓM LỖI DỮ LIỆU ĐẦU VÀO (BAD REQUEST - 400xx) ---
	ErrCodeParamInvalid  = 40001 // Tham số URL hoặc Query không hợp lệ / thiếu
	ErrCodeHeaderInvalid = 40002 // Header yêu cầu không hợp lệ / thiếu
	ErrCodeBodyInvalid   = 40003 // Dữ liệu JSON gửi lên không hợp lệ / thiếu

	// --- 3. NHÓM LỖI XÁC THỰC (UNAUTHORIZED - 401xx) ---
	ErrCodeAuthFailed    = 40101 // Đăng nhập hoặc xác thực tài khoản thất bại
	ErrCodeTokenInvalid  = 40102 // Token bảo mật không hợp lệ (sai signature)
	ErrCodeTokenExpired  = 40103 // Token đã hết hạn sử dụng
	ErrCodeTokenNotFound = 40104 // Không tìm thấy Token trong header Authorization

	// --- 4. NHÓM LỖI QUYỀN TRUY CẬP (FORBIDDEN - 403xx) ---
	ErrCodePermissionDenied = 40301 // Người dùng không có quyền truy cập tài nguyên này

	// --- 5. NHÓM LỖI KHÔNG TÌM THẤY (NOT FOUND - 404xx) ---
	ErrCodeNotFound     = 40401 // Không tìm thấy tài nguyên yêu cầu
	ErrCodeUserNotFound = 40402 // Không tìm thấy người dùng trên hệ thống

	// --- 6. NHÓM LỖI XUNG ĐỘT DỮ LIỆU (CONFLICT - 409xx) ---
	ErrCodeUserExist = 40901 // Người dùng đã tồn tại (Trùng email/username)

	// --- 7. GIỚI HẠN TẦN SUẤT (RATE LIMIT - 429xx) ---
	ErrCodeTooManyRequests = 42901 // Gửi quá nhiều yêu cầu trong thời gian ngắn

	// --- 8. NHÓM LỖI HỆ THỐNG / DATABASE (INTERNAL SERVER ERROR - 500xx) ---
	ErrCodeServer       = 50000 // Lỗi hệ thống nội bộ máy chủ
	ErrCodeCreateFailed = 50001 // Tạo cơ sở dữ liệu thất bại
	ErrCodeGetFailed    = 50002 // Query cơ sở dữ liệu thất bại
	ErrCodeUpdateFailed = 50003 // Cập nhật cơ sở dữ liệu thất bại
	ErrCodeDeleteFailed = 50004 // Xóa cơ sở dữ liệu thất bại
	ErrCodeCache        = 50005
)

// Bản đồ thông báo lỗi chi tiết phục vụ hiển thị Client
var msg = map[int]string{
	// Success
	ErrCodeSuccess: "Success",

	// Request Validation
	ErrCodeParamInvalid:  "Parameter is missing or invalid",
	ErrCodeHeaderInvalid: "Header is missing or invalid",
	ErrCodeBodyInvalid:   "Body is missing or invalid",

	// Authentication & Authorization
	ErrCodeAuthFailed:       "Authentication failed",
	ErrCodeTokenInvalid:     "Token is invalid",
	ErrCodeTokenExpired:     "Token has expired",
	ErrCodeTokenNotFound:    "Authorization token not found",
	ErrCodePermissionDenied: "Permission denied",

	// Not Found
	ErrCodeNotFound:     "Resource not found",
	ErrCodeUserNotFound: "User not found",

	// Conflicts
	ErrCodeUserExist: "User has already existed",

	// Rate Limiting
	ErrCodeTooManyRequests: "Too many requests",

	// Server Errors
	ErrCodeServer:       "Internal server error",
	ErrCodeCreateFailed: "Create failed",
	ErrCodeGetFailed:    "Get failed",
	ErrCodeUpdateFailed: "Update failed",
	ErrCodeDeleteFailed: "Delete failed",
	ErrCodeCache:        "Cache not existed",
}
