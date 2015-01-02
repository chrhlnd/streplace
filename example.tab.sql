-- ---->>> example.tab  -----
-- --- using:  mysql.gram  ----- 
DELIMITER $$
DROP PROCEDURE IF EXISTS `crm`.`_setup_account` $$
CREATE PROCEDURE `crm`.`_setup_account`()
proc:BEGIN
  SELECT "CHECKING table crm.account" as "Log";
  IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account") = 0 THEN

    SELECT "CREATING Missing table crm.account" as "Log";
    
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
      UNIQUE KEY `uidx_name` (`name` ASC),
      KEY `idx_name_last_invoice` (`name` ASC,`last_invoice` DESC)
    ) ENGINE=innodb AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8;
    -- FULL table create

  ELSE

      SELECT "CHECKING column crm.account.id" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "id") = 0 THEN
        SELECT "CREATING Missing column crm.account.id" as "Log";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT;
      END IF;
      
      SELECT "CHECKING column crm.account.name" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "name") = 0 THEN
        SELECT "CREATING Missing column crm.account.name" as "Log";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `name` VARCHAR(200) NULL AFTER `id`;
      END IF;
      
      SELECT "CHECKING column crm.account.balance" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "balance") = 0 THEN
        SELECT "CREATING Missing column crm.account.balance" as "Log";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `balance` BIGINT NOT NULL DEFAULT 0 AFTER `name`;
      END IF;
      
      SELECT "CHECKING column crm.account.important" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "important") = 0 THEN
        SELECT "CREATING Missing column crm.account.important" as "Log";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `important` BIT(1) NULL AFTER `balance`;
      END IF;
      
      SELECT "CHECKING column crm.account.main_phone" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "main_phone") = 0 THEN
        SELECT "CREATING Missing column crm.account.main_phone" as "Log";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `main_phone` VARCHAR(20) NULL AFTER `important`;
      END IF;
      
      SELECT "CHECKING column crm.account.main_email" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "main_email") = 0 THEN
        SELECT "CREATING Missing column crm.account.main_email" as "Log";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `main_email` VARCHAR(100) NULL AFTER `main_phone`;
      END IF;
      
      SELECT "CHECKING column crm.account.last_invoice" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "last_invoice") = 0 THEN
        SELECT "CREATING Missing column crm.account.last_invoice" as "Log";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `last_invoice` DATETIME NULL AFTER `main_email`;
      END IF;
      
      SELECT "CHECKING column crm.account.last_contact" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "last_contact") = 0 THEN
        SELECT "CREATING Missing column crm.account.last_contact" as "Log";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `last_contact` DATETIME NULL AFTER `last_invoice`;
      END IF;
      
      SELECT "CHECKING column crm.account.created" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account" AND column_name = "created") = 0 THEN
        SELECT "CREATING Missing column crm.account.created" as "Log";
        
        ALTER TABLE `crm`.`account`
         ADD COLUMN `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER `last_contact`;
        UPDATE `crm`.`account` SET `created` = CURRENT_TIMESTAMP;
      END IF;
      
      
  END IF;


  SELECT "Checking Unique constraint crm.account.uidx_name" as "Log";
  
  IF (SELECT count(*) FROM information_schema.statistics
       WHERE table_schema = "crm" AND
             table_name   = "account" AND
             column_name in ("name")
       GROUP BY index_name HAVING count(*) > 1) IS NULL 
    THEN
    SELECT "Creating Unique constraint crm.account.uidx_name" as "Log";
    
    CREATE UNIQUE INDEX `uidx_name`
     ON `crm`.`account`
     (`name` ASC);
    
    
  END IF;
  SELECT "Checking constraint crm.account.idx_name_last_invoice" as "Log";
  
  IF (SELECT count(*) FROM information_schema.statistics
       WHERE table_schema = "crm" AND
             table_name   = "account" AND
             column_name in ("name","last_invoice")
       GROUP BY index_name HAVING count(*) > 1+1) IS NULL 
    THEN
    SELECT "Creating constraint crm.account.idx_name_last_invoice" as "Log";
    
    CREATE INDEX `idx_name_last_invoice`
     ON `crm`.`account`
     (`name` ASC,`last_invoice` DESC);
    
  END IF;
  
END$$
DELIMITER ;

CALL `crm`.`_setup_account`();
DROP PROCEDURE `crm`.`_setup_account`;

DELIMITER $$
DROP PROCEDURE IF EXISTS `crm`.`_setup_account_note` $$
CREATE PROCEDURE `crm`.`_setup_account_note`()
proc:BEGIN
  SELECT "CHECKING table crm.account_note" as "Log";
  IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account_note") = 0 THEN

    SELECT "CREATING Missing table crm.account_note" as "Log";
    
    -- FULL table create
    CREATE TABLE `crm`.`account_note` (
      `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
      `account_id` BIGINT UNSIGNED NOT NULL,
      `note` TEXT NOT NULL,
      `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      PRIMARY KEY (`id` ASC),
      KEY `idx_account_id` (`account_id` ASC)
    ) ENGINE=innodb AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8;
    -- FULL table create

  ELSE

      SELECT "CHECKING column crm.account_note.id" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account_note" AND column_name = "id") = 0 THEN
        SELECT "CREATING Missing column crm.account_note.id" as "Log";
        
        ALTER TABLE `crm`.`account_note`
         ADD COLUMN `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT;
      END IF;
      
      SELECT "CHECKING column crm.account_note.account_id" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account_note" AND column_name = "account_id") = 0 THEN
        SELECT "CREATING Missing column crm.account_note.account_id" as "Log";
        
        ALTER TABLE `crm`.`account_note`
         ADD COLUMN `account_id` BIGINT UNSIGNED NOT NULL AFTER `id`;
      END IF;
      
      SELECT "CHECKING column crm.account_note.note" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account_note" AND column_name = "note") = 0 THEN
        SELECT "CREATING Missing column crm.account_note.note" as "Log";
        
        ALTER TABLE `crm`.`account_note`
         ADD COLUMN `note` TEXT NOT NULL AFTER `account_id`;
      END IF;
      
      SELECT "CHECKING column crm.account_note.created" as "Log";
      
      IF (SELECT count(*) FROM information_schema.columns WHERE table_schema = "crm" AND table_name = "account_note" AND column_name = "created") = 0 THEN
        SELECT "CREATING Missing column crm.account_note.created" as "Log";
        
        ALTER TABLE `crm`.`account_note`
         ADD COLUMN `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER `note`;
        UPDATE `crm`.`account_note` SET `created` = CURRENT_TIMESTAMP;
      END IF;
      
      
  END IF;


  SELECT "Checking constraint crm.account_note.idx_account_id" as "Log";
  
  IF (SELECT count(*) FROM information_schema.statistics
       WHERE table_schema = "crm" AND
             table_name   = "account_note" AND
             column_name in ("account_id")
       GROUP BY index_name HAVING count(*) > 1) IS NULL 
    THEN
    SELECT "Creating constraint crm.account_note.idx_account_id" as "Log";
    
    CREATE INDEX `idx_account_id`
     ON `crm`.`account_note`
     (`account_id` ASC);
    
  END IF;
  
END$$
DELIMITER ;

CALL `crm`.`_setup_account_note`();
DROP PROCEDURE `crm`.`_setup_account_note`;


-- ----<<<  example.tab  -----
