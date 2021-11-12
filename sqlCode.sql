CREATE DATABASE serverstorage;
use serverstorage;

CREATE TABLE User (
        Username VARCHAR(24) NOT NULL PRIMARY KEY,
        Display_Name VARCHAR(24) NOT NULL,
        Password VARCHAR(60) NOT NULL,
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
        Text VARCHAR(500),
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

CREATE TABLE Ads (
        Ad_ID BIGINT NOT NULL PRIMARY KEY,
        Title VARCHAR(50) NOT NULL,
        Text VARCHAR(600),
        Ad_City VARCHAR(50) NOT NULL
) ENGINE = InnoDB;


CREATE TABLE Events (
        Event_ID BIGINT NOT NULL PRIMARY KEY,
        Event_Name VARCHAR(50) NOT NULL,
        Date VARCHAR(15) NOT NULL, 
        Time VARCHAR(15) NOT NULL,
        Event_Location VARCHAR(600),
        What VARCHAR(600),
        Home_Community BIGINT NOT NULL,
        CONSTRAINT `Home_Community`
                FOREIGN KEY (Home_Community) REFERENCES Communities (Community_ID)
                ON DELETE CASCADE
                ON UPDATE CASCADE
) ENGINE = InnoDB;

ALTER TABLE Events MODIFY COLUMN Date VARCHAR(25) NOT NULL;
ALTER TABLE Communities ADD Longitude Float(10,10) SIGNED AFTER City;
ALTER TABLE Communities ADD Latitude Float(10,10) SIGNED AFTER Longitude;