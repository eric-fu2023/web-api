package service

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"os"
	"strconv"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

var (
	kycFileContentTypeToExtension = map[string]string{
		"image/jpeg": ".jpg",
		"image/jpg":  ".jpg",
		"image/png":  ".png",
	}

	errUpdateCompletedKyc = errors.New("cannot update completed kyc")
)

type SubmitKycService struct {
	FirstName        string                 `form:"first_name" json:"first_name" binding:"required"`
	MiddleName       string                 `form:"middle_name" json:"middle_name"`
	LastName         string                 `form:"last_name" json:"last_name" binding:"required"`
	Birthday         string                 `form:"birthday" json:"birthday" binding:"required"`
	DocumentType     int                    `form:"document_type" json:"document_type" binding:"required"`
	Documents        []multipart.FileHeader `form:"documents" json:"documents" binding:"gt=0"`
	Nationality      int                    `form:"nationality" json:"nationality" binding:"required"`
	CurrentAddress   string                 `form:"current_address" json:"current_address" binding:"required"`
	PermanentAddress string                 `form:"permanent_address" json:"permanent_address" binding:"required"`
	EmploymentType   int                    `form:"employment_type" json:"employment_type" binding:"required"`
	IncomeSource     int                    `form:"income_source" json:"income_source" binding:"required"`
	AssociatedStore  int                    `form:"associated_store" json:"associated_store" binding:"required"`
}

func (service *SubmitKycService) SubmitKyc(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)

	// Validate birthday format
	_, err := time.Parse(time.DateOnly, service.Birthday)
	if err != nil {
		return serializer.ParamErr(c, service, fmt.Sprintf(i18n.T("kyc_invalid_birthday"), service.Birthday), err)
	}

	for _, d := range service.Documents {
		mt := d.Header.Get("Content-Type")
		if _, exists := kycFileContentTypeToExtension[mt]; !exists {
			return serializer.ParamErr(c, service, fmt.Sprintf(i18n.T("file_type_not_allowed"), mt), err)
		}
		if d.Size > 2*1024*1024 {
			return serializer.ParamErr(c, service, fmt.Sprintf(i18n.T("file_size_exceeded"), "2MB"), err)
		}
	}

	kyc, err := service.createOrUpdateKyc(c)
	if err != nil && errors.Is(err, errUpdateCompletedKyc) {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("kyc_cannot_update_completed_kyc"), err)
	} else if err != nil {
		util.GetLoggerEntry(c).Errorf("Create or update kyc error: %s", err.Error())
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("kyc_submit_failed"), err)
	}

	go service.verifyDocuments(c, kyc.ID)

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}

func (service *SubmitKycService) createOrUpdateKyc(c *gin.Context) (kyc model.Kyc, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)

	tx := model.DB.Begin()

	curKyc, err := model.GetKycWithLock(tx, user.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		util.GetLoggerEntry(c).Errorf("Get kyc with lock error: %s", err.Error())
		return
	}
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		kyc, err = service.createKycInfo(c, tx)
	} else {
		kyc, err = service.updateKycInfo(c, tx, curKyc)
	}

	if err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()
	return
}

func (service *SubmitKycService) createKycInfo(c *gin.Context, tx *gorm.DB) (kyc model.Kyc, err error) {
	// Create KYC
	kyc = service.buildKyc(c)
	err = tx.Create(&kyc).Error
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Create kyc error: %s", err.Error())
		return
	}

	// Create Kyc Documents
	for _, d := range service.Documents {
		err = service.uploadKycDoc(c, tx, kyc.ID, d)
		if err != nil {
			util.GetLoggerEntry(c).Errorf("Upload kyc doc error: %s", err.Error())
			return
		}
	}

	return
}

func (service *SubmitKycService) updateKycInfo(c *gin.Context, tx *gorm.DB, curKyc model.Kyc) (kyc model.Kyc, err error) {
	if curKyc.Status == consts.KycStatusCompleted {
		err = errUpdateCompletedKyc
		return
	}

	kyc = service.buildKycForUpdate(c, curKyc.ID)
	err = model.UpdateKyc(tx, kyc)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Update kyc error: %s", err.Error())
		return
	}

	err = service.updateKycDocuments(c, tx, curKyc.ID)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Update kyc documents error: %s", err.Error())
		return
	}

	return
}

func (service *SubmitKycService) updateKycDocuments(c *gin.Context, tx *gorm.DB, kycId int64) error {
	// Get current docs from db
	curDocs, err := model.GetKycDocumentsWithLock(tx, kycId)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Get cur kyc docs error: %s", err.Error())
		return err
	}

	// Delete all current docs from db
	err = model.DeleteKycDocuments(tx, kycId)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Delete kyc docs error: %s", err.Error())
		return err
	}

	// Upload all docs given in request to oss and insert to db
	for _, d := range service.Documents {
		err = service.uploadKycDoc(c, tx, kycId, d)
		if err != nil {
			util.GetLoggerEntry(c).Errorf("Upload kyc doc error: %s", err.Error())
			return err
		}
	}

	// Delete old docs from oss
	oss, err := util.InitAliyunOSS()
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Init oss error: %s", err.Error())
		return err
	}
	for _, d := range curDocs {
		err = oss.DeleteFile(d.Url)
		if err != nil {
			util.GetLoggerEntry(c).Errorf("OSS delete file error: %s", err.Error())
			return err
		}
	}

	return nil
}

func (service *SubmitKycService) uploadKycDoc(
	c *gin.Context,
	tx *gorm.DB,
	kycId int64,
	document multipart.FileHeader,
) error {
	u, _ := c.Get("user")
	user := u.(model.User)

	oss, err := util.InitAliyunOSS()
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Init oss error: %s", err.Error())
		return err
	}
	path, err := oss.UploadFile(
		util.AliyunOssKyc,
		user.ID,
		&document,
		kycFileContentTypeToExtension[document.Header.Get("Content-Type")])
	if err != nil {
		util.GetLoggerEntry(c).Errorf("OSS upload file error: %s", err.Error())
		return err
	}

	kycDocument := service.buildKycDocument(kycId, path)
	err = tx.Create(&kycDocument).Error
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Create kyc document error: %s", err.Error())

		err = oss.DeleteFile(path)
		if err != nil {
			util.GetLoggerEntry(c).Errorf("OSS delete file error: %s", err.Error())
		}
		return err
	}

	return nil
}

func (service *SubmitKycService) buildKycForUpdate(c *gin.Context, kycId int64) model.Kyc {
	kyc := service.buildKyc(c)
	kyc.ID = kycId
	return kyc
}

func (service *SubmitKycService) buildKyc(c *gin.Context) model.Kyc {
	u, _ := c.Get("user")
	user := u.(model.User)

	return model.Kyc{
		KycC: models.KycC{
			UserId:           user.ID,
			FirstName:        service.FirstName,
			MiddleName:       service.MiddleName,
			LastName:         service.LastName,
			Birthday:         service.Birthday,
			DocumentType:     service.DocumentType,
			Nationality:      service.Nationality,
			CurrentAddress:   service.CurrentAddress,
			PermanentAddress: service.PermanentAddress,
			EmploymentType:   service.EmploymentType,
			IncomeSource:     service.IncomeSource,
			AssociatedStore:  service.AssociatedStore,
			Status:           consts.KycStatusPending,
		},
	}
}

func (service *SubmitKycService) buildKycDocument(kycId int64, url string) model.KycDocument {
	return model.KycDocument{KycDocumentC: models.KycDocumentC{
		KycId: kycId,
		Url:   url,
	}}
}

func (service *SubmitKycService) verifyDocuments(c *gin.Context, kycId int64) {
	var images [][]byte
	shufti := util.Shufti{
		ClientId:  os.Getenv("SHUFTI_CLIENT_ID"),
		SecretKey: os.Getenv("SHUFTI_SECRET_KEY"),
	}
	for _, d := range service.Documents {
		if f, e := d.Open(); e == nil {
			defer f.Close()
			if b, e := io.ReadAll(f); e == nil {
				images = append(images, b)
			}
		}
	}
	isAccepted, reason, err := shufti.VerifyDocument(kycId, service.FirstName, service.MiddleName, service.LastName, service.Birthday, strconv.Itoa(service.Nationality), images[0], images[1])
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Shufti document verification error: %s", err.Error())
		return
	}
	if isAccepted {
		err = model.AcceptKyc(kycId)
		if err != nil {
			util.GetLoggerEntry(c).Errorf("Update kyc status error: %s", err.Error())
			return
		}
	} else {
		err = model.RejectKycWithReason(kycId, reason)
		if err != nil {
			util.GetLoggerEntry(c).Errorf("Update kyc status error: %s", err.Error())
			return
		}
	}
}
