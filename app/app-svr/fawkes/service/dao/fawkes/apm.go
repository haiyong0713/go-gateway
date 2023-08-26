package fawkes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/database/xsql"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/model/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

//go:generate ../sqlgenerate/gensql -filter _eventSampleList

const (
	_busByID      = `SELECT id,name,app_keys,description,owner,datacenter_bus_key,datacenter_dwd_table_names,is_shared,operator,unix_timestamp(ctime),unix_timestamp(mtime) FROM apm_bus WHERE id=?`
	_busListCount = `SELECT count(*) FROM apm_bus %s`
	_busList      = `SELECT id,name,app_keys,description,owner,datacenter_bus_key,datacenter_dwd_table_names,is_shared,operator,unix_timestamp(ctime),unix_timestamp(mtime) FROM apm_bus %s`
	_addBus       = `INSERT INTO apm_bus (name,app_keys,description,owner,datacenter_bus_key,is_shared,operator,datacenter_dwd_table_names) VALUES (?,?,?,?,?,?,?,?)`
	_upBus        = "UPDATE apm_bus SET name=?,app_keys=?,description=?,owner=?,datacenter_bus_key=?,is_shared=?,operator=?,datacenter_dwd_table_names=? WHERE id=?"
	_delBus       = "DELETE FROM apm_bus WHERE id=?"

	_event          = `SELECT e.id,e.app_keys,e.bus_id,e.db_name,e.table_name,e.distributed_table_name,e.name,e.description,e.owner,e.operator,e.is_shared,unix_timestamp(e.ctime),unix_timestamp(e.mtime),e.log_id,e.is_activity,e.kafka_topic,e.sample_rate,e.datacenter_event_id,e.datacenter_app_id,b.name,e.state,e.data_count,e.level,e.datacenter_dwd_table_name,e.is_wide_table,e.storage_count,e.storage_capacity,e.lowest_sample_rate FROM (apm_event as e INNER JOIN apm_bus as b ON e.bus_id=b.id) WHERE e.id=?`
	_eventByIds     = `SELECT e.id,e.app_keys,e.bus_id,e.db_name,e.table_name,e.distributed_table_name,e.name,e.description,e.owner,e.operator,e.is_shared,unix_timestamp(e.ctime),unix_timestamp(e.mtime),e.log_id,e.is_activity,e.kafka_topic,e.sample_rate,e.datacenter_event_id,e.datacenter_app_id,b.name,e.state,e.data_count,e.level,e.datacenter_dwd_table_name,e.is_wide_table,e.storage_count,e.storage_capacity,e.lowest_sample_rate FROM (apm_event as e INNER JOIN apm_bus as b ON e.bus_id=b.id) WHERE e.id IN (%s)`
	_eventByName    = `SELECT id,app_keys,bus_id,db_name,table_name,distributed_table_name,name,description,owner,operator,is_shared,unix_timestamp(ctime),unix_timestamp(mtime),log_id,is_activity,kafka_topic,sample_rate,datacenter_event_id,datacenter_app_id,state,data_count,level,datacenter_dwd_table_name,is_wide_table,storage_count,storage_capacity,lowest_sample_rate FROM apm_event WHERE name=?`
	_eventListCount = `SELECT COUNT(*) FROM apm_event as e, apm_bus as b WHERE e.bus_id=b.id %s`
	_eventList      = `SELECT e.id,e.app_keys,e.bus_id,e.db_name,e.table_name,e.distributed_table_name,e.name,e.description,e.owner,e.operator,e.is_shared,unix_timestamp(e.ctime),unix_timestamp(e.mtime),e.log_id,e.is_activity,e.kafka_topic,b.name,e.sample_rate,e.datacenter_event_id,e.datacenter_app_id,e.state,e.data_count,e.level,e.datacenter_dwd_table_name,e.is_wide_table,e.storage_count,e.storage_capacity,e.lowest_sample_rate FROM (apm_event as e INNER JOIN apm_bus as b ON e.bus_id=b.id) %s`
	_addEvent       = `INSERT INTO apm_event (name,app_keys,description,owner,operator,is_shared,bus_id,log_id,db_name,table_name,distributed_table_name,is_activity,kafka_topic,sample_rate,datacenter_event_id,datacenter_app_id,level,data_count,datacenter_dwd_table_name,is_wide_table,lowest_sample_rate) VALUES (?,?,?,?,?,?,?,?,?,?,?,-1,?,?,?,?,?,?,?,?,?)`
	_upEvent        = `UPDATE apm_event SET app_keys=?,description=?,owner=?,operator=?,is_shared=?,log_id=?,db_name=?,table_name=?,distributed_table_name=?,kafka_topic=?,sample_rate=?,datacenter_app_id=?,bus_id=?,name=?,datacenter_event_id=?,state=?,level=?,data_count=?,datacenter_dwd_table_name=?,is_wide_table=?,lowest_sample_rate=? %s`
	_delEvent       = `DELETE FROM apm_event WHERE id =?`
	_upEventStorage = `INSERT INTO apm_event (id,storage_count,storage_capacity) VALUES %s ON DUPLICATE KEY UPDATE storage_count=VALUES(storage_count),storage_capacity=VALUES(storage_capacity)`

	// event with appId
	_eventListCountWithAppId = `SELECT COUNT(*) FROM (apm_event_datacenter_relation as r INNER JOIN apm_event as e ON r.event_id=e.id) INNER JOIN apm_bus as b ON b.id=e.bus_id WHERE r.datacenter_app_id=? %s`
	_eventListWithAppId      = `SELECT e.id,e.app_keys,e.bus_id,e.db_name,e.table_name,e.distributed_table_name,e.name,e.description,e.owner,e.operator,e.is_shared,unix_timestamp(e.ctime),unix_timestamp(e.mtime),e.log_id,e.is_activity,e.kafka_topic,b.name,e.sample_rate,e.datacenter_event_id,e.datacenter_app_id,e.state,e.data_count,e.level,e.datacenter_dwd_table_name,e.storage_count,e.storage_capacity,e.lowest_sample_rate FROM (apm_event_datacenter_relation as r INNER JOIN apm_event as e ON r.event_id=e.id) INNER JOIN apm_bus as b ON b.id=e.bus_id WHERE r.datacenter_app_id=? %s`
	_appEventRelAdd          = `INSERT INTO apm_event_datacenter_relation (event_id,datacenter_app_id,datacenter_event_id,operator) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE event_id=VALUES(event_id),datacenter_app_id=VALUES(datacenter_app_id),datacenter_event_id=VALUES(datacenter_event_id),operator=VALUES(operator)`
	_appEventRelList         = `SELECT event_id,datacenter_app_id,datacenter_event_id,operator FROM apm_event_datacenter_relation WHERE event_id=? %s`

	_eventFieldList             = `SELECT id,event_id,field_key,example,description,field_type,type,default_value,field_index,state,operator,unix_timestamp(ctime),unix_timestamp(mtime),is_clickhouse,is_elasticsearch_index,elasticsearch_field_type FROM apm_event_field WHERE event_id=?`
	_addEventField              = `INSERT INTO apm_event_field (event_id,field_key,example,description,field_type,default_value,state,type,field_index,operator,is_clickhouse,is_elasticsearch_index,elasticsearch_field_type) VALUES %s`
	_updateEventField           = `UPDATE apm_event_field SET example=?,description=?,default_value=?,operator=?,field_type=?,is_clickhouse=?,is_elasticsearch_index=?,field_index=?,type=?,state=?,elasticsearch_field_type=? WHERE id=?`
	_delEventField              = `DELETE FROM apm_event_field WHERE event_id=? %s`
	_delEventFieldById          = `DELETE FROM apm_event_field WHERE id=?`
	_updateEventFieldState      = `UPDATE apm_event_field SET state=? WHERE id=?`
	_upEventFieldStateByEventId = `UPDATE apm_event_field SET state=? WHERE event_id=?`
	_eventFieldModifyCount      = `SELECT event_id,COUNT(*) FROM apm_event_field WHERE state<>3 GROUP BY event_id`

	// Event Advanced
	_eventAdvancedCount  = `SELECT  COUNT(*) FROM apm_event_advanced WHERE event_id=?`
	_eventAdvancedList   = `SELECT id,event_id,field_name,title,description,display_type,query_type,mapping_group,custom_sql,operator,ctime,mtime FROM apm_event_advanced WHERE event_id=?`
	_eventAdvancedAdd    = `INSERT INTO apm_event_advanced (event_id,field_name,title,description,display_type,query_type,mapping_group,custom_sql,operator) VALUES (?,?,?,?,?,?,?,?,?)`
	_eventAdvancedDel    = `DELETE FROM apm_event_advanced WHERE id=?`
	_eventAdvancedUpdate = `UPDATE apm_event_advanced SET title=?,description=?,display_type=?,query_type=?,mapping_group=?,custom_sql=?,operator=? WHERE id=?`

	// Event Storage
	_eventStorageList = `SELECT event_id,datacenter_app_id,cnt,part_real_size FROM apm_event_technology_storage WHERE log_date=?`

	// 业务组
	_commandGroupListCount = `SELECT count(*) FROM apm_command_group WHERE app_key=? AND event_id=? %s`
	_commandGroupList      = `SELECT a.id,a.app_key,a.bus_id,a.event_id,a.name,a.description,a.operator,b.name,unix_timestamp(a.ctime),unix_timestamp(a.mtime) FROM apm_command_group as a, apm_bus as b WHERE a.bus_id=b.id AND app_key=? AND event_id=? %s`
	_commandGroupAdd       = `INSERT INTO apm_command_group (app_key,bus_id,event_id,name,description,operator) VALUES (?,?,?,?,?,?)`
	_commandGroupDel       = `DELETE FROM apm_command_group WHERE app_key=? AND id =? AND event_id=?`
	_commandGroupUpdate    = `UPDATE apm_command_group SET description=?,operator=? WHERE app_key=? AND id=? AND event_id=?`
	_commandGroupByBusID   = `SELECT id,app_key,bus_id,event_id,name,description,operator,unix_timestamp(ctime),unix_timestamp(mtime) FROM apm_command_group WHERE app_key=? AND event_id=? AND bus_id=?`
	_commandGroupByGroupID = `SELECT id,app_key,bus_id,event_id,name,description,operator,unix_timestamp(ctime),unix_timestamp(mtime) FROM apm_command_group WHERE app_key=? AND event_id=? AND id=?`

	// 业务项
	_commandList = `SELECT a.id,a.app_key,a.group_id,a.command,a.operator,unix_timestamp(a.ctime),unix_timestamp(a.mtime) FROM apm_command as a, apm_command_group as b WHERE a.group_id=b.id AND a.app_key=? AND b.event_id=? %s`
	_commandAdd  = `INSERT INTO apm_command (app_key,group_id,command,operator) VALUES %s`
	_commandDel  = `DELETE FROM apm_command WHERE app_key=? AND id=?`
	_commandsDel = `DELETE FROM apm_command WHERE app_key=? AND group_id=?`

	// 事件组高级配置项
	_commandGroupAdvancedList   = `SELECT id,app_key,event_id,field_name,title,description,display_type,query_type,mapping,operator,unix_timestamp(ctime),unix_timestamp(mtime) FROM apm_command_group_advanced WHERE app_key=? AND event_id=? AND group_id=?`
	_commandGroupAdvancedAdd    = `INSERT INTO apm_command_group_advanced (app_key,event_id,group_id,field_name,title,description,display_type,mapping,query_type,operator) VALUES (?,?,?,?,?,?,?,?,?,?)`
	_commandGroupAdvancedUpdate = `UPDATE apm_command_group_advanced SET title=?,description=?,display_type=?,query_type=?,mapping=?,operator=? WHERE app_key=? AND event_id=? AND group_id=? AND id=? `
	_commandGroupAdvancedDel    = `DELETE FROM apm_command_group_advanced WHERE app_key=? AND event_id=? AND group_id=? AND id=? `

	// 流量图理由别名
	_apmFlowmapRouteAliasList   = `SELECT id,app_key,bus_id,route_name,route_alias,unix_timestamp(ctime),unix_timestamp(mtime),operator FROM apm_route_alias WHERE app_key=? AND state=1 %s ORDER BY id DESC`
	_apmFlowMapRouteAliasAdd    = `INSERT INTO apm_route_alias (app_key, route_name, route_alias, bus_id, operator) VALUES (?,?,?,?,?)`
	_apmFlowMapRouteAliasUpdate = `UPDATE apm_route_alias SET route_name=?,route_alias=?,bus_id=?,operator=? WHERE id=?`
	_apmFlowMapRouteAliasDel    = `UPDATE apm_route_alias SET operator=?,state=0 WHERE id=?`

	// APM 应用配置表
	_apmEventSettiing = `SELECT id,app_key,event_id,sample_desc,sample_conf_key,unix_timestamp(ctime),unix_timestamp(mtime) FROM apm_event_setting WHERE app_key=? AND event_id=?`

	// APM Metric
	_apmMetricCount                  = `SELECT COUNT(*) FROM apm_ck_protheusme_metric as m,apm_bus as b WHERE m.bus_id=b.id %s`
	_apmMetricList                   = `SELECT m.id,m.metric,m.metric_type,m.exec_sql,m.labeled_keys,m.value_key,m.timestamp_key,m.description,m.apm_database_name,m.apm_table_name,m.time_filter,m.time_offset,m.state,m.status,m.url,m.bus_id,b.name,m.operator,m.ctime,m.mtime FROM apm_ck_protheusme_metric as m,apm_bus as b WHERE m.bus_id=b.id %s`
	_apmMetricByMetric               = `SELECT id,metric,metric_type,exec_sql,labeled_keys,value_key,timestamp_key,description,apm_database_name,apm_table_name,time_filter,time_offset,operator,ctime,mtime,state,status,url,bus_id FROM apm_ck_protheusme_metric WHERE metric=?`
	_apmMetricAdd                    = `INSERT INTO apm_ck_protheusme_metric(metric,metric_type,exec_sql,labeled_keys,value_key,timestamp_key,description,apm_database_name,apm_table_name,time_filter,time_offset,operator,state,status,url,bus_id) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,1,?,?,?)`
	_apmMetricUpdate                 = `UPDATE apm_ck_protheusme_metric SET metric_type=?,exec_sql=?,labeled_keys=?,value_key=?,timestamp_key=?,description=?,apm_database_name=?,apm_table_name=?,time_filter=?,time_offset=?,operator=?,state=?,status=?,url=?,bus_id=? WHERE metric=?`
	_apmMetricDelByUpdate            = `UPDATE apm_ck_protheusme_metric SET state=? WHERE metric=?`
	_apmMetricDel                    = `DELETE FROM apm_ck_protheusme_metric WHERE metric=?`
	_apmMetricPublish                = `INSERT INTO apm_ck_protheusme_metric_publish (md5,local_path,description,is_active_version,operator) VALUES (?,?,?,1,?)`
	_apmMetricPublishCount           = `SELECT COUNT(*) FROM apm_ck_protheusme_metric_publish %s`
	_apmMetricPublishList            = `SELECT id,md5,local_path,description,is_active_version,operator,ctime,mtime FROM apm_ck_protheusme_metric_publish %s`
	_apmMetricPublishById            = `SELECT id,md5,local_path,description,is_active_version,operator,ctime,mtime FROM apm_ck_protheusme_metric_publish WHERE id=?`
	_apmMetricPublishActive          = `SELECT id,md5,local_path,description,is_active_version,operator,ctime,mtime FROM apm_ck_protheusme_metric_publish WHERE is_active_version=1`
	_apmMetricPublishDiff            = `SELECT local_path FROM apm_ck_protheusme_metric_publish ORDER BY ctime DESC LIMIT 0,1`
	_apmMetricPublishStateUpdate     = `UPDATE apm_ck_protheusme_metric SET state=?`
	_apmMetricPublishActiveVerUpdate = `UPDATE apm_ck_protheusme_metric_publish SET is_active_version=? WHERE id=?`
	_apmMetricPublishDel             = `DELETE FROM apm_ck_protheusme_metric WHERE state=-1`

	// APM flink任务表
	_apmFlinkJobByID   = `SELECT id, log_id, name, description, owner, operator, state, ctime, mtime FROM apm_flink_job WHERE id=?`
	_apmFlinkJobList   = `SELECT id, log_id, name, description, owner, operator, state, ctime, mtime FROM apm_flink_job %s`
	_apmFlinkJobCount  = `SELECT COUNT(*) FROM apm_flink_job %s`
	_apmFlinkJobAdd    = `INSERT INTO apm_flink_job (log_id,name,description,owner,operator,state) VALUES (?,?,?,?,?,?)`
	_apmFlinkJobUpdate = `UPDATE apm_flink_job SET operator=? %s`
	_apmFlinkJobDel    = `DELETE FROM apm_flink_job WHERE id=?`

	// APM flink和event关系表
	_apmFlinkJobRelation            = `SELECT id, event_id, flink_job_id, operator, ctime, mtime, state FROM apm_event_flink_relation WHERE flink_job_id=? AND event_id=?`
	_apmFlinkJobRelationAdd         = `INSERT INTO apm_event_flink_relation (event_id,flink_job_id,operator,state) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE state=1`
	_apmFlinkJobRelationList        = `SELECT e.id,e.app_keys,e.bus_id,e.db_name,e.table_name,e.name,e.description,e.owner,e.is_shared,e.operator,e.state,unix_timestamp(e.ctime),unix_timestamp(e.mtime),e.kafka_topic,e.log_id,e.distributed_table_name,e.sample_rate,e.is_wide_table FROM (apm_event as e INNER JOIN apm_event_flink_relation as r ON e.id=r.event_id) INNER JOIN apm_flink_job as j ON j.id=r.flink_job_id WHERE j.id=? AND r.state>0 ORDER BY e.mtime DESC`
	_apmFlinkJobRelationStateUpdate = `UPDATE apm_event_flink_relation SET state=? WHERE flink_job_id=? %s`
	_apmFlinkJobRelationDel         = `DELETE FROM apm_event_flink_relation WHERE flink_job_id=? AND event_id=? AND state=1`
	_apmFlinkJobPublishCount        = `SELECT COUNT(*) FROM apm_flink_job_publish where flink_job_id=?`
	_apmFlinkJobPublishList         = `SELECT id,flink_job_id,md5,local_path,description,operator,ctime,mtime FROM apm_flink_job_publish WHERE flink_job_id=? %s`
	_apmFlinkJobPublish             = `INSERT INTO apm_flink_job_publish (flink_job_id,md5, local_path, description, operator) VALUES (?,?,?,?,?)`
	_apmFlinkJobLastPath            = `SELECT local_path FROM apm_flink_job_publish WHERE flink_job_id=? ORDER BY ctime DESC LIMIT 0,1`
	_apmFlinkJobPublishModifyCount  = `SELECT COUNT(*) FROM apm_event_flink_relation WHERE state<>3 AND flink_job_id=?`
	_apmFlinkJobPublishStateUpdate  = `UPDATE apm_event_flink_relation SET state=? WHERE state>0 AND flink_job_id=?`
	_apmFlinkJobPublishDel          = `DELETE FROM apm_event_flink_relation WHERE state=-1 AND flink_job_id=?`

	// APM crash rule
	_apmCrashRule       = `SELECT c.id,c.app_keys,c.bus_id,c.rule_name,c.keywords,c.page_keywords,c.operator,c.ctime,c.mtime,c.description,b.name FROM apm_crash_rule AS c,apm_bus AS b WHERE c.bus_id=b.id AND c.id=?`
	_apmCrashRuleCount  = `SELECT COUNT(*) FROM apm_crash_rule AS c,apm_bus as b WHERE c.bus_id=b.id %s`
	_apmCrashRuleList   = `SELECT c.id,c.app_keys,c.bus_id,c.rule_name,c.keywords,c.page_keywords,c.operator,c.ctime,c.mtime,c.description,b.name FROM apm_crash_rule AS c,apm_bus AS b WHERE c.bus_id=b.id %s`
	_apmCrashRuleAdd    = `INSERT INTO apm_crash_rule (app_keys,bus_id,rule_name,keywords,page_keywords,operator,description) VALUES (?,?,?,?,?,?,?)`
	_apmCrashRuleDel    = `DELETE FROM apm_crash_rule WHERE id=?`
	_apmCrashRuleUpdate = `UPDATE apm_crash_rule SET app_keys=?,bus_id=?,rule_name=?,keywords=?,page_keywords=?,operator=?,description=? WHERE id=?`

	// APM vada config
	_apmVedaConfig = `SELECT id,event_id,event_name,veda_db_name,veda_index_table,veda_stack_table,hash_column,unix_timestamp(ctime),unix_timestamp(mtime) FROM apm_event_veda_config WHERE event_id=?`

	// APM event field file
	_eventFieldFileAdd    = `INSERT INTO apm_event_field_file (event_id,field_id,field_key,example,field_type,field_index,description,type,default_value,is_clickhouse,is_elasticsearch_index,elasticsearch_field_type,field_state,field_version,operator) VALUES %s`
	_eventFieldFileLastFV = `SELECT field_version FROM apm_event_field_file WHERE event_id=? ORDER BY field_version DESC LIMIT 1`
	_eventFieldFileList   = `SELECT id,event_id,field_id,field_key,example,field_type,description,type,default_value,is_clickhouse,is_elasticsearch_index,field_state,field_index,field_version,operator,ctime,mtime,elasticsearch_field_type FROM apm_event_field_file WHERE event_id=? AND field_version=?`

	// APM event field publish
	_eventFieldPublishCount       = `SELECT COUNT(*) FROM apm_event_field_publish WHERE event_id=?`
	_eventFieldPublishAdd         = `INSERT INTO apm_event_field_publish (event_id,version,operator) VALUES (?,?,?)`
	_eventFieldPublishList        = `SELECT id,event_id,version,operator,ctime,mtime FROM apm_event_field_publish WHERE event_id=? ORDER BY version DESC LIMIT ?,?`
	_eventFieldPublishLastVersion = `SELECT version FROM apm_event_field_publish WHERE event_id=? AND version<? ORDER BY id DESC LIMIT 1`

	// APM event field type sync
	_eventFieldTypeSync = `UPDATE apm_event_field SET elasticsearch_field_type=? WHERE id=?`
	// APM app event common field group
	_eventCommonFieldGroupAdd    = `INSERT INTO apm_app_event_common_field_group (app_key,name,description,is_default,operator) VALUES (?,?,?,?,?)`
	_eventCommonFieldGroupById   = `SELECT id,app_key,name,description,is_default,operator,ctime,mtime FROM apm_app_event_common_field_group WHERE id=?`
	_eventCommonFieldGroupCount  = `SELECT COUNT(*) FROM apm_app_event_common_field_group WHERE app_key=?`
	_eventCommonFieldGroupList   = `SELECT id,app_key,name,description,is_default,operator,ctime,mtime FROM apm_app_event_common_field_group WHERE app_key=? LIMIT ?,?`
	_eventCommonFieldGroupUpdate = `UPDATE apm_app_event_common_field_group SET name=?,description=?,is_default=?,operator=? WHERE id=?`
	_eventCommonFieldGroupDel    = `DELETE FROM apm_app_event_common_field_group WHERE id=?`
	// APM app event  common field
	_eventCommonFieldAdd          = `INSERT INTO apm_app_event_common_field (app_key,group_id,field_key,field_type,field_index,description,default_value,is_clickhouse,is_elasticsearch_index,elasticsearch_field_type,operator) VALUES %s`
	_eventCommonFieldList         = `SELECT id,app_key,group_id,field_key,field_type,field_index,description,default_value,state,is_clickhouse,is_elasticsearch_index,elasticsearch_field_type,operator,ctime,mtime FROM apm_app_event_common_field WHERE group_id=?`
	_eventCommonFieldDel          = `DELETE FROM apm_app_event_common_field WHERE id=?`
	_eventCommonFieldDelByGroupId = `DELETE FROM apm_app_event_common_field WHERE group_id=?`
	_eventCommonFieldUpdate       = `UPDATE apm_app_event_common_field SET field_type=?,field_index=?,description=?,default_value=?,is_clickhouse=?,is_elasticsearch_index=?,elasticsearch_field_type=?,operator=? WHERE id=?`

	// APM event alert
	_eventAlertAdd    = `INSERT INTO apm_event_alert_rule (event_id,datacenter_app_id,billion_id,title,description,intervals,time_field,cluster,level,time_frame,agg_type,agg_field,agg_percentile,filter_query,denominator_filter_query,trigger_condition,group_field,notify_fields,notify_duration,channels,targets,bot_webhook,webhook,mute_type,mute_period,version,min_log_count,is_enable,is_log_detail,creator,operator) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_eventAlertCount  = `SELECT COUNT(*) FROM (apm_event_alert_rule as r INNER JOIN apm_event as e ON r.event_id=e.id) %s`
	_eventAlertList   = `SELECT r.id,r.event_id,r.datacenter_app_id,e.name as event_name,r.billion_id,r.title,r.description,r.intervals,r.time_field,r.cluster,r.level,r.time_frame,r.agg_type,r.agg_field,r.agg_percentile,r.filter_query,r.denominator_filter_query,r.trigger_condition,r.group_field,r.notify_fields,r.notify_duration,r.channels,r.targets,r.bot_webhook,r.webhook,r.mute_type,r.mute_period,r.version,r.min_log_count,r.is_enable,r.is_log_detail,r.creator,r.operator,unix_timestamp(r.ctime) as ctime,unix_timestamp(r.mtime) as mtime FROM (apm_event_alert_rule as r INNER JOIN apm_event as e ON r.event_id=e.id) %s`
	_eventAlertInfo   = `SELECT r.id,r.event_id,r.datacenter_app_id,e.name as event_name,r.billion_id,r.title,r.description,r.intervals,r.time_field,r.cluster,r.level,r.time_frame,r.agg_type,r.agg_field,r.agg_percentile,r.filter_query,r.denominator_filter_query,r.trigger_condition,r.group_field,r.notify_fields,r.notify_duration,r.channels,r.targets,r.bot_webhook,r.webhook,r.mute_type,r.mute_period,r.version,r.min_log_count,r.is_enable,r.is_log_detail,r.creator,r.operator,unix_timestamp(r.ctime) as ctime,unix_timestamp(r.mtime) as mtime FROM (apm_event_alert_rule as r INNER JOIN apm_event as e ON r.event_id=e.id) WHERE r.id=?`
	_eventAlertUpdate = `UPDATE apm_event_alert_rule SET datacenter_app_id=?,title=?,description=?,version=?,min_log_count=?,intervals=?,time_field=?,cluster=?,level=?,time_frame=?,agg_type=?,agg_field=?,agg_percentile=?,filter_query=?,denominator_filter_query=?,trigger_condition=?,group_field=?,notify_fields=?,notify_duration=?,channels=?,targets=?,bot_webhook=?,webhook=?,mute_type=?,mute_period=?,is_log_detail=?,operator=? WHERE id=?`
	_eventAlertDel    = `DELETE FROM apm_event_alert_rule WHERE id=?`
	_eventAlertSwitch = `UPDATE apm_event_alert_rule SET is_enable=? WHERE id=?`

	// APM event data completion
	_eventCompletionList = `SELECT datacenter_event_name,datacenter_app_id,log_date FROM apm_event_technology_completion WHERE log_date=?`

	// APM alert rule
	_apmAlertRuleByHawkeyeIds  = `SELECT id,hawkeye_id,name,trigger_condition,species,query_exprs,rule_type,IFNULL(markdown,'') as markdown,operator,unix_timestamp(ctime) as ctime,unix_timestamp(mtime) as mtime FROM apm_alert_rule WHERE hawkeye_id IN (%s)`
	_apmAlertRuleCount         = `SELECT COUNT(*) FROM apm_alert_rule %s`
	_apmAlertRuleList          = `SELECT id,hawkeye_id,name,trigger_condition,species,query_exprs,rule_type,IFNULL(markdown,'') as markdown,operator,unix_timestamp(ctime) as ctime,unix_timestamp(mtime) as mtime FROM apm_alert_rule %s`
	_apmAlertRuleAdd           = `INSERT INTO apm_alert_rule (hawkeye_id,name,trigger_condition,species,query_exprs,operator,rule_type) VALUES (?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE name=?,trigger_condition=?,species=?,query_exprs=?,operator=?,rule_type=?`
	_apmAlertRuleMDUpdate      = `UPDATE apm_alert_rule SET markdown=? WHERE id=?`
	_apmAlertRuleDel           = `DELETE FROM apm_alert_rule WHERE id=?`
	_apmAlertRuleRelAdd        = `INSERT INTO apm_alert_rule_relation (rule_id,adjust_rule_id,operator) VALUES (?,?,?) ON DUPLICATE KEY UPDATE rule_id=?,adjust_rule_id=?,operator=?`
	_apmAlertRuleRelByRuleIds  = `SELECT id,rule_id,adjust_rule_id,operator,unix_timestamp(ctime) as ctime,unix_timestamp(mtime) as mtime FROM apm_alert_rule_relation WHERE rule_id IN(%s)`
	_apmAlertRuleRelByAdjustId = `SELECT id,rule_id,adjust_rule_id,operator,unix_timestamp(ctime) as ctime,unix_timestamp(mtime) as mtime FROM apm_alert_rule_relation WHERE adjust_rule_id=?`

	// APM alert
	_apmAlert       = `SELECT id,rule_id,alert_md5,app_key,alert_type,alert_status,description,duration,labels,trigger_value,operator,start_time,ctime,mtime FROM apm_alert WHERE id=?`
	_apmAlertByMd5  = `SELECT id,rule_id,alert_md5,app_key,alert_type,alert_status,description,duration,labels,trigger_value,operator,start_time,ctime,mtime FROM apm_alert WHERE alert_md5=?`
	_apmAlertCount  = `SELECT COUNT(*) FROM apm_alert %s`
	_apmAlertList   = `SELECT id,rule_id,app_key,env,alert_md5,alert_type,alert_status,description,duration,labels,trigger_value,operator,start_time,ctime,mtime FROM apm_alert %s`
	_apmAlertAdd    = `INSERT INTO apm_alert (rule_id,app_key,env,alert_md5,alert_status,duration,labels,trigger_value,alert_type,operator,start_time) VALUES (?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE rule_id=?,app_key=?,env=?,alert_md5=?,alert_status=?,duration=?,labels=?,trigger_value=?,operator=?,start_time=?`
	_apmAlertUpdate = `UPDATE apm_alert SET alert_type=?,alert_status=?,description=?,operator=?,duration=? WHERE id=?`

	// APM alert reason config
	_apmAlertReasonConfig       = `SELECT id,rule_id,event_id,query_type,query_sql,query_condition,impact_factor_fields,description,operator,ctime,mtime FROM apm_alert_reason_config WHERE rule_id=?`
	_apmAlertReasonConfigAdd    = `INSERT INTO apm_alert_reason_config (rule_id,event_id,query_type,query_sql,query_condition,impact_factor_fields,description,operator) VALUES (?,?,?,?,?,?,?,?)`
	_apmAlertReasonConfigUpdate = `UPDATE apm_alert_reason_config SET event_id=?,query_type=?,query_sql=?,query_condition=?,impact_factor_fields=?,description=?,operator=? WHERE id=?`
	_apmAlertReasonConfigDelete = `DELETE FROM apm_alert_reason_config WHERE id=?`

	// 采样率
	_evenSampleRateAdd         = "INSERT INTO apm_event_sample_rate (datacenter_app_id,event_id,event_name,sample_rate,log_id) values (?,?,?,?,?) ON DUPLICATE KEY UPDATE sample_rate=VALUES(sample_rate),event_name=VALUES(event_name)"
	_evenSampleRateAppAdd      = "INSERT INTO apm_event_sample_rate_app (app_key,event_id,event_name,sample_rate,log_id) values (?,?,?,?,?) ON DUPLICATE KEY UPDATE sample_rate=VALUES(sample_rate),event_name=VALUES(event_name)"
	_eventSampleList           = "SELECT aa.app_key,er.datacenter_app_id,er.sample_rate,er.event_id,er.event_name, er.mtime,er.ctime,er.log_id FROM apm_event_sample_rate as er, app_attribute as aa WHERE er.datacenter_app_id = aa.datacenter_app_id AND aa.datacenter_app_id=? AND aa.app_key=? %s;"
	_eventSampleAppList        = "SELECT app_key,sample_rate,event_id,event_name,mtime,ctime,log_id FROM apm_event_sample_rate_app WHERE app_key=? %s;"
	_eventSampleBatchDelete    = "DELETE FROM apm_event_sample_rate WHERE (datacenter_app_id, event_id) IN (%s)"
	_eventSampleAppBatchDelete = "DELETE FROM apm_event_sample_rate_app WHERE (app_key, event_id) IN (%s)"

	// 埋点监测通知配置
	_eventMonitorNotifyConfig         = `SELECT id,event_id,app_key,is_notify,is_mute,mute_start_time,mute_end_time,operator,ctime,mtime FROM apm_event_monitor_notify_config WHERE event_id=? AND app_key=?`
	_eventMonitorNotifyConfigCount    = `SELECT COUNT(*) FROM apm_event_monitor_notify_config %s`
	_eventMonitorNotifyConfigList     = `SELECT id,event_id,app_key,is_notify,is_mute,mute_start_time,mute_end_time,operator,ctime,mtime FROM apm_event_monitor_notify_config %s`
	_eventMonitorNotifyConfigSet      = `INSERT INTO apm_event_monitor_notify_config (event_id,app_key,is_notify,is_mute,mute_start_time,mute_end_time,operator) VALUES (?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE is_notify=VALUES(is_notify),is_mute=VALUES(is_mute),mute_start_time=VALUES(mute_start_time),mute_end_time=VALUES(mute_end_time),operator=VALUES(operator)`
	_eventMonitorNotifyConfigBatchSet = `INSERT IGNORE INTO apm_event_monitor_notify_config (event_id,app_key,is_notify,is_mute,mute_start_time,mute_end_time,operator) VALUES %s`
	_eventMonitorNotifyMuteUpdate     = `UPDATE apm_event_monitor_notify_config SET is_mute=0 WHERE id IN (%s)`
)
const (
	_contentTypeJson  = "application/json; charset=utf-8"
	_billionsUsername = "fawkes"
	_esAppId          = "mobile"

	// 数据平台openapi
	_ckCreateApiName = "CreateTableWithCode"
	_metaAppId       = "datacenter.keeper.keeper"
	_ckGroupName     = "KeeperMultiTable"
	_ckRequestId     = "keeper_CreateTable"
)

func (d *Dao) TxApmAddEventField(tx *sql.Tx, eventID int64, eventFields []*apm.EventField, operator string) (r int64, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, field := range eventFields {
		sqls = append(sqls, "(?,?,?,?,?,?,?,?,?,?,?,?,?)")
		args = append(args, eventID, field.Key, field.Example, field.Description, field.Type, field.DefaultValue, apm.EventFieldStateAdd, apm.EventFieldExtendedType, field.Index, operator, field.IsClickhouse, field.ISElasticsearchIndex, field.ElasticSearchFieldType)
	}
	res, err := tx.Exec(fmt.Sprintf(_addEventField, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("TxApmAddEventField tx.Exec error(%v)", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) ApmBusListCount(c context.Context, appKeys, filterKey string) (r int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if appKeys != "" {
		args = append(args, appKeys)
		sqlAdd += "AND (FIND_IN_SET(?, app_keys) OR is_shared=1)"
	}
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey)
		sqlAdd += "AND (name LIKE ?)"
	}
	if len(sqlAdd) > 0 {
		sqlAdd = strings.Replace(sqlAdd, "AND", "WHERE", 1)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_busListCount, sqlAdd), args...)
	if err = row.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("d.ApmBusListCount row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmBusByID(c context.Context, id int64) (re *apm.Bus, err error) {
	row := d.db.QueryRow(c, _busByID, id)
	re = &apm.Bus{}
	if err = row.Scan(&re.ID, &re.Name, &re.AppKeys, &re.Description, &re.Owner, &re.DatacenterBusKey, &re.DatacenterDwdTableNames, &re.Shared, &re.Operator, &re.Ctime, &re.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("ApmBusByID row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmBusList(c context.Context, appKeys, filterKey string, ps, pn int) (res []*apm.Bus, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if appKeys != "" {
		args = append(args, appKeys)
		sqlAdd += "AND (FIND_IN_SET(?, app_keys) OR is_shared=1)"
	}
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey)
		sqlAdd = "AND (name LIKE ?)"
	}
	if len(sqlAdd) > 0 {
		sqlAdd = strings.Replace(sqlAdd, "AND", "WHERE", 1)
	}
	sqlAdd += " ORDER BY ctime DESC"
	args = append(args, (pn-1)*ps, ps)
	sqlAdd += " LIMIT ?,?"
	rows, err := d.db.Query(c, fmt.Sprintf(_busList, sqlAdd), args...)
	if err != nil {
		log.Error("ApmBusList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.Bus{}
		if err = rows.Scan(&re.ID, &re.Name, &re.AppKeys, &re.Description, &re.Owner, &re.DatacenterBusKey, &re.DatacenterDwdTableNames, &re.Shared, &re.Operator, &re.Ctime, &re.Mtime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmBusAdd(tx *sql.Tx, name, appKeys, description, owner, datacenterBusinessKey, userName, datacenterDwdTableNames string, shared int) (r int64, err error) {
	res, err := tx.Exec(_addBus, name, appKeys, description, owner, datacenterBusinessKey, shared, userName, datacenterDwdTableNames)
	if err != nil {
		log.Error("ApmBusAdd %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmBusUpdate(tx *sql.Tx, name, appKeys, description, owner, datacenterBusinessKey, userName, datacenterDwdTableNames string, busId int64, shared int) (r int64, err error) {
	res, err := tx.Exec(_upBus, name, appKeys, description, owner, datacenterBusinessKey, shared, userName, datacenterDwdTableNames, busId)
	if err != nil {
		log.Error("ApmBusUpdate %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmBusDel(tx *sql.Tx, busId int64) (r int64, err error) {
	res, err := tx.Exec(_delBus, busId)
	if err != nil {
		log.Error("ApmBusDel %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmEventAdd(tx *sql.Tx, name, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName string, level, isWideTable int8, shared, sampleRate int, busId, datacenterEventID, datacenterAppID, dataCount int64, lowestSampleRate float64) (r int64, err error) {
	res, err := tx.Exec(_addEvent, name, appKeys, description, owner, userName, shared, busId, logID, dbName, tableName, distributedTableName, topic, sampleRate, datacenterEventID, datacenterAppID, level, dataCount, dwdTableName, isWideTable, lowestSampleRate)
	if err != nil {
		log.Error("ApmEventAdd %v", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

func (d *Dao) TxApmEventDel(tx *sql.Tx, eventID int64) (r int64, err error) {
	res, err := tx.Exec(_delEvent, eventID)
	if err != nil {
		log.Error("ApmEventDel %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmEventUpdate(tx *sql.Tx, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, name, dwdTableName string, activity, state, level, isWideTable int8, shared, sampleRate int, eventId, datacenterAppID, busID, datacenterEventID, dataCount int64, lowestSampleRate float64) (r int64, err error) {
	var (
		args   []interface{}
		sqlAdd string
	)
	args = append(args, appKeys, description, owner, userName, shared, logID, dbName, tableName, distributedTableName, topic, sampleRate, datacenterAppID, busID, name, datacenterEventID, state, level, dataCount, dwdTableName, isWideTable, lowestSampleRate)
	if activity != 0 {
		args = append(args, activity)
		sqlAdd += ",is_activity=?"
	}
	args = append(args, eventId)
	sqlAdd += " WHERE id = ?"
	res, err := tx.Exec(fmt.Sprintf(_upEvent, sqlAdd), args...)
	if err != nil {
		log.Error("ApmEventUpdateBaseInfo %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) ApmEvent(c context.Context, id int64) (re *apm.Event, err error) {
	row := d.db.QueryRow(c, _event, id)
	re = &apm.Event{}
	if err = row.Scan(&re.ID, &re.AppKeys, &re.BusID, &re.Databases, &re.TableName, &re.DistributedTableName, &re.Name, &re.Description, &re.Owner, &re.Operator, &re.Shared, &re.Ctime, &re.Mtime, &re.LogID, &re.Activity, &re.Topic, &re.SampleRate, &re.DatacenterEventID, &re.DatacenterAppID, &re.BusName, &re.State, &re.DataCount, &re.Level, &re.DatacenterDwdTableName, &re.IsWideTable, &re.StorageCount, &re.StorageCapacity, &re.LowestSampleRate); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("%v", err)
		}
	}
	return
}

func (d *Dao) ApmEventByIds(c context.Context, ids []int64) (res map[int64]*apm.Event, err error) {
	if len(ids) < 1 {
		log.Warnc(c, "ids %v is empty", ids)
		return
	}
	var (
		sqls []string
		args []interface{}
	)
	for _, id := range ids {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_eventByIds, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*apm.Event)
	for rows.Next() {
		re := &apm.Event{}
		if err = rows.Scan(&re.ID, &re.AppKeys, &re.BusID, &re.Databases, &re.TableName, &re.DistributedTableName, &re.Name, &re.Description, &re.Owner, &re.Operator, &re.Shared, &re.Ctime, &re.Mtime, &re.LogID, &re.Activity, &re.Topic, &re.SampleRate, &re.DatacenterEventID, &re.DatacenterAppID, &re.BusName, &re.State, &re.DataCount, &re.Level, &re.DatacenterDwdTableName, &re.IsWideTable, &re.StorageCount, &re.StorageCapacity, &re.LowestSampleRate); err != nil {
			log.Error("ConfigVersionByIDs %v", err)
			return
		}
		res[re.ID] = re
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmEventByName(c context.Context, name string) (re *apm.Event, err error) {
	row := d.db.QueryRow(c, _eventByName, name)
	re = &apm.Event{}
	if err = row.Scan(&re.ID, &re.AppKeys, &re.BusID, &re.Databases, &re.TableName, &re.DistributedTableName, &re.Name, &re.Description, &re.Owner, &re.Operator, &re.Shared, &re.Ctime, &re.Mtime, &re.LogID, &re.Activity, &re.Topic, &re.SampleRate, &re.DatacenterEventID, &re.DatacenterAppID, &re.State, &re.DataCount, &re.Level, &re.DatacenterDwdTableName, &re.IsWideTable, &re.StorageCount, &re.StorageCapacity, &re.LowestSampleRate); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("%v", err)
		}
	}
	return
}

func (d *Dao) ApmEventListCount(c context.Context, name, appKeys, logID, busName, topic, dbName, tableName, distTabName, dwdTableName string, busId, appId int64, activity, dtCondition, state int8) (r int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if appKeys != "" {
		args = append(args, appKeys)
		sqlAdd += "AND (FIND_IN_SET(?, e.app_keys) OR e.is_shared=1)"
	}
	if busId != 0 {
		args = append(args, busId)
		sqlAdd += "AND (e.bus_id = ?)"
	}
	if logID != "" {
		args = append(args, logID)
		sqlAdd += "AND (e.log_id = ?)"
	}
	if busName != "" {
		args = append(args, busName)
		sqlAdd += "AND (b.name = ?)"
	}
	if name != "" {
		name = "%" + name + "%"
		args = append(args, name)
		sqlAdd += "AND (e.name LIKE ?)"
	}
	if topic != "" {
		topic = "%" + topic + "%"
		args = append(args, topic)
		sqlAdd += "AND (e.kafka_topic LIKE ?)"
	}
	if activity != 0 {
		args = append(args, activity)
		sqlAdd += "AND (e.is_activity = ?)"
	}
	if dtCondition != 0 {
		sqlAdd += "AND (e.db_name !='' AND e.table_name !='')"
	}
	if dbName != "" {
		args = append(args, dbName)
		sqlAdd += " AND (e.db_name = ?) "
	}
	if tableName != "" {
		args = append(args, tableName)
		sqlAdd += "AND (e.table_name = ?)"
	}
	if distTabName != "" {
		args = append(args, distTabName)
		sqlAdd += " AND (e.distributed_table_name = ?) "
	}
	if appId != 0 {
		args = append(args, appId)
		sqlAdd += "AND (e.datacenter_app_id = ?)"
	}
	if state != 0 {
		args = append(args, state)
		sqlAdd += "AND (e.state = ?)"
	}
	if dwdTableName != "" {
		args = append(args, dwdTableName)
		sqlAdd += "AND (e.datacenter_dwd_table_name = ?)"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_eventListCount, sqlAdd), args...)
	if err = row.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("d.ApmEventListCount row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmEventList(c context.Context, name, appKeys, logID, busName, topic, dbName, tableName, distTabName, orderBy, dwdTableName string, ps, pn int, busId, appId int64, activity, dtCondition, state int8) (res []*apm.Event, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if appKeys != "" {
		args = append(args, appKeys)
		sqlAdd += " AND (FIND_IN_SET(?, e.app_keys) OR e.is_shared=1) "
	}
	if busId != 0 {
		args = append(args, busId)
		sqlAdd += " AND (e.bus_id = ?) "
	}
	if logID != "" {
		args = append(args, logID)
		sqlAdd += " AND (e.log_id = ?) "
	}
	if busName != "" {
		args = append(args, busName)
		sqlAdd += " AND (b.name = ?) "
	}
	if name != "" {
		name = "%" + name + "%"
		args = append(args, name)
		sqlAdd += " AND (e.name LIKE ?) "
	}
	if topic != "" {
		topic = "%" + topic + "%"
		args = append(args, topic)
		sqlAdd += " AND (e.kafka_topic LIKE ?) "
	}
	if activity != 0 {
		args = append(args, activity)
		sqlAdd += " AND (e.is_activity = ?) "
	}
	if dtCondition != 0 {
		sqlAdd += " AND (e.db_name !='' AND e.table_name !='') "
	}
	if dbName != "" {
		args = append(args, dbName)
		sqlAdd += " AND (e.db_name = ?) "
	}
	if tableName != "" {
		args = append(args, tableName)
		sqlAdd += " AND (e.table_name = ?) "
	}
	if distTabName != "" {
		args = append(args, distTabName)
		sqlAdd += " AND (e.distributed_table_name = ?) "
	}
	if appId != 0 {
		args = append(args, appId)
		sqlAdd += " AND (e.datacenter_app_id = ?) "
	}
	if state != 0 {
		args = append(args, state)
		sqlAdd += " AND (e.state = ?) "
	}
	if dwdTableName != "" {
		args = append(args, dwdTableName)
		sqlAdd += " AND (e.datacenter_dwd_table_name = ?) "
	}
	if orderBy != "" {
		sqlAdd += fmt.Sprintf(" ORDER BY %v ", orderBy)
	} else {
		sqlAdd += " ORDER BY e.ctime DESC"
	}
	if pn != 0 && ps != 0 {
		args = append(args, (pn-1)*ps, ps)
		sqlAdd += " LIMIT ?,?"
	}
	sqlAdd = strings.Replace(sqlAdd, "AND", "WHERE", 1)
	rows, err := d.db.Query(c, fmt.Sprintf(_eventList, sqlAdd), args...)
	if err != nil {
		log.Error("ApmEventList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.Event{}
		if err = rows.Scan(&re.ID, &re.AppKeys, &re.BusID, &re.Databases, &re.TableName, &re.DistributedTableName, &re.Name, &re.Description, &re.Owner, &re.Operator, &re.Shared, &re.Ctime, &re.Mtime, &re.LogID, &re.Activity, &re.Topic, &re.BusName, &re.SampleRate, &re.DatacenterEventID, &re.DatacenterAppID, &re.State, &re.DataCount, &re.Level, &re.DatacenterDwdTableName, &re.IsWideTable, &re.StorageCount, &re.StorageCapacity, &re.LowestSampleRate); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmEventListCountWithAppId(c context.Context, name, busName, logId string, appId int64, state int8) (count int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appId)
	if name != "" {
		name = "%" + name + "%"
		args = append(args, name)
		sqlAdd += "AND (e.name LIKE ?)"
	}
	if busName != "" {
		args = append(args, busName)
		sqlAdd += "AND (b.name = ?)"
	}
	if logId != "" {
		args = append(args, logId)
		sqlAdd += "AND (e.log_id = ?)"
	}
	if state != 0 {
		args = append(args, state)
		sqlAdd += "AND (e.state = ?)"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_eventListCountWithAppId, sqlAdd), args...)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("d.ApmEventListCount row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmEventListWithAppId(c context.Context, name, busName, logId, orderBy string, appId int64, pn, ps int, state int8) (res []*apm.Event, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appId)
	if name != "" {
		name = "%" + name + "%"
		args = append(args, name)
		sqlAdd += "AND (e.name LIKE ?)"
	}
	if busName != "" {
		args = append(args, busName)
		sqlAdd += "AND (b.name = ?)"
	}
	if logId != "" {
		args = append(args, logId)
		sqlAdd += "AND (e.log_id = ?)"
	}
	if state != 0 {
		args = append(args, state)
		sqlAdd += "AND (e.state = ?)"
	}
	if orderBy != "" {
		sqlAdd += fmt.Sprintf(" ORDER BY %v ", orderBy)
	} else {
		sqlAdd += " ORDER BY e.ctime DESC"
	}
	args = append(args, (pn-1)*ps, ps)
	sqlAdd += " LIMIT ?,?"
	rows, err := d.db.Query(c, fmt.Sprintf(_eventListWithAppId, sqlAdd), args...)
	if err != nil {
		log.Error("ApmEventListByAppKey d.db.Query")
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.Event{}
		if err = rows.Scan(&re.ID, &re.AppKeys, &re.BusID, &re.Databases, &re.TableName, &re.DistributedTableName, &re.Name, &re.Description, &re.Owner, &re.Operator, &re.Shared, &re.Ctime, &re.Mtime, &re.LogID, &re.Activity, &re.Topic, &re.BusName, &re.SampleRate, &re.DatacenterEventID, &re.DatacenterAppID, &re.State, &re.DataCount, &re.Level, &re.DatacenterDwdTableName, &re.StorageCount, &re.StorageCapacity, &re.LowestSampleRate); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmEventFieldList(c context.Context, eventId int64) (res []*apm.EventField, err error) {
	rows, err := d.db.Query(c, _eventFieldList, eventId)
	if err != nil {
		log.Error("ApmEventFieldList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.EventField{}
		if err = rows.Scan(&re.ID, &re.EventID, &re.Key, &re.Example, &re.Description, &re.Type, &re.Mode, &re.DefaultValue, &re.Index, &re.State, &re.Operator, &re.Ctime, &re.Mtime, &re.IsClickhouse, &re.ISElasticsearchIndex, &re.ElasticSearchFieldType); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmEventFieldDelByEventID(tx *sql.Tx, eventId int64, state int8) (r int64, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, eventId)
	if state != 0 {
		sqlAdd += " AND state=? "
		args = append(args, state)
	}
	res, err := tx.Exec(fmt.Sprintf(_delEventField, sqlAdd), args...)
	if err != nil {
		log.Error("ApmEventFieldDelByEventID %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmEventFieldUpdate(tx *sql.Tx, example, description, defaultValue, operator string, isClickhouse, isElasticSearchIndex, state, elasticSearchFieldType int8, fieldIndex, id int64, fieldType, mode int8) (r int64, err error) {
	res, err := tx.Exec(_updateEventField, example, description, defaultValue, operator, fieldType, isClickhouse, isElasticSearchIndex, fieldIndex, mode, state, elasticSearchFieldType, id)
	if err != nil {
		log.Error("ApmEventFieldUpdate %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) ApmEventAdvancedCount(c context.Context, eventID int64) (r int64, err error) {
	row := d.db.QueryRow(c, _eventAdvancedCount, eventID)
	if err = row.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("ApmEventAdvancedCount row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmEventAdvancedList(c context.Context, eventID int64) (res []*apm.EventAdvanced, err error) {
	rows, err := d.db.Query(c, _eventAdvancedList, eventID)
	if err != nil {
		log.Error("ApmEventAdvancedList d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Error("ApmEventAdvancedList rows.Scan error(%v)", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmEventAdvancedAdd(c context.Context, eventID, displayType int64, fieldName, title, description, queryType, mappingGroup, customSql, operator string) (err error) {
	if _, err = d.db.Exec(c, _eventAdvancedAdd, eventID, fieldName, title, description, displayType, queryType, mappingGroup, customSql, operator); err != nil {
		log.Error("ApmEventAdvancedAdd d.db.Exec error(%v)", err)
	}
	return
}

func (d *Dao) ApmEventAdvancedDel(c context.Context, id int64) (err error) {
	if _, err = d.db.Exec(c, _eventAdvancedDel, id); err != nil {
		log.Error("ApmEventAdvancedDel d.db.Exec error(%v)", err)
	}
	return
}

func (d *Dao) ApmEventAdvancedUpdate(c context.Context, id, displayType int64, title, description, queryType, mappingGroup, customSql, operator string) (err error) {
	if _, err = d.db.Exec(c, _eventAdvancedUpdate, title, description, displayType, queryType, mappingGroup, customSql, operator, id); err != nil {
		log.Error("ApmEventAdvancedUpdate d.db.Exec error(%v)", err)
	}
	return
}

func (d *Dao) ApmCommandGroupListCount(c context.Context, appKey, filterKey string, eventId, busId int64, ps, pn int) (r int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey, eventId)
	if busId != 0 {
		args = append(args, busId)
		sqlAdd += " AND (bus_id = ?)"
	}
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey)
		sqlAdd += "AND (name LIKE ?)"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_commandGroupListCount, sqlAdd), args...)
	if err = row.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("d.ApmCommandGroupListCount row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmCommandGroupList(c context.Context, appKey, filterKey string, eventId, busId int64, ps, pn int) (res []*apm.CommandGroup, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey, eventId)
	if busId != 0 {
		args = append(args, busId)
		sqlAdd += " AND (a.bus_id = ?)"
	}
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey)
		sqlAdd += "AND (a.name LIKE ?)"
	}
	sqlAdd += " ORDER BY a.ctime DESC"
	sqlAdd += " LIMIT ?,?"
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_commandGroupList, sqlAdd), args...)
	if err != nil {
		log.Error("ApmCommandGroupList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.CommandGroup{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.BusID, &re.EventID, &re.Name, &re.Description, &re.Operator, &re.BusName, &re.Ctime, &re.Mtime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmCommandGroupAdd(tx *sql.Tx, appKey, name, description, userName string, busId, eventId int64) (r int64, err error) {
	res, err := tx.Exec(_commandGroupAdd, appKey, busId, eventId, name, description, userName)
	if err != nil {
		log.Error("ApmCommandGroupAdd %v", err)
		return
	}
	return res.LastInsertId()
}

func (d *Dao) TxApmCommandGroupUpdate(tx *sql.Tx, appKey, description, userName string, id, eventId int64) (r int64, err error) {
	res, err := tx.Exec(_commandGroupUpdate, description, userName, appKey, id, eventId)
	if err != nil {
		log.Error("ApmCommandGroupUpdate %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmCommandGroupDel(tx *sql.Tx, appKey string, id, eventId int64) (r int64, err error) {
	res, err := tx.Exec(_commandGroupDel, appKey, id, eventId)
	if err != nil {
		log.Error("ApmCommandGroupDel %v", err)
		return
	}
	return res.RowsAffected()
}
func (d *Dao) ApmCommandGroupByBusID(c context.Context, appKey string, eventId, busId int64) (res []*apm.CommandGroup, err error) {
	rows, err := d.db.Query(c, _commandGroupByBusID, appKey, eventId, busId)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.CommandGroup{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.BusID, &re.EventID, &re.Name, &re.Description, &re.Operator, &re.Ctime, &re.Mtime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmCommandByGroupID(c context.Context, appKey string, eventId, groupId int64) (res []*apm.CommandGroup, err error) {
	rows, err := d.db.Query(c, _commandGroupByGroupID, appKey, eventId, groupId)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.CommandGroup{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.BusID, &re.EventID, &re.Name, &re.Description, &re.Operator, &re.Ctime, &re.Mtime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmCommandList(c context.Context, appKey, filterKey string, eventId, groupId int64) (res []*apm.Command, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	args = append(args, eventId)
	if groupId != 0 {
		args = append(args, groupId)
		sqlAdd += " AND a.group_id=? "
	}
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey)
		sqlAdd += " AND (a.command LIKE ?) "
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_commandList, sqlAdd), args...)
	if err != nil {
		log.Error("ApmCommandList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.Command{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.GroupId, &re.Command, &re.Operator, &re.Ctime, &re.Mtime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmCommandAdd(tx *sql.Tx, sqls []string, args []interface{}) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_commandAdd, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("TxApmCommandAdd %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmCommandDel(tx *sql.Tx, appKey string, id int64) (r int64, err error) {
	res, err := tx.Exec(_commandDel, appKey, id)
	if err != nil {
		log.Error("ApmCommandDel %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmCommandsDel(tx *sql.Tx, appKey string, id int64) (r int64, err error) {
	res, err := tx.Exec(_commandsDel, appKey, id)
	if err != nil {
		log.Error("ApmCommandsDel %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) ApmCommandGroupAdvancedList(c context.Context, appKey string, eventId, groupId int64) (res []*apm.CommandGroupAdvanced, err error) {
	rows, err := d.db.Query(c, _commandGroupAdvancedList, appKey, eventId, groupId)
	if err != nil {
		log.Error("ApmCommandGroupAdvancedList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.CommandGroupAdvanced{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.EventID, &re.FieldName, &re.Title, &re.Description, &re.DisplayType, &re.QueryType, &re.Mapping, &re.Operator, &re.Ctime, &re.Mtime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmCommandGroupAdvancedAdd(tx *sql.Tx, appKey, fieldName, title, description, queryType, mapping, operator string, displayType int, eventId, groupId int64) (r int64, err error) {
	res, err := tx.Exec(_commandGroupAdvancedAdd, appKey, eventId, groupId, fieldName, title, description, displayType, mapping, queryType, operator)
	if err != nil {
		log.Error("ApmCommandGroupAdvancedAdd %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmCommandGroupAdvancedUpdate(tx *sql.Tx, appKey, title, description, queryType, mapping, operator string, displayType int, eventId, groupId, itemId int64) (r int64, err error) {
	res, err := tx.Exec(_commandGroupAdvancedUpdate, title, description, displayType, queryType, mapping, operator, appKey, eventId, groupId, itemId)
	if err != nil {
		log.Error("ApmCommandGroupAdvancedUpdate %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmCommandGroupAdvancedDel(tx *sql.Tx, appKey string, eventId, groupId, itemId int64) (r int64, err error) {
	res, err := tx.Exec(_commandGroupAdvancedDel, appKey, eventId, groupId, itemId)
	if err != nil {
		log.Error("ApmCommandGroupAdvancedDel %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) ApmFlowmapRouteAliasList(c context.Context, appKey, filterKey string, busID int64) (res []*apm.FlowmapRouteAlias, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		sqlAdd += "AND ((route_name LIKE ?) OR (route_alias LIKE ?))"
		args = append(args, filterKey, filterKey)
	}
	if busID != -1 {
		sqlAdd += "AND bus_id=?"
		args = append(args, busID)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_apmFlowmapRouteAliasList, sqlAdd), args...)
	if err != nil {
		log.Error("ApmFlowmapRouteAliasList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.FlowmapRouteAlias{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.BusID, &re.RouteName, &re.RouteAlias, &re.Ctime, &re.Mtime, &re.Operator); err != nil {
			log.Error("ApmFlowmapRouteAliasList scan %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmFlowMapRouteAliasAdd(tx *sql.Tx, appKey, routeName, routeAlias, userName string, busID int64) (r int64, err error) {
	res, err := tx.Exec(_apmFlowMapRouteAliasAdd, appKey, routeName, routeAlias, busID, userName)
	if err != nil {
		log.Error("TxApmFlowMapRouteAliasAdd %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmFlowMapRouteAliasUpdate(tx *sql.Tx, id int64, routeName, routeAlias, userName string, busID int64) (r int64, err error) {
	res, err := tx.Exec(_apmFlowMapRouteAliasUpdate, routeName, routeAlias, busID, userName, id)
	if err != nil {
		log.Error("TxApmFlowMapRouteAliasUpdate %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxApmFlowMapRouteAliasDel(tx *sql.Tx, id int64, userName string) (r int64, err error) {
	res, err := tx.Exec(_apmFlowMapRouteAliasDel, userName, id)
	if err != nil {
		log.Error("TxApmFlowMapRouteAliasDel %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) ApmEventSetting(c context.Context, appKey string, eventId int64) (res []*apm.ApmEventSetting, err error) {
	rows, err := d.db.Query(c, _apmEventSettiing, appKey, eventId)
	if err != nil {
		log.Error("ApmEventSettings %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.ApmEventSetting{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.EventID, &re.SampleDesc, &re.SampleConfigKey, &re.Ctime, &re.Mtime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmMetricCount(c context.Context, metric, databaseName, tableName, operator string, state, status int8, busID int64) (r int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, state)
	sqlAdd += " AND m.state<>?"
	if status != 0 {
		args = append(args, status)
		sqlAdd += " AND m.status=?"
	}
	if metric != "" {
		args = append(args, "%"+metric+"%")
		sqlAdd += " AND m.metric like ?"
	}
	if databaseName != "" {
		args = append(args, "%"+databaseName+"%")
		sqlAdd += " AND m.apm_database_name like ?"
	}
	if tableName != "" {
		args = append(args, "%"+tableName+"%")
		sqlAdd += " AND m.apm_table_name like ?"
	}
	if operator != "" {
		args = append(args, operator)
		sqlAdd += " AND m.operator=?"
	}
	if busID != 0 {
		args = append(args, busID)
		sqlAdd += " AND m.bus_id=?"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_apmMetricCount, sqlAdd), args...)
	if err = row.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("ApmMetricCount row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmMetricList(c context.Context, metric, databaseName, tableName, operator string, pn, ps int, state, status int8, busID int64, isYamlOrderBy bool) (res []*apm.PrometheusMetric, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, state)
	sqlAdd += " AND m.state<>?"
	if status != 0 {
		args = append(args, status)
		sqlAdd += " AND m.status=?"
	}
	if metric != "" {
		args = append(args, "%"+metric+"%")
		sqlAdd += " AND m.metric like ?"
	}
	if databaseName != "" {
		args = append(args, "%"+databaseName+"%")
		sqlAdd += " AND m.apm_database_name like ?"
	}
	if tableName != "" {
		args = append(args, "%"+tableName+"%")
		sqlAdd += " AND m.apm_table_name like ?"
	}
	if operator != "" {
		args = append(args, operator)
		sqlAdd += " AND m.operator=?"
	}
	if busID != 0 {
		args = append(args, busID)
		sqlAdd += " AND m.bus_id=?"
	}
	if isYamlOrderBy {
		sqlAdd += " ORDER BY m.id DESC"
	} else {
		sqlAdd += " ORDER BY m.state ASC,m.id DESC"
	}
	if pn != -1 && ps != -1 {
		args = append(args, (pn-1)*ps, ps)
		sqlAdd += " LIMIT ?,?"
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_apmMetricList, sqlAdd), args...)
	if err != nil {
		log.Error("ApmMetricList d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var cpm = &apm.PrometheusMetric{}
		if err = rows.Scan(&cpm.ID, &cpm.Metric, &cpm.MetricType, &cpm.ExecSQL, &cpm.LabeledKeys, &cpm.ValueKey,
			&cpm.TimestampKey, &cpm.Description, &cpm.ApmDatabaseName, &cpm.ApmTableName, &cpm.TimeFilter,
			&cpm.TimeOffset, &cpm.State, &cpm.Status, &cpm.URL, &cpm.BusID, &cpm.BusName, &cpm.Operator, &cpm.CTime, &cpm.MTime); err != nil {
			log.Error("ApmMetricList rows.Scan error(%v)", err)
			return
		}
		res = append(res, cpm)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmMetricByMetric(c context.Context, metric string) (res *apm.PrometheusMetric, err error) {
	row := d.db.QueryRow(c, _apmMetricByMetric, metric)
	res = &apm.PrometheusMetric{}
	if err = row.Scan(&res.ID, &res.Metric, &res.MetricType, &res.ExecSQL, &res.LabeledKeys, &res.ValueKey,
		&res.TimestampKey, &res.Description, &res.ApmDatabaseName, &res.ApmTableName, &res.TimeFilter,
		&res.TimeOffset, &res.Operator, &res.CTime, &res.MTime, &res.State, &res.Status, &res.URL, &res.BusID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Error("ApmMetricByID row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmMetricAdd(c context.Context, metric, mType, sqlName, labeledKeys, valueKey, timestampKey, description, databaseName, tableName, operator, url string,
	timeFilter, timeOffset, busID int64, status int8) (r int64, err error) {
	res, err := d.db.Exec(c, _apmMetricAdd, metric, mType, sqlName, labeledKeys, valueKey, timestampKey, description, databaseName, tableName, timeFilter, timeOffset, operator, status, url, busID)
	if err != nil {
		log.Error("d.db.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

func (d *Dao) ApmMetricUpdate(c context.Context, metric, mType, sqlName, labeledKeys, valueKey, timestampKey, description, databaseName, tableName, operator, url string,
	timeFilter, timeOffset, busID int64, state, status int8) (effect int64, err error) {
	res, err := d.db.Exec(c, _apmMetricUpdate, mType, sqlName, labeledKeys, valueKey, timestampKey, description, databaseName, tableName, timeFilter, timeOffset, operator, state, status, url, busID, metric)
	if err != nil {
		log.Error("d.db.Exec error(%v)", err)
		return
	}
	effect, err = res.RowsAffected()
	return
}

func (d *Dao) ApmMetricDelByUpdate(c context.Context, state int8, metric string) (effect int64, err error) {
	res, err := d.db.Exec(c, _apmMetricDelByUpdate, state, metric)
	if err != nil {
		log.Error("d.db.Exec error(%v)", err)
		return
	}
	effect, err = res.RowsAffected()
	return
}

func (d *Dao) ApmMetricDel(c context.Context, metric string) (effect int64, err error) {
	res, err := d.db.Exec(c, _apmMetricDel, metric)
	if err != nil {
		log.Error("d.db.Exec error(%v)", err)
		return
	}
	effect, err = res.RowsAffected()
	return
}
func (d *Dao) TxApmMetricPublish(tx *sql.Tx, fileMD5, localPath, operator, description string) (r int64, err error) {
	res, err := tx.Exec(_apmMetricPublish, fileMD5, localPath, description, operator)
	if err != nil {
		log.Error("ApmMetricPublish d.db.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

func (d *Dao) ApmMetricPublishStateUpdate(tx *sql.Tx, publishState int) (err error) {
	_, err = tx.Exec(_apmMetricPublishStateUpdate, publishState)
	if err != nil {
		return
	}
	return
}

func (d *Dao) TxApmMetricPublishDel(tx *sql.Tx) (err error) {
	_, err = tx.Exec(_apmMetricPublishDel)
	if err != nil {
		return
	}
	return
}

func (d *Dao) ApmMetricPublishCount(c context.Context, fileMD5, localPath, description, operator string) (r int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if fileMD5 != "" {
		args = append(args, fileMD5)
		sqlAdd += " AND md5=?"
	}
	if localPath != "" {
		args = append(args, localPath)
		sqlAdd += " AND local_path=?"
	}
	if description != "" {
		args = append(args, description)
		sqlAdd += " AND description=?"
	}
	if operator != "" {
		args = append(args, operator)
		sqlAdd += " AND operator=?"
	}
	if len(args) > 0 {
		sqlAdd = strings.Replace(sqlAdd, " AND", " WHERE", 1)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_apmMetricPublishCount, sqlAdd), args...)
	if err = row.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("ApmMetricPublishCount row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmMetricPublishList(c context.Context, fileMD5, localPath, description, operator string, pn, ps int) (res []*apm.PrometheusMetricPublish, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if fileMD5 != "" {
		args = append(args, fileMD5)
		sqlAdd += " AND md5=?"
	}
	if localPath != "" {
		args = append(args, localPath)
		sqlAdd += " AND local_path=?"
	}
	if description != "" {
		args = append(args, description)
		sqlAdd += " AND description=?"
	}
	if operator != "" {
		args = append(args, operator)
		sqlAdd += " AND operator=?"
	}
	if len(args) > 0 {
		sqlAdd = strings.Replace(sqlAdd, " AND", " WHERE", 1)
	}
	sqlAdd += " ORDER BY ctime DESC"
	if pn != -1 && ps != -1 {
		args = append(args, (pn-1)*ps, ps)
		sqlAdd += " LIMIT ?,?"
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_apmMetricPublishList, sqlAdd), args...)
	if err != nil {
		log.Error("ApmMetricPublishList d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var sp = &apm.PrometheusMetricPublish{}
		if err = rows.Scan(&sp.ID, &sp.MD5, &sp.LocalPath, &sp.Description, &sp.IsActiveVersion, &sp.Operator, &sp.CTime, &sp.MTime); err != nil {
			log.Error("ApmMetricPublishList rows.Scan error(%v)", err)
			return
		}
		res = append(res, sp)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmMetricPublishDiff(c context.Context) (localPath string, err error) {
	row := d.db.QueryRow(c, _apmMetricPublishDiff)
	if err = row.Scan(&localPath); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("ApmMetricPublishDiff row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmMetricPublishById(c context.Context, id int64) (res *apm.PrometheusMetricPublish, err error) {
	rows, err := d.db.Query(c, _apmMetricPublishById, id)
	if err != nil {
		return
	}
	defer rows.Close()
	var list []*apm.PrometheusMetricPublish
	if err = xsql.ScanSlice(rows, &list); err != nil {
		return
	}
	err = rows.Err()
	if len(list) == 1 {
		return list[0], err
	}
	return nil, err
}

func (d *Dao) ApmMetricPublishActive(c context.Context) (res *apm.PrometheusMetricPublish, err error) {
	rows, err := d.db.Query(c, _apmMetricPublishActive)
	if err != nil {
		return
	}
	defer rows.Close()
	var list []*apm.PrometheusMetricPublish
	if err = xsql.ScanSlice(rows, &list); err != nil {
		return
	}
	err = rows.Err()
	if len(list) == 1 {
		return list[0], err
	}
	return
}

func (d *Dao) TxApmMetricPublishActiveVerUpdate(tx *sql.Tx, id int64, isActive int8) (err error) {
	_, err = tx.Exec(_apmMetricPublishActiveVerUpdate, isActive, id)
	return
}

// ApmFlinkJobCount Flink任务数量统计
func (d *Dao) ApmFlinkJobCount(c context.Context, logId, name, description, owner, operator string, state int, startTime, endTime xtime.Time) (r int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if logId != "" {
		args = append(args, logId)
		sqlAdd += " AND log_id=?"
	}
	if name != "" {
		args = append(args, "%"+name+"%")
		sqlAdd += " AND name LIKE ?"
	}
	if description != "" {
		args = append(args, "%"+description+"%")
		sqlAdd += " AND description LIKE ?"
	}
	if owner != "" {
		args = append(args, owner)
		sqlAdd += " AND owner=?"
	}
	if operator != "" {
		args = append(args, operator)
		sqlAdd += " AND operator=?"
	}
	if state != 0 {
		args = append(args, state)
		sqlAdd += " AND state=?"
	}
	if startTime != 0 {
		args = append(args, startTime)
		sqlAdd += " AND ctime>?"
	}
	if endTime != 0 {
		args = append(args, endTime)
		sqlAdd += " AND ctime<?"
	}
	if len(args) > 0 {
		sqlAdd = strings.Replace(sqlAdd, " AND", " WHERE", 1)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_apmFlinkJobCount, sqlAdd), args...)
	if err = row.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("ApmFlinkJobCount row.Scan error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) ApmFlinkJobById(c context.Context, id int64) (res *apm.FlinkJobDB, err error) {
	row := d.db.QueryRow(c, _apmFlinkJobByID, id)
	res = &apm.FlinkJobDB{}
	if err = row.Scan(&res.ID, &res.LogID, &res.Name, &res.Description, &res.Owner, &res.Operator, &res.State, &res.CTime, &res.MTime); err != nil {
		log.Error("ApmFlinkJobById row.Scan error(%v)", err)
		return
	}
	return
}

// ApmFlinkJobList Flink任务列表查询
func (d *Dao) ApmFlinkJobList(c context.Context, logId, name, description, owner, operator string, state int, startTime, endTime xtime.Time, pn, ps int) (res []*apm.FlinkJobDB, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if logId != "" {
		args = append(args, logId)
		sqlAdd += " AND log_id=?"
	}
	if name != "" {
		args = append(args, "%"+name+"%")
		sqlAdd += " AND name LIKE ?"
	}
	if description != "" {
		args = append(args, "%"+description+"%")
		sqlAdd += " AND description LIKE ?"
	}
	if owner != "" {
		args = append(args, owner)
		sqlAdd += " AND owner=?"
	}
	if operator != "" {
		args = append(args, operator)
		sqlAdd += " AND operator=?"
	}
	if state != 0 {
		args = append(args, state)
		sqlAdd += " AND state=?"
	}
	if startTime != 0 {
		args = append(args, startTime)
		sqlAdd += " AND ctime>?"
	}
	if endTime != 0 {
		args = append(args, endTime)
		sqlAdd += " AND ctime<?"
	}
	if len(args) > 0 {
		sqlAdd = strings.Replace(sqlAdd, " AND", " WHERE", 1)
	}
	sqlAdd += " ORDER BY ctime DESC"
	if pn != -1 && ps != -1 {
		args = append(args, (pn-1)*ps, ps)
		sqlAdd += " LIMIT ?,?"
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_apmFlinkJobList, sqlAdd), args...)
	if err != nil {
		log.Error("ApmFlinkJobList d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var fjd = &apm.FlinkJobDB{}
		if err = rows.Scan(&fjd.ID, &fjd.LogID, &fjd.Name, &fjd.Description,
			&fjd.Owner, &fjd.Operator, &fjd.State, &fjd.CTime, &fjd.MTime); err != nil {
			log.Error("ApmFlinkJobList rows.Scan error(%v)", err)
			return
		}
		res = append(res, fjd)
	}
	err = rows.Err()
	return
}

// ApmFlinkJobAdd Flink任务增加
func (d *Dao) ApmFlinkJobAdd(c context.Context, logId, name, description, owner, operator string, state int) (r int64, err error) {
	res, err := d.db.Exec(c, _apmFlinkJobAdd, logId, name, description, owner, operator, state)
	if err != nil {
		log.Error("ApmFlinkJobAdd d.db.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

// ApmFlinkJobUpdate Flink任务更新
func (d *Dao) ApmFlinkJobUpdate(c context.Context, logId, name, description, owner, operator string, state int, id int64) (effect int64, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, operator)
	if logId != "" {
		args = append(args, logId)
		sqlAdd += ",log_id=?"
	}
	if name != "" {
		args = append(args, name)
		sqlAdd += ",name=?"
	}
	if description != "" {
		args = append(args, description)
		sqlAdd += ",description=?"
	}
	if owner != "" {
		args = append(args, owner)
		sqlAdd += ",owner=?"
	}
	if state != 0 {
		args = append(args, state)
		sqlAdd += ",state=?"
	}
	args = append(args, id)
	sqlAdd += " WHERE id=?"
	res, err := d.db.Exec(c, fmt.Sprintf(_apmFlinkJobUpdate, sqlAdd), args...)
	if err != nil {
		log.Error("ApmFlinkJobUpdate d.db.Exec error(%v)", err)
		return
	}
	effect, err = res.RowsAffected()
	return
}

// ApmFlinkJobDel Flink任务删除
func (d *Dao) ApmFlinkJobDel(c context.Context, id int64) (effect int64, err error) {
	res, err := d.db.Exec(c, _apmFlinkJobDel, id)
	if err != nil {
		log.Error("ApmFlinkDel d.db.Exec error(%v)", err)
		return
	}
	effect, err = res.RowsAffected()
	return
}

// ApmFlinkJobRelationList Flink任务和Events关联关系查询
func (d *Dao) ApmFlinkJobRelationList(c context.Context, jobID int64) (res []*apm.Event, err error) {
	rows, err := d.db.Query(c, _apmFlinkJobRelationList, jobID)
	if err != nil {
		log.Error("ApmEventFlinkRelList d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var eventTmp = &apm.Event{}
		if err = rows.Scan(&eventTmp.ID, &eventTmp.AppKeys, &eventTmp.BusID, &eventTmp.Databases, &eventTmp.TableName, &eventTmp.Name, &eventTmp.Description, &eventTmp.Owner, &eventTmp.Shared, &eventTmp.Operator, &eventTmp.State, &eventTmp.Ctime, &eventTmp.Mtime, &eventTmp.Topic, &eventTmp.LogID, &eventTmp.DistributedTableName, &eventTmp.SampleRate, &eventTmp.IsWideTable); err != nil {
			log.Error("ApmEventFlinkRelList rows.Scan error(%v)", err)
			return
		}
		res = append(res, eventTmp)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmFlinkJobRelation(c context.Context, jobID, eventID int64) (res *apm.EventFlinkRelDB, err error) {
	row := d.db.QueryRow(c, _apmFlinkJobRelation, jobID, eventID)
	res = &apm.EventFlinkRelDB{}
	if err = row.Scan(&res.ID, &res.EventID, &res.JobID, &res.Operator, &res.CTime, &res.MTime, &res.State); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("ApmFlinkJobPublishModifyCount row.Scan error(%v)", err)
		}
	}
	return
}

// ApmFlinkJobRelationAdd Flink任务和Events关联关系增加
func (d *Dao) ApmFlinkJobRelationAdd(c context.Context, eventID, jobID int64, operator string, state int) (r int64, err error) {
	res, err := d.db.Exec(c, _apmFlinkJobRelationAdd, eventID, jobID, operator, state)
	if err != nil {
		log.Error("ApmEventJobRelAdd d.db.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

// ApmFlinkJobRelationDelByUpdate Flink任务和Events关联关系通过更新删除
func (d *Dao) ApmFlinkJobRelationDelByUpdate(c context.Context, jobID, eventID int64, state int) (effect int64, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, state, jobID)
	if eventID != 0 {
		sqlAdd += " AND event_id=?"
		args = append(args, eventID)
	}
	res, err := d.db.Exec(c, fmt.Sprintf(_apmFlinkJobRelationStateUpdate, sqlAdd), args...)
	if err != nil {
		log.Error("ApmEventFlinkRelDel d.db.Exec error(%v)", err)
		return
	}
	effect, err = res.RowsAffected()
	return
}

// ApmFlinkJobRelationDel Flink任务和Events关联关系删除
func (d *Dao) ApmFlinkJobRelationDel(c context.Context, jobID, eventID int64) (effect int64, err error) {
	res, err := d.db.Exec(c, _apmFlinkJobRelationDel, jobID, eventID)
	if err != nil {
		log.Error("ApmFlinkJobRelationDel d.db.Exec error(%v)", err)
		return
	}
	effect, err = res.RowsAffected()
	return
}

// ApmFlinkJobPublishCount Flink任务publish统计
func (d *Dao) ApmFlinkJobPublishCount(c context.Context, jobID int64) (r int, err error) {
	row := d.db.QueryRow(c, _apmFlinkJobPublishCount, jobID)
	if err = row.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("ApmFlinkJobPublishCount row.Scan error(%v)", err)
			return
		}
	}
	return
}

// ApmFlinkJobPublishList Flink publish列表
func (d *Dao) ApmFlinkJobPublishList(c context.Context, jobID int64, pn, ps int) (res []*apm.EventFlinkRelPublish, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, jobID)
	sqlAdd += " ORDER BY ctime DESC"
	if pn > 0 && ps > 0 {
		sqlAdd += " LIMIT ?,?"
		args = append(args, (pn-1)*ps, ps)
	}

	rows, err := d.db.Query(c, fmt.Sprintf(_apmFlinkJobPublishList, sqlAdd), args...)
	if err != nil {
		log.Error("ApmFlinkJobPublishList d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var re = &apm.EventFlinkRelPublish{}
		if err = rows.Scan(&re.ID, &re.FlinkJobID, &re.MD5, &re.LocalPath, &re.Description, &re.Operator, &re.CTime, &re.MTime); err != nil {
			log.Error("ApmFlinkJobPublishList rows.Scan error(%v)", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmFlinkJobPublishModifyCount(c context.Context, id int64) (r int64, err error) {
	row := d.db.QueryRow(c, _apmFlinkJobPublishModifyCount, id)
	if err = row.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("ApmFlinkJobPublishModifyCount row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmFlinkJobPublish(c context.Context, flinkJobID int64, md5, localPath, description, operator string) (r int64, err error) {
	res, err := d.db.Exec(c, _apmFlinkJobPublish, flinkJobID, md5, localPath, description, operator)
	if err != nil {
		log.Error("ApmFlinkJobPublish d.db.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

func (d *Dao) ApmFlinkJobPublishStateUpdate(c context.Context, state int, jobID int64) (effect int64, err error) {
	res, err := d.db.Exec(c, _apmFlinkJobPublishStateUpdate, state, jobID)
	if err != nil {
		log.Error("ApmFlinkJobPublishStateUpdate d.db.Exec error(%v)", err)
		return
	}
	effect, err = res.RowsAffected()
	return

}

func (d *Dao) ApmFlinkJobPublishDel(c context.Context, jobID int64) (effect int64, err error) {
	res, err := d.db.Exec(c, _apmFlinkJobPublishDel, jobID)
	if err != nil {
		log.Error("ApmFlinkJobPublishDel d.db.Exec error(%v)", err)
		return
	}
	effect, err = res.RowsAffected()
	return

}

func (d *Dao) ApmFlinkJobLastPath(c context.Context, jobID int64) (localPath string, err error) {
	row := d.db.QueryRow(c, _apmFlinkJobLastPath, jobID)
	if err = row.Scan(&localPath); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("ApmFlinkJobPublishDiff row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) ApmCrashRule(c context.Context, id int64) (res *apm.CrashRule, err error) {
	row := d.db.QueryRow(c, _apmCrashRule, id)
	res = &apm.CrashRule{}
	if err = row.Scan(&res.ID, &res.AppKeys, &res.BusID, &res.RuleName, &res.KeyWords, &res.PageKeyWords, &res.Operator, &res.Ctime, &res.Mtime, &res.Description, &res.BusName); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("ApmCrashRule row.Scan error(%v)", err)
		}
	}
	return
}

// ApmCrashRuleListCount 堆栈解析规则表的计数
func (d *Dao) ApmCrashRuleListCount(c context.Context, keyWords, pageKeyWords, appKeys string, busID int64) (total int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if appKeys != "" {
		args = append(args, appKeys)
		sqlAdd += "AND FIND_IN_SET(?, c.app_keys)"
	}
	if busID != 0 {
		args = append(args, busID)
		sqlAdd += " AND bus_id=?"
	}
	if keyWords != "" {
		keyWordArr := strings.Split(keyWords, ",")
		sqlAddTmp := " AND ("
		for _, word := range keyWordArr {
			sqlAddTmp += " c.keywords LIKE BINARY ? OR"
			args = append(args, "%"+string(word)+"%")
		}
		sqlAdd += strings.TrimRight(sqlAddTmp, "OR")
		sqlAdd += ")"
	}
	if pageKeyWords != "" {
		keyWordArr := strings.Split(pageKeyWords, ",")
		sqlAddTmp := " AND ("
		for _, word := range keyWordArr {
			sqlAddTmp += " c.page_keywords LIKE BINARY ? OR"
			args = append(args, "%"+string(word)+"%")
		}
		sqlAdd += strings.TrimRight(sqlAddTmp, "OR")
		sqlAdd += ")"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_apmCrashRuleCount, sqlAdd), args...)
	if err = row.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("ApmCrashListCount row.Scan error(%v)", err)
		}
	}
	return
}

// ApmCrashRuleList 堆栈解析规则列表
func (d *Dao) ApmCrashRuleList(c context.Context, keyWords, pageKeyWords, appKeys string, busID int64, pn, ps int) (res []*apm.CrashRule, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if appKeys != "" {
		args = append(args, appKeys)
		sqlAdd += "AND FIND_IN_SET(?, c.app_keys)"
	}
	if busID != 0 {
		args = append(args, busID)
		sqlAdd += " AND bus_id=?"
	}
	if keyWords != "" {
		keyWordArr := strings.Split(keyWords, ",")
		sqlAddTmp := " AND ("
		for _, word := range keyWordArr {
			sqlAddTmp += " c.keywords LIKE BINARY ? OR"
			args = append(args, "%"+string(word)+"%")
		}
		sqlAdd += strings.TrimRight(sqlAddTmp, "OR")
		sqlAdd += ")"
	}
	if pageKeyWords != "" {
		keyWordArr := strings.Split(pageKeyWords, ",")
		sqlAddTmp := " AND ("
		for _, word := range keyWordArr {
			sqlAddTmp += " c.page_keywords LIKE BINARY ? OR"
			args = append(args, "%"+string(word)+"%")
		}
		sqlAdd += strings.TrimRight(sqlAddTmp, "OR")
		sqlAdd += ")"
	}
	sqlAdd += " ORDER BY c.ctime DESC"
	args = append(args, (pn-1)*ps, ps)
	sqlAdd += " LIMIT ?,?"
	rows, err := d.db.Query(c, fmt.Sprintf(_apmCrashRuleList, sqlAdd), args...)
	if err != nil {
		log.Error("ApmCrashRuleList d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.CrashRule{}
		if err = rows.Scan(&re.ID, &re.AppKeys, &re.BusID, &re.RuleName, &re.KeyWords, &re.PageKeyWords, &re.Operator, &re.Ctime, &re.Mtime, &re.Description, &re.BusName); err != nil {
			log.Error("ApmCrashRuleList rows.Scan error(%v)", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// ApmCrashRuleAdd 堆栈解析规则增加
func (d *Dao) ApmCrashRuleAdd(c context.Context, appKeys, ruleName, keyWords, pageKeyWords, operator, description string, busID int64) (err error) {
	if _, err = d.db.Exec(c, _apmCrashRuleAdd, appKeys, busID, ruleName, keyWords, pageKeyWords, operator, description); err != nil {
		log.Error("ApmCrashRuleAdd d.db.Exec error(%v)", err)
	}
	return
}

// ApmCrashRuleDel 堆栈解析规则删除
func (d *Dao) ApmCrashRuleDel(c context.Context, id int64) (err error) {
	if _, err = d.db.Exec(c, _apmCrashRuleDel, id); err != nil {
		log.Error("ApmCrashRuleDel d.db.Exec error(%v)", err)
	}
	return
}

// ApmCrashRuleUpdate 堆栈解析规则更新
func (d *Dao) ApmCrashRuleUpdate(c context.Context, appKeys, ruleName, keyWords, pageKeyWords, operator, description string, busID, id int64) (err error) {
	if _, err = d.db.Exec(c, _apmCrashRuleUpdate, appKeys, busID, ruleName, keyWords, pageKeyWords, operator, description, id); err != nil {
		log.Error("ApmCrashRuleUpdate d.db.Exec error(%v)", err)
	}
	return
}

// TxApmAppEventRelAdd app和event关联增加
func (d *Dao) TxApmAppEventRelAdd(tx *sql.Tx, eventId, datacenterAppId, datacenterEventId int64, operator string) (err error) {
	_, err = tx.Exec(_appEventRelAdd, eventId, datacenterAppId, datacenterEventId, operator)
	return
}

func (d *Dao) ApmAppEventRelList(c context.Context, eventId, datacenterAppId int64) (res []*apm.EventDatacenterRel, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, eventId)
	if datacenterAppId != 0 {
		sqlAdd += "AND datacenter_app_id=?"
		args = append(args, datacenterAppId)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_appEventRelList, sqlAdd), args...)
	if err != nil {
		return
	}
	if rows.Err() != nil {
		return
	}
	err = xsql.ScanSlice(rows, &res)
	return
}

func (d *Dao) ApmEventVedaConfig(c context.Context, eventId int64) (re *apm.EventVedaConfig, err error) {
	row := d.db.QueryRow(c, _apmVedaConfig, eventId)
	re = &apm.EventVedaConfig{}
	if err = row.Scan(&re.ID, &re.EventID, &re.EventName, &re.VedaDBName, &re.VedaIndexTable, &re.VedaStackTable, &re.HashColumn, &re.Ctime, &re.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("%v", err)
		}
	}
	return
}

// ApmEventStorageUpdate 技术埋点监控存储更新
func (d *Dao) ApmEventStorageUpdate(c context.Context, monitors []*apm.EventMonitor) (err error) {
	var (
		sqlAdd []string
		args   []interface{}
	)
	if len(monitors) < 1 {
		return
	}
	for _, monitor := range monitors {
		sqlAdd = append(sqlAdd, "(?,?,?)")
		args = append(args, monitor.EventId, monitor.StorageCount, monitor.StorageCapacity)
	}
	_, err = d.db.Exec(c, fmt.Sprintf(_upEventStorage, strings.Join(sqlAdd, ",")), args...)
	return
}

func (d *Dao) ApmEventStorageList(c context.Context, logData string) (res map[string][]*apm.EventMonitor, err error) {
	res = make(map[string][]*apm.EventMonitor)
	rows, err := d.db.Query(c, _eventStorageList, logData)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.EventMonitor{}
		if err = rows.Scan(&re.EventName, &re.DatacenterAppId, &re.StorageCount, &re.StorageCapacity); err != nil {
			return
		}
		res[re.EventName] = append(res[re.EventName], re)
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmEventFieldStateUpdateById(tx *sql.Tx, id int64, state int8) (err error) {
	_, err = tx.Exec(_updateEventFieldState, state, id)
	return
}

func (d *Dao) ApmEventFieldModifyCount(c context.Context) (res map[int64]int64, err error) {
	res = make(map[int64]int64)
	rows, err := d.db.Query(c, _eventFieldModifyCount)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var re struct {
			EventId int64
			Count   int64
		}
		if err = rows.Scan(&re.EventId, &re.Count); err != nil {
			return
		}
		res[re.EventId] = re.Count
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmEventFieldFileAdd(tx *sql.Tx, fieldVersion int64, eventFields []*apm.EventField, operator string) (err error) {
	var (
		sqlAdd []string
		args   []interface{}
	)
	if len(eventFields) < 1 {
		return
	}
	for _, field := range eventFields {
		sqlAdd = append(sqlAdd, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		args = append(args, field.EventID, field.ID, field.Key, field.Example, field.Type, field.Index, field.Description, field.Mode, field.DefaultValue, field.IsClickhouse, field.ISElasticsearchIndex, field.ElasticSearchFieldType, field.State, fieldVersion, operator)
	}
	_, err = tx.Exec(fmt.Sprintf(_eventFieldFileAdd, strings.Join(sqlAdd, ",")), args...)
	return
}

func (d *Dao) TxApmEventFieldStateUpdate(tx *sql.Tx, eventId int64, state int8) (err error) {
	_, err = tx.Exec(_upEventFieldStateByEventId, state, eventId)
	return
}

func (d *Dao) ApmEventFieldFileLastFV(c context.Context, eventId int64) (fv int64, err error) {
	row := d.db.QueryRow(c, _eventFieldFileLastFV, eventId)
	if err = row.Scan(&fv); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

func (d *Dao) ApmEventFieldFileList(c context.Context, eventId, fv int64) (res []*apm.EventFieldFile, err error) {
	rows, err := d.db.Query(c, _eventFieldFileList, eventId, fv)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "scan error: %#v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmEventFieldDelById(tx *sql.Tx, id int64) (err error) {
	_, err = tx.Exec(_delEventFieldById, id)
	return
}

func (d *Dao) TxApmEventFieldPublish(tx *sql.Tx, eventId, fv int64, operator string) (err error) {
	_, err = tx.Exec(_eventFieldPublishAdd, eventId, fv, operator)
	return

}

func (d *Dao) ApmEventFieldPublishHistoryCount(c context.Context, eventId int64) (count int, err error) {
	row := d.db.QueryRow(c, _eventFieldPublishCount, eventId)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

func (d *Dao) ApmEventFieldPublishHistory(c context.Context, eventId int64, pn, ps int) (res []*apm.EventFieldPublish, err error) {
	rows, err := d.db.Query(c, _eventFieldPublishList, eventId, (pn-1)*ps, ps)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "scan error: %#v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmEventFieldPublishLastVersion(c context.Context, eventId, version int64) (lastVersion int64, err error) {
	row := d.db.QueryRow(c, _eventFieldPublishLastVersion, eventId, version)
	if err = row.Scan(&lastVersion); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

func (d *Dao) ApmEventFieldStateSync(c context.Context, id int64, esType int8) (err error) {
	_, err = d.db.Exec(c, _eventFieldTypeSync, esType, id)
	return
}

func (d *Dao) TxApmAppEventCommonFieldGroupAdd(tx *sql.Tx, appKey, name, description, operator string, isDefault int8) (lastId int64, err error) {
	row, err := tx.Exec(_eventCommonFieldGroupAdd, appKey, name, description, isDefault, operator)
	if err != nil {
		return
	}
	return row.LastInsertId()
}

func (d *Dao) TxApmAppEventCommonFieldGroupUpdate(tx *sql.Tx, name, description, operator string, isDefault int8, id int64) (err error) {
	_, err = tx.Exec(_eventCommonFieldGroupUpdate, name, description, isDefault, operator, id)
	return
}

func (d *Dao) TxApmAppEventCommonFieldGroupDel(tx *sql.Tx, id int64) (err error) {
	_, err = tx.Exec(_eventCommonFieldGroupDel, id)
	return
}

func (d *Dao) ApmAppEventCommonFieldGroupCount(c context.Context, appKey string) (count int64, err error) {
	row := d.db.QueryRow(c, _eventCommonFieldGroupCount, appKey)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

func (d *Dao) ApmAppEventCommonFieldGroupList(c context.Context, appKey string, pn, ps int) (res []*apm.EventCommonFieldGroup, err error) {
	rows, err := d.db.Query(c, _eventCommonFieldGroupList, appKey, (pn-1)*ps, ps)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmAppEventCommonFieldGroupById(c context.Context, id int64) (res *apm.EventCommonFieldGroup, err error) {
	rows, err := d.db.Query(c, _eventCommonFieldGroupById, id)
	if err != nil {
		return
	}
	defer rows.Close()
	var l []*apm.EventCommonFieldGroup
	if err = xsql.ScanSlice(rows, &l); err != nil {
		return
	}
	err = rows.Err()
	return l[0], err
}

func (d *Dao) TxApmAppEventCommonFieldAdd(tx *sql.Tx, appKey string, groupId int64, operator string, fields []*apm.AppEventCommonField) (err error) {
	var (
		sqlAdd []string
		args   []interface{}
	)
	if len(fields) < 1 {
		return
	}
	for _, field := range fields {
		sqlAdd = append(sqlAdd, "(?,?,?,?,?,?,?,?,?,?,?)")
		args = append(args, appKey, groupId, field.Key, field.Type, field.Index, field.Description, field.DefaultValue, field.IsClickhouse, field.IsElasticsearchIndex, field.ElasticsearchFieldType, operator)
	}
	_, err = tx.Exec(fmt.Sprintf(_eventCommonFieldAdd, strings.Join(sqlAdd, ",")), args...)
	return
}

func (d *Dao) TxApmAppEventCommonFieldDel(tx *sql.Tx, id int64) (err error) {
	_, err = tx.Exec(_eventCommonFieldDel, id)
	return
}

func (d *Dao) TxApmAppEventCommonFieldDelByGroupId(tx *sql.Tx, groupId int64) (err error) {
	_, err = tx.Exec(_eventCommonFieldDelByGroupId, groupId)
	return
}

func (d *Dao) TxApmAppEventCommonFieldUpdate(tx *sql.Tx, description, defaultValue, operator string, fieldType, isClickhouse, isElasticsearchIndex, elasticsearchFieldType int8, id, fieldIndex int64) (err error) {
	_, err = tx.Exec(_eventCommonFieldUpdate, fieldType, fieldIndex, description, defaultValue, isClickhouse, isElasticsearchIndex, elasticsearchFieldType, operator, id)
	return
}

func (d *Dao) ApmAppEventCommonFieldList(c context.Context, groupId int64) (res []*apm.AppEventCommonField, err error) {
	rows, err := d.db.Query(c, _eventCommonFieldList, groupId)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmEventAlertAdd(tx *sql.Tx, title, description, timeField, aggField, filterQuery, denominatorFilterQuery, triggerCondition, groupField, notifyFields, channels, target, botWebhook, webhook, mutePeriod, creator, operator string, eventId, billionsId, intervals, timeFrame, notifyDur, version, minLogCount, datacenterAppId int64, cluster, level, aggType, muteType, isEnable, aggPercentile, isLogDetail int8) (err error) {
	_, err = tx.Exec(_eventAlertAdd, eventId, datacenterAppId, billionsId, title, description, intervals, timeField, cluster, level, timeFrame, aggType, aggField, aggPercentile, filterQuery, denominatorFilterQuery, triggerCondition, groupField, notifyFields, notifyDur, channels, target, botWebhook, webhook, muteType, mutePeriod, version, minLogCount, isEnable, isLogDetail, creator, operator)
	return
}

func (d *Dao) ApmEventAlertUpdate(tx *sql.Tx, title, description, timeField, aggField, filterQuery, denominatorFilterQuery, triggerCondition, groupField, notifyFields, channels, target, botWebhook, webhook, mutePeriod, operator string, intervals, timeFrame, notifyDur, version, minLogCount, datacenterAppId, id int64, cluster, level, aggType, muteType, isLogDetail, aggPercentile int8) (err error) {
	_, err = tx.Exec(_eventAlertUpdate, datacenterAppId, title, description, version, minLogCount, intervals, timeField, cluster, level, timeFrame, aggType, aggField, aggPercentile, filterQuery, denominatorFilterQuery, triggerCondition, groupField, notifyFields, notifyDur, channels, target, botWebhook, webhook, muteType, mutePeriod, isLogDetail, operator, id)
	return
}

func (d *Dao) ApmEventAlertDel(tx *sql.Tx, id int64) (err error) {
	_, err = tx.Exec(_eventAlertDel, id)
	return
}

func (d *Dao) ApmEventAlert(c context.Context, id int64) (res *apm.EventAlertDB, err error) {
	rows, err := d.db.Query(c, _eventAlertInfo, id)
	if err != nil {
		return
	}
	defer rows.Close()
	var li []*apm.EventAlertDB
	if err = xsql.ScanSlice(rows, &li); err != nil {
		log.Error("ScanSlice error %v", err)
		return
	}
	if len(li) == 1 {
		return li[0], err
	}
	err = rows.Err()
	return nil, err
}

func (d *Dao) ApmEventAlertCount(c context.Context, eventId, datacenterAppId int64, title, eventName string, isEnable, level int8) (count int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if eventId != 0 {
		sqlAdd += " AND r.event_id=? "
		args = append(args, eventId)
	}
	if eventName != "" {
		eventName = "%" + eventName + "%"
		sqlAdd += " AND e.name LIKE ? "
		args = append(args, eventName)
	}
	if datacenterAppId != 0 {
		sqlAdd += " AND r.datacenter_app_id=? "
		args = append(args, datacenterAppId)
	}
	if title != "" {
		title = "%" + title + "%"
		sqlAdd += " AND r.title LIKE ? "
		args = append(args, title)
	}
	if isEnable != 0 {
		sqlAdd += " AND r.is_enable=? "
		args = append(args, isEnable)
	}
	if level != 0 {
		sqlAdd += " AND r.level=? "
		args = append(args, level)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_eventAlertCount, strings.Replace(sqlAdd, "AND", "WHERE", 1)), args...)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

func (d *Dao) ApmEventAlertList(c context.Context, eventId, datacenterAppId int64, title, eventName string, isEnable, level int8, pn, ps int) (res []*apm.EventAlertDB, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if eventId != 0 {
		sqlAdd += " AND r.event_id=? "
		args = append(args, eventId)
	}
	if eventName != "" {
		eventName = "%" + eventName + "%"
		sqlAdd += " AND e.name LIKE ? "
		args = append(args, eventName)
	}
	if datacenterAppId != 0 {
		sqlAdd += " AND r.datacenter_app_id=?"
		args = append(args, datacenterAppId)
	}
	if title != "" {
		title = "%" + title + "%"
		sqlAdd += " AND r.title LIKE ? "
		args = append(args, title)
	}
	if isEnable != 0 {
		sqlAdd += " AND r.is_enable=? "
		args = append(args, isEnable)
	}
	if level != 0 {
		sqlAdd += " AND r.level=? "
		args = append(args, level)
	}
	sqlAdd += " ORDER BY r.ctime DESC "
	if pn != 0 && ps != 0 {
		args = append(args, (pn-1)*ps, ps)
		sqlAdd += " LIMIT ?,?"
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_eventAlertList, strings.Replace(sqlAdd, "AND", "WHERE", 1)), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmEventAlertSwitch(tx *sql.Tx, isEnable int8, id int64) (err error) {
	_, err = tx.Exec(_eventAlertSwitch, isEnable, id)
	return
}

// ApmEventCompletionList 未注册的Appid的技术埋点的数据补全
func (d *Dao) ApmEventCompletionList(c context.Context, logData string) (res []*apm.EventCompletion, err error) {
	if logData == "" {
		return
	}
	rows, err := d.db.Query(c, _eventCompletionList, logData)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "ScanSlice error %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmAlertRuleByHawkeyeIds(c context.Context, hawkeyeIds []int64) (res map[int64]*apm.AlertRule, err error) {
	var (
		sqls []string
		args []interface{}
	)
	if len(hawkeyeIds) == 0 {
		return
	}
	for _, hawkeyeId := range hawkeyeIds {
		sqls = append(sqls, "?")
		args = append(args, hawkeyeId)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_apmAlertRuleByHawkeyeIds, strings.Join(sqls, ",")), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	res = make(map[int64]*apm.AlertRule)
	for rows.Next() {
		re := &apm.AlertRule{}
		if err = rows.Scan(&re.Id, &re.HawkeyeId, &re.Name, &re.TriggerCondition, &re.Species, &re.QueryExprs, &re.RuleType,
			&re.Markdown, &re.Operator, &re.CTime, &re.MTime); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		res[re.HawkeyeId] = re
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmAlertRuleCount(c context.Context, hawkeyeId int64, name, species, queryExprs string, ruleType int8) (count int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if ruleType != -1 {
		args = append(args, ruleType)
		sqlAdd += " AND rule_type=? "
	}
	if hawkeyeId != 0 {
		args = append(args, hawkeyeId)
		sqlAdd += " AND hawkeye_id=? "
	}
	if name != "" {
		name = "%" + name + "%"
		args = append(args, name)
		sqlAdd += " AND name LIKE ? "
	}
	if species != "" {
		args = append(args, species)
		sqlAdd += " AND species LIKE ? "
	}
	if queryExprs != "" {
		queryExprs = "%" + queryExprs + "%"
		args = append(args, queryExprs)
		sqlAdd += " AND query_exprs LIKE ? "
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_apmAlertRuleCount, strings.Replace(sqlAdd, "AND", "WHERE", 1)), args...)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

func (d *Dao) ApmAlertRuleList(c context.Context, hawkeyeId int64, name, species, queryExprs string, ruleType int8, pn, ps int) (res []*apm.AlertRule, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if ruleType != -1 {
		args = append(args, ruleType)
		sqlAdd += " AND rule_type=? "
	}
	if hawkeyeId != 0 {
		args = append(args, hawkeyeId)
		sqlAdd += " AND hawkeye_id=? "
	}
	if name != "" {
		name = "%" + name + "%"
		args = append(args, name)
		sqlAdd += " AND name LIKE ? "
	}
	if species != "" {
		args = append(args, species)
		sqlAdd += " AND species LIKE ? "
	}
	if queryExprs != "" {
		queryExprs = "%" + queryExprs + "%"
		args = append(args, queryExprs)
		sqlAdd += " AND query_exprs LIKE ? "
	}
	sqlAdd += " ORDER BY ctime DESC "
	if pn != 0 && ps != 0 {
		args = append(args, (pn-1)*ps, ps)
		sqlAdd += " LIMIT ?,? "
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_apmAlertRuleList, strings.Replace(sqlAdd, "AND", "WHERE", 1)), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "ScanSlice error %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmAlertRuleAdd(tx *sql.Tx, hawkeyeId int64, name, triggerCond, species, queryExprs, operator string, ruleType int8) (err error) {
	_, err = tx.Exec(_apmAlertRuleAdd, hawkeyeId, name, triggerCond, species, queryExprs, operator, ruleType, name, triggerCond, species, queryExprs, operator, ruleType)
	return
}

func (d *Dao) TxApmAlertRuleMDUpdate(tx *sql.Tx, id int64, markdown string) (err error) {
	_, err = tx.Exec(_apmAlertRuleMDUpdate, markdown, id)
	return
}

func (d *Dao) TxApmAlertRuleDel(tx *sql.Tx, id int64) (err error) {
	_, err = tx.Exec(_apmAlertRuleDel, id)
	return
}

func (d *Dao) TxApmAlertRuleRelAdd(tx *sql.Tx, ruleId, adjustRuleId int64, operator string) (err error) {
	_, err = tx.Exec(_apmAlertRuleRelAdd, ruleId, adjustRuleId, operator, ruleId, adjustRuleId, operator)
	return
}

func (d *Dao) ApmAlertRuleRelByRuleIds(c context.Context, ruleIds []int64) (res []*apm.AlertRuleRel, err error) {
	var (
		sqls []string
		args []interface{}
	)
	if len(ruleIds) == 0 {
		return
	}
	for _, ruleId := range ruleIds {
		sqls = append(sqls, "?")
		args = append(args, ruleId)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_apmAlertRuleRelByRuleIds, strings.Join(sqls, ",")), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "ScanSlice error %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmAlertRuleRelByAdjustId(c context.Context, adjustId int64) (res *apm.AlertRuleRel, err error) {
	rows, err := d.db.Query(c, _apmAlertRuleRelByAdjustId, adjustId)
	if err != nil {
		return
	}
	defer rows.Close()
	var li []*apm.AlertRuleRel
	if err = xsql.ScanSlice(rows, &li); err != nil {
		log.Error("ScanSlice error %v", err)
		return
	}
	if len(li) == 1 {
		return li[0], err
	}
	err = rows.Err()
	return nil, err
}

func (d *Dao) ApmAlertById(c context.Context, id int64) (res *apm.Alert, err error) {
	rows, err := d.db.Query(c, _apmAlert, id)
	if err != nil {
		return
	}
	defer rows.Close()
	var li []*apm.Alert
	if err = xsql.ScanSlice(rows, &li); err != nil {
		log.Error("ScanSlice error %v", err)
		return
	}
	if len(li) == 1 {
		return li[0], err
	}
	err = rows.Err()
	return nil, err
}

func (d *Dao) ApmAlertByMd5(c context.Context, alertMd5 string) (res *apm.Alert, err error) {
	rows, err := d.db.Query(c, _apmAlertByMd5, alertMd5)
	if err != nil {
		return
	}
	defer rows.Close()
	var li []*apm.Alert
	if err = xsql.ScanSlice(rows, &li); err != nil {
		log.Error("ScanSlice error %v", err)
		return
	}
	if len(li) == 1 {
		return li[0], err
	}
	err = rows.Err()
	return nil, err
}

func (d *Dao) ApmAlertCount(c context.Context, appKey string, env apm.Env, ruleIds []int64, alertType, status int8, alertMd5 string, startTime, endTime int64) (count int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if appKey != "" {
		sqlAdd += " AND app_key=? "
		args = append(args, appKey)
	}
	if env != "" {
		sqlAdd += " AND env=? "
		args = append(args, env)
	}
	if len(ruleIds) > 0 {
		var sqls []string
		for _, ruleId := range ruleIds {
			args = append(args, ruleId)
			sqls = append(sqls, "?")
		}
		sqlAdd += fmt.Sprintf(" AND rule_id IN (%s) ", strings.Join(sqls, ","))
	}
	if alertType != 0 {
		sqlAdd += " AND alert_type=? "
		args = append(args, alertType)
	}
	if status != 0 {
		sqlAdd += " AND alert_status=? "
		args = append(args, status)
	}
	if alertMd5 != "" {
		sqlAdd += " AND alert_md5=? "
		args = append(args, alertMd5)
	}
	if startTime != 0 {
		sqlAdd += " AND start_time > FROM_UNIXTIME(?,'%Y-%m-%d %H:%i:%s') "
		args = append(args, startTime)
	}
	if endTime != 0 {
		sqlAdd += " AND start_time < FROM_UNIXTIME(?,'%Y-%m-%d %H:%i:%s') "
		args = append(args, endTime)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_apmAlertCount, strings.Replace(sqlAdd, "AND", "WHERE", 1)), args...)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

func (d *Dao) ApmAlertList(c context.Context, appKey string, env apm.Env, ruleIds []int64, alertType, status int8, alertMd5 string, startTime, endTime int64, pn, ps int) (res []*apm.Alert, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if appKey != "" {
		sqlAdd += " AND app_key=? "
		args = append(args, appKey)
	}
	if env != "" {
		sqlAdd += " AND env=? "
		args = append(args, env)
	}
	if len(ruleIds) > 0 {
		var sqls []string
		for _, ruleId := range ruleIds {
			args = append(args, ruleId)
			sqls = append(sqls, "?")
		}
		sqlAdd += fmt.Sprintf(" AND rule_id IN (%s) ", strings.Join(sqls, ","))
	}
	if alertType != 0 {
		sqlAdd += " AND alert_type=? "
		args = append(args, alertType)
	}
	if status != 0 {
		sqlAdd += " AND alert_status=? "
		args = append(args, status)
	}
	if alertMd5 != "" {
		sqlAdd += " AND alert_md5=? "
		args = append(args, alertMd5)
	}
	if startTime != 0 {
		sqlAdd += " AND start_time > FROM_UNIXTIME(?,'%Y-%m-%d %H:%i:%s') "
		args = append(args, startTime)
	}
	if endTime != 0 {
		sqlAdd += " AND start_time < FROM_UNIXTIME(?,'%Y-%m-%d %H:%i:%s') "
		args = append(args, endTime)
	}
	sqlAdd += " ORDER BY ctime DESC "
	if pn != 0 && ps != 0 {
		args = append(args, (pn-1)*ps, ps)
		sqlAdd += " LIMIT ?,? "
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_apmAlertList, strings.Replace(sqlAdd, "AND", "WHERE", 1)), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "ScanSlice error %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmAlertAdd(tx *sql.Tx, ruleId int64, appKey string, env apm.Env, duration int64, alertMd5, labels, operator string, alertType, status int8, triggerValue interface{}, startTime time.Time) (err error) {
	_, err = tx.Exec(_apmAlertAdd, ruleId, appKey, env, alertMd5, status, duration, labels, triggerValue, alertType, operator, startTime, ruleId, appKey, env, alertMd5, status, duration, labels, triggerValue, operator, startTime)
	return
}

func (d *Dao) TxApmAlertUpdate(tx *sql.Tx, alertType, status int8, description, operator string, duration, id int64) (err error) {
	_, err = tx.Exec(_apmAlertUpdate, alertType, status, description, operator, duration, id)
	return
}

func (d *Dao) ApmAlertReasonConfig(c context.Context, ruleId int64) (res []*apm.AlertReasonConfig, err error) {
	rows, err := d.db.Query(c, _apmAlertReasonConfig, ruleId)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "ScanSlice error %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmAlertReasonConfigAdd(tx *sql.Tx, ruleId, eventId int64, queryType, customQuerySql, queryCondition, fields, description, operator string) (err error) {
	querySql, err := getColumnSQL(queryType, customQuerySql)
	if err != nil {
		return
	}
	_, err = tx.Exec(_apmAlertReasonConfigAdd, ruleId, eventId, queryType, querySql, queryCondition, fields, description, operator)
	return
}

func (d *Dao) TxApmAlertReasonConfigUpdate(tx *sql.Tx, id, eventId int64, queryType, customQuerySql, queryCondition, fields, description, operator string) (err error) {
	querySql, err := getColumnSQL(queryType, customQuerySql)
	if err != nil {
		return
	}
	_, err = tx.Exec(_apmAlertReasonConfigUpdate, eventId, queryType, querySql, queryCondition, fields, description, operator, id)
	return
}

func (d *Dao) TxApmAlertReasonConfigDelete(tx *sql.Tx, id int64) (err error) {
	_, err = tx.Exec(_apmAlertReasonConfigDelete, id)
	return
}

func (d *Dao) ApmEventMonitorNotifyConfig(c context.Context, eventId int64, appKey string) (res *apm.EventMonitorNotifyConfig, err error) {
	rows, err := d.db.Query(c, _eventMonitorNotifyConfig, eventId, appKey)
	if err != nil {
		return
	}
	defer rows.Close()
	var list []*apm.EventMonitorNotifyConfig
	if err = xsql.ScanSlice(rows, &list); err != nil {
		log.Error("ScanSlice error %v", err)
		return
	}
	if len(list) == 1 {
		return list[0], err
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmEventMonitorNotifyConfigCount(c context.Context, eventId int64, appKey string, isNotify, isMute int8) (count int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if eventId != 0 {
		sqlAdd += " AND event_id=? "
		args = append(args, eventId)
	}
	if appKey != "" {
		sqlAdd += " AND app_key=? "
		args = append(args, appKey)
	}
	if isNotify != 0 {
		sqlAdd += " AND is_notify=? "
		args = append(args, isNotify)
	}
	if isMute != 0 {
		sqlAdd += " AND is_mute=? "
		args = append(args, isMute)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_eventMonitorNotifyConfigCount, strings.Replace(sqlAdd, "AND", "WHERE", 1)), args...)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

func (d *Dao) ApmEventMonitorNotifyConfigList(c context.Context, eventId int64, appKey string, isNotify, isMute int8, pn, ps int) (res []*apm.EventMonitorNotifyConfig, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if eventId != 0 {
		sqlAdd += " AND event_id=? "
		args = append(args, eventId)
	}
	if appKey != "" {
		sqlAdd += " AND app_key=? "
		args = append(args, appKey)
	}
	if isNotify != 0 {
		sqlAdd += " AND is_notify=? "
		args = append(args, isNotify)
	}
	if isMute != 0 {
		sqlAdd += " AND is_mute=? "
		args = append(args, isMute)
	}
	sqlAdd += " ORDER BY id DESC "
	if pn != 0 && ps != 0 {
		args = append(args, (pn-1)*ps, ps)
		sqlAdd += " LIMIT ?,? "
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_eventMonitorNotifyConfigList, strings.Replace(sqlAdd, "AND", "WHERE", 1)), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "ScanSlice error %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) TxApmEventMonitorNotifyConfigSet(tx *sql.Tx, eventId int64, appKey string, isNotify, isMute int8, muteStartTime, muteEndTime time.Time, operator string) (err error) {
	_, err = tx.Exec(_eventMonitorNotifyConfigSet, eventId, appKey, isNotify, isMute, muteStartTime, muteEndTime, operator)
	return
}

func (d *Dao) TxApmEventMonitorNotifyConfigBatchSet(tx *sql.Tx, configs []*apm.EventMonitorNotifyConfig) (err error) {
	if len(configs) < 1 {
		log.Warn("configs is empty")
		return
	}
	var (
		sqls []string
		args []interface{}
	)
	for _, config := range configs {
		sqls = append(sqls, "(?,?,?,?,?,?,?)")
		args = append(args, config.EventId, config.AppKey, config.IsNotify, config.IsMute, config.MuteStartTime, config.MuteEndTime, config.Operator)
	}
	_, err = tx.Exec(fmt.Sprintf(_eventMonitorNotifyConfigBatchSet, strings.Join(sqls, ",")), args...)
	return
}

func (d *Dao) TxApmEventMonitorNotifyMuteUpdate(tx *sql.Tx, ids []int64) (err error) {
	var (
		sqls []string
		args []interface{}
	)
	if len(ids) < 1 {
		log.Warn("ids is empty")
		return
	}
	for _, id := range ids {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	_, err = tx.Exec(fmt.Sprintf(_eventMonitorNotifyMuteUpdate, strings.Join(sqls, ",")), args...)
	return
}

/**********************************billions平台 埋点注册与字段添加**********************************/

// ApmBillionsEventAdd 日志平台监控事件增加
func (d *Dao) ApmBillionsEventAdd(c context.Context, billionsConf *conf.Billions, name, description, operator string) (err error) {
	var billionsEvent = &apm.BillionsEvent{
		TreeID:           billionsConf.TreeID,
		AppID:            name,
		AppName:          description,
		ServicePrincipal: operator,
		DeployLocations:  billionsConf.DeployLocations,
	}
	queryBytes, _ := json.Marshal(&billionsEvent)
	payload := strings.NewReader(string(queryBytes))
	toRequestUrl := billionsConf.Host + billionsConf.Dir + billionsConf.AutoAdd
	req, _ := http.NewRequest(http.MethodPost, toRequestUrl, payload)
	req.Header.Add("content-type", _contentTypeJson)
	req.Header.Add("X-Authorization-Token", billionsConf.AuthorizationToken)
	req.Header.Add("username", _billionsUsername)
	var re struct {
		Msg  string `json:"message"`
		Code int64  `json:"code"`
		Data bool   `json:"data"`
	}
	if err = d.httpClient.Do(c, req, &re); err != nil {
		log.Error("d.httpClient.Do err:%v", err)
		return
	}
	if !re.Data && !strings.Contains(re.Msg, "Duplicate entry") {
		err = d.ExternalErrorc(c, apm.BillionsOkStatus, re.Code, fmt.Sprintf("billions event add %v,\n[query_content]:%+v", re.Msg, billionsEvent))
	} else {
		log.Warn("billions event add duplicate entry %+v", billionsEvent)
	}
	return err
}

// ApmBillionsAddEventField 日志平台扩展字段增加或修改
func (d *Dao) ApmBillionsAddEventField(c context.Context, billionsConf *conf.Billions, mapping *apm.BillionsEventFieldMapping) (err error) {
	log.Warnc(c, "ApmBillionsAddEventField mapping:\n%+v", mapping)
	queryBytes, _ := json.Marshal(mapping)
	payload := strings.NewReader(string(queryBytes))
	toRequestUrl := billionsConf.Host + billionsConf.Dir + billionsConf.MappingUpdate
	req, _ := http.NewRequest(http.MethodPost, toRequestUrl, payload)
	req.Header.Add("content-type", _contentTypeJson)
	req.Header.Add("X-Authorization-Token", billionsConf.AuthorizationToken)
	req.Header.Add("username", _billionsUsername)
	var re struct {
		Msg  string `json:"message"`
		Code int64  `json:"code"`
		Data int64  `json:"data"`
	}
	if err = d.httpClient.Do(c, req, &re); err != nil {
		log.Errorc(c, "d.httpClient.Do err:%v", err)
		return
	}
	err = d.ExternalErrorc(c, apm.BillionsOkStatus, re.Code, fmt.Sprintf("billions event field add %v,\n[query_content]:%+v", re.Msg, mapping))
	return err
}

/**********************************billions平台 日志信息查询**********************************/

// BillionsLifecycle 日志平台生命周期（容量、索引打开时间、日志留存时间）
func (d *Dao) BillionsLifecycle(c context.Context, pn, ps int64, billionsConf *conf.Billions) (res map[string]*apm.BillionsLifecycle, err error) {
	var req *http.Request
	toRequestUrl := fmt.Sprintf("%v%v%v?cluster=%v&page=%v&size=%v", billionsConf.Host, billionsConf.Dir, billionsConf.Lifecycle, billionsConf.Cluster, pn, ps)
	if req, err = http.NewRequest(http.MethodGet, toRequestUrl, nil); err != nil {
		return
	}
	req.Header.Add("X-Authorization-Token", billionsConf.AuthorizationToken)
	req.Header.Add("username", _billionsUsername)
	var re struct {
		Data struct {
			Lifecycle []*apm.BillionsLifecycle `json:"list"`
		} `json:"data"`
	}
	if err = d.httpClient.Do(c, req, &re); err != nil {
		return
	}
	res = make(map[string]*apm.BillionsLifecycle)
	for _, data := range re.Data.Lifecycle {
		res[data.AppId] = data
	}
	return res, nil
}

// BillionsLogCount 日志平台log数量
func (d *Dao) BillionsLogCount(c context.Context, eventName, queryBody string, esProxyConf *conf.ElasticsearchProxy) (count int64, err error) {
	var (
		queryByte []byte
		req       *http.Request
	)
	queryByte = []byte(queryBody)
	payload := bytes.NewBuffer(queryByte)
	toRequestUrl := fmt.Sprintf("%v%v/billions-%v-@*%v?cluster=%v", esProxyConf.Host, esProxyConf.Dir, eventName, esProxyConf.Search, esProxyConf.Cluster)
	if req, err = http.NewRequest(http.MethodPost, toRequestUrl, payload); err != nil {
		return
	}
	req.Header.Add("Content-Type", _contentTypeJson)
	req.Header.Add("Appid", _esAppId)
	req.Header.Add("Appkey", esProxyConf.Token)
	var re struct {
		Hit struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
		} `json:"hits"`
	}
	if err = d.httpClient.Do(c, req, &re); err != nil {
		return
	}
	return re.Hit.Total.Value, nil
}

/**********************************billions平台 告警规则操作**********************************/

// BillionsAlertAdd 告警规则添加
func (d *Dao) BillionsAlertAdd(c context.Context, addReq *apm.EventAlertBillionsReq, billionsConf *conf.BillionsAlert) (id int64, err error) {
	addBytes, err := json.MarshalIndent(addReq, "", " ")
	if err != nil {
		log.Errorc(c, "MarshalIndent error %v", err)
		return
	}
	fmt.Println(string(addBytes))
	payload := strings.NewReader(string(addBytes))
	toRequestUrl := billionsConf.Host + billionsConf.Dir + billionsConf.Alert
	req, _ := http.NewRequest(http.MethodPost, toRequestUrl, payload)
	req.Header.Add("content-type", _contentTypeJson)
	req.Header.Add("X-Authorization-Token", billionsConf.Token)
	req.Header.Add("username", _billionsUsername)
	var re struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Id int64 `json:"id"`
		} `json:"data"`
	}
	if err = d.httpClient.Do(c, req, &re); err != nil {
		log.Errorc(c, "d.httpClient.Do error %v", err)
		return
	}
	err = d.ExternalErrorc(c, apm.BillionsOkStatus, re.Code, fmt.Sprintf("billions alert add %v,\n[query_content]:%+v", re.Message, addReq))
	return re.Data.Id, err
}

// BillionsAlertUpdate 告警规则更新
func (d *Dao) BillionsAlertUpdate(c context.Context, id int64, updateReq *apm.EventAlertBillionsReq, billionsConf *conf.BillionsAlert) (err error) {
	updateBytes, err := json.MarshalIndent(updateReq, "", " ")
	if err != nil {
		log.Errorc(c, "MarshalIndent error %v", err)
		return
	}
	fmt.Println(string(updateBytes))
	payload := strings.NewReader(string(updateBytes))
	toRequestUrl := fmt.Sprintf("%s%s%s/%d", billionsConf.Host, billionsConf.Dir, billionsConf.Alert, id)
	req, _ := http.NewRequest(http.MethodPut, toRequestUrl, payload)
	req.Header.Add("content-type", _contentTypeJson)
	req.Header.Add("X-Authorization-Token", billionsConf.Token)
	req.Header.Add("username", _billionsUsername)
	var re struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
	}
	if err = d.httpClient.Do(c, req, &re); err != nil {
		log.Errorc(c, "d.httpClient.Do error %v", err)
		return
	}
	err = d.ExternalErrorc(c, apm.BillionsOkStatus, re.Code, fmt.Sprintf("billions alert update error %v,\n[query_content]:%+v", re.Message, updateReq))
	return err
}

// BillionsAlertRuleOpt 告警规则操作:删除和启用开关
func (d *Dao) BillionsAlertRuleOpt(c context.Context, alertOpt *apm.BillionsAlertOpt, billionsConf *conf.BillionsAlert) (err error) {
	alertOptBytes, err := json.MarshalIndent(alertOpt, "", " ")
	if err != nil {
		log.Errorc(c, "MarshalIndent error %v", err)
		return
	}
	payload := strings.NewReader(string(alertOptBytes))
	toRequestUrl := billionsConf.Host + billionsConf.Dir + billionsConf.RuleOpt
	req, _ := http.NewRequest(http.MethodPost, toRequestUrl, payload)
	req.Header.Add("content-type", _contentTypeJson)
	req.Header.Add("X-Authorization-Token", billionsConf.Token)
	req.Header.Add("username", _billionsUsername)
	var re struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
	}
	if err = d.httpClient.Do(c, req, &re); err != nil {
		log.Errorc(c, "d.httpClient.Do error %v", err)
		return
	}
	err = d.ExternalErrorc(c, apm.BillionsOkStatus, re.Code, fmt.Sprintf("billions alert rule opt error %vv,\n[query_content]:%+v", re.Message, alertOpt))
	return err
}

/*********************************数据平台 技术埋点操作**********************************/

// ApmDatacenterEventAdd 数据平台监控事件增加
func (d *Dao) ApmDatacenterEventAdd(c context.Context, addReq *apm.DatacenterEvent, datacenterConf *conf.Datacenter) (lastID int64, err error) {
	queryBytes, err := json.MarshalIndent(addReq, "", " ")
	if err != nil {
		log.Errorc(c, "MarshalIndent error %v", err)
		return
	}
	payload := strings.NewReader(string(queryBytes))
	toRequestUrl := datacenterConf.Host + datacenterConf.Dir + datacenterConf.Add
	req, _ := http.NewRequest(http.MethodPost, toRequestUrl, payload)
	req.Header.Add("content-type", _contentTypeJson)
	var re struct {
		Msg  string `json:"msg"`
		Code int64  `json:"code"`
		ID   int64  `json:"id"`
	}
	if err = d.httpClient.Do(c, req, &re); err != nil {
		log.Errorc(c, "d.httpClient.Do err:%v", err)
		return
	}
	err = d.ExternalErrorc(c, apm.DatacenterOkStatus, re.Code, fmt.Sprintf("datacenter event add %v,\n[query_content]:%+v", re.Msg, addReq))
	return re.ID, err
}

// ApmDatacenterEventUpdate 数据平台监控事件修改
func (d *Dao) ApmDatacenterEventUpdate(c context.Context, updateReq *apm.DatacenterEvent, datacenterConf *conf.Datacenter) (err error) {
	queryBytes, err := json.MarshalIndent(updateReq, "", " ")
	if err != nil {
		log.Errorc(c, "MarshalIndent error %v", err)
		return
	}
	payload := strings.NewReader(string(queryBytes))
	toRequestUrl := datacenterConf.Host + datacenterConf.Dir + datacenterConf.Update
	req, _ := http.NewRequest(http.MethodPost, toRequestUrl, payload)
	req.Header.Add("content-type", _contentTypeJson)
	var re struct {
		Msg  string `json:"msg"`
		Code int64  `json:"code"`
	}
	if err = d.httpClient.Do(c, req, &re); err != nil {
		log.Errorc(c, "d.httpClient.Do err:%v", err)
		return
	}
	err = d.ExternalErrorc(c, apm.DatacenterOkStatus, re.Code, fmt.Sprintf("datacenter event update %v,\n[query_content]:%+v", re.Msg, updateReq))
	return err
}

func (d *Dao) ApmDataCenterCKTableCreate(c context.Context, tabData *apm.CKTableCreateData, conf *conf.Datacenter) (err error) {
	dataByte, err := json.Marshal(tabData)
	if err != nil {
		log.Errorc(c, "ApmDataCenterCreateTable error %v", err)
		return
	}
	dataRe := strings.ReplaceAll(string(dataByte), "`", "")
	sb := strings.Builder{}
	sb.WriteString(_metaAppId)
	sb.WriteString(_ckGroupName)
	sb.WriteString(_ckCreateApiName)
	sb.WriteString(conf.OpenAPI.Account)
	sb.WriteString(_ckRequestId)
	sb.WriteString(dataRe)
	sb.WriteString(conf.OpenAPI.SecretKey)
	signature := utils.MD5HashString(sb.String())
	query := &apm.DatacenterOpenAPI{
		Account:   conf.OpenAPI.Account,
		APIName:   "CreateTableWithCode",
		AppId:     "datacenter.keeper.keeper",
		Data:      dataRe,
		GroupName: "KeeperMultiTable",
		RequestId: "keeper_CreateTable",
		Signature: signature,
	}
	queryBytes, _ := json.MarshalIndent(query, "", " ")
	payload := strings.NewReader(string(queryBytes))
	toRequestUrl := fmt.Sprintf("%s%s", conf.Host, conf.OpenAPI.Dir)
	req, _ := http.NewRequest(http.MethodPost, toRequestUrl, payload)
	req.Header.Add("Content-Type", _contentTypeJson)
	var re struct {
		Code      int64  `json:"code"`
		Message   string `json:"message"`
		Data      string `json:"data"`
		RequestId string `json:"requestId"`
		TraceId   string `json:"traceId"`
	}
	if err = d.httpClient.Do(c, req, &re); err != nil {
		log.Error("d.httpClient.Do err:%v", err)
		return
	}
	err = d.ExternalErrorc(c, apm.DatacenterOkStatus, re.Code, fmt.Sprintf("datacenter clickhouse table create %v,\n[query_content]:%+v", re.Message, string(queryBytes)))
	return
}

func (d *Dao) ApmEventSampleRateAdd(ctx context.Context, appId int64, eventId, eventName, logId string, rate float64) (id int64, err error) {
	// insert on duplicate key update
	row, err := d.db.Exec(ctx, _evenSampleRateAdd, appId, eventId, eventName, rate, logId)
	if err != nil {
		return
	}
	return row.LastInsertId()
}

func (d *Dao) ApmEventSampleRateAppAdd(ctx context.Context, appKey, eventId, eventName, logId string, rate float64) (id int64, err error) {
	// insert on duplicate key update
	row, err := d.db.Exec(ctx, _evenSampleRateAppAdd, appKey, eventId, eventName, rate, logId)
	if err != nil {
		return
	}
	return row.LastInsertId()
}

func (d *Dao) ApmEventSampleRateDelete(ctx context.Context, items []*apm.DeleteSampleItem) (id int64, err error) {
	if len(items) == 0 {
		return
	}
	var (
		args   []interface{}
		sqlAdd []string
	)
	for _, v := range items {
		sqlAdd = append(sqlAdd, "(?,?)")
		args = append(args, v.DatacenterAppId, v.EventId)
	}
	row, err := d.db.Exec(ctx, fmt.Sprintf(_eventSampleBatchDelete, strings.Join(sqlAdd, ",")), args...)
	if err != nil {
		return
	}
	return row.RowsAffected()
}

func (d *Dao) ApmEventSampleRateAppDelete(ctx context.Context, items []*apm.DeleteSampleItem) (id int64, err error) {
	if len(items) == 0 {
		return
	}
	var (
		args   []interface{}
		sqlAdd []string
	)
	for _, v := range items {
		sqlAdd = append(sqlAdd, "(?,?)")
		args = append(args, v.AppKey, v.EventId)
	}
	row, err := d.db.Exec(ctx, fmt.Sprintf(_eventSampleAppBatchDelete, strings.Join(sqlAdd, ",")), args...)
	if err != nil {
		return
	}
	return row.RowsAffected()
}

func (d *Dao) SelectApmEventSampleRate(ctx context.Context, dataCenterAppId int64, appKey, eventId string, logId []string) (res []*apm.EventSampleRate, err error) {
	// 列表页 条件查询
	var (
		rows   *sql.Rows
		sqlAdd string
		args   []interface{}
	)
	args = append(args, dataCenterAppId, appKey)
	if len(eventId) != 0 {
		sqlAdd = "AND event_id LIKE ?"
		args = append(args, "%"+eventId+"%")
	}
	if len(logId) != 0 {
		var logIdStr []string
		for _, v := range logId {
			logIdStr = append(logIdStr, "?")
			args = append(args, v)
		}
		sqlAdd += fmt.Sprintf(" AND log_id IN (%s)", strings.Join(logIdStr, ","))
	}
	if rows, err = d.db.Query(ctx, fmt.Sprintf(_eventSampleList, sqlAdd), args...); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &res); err != nil {
		return
	}
	return
}

func (d *Dao) SelectApmEventSampleRateApp(ctx context.Context, appKey, eventId string, logId []string) (res []*apm.EventSampleRateApp, err error) {
	// 列表页 条件查询
	var (
		rows   *sql.Rows
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if len(eventId) != 0 {
		sqlAdd = "AND event_id LIKE ?"
		args = append(args, "%"+eventId+"%")
	}
	if len(logId) != 0 {
		var logIdStr []string
		for _, v := range logId {
			logIdStr = append(logIdStr, "?")
			args = append(args, v)
		}
		sqlAdd += fmt.Sprintf(" AND log_id IN (%s)", strings.Join(logIdStr, ","))
	}
	if rows, err = d.db.Query(ctx, fmt.Sprintf(_eventSampleAppList, sqlAdd), args...); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &res); err != nil {
		return
	}
	return
}
