package bwsonline

import (
	"context"

	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
)

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -struct_name=Dao
	AwardPackageList(ctx context.Context, bid int64) ([]int64, error)
	// bts: -struct_name=Dao
	AwardPackage(ctx context.Context, id int64) (*bwsonline.AwardPackage, error)
	// bts: -struct_name=Dao
	AwardPackageByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.AwardPackage, error)
	// bts: -struct_name=Dao
	Award(ctx context.Context, id int64) (*bwsonline.Award, error)
	// bts: -struct_name=Dao
	AwardByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.Award, error)
	// bts: -struct_name=Dao -nullcache=[]*bwsonline.AwardPackage{{ID:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].ID==-1
	UserPackage(ctx context.Context, mid int64) ([]*bwsonline.AwardPackage, error)
	// bts: -struct_name=Dao -nullcache=[]*bwsonline.UserAward{{Award:&bwsonline.Award{ID:-1}}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].Award!=nil&&$[0].Award.ID==-1
	UserAward(ctx context.Context, mid int64, bid int64) ([]*bwsonline.UserAward, error)
	// bts: -struct_name=Dao -nullcache=map[int64]int64{-1:-1} -check_null_code=len($)==1&&$[-1]==-1
	UserCurrency(ctx context.Context, mid int64, bid int64) (map[int64]int64, error)
	// bts: -struct_name=Dao
	Dress(ctx context.Context, id int64) (*bwsonline.Dress, error)
	// bts: -struct_name=Dao
	DressByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.Dress, error)
	// bts: -struct_name=Dao -nullcache=[]*bwsonline.UserDress{{ID:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].ID==-1
	UserDress(ctx context.Context, mid int64) ([]*bwsonline.UserDress, error)
	// bts: -struct_name=Dao
	Piece(ctx context.Context, id int64) (*bwsonline.Piece, error)
	// bts: -struct_name=Dao -nullcache=[]*bwsonline.UserPiece{{Pid:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].Pid==-1
	UserPiece(ctx context.Context, mid int64, bid int64) ([]*bwsonline.UserPiece, error)
	// bts: -struct_name=Dao
	PrintList(ctx context.Context, bid int64) ([]int64, error)
	// bts: -struct_name=Dao
	Print(ctx context.Context, id int64) (*bwsonline.Print, error)
	// bts: -struct_name=Dao
	PrintByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.Print, error)
	// bts: -struct_name=Dao -nullcache=map[int64]string{-1:"-1"} -check_null_code=len($)==1&&$[-1]=="-1"
	UserPrint(ctx context.Context, mid int64) (map[int64]string, error)
	// bts: -struct_name=Dao
	PieceUsedLog(ctx context.Context, mid int64, batchIDs []string) (map[string]map[int64]int64, error)
	// bts: -struct_name=Dao -nullcache=-1 -check_null_code=$==-1
	LastAutoEnergy(ctx context.Context, mid int64, bid int64) (int64, error)
}
