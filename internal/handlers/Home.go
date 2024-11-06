package handlers

import (
	"abir-el-hamd/internal/config"
	"abir-el-hamd/internal/middleware"
	"abir-el-hamd/internal/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func HomeApi(w http.ResponseWriter, r *http.Request) {
	database := config.Database()
	defer database.Close()

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))

	if err != nil {
		er := errors.New("invalid number")
		middleware.ErrorResopnse(w, er)
		return
	}

	from, err := strconv.Atoi(r.URL.Query().Get("from"))

	if err != nil {
		er := errors.New("invalid number")
		middleware.ErrorResopnse(w, er)
		return
	}

	to, err := strconv.Atoi(r.URL.Query().Get("to"))

	if err != nil {
		er := errors.New("invalid number")
		middleware.ErrorResopnse(w, er)
		return
	}

	district := r.URL.Query().Get("district")

	offset := limit - 30

	CasesTable := models.Cases{}

	Cases, err := CasesTable.GetAll(database, 30, offset, from, to, district)

	if err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

	Districts, err := CasesTable.GetAllDistinct(database)

	if err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

	pages, err := CasesTable.NumberOfPages(database, district, from, to)

	if err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	Res := map[string]interface{}{
		"Cases":     Cases,
		"Pages":     pages,
		"Districts": Districts,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(Res); err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}
}

func DeletedCases(w http.ResponseWriter, r *http.Request) {
	database := config.Database()
	defer database.Close()

	CasesTable := models.Cases{}

	Cases, err := CasesTable.DeletedCases(database)

	if err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(Cases); err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}
}

func FilterKids(w http.ResponseWriter, r *http.Request) {
	database := config.Database()
	defer database.Close()

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))

	if err != nil {
		er := errors.New("invalid number")
		middleware.ErrorResopnse(w, er)
		return
	}

	from, err := strconv.Atoi(r.URL.Query().Get("from"))

	if err != nil {
		er := errors.New("invalid number")
		middleware.ErrorResopnse(w, er)
		return
	}

	to, err := strconv.Atoi(r.URL.Query().Get("to"))

	if err != nil {
		er := errors.New("invalid number")
		middleware.ErrorResopnse(w, er)
		return
	}

	district := r.URL.Query().Get("district")

	offset := limit - 30

	FilterdCases := models.FilterdCases{}

	CasesTable := models.Cases{}

	Cases, err := FilterdCases.FilterCasesByRelativeAge(database, district, from, to, limit, offset)

	if err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

	Districts, err := CasesTable.GetAllDistinct(database)

	if err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

	Res := map[string]interface{}{
		"Cases":     Cases,
		"Pages":     len(Cases),
		"Districts": Districts,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(Res); err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}
}

func AddCase(w http.ResponseWriter, r *http.Request) {
	token, err := middleware.GetToken(r)

	if err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	userData, err := middleware.ValidateToken(token)

	if err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	db := config.Database()
	defer db.Close()

	err = r.ParseMultipartForm(30 << 20)

	if err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}
	id := uuid.New().String()

	err = os.MkdirAll("./uploads/"+id, 0755)

	if err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	files := r.MultipartForm.File["files"]

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {

			middleware.ErrorResopnse(w, err)
			return
		}
		defer file.Close()

		dst, err := os.Create("./uploads/" + id + "/" + fileHeader.Filename)
		if err != nil {
			fmt.Println(err)
			middleware.ErrorResopnse(w, err)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			fmt.Println(err)
			middleware.ErrorResopnse(w, err)
			return
		}
	}

	totalIncome, err := strconv.Atoi(r.FormValue("total_income"))
	if err != nil {
		fmt.Println(err)
		log.Println("Error converting total_income:", err)
		return
	}

	fixedExpenses, err := strconv.Atoi(r.FormValue("fixed_expenses"))
	if err != nil {
		log.Println("Error converting fixed_expenses:", err)
		return
	}

	Age, err := strconv.Atoi(r.FormValue("age"))
	if err != nil {
		log.Println("Error converting age:", err)
		return
	}

	now := time.Now()
	createdAt := sql.NullString{String: now.Format(time.RFC3339), Valid: true}

	Case := models.Cases{
		Id:                            id,
		Case_name:                     r.FormValue("case_name"),
		National_id:                   r.FormValue("national_id"),
		Devices_needed_for_the_case:   r.FormValue("devices_needed_for_the_case"),
		Total_income:                  totalIncome,
		Fixed_expenses:                fixedExpenses,
		Pension_from_husband:          r.FormValue("pension_from_husband"),
		Pension_from_father:           r.FormValue("pension_from_father"),
		Debts:                         r.FormValue("debts"),
		Case_type:                     r.FormValue("case_type"),
		Date_of_birth:                 sql.NullString{String: r.FormValue("date_of_birth"), Valid: true},
		Age:                           Age,
		Gender:                        r.FormValue("gender"),
		Job:                           r.FormValue("job"),
		Social_situation:              r.FormValue("social_situation"),
		Address_from_national_id_card: r.FormValue("address_from_national_id_card"),
		Actual_address:                r.FormValue("actual_address"),
		District:                      r.FormValue("district"),
		PhoneNumbers:                  sql.NullString{String: r.FormValue("phone_numbers"), Valid: true},
		Created_at:                    createdAt,
		Updated_at:                    createdAt,
	}

	if err := Case.Create(db); err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

	if err := models.CreateLogs(db, id, "إنشاء حالة جديدة", userData.Id); err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "تمت إضافة الحالة بنجاح"}); err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}
}

func DeleteCase(w http.ResponseWriter, r *http.Request) {
	token, err := middleware.GetToken(r)

	if err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	userData, err := middleware.ValidateToken(token)

	if err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	middleware.VerifyAdmin(w, r)

	db := config.Database()
	defer db.Close()

	id := r.PathValue("id")

	Case := models.Cases{
		Id: id,
	}

	if err := Case.Delete(db); err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	Response := map[string]interface{}{}
	Response["status"] = "تمت العملية بنجاح"

	if err := models.CreateLogs(db, Case.Id, "حذف الحالة", userData.Id); err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(Response); err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}
}

func CaseApi(w http.ResponseWriter, r *http.Request) {
	db := config.Database()
	defer db.Close()

	id := r.PathValue("id")

	Case := models.CaseDitails{
		Id: sql.NullString{
			String: id,
			Valid:  true,
		},
	}

	cas, err := Case.Get(db)

	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			fmt.Println(err)
			middleware.ErrorResopnse(w, err)
			return
		}
	}

	path := fmt.Sprintf("uploads/%s", id)

	exists, err := middleware.Exists(path)

	errText := fmt.Sprintf("CreateFile %s: The system cannot find the file specified.", path)

	if err != nil {
		if err.Error() != errText {
			fmt.Println(err)
			middleware.ErrorResopnse(w, err)
			return
		}
	}

	Response := map[string]interface{}{
		"Data":     cas,
		"hasFiles": exists,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(Response); err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}
}

func SearchCase(w http.ResponseWriter, r *http.Request) {
	db := config.Database()
	defer db.Close()

	search := "%" + r.URL.Query().Get("search") + "%"

	Case := models.Cases{}

	cases, err := Case.Search(db, search)

	if err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(cases); err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

}

func UpdateCase(w http.ResponseWriter, r *http.Request) {
	token, err := middleware.GetToken(r)

	if err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	userData, err := middleware.ValidateToken(token)

	if err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	var casesUpdate map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&casesUpdate); err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	age, err := strconv.Atoi(casesUpdate["age"].(string))

	if err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	total_income, err := strconv.Atoi(casesUpdate["total_income"].(string))

	if err != nil {
		fmt.Println(err)
		middleware.ErrorResopnse(w, err)
		return
	}

	cases := models.Cases{
		Id:                            casesUpdate["id"].(string),
		Case_name:                     casesUpdate["case_name"].(string),
		National_id:                   casesUpdate["national_id"].(string),
		Devices_needed_for_the_case:   casesUpdate["devices_needed_for_the_case"].(string),
		Total_income:                  total_income,
		Age:                           age,
		Gender:                        casesUpdate["gender"].(string),
		Job:                           casesUpdate["job"].(string),
		Social_situation:              casesUpdate["social_situation"].(string),
		Address_from_national_id_card: casesUpdate["address_from_national_id_card"].(string),
		Actual_address:                casesUpdate["actual_address"].(string),
		District:                      casesUpdate["district"].(string),
		Debts:                         casesUpdate["debts"].(string),
		Pension_from_husband:          casesUpdate["pension_from_husband"].(string),
		Pension_from_father:           casesUpdate["pension_from_father"].(string),
		PhoneNumbers:                  sql.NullString{String: casesUpdate["phone_numbers"].(string), Valid: true},
		Updated_at:                    sql.NullString{String: time.Now().Format("2006-01-02 15:04:05"), Valid: true},
	}

	db := config.Database()
	defer db.Close()

	if err := cases.Update(db); err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

	if err := models.CreateLogs(db, cases.Id, "تعديل على معلومات الحالة", userData.Id); err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}

	Response := map[string]interface{}{"status": "تمت العملية بنجاح"}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(Response); err != nil {
		middleware.ErrorResopnse(w, err)
		return
	}
}
