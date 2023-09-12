package serializer

import (
	"web-api/model"
	"web-api/util"
)

type Kyc struct {
	Id               int64    `json:"id"`
	FirstName        string   `json:"first_name"`
	MiddleName       string   `json:"middle_name"`
	LastName         string   `json:"last_name"`
	Birthday         string   `json:"birthday"`
	DocumentType     int      `json:"document_type"`
	DocumentUrls     []string `json:"document_urls"`
	Nationality      int      `json:"nationality"`
	CurrentAddress   string   `json:"current_address"`
	PermanentAddress string   `json:"permanent_address"`
	EmploymentType   int      `json:"employment_type"`
	IncomeSource     int      `json:"income_source"`
	AssociatedStore  int      `json:"associated_store"`
	Status           int      `json:"status"`
	Remark           string   `json:"remark"`
}

func BuildKyc(dbKyc model.Kyc, dbKycDocs []model.KycDocument) Kyc {
	var docUrls []string
	for _, d := range dbKycDocs {
		docUrls = append(docUrls, util.BuildAliyunOSSUrl(d.Url))
	}

	return Kyc{
		Id:               dbKyc.ID,
		FirstName:        dbKyc.FirstName,
		MiddleName:       dbKyc.MiddleName,
		LastName:         dbKyc.LastName,
		Birthday:         dbKyc.Birthday,
		DocumentType:     dbKyc.DocumentType,
		DocumentUrls:     docUrls,
		Nationality:      dbKyc.Nationality,
		CurrentAddress:   dbKyc.CurrentAddress,
		PermanentAddress: dbKyc.PermanentAddress,
		EmploymentType:   dbKyc.EmploymentType,
		IncomeSource:     dbKyc.IncomeSource,
		AssociatedStore:  dbKyc.AssociatedStore,
		Status:           dbKyc.Status,
		Remark:           dbKyc.Remark,
	}
}
