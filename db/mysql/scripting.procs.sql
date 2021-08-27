DELIMITER $$
DROP FUNCTION IF EXISTS `ttab`.`fScriptScrubColumnType` $$

CREATE FUNCTION `ttab`.`fScriptScrubColumnType`(column_type_ VARCHAR(255)) RETURNS VARCHAR(255) CHARSET utf8
    DETERMINISTIC
BEGIN

	IF LOCATE('VARCHAR',column_type_) > 0 THEN
		RETURN CONCAT(' (',REPLACE(REPLACE(REPLACE(column_type_,')',''),'(',' '),'varchar','vc'),')');
	END IF;
	
	IF LOCATE('UNSIGNED',column_type_) > 0 THEN
		IF LOCATE('BIGINT',column_type_) > 0 THEN
			RETURN ' u64';
		END IF;
		IF LOCATE('SMALLINT',column_type_) > 0 THEN
			RETURN ' u16';
		END IF;
		RETURN ' u32';
	END IF;
	
	IF LOCATE('BIGINT',column_type_) > 0 THEN
		RETURN ' i64';
	END IF;
	
	IF LOCATE('SMALLINT',column_type_) > 0 THEN
		RETURN ' i16';
	END IF;
	
	IF LOCATE('INT',column_type_) > 0 THEN
		RETURN ' i32';
	END IF;	
	
	IF LOCATE('BIT',column_type_) > 0 THEN
		RETURN CONCAT(' (',REPLACE(REPLACE(REPLACE(column_type_,')',''),'(',' '),'BIT','bit'),')');
	END IF;	

	IF LOCATE('TEXT',column_type_) > 0 THEN
		RETURN ' txt';
	END IF;	
	
	IF LOCATE('TIMESTAMP',column_type_) > 0 THEN
		RETURN ' ts';
	END IF;	
	
	IF LOCATE('DATETIME',column_type_) > 0 THEN
		RETURN ' dt';
	END IF;	
	
	IF LOCATE('TINYINT',column_type_) > 0 THEN
		RETURN ' bool';
	END IF;	
	
	IF LOCATE('VARBINARY',column_type_) > 0 THEN
		RETURN CONCAT(' (',REPLACE(REPLACE(REPLACE(column_type_,')',''),'(',' '),'varbinary','vb'),')');
	END IF;	
	
	IF LOCATE('ENUM',column_type_) > 0 THEN
		RETURN CONCAT(' (',REPLACE(REPLACE(REPLACE(column_type_,')','"'),'(',' "'),'enum','enum'),')');
	END IF;	
	
	IF LOCATE('CHAR',column_type_) > 0 THEN
		RETURN CONCAT(' (',REPLACE(REPLACE(REPLACE(column_type_,')',''),'(',' '),'char','ch'),')');
	END IF;	
	
	IF LOCATE('FLOAT',column_type_) > 0 THEN
		RETURN ' f32';
	END IF;	
	
	IF LOCATE('DECIMAL',column_type_) > 0 THEN
		RETURN CONCAT(' (',REPLACE(REPLACE(REPLACE(column_type_,')','"'),'(','"'),'decimal','dec'),')');
	END IF;	
	
	IF LOCATE('BINARY',column_type_) > 0 THEN
		RETURN CONCAT(' (',REPLACE(REPLACE(REPLACE(column_type_,')',''),'(',' '),'binary','bin'),')');
	END IF;	
	
	IF LOCATE('BLOB',column_type_) > 0 THEN
		RETURN ' blob';
	END IF;
	
	IF LOCATE('DOUBLE',column_type_) > 0 THEN
		RETURN ' f64';
	END IF;
	
	RETURN column_type_;
END$$

DELIMITER ;

DELIMITER $$
DROP FUNCTION IF EXISTS `ttab`.`fScriptScriptIsNullable` $$

CREATE FUNCTION `ttab`.`fScriptScriptIsNullable`(nullable_ VARCHAR(255)) RETURNS VARCHAR(255) CHARSET utf8
    DETERMINISTIC
BEGIN

	IF LOCATE('NO',nullable_) > 0 THEN
		RETURN ' notnull';
	END IF;
	
	RETURN ' null';

END$$

DELIMITER ;

DELIMITER $$
DROP FUNCTION IF EXISTS `ttab`.`fScriptScrubDefault` $$

CREATE FUNCTION `ttab`.`fScriptScrubDefault`(def_ VARCHAR(255), type_ VARCHAR(255)) RETURNS VARCHAR(255) CHARSET utf8
    DETERMINISTIC
BEGIN

	IF LOCATE('_none_',def_) > 0 THEN
		RETURN '';
	END IF;
	
	IF LOCATE('0000-00',def_) > 0 THEN
		RETURN '';
	END IF;
	
	IF LOCATE('0000-00',def_) > 0 THEN
		RETURN '';
	END IF;
	
	IF LOCATE('vc',type_) > 0 OR LOCATE('ch',type_) > 0 OR LOCATE('txt',type_) > 0 OR LOCATE('enum',type_) > 0 THEN
		RETURN CONCAT(' default \"\'',def_,'\'\"');
	END IF;
	
	IF LOCATE('bit',type_) > 0 THEN
		RETURN CONCAT(" default ", REPLACE(REPLACE(def_,"b'",''),"'",""));
	END IF;
	
	RETURN CONCAT(" default ",def_);
END$$

DELIMITER ;

DELIMITER $$
DROP FUNCTION IF EXISTS `ttab`.`fScriptScrubExtra` $$

CREATE FUNCTION `ttab`.`fScriptScrubExtra`(extra_ VARCHAR(255)) RETURNS VARCHAR(255) CHARSET utf8
    DETERMINISTIC
BEGIN
	IF LOCATE('AUTO_INCREMENT',extra_) > 0 THEN
		RETURN ' ai';
	END IF;
	RETURN '';
END$$

DELIMITER ;

DELIMITER $$
DROP PROCEDURE IF EXISTS `ttab`.`pScriptTableToTab` $$
CREATE PROCEDURE `ttab`.`pScriptTableToTab`(schema_ VARCHAR(100), table_ VARCHAR(100))
proc:BEGIN
	DECLARE OLDLEN INT;
	
	DROP TEMPORARY TABLE IF EXISTS _lines;
	CREATE TEMPORARY TABLE _lines (line TEXT);
	
	SELECT @@group_concat_max_len INTO OLDLEN;
	
	SET @@group_concat_max_len = 65536;
	
	-- leader
	INSERT INTO _lines VALUES (CONCAT("tab ",table_," innodb"));
	INSERT INTO _lines VALUES (CONCAT("	,(dbname ",schema_,")"));
	INSERT INTO _lines VALUES (CONCAT("	,(autoinc 10000)"));
	INSERT INTO _lines VALUES (CONCAT("	,(charset utf8mb4)"));
	
	-- rows
	insert into _lines
	select CONCAT('	,(col '
		,column_name
		,fScriptScrubColumnType(replace(column_type,' ','_'))
		,fScriptScriptIsNullable(is_nullable)
		,fScriptScrubDefault(coalesce(column_default,'_none_'),fScriptScrubColumnType(replace(column_type,' ','_')))
		,fScriptScrubExtra(coalesce(extra,'_none_'))
		,')')
	from information_schema.columns
	where table_schema = schema_
		AND table_name = table_
	order by ordinal_position asc;
	
	-- primary indexes
	insert into _lines
	select concat('	,(primary ',line,')')
	from (
		select CONCAT('(',group_concat(CONCAT_WS(' '
			,CASE COALESCE(collation,'A') WHEN 'A' THEN 'asc' ELSE 'desc' END
			,column_name
		) SEPARATOR ') ('),')') as 'line'
		from information_schema.statistics
		where table_schema = schema_
			and table_name = table_
			and index_name = "PRIMARY"
		group by index_name
		order by seq_in_index
	) tab;
	
	-- unique indexes
	insert into _lines
	select concat('	,(unique ',line,')')
	from (
		select CONCAT('(',group_concat(CONCAT_WS(' '
			,CASE COALESCE(collation,'A') WHEN 'A' THEN 'asc' ELSE 'desc' END
			,column_name
		) SEPARATOR ') ('),')') as 'line'
		from information_schema.statistics
		where table_schema = schema_
			and table_name = table_
			and non_unique = 0
			and index_name != "PRIMARY"
		group by index_name
		order by index_name,seq_in_index
	) tab;

	-- normal indexes
	insert into _lines
	select concat('	,(key ',line,')')
	from (
		select CONCAT('(',group_concat(CONCAT_WS(' '
			,CASE COALESCE(collation,'A') WHEN 'A' THEN 'asc' ELSE 'desc' END
			,column_name
		) SEPARATOR ') ('),')') as 'line'
		from information_schema.statistics
		where table_schema = schema_
			and table_name = table_
			and non_unique = 1
			and index_name != "PRIMARY"
		group by index_name
		order by index_name,seq_in_index
	) tab;
	
	
	select GROUP_CONCAT(line SEPARATOR '\n') as Script from _lines;	
	
	SET @@group_concat_max_len = OLDLEN;
	
	DROP TEMPORARY TABLE _lines;
	
	
END$$
DELIMITER ;

DELIMITER $$
DROP FUNCTION IF EXISTS `ttab`.`fScriptColumnDataSelect` $$

CREATE FUNCTION `ttab`.`fScriptColumnDataSelect`(coln_ VARCHAR(255), type_ VARCHAR(255)) RETURNS VARCHAR(255) CHARSET utf8
    DETERMINISTIC
BEGIN
	DECLARE BUF VARCHAR(255);
	
	IF LOCATE('TEXT',type_) > 0 OR LOCATE('BLOB',type_) > 0 OR LOCATE('BINARY',type_) > 0 OR LOCATE('CHAR',type_) > 0 OR LOCATE('ENUM',type_) > 0 OR LOCATE('TIME',type_) > 0 OR LOCATE('DATE',type_) > 0 THEN
		SET BUF = CONCAT('REPLACE(`',coln_,"`,'\\\\','\\\\\\\\')");
		SET BUF = CONCAT('REPLACE(',BUF,",'\\\'','\\\\\\\'')");
		SET BUF = CONCAT('REPLACE(',BUF,",'\"','\\\\\\\\\"')");
		SET BUF = CONCAT(' CONCAT("\'\\\"\",',BUF,',"\\\"\'\") as "', coln_, '"');
		
		RETURN BUF;
		
		
	END IF;
	RETURN CONCAT('`',coln_,'`');
END$$

DELIMITER ;

DELIMITER $$
DROP PROCEDURE IF EXISTS `ttab`.`pScriptDataToTab` $$
CREATE PROCEDURE `ttab`.`pScriptDataToTab`(schema_ VARCHAR(100), table_ VARCHAR(100))
proc:BEGIN
	DECLARE OLDLEN INT;
	
	DROP TEMPORARY TABLE IF EXISTS _lines;
	CREATE TEMPORARY TABLE _lines (line TEXT);
	
	SELECT @@group_concat_max_len INTO OLDLEN;
	
	SET @@group_concat_max_len = 16777216;
	
	-- leader
	INSERT INTO _lines VALUES (CONCAT("tab ",table_));
	INSERT INTO _lines VALUES (CONCAT("	,(dbname ",schema_,")"));
	
	-- output columns
	INSERT INTO _lines
	SELECT CONCAT('	,(col ',c.column_name,')')
	from information_schema.columns c
	where c.table_schema = schema_ and
	      c.table_name = table_ and
	      (c.extra != "AUTO_INCREMENT");
	
	DROP TEMPORARY TABLE IF EXISTS _data;
	-- output data
	SELECT CONCAT('CREATE TEMPORARY TABLE _data as SELECT '
			,GROUP_CONCAT(fScriptColumnDataSelect(c.column_name,c.column_type) SEPARATOR ',')
			,' FROM ',schema_,'.',table_)
	INTO @_QUERY
	from information_schema.columns c
	where c.table_schema = schema_ and
	      c.table_name = table_ and
	      (c.extra != "AUTO_INCREMENT");

	PREPARE stmt1 FROM @_QUERY;
	
	EXECUTE stmt1;

	DEALLOCATE PREPARE stmt1;
	
	SELECT CONCAT('INSERT INTO _lines SELECT CONCAT("	,(d ",COALESCE(`'
		,GROUP_CONCAT(c.column_name SEPARATOR '`,"NULL")," ",COALESCE(`')
		,'`,"NULL"),")") FROM _data')
	INTO @_QUERY
	from information_schema.columns c
	where c.table_schema = schema_ and
	      c.table_name = table_ and
	      (c.extra != "AUTO_INCREMENT");
	
	PREPARE stmt1 FROM @_QUERY;
	
	EXECUTE stmt1;

	DEALLOCATE PREPARE stmt1;
	
	select GROUP_CONCAT(line SEPARATOR '\n') as Script from _lines;	
	
	SET @@group_concat_max_len = OLDLEN;
	
	DROP TEMPORARY TABLE IF EXISTS _data;
	DROP TEMPORARY TABLE IF EXISTS _lines;
END$$
DELIMITER ;