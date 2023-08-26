// update all status as 4(approve passed) if status is 2
update es_archives set is_deleted = 4 where is_deleted = 2;