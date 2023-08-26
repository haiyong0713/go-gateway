package fawkes

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	xsql "go-common/library/database/sql"

	mdlmdl "go-gateway/app/app-svr/fawkes/service/model/modules"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_addGroup               = `INSERT INTO module_group (app_key,name,c_name) VALUES (?,?,?)`
	_editGroup              = `UPDATE module_group SET %s WHERE id=?`
	_delGroupModuleRelation = `DELETE FROM module_group_relation WHERE group_id=?`
	_delGroup               = `DELETE FROM module_group WHERE id=?`
	_getGroupIDByName       = `SELECT id,c_name FROM module_group WHERE app_key=? AND name=?`
	_listModuleGroup        = `SELECT m.id AS mid,m.name AS m_name,m.c_name AS m_cname,g.id AS gid,g.name AS g_name,g.c_name AS g_cname FROM module AS m,module_group AS g,module_group_relation AS r WHERE m.id=r.module_id AND g.id=r.group_id AND m.app_key=? ORDER BY g_name ASC`
	_listModuleUngroup      = `SELECT id,name,c_name FROM module WHERE app_key=? AND id NOT IN (SELECT module_id FROM module_group_relation) ORDER BY ctime DESC`
	_listEmptyGroup         = `SELECT id,name,c_name FROM module_group WHERE app_key=? AND id NOT IN (SELECT group_id FROM module_group_relation) ORDER BY ctime DESC`
	_listAllGroup           = `SELECT id,name,c_name FROM module_group WHERE app_key=?`

	_addModule         = `INSERT INTO module (app_key,name) VALUES (?,?)`
	_getModuleIDByName = `SELECT id FROM module WHERE app_key=? AND name=?`

	_existsRelation          = `SELECT count(*) FROM module_group_relation WHERE module_id=?`
	_existsRelationModuleIDs = `SELECT module_id FROM module_group_relation WHERE module_id IN (%v)`
	_setModuleGroupRelation  = `INSERT INTO module_group_relation (module_id,group_id,operator) VALUES (?,?,?) ON DUPLICATE KEY UPDATE group_id=VALUES(group_id),operator=VALUES(operator)`

	_addModuleSize     = `INSERT INTO module_size (app_key,build_id,module_id,lib_ver,size_type,size) VALUES (?,?,?,?,?,?)`
	_moduleSize        = `SELECT m.id,m.name,m.c_name,s.lib_ver,s.build_id,SUM(s.size),v.version,v.version_code FROM pack AS p,module AS m,module_size AS s,pack_version AS v WHERE s.module_id=m.id AND s.build_id=p.build_id AND p.version_id=v.id AND s.app_key=? AND p.env='prod' AND m.name=? %s GROUP BY p.build_id ORDER BY v.version_code DESC LIMIT 20`
	_listLatestVers    = `SELECT DISTINCT pv.version_code FROM pack AS p, pack_version AS pv WHERE p.env='prod' AND p.version_id=pv.id AND p.app_key=? ORDER BY p.id DESC LIMIT ?`
	_groupSize         = `SELECT %s AS g_size,p.build_id,g.id AS gid,g.name AS g_name,g.c_name AS g_cname,v.version,v.version_code FROM pack AS p,module AS m,module_group AS g,module_group_relation AS r,module_size AS s,pack_version AS v WHERE s.module_id=m.id AND m.id=r.module_id AND r.group_id=g.id AND s.build_id=p.build_id AND p.version_id=v.id AND s.app_key=? AND p.env='prod' AND g.name=? %s AND v.version_code IN (%s) GROUP BY p.build_id ORDER BY v.version_code DESC`
	_listSizeType      = `SELECT DISTINCT s.size_type FROM pack AS p,module AS m,module_group AS g,module_group_relation AS r,module_size AS s,pack_version AS v WHERE s.module_id=m.id AND m.id=r.module_id AND r.group_id=g.id AND s.build_id=p.build_id AND p.version_id=v.id AND s.app_key=? AND p.env='prod' AND v.version_code IN (%s)`
	_moduleSizeInGroup = `SELECT m.id,m.name,m.c_name,s.lib_ver,%s AS sum_size FROM module AS m,module_size AS s,module_group AS g,module_group_relation AS r WHERE s.module_id=m.id AND r.module_id=m.id AND g.id=r.group_id AND s.app_key=? AND g.name=? AND s.build_id=? %s GROUP BY m.name ORDER BY sum_size DESC`
	_groupSizeInBuild  = `SELECT %s AS g_size,g.id AS gid,g.name AS g_name,g.c_name AS g_cname FROM module AS m,module_group AS g,module_group_relation AS r,module_size AS s WHERE s.module_id=m.id AND m.id=r.module_id AND r.group_id=g.id AND s.app_key=? AND s.build_id=? %s GROUP BY g_name ORDER BY g_name DESC`

	_getPackBuildID = `SELECT gl_job_id FROM build_pack WHERE id=?`

	_setModuleConfigTotalSize = `INSERT INTO module_config (app_key,version,module_group_id,total_size,operator) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE total_size=?,fixed_size=?*percentage,operator=?`
	_setModuleConfig          = `INSERT INTO module_config (app_key,version,module_group_id,total_size,percentage,fixed_size,apply_normal_size,apply_force_size,external_size,description,operator) VALUES (?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE total_size=?,percentage=?,fixed_size=?,apply_normal_size=?,apply_force_size=?,external_size=?,description=?,operator=?`
	_getModuleConfig          = `SELECT mc.app_key,mc.version,mc.module_group_id,mc.total_size,mc.percentage,mc.fixed_size,mc.apply_normal_size,mc.apply_force_size,mc.external_size,mc.description,mc.operator FROM module_config AS mc, module_group AS mg WHERE mc.module_group_id =mg.id AND mc.app_key=? AND mc.version=?`

	_getNewestModulesConfVersion = `SELECT mc.version FROM module_config mc INNER JOIN pack_version pv ON mc.version = pv.version WHERE mc.app_key=? AND pv.env='prod' ORDER BY pv.version_code DESC LIMIT 1`
	_getPreciousVersion          = `SELECT version FROM pack_version WHERE version_code = (SELECT MAX(version_code) FROM pack_version WHERE env='prod' AND app_key=? and version!=?) limit 1`
	_getNewestVersionByTime      = `SELECT MAX(version_code) FROM pack_version WHERE env='prod' AND app_key=? AND ctime < ? LIMIT ?`
)

// TxAddGroup add group.
func (d *Dao) TxAddGroup(tx *xsql.Tx, appKey, name, cName string) (gid int64, err error) {
	res, err := tx.Exec(_addGroup, appKey, name, cName)
	if err != nil {
		log.Error("TxAddGroup %v", err)
		return
	}
	return res.LastInsertId()
}

// TxEditGroup edit group.
func (d *Dao) TxEditGroup(tx *xsql.Tx, gID int64, name, cName string) (r int64, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if name != "" {
		args = append(args, name)
		sqlAdd += "name=?"
	}
	if cName != "" {
		if len(args) > 0 {
			sqlAdd += ","
		}
		args = append(args, cName)
		sqlAdd += "c_name=?"
	}
	args = append(args, gID)
	res, err := tx.Exec(fmt.Sprintf(_editGroup, sqlAdd), args...)
	if err != nil {
		log.Error("TxEditGroup %v", err)
	}
	return res.RowsAffected()
}

// TxDelGroupModuleRelation delete relation with group & modules.
func (d *Dao) TxDelGroupModuleRelation(tx *xsql.Tx, gID int64) (r int64, err error) {
	res, err := tx.Exec(_delGroupModuleRelation, gID)
	if err != nil {
		log.Error("TxAddGroup %v", err)
		return
	}
	return res.RowsAffected()
}

// TxDelGroup delete a group.
func (d *Dao) TxDelGroup(tx *xsql.Tx, gID int64) (r int64, err error) {
	res, err := tx.Exec(_delGroup, gID)
	if err != nil {
		log.Error("TxDelGroup %v", err)
		return
	}
	return res.RowsAffected()
}

// GetGroupID get group id by app key & name.
func (d *Dao) GetGroupID(c context.Context, appKey, name string) (groupID int64, gCName string, err error) {
	res := d.db.QueryRow(c, _getGroupIDByName, appKey, name)
	if err = res.Scan(&groupID, &gCName); err != nil {
		if err != sql.ErrNoRows {
			log.Error("GetGroupID %v", err)
		}
	}
	return
}

// ListModuleGroup list grouped modules.
func (d *Dao) ListModuleGroup(c context.Context, appKey string) (res []*mdlmdl.Module, err error) {
	rows, err := d.db.Query(c, _listModuleGroup, appKey)
	if err != nil {
		log.Error("d.ListModuleGroup d.db.Query(%v) error(%v)", appKey, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var m = &mdlmdl.Module{}
		if err = rows.Scan(&m.MID, &m.MName, &m.MCName, &m.GID, &m.GName, &m.GCName); err != nil {
			log.Error("d.ListModuleGroup rows.Scan error(%v)", err)
			return
		}
		res = append(res, m)
	}
	err = rows.Err()
	return
}

// ListModuleUngroup list ungrouped modules
func (d *Dao) ListModuleUngroup(c context.Context, appKey string) (res []*mdlmdl.Module, err error) {
	rows, err := d.db.Query(c, _listModuleUngroup, appKey)
	if err != nil {
		log.Error("d.ListModuleUngroup d.db.Query(%v) error(%v)", appKey, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var m = &mdlmdl.Module{}
		if err = rows.Scan(&m.MID, &m.MName, &m.MCName); err != nil {
			log.Error("d.ListModuleUngroup rows.Scan error(%v)", err)
			return
		}
		res = append(res, m)
	}
	err = rows.Err()
	return
}

// ListEmptyGroups list empty groups
func (d *Dao) ListEmptyGroups(c context.Context, appKey string) (res []*mdlmdl.Group, err error) {
	rows, err := d.db.Query(c, _listEmptyGroup, appKey)
	if err != nil {
		log.Error("d.ListEmptyGroup d.db.Query(%v) error(%v)", appKey, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var g = &mdlmdl.Group{}
		if err = rows.Scan(&g.GID, &g.GName, &g.GCName); err != nil {
			log.Error("d.ListEmptyGroup rows.Scan error(%v)", err)
			return
		}
		g.Modules = []*mdlmdl.Module{}
		res = append(res, g)
	}
	err = rows.Err()
	return
}

// ListAllGroups list all groups
func (d *Dao) ListAllGroups(c context.Context, appKey string) (res []*mdlmdl.Group, err error) {
	rows, err := d.db.Query(c, _listAllGroup, appKey)
	if err != nil {
		log.Error("d.ListAllGroups d.db.Query(%v) error(%v)", appKey, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var g = &mdlmdl.Group{}
		if err = rows.Scan(&g.GID, &g.GName, &g.GCName); err != nil {
			log.Error("d.ListAllGroups rows.Scan error(%v)", err)
			return
		}
		res = append(res, g)
	}
	err = rows.Err()
	return
}

// TxAddModule add gmodule.
func (d *Dao) TxAddModule(tx *xsql.Tx, appKey, name string) (itemID int64, err error) {
	res, err := tx.Exec(_addModule, appKey, name)
	if err != nil {
		log.Error("TxAddModule %v", err)
		return
	}
	return res.LastInsertId()
}

// GetModuleID get group id by app key & name.
func (d *Dao) GetModuleID(c context.Context, appKey, name string) (moduleID int64, err error) {
	res := d.db.QueryRow(c, _getModuleIDByName, appKey, name)
	if err = res.Scan(&moduleID); err != nil {
		if err != sql.ErrNoRows {
			log.Error("GetModuleID %v", err)
		}
	}
	return
}

// ExistsRelation Search if there is a relation.
func (d *Dao) ExistsRelation(c context.Context, moduleID int64) (count int, err error) {
	row := d.db.QueryRow(c, _existsRelation, moduleID)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("AppInfo %v", err)
		}
	}
	return
}

// ExistsRelationModuleIDs Search modules_ids there is exist relation
func (d *Dao) ExistsRelationModuleIDs(c context.Context, mdlIDs []int64) (extMdlIDs []int64, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, mdlID := range mdlIDs {
		sqls = append(sqls, "?")
		args = append(args, mdlID)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_existsRelationModuleIDs, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("d.ExistsRelationModuleIDs d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var mID int64
		if err = rows.Scan(&mID); err != nil {
			log.Error("d.ExistsRelationModuleIDs rows.Scan error(%v)", err)
			return
		}
		extMdlIDs = append(extMdlIDs, mID)
	}
	err = rows.Err()
	return
}

// TxSetModuleGroupRalation add a module to a group.
func (d *Dao) TxSetModuleGroupRalation(tx *xsql.Tx, moduleID, groupID int64, operator string) (r int64, err error) {
	res, err := tx.Exec(_setModuleGroupRelation, moduleID, groupID, operator)
	if err != nil {
		log.Error("TxAddToGroup %v", err)
		return
	}
	return res.RowsAffected()
}

// TxAddModuleSize add module size.
func (d *Dao) TxAddModuleSize(tx *xsql.Tx, appKey string, buildID int64, moduleID int64, libVer, sizeType string, size int64) (r int64, err error) {
	res, err := tx.Exec(_addModuleSize, appKey, buildID, moduleID, libVer, sizeType, size)
	if err != nil {
		log.Error("TxAddModuleSize %v", err)
		return
	}
	return res.RowsAffected()
}

// ListModuleSize list module size.
func (d *Dao) ListModuleSize(c context.Context, appKey, moduleName, sizeType string) (res []*mdlmdl.ModuleSize, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey, moduleName)
	if sizeType != "" {
		args = append(args, sizeType)
		sqlAdd += " AND s.size_type=?"
	} else {
		sqlAdd += " AND s.size_type IN ('res', 'code')"
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_moduleSize, sqlAdd), args...)
	if err != nil {
		log.Error("d.ListModuleSize d.db.Query(%v) error(%v)", appKey, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ms = &mdlmdl.ModuleSize{}
		if err = rows.Scan(&ms.ID, &ms.Name, &ms.CName, &ms.LibVer, &ms.BuildID, &ms.Size, &ms.PackVersion, &ms.VersionCode); err != nil {
			log.Error("d.ListModuleSize rows.Scan error(%v)", err)
			return
		}
		res = append(res, ms)
	}
	err = rows.Err()
	return
}

// ListGroupSize list group size.
func (d *Dao) ListGroupSize(c context.Context, appKey, groupName, sizeType string, verCodes []int64, resRatio, codeRatio, xcassetsRatio float64) (res []*mdlmdl.GroupSize, err error) {
	var (
		sqlSum, sqlAdd string
		args           []interface{}
		inVersions     []string
	)
	args = append(args, appKey, groupName)
	if sizeType != "" {
		sqlSum += " SUM(s.size)"
		args = append(args, sizeType)
		sqlAdd += " AND s.size_type=?"
	} else {
		sqlSum += fmt.Sprintf(" FLOOR ( %.2f * SUM(IF( s.size_type = 'res', s.size, 0 )) + %.2f * sum(IF( s.size_type = 'code', s.size, 0 )) + %.2f * sum(IF( s.size_type = 'xcassets', s.size, 0 )) )", resRatio, codeRatio, xcassetsRatio)
		sqlAdd += " AND s.size_type IN ('res', 'code', 'xcassets')"
	}
	for _, verCode := range verCodes {
		inVersions = append(inVersions, strconv.FormatInt(verCode, 10))
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_groupSize, sqlSum, sqlAdd, strings.Join(inVersions, ",")), args...)
	if err != nil {
		log.Error("d.ListGroupSize d.db.Query(%v,%v,%v) error(%v)", appKey, groupName, sizeType, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var gs = &mdlmdl.GroupSize{}
		if err = rows.Scan(&gs.Size, &gs.BuildID, &gs.GID, &gs.GName, &gs.GCName, &gs.PackVersion, &gs.VersionCode); err != nil {
			log.Error("d.ListGroupSize rows.Scan error(%v)", err)
			return
		}
		res = append(res, gs)
	}
	err = rows.Err()
	return
}

// ListModuleSizeInGroup list module size in group.
func (d *Dao) ListModuleSizeInGroup(c context.Context, appKey, groupName, sizeType string, buildID int64, resRatio, codeRatio, xcassetsRatio float64) (res []*mdlmdl.ModuleGroupSize, err error) {
	var (
		sqlSum, sqlAdd string
		args           []interface{}
	)
	args = append(args, appKey, groupName, buildID)
	if sizeType != "" {
		args = append(args, sizeType)
		sqlSum += " SUM(s.size)"
		sqlAdd += " AND s.size_type=?"
	} else {
		sqlSum += fmt.Sprintf(" FLOOR ( %.2f * SUM(IF( s.size_type = 'res', s.size, 0 )) + %.2f * sum(IF( s.size_type = 'code', s.size, 0 )) + %.2f * sum(IF( s.size_type = 'xcassets', s.size, 0 )) )", resRatio, codeRatio, xcassetsRatio)
		sqlAdd += " AND s.size_type IN ('res', 'code', 'xcassets')"
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_moduleSizeInGroup, sqlSum, sqlAdd), args...)
	if err != nil {
		log.Error("d.ListModuleSizeInGroup d.db.Query(%v,%v,%v,%v) error(%v)", appKey, groupName, sizeType, buildID, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var mgs = &mdlmdl.ModuleGroupSize{}
		if err = rows.Scan(&mgs.MID, &mgs.MName, &mgs.MCName, &mgs.LibVer, &mgs.Size); err != nil {
			log.Error("d.ListModuleSizeInGroup rows.Scan error(%v)", err)
			return
		}
		res = append(res, mgs)
	}
	err = rows.Err()
	return
}

// ListGroupSizeInBuild list groups' size in a build
func (d *Dao) ListGroupSizeInBuild(c context.Context, appKey, sizeType string, buildID int64, resRatio, codeRatio, xcassetsRatio float64) (res []*mdlmdl.GroupSizeInBuildRes, err error) {
	var (
		sqlSum, sqlAdd string
		args           []interface{}
	)
	args = append(args, appKey, buildID)
	if sizeType != "" {
		args = append(args, sizeType)
		sqlSum += " SUM(s.size)"
		sqlAdd += " AND s.size_type=?"
	} else {
		sqlSum += fmt.Sprintf(" FLOOR ( %.2f * SUM(IF( s.size_type = 'res', s.size, 0 )) + %.2f * sum(IF( s.size_type = 'code', s.size, 0 )) + %.2f * sum(IF( s.size_type = 'xcassets', s.size, 0 )) )", resRatio, codeRatio, xcassetsRatio)
		sqlAdd += " AND s.size_type IN ('res', 'code', 'xcassets')"
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_groupSizeInBuild, sqlSum, sqlAdd), args...)
	if err != nil {
		log.Error("d.ListGroupSizeInBuild d.db.Query(%v,%v,%v) error(%v)", appKey, sizeType, buildID, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var gs = &mdlmdl.GroupSizeInBuildRes{}
		if err = rows.Scan(&gs.Size, &gs.GID, &gs.GName, &gs.GCName); err != nil {
			log.Error("d.ListGroupSizeInBuild rows.Scan error(%v)", err)
			return
		}
		res = append(res, gs)
	}
	err = rows.Err()
	return
}

// GetPackBuildID get pack build id.
func (d *Dao) GetPackBuildID(c context.Context, buildPackID int64) (buildID int64, err error) {
	res := d.db.QueryRow(c, _getPackBuildID, buildPackID)
	if err = res.Scan(&buildID); err != nil {
		log.Error("GetPackBuildID %v error(%v)", buildPackID, err)
		return
	}
	return
}

// ListSizeTypes list all size types for an app.
func (d *Dao) ListSizeTypes(c context.Context, appKey string, verCodes []int64) (res []string, err error) {
	var inVersions []string
	for _, verCode := range verCodes {
		inVersions = append(inVersions, strconv.FormatInt(verCode, 10))
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_listSizeType, strings.Join(inVersions, ",")), appKey)
	if err != nil {
		log.Error("d.ListSizeTypes d.db.Query(%v) error(%v)", appKey, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var sizeType string
		if err = rows.Scan(&sizeType); err != nil {
			log.Error("d.ListSizeTypes rows.Scan error(%v)", err)
			return
		}
		res = append(res, sizeType)
	}
	err = rows.Err()
	return
}

// LatestVersions list latest version codes.
func (d *Dao) LatestVersions(c context.Context, appKey string, limit int) (verCodes []int64, err error) {
	rows, err := d.db.Query(c, _listLatestVers, appKey, limit)
	if err != nil {
		log.Error("d.LatestVersions d.db.Query(%v) error(%v)", appKey, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var verCode int64
		if err = rows.Scan(&verCode); err != nil {
			log.Error("d.LatestVersions rows.Scan error(%v)", err)
			return
		}
		verCodes = append(verCodes, verCode)
	}
	err = rows.Err()
	return
}

// TxModulesConfTotalSizeSet set module config totalsize
func (d *Dao) TxModulesConfTotalSizeSet(tx *xsql.Tx, appKey, version, operator string, moduleGroupIDList []int64, totalSize int64) (err error) {
	for _, moduleGroupID := range moduleGroupIDList {
		if _, err = tx.Exec(_setModuleConfigTotalSize, appKey, version, moduleGroupID, totalSize, operator, totalSize, totalSize, operator); err != nil {
			log.Error("TxModulesConfTotalSizeSet %v", err)
			return
		}
	}
	return
}

// TxModulesConfSet set module config
func (d *Dao) TxModulesConfSet(tx *xsql.Tx, appKey, version, description, operator string, percentage float64, moduleGroupID, totalSize, fixedSize, applyNormalSize, applyForceSize, externalSize int64) (err error) {
	_, err = tx.Exec(_setModuleConfig, appKey, version, moduleGroupID, totalSize, percentage, fixedSize, applyNormalSize, applyForceSize, externalSize, description, operator, totalSize, percentage, fixedSize, applyNormalSize, applyForceSize, externalSize, description, operator)
	if err != nil {
		log.Error("TxModulesConfSet tx.Exec error(%v)", err)
	}
	return
}

// GetModulesConf get module config
func (d *Dao) GetModulesConf(c context.Context, appKey, version string) (res []*mdlmdl.ModuleConfig, err error) {
	rows, err := d.db.Query(c, _getModuleConfig, appKey, version)
	if err != nil {
		log.Error("Dao GetModulesConf: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mdlmdl.ModuleConfig{}
		if err = rows.Scan(&re.AppKey, &re.Version, &re.ModuleGroupID, &re.TotalSize, &re.Percentage, &re.FixedSize, &re.ApplyNormalSize,
			&re.ApplyForceSize, &re.ExternalSize, &re.Description, &re.OPERATOR); err != nil {
			log.Error("GetModulesConf Scan: %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// GetPreciousVersion get precious version
func (d *Dao) GetNewestModulesConfVersion(c context.Context, appKey string) (newestVersion string, err error) {
	row := d.db.QueryRow(c, _getNewestModulesConfVersion, appKey)
	if err = row.Scan(&newestVersion); err != nil {
		if err == sql.ErrNoRows {
			newestVersion = ""
			err = nil
		} else {
			log.Error("GetNewestModulesConfVersion %v", err)
		}
	}
	return
}

// GetPreciousVersion get precious version
func (d *Dao) GetPreciousVersion(c context.Context, appKey, version string) (preVersion string, err error) {
	row := d.db.QueryRow(c, _getPreciousVersion, appKey, version)
	if err = row.Scan(&preVersion); err != nil {
		if err == sql.ErrNoRows {
			preVersion = ""
			err = nil
		} else {
			log.Error("GetPreciousVersion %v", err)
		}
	}
	return
}

// GetPreciousVersion get precious version
func (d *Dao) GetMaxVersionCodeByTime(c context.Context, appKey string, endTime time.Time, count int) (maxVersionCode string, err error) {
	row := d.db.QueryRow(c, _getNewestVersionByTime, appKey, endTime, count)
	if err = row.Scan(&maxVersionCode); err != nil {
		if err == sql.ErrNoRows {
			maxVersionCode = ""
			err = nil
		} else {
			log.Error("GetPreciousVersion %v", err)
		}
	}
	return
}
