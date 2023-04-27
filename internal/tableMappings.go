package internal

import "time"

type UploadedFile struct {
    ID               int32       `db:"id"`
    FileName         string      `db:"file_name"`
    UploaderAddress  string      `db:"uploader_address"`
    ExpirationTime   time.Time   `db:"expirationtime"`
}

type PermissionCode struct {
    Code            string      `db:"code"`
    GrantedBy       string      `db:"granted_by"`
    ExpirationTime  time.Time   `db:"expiration_time"`
}