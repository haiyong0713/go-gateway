package cd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	"go-gateway/app/app-svr/fawkes/service/tools/appstoreconnect"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"

	goGitlab "github.com/xanzy/go-gitlab"
)

// fawkes consts.
const (
	BetaStateDisable   = -1
	BetaStateInit      = 0
	BetaStateUploaded  = 1
	BetaStateProcessed = 2
	BetaStateInReview  = 3
	BetaStateApproved  = 4
	BetaStateTesting   = 5
	BetaStateStopped   = 6

	BuglyPrefix = "bugly-"
)

func (s *Service) registerToAppstoreConnectClient() (err error) {
	var (
		apps []*cdmdl.TestFlightAppInfo
	)
	if apps, err = s.fkDao.TFAllAppsInfo(context.Background()); err != nil {
		log.Error("registerToAppstoreConnectClient: %v", err)
		return
	}
	for _, app := range apps {
		s.appstoreClient.RegisterApp(app.AppKey, utils.FawkesDecode(app.KeyID), utils.FawkesDecode(app.IssuerID))
	}
	return
}

// TestFlightAppInfoSet set testflight app info.
func (s *Service) TestFlightAppInfoSet(c context.Context, appKey, storeAppID, issuerID, keyID, tagPrefix, buglyAppID, buglyAppKey string, file multipart.File, header *multipart.FileHeader) (err error) {
	var (
		r        *cdmdl.TestFlightAppInfo
		tx       *sql.Tx
		destFile *os.File
	)
	pkPath := s.c.AppstoreConnect.KeyPath + "AuthKey_" + keyID + ".p8"
	if _, err = os.Stat(pkPath); err != nil {
		if os.IsNotExist(err) {
			if destFile, err = os.Create(pkPath); err != nil {
				log.Error("%v", err)
				return
			}
			if _, err = io.Copy(destFile, file); err != nil {
				log.Error("%v", err)
				return
			}
			defer destFile.Close()
		}
	}
	defer file.Close()
	encryptIssuerID := utils.FawkesEncode(issuerID)
	encryptKeyID := utils.FawkesEncode(keyID)
	if r, err = s.fkDao.TFAppInfo(c, appKey); err != nil {
		log.Error("TestFlightAppInfoSet: %v", err)
		return
	}
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if r != nil {
		if err = s.fkDao.TxUpTFAppInfo(tx, appKey, storeAppID, encryptIssuerID, encryptKeyID, tagPrefix, buglyAppID, buglyAppKey); err != nil {
			log.Error("TestFlightAppInfoSet: %v", err)
			return
		}
	} else {
		if err = s.fkDao.TxSetTFAppInfo(tx, appKey, storeAppID, encryptIssuerID, encryptKeyID, tagPrefix, buglyAppID, buglyAppKey); err != nil {
			log.Error("TestFlightAppInfoSet: %v", err)
			return
		}
	}
	return
}

// TestFlightAppInfo get testflight app info.
func (s *Service) TestFlightAppInfo(c context.Context, appKey string) (r *cdmdl.TestFlightAppInfo, err error) {
	if r, err = s.fkDao.TFAppInfo(c, appKey); err != nil {
		log.Error("TestFlightAppInfo: %v", err)
		return
	}
	return
}

// UploadToAppleStoreConnect upload ipa to app store connect
func (s *Service) UploadToAppleStoreConnect(appKey, ipaPath string, packID int64, packType int8, gitlabJobId int64) (err error) {
	ipaDescPath := filepath.Dir(ipaPath) + "/AppStoreInfo.plist"
	if _, err = os.Stat(ipaDescPath); err != nil {
		log.Error("Can't upload ipa without AppStoreInfo.plist: %v", ipaDescPath)
		return
	}
	out, err := s.appstoreClient.UploadIPA(appKey, ipaPath, ipaDescPath)
	resultReport(err == nil, out, appKey, filepath.Dir(ipaPath)+"/upload_log.txt", gitlabJobId, s.fkDao, s.c)
	if err != nil {
		log.Error("UploadIPA: %v", err)
		return
	}
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var guideTFTxt, remindUpdTxt, forceUpdTxt string
	//nolint:gomnd
	if packType == 9 {
		guideTFTxt = "邀请您参与B站内测，第一时间体验新功能 （提示：测试版本不支持内购）"
		remindUpdTxt = "有新的测试版本，是否立即升级？"
		forceUpdTxt = "感谢您参与测试，当前内测版本已经结束，请更新版本"
	} else if packType == 4 {
		remindUpdTxt = "感谢您参与测试，为避免影响您的日常使用，建议及时升级为正式版"
		forceUpdTxt = "感谢您参与测试，当前内测版本已经结束，请更新至正式版"
	}
	if err = s.fkDao.TxSetTFPackInfo(tx, appKey, packID, guideTFTxt, remindUpdTxt, forceUpdTxt); err != nil {
		log.Error("TxSetTFPackInfo: %v", err)
	}
	return
}

// SubmitBetaReview submit a pack to beta review
func (s *Service) SubmitBetaReview(appKey string, packTFID int64, betaBuildID string) (err error) {
	var (
		res *appstoreconnect.BetaAppReviewSubmissionResponse
		tx  *sql.Tx
	)
	if res, _, err = s.appstoreClient.Submissions.SubmitForBetaReview(appKey, betaBuildID); err != nil {
		log.Error("SubmitForBetaReview: %v", err)
		return
	}
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpTFReviewState(tx, packTFID, res.Data.Attributes.BetaReviewState, BetaStateInReview); err != nil {
		log.Error("TxUpTFReviewState: %v", err)
	}
	return
}

// DistributeTestFlight distribute a testflight pack for external testing.
func (s *Service) DistributeTestFlight(appKey string, packTFID int64, disPermil int, disLimit int64) (err error) {
	var (
		buildsIDs   []string
		betaGroupID string
		tfAttr      *cdmdl.TestFlightPackInfo
	)
	if tfAttr, err = s.fkDao.TFPackByPackTFID(context.Background(), packTFID); err != nil {
		log.Error("s.fkDao.TFPackByPackTFID %v", err)
		return
	}
	buildsIDs = append(buildsIDs, tfAttr.BetaBuildID)
	if tfAttr.Env == "prod" {
		betaGroupID = tfAttr.BetaGroupID
	} else if tfAttr.Env == "test" {
		betaGroupID = tfAttr.BetaGroupIDTest
	} else {
		err = errors.New("unknown env")
		log.Error("unknown env: %v", tfAttr.Env)
		return
	}
	if _, err = s.appstoreClient.BetaGroups.AddBuilds(appKey, betaGroupID, buildsIDs); err != nil {
		log.Error("s.appstoreClient.BetaGroups.AddBuilds: %v", err)
		return
	}
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxTFPackDistribute(tx, packTFID, disPermil, disLimit); err != nil {
		log.Error("TxTFPackDistribute: %v", err)
	}
	return
}

// StopTestFlight stop a testflight pack for external testing.
func (s *Service) StopTestFlight(appKey string, packTFID int64) (err error) {
	var (
		buildsIDs   []string
		betaGroupID string
		tfAttr      *cdmdl.TestFlightPackInfo
	)
	if tfAttr, err = s.fkDao.TFPackByPackTFID(context.Background(), packTFID); err != nil {
		log.Error("s.fkDao.TFPackByPackTFID %v", err)
		return
	}
	buildsIDs = append(buildsIDs, tfAttr.BetaBuildID)
	if tfAttr.Env == "prod" {
		betaGroupID = tfAttr.BetaGroupID
	} else if tfAttr.Env == "test" {
		betaGroupID = tfAttr.BetaGroupIDTest
	} else {
		err = errors.New("unknown env")
		log.Error("unknown env: %v", tfAttr.Env)
		return
	}
	if _, err = s.appstoreClient.BetaGroups.RemoveBuilds(appKey, betaGroupID, buildsIDs); err != nil {
		log.Error("RemoveBuilds: %v", err)
		return
	}
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpTFBetaState(tx, packTFID, BetaStateStopped); err != nil {
		log.Error("TxUpTFBetaState: %v", err)
	}
	return
}

// UpdateRemindUpdTime update remind update time
func (s *Service) UpdateRemindUpdTime(packTFID, remindUpdTime int64) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpRemindUpdTime(tx, packTFID, remindUpdTime); err != nil {
		log.Error("TxUpRemindUpdTime: %v", err)
	}
	return
}

// UpdateForceUpdTime update force update time
func (s *Service) UpdateForceUpdTime(packTFID, forceUpdTime int64) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpForceUpdTime(tx, packTFID, forceUpdTime); err != nil {
		log.Error("TxUpForceUpdTime: %v", err)
	}
	return
}

// SetBetaGroups set beta groups for testflight
func (s *Service) SetBetaGroups(appKey, publicLink, publicLinkTest string) (err error) {
	var (
		res                          *appstoreconnect.BetaGroupsResponse
		app                          *cdmdl.TestFlightAppInfo
		tx                           *sql.Tx
		betaGroupID, betaGroupIDTest string
	)
	if app, err = s.fkDao.TFAppInfo(context.Background(), appKey); err != nil {
		log.Error("TFAppInfo: %v", err)
		return
	}
	if res, _, err = s.appstoreClient.Apps.BetaGroups(appKey, app.StoreAppID); err != nil {
		log.Error("BetaGroups: %v", err)
		return
	}
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	for _, betaGroup := range res.BetaGroups {
		if betaGroup.Attributes.PublicLink == publicLink {
			betaGroupID = betaGroup.ID
		} else if betaGroup.Attributes.PublicLink == publicLinkTest {
			betaGroupIDTest = betaGroup.ID
		}
	}
	if betaGroupID == "" || betaGroupIDTest == "" {
		err = errors.New("cannot find the beta groups")
		return
	}
	if err = s.fkDao.TxUpBetaGroup(tx, appKey, betaGroupID, publicLink, betaGroupIDTest, publicLinkTest); err != nil {
		log.Error("TxUpBetaGroup: %v", err)
	}
	return
}

// TFSetUpdTxt set testflight update text
func (s *Service) TFSetUpdTxt(packTFID int64, guideTFTxt, remindUpdTxt, forceUpdTxt string) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpTFUpdTxt(tx, packTFID, guideTFTxt, remindUpdTxt, forceUpdTxt); err != nil {
		log.Error("TxUpTFUpdTxt: %v", err)
	}
	return
}

// UploadStateProc refresh the ipa upload state
func (s *Service) UploadStateProc() (err error) {
	var tfPacks []*cdmdl.TestFlightPackInfo
	if tfPacks, err = s.fkDao.TFPackInfoWithState(context.Background(), BetaStateInit); err != nil {
		log.Error("UploadStateProc: %v", err)
		return
	}
	for _, pack := range tfPacks {
		outputPath := filepath.Dir(pack.PackPath) + "/upload_log.txt"
		if _, err = os.Stat(outputPath); err != nil {
			err = nil
			continue
		}
		var read []byte
		if read, err = ioutil.ReadFile(outputPath); err != nil {
			log.Error("ReadFile %v: %v", outputPath, err)
			continue
		}
		if strings.Contains(string(read), "uploaded successfully:") {
			opt := &appstoreconnect.BuildsOptions{
				AppFilter:     pack.StoreAppID,
				VersionFilter: strconv.FormatInt(pack.VersionCode, 10),
			}
			var res *appstoreconnect.BuildsResponse
			if res, _, err = s.appstoreClient.Builds.Builds(pack.AppKey, opt); err != nil {
				log.Error("s.appstoreClient.Builds.Builds: %v", err)
				continue
			}
			if len(res.Builds) > 0 {
				if err = s.handleUploadedState(pack, res.Builds[0]); err != nil {
					log.Error("handleUploadedState: %v", err)
					continue
				}
			}
		} else {
			if err = s.upBetaState(pack.ID, BetaStateDisable); err != nil {
				log.Error("upBetaState: %v", err)
				continue
			}
		}
	}
	return
}

func (s *Service) handleUploadedState(packInfo *cdmdl.TestFlightPackInfo, build appstoreconnect.Build) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpTFPackInfo(tx, packInfo.ID, build.ID, build.Attributes.ProcessingState, build.Attributes.ExpirationDate.Unix()); err != nil {
		log.Error("TxUpTFPackInfo: %v", err)
		return
	}
	//nolint:gomnd
	remindUpdTime := build.Attributes.ExpirationDate.Unix() - 60*60*24*7*2
	//nolint:gomnd
	forceUpdTime := build.Attributes.ExpirationDate.Unix() - 60*60*24*7
	if err = s.fkDao.TxUpRemindUpdTime(tx, packInfo.ID, remindUpdTime); err != nil {
		log.Error("TxUpTFPackInfo: %v", err)
		return
	}
	if err = s.fkDao.TxUpForceUpdTime(tx, packInfo.ID, forceUpdTime); err != nil {
		log.Error("TxUpTFPackInfo: %v", err)
	}
	return
}

func (s *Service) upBetaState(packTFID int64, betaState int) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpTFBetaState(tx, packTFID, betaState); err != nil {
		log.Error("TxUpTFBetaState error(%v)", err)
	}
	return
}

// PackStateProc refresh the ipa build state
func (s *Service) PackStateProc() (err error) {
	var tfPacks []*cdmdl.TestFlightPackInfo
	if tfPacks, err = s.fkDao.TFPackInfoWithState(context.Background(), BetaStateUploaded); err != nil {
		log.Error("PackStateProc: %v", err)
		return
	}
	for _, pack := range tfPacks {
		opt := &appstoreconnect.BuildsOptions{
			AppFilter:     pack.StoreAppID,
			VersionFilter: strconv.FormatInt(pack.VersionCode, 10),
		}
		var res *appstoreconnect.BuildsResponse
		if res, _, err = s.appstoreClient.Builds.Builds(pack.AppKey, opt); err != nil {
			log.Error("s.appstoreClient.Builds.Builds: %v", err)
			continue
		}
		if len(res.Builds) > 0 {
			packState := res.Builds[0].Attributes.ProcessingState
			if packState == "FAILED" || packState == "INVALID" {
				if err = s.upPackState(pack.ID, packState, BetaStateDisable); err != nil {
					log.Error("TxUpTFPackState: %v", err)
					continue
				}
			} else if packState == "VALID" {
				if err = s.upPackState(pack.ID, packState, BetaStateProcessed); err != nil {
					log.Error("upPackState: %v", err)
					continue
				}
				if _, _, err = s.appstoreClient.Builds.ModifyBuild(pack.AppKey, pack.BetaBuildID, false, false); err != nil {
					log.Error("ModifyBuild: %v", err)
					continue
				}
			}
		}
	}
	return
}

// UpdateAppStoreConnectProc update app store connect app info
func (s *Service) UpdateAppStoreConnectProc() (err error) {
	if err = s.registerToAppstoreConnectClient(); err != nil {
		log.Error("UpdateAppStoreConnectProc: %v", err)
	}
	return
}

func (s *Service) upPackState(packTFID int64, packState string, betaState int) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpTFPackState(tx, packTFID, packState, betaState); err != nil {
		log.Error("TxUpTFPackState: %v", err)
	}
	return
}

// ReviewStateProc refresh the reivew state of builds
func (s *Service) ReviewStateProc() (err error) {
	var tfPacks []*cdmdl.TestFlightPackInfo
	if tfPacks, err = s.fkDao.TFPackInfoWithState(context.Background(), BetaStateInReview); err != nil {
		log.Error("PackStateProc: %v", err)
		return
	}
	for _, pack := range tfPacks {
		var res *appstoreconnect.BetaAppReviewSubmissionResponse
		if res, _, err = s.appstoreClient.Builds.BetaAppReviewSubmission(pack.AppKey, pack.BetaBuildID); err != nil {
			log.Error("BetaAppReviewSubmission: %v", err)
			continue
		}
		if res.Data.Attributes.BetaReviewState == "REJECTED" {
			if err = s.upReviewState(pack.ID, res.Data.Attributes.BetaReviewState, BetaStateDisable); err != nil {
				log.Error("upReviewState: %v", err)
			}
		} else if res.Data.Attributes.BetaReviewState == "APPROVED" {
			if err = s.upReviewState(pack.ID, res.Data.Attributes.BetaReviewState, BetaStateApproved); err != nil {
				log.Error("upReviewState: %v", err)
			}
		} else if res.Data.Attributes.BetaReviewState == "IN_REVIEW" {
			if err = s.upReviewState(pack.ID, res.Data.Attributes.BetaReviewState, BetaStateInReview); err != nil {
				log.Error("upReviewState: %v", err)
			}
		}
	}
	return
}

func (s *Service) upReviewState(packTFID int64, reviewState string, betaState int) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpTFReviewState(tx, packTFID, reviewState, betaState); err != nil {
		log.Error("TxUpTFReviewState: %v", err)
	}
	return
}

// DistributeNumProc refresh the user number of the testflight ipa.
func (s *Service) DistributeNumProc() (err error) {
	var tfPacks []*cdmdl.TestFlightPackInfo
	if tfPacks, err = s.fkDao.TFPackInfoWithState(context.Background(), BetaStateTesting); err != nil {
		log.Error("PackStateProc: %v", err)
		return
	}
	for _, pack := range tfPacks {
		// 更新用户人数
		var disNum int64
		if disNum, err = s.fkDao.TestFlightPackUser(context.Background(), pack.AppKey, pack.VersionCode, pack.CTime.Time().Unix()); err != nil {
			log.Error("TestFlightPackUser: %v", err)
			continue
		}
		if err = s.upDisNum(pack.ID, disNum); err != nil {
			log.Error("upDisNum: %v", err)
			continue
		}
		if disNum >= pack.DisLimit {
			var (
				buildIDs    []string
				betaGroupID string
			)
			buildIDs = append(buildIDs, pack.BetaBuildID)
			if err = s.upBetaState(pack.ID, BetaStateStopped); err != nil {
				log.Error("upBetaState: %v", err)
				continue
			}
			if pack.Env == "prod" {
				betaGroupID = pack.BetaGroupID
			} else if pack.Env == "test" {
				betaGroupID = pack.BetaGroupIDTest
			} else {
				err = errors.New("unknown env")
				log.Error("unknown env: %v", pack.Env)
				return
			}
			if _, err = s.appstoreClient.BetaGroups.RemoveBuilds(pack.AppKey, betaGroupID, buildIDs); err != nil {
				log.Error("RemoveBuilds: %v", err)
				continue
			}
		}
	}
	return
}

func (s *Service) upDisNum(packTFID, disNum int64) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpTFPackDisNum(tx, packTFID, disNum); err != nil {
		log.Error("TxUpTFReviewState: %v", err)
	}
	return
}

// DeleteTestersProc delete testers from betagroups.
func (s *Service) DeleteTestersProc() (err error) {
	var (
		apps []*cdmdl.TestFlightAppInfo
		res  *appstoreconnect.BetaTestersResponse
	)
	if apps, err = s.fkDao.TFAllAppsInfo(context.Background()); err != nil {
		log.Error("DeleteTestersProc: %v", err)
		return
	}
	for _, app := range apps {
		if len(app.BetaGroupID) == 0 {
			continue
		}
		if res, _, err = s.appstoreClient.BetaGroups.BetaTesters(app.AppKey, app.BetaGroupID); err != nil {
			log.Error("DeleteTestersProc: %v %v %v", app.AppKey, app.BetaGroupID, err)
			continue
		}
		if res.Meta.Paging.Total > s.c.AppstoreConnect.TestersThreshold {
			var betaTestersIDs []string
			for _, tester := range res.BetaTesters {
				betaTestersIDs = append(betaTestersIDs, tester.ID)
			}
			if _, err = s.appstoreClient.Apps.RemoveBetaTesters(app.AppKey, app.StoreAppID, betaTestersIDs); err != nil {
				log.Error("DeleteTestersProc: %v %v %v", app.AppKey, app.BetaGroupID, err)
			}
		}
	}
	return
}

// DisableExpiredProc set the beta state to disable for the expired testflight package.
func (s *Service) DisableExpiredProc() (err error) {
	var tfPacks []*cdmdl.TestFlightPackInfo
	if tfPacks, err = s.fkDao.TFPackInfoValid(context.Background()); err != nil {
		log.Error("DisableExpiredProc: %v", err)
		return
	}
	for _, pack := range tfPacks {
		if pack.ExpireTime.Time().Unix() < time.Now().Unix() {
			if err = s.upBetaState(pack.ID, BetaStateDisable); err != nil {
				log.Error("upBetaState: %v", err)
				continue
			}
		}
	}
	return
}

// UpdateOnlineVersProc update online version
func (s *Service) UpdateOnlineVersProc() (err error) {
	var (
		apps     []*cdmdl.TestFlightAppInfo
		verRes   *appstoreconnect.AppStoreVersionsResponse
		buildRes *appstoreconnect.BuildResponse
	)
	if apps, err = s.fkDao.TFAllAppsInfo(context.Background()); err != nil {
		log.Error("UpdateOnlineVersProc: %v", err)
		return
	}
	for _, app := range apps {
		opt := &appstoreconnect.AppStoreVersionsOption{
			AppStoreStateFilter: "READY_FOR_SALE",
		}
		// 查找 AppStore 线上版本号
		if verRes, _, err = s.appstoreClient.Apps.AppStoreVersions(app.AppKey, app.StoreAppID, opt); err != nil {
			log.Error("UpdateOnlineVersProc: %v", err)
			continue
		}
		if len(verRes.AppStoreVersion) == 1 {
			onlineVer := verRes.AppStoreVersion[0]
			if buildRes, _, err = s.appstoreClient.AppStoreVersions.Build(app.AppKey, onlineVer.ID); err != nil {
				log.Error("UpdateOnlineVersProc: %v", err)
				continue
			}
			var (
				buildID, versionCode    int64
				commit, gitlabProjectID string
				opt                     *goGitlab.CreateTagOptions
			)
			if versionCode, err = strconv.ParseInt(buildRes.Build.Attributes.Version, 10, 64); err != nil {
				log.Error("UpdateOnlineVersProc: %v", err)
				continue
			}
			if app.OnlineVersionCode != versionCode {
				if _, buildID, commit, err = s.fkDao.OnlineBuildID(context.Background(), buildRes.Build.ID); err != nil {
					log.Error("UpdateOnlineVersProc: %v", err)
					continue
				}
				// 1.自动打 Tag
				opt = &goGitlab.CreateTagOptions{
					TagName: goGitlab.String(app.TagPrefix + onlineVer.Attributes.VersionString),
					Ref:     goGitlab.String(commit),
				}
				if gitlabProjectID, err = s.fkDao.GitlabProjectID(context.Background(), app.AppKey); err != nil {
					log.Error("UpdateOnlineVersProc GitlabProjectID: %v", err)
					continue
				}
				if _, _, err = s.gitlabClient.Tags.CreateTag(gitlabProjectID, opt); err != nil {
					log.Error("UpdateOnlineVersProc CreateTag: %v", err)
					continue
				}
				// 2.自动推正式环境
				if _, _, err = s.CDEvolution(context.Background(), app.AppKey, "test", "fawkes", 0, 0, buildID); err != nil {
					log.Error("UpdateOnlineVersProc CDEvolution: %v", err)
					continue
				}
				// 3.更新 App 的线上包信息
				if err = s.upOnlineInfo(app.AppKey, onlineVer.Attributes.VersionString, versionCode, buildID); err != nil {
					log.Error("UpdateOnlineVersProc: %v", err)
					continue
				}
			}
		}
	}
	return
}

func (s *Service) upOnlineInfo(appKey, onlineVer string, onlineVerCode, buildID int64) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpOnlineInfo(tx, appKey, onlineVer, onlineVerCode, buildID); err != nil {
		log.Error("TxUpOnlineInfo: %v", err)
	}
	return
}

func (s *Service) setTFPackInfoProd(packInfo *cdmdl.TestFlightPackInfo, betaState, disPermil int, disLimit, prodPackID int64) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxSetTFProdPack(tx, packInfo.AppKey, prodPackID, packInfo.BetaBuildID, int64(packInfo.ExpireTime), packInfo.PackState, packInfo.ReviewState,
		betaState, disPermil, packInfo.DisNum, disLimit, int64(packInfo.RemindUpdTime), int64(packInfo.ForceupdTime), packInfo.GuideTFTxt, packInfo.RemindUpdTxt, packInfo.ForceUpdTxt); err != nil {
		log.Error("TxSetTFProdPack error(%v)", err)
	}
	return
}

// TFUploadBugly upload dsym to bugly
func (s *Service) TFUploadBugly(c context.Context, appKey string, buildID int64) (err error) {
	var (
		pack        *cimdl.BuildPack
		appInfo     *cdmdl.TestFlightAppInfo
		appBaseInfo *appmdl.APP
	)
	if pack, err = s.fkDao.BuildPack(c, appKey, buildID); err != nil {
		log.Error("TFUploadBugly: %v", err)
		return
	}
	if appInfo, err = s.fkDao.TFAppInfo(c, appKey); err != nil {
		log.Error("TFUploadBugly: %v", err)
		return
	}
	if appBaseInfo, err = s.fkDao.AppInfo(c, appKey, -1); err != nil {
		log.Error("TFUploadBugly: %v", err)
		return
	}
	s.AddHandlerProc(func() {
		err = s.uploadBugly(appInfo, pack.PkgPath, appBaseInfo.AppID, pack.VersionCode)
	})
	return
}

func (s *Service) uploadBugly(appInfo *cdmdl.TestFlightAppInfo, packPath, bundleID string, versionCode int64) (err error) {
	var (
		f      *os.File
		output string
	)
	packDir := filepath.Dir(packPath)
	buglyDir := filepath.Join(filepath.Dir(packPath), "bugly")
	if err = os.MkdirAll(buglyDir, os.ModePerm); err != nil {
		log.Error("os.MkdirAll error %v", err)
		return
	}
	defer os.RemoveAll(buglyDir)
	f, err = os.Create(buglyDir + "/dsymUpload.txt")
	defer func() {
		_, err = f.Write([]byte(output))
		f.Close()
	}()
	if err != nil {
		log.Error("os.Create: %v", err)
		return
	}
	err = filepath.Walk(packDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Error("filepath.Walk error(%v)", err)
			return err
		}
		if f == nil {
			errMsg := "found no file"
			err = fmt.Errorf(errMsg)
			log.Error(errMsg)
			return err
		}
		if f.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".dSYM.zip") {
			// buglySymboliOS.jar会重新打包成一个已损坏的同名zip包，导致rhino服务出错，为了不污染原先的zip包，需要拷贝一份zip包来操作
			_, fileName := filepath.Split(path)
			copyPath := filepath.Join(buglyDir, fmt.Sprintf("%v%v", BuglyPrefix, fileName))
			if err = utils.FileCopy(path, copyPath); err != nil {
				log.Error("FileCopy error %v", err)
				return err
			}
			unzipPath := strings.TrimSuffix(copyPath, ".zip")
			if !isDir(unzipPath) {
				if err = utils.Unzip(copyPath, buglyDir); err != nil {
					log.Error("unzip(%s, %s) error(%v)", copyPath, buglyDir, err)
					return err
				}
			}
			var (
				out    bytes.Buffer
				errOut bytes.Buffer
			)
			dir, file := filepath.Split(unzipPath)
			execPath := filepath.Join(dir, strings.TrimPrefix(file, BuglyPrefix))
			execJava := fmt.Sprintf("%s/bin/java", os.Getenv("JAVA_HOME"))
			cmd := exec.Command(execJava, "-jar", s.c.AppstoreConnect.BuglyUploader, "-appid", appInfo.BuglyAppID, "-appkey", appInfo.BuglyAppKey, "-bundleid", bundleID, "-version", strconv.FormatInt(versionCode, 10), "-platform", "IOS", "-inputSymbol", execPath)
			cmd.Stdout = &out
			cmd.Stderr = &errOut
			cmdStr := fmt.Sprintf("%s -jar %s -appid %s -appkey %s -bundleid %s -version %s -platform IOS -inputSymbol %s", execJava, s.c.AppstoreConnect.BuglyUploader, appInfo.BuglyAppID, appInfo.BuglyAppKey, bundleID, strconv.FormatInt(versionCode,
				10), execPath)
			if err = cmd.Run(); err != nil {
				output = output + "\n\n" + out.String() + "\n" + errOut.String()
				log.Error("Run cmd error: %s, stdout=(%s) stderr=(%s) error(%v)", cmdStr, out.String(), errOut.String(), err)
			} else {
				log.Warn("Run cmd success: %v", cmdStr)
				output = output + "\n\n" + out.String()
			}
			// 删掉bugly的临时zip包
			if err = os.Remove(copyPath); err != nil {
				log.Error("Remove error %v", err)
				return err
			}
		}
		return nil
	})
	return
}

// TestflightBWAdd add a user to testflight black/white list.
func (s *Service) TestflightBWAdd(appKey, env string, mid int64, nick, operator, listType string) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxSetTFBlackWhite(tx, appKey, env, mid, nick, operator, listType); err != nil {
		log.Error("TxSetTFBlackWhite %v", err)
	}
	return
}

// TestflightBWList list users in testflight black/white list.
func (s *Service) TestflightBWList(c context.Context, appKey, listType, env string) (res []*cdmdl.TestFlightBWList, err error) {
	if res, err = s.fkDao.TFBlackWhiteList(c, appKey, listType, env); err != nil {
		log.Error("TFBlackWhiteList: %v", err)
	}
	return
}

// TestflightBWDel remove the user from testflight black/white list.
func (s *Service) TestflightBWDel(ID int64) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxDelBlackWhite(tx, ID); err != nil {
		log.Error("TxDelBlackWhite %v", err)
	}
	return
}

// 判断所给路径是否为文件夹
func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// BetaPacks get beta testing packs
func (s *Service) BetaPacks(c context.Context) (res []*cdmdl.TestFlightPackInfo, err error) {
	if res, err = s.fkDao.TFPackInfoWithState(c, BetaStateTesting); err != nil {
		log.Error("TFPackInfoWithState: %v", err)
	}
	return
}

func resultReport(success bool, brief bytes.Buffer, appKey, logPath string, gitlabJobId int64, d *fkdao.Dao, conf *conf.Config) {
	title := "构建包上传AppStore失败"
	if success {
		title = "构建包上传AppStore成功"
	}
	users, _ := d.AuthUserNamesDistinct(context.Background(), appKey, "1")
	receivers := append(users, conf.AlarmReceiver.UploadMonitorReceiver...)
	if err := d.WechatCardMessageNotify(
		title,
		fmt.Sprintf("应用【%s】构建号【%d】\n %s", appKey, gitlabJobId, brief.String()),
		conf.LocalPath.LocalDomain+strings.TrimPrefix(logPath, conf.LocalPath.LocalDir),
		"",
		strings.Join(receivers, "|"),
		conf.Comet.FawkesAppID); err != nil {
		log.Error("error: %v", err)
	}
}

// logAnalyse 分析日志文件
/*func logAnalyse(logPath string) (bool, string) {
	file, err := os.Open(logPath)
	if err != nil {
		log.Error("open file error: %v", err)
		return false, ""
	}
	defer file.Close()
	var cursor int64 = 0
	var line string
	var fileContent string
	stat, _ := file.Stat()
	filesize := stat.Size()
	for {
		line, cursor, _ = tailLine(file, cursor)
		if strings.Contains(line, "uploaded successfully") {
			fileContent = fileContext(file, cursor)
			log.Error("file info: " + fileContent)
			return true, fileContent
		}
		if strings.Contains(line, "not uploaded because they had problems") {
			fileContent = fileContext(file, cursor)
			log.Error("file info: " + fileContent)
			return false, fileContent
		}
		if strings.Contains(line, "Package Summary:") {
			fileContent = fileContext(file, cursor)
			log.Error("get summary but don't know if success or not.")
			log.Error("file info: " + fileContent)
			return false, fileContent
		}
		if cursor == -filesize {
			log.Error("can't find summary")
			fileContent = "NO Package Summary"
			return false, fileContent
		}
	}
}

// tailLine 从文件结尾向前逐行扫描
func tailLine(file *os.File, start int64) (line string, cursor int64, err error) {
	line = ""
	cursor = start
	stat, _ := file.Stat()
	filesize := stat.Size()
	for {
		cursor--
		if _, err = file.Seek(cursor, io.SeekEnd); err != nil {
			log.Error("seek file error: %v", err)
			return
		}
		char := make([]byte, 1)
		if _, err = file.Read(char); err != nil {
			log.Error("read char error: %v", err)
			return
		}
		if cursor != -1 && (char[0] == 10 || char[0] == 13) {
			break
		}
		line = fmt.Sprintf("%s%s", string(char), line)
		if cursor == -filesize {
			break
		}
	}
	return
}

func fileContext(file *os.File, cursor int64) string {
	_, _ = file.Seek(cursor, io.SeekEnd)
	char := make([]byte, -cursor)
	_, _ = file.Read(char)
	return string(char)
}*/
