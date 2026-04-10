package helpers

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	MaxFileSize = 5 * 1024 * 1024 // 5 MB
)

func MaskPassword(pass string) string {
    if len(pass) <= 4 {
        return strings.Repeat("*", len(pass))
    }
    return pass[:2] + strings.Repeat("*", len(pass)-4) + pass[len(pass)-2:]
}

func MaskCardNumber(number string) string {
    if len(number) <= 4 {
        return "****"
    }
    return "**** **** **** " + number[len(number)-4:]
}

// isValidCardNumber проверяет номер карты по алгоритму Луна
func IsValidCardNumber(number string) bool {
    if len(number) < 13 || len(number) > 19 {
        return false
    }
    
    // Проверка что только цифры
    for _, r := range number {
        if r < '0' || r > '9' {
            return false
        }
    }
    
    // Алгоритм Луна
    var sum int
    alternate := false
    for i := len(number) - 1; i >= 0; i-- {
        digit := int(number[i] - '0')
        if alternate {
            digit *= 2
            if digit > 9 {
                digit -= 9
            }
        }
        sum += digit
        alternate = !alternate
    }
    
    return sum%10 == 0
}

// isValidCVV проверяет CVV код
func IsValidCVV(cvv string) bool {
    if len(cvv) != 3 && len(cvv) != 4 {
        return false
    }
    for _, r := range cvv {
        if r < '0' || r > '9' {
            return false
        }
    }
    return true
}

func FormatFileSize(size int64) string {
    const (
        KB = 1024
        MB = KB * 1024
        GB = MB * 1024
    )
    
    switch {
    case size >= GB:
        return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
    case size >= MB:
        return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
    case size >= KB:
        return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
    default:
        return fmt.Sprintf("%d B", size)
    }
}

func DetectMimeType(filePath string) string {
    ext := strings.ToLower(filepath.Ext(filePath))
    
    mimeTypes := map[string]string{
        ".txt":  "text/plain",
        ".md":   "text/markdown",
        ".json": "application/json",
        ".yaml": "application/yaml",
        ".yml":  "application/yaml",
        ".xml":  "application/xml",
        ".html": "text/html",
        ".htm":  "text/html",
        ".css":  "text/css",
        ".js":   "application/javascript",
        ".pdf":  "application/pdf",
        ".jpg":  "image/jpeg",
        ".jpeg": "image/jpeg",
        ".png":  "image/png",
        ".gif":  "image/gif",
        ".webp": "image/webp",
        ".svg":  "image/svg+xml",
        ".zip":  "application/zip",
        ".tar":  "application/x-tar",
        ".gz":   "application/gzip",
        ".7z":   "application/x-7z-compressed",
        ".mp3":  "audio/mpeg",
        ".mp4":  "video/mp4",
        ".mov":  "video/quicktime",
        ".avi":  "video/x-msvideo",
    }
    
    if mime, ok := mimeTypes[ext]; ok {
        return mime
    }
    
    return "application/octet-stream"
}