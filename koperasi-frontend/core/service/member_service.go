// Package service memuat logika bisnis (di antara handler dan repository).
package service

import (
	"context"
	"fmt"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

// MemberService: aturan bisnis domain Keanggotaan (M1). Handler memanggil
// service ini; service mendelegasikan akses data ke MemberRepository.
type MemberService struct {
	repo repository.MemberRepository
}

func NewMemberService(repo repository.MemberRepository) *MemberService {
	return &MemberService{repo: repo}
}

// List mengembalikan anggota (terfilter status) beserta hitungan per status.
func (s *MemberService) List(ctx context.Context, status string) ([]model.Member, repository.MemberCounts, error) {
	members, err := s.repo.List(ctx, status)
	if err != nil {
		return nil, repository.MemberCounts{}, err
	}
	counts, err := s.repo.Counts(ctx)
	if err != nil {
		return nil, repository.MemberCounts{}, err
	}
	return members, counts, nil
}

func (s *MemberService) FindByID(ctx context.Context, id int) (*model.Member, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *MemberService) SimpananHistory(ctx context.Context, memberID int) ([]model.SimpananTransaction, error) {
	return s.repo.SimpananHistory(ctx, memberID)
}

func (s *MemberService) LoansByMember(ctx context.Context, memberID int) ([]model.Loan, error) {
	return s.repo.LoansByMember(ctx, memberID)
}

// Register membuat anggota baru berstatus PENDING (NomorAnggota "-").
func (s *MemberService) Register(ctx context.Context, nama, nik, alamat, noHP string) (int, error) {
	return s.repo.Create(ctx, nama, nik, alamat, noHP)
}

// Approve menyetujui anggota PENDING -> AKTIF dan mengembalikan nomor anggota baru.
func (s *MemberService) Approve(ctx context.Context, id int) (string, error) {
	return s.repo.Approve(ctx, id)
}

// Reject menolak pendaftaran (PENDING -> NON_AKTIF).
func (s *MemberService) Reject(ctx context.Context, id int) error {
	return s.repo.SetStatus(ctx, id, "NON_AKTIF")
}

// Resign memproses pengunduran diri (-> NON_AKTIF).
func (s *MemberService) Resign(ctx context.Context, id int) error {
	loans, err := s.repo.LoansByMember(ctx, id)
	if err != nil {
		return err
	}
	for _, loan := range loans {
		if loan.Status == "PENDING" || loan.Status == "DISETUJUI" || loan.Status == "AKTIF" {
			return fmt.Errorf("anggota masih memiliki pinjaman %s", loan.Status)
		}
	}
	return s.repo.SetStatus(ctx, id, "NON_AKTIF")
}

// RecordSimpanan mencatat transaksi simpanan secara transaksional (saldo + jurnal).
func (s *MemberService) RecordSimpanan(ctx context.Context, memberID int, jenis, tipe string, nominal float64, keterangan string) error {
	return s.repo.RecordSimpanan(ctx, memberID, jenis, tipe, nominal, keterangan)
}
