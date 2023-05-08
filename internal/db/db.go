package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Brightwater/goDrive/internal/service"
	"github.com/jackc/pgx/v4/pgxpool"
)

var Pool *pgxpool.Pool

func InitPgPool() {

	config := service.AppConfig

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.DBUser, config.Password, config.DBName)

	poolConfig, err := pgxpool.ParseConfig(psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	poolConfig.MaxConns = 10
	poolConfig.MaxConnIdleTime = 15 * time.Minute
	poolConfig.MaxConnLifetime = 30 * time.Minute

	Pool, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Database connected")
}

func CloseDb() {
	if Pool != nil {
		Pool.Close()
	}
}

func PersistPermissionCode(code string, hours int64, uses int64) error {
	exp := time.Now().Add(time.Duration(hours) * time.Hour)
	permissionCode := PermissionCode{Code: code, GrantedBy: "Jeremiah", ExpirationTime: exp, UsesRemaining: uses}

	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	queryStr := `INSERT INTO permission_codes
				 (code, granted_by, expiration_time, uses_remaining)
				 VALUES($1, $2, $3, $4);`

	result, err := conn.Query(context.Background(), queryStr, permissionCode.Code,
		permissionCode.GrantedBy, permissionCode.ExpirationTime, permissionCode.UsesRemaining)
	// result, err := conn.Exec(context.Background(), queryStr, permissionCode.Code, permissionCode.GrantedBy, permissionCode.ExpirationTime)
	if err != nil {
		log.Println("Error persisting")
		return err
	}
	result.Close()

	// Check if the insert was successful
	if result.CommandTag().RowsAffected() == 0 {
		return fmt.Errorf("insert failed")
	}
	log.Printf("persisted in db %d code", result.CommandTag().RowsAffected())

	return nil
}

func UpdatePermissionCodeUsesCount(code string) error {
	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	queryStr := `update permission_codes set uses_remaining = uses_remaining - 1 where code = $1`

	result, err := conn.Query(context.Background(), queryStr, code)
	if err != nil {
		log.Println("Error persisting")
		return err
	}
	result.Close()

	// Check if the insert was successful
	if result.CommandTag().RowsAffected() == 0 {
		return fmt.Errorf("update failed")
	}
	log.Printf("persisted in db %d code update", result.CommandTag().RowsAffected())

	return nil
}

func PersistLocalFileDl(filePath string, fileName string, permissionCode string, ttlHours int64) error {
	exp := time.Now().Add(time.Duration(ttlHours) * time.Hour)

	fileUp := LocalFile{FileName: fileName, User: "Jeremiah",
		Code: permissionCode, ExpirationTime: exp, FilePath: filePath}

	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	queryStr := `INSERT INTO public.local_files
				 (file_name, "user", permission_code, expiration_time, file_path)
				 VALUES($1, $2, $3, $4, $5)`

	result, err := conn.Query(context.Background(), queryStr, fileUp.FileName,
		fileUp.User, fileUp.Code, fileUp.ExpirationTime, fileUp.FilePath)
	if err != nil {
		log.Println("Error persisting")
		return err
	}
	result.Close()

	// Check if the insert was successful
	if result.CommandTag().RowsAffected() == 0 {
		return fmt.Errorf("insert failed")
	}
	log.Printf("persisted in db %d local file", result.CommandTag().RowsAffected())

	return nil
}



func GetLocalFileDownload(fileName string, code string) (*LocalFile, error) {
	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer conn.Release()

	queryStr := `select file_name, user, permission_code, expiration_time, file_path
				 from local_files
				 where permission_code = $1 and file_name = $2`

	var file *LocalFile = new(LocalFile)
	err = conn.QueryRow(context.Background(), queryStr, code, fileName).Scan(&file.FileName,
		&file.User, &file.Code, &file.ExpirationTime, &file.FilePath)
	if err != nil {
		return file, err
	}

	if file.Code == code {
		return file, nil
	}

	// can also do temp file dl check here

	return file, fmt.Errorf("code not found")
}

func GetUploadedFile(fileName string, code string) (*UploadedFile, error) {
	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer conn.Release()

	queryStr := `select file_name, permission_code, expirationtime
				 from uploaded_files
				 where permission_code = $1 and file_name = $2`

	var file *UploadedFile = new(UploadedFile)
	err = conn.QueryRow(context.Background(), queryStr, code, fileName).Scan(&file.FileName,
		&file.Code, &file.ExpirationTime)
	if err != nil {
		return file, err
	}

	if file.Code == code {
		return file, nil
	}

	// can also do temp file dl check here

	return file, fmt.Errorf("code not found")
}

func GetPermissionCode(code string) (*PermissionCode, error) {

	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer conn.Release()

	queryStr := `select code, granted_by, expiration_time, uses_remaining
				 from permission_codes pc
				 where code = $1`

	var retCode *PermissionCode = new(PermissionCode)
	err = conn.QueryRow(context.Background(), queryStr, code).Scan(&retCode.Code, &retCode.GrantedBy, &retCode.ExpirationTime, &retCode.UsesRemaining)
	if err != nil {
		return retCode, err
	}

	if retCode.Code == code {
		return retCode, nil
	}

	return retCode, fmt.Errorf("code not found")
}

// clears expired codes
func CleanPermissionCodes() {
	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer conn.Release()

	queryStr := `DELETE FROM permission_codes WHERE expiration_time < NOW()`

	result, err := conn.Query(context.Background(), queryStr)
	if err != nil {
		log.Println(err)
	}
	result.Close()

	log.Printf("Cleaned %d permission codes", result.CommandTag().RowsAffected())
}

func PersistUploadFile(fileName string, permissionCode string, address string) error {
	exp := time.Now().Add(24 * time.Hour)

	fileUp := UploadedFile{FileName: fileName, UploaderAddress: address, ExpirationTime: exp, Code: permissionCode}

	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	queryStr := `INSERT INTO uploaded_files
				 (file_name, uploader_address, expirationtime, permission_code)
				 VALUES($1, $2, $3, $4)`

	result, err := conn.Query(context.Background(), queryStr, fileUp.FileName,
		fileUp.UploaderAddress, fileUp.ExpirationTime, fileUp.Code)
	if err != nil {
		log.Println("Error persisting")
		return err
	}
	result.Close()

	// Check if the insert was successful
	if result.CommandTag().RowsAffected() == 0 {
		return fmt.Errorf("insert failed")
	}
	log.Printf("persisted in db %d upload file", result.CommandTag().RowsAffected())

	return nil
}

func CleanLocalFileDbEntries() {
	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer conn.Release()

	queryStr := `DELETE FROM local_files WHERE expiration_time < NOW()`

	result, err := conn.Query(context.Background(), queryStr)
	if err != nil {
		log.Println(err)
	}
	result.Close()

	log.Printf("Cleaned %d local file entries", result.CommandTag().RowsAffected())
}

func GetListOfUploadFiles() ([]string, error) {
	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer conn.Release()

	queryStr := `select file_name
				 from uploaded_files pc`

	names := []string{}
	rows, err := conn.Query(context.Background(), queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}

func CleanFileUploadsEntries() {
	conn, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer conn.Release()

	queryStr := `DELETE FROM uploaded_files WHERE expirationtime < NOW()`

	result, err := conn.Query(context.Background(), queryStr)
	if err != nil {
		log.Println(err)
	}
	result.Close()

	log.Printf("Cleaned %d db upload file entries", result.CommandTag().RowsAffected())
}
