package util

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliyunOssStruct struct {
	Info struct {
		Endpoint        string
		AccessKeyId     string
		AccessKeySecret string
		BucketName      string
		BasePath        string
	}
	Client *oss.Client
	Bucket *oss.Bucket
}

const AliyunOssFileRandStringNum = 6
const AliyunOssFileTimeFormat = "20060102150405"

const AliyunOssAvatar = "AVATAR"
const AliyunOssCoverImage = "COVER_IMAGE"
const AliyunOssRoomImage = "ROOM_IMAGE"
const AliyunOssGallery = "GALLERY"
const AliyunOssKyc = "KYC"

var AliyunOssFolder = map[string]string{
	AliyunOssAvatar:     "avatar",
	AliyunOssCoverImage: "cover_image",
	AliyunOssRoomImage:  "room_image",
	AliyunOssGallery:    "gallery",
	AliyunOssKyc:        "kyc",
}

func (a *AliyunOssStruct) getFileName(aliyunOssFolder string, userId int64, extension string) (fileName string, err error) {
	if AliyunOssFolder[aliyunOssFolder] == "" {
		err = errors.New("folder does not exist")
		return
	}
	fileName = fmt.Sprintf(
		"%d-%s-%s-%s%s",
		userId,
		AliyunOssFolder[aliyunOssFolder],
		time.Now().Format(AliyunOssFileTimeFormat),
		RandStringRunes(AliyunOssFileRandStringNum),
		extension,
	)
	return
}

func (a *AliyunOssStruct) getFolderPath(aliyunOssFolder string, userId int64) (path string, err error) {
	if AliyunOssFolder[aliyunOssFolder] == "" {
		err = errors.New("folder does not exist")
		return
	}
	path = fmt.Sprintf("%s/user/%d/%s", a.Info.BasePath, userId, AliyunOssFolder[aliyunOssFolder])
	return
}

func (a *AliyunOssStruct) UploadFile(aliyunOssFolder string, userId int64, file *multipart.FileHeader, extension string) (path string, err error) {
	var folderPath, filePath string
	folderPath, err = a.getFolderPath(aliyunOssFolder, userId)
	if err != nil {
		err = errors.New("oss path error, " + err.Error())
		return
	}
	filePath, err = a.getFileName(aliyunOssFolder, userId, extension)
	if err != nil {
		err = errors.New("oss path error, " + err.Error())
		return
	}

	f, err := file.Open()
	if err != nil {
		err = errors.New("file open failed, " + err.Error())
		return
	}
	defer f.Close()
	path = folderPath + "/" + filePath

	if err = a.Bucket.PutObject(path, f); err != nil {
		err = errors.New("upload file failed, " + err.Error())
		return
	}

	return
}

func (a *AliyunOssStruct) DeleteFile(key string) error {
	if err := a.Bucket.DeleteObject(key); err != nil {
		return errors.New("AliyunOss Delete " + key + " Failed, err: " + err.Error())
	}

	return nil
}

func InitAliyunOSS() (aliyunOss AliyunOssStruct, err error) {
	aliyunOss.Info.Endpoint = os.Getenv("ALIYUN_OSS_ENDPOINT")
	aliyunOss.Info.AccessKeyId = os.Getenv("ALIYUN_OSS_ACCESS_KEY_ID")
	aliyunOss.Info.AccessKeySecret = os.Getenv("ALIYUN_OSS_ACCESS_KEY_SECRET")
	aliyunOss.Info.BucketName = os.Getenv("ALIYUN_OSS_BUCKET_NAME")
	aliyunOss.Info.BasePath = os.Getenv("ALIYUN_OSS_BASE_PATH")

	if aliyunOss.Info.BasePath == "" {
		fmt.Println("ALIYUN_OSS_BASE_PATH is empty")
	}

	// 创建OSSClient实例。
	client, err := oss.New(aliyunOss.Info.Endpoint, aliyunOss.Info.AccessKeyId, aliyunOss.Info.AccessKeySecret)
	if err != nil {
		return
	}
	aliyunOss.Client = client

	// 获取存储空间。
	bucket, err := client.Bucket(aliyunOss.Info.BucketName)
	if err != nil {
		return
	}
	aliyunOss.Bucket = bucket
	return
}

func BuildAliyunOSSUrl(path string) string {
	bucketUrl := os.Getenv("ALIYUN_OSS_BUCKET_URL")
	return bucketUrl + path
}
