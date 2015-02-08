--  begin  example.tab
--  applying  mysql.gram
DELIMITER $$
DROP PROCEDURE IF EXISTS `crm`.`_setup_account` $$
CREATE PROCEDURE `crm`.`_setup_account`()
proc:BEGIN
  SELECT "check TABLE crm.account";
  IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account") = 0 THEN

    SELECT "create TABLE crm.account";
    
    -- FULL table create
    CREATE TABLE `crm`.`account` (
      `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
      `name` VARCHAR(200) NULL,
      `balance` BIGINT NOT NULL DEFAULT 0,
      `important` BIT(1) NULL,
      `main_phone` VARCHAR(20) NULL,
      `main_email` VARCHAR(100) NULL,
      `last_invoice` DATETIME NULL,
      `last_contact` DATETIME NULL,
      `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      PRIMARY KEY (`id` ASC),
      UNIQUE KEY `uidx_b068931cc450442b63f5b3d276ea4297` (`name` ASC),
      KEY `idx_33408cfee9e6bc2255a4906e219aba39` (`name` ASC,`last_invoice` DESC)
    ) ENGINE=innodb AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8;
    -- FULL table create

  ELSE

      SELECT "check COLUMN crm.account.id";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "id") = 0 THEN
        SELECT "create COLUMN crm.account.id";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `id` BIGINT UNSIGNED NOT NULL;
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account"
         AND column_name = "id"
         AND column_type like replace("%BIGINT UNSIGNED%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NOT NULL"
         AND extra = "AUTO_INCREMENT"
         AND column_default IS NULL ) = 0 
      THEN
        SELECT "change COLUMN crm.account.id";
        
        ALTER TABLE `crm`.`account`
         CHANGE COLUMN `id` `id` BIGINT UNSIGNED NOT NULL;
      END IF;
      
      SELECT "check COLUMN crm.account.name";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "name") = 0 THEN
        SELECT "create COLUMN crm.account.name";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `name` VARCHAR(200) NULL AFTER `id`;
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account"
         AND column_name = "name"
         AND column_type like replace("%VARCHAR(200)%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NULL"
         AND extra != "AUTO_INCREMENT"
         AND column_default IS NULL ) = 0 
      THEN
        SELECT "change COLUMN crm.account.name";
        
        ALTER TABLE `crm`.`account`
         CHANGE COLUMN `name` `name` VARCHAR(200) NULL AFTER `id`;
      END IF;
      
      SELECT "check COLUMN crm.account.balance";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "balance") = 0 THEN
        SELECT "create COLUMN crm.account.balance";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `balance` BIGINT NOT NULL DEFAULT 0 AFTER `name`;
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account"
         AND column_name = "balance"
         AND column_type like replace("%BIGINT%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NOT NULL"
         AND extra != "AUTO_INCREMENT"
         AND column_default = 0) = 0 
      THEN
        SELECT "change COLUMN crm.account.balance";
        
        ALTER TABLE `crm`.`account`
         CHANGE COLUMN `balance` `balance` BIGINT NOT NULL DEFAULT 0 AFTER `name`;
      END IF;
      
      SELECT "check COLUMN crm.account.important";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "important") = 0 THEN
        SELECT "create COLUMN crm.account.important";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `important` BIT(1) NULL AFTER `balance`;
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account"
         AND column_name = "important"
         AND column_type like replace("%BIT(1)%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NULL"
         AND extra != "AUTO_INCREMENT"
         AND column_default IS NULL ) = 0 
      THEN
        SELECT "change COLUMN crm.account.important";
        
        ALTER TABLE `crm`.`account`
         CHANGE COLUMN `important` `important` BIT(1) NULL AFTER `balance`;
      END IF;
      
      SELECT "check COLUMN crm.account.main_phone";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "main_phone") = 0 THEN
        SELECT "create COLUMN crm.account.main_phone";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `main_phone` VARCHAR(20) NULL AFTER `important`;
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account"
         AND column_name = "main_phone"
         AND column_type like replace("%VARCHAR(20)%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NULL"
         AND extra != "AUTO_INCREMENT"
         AND column_default IS NULL ) = 0 
      THEN
        SELECT "change COLUMN crm.account.main_phone";
        
        ALTER TABLE `crm`.`account`
         CHANGE COLUMN `main_phone` `main_phone` VARCHAR(20) NULL AFTER `important`;
      END IF;
      
      SELECT "check COLUMN crm.account.main_email";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "main_email") = 0 THEN
        SELECT "create COLUMN crm.account.main_email";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `main_email` VARCHAR(100) NULL AFTER `main_phone`;
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account"
         AND column_name = "main_email"
         AND column_type like replace("%VARCHAR(100)%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NULL"
         AND extra != "AUTO_INCREMENT"
         AND column_default IS NULL ) = 0 
      THEN
        SELECT "change COLUMN crm.account.main_email";
        
        ALTER TABLE `crm`.`account`
         CHANGE COLUMN `main_email` `main_email` VARCHAR(100) NULL AFTER `main_phone`;
      END IF;
      
      SELECT "check COLUMN crm.account.last_invoice";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "last_invoice") = 0 THEN
        SELECT "create COLUMN crm.account.last_invoice";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `last_invoice` DATETIME NULL AFTER `main_email`;
        
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account"
         AND column_name = "last_invoice"
         AND column_type like replace("%DATETIME%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NULL"
         AND extra != "AUTO_INCREMENT"
         AND (column_default IS NULL OR column_default = "0000-00-00 00:00:00")) = 0 
      THEN
        SELECT "change COLUMN crm.account.last_invoice";
        
        ALTER TABLE `crm`.`account`
         CHANGE COLUMN `last_invoice` `last_invoice` DATETIME NULL AFTER `main_email`;
      END IF;
      
      SELECT "check COLUMN crm.account.last_contact";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "last_contact") = 0 THEN
        SELECT "create COLUMN crm.account.last_contact";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `last_contact` DATETIME NULL AFTER `last_invoice`;
        
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account"
         AND column_name = "last_contact"
         AND column_type like replace("%DATETIME%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NULL"
         AND extra != "AUTO_INCREMENT"
         AND (column_default IS NULL OR column_default = "0000-00-00 00:00:00")) = 0 
      THEN
        SELECT "change COLUMN crm.account.last_contact";
        
        ALTER TABLE `crm`.`account`
         CHANGE COLUMN `last_contact` `last_contact` DATETIME NULL AFTER `last_invoice`;
      END IF;
      
      SELECT "check COLUMN crm.account.created";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "created") = 0 THEN
        SELECT "create COLUMN crm.account.created";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER `last_contact`;
        UPDATE `crm`.`account` SET `created` = CURRENT_TIMESTAMP;
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account"
         AND column_name = "created"
         AND column_type like replace("%TIMESTAMP%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NOT NULL"
         AND extra != "AUTO_INCREMENT"
         AND column_default = "CURRENT_TIMESTAMP") = 0 
      THEN
        SELECT "change COLUMN crm.account.created";
        
        ALTER TABLE `crm`.`account`
         CHANGE COLUMN `created` `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER `last_contact`;
      END IF;
      
      
  END IF;


  SELECT "check PRIMARY INDEX crm.account.PRIMARY";
  
  IF (SELECT count(*) FROM information_schema.statistics
       WHERE table_schema = "crm" AND
             table_name   = "account" AND
             column_name in ("id") AND
             index_name = "PRIMARY"
       GROUP BY index_name HAVING count(*) > 1 - 1 LIMIT 1) IS NULL 
    THEN
    SELECT "adding PRIMARY KEY crm.account.PRIMARY";
    
    IF (SELECT count(*) FROM information_schema.statistics
         WHERE table_schema = "crm" AND
               table_name   = "account" AND
               index_name = "PRIMARY") = 0 THEN
      ALTER TABLE `crm`.`account` ADD PRIMARY KEY (id ASC);
    ELSE 
      ALTER TABLE `crm`.`account` DROP PRIMARY KEY, ADD PRIMARY KEY (id ASC);
    END IF;
  END IF;
  
  SELECT "check UNIQUE INDEX crm.account.uidx_b068931cc450442b63f5b3d276ea4297";
  
  IF (SELECT count(*) FROM information_schema.statistics
       WHERE table_schema = "crm" AND
             table_name   = "account" AND
             column_name in ("name")
       GROUP BY index_name HAVING count(*) > 1 - 1 LIMIT 1) IS NULL 
    THEN
    SELECT "create UNIQUE INDEX crm.account.uidx_b068931cc450442b63f5b3d276ea4297";
    
    CREATE UNIQUE INDEX `uidx_b068931cc450442b63f5b3d276ea4297`
     ON `crm`.`account`
     (`name` ASC);
    
    
  END IF;
  SELECT "check INDEX crm.account.idx_33408cfee9e6bc2255a4906e219aba39";
  
  IF (SELECT count(*) FROM information_schema.statistics
       WHERE table_schema = "crm" AND
             table_name   = "account" AND
             column_name in ("name","last_invoice")
       GROUP BY index_name HAVING count(*) > 1+1 - 1 LIMIT 1) IS NULL 
    THEN
    SELECT "create INDEX crm.account.idx_33408cfee9e6bc2255a4906e219aba39";
    
    CREATE INDEX `idx_33408cfee9e6bc2255a4906e219aba39`
     ON `crm`.`account`
     (`name` ASC,`last_invoice` DESC);
    
  END IF;
  
  SELECT "patching AUTO_INCREMENT crm.account";
  IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
   AND table_name = "account"
   AND column_name = "id"
   AND extra = "AUTO_INCREMENT") = 0 
  THEN
  ALTER TABLE `crm`.`account`
   CHANGE COLUMN `id` `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT ;
  END IF;
  
  
END$$
DELIMITER ;

CALL `crm`.`_setup_account`();
DROP PROCEDURE `crm`.`_setup_account`;

SELECT "DONE with account";

DELIMITER $$
DROP PROCEDURE IF EXISTS `crm`.`_setup_account_note` $$
CREATE PROCEDURE `crm`.`_setup_account_note`()
proc:BEGIN
  SELECT "check TABLE crm.account_note";
  IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account_note") = 0 THEN

    SELECT "create TABLE crm.account_note";
    
    -- FULL table create
    CREATE TABLE `crm`.`account_note` (
      `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
      `account_id` BIGINT UNSIGNED NOT NULL,
      `note` TEXT NOT NULL,
      `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      PRIMARY KEY (`id` ASC),
      KEY `idx_8a089c2a7e6c77be2e2e68c5c366f460` (`account_id` ASC)
    ) ENGINE=innodb AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8;
    -- FULL table create

  ELSE

      SELECT "check COLUMN crm.account_note.id";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account_note" AND column_name = "id") = 0 THEN
        SELECT "create COLUMN crm.account_note.id";
        
        ALTER TABLE `crm`.`account_note`
         ADD COLUMN `id` BIGINT UNSIGNED NOT NULL;
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account_note"
         AND column_name = "id"
         AND column_type like replace("%BIGINT UNSIGNED%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NOT NULL"
         AND extra = "AUTO_INCREMENT"
         AND column_default IS NULL ) = 0 
      THEN
        SELECT "change COLUMN crm.account_note.id";
        
        ALTER TABLE `crm`.`account_note`
         CHANGE COLUMN `id` `id` BIGINT UNSIGNED NOT NULL;
      END IF;
      
      SELECT "check COLUMN crm.account_note.account_id";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account_note" AND column_name = "account_id") = 0 THEN
        SELECT "create COLUMN crm.account_note.account_id";
        
        ALTER TABLE `crm`.`account_note`
         ADD COLUMN `account_id` BIGINT UNSIGNED NOT NULL AFTER `id`;
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account_note"
         AND column_name = "account_id"
         AND column_type like replace("%BIGINT UNSIGNED%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NOT NULL"
         AND extra != "AUTO_INCREMENT"
         AND column_default IS NULL ) = 0 
      THEN
        SELECT "change COLUMN crm.account_note.account_id";
        
        ALTER TABLE `crm`.`account_note`
         CHANGE COLUMN `account_id` `account_id` BIGINT UNSIGNED NOT NULL AFTER `id`;
      END IF;
      
      SELECT "check COLUMN crm.account_note.note";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account_note" AND column_name = "note") = 0 THEN
        SELECT "create COLUMN crm.account_note.note";
        
        ALTER TABLE `crm`.`account_note`
         ADD COLUMN `note` TEXT NOT NULL AFTER `account_id`;
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account_note"
         AND column_name = "note"
         AND column_type like replace("%TEXT%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NOT NULL"
         AND extra != "AUTO_INCREMENT"
         AND column_default IS NULL ) = 0 
      THEN
        SELECT "change COLUMN crm.account_note.note";
        
        ALTER TABLE `crm`.`account_note`
         CHANGE COLUMN `note` `note` TEXT NOT NULL AFTER `account_id`;
      END IF;
      
      SELECT "check COLUMN crm.account_note.created";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account_note" AND column_name = "created") = 0 THEN
        SELECT "create COLUMN crm.account_note.created";
        
        ALTER TABLE `crm`.`account_note`
         ADD COLUMN `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER `note`;
        UPDATE `crm`.`account_note` SET `created` = CURRENT_TIMESTAMP;
      END IF;
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
         AND table_name = "account_note"
         AND column_name = "created"
         AND column_type like replace("%TIMESTAMP%"," ","%")
         AND replace(replace(is_nullable,"NO","NOT NULL"),"YES","NULL") = "NOT NULL"
         AND extra != "AUTO_INCREMENT"
         AND column_default = "CURRENT_TIMESTAMP") = 0 
      THEN
        SELECT "change COLUMN crm.account_note.created";
        
        ALTER TABLE `crm`.`account_note`
         CHANGE COLUMN `created` `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER `note`;
      END IF;
      
      
  END IF;


  SELECT "check PRIMARY INDEX crm.account_note.PRIMARY";
  
  IF (SELECT count(*) FROM information_schema.statistics
       WHERE table_schema = "crm" AND
             table_name   = "account_note" AND
             column_name in ("id") AND
             index_name = "PRIMARY"
       GROUP BY index_name HAVING count(*) > 1 - 1 LIMIT 1) IS NULL 
    THEN
    SELECT "adding PRIMARY KEY crm.account_note.PRIMARY";
    
    IF (SELECT count(*) FROM information_schema.statistics
         WHERE table_schema = "crm" AND
               table_name   = "account_note" AND
               index_name = "PRIMARY") = 0 THEN
      ALTER TABLE `crm`.`account_note` ADD PRIMARY KEY (id ASC);
    ELSE 
      ALTER TABLE `crm`.`account_note` DROP PRIMARY KEY, ADD PRIMARY KEY (id ASC);
    END IF;
  END IF;
  
  SELECT "check INDEX crm.account_note.idx_8a089c2a7e6c77be2e2e68c5c366f460";
  
  IF (SELECT count(*) FROM information_schema.statistics
       WHERE table_schema = "crm" AND
             table_name   = "account_note" AND
             column_name in ("account_id")
       GROUP BY index_name HAVING count(*) > 1 - 1 LIMIT 1) IS NULL 
    THEN
    SELECT "create INDEX crm.account_note.idx_8a089c2a7e6c77be2e2e68c5c366f460";
    
    CREATE INDEX `idx_8a089c2a7e6c77be2e2e68c5c366f460`
     ON `crm`.`account_note`
     (`account_id` ASC);
    
  END IF;
  
  SELECT "patching AUTO_INCREMENT crm.account_note";
  IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm"
   AND table_name = "account_note"
   AND column_name = "id"
   AND extra = "AUTO_INCREMENT") = 0 
  THEN
  ALTER TABLE `crm`.`account_note`
   CHANGE COLUMN `id` `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT ;
  END IF;
  
  
END$$
DELIMITER ;

CALL `crm`.`_setup_account_note`();
DROP PROCEDURE `crm`.`_setup_account_note`;

SELECT "DONE with account_note";

--  end  example.tab
