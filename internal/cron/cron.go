package cron

import (
	"github.com/Brightwater/goDrive/internal/db"
	"github.com/Brightwater/goDrive/internal/file"
	"github.com/robfig/cron"
)

var cronJob *cron.Cron

func StartCronService() {
	cronJob = cron.New()

	// cronJob.AddFunc("0 * * * * *", testCron)
	cronJob.AddFunc("0 0 * * * *", db.CleanPermissionCodes)
	cronJob.AddFunc("0 0 * * * *", db.CleanLocalFileDbEntries)
	cronJob.AddFunc("0 0 * * * *", db.CleanFileUploadsEntries)
	cronJob.AddFunc("0 0 23 * * *", file.CleanUploadedFiles)
	


	cronJob.Start()
}

func StopCronService() {
	cronJob.Stop()
}

// func testCron() {
// 	fmt.Println("TEST")
// }
