package long_conn

import (
	"net/http"
	"time"
)

type LongConn interface {
	//Close this connection
	Close() error
	// WriteMessage Write message to connection,messageType means data type,can be set binary(2) and text(1).
	WriteMessage(messageType int, message []byte) error
	// ReadMessage Read message from connection.
	ReadMessage() (int, []byte, error)
	// SetReadDeadline sets the read deadline on the underlying network connection,
	//after a read has timed out, will return an error.
	SetReadDeadline(timeout time.Duration) error
	// SetWriteDeadline sets to write deadline when send message,when read has timed out,will return error.
	SetWriteDeadline(timeout time.Duration) error
	// Dial Try to dial a connection,url must set auth args,header can control compress data
	Dial(urlStr string, requestHeader http.Header) (*http.Response, error)
	// IsNil Whether the connection of the current long connection is nil
	IsNil() bool
	// SetReadLimit sets the maximum size for a message read from the peer.bytes
	SetReadLimit(limit int64)
	LocalAddr() string
	CheckHeartbeat(interval time.Duration)
}
