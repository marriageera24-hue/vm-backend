package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
)

type Media struct {
	Model
	UserUUID string `gorm:"type:uuid;index;" json:"user_uuid"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
}

func upload(ctx echo.Context) error {
	u, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, err.Error())
	}

	filetype := ctx.Param("type")
	UUID := ctx.Param("uuid")

	var (
		m  Media
		m2 Media
	)
	m.UserUUID = u.UUID
	m.Type = filetype

	// Parsing UUID from string input
	u2, err := uuid.FromString(UUID)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		// return ctx.JSON(http.StatusBadRequest, err.Error())
		u2 = uuid.NewV4() // generate new valid uuid
	}

	m.UUID = u2.String()

	db.Model(&Media{}).Where("uuid = ?", m.UUID).First(&m2)

	if m2.ID > 0 {
		m.ID = m2.ID
	}

	// Multipart form
	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["pics"]
	for _, file := range files {
		// Create a single AWS session (we can re use this if we're uploading many files)
		/*s,*/_, err := session.NewSession(&aws.Config{Region: aws.String(os.Getenv("AWS_REGION"))})
		if err != nil {
			log.Fatal(err)
		}

		//upload to public/pictures
		src, err := file.Open()
		if err != nil {
			return err
		}
		
	
		// Destination
		dst, err := os.Create("public/pictures/"+file.Filename)
		if err != nil {
			return err
		}
	// Copy
		if _, err = io.Copy(dst, src); err != nil {
			return err
		}

		defer dst.Close()
		defer src.Close()
	
		// Upload to s3
		// m.Path, err = AddFileToS3(s, m.UUID, file)
		// if err != nil {
		// 	log.Fatal(err)
		// 	return ctx.JSON(http.StatusBadRequest, err.Error())
		// }

		// m.Name = file.Filename
		// m.Path = "https://vm-pictures.s3.ap-south-1.amazonaws.com/pictures/" + m.Path

		m.Path = "https://marriage-era.com/pictures/"+file.Filename

		db.Save(&m)

	}
	// db.Where("from_user_uuid = ?", u.UUID).Find(&uu)

	return ctx.JSON(http.StatusCreated, m)
}

func download(ctx echo.Context) error {

	_, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, err.Error())
	}

	var m Media
	UUID := ctx.Param("uuid")
	log.Println(UUID)

	err = db.Model(&Media{}).Where("uuid = ?", UUID).First(&m).Error

	if err != nil {
		log.Println(err.Error())
		return ctx.NoContent(http.StatusNotFound)
	}

	resp, err := http.Get(m.Path)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Status error: %v", resp.StatusCode)
	}
	data, _ := ioutil.ReadAll(resp.Body)
	return ctx.Blob(http.StatusOK, mime.TypeByExtension(filepath.Ext(m.Path)), data)

}

// AddFileToS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func AddFileToS3(s *session.Session, uuid string, file *multipart.FileHeader) (string, error) {

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// // Open the file for use
	// file, err := os.Open(fileDir)
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()

	// Get file size and read the file content into a buffer
	var size int64 = file.Size
	buffer := make([]byte, size)
	src.Read(buffer)

	// create a unique file name for the file
	tempFileName := "pictures/" + uuid + filepath.Ext(file.Filename)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(os.Getenv("AWS_S3_BUCKET")),
		Key:                  aws.String(tempFileName),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	if err != nil {
		log.Println(err.Error)
		return "", err
	}

	return tempFileName, err
}
