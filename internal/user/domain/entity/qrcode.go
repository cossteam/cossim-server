package entity

// QRCode represents a QR code with a token, status, and user ID.
type QRCode struct {
	Token  string       `json:"token"`
	Status QRCodeStatus `json:"status"`
	UserID string       `json:"user_id"`
}

// QRCodeStatus represents the status of a QR code.
type QRCodeStatus uint8

const (
	// QRCodeStatusNotScanned indicates that the QR code has not been scanned.
	QRCodeStatusNotScanned QRCodeStatus = iota

	// QRCodeStatusScanned indicates that the QR code has been scanned.
	QRCodeStatusScanned

	// QRCodeStatusConfirmed indicates that the QR code has been confirmed.
	QRCodeStatusConfirmed

	// QRCodeStatusExpired indicates that the QR code has expired.
	QRCodeStatusExpired
)

// statusDescriptions provides a human-readable description for each QRCodeStatus.
var statusDescriptions = map[QRCodeStatus]string{
	QRCodeStatusNotScanned: "未扫描",
	QRCodeStatusScanned:    "已扫描",
	QRCodeStatusConfirmed:  "已确认",
	QRCodeStatusExpired:    "已过期",
}

// String returns the description of the QRCodeStatus.
func (status QRCodeStatus) String() string {
	return statusDescriptions[status]
}
