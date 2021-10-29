CREATE DATABASE serverstorage;
use serverstorage;

CREATE TABLE User (
        Username VARCHAR(24) NOT NULL PRIMARY KEY,
        Display_Name VARCHAR(24) NOT NULL,
        Password VARCHAR(24) NOT NULL,
        Email VARCHAR(50) NOT NULL,
        Bio VARCHAR(250)
) ENGINE = InnoDB;


CREATE TABLE Communities (
        Community_ID BIGINT NOT NULL PRIMARY KEY,
        Name VARCHAR(24) NOT NULL,
        Description VARCHAR(250) NOT NULL,
        City VARCHAR(50) NOT NULL
) ENGINE = InnoDB;


CREATE TABLE Posts (
        Post_ID BIGINT NOT NULL PRIMARY KEY,
        Title VARCHAR(50) NOT NULL,
        Text VARCHAR(50),
        Community BIGINT NOT NULL,
        Username_FID VARCHAR(24) NOT NULL,
        CONSTRAINT `Post`
                FOREIGN KEY (Community) REFERENCES Communities (Community_ID)
                ON DELETE CASCADE
                ON UPDATE CASCADE,
        CONSTRAINT `Post2`
                FOREIGN KEY (Username_FID) REFERENCES User (Username)
                ON DELETE CASCADE
                ON UPDATE CASCADE
) ENGINE = InnoDB;

CREATE TABLE Sub_List (
    User_ID_Sub VARCHAR(24) NOT NULL,
    Community_Sub BIGINT NOT NULL,
    CONSTRAINT `List`    
        FOREIGN KEY (User_ID_Sub) REFERENCES User (Username)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT `List2`
        FOREIGN KEY (Community_Sub) REFERENCES Communities (Community_ID)
        ON DELETE CASCADE
        ON UPDATE CASCADE
) ENGINE = InnoDB;
