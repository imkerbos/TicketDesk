-- remove_deleted_at.sql
-- 目的：
-- 1) 清理所有表中已软删除的数据（deleted_at IS NOT NULL）
-- 2) 删除所有表的 deleted_at 字段
--
-- 适用：MySQL 8+
-- 注意：请在业务低峰执行，并先做好备份。

DELIMITER $$

DROP PROCEDURE IF EXISTS remove_deleted_at_columns $$

CREATE PROCEDURE remove_deleted_at_columns()
BEGIN
  DECLARE done INT DEFAULT 0;
  DECLARE tbl VARCHAR(128);

  DECLARE cur CURSOR FOR
    SELECT DISTINCT c.table_name
    FROM information_schema.columns c
    WHERE c.table_schema = DATABASE()
      AND c.column_name = 'deleted_at';

  DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1;

  OPEN cur;

  table_loop: LOOP
    FETCH cur INTO tbl;
    IF done = 1 THEN
      LEAVE table_loop;
    END IF;

    -- 先清理该表所有软删除记录
    SET @cleanup_sql = CONCAT('DELETE FROM `', tbl, '` WHERE `deleted_at` IS NOT NULL');
    PREPARE stmt_cleanup FROM @cleanup_sql;
    EXECUTE stmt_cleanup;
    DEALLOCATE PREPARE stmt_cleanup;

    -- 再删除 deleted_at 列
    SET @drop_sql = CONCAT('ALTER TABLE `', tbl, '` DROP COLUMN `deleted_at`');
    PREPARE stmt_drop FROM @drop_sql;
    EXECUTE stmt_drop;
    DEALLOCATE PREPARE stmt_drop;
  END LOOP;

  CLOSE cur;
END $$

DELIMITER ;

CALL remove_deleted_at_columns();
DROP PROCEDURE IF EXISTS remove_deleted_at_columns;
