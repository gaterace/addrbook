use addrbook;

DROP TABLE IF EXISTS tb_Party;

-- address book party entity
CREATE TABLE tb_Party
(

    -- party identifier
    inbPartyId BIGINT AUTO_INCREMENT NOT NULL,
    -- creation date
    dtmCreated DATETIME NOT NULL,
    -- modification date
    dtmModified DATETIME NOT NULL,
    -- deletion date
    dtmDeleted DATETIME NOT NULL,
    -- has record been deleted?
    bitIsDeleted BOOL NOT NULL,
    -- version of this record
    intVersion INT NOT NULL,
    -- mservice account identifier
    inbMserviceId BIGINT NOT NULL,
    -- type of party record, int value of PartyType
    intPartyType INT NOT NULL,
    -- party last name
    chvLastName VARCHAR(50) NOT NULL,
    -- party middle name
    chvMiddleName VARCHAR(50) NOT NULL,
    -- party first name
    chvFirstName VARCHAR(50) NOT NULL,
    -- party nickname
    chvNickname VARCHAR(50) NOT NULL,
    -- party company
    chvCompany VARCHAR(100) NOT NULL,
    -- party email
    chvEmail VARCHAR(50) NOT NULL,


    PRIMARY KEY (inbPartyId),
    UNIQUE (inbMserviceId,inbPartyId)
) ENGINE=InnoDB;

