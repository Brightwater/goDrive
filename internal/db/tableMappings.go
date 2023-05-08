package db

import "time"

type UploadedFile struct {
	FileName        string    `db:"file_name"`
	UploaderAddress string    `db:"uploader_address"`
	ExpirationTime  time.Time `db:"expirationtime"`
	Code            string    `db:"permission_code"`
}

type PermissionCode struct {
	Code           string    `db:"code"`
	GrantedBy      string    `db:"granted_by"`
	ExpirationTime time.Time `db:"expiration_time"`
	UsesRemaining  int64     `db:"uses_remaining"`
}

type LocalFile struct {
	FileName        string    `db:"file_name"`
    User            string    `db:"user"`
    Code            string    `db:"permission_code"`
	ExpirationTime  time.Time `db:"expirationtime"`
    FilePath        string    `db:"file_path"`
}
