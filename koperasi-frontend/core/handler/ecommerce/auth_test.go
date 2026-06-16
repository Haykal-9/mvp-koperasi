package ecommerce

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
	"koperasi-frontend/core/service"

	"github.com/gin-gonic/gin"
)

type fakeMemberSearchRepo struct {
	repository.ECommerceRepository
}

func (f *fakeMemberSearchRepo) SearchMembers(_ context.Context, q string) ([]model.Member, error) {
	if q != "rina" {
		return nil, nil
	}
	return []model.Member{{
		ID:           2,
		Nama:         "Rina Pertiwi",
		NomorAnggota: "KOP-0002",
		Status:       "AKTIF",
	}}, nil
}

func TestSearchKoperasiMemberReturnsResultsEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(nil, service.NewECAccountService(&fakeMemberSearchRepo{}))
	r := gin.New()
	r.GET("/ecommerce/api/members/search", h.SearchKoperasiMember)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ecommerce/api/members/search?q=rina", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, mau %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Query   string `json:"query"`
		Results []struct {
			ID           int    `json:"id"`
			Nama         string `json:"nama"`
			NomorAnggota string `json:"nomor_anggota"`
			Status       string `json:"status"`
		} `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Query != "rina" || len(body.Results) != 1 {
		t.Fatalf("body = %+v, mau query rina dengan 1 hasil", body)
	}
	got := body.Results[0]
	if got.ID != 2 || got.Nama != "Rina Pertiwi" || got.NomorAnggota != "KOP-0002" || got.Status != "AKTIF" {
		t.Fatalf("hasil search = %+v", got)
	}
}
